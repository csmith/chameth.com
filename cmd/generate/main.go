package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type provider struct {
	importPath string
	typeName   string
	fieldName  string
}

type pkg struct {
	importPath string
	alias      string
	// params maps each registration function the package defines to the
	// indices of the providers it takes.
	params map[string][]int
}

func main() {
	root := repoRoot()
	mod := moduleName(filepath.Join(root, "go.mod"))
	providers := parseSite(root)
	gen(root, scan(root, mod, providers), providers)
}

func die(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func repoRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		die("failed to get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			die("could not find go.mod")
		}
		dir = parent
	}
}

func moduleName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		die("failed to read go.mod: %v", err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after)
		}
	}
	die("module directive not found in go.mod")
	return ""
}

func makeAlias(mod, importPath string) string {
	parts := strings.Split(strings.TrimPrefix(importPath, mod+"/"), "/")
	var a strings.Builder
	a.WriteString(parts[0])
	for _, p := range parts[1:] {
		a.WriteString(strings.ToUpper(p[:1]))
		a.WriteString(p[1:])
	}
	return a.String()
}

func parseSite(root string) []provider {
	siteFile := filepath.Join(root, "cmd", "serve", "site.go")
	f, err := parser.ParseFile(token.NewFileSet(), siteFile, nil, 0)
	if err != nil {
		die("failed to parse site.go: %v", err)
	}

	fileImports := fileImports(f)

	var providers []provider
	for _, d := range f.Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != "site" {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok {
				continue
			}
			providers = append(providers, providersFromFields(st.Fields.List, fileImports)...)
		}
	}

	if len(providers) == 0 {
		die("no providers found in site struct")
	}
	return providers
}

func fileImports(f *ast.File) map[string]string {
	imports := map[string]string{}
	for _, imp := range f.Imports {
		impPath := strings.Trim(imp.Path.Value, "\"")
		if imp.Name != nil {
			imports[imp.Name.Name] = impPath
		} else {
			parts := strings.Split(impPath, "/")
			imports[parts[len(parts)-1]] = impPath
		}
	}
	return imports
}

func providersFromFields(fields []*ast.Field, fileImports map[string]string) []provider {
	var providers []provider
	for _, field := range fields {
		if len(field.Names) != 1 {
			continue
		}
		if p, ok := providerForField(field, fileImports); ok {
			providers = append(providers, p)
		}
	}
	return providers
}

func providerForField(field *ast.Field, fileImports map[string]string) (provider, bool) {
	typeExpr := field.Type
	if star, ok := typeExpr.(*ast.StarExpr); ok {
		typeExpr = star.X
	}
	sel, ok := typeExpr.(*ast.SelectorExpr)
	if !ok {
		return provider{}, false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return provider{}, false
	}
	impPath, ok := fileImports[ident.Name]
	if !ok {
		return provider{}, false
	}
	return provider{
		importPath: impPath,
		typeName:   sel.Sel.Name,
		fieldName:  field.Names[0].Name,
	}, true
}

func matchProviders(fn *ast.FuncDecl, fileImports map[string]string, providers []provider, pkgImportPath string) []int {
	if fn.Type.Params == nil || len(fn.Type.Params.List) == 0 {
		return nil
	}
	var matched []int
	for _, param := range fn.Type.Params.List {
		typeExpr := param.Type
		if star, ok := typeExpr.(*ast.StarExpr); ok {
			typeExpr = star.X
		}
		switch x := typeExpr.(type) {
		case *ast.SelectorExpr:
			ident, ok := x.X.(*ast.Ident)
			if !ok {
				continue
			}
			impPath, ok := fileImports[ident.Name]
			if !ok {
				continue
			}
			for i, p := range providers {
				if impPath == p.importPath && x.Sel.Name == p.typeName {
					matched = append(matched, i)
					break
				}
			}
		case *ast.Ident:
			for i, p := range providers {
				if pkgImportPath == p.importPath && x.Name == p.typeName {
					matched = append(matched, i)
					break
				}
			}
		}
	}
	return matched
}

func scan(root, mod string, providers []provider) map[string]*pkg {
	pkgs := map[string]*pkg{}
	filepath.Walk(
		root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if skippedDir(info.Name()) {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			scanFile(root, mod, path, providers, pkgs)
			return nil
		},
	)
	return pkgs
}

func skippedDir(name string) bool {
	// Like the go tool, ignore directories starting with a dot or underscore.
	return name == "tsdata" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// scanFile records the registration functions found in a single Go file. A
// package whose files register nothing is removed from the map again.
func scanFile(root, mod, path string, providers []provider, pkgs map[string]*pkg) {
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return
	}

	fileImports := fileImports(f)

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return
	}
	dir := filepath.Dir(rel)
	if dir == "." {
		return
	}
	importPath := mod + "/" + filepath.ToSlash(dir)

	p, ok := pkgs[importPath]
	if !ok {
		p = &pkg{
			importPath: importPath,
			alias:      makeAlias(mod, importPath),
			params:     map[string][]int{},
		}
		pkgs[importPath] = p
	}

	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || !isRegistration(fn.Name.Name) {
			continue
		}
		p.params[fn.Name.Name] = append(
			p.params[fn.Name.Name],
			matchProviders(fn, fileImports, providers, importPath)...,
		)
	}

	if len(p.params) == 0 {
		delete(pkgs, importPath)
	}
}

// registrations lists each registration function and the site method that
// calls it, in the order the methods are generated.
var registrations = []struct {
	funcName   string
	methodName string
	call       string
}{
	{"RegisterAssets", "registerAssets", "%s.RegisterAssets(%s)"},
	{"RegisterShortcodes", "registerShortcodes", "%s.RegisterShortcodes(%s)"},
	{"RegisterRoutes", "registerRoutes", "%s.RegisterRoutes(%s)"},
	{"RegisterGoroutine", "launchGoroutines", "go %s.RegisterGoroutine(%s)()"},
	{"RegisterContentTypes", "registerContentTypes", "%s.RegisterContentTypes(%s)"},
}

func isRegistration(name string) bool {
	for _, r := range registrations {
		if r.funcName == name {
			return true
		}
	}
	return false
}

func buildArgs(paramIndices []int, providers []provider) string {
	args := make([]string, 0, len(paramIndices))
	for _, idx := range paramIndices {
		args = append(args, "s."+providers[idx].fieldName)
	}
	return strings.Join(args, ", ")
}

func gen(root string, pkgs map[string]*pkg, providers []provider) {
	sortedPaths := make([]string, 0, len(pkgs))
	for p := range pkgs {
		sortedPaths = append(sortedPaths, p)
	}
	sort.Strings(sortedPaths)

	var buf bytes.Buffer
	buf.WriteString("// Code generated by cmd/generate. DO NOT EDIT.\n")
	buf.WriteString("//go:generate go run ../../cmd/generate\n\n")
	buf.WriteString("package main\n\n")
	buf.WriteString("import (\n")
	for _, path := range sortedPaths {
		fmt.Fprintf(&buf, "\t%s %q\n", pkgs[path].alias, path)
	}
	buf.WriteString(")\n")

	for _, r := range registrations {
		fmt.Fprintf(&buf, "\nfunc (s *site) %s() {\n", r.methodName)
		for _, path := range sortedPaths {
			p := pkgs[path]
			if params, ok := p.params[r.funcName]; ok {
				fmt.Fprintf(&buf, "\t"+r.call+"\n", p.alias, buildArgs(params, providers))
			}
		}
		buf.WriteString("}\n")
	}

	out := filepath.Join(root, "cmd", "serve", "register.go")
	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		die("failed to write %s: %v", out, err)
	}
}
