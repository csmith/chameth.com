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
	importPath        string
	alias             string
	hasShortcodes     bool
	hasAssets         bool
	hasRoutes         bool
	hasGoroutines     bool
	hasContentTypes   bool
	shortcodeParams   []int
	assetParams       []int
	routeParams       []int
	goroutineParams   []int
	contentTypeParams []int
}

func main() {
	root := repoRoot()
	mod := moduleName(filepath.Join(root, "go.mod"))
	providers := parseSite(root)
	pkgs := scan(root, mod, providers)
	gen(root, pkgs, providers)
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
	rel := strings.TrimPrefix(importPath, mod+"/")
	parts := strings.Split(rel, "/")
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
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, siteFile, nil, 0)
	if err != nil {
		die("failed to parse site.go: %v", err)
	}

	fileImports := map[string]string{}
	for _, imp := range f.Imports {
		impPath := strings.Trim(imp.Path.Value, "\"")
		if imp.Name != nil {
			fileImports[imp.Name.Name] = impPath
		} else {
			parts := strings.Split(impPath, "/")
			fileImports[parts[len(parts)-1]] = impPath
		}
	}

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
			for _, field := range st.Fields.List {
				if len(field.Names) != 1 {
					continue
				}
				typeExpr := field.Type
				if star, ok := typeExpr.(*ast.StarExpr); ok {
					typeExpr = star.X
				}
				sel, ok := typeExpr.(*ast.SelectorExpr)
				if !ok {
					continue
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					continue
				}
				impPath, ok := fileImports[ident.Name]
				if !ok {
					continue
				}
				providers = append(providers, provider{
					importPath: impPath,
					typeName:   sel.Sel.Name,
					fieldName:  field.Names[0].Name,
				})
			}
		}
	}

	if len(providers) == 0 {
		die("no providers found in site struct")
	}
	return providers
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
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			switch info.Name() {
			case ".git", ".postgres", "tsdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}

		fileImports := map[string]string{}
		for _, imp := range f.Imports {
			impPath := strings.Trim(imp.Path.Value, "\"")
			if imp.Name != nil {
				fileImports[imp.Name.Name] = impPath
			} else {
				parts := strings.Split(impPath, "/")
				fileImports[parts[len(parts)-1]] = impPath
			}
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		dir := filepath.Dir(rel)
		if dir == "." {
			return nil
		}
		importPath := mod + "/" + filepath.ToSlash(dir)

		p, ok := pkgs[importPath]
		if !ok {
			p = &pkg{importPath: importPath, alias: makeAlias(mod, importPath)}
			pkgs[importPath] = p
		}

		var found bool
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			switch fn.Name.Name {
			case "RegisterShortcodes":
				found = true
				p.hasShortcodes = true
				p.shortcodeParams = append(p.shortcodeParams, matchProviders(fn, fileImports, providers, importPath)...)
			case "RegisterAssets":
				found = true
				p.hasAssets = true
				p.assetParams = append(p.assetParams, matchProviders(fn, fileImports, providers, importPath)...)
			case "RegisterRoutes":
				found = true
				p.hasRoutes = true
				p.routeParams = append(p.routeParams, matchProviders(fn, fileImports, providers, importPath)...)
			case "RegisterGoroutine":
				found = true
				p.hasGoroutines = true
				p.goroutineParams = append(p.goroutineParams, matchProviders(fn, fileImports, providers, importPath)...)
			case "RegisterContentTypes":
				found = true
				p.hasContentTypes = true
				p.contentTypeParams = append(p.contentTypeParams, matchProviders(fn, fileImports, providers, importPath)...)
			}
		}
		if !found && !p.hasShortcodes && !p.hasAssets && !p.hasRoutes && !p.hasGoroutines && !p.hasContentTypes {
			delete(pkgs, importPath)
		}
		return nil
	})
	return pkgs
}

func buildArgs(paramIndices []int, providers []provider) string {
	var args []string
	for _, idx := range paramIndices {
		args = append(args, "s."+providers[idx].fieldName)
	}
	return strings.Join(args, ", ")
}

func gen(root string, pkgs map[string]*pkg, providers []provider) {
	allImports := map[string]string{}
	for _, p := range pkgs {
		allImports[p.alias] = p.importPath
	}

	sortedAliases := make([]string, 0, len(allImports))
	for a := range allImports {
		sortedAliases = append(sortedAliases, a)
	}
	sort.Strings(sortedAliases)

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
	for _, a := range sortedAliases {
		fmt.Fprintf(&buf, "\t%s %q\n", a, allImports[a])
	}
	buf.WriteString(")\n\n")

	var assetPkgs, shortcodePkgs, routePkgs, goroutinePkgs, contentTypePkgs []*pkg
	for _, path := range sortedPaths {
		p := pkgs[path]
		if p.hasAssets {
			assetPkgs = append(assetPkgs, p)
		}
		if p.hasShortcodes {
			shortcodePkgs = append(shortcodePkgs, p)
		}
		if p.hasRoutes {
			routePkgs = append(routePkgs, p)
		}
		if p.hasGoroutines {
			goroutinePkgs = append(goroutinePkgs, p)
		}
		if p.hasContentTypes {
			contentTypePkgs = append(contentTypePkgs, p)
		}
	}

	buf.WriteString("func (s *site) registerAssets() {\n")
	for _, p := range assetPkgs {
		fmt.Fprintf(&buf, "\t%s.RegisterAssets(%s)\n", p.alias, buildArgs(p.assetParams, providers))
	}
	buf.WriteString("}\n\n")

	buf.WriteString("func (s *site) registerShortcodes() {\n")
	for _, p := range shortcodePkgs {
		fmt.Fprintf(&buf, "\t%s.RegisterShortcodes(%s)\n", p.alias, buildArgs(p.shortcodeParams, providers))
	}
	buf.WriteString("}\n\n")

	if len(routePkgs) > 0 {
		buf.WriteString("func (s *site) registerRoutes() {\n")
		for _, p := range routePkgs {
			fmt.Fprintf(&buf, "\t%s.RegisterRoutes(%s)\n", p.alias, buildArgs(p.routeParams, providers))
		}
		buf.WriteString("}\n")
	}

	if len(goroutinePkgs) > 0 {
		buf.WriteString("\nfunc (s *site) launchGoroutines() {\n")
		for _, p := range goroutinePkgs {
			fmt.Fprintf(&buf, "\tgo %s.RegisterGoroutine(%s)()\n", p.alias, buildArgs(p.goroutineParams, providers))
		}
		buf.WriteString("}\n")
	}

	if len(contentTypePkgs) > 0 {
		buf.WriteString("\nfunc (s *site) registerContentTypes() {\n")
		for _, p := range contentTypePkgs {
			fmt.Fprintf(&buf, "\t%s.RegisterContentTypes(%s)\n", p.alias, buildArgs(p.contentTypeParams, providers))
		}
		buf.WriteString("}\n")
	}

	out := filepath.Join(root, "cmd", "serve", "register.go")
	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		die("failed to write %s: %v", out, err)
	}
}
