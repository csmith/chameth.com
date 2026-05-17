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
	varName    string
}

type pkg struct {
	importPath    string
	alias         string
	hasShortcodes bool
	hasAssets     bool
	hasRoutes     bool
	routeParams   []int
}

func main() {
	root := repoRoot()
	mod := moduleName(filepath.Join(root, "go.mod"))
	providers := []provider{
		{importPath: mod + "/assets", typeName: "Manager", varName: "assetsManager"},
	}
	pkgs := scan(root, mod, providers)
	gen(root, mod, pkgs, providers)
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
		a.WriteString(strings.ToUpper(p[:1]) + p[1:])
	}
	return a.String()
}

func matchProviders(fn *ast.FuncDecl, fileImports map[string]string, providers []provider) []int {
	if fn.Type.Params == nil || len(fn.Type.Params.List) < 2 {
		return nil
	}
	var matched []int
	for _, param := range fn.Type.Params.List[1:] {
		star, ok := param.Type.(*ast.StarExpr)
		if !ok {
			continue
		}
		sel, ok := star.X.(*ast.SelectorExpr)
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
		for i, p := range providers {
			if impPath == p.importPath && sel.Sel.Name == p.typeName {
				matched = append(matched, i)
				break
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

		var sc, as, rt bool
		var rp []int
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			switch fn.Name.Name {
			case "RegisterShortcodes":
				sc = true
			case "RegisterAssets":
				as = true
			case "RegisterRoutes":
				rt = true
				rp = matchProviders(fn, fileImports, providers)
			}
		}
		if !sc && !as && !rt {
			return nil
		}

		p, ok := pkgs[importPath]
		if !ok {
			p = &pkg{importPath: importPath, alias: makeAlias(mod, importPath)}
			pkgs[importPath] = p
		}
		p.hasShortcodes = p.hasShortcodes || sc
		p.hasAssets = p.hasAssets || as
		p.hasRoutes = p.hasRoutes || rt
		p.routeParams = append(p.routeParams, rp...)
		return nil
	})
	return pkgs
}

func gen(root, mod string, pkgs map[string]*pkg, providers []provider) {
	shortcodesPath := mod + "/features/shortcodes"
	shortcodesAlias := makeAlias(mod, shortcodesPath)

	allImports := map[string]string{}
	for _, p := range pkgs {
		allImports[p.alias] = p.importPath
	}
	allImports[shortcodesAlias] = shortcodesPath

	usedProviders := map[int]bool{}
	for _, p := range pkgs {
		for _, idx := range p.routeParams {
			usedProviders[idx] = true
		}
	}

	hasRoutes := false
	for _, p := range pkgs {
		if p.hasRoutes {
			hasRoutes = true
			break
		}
	}

	if hasRoutes {
		allImports["http"] = "net/http"
		for idx := range usedProviders {
			allImports[makeAlias(mod, providers[idx].importPath)] = providers[idx].importPath
		}
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

	var assetPkgs, shortcodePkgs []*pkg
	for _, path := range sortedPaths {
		p := pkgs[path]
		if p.hasAssets {
			assetPkgs = append(assetPkgs, p)
		}
		if p.hasShortcodes {
			shortcodePkgs = append(shortcodePkgs, p)
		}
	}

	buf.WriteString("func registerAssets(mgr *assets.Manager) {\n")
	for _, p := range assetPkgs {
		fmt.Fprintf(&buf, "\t%s.RegisterAssets(mgr)\n", p.alias)
	}
	buf.WriteString("}\n\n")

	fmt.Fprintf(&buf, "func registerShortcodes(mgr *%s.Manager) {\n", shortcodesAlias)
	for _, p := range shortcodePkgs {
		fmt.Fprintf(&buf, "\t%s.RegisterShortcodes(mgr)\n", p.alias)
	}
	buf.WriteString("}\n\n")

	var routePkgs []*pkg
	for _, path := range sortedPaths {
		p := pkgs[path]
		if p.hasRoutes {
			routePkgs = append(routePkgs, p)
		}
	}

	if len(routePkgs) > 0 {
		var params []string
		params = append(params, "mux *http.ServeMux")
		sortedUsedProviders := make([]int, 0, len(usedProviders))
		for idx := range usedProviders {
			sortedUsedProviders = append(sortedUsedProviders, idx)
		}
		sort.Ints(sortedUsedProviders)
		for _, idx := range sortedUsedProviders {
			prov := providers[idx]
			params = append(params, fmt.Sprintf("%s *%s.%s", prov.varName, makeAlias(mod, prov.importPath), prov.typeName))
		}

		fmt.Fprintf(&buf, "func registerRoutes(%s) {\n", strings.Join(params, ", "))
		for _, p := range routePkgs {
			args := []string{"mux"}
			for _, idx := range p.routeParams {
				args = append(args, providers[idx].varName)
			}
			fmt.Fprintf(&buf, "\t%s.RegisterRoutes(%s)\n", p.alias, strings.Join(args, ", "))
		}
		buf.WriteString("}\n")
	}

	out := filepath.Join(root, "cmd", "serve", "register.go")
	if err := os.WriteFile(out, buf.Bytes(), 0644); err != nil {
		die("failed to write %s: %v", out, err)
	}
}
