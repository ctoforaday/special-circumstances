package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// NO CHECKOUT-BUILT TELEPATHY RUNS AGAINST THE HOST'S RUNTIME DIRECTORY.
//
// Every verb prunes the Antigravity server records under $XDG_RUNTIME_DIR, and a test that execs
// the built binary with the inherited environment prunes the host's records with an unreleased
// build. This reads every test file in the module that is not a live test (those carry
// HOOKCORE_LIVE and build their own environment through testhome) and fails any exec that may run
// the built telepathy whose command's Env has no XDG_RUNTIME_DIR entry. An exec may run it when its
// program names telepathy in any form, or when the file does not show the program to be another
// one (a helper's parameter, a field, a call's result). A sweep that finds no exec at all fails
// too: a pattern that stopped matching would otherwise read as a clean module.
func TestBuiltBinaryNeverSeesHostRuntimeDir(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	sites := 0
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if name := d.Name(); name == "testdata" || (strings.HasPrefix(name, ".") && path != root) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		if fileMentions(f, "HOOKCORE_LIVE") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		n, bad := sweepFile(f)
		sites += n
		for _, b := range bad {
			t.Errorf("%s: %s", filepath.ToSlash(rel), b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if sites == 0 {
		t.Fatal("found no exec of the built telepathy in the module's tests: the sweep no longer recognises one, so it measured nothing")
	}
}

// sweepFile counts the execs in f that may run the built telepathy and describes each one whose
// command has no XDG_RUNTIME_DIR entry in its Env.
func sweepFile(f *ast.File) (int, []string) {
	names := binaryNames(f)
	others := otherPrograms(f, names)
	sites := 0
	var bad []string
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		n, b := execsWithoutRuntimeDir(fn.Body, names, others)
		sites += n
		for _, d := range b {
			bad = append(bad, fn.Name.Name+" execs what may be the built telepathy "+d)
		}
	}
	return sites, bad
}

// otherPrograms is every identifier in f assigned an expression that holds a string literal and
// names no telepathy, such as exec.LookPath("rg"): a program shown to be another one.
func otherPrograms(f *ast.File, names map[string]bool) map[string]bool {
	others := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range as.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || names[id.Name] {
				continue
			}
			rhs := as.Rhs[0]
			if i < len(as.Rhs) {
				rhs = as.Rhs[i]
			}
			if holdsStringLiteral(rhs) && !namesTelepathy(rhs, names) {
				others[id.Name] = true
			}
		}
		return true
	})
	return others
}

func holdsStringLiteral(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			found = true
		}
		return !found
	})
	return found
}

// mayRunTelepathy reports whether prog, an exec's program argument, can be the built telepathy.
func mayRunTelepathy(prog ast.Expr, names, others map[string]bool) bool {
	if namesTelepathy(prog, names) {
		return true
	}
	switch v := prog.(type) {
	case *ast.BasicLit:
		return false // a literal naming another program
	case *ast.Ident:
		return !others[v.Name]
	case *ast.IndexExpr:
		// os.Args[0]: the test binary itself, re-run as a helper.
		if sel, ok := v.X.(*ast.SelectorExpr); ok {
			if pkg, ok := sel.X.(*ast.Ident); ok && pkg.Name == "os" && sel.Sel.Name == "Args" {
				return false
			}
		}
	}
	return true
}

// fileMentions reports whether f holds the string literal s.
func fileMentions(f *ast.File, s string) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if v, err := strconv.Unquote(lit.Value); err == nil && v == s {
				found = true
			}
		}
		return !found
	})
	return found
}

// binaryNames is every identifier in f that is assigned an expression naming the telepathy binary:
// one holding the literal "telepathy", or built from an identifier that already names it.
func binaryNames(f *ast.File) map[string]bool {
	names := map[string]bool{}
	for changed := true; changed; {
		changed = false
		ast.Inspect(f, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range as.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || names[id.Name] || i >= len(as.Rhs) && len(as.Rhs) != 1 {
					continue
				}
				rhs := as.Rhs[0]
				if i < len(as.Rhs) {
					rhs = as.Rhs[i]
				}
				if namesTelepathy(rhs, names) {
					names[id.Name] = true
					changed = true
				}
			}
			return true
		})
	}
	return names
}

func namesTelepathy(e ast.Expr, names map[string]bool) bool {
	if call, ok := e.(*ast.CallExpr); ok && isExecCommand(call) {
		return false // a command is not a path to the binary, whatever it runs
	}
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			if s, err := strconv.Unquote(v.Value); err == nil && v.Kind == token.STRING && (s == "telepathy" || s == "telepathy.exe") {
				found = true
			}
		case *ast.Ident:
			if names[v.Name] {
				found = true
			}
		}
		return !found
	})
	return found
}

func isExecCommand(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "exec" && (sel.Sel.Name == "Command" || sel.Sel.Name == "CommandContext")
}

// execsWithoutRuntimeDir counts the execs in body that may run the built telepathy and describes
// each one whose command is not assigned an Env holding an XDG_RUNTIME_DIR entry in the same
// function.
func execsWithoutRuntimeDir(body *ast.BlockStmt, names, others map[string]bool) (int, []string) {
	envSet := map[string]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range as.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Env" || i >= len(as.Rhs) {
				continue
			}
			x, ok := sel.X.(*ast.Ident)
			if ok && holdsRuntimeDirEntry(as.Rhs[i]) {
				envSet[x.Name] = true
			}
		}
		return true
	})
	count := 0
	var bad []string
	assigned := map[*ast.CallExpr]string{}
	ast.Inspect(body, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok {
			for i, rhs := range as.Rhs {
				if call, ok := rhs.(*ast.CallExpr); ok && i < len(as.Lhs) {
					if id, ok := as.Lhs[i].(*ast.Ident); ok {
						assigned[call] = id.Name
					}
				}
			}
		}
		call, ok := n.(*ast.CallExpr)
		if !ok || !isExecCommand(call) {
			return true
		}
		prog := 0
		if sel := call.Fun.(*ast.SelectorExpr); sel.Sel.Name == "CommandContext" {
			prog = 1
		}
		if len(call.Args) <= prog {
			return true
		}
		if !mayRunTelepathy(call.Args[prog], names, others) {
			return true
		}
		count++
		v, ok := assigned[call]
		switch {
		case !ok:
			bad = append(bad, "through a command that is never assigned, so it has no Env")
		case !envSet[v]:
			bad = append(bad, "as "+v+", whose Env sets no XDG_RUNTIME_DIR")
		}
		return true
	})
	return count, bad
}

func holdsRuntimeDirEntry(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil && strings.HasPrefix(s, "XDG_RUNTIME_DIR=") {
				found = true
			}
		}
		return !found
	})
	return found
}

// THE SWEEP SEES EVERY SHAPE OF EXEC, and passes only the ones that cannot run telepathy or that
// give it a runtime directory of its own.
func TestTheSweepRecognisesEveryShape(t *testing.T) {
	for _, tc := range []struct {
		name  string
		src   string
		sites int
		bad   int
	}{
		{"a named binary with an env", `bin := filepath.Join(dir, "telepathy"); c := exec.Command(bin); c.Env = []string{"XDG_RUNTIME_DIR=" + rt}`, 1, 0},
		{"a named binary without one", `bin := filepath.Join(dir, "telepathy"); c := exec.Command(bin); _ = c`, 1, 1},
		{"a bare literal", `c := exec.Command("telepathy", "agents"); _ = c`, 1, 1},
		{"an inline path", `c := exec.Command(filepath.Join(dir, "telepathy"), "agents"); _ = c`, 1, 1},
		{"a parameter", `c := exec.CommandContext(ctx, prog, "agents"); _ = c`, 1, 1},
		{"a parameter with an env", `c := exec.Command(prog); c.Env = append(os.Environ(), "XDG_RUNTIME_DIR="+rt)`, 1, 0},
		{"a never-assigned command", `_ = exec.Command(prog).Run()`, 1, 1},
		{"another program by literal", `c := exec.Command("go", "build"); _ = c`, 0, 0},
		{"another program found on PATH", `rg, _ := exec.LookPath("rg"); c := exec.Command(rg); _ = c`, 0, 0},
		{"the test binary", `c := exec.Command(os.Args[0], "-test.run=X"); _ = c`, 0, 0},
	} {
		f, err := parser.ParseFile(token.NewFileSet(), "x_test.go", "package x\nfunc run(prog string) {\n"+tc.src+"\n}\n", 0)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		sites, bad := sweepFile(f)
		if sites != tc.sites || len(bad) != tc.bad {
			t.Errorf("%s: %d sites, %d flagged %v; want %d and %d", tc.name, sites, len(bad), bad, tc.sites, tc.bad)
		}
	}
}
