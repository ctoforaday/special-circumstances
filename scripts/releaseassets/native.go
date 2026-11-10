package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

const (
	feovTools        = "plugins/frank-exchange-of-views/tools"
	nativehelperPkg  = "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	nativehelperName = "nativehelper"
)

// NativeSet is the paths whose change makes a pull request build every helper for every target
// and run native-run and native-fuzz (plan §III.1.5): nativedigest's inputs, nativedigest itself,
// internal/nativehelper, the directory of every package importing internal/nativehelper (its
// tests included), internal/tessocr's whole tree (the corpus: its pages, the golden render format,
// and every target's goldens, which only native-run checks), the page renderer that bounds every
// feov-ocr request, the jobs' own tooling and the workflow with the action it installs helpers
// through.
type NativeSet struct {
	// Trees hold every path under them; Dirs hold only the files directly in them (a package is
	// one directory, never its subdirectories); Files are single paths. All are repository-relative
	// with forward slashes.
	Trees, Dirs, Files []string
}

// Contains reports whether a repository-relative path is in the set.
func (s NativeSet) Contains(p string) bool {
	p = filepath.ToSlash(p)
	for _, t := range s.Trees {
		if strings.HasPrefix(p, t+"/") {
			return true
		}
	}
	if slices.Contains(s.Dirs, path.Dir(p)) {
		return true
	}
	return slices.Contains(s.Files, p)
}

// Touched reports whether any changed file is in the set.
func (s NativeSet) Touched(changed []string) bool {
	return slices.ContainsFunc(changed, s.Contains)
}

// DeriveNativeSet builds the set for the repository at root. The importers of internal/nativehelper
// come from `go list` over FEOV's tools module, test imports included, so a package joins the set
// by importing it.
func DeriveNativeSet(root string) (NativeSet, error) {
	cmd := exec.Command("go", "list", "-json=ImportPath,Dir,Imports,TestImports,XTestImports", "./...")
	cmd.Dir = filepath.Join(root, feovTools)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		return NativeSet{}, fmt.Errorf("go list over %s: %v %s", feovTools, err, stderr)
	}
	s := NativeSet{
		Trees: []string{
			"plugins/frank-exchange-of-views/native",
			feovTools + "/devcmd/nativedigest",
			feovTools + "/internal/nativehelper",
			feovTools + "/internal/tessocr",
			"scripts/releaseassets",
			"scripts/internal/staticproof",
			"scripts/staticproof",
			".github/actions/install-helpers",
		},
		Files: []string{
			feovTools + "/internal/fetchcache/pagerender.go",
			".github/workflows/hooks.yml",
		},
	}
	dec := json.NewDecoder(strings.NewReader(string(out)))
	for {
		var p struct {
			ImportPath, Dir                    string
			Imports, TestImports, XTestImports []string
		}
		if err := dec.Decode(&p); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return NativeSet{}, fmt.Errorf("reading go list: %w", err)
		}
		if p.ImportPath == nativehelperPkg {
			continue
		}
		if slices.Contains(p.Imports, nativehelperPkg) || slices.Contains(p.TestImports, nativehelperPkg) || slices.Contains(p.XTestImports, nativehelperPkg) {
			rel, err := filepath.Rel(root, p.Dir)
			if err != nil {
				return NativeSet{}, err
			}
			s.Dirs = append(s.Dirs, filepath.ToSlash(rel))
		}
	}
	if len(s.Dirs) == 0 {
		return NativeSet{}, fmt.Errorf("no package in %s imports %s, so the native set has no client", feovTools, nativehelperPkg)
	}
	slices.Sort(s.Dirs)
	return s, nil
}

// RequirePackages is every package directory under FEOV's tools module whose Go files call
// nativehelper.Require (Require itself, inside package nativehelper), as ./-relative paths from
// the tools directory: the packages native-run runs. Found by parsing each file's calls, so a
// comment or a string naming Require does not count.
func RequirePackages(root string) ([]string, error) {
	tools := filepath.Join(root, feovTools)
	found := map[string]bool{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(tools, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		if callsRequire(f) {
			rel, err := filepath.Rel(tools, filepath.Dir(p))
			if err != nil {
				return err
			}
			found["./"+filepath.ToSlash(rel)] = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var pkgs []string
	for p := range found {
		pkgs = append(pkgs, p)
	}
	slices.Sort(pkgs)
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no file under %s calls nativehelper.Require", feovTools)
	}
	return pkgs, nil
}

func callsRequire(f *ast.File) bool {
	inPackage := f.Name.Name == nativehelperName
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || found {
			return !found
		}
		switch fn := call.Fun.(type) {
		case *ast.SelectorExpr:
			if id, ok := fn.X.(*ast.Ident); ok && id.Name == nativehelperName && fn.Sel.Name == "Require" {
				found = true
			}
		case *ast.Ident:
			if inPackage && fn.Name == "Require" {
				found = true
			}
		}
		return !found
	})
	return found
}

// TestFuncs is the names of every top-level `func TestX(t *testing.T)` in the _test.go files of the
// package directories given, relative to FEOV's tools directory.
func TestFuncs(root string, pkgs []string) (map[string]string, error) {
	tools := filepath.Join(root, feovTools)
	fset := token.NewFileSet()
	names := map[string]string{}
	for _, pkg := range pkgs {
		dir := filepath.Join(tools, filepath.FromSlash(pkg))
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			f, err := parser.ParseFile(fset, filepath.Join(dir, e.Name()), nil, parser.SkipObjectResolution)
			if err != nil {
				return nil, err
			}
			for _, d := range f.Decls {
				fn, ok := d.(*ast.FuncDecl)
				if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") || !takesTestingT(fn) {
					continue
				}
				names[fn.Name.Name] = pkg
			}
		}
	}
	return names, nil
}

func takesTestingT(fn *ast.FuncDecl) bool {
	params := fn.Type.Params.List
	if len(params) != 1 || len(params[0].Names) > 1 {
		return false
	}
	star, ok := params[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	sel, ok := star.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "testing" && sel.Sel.Name == "T"
}
