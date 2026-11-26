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
// HOOKCORE_LIVE and build their own environment through testhome) and fails any exec of the built
// telepathy whose command's Env has no XDG_RUNTIME_DIR entry. A sweep that finds no exec at all
// fails too: a pattern that stopped matching would otherwise read as a clean module.
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
		names := binaryNames(f)
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			n, bad := execsWithoutRuntimeDir(fn.Body, names)
			sites += n
			for _, b := range bad {
				t.Errorf("%s: %s execs the built telepathy %s", filepath.ToSlash(rel), fn.Name.Name, b)
			}
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

// execsWithoutRuntimeDir counts the execs of the built telepathy in body and describes each one
// whose command is not assigned an Env holding an XDG_RUNTIME_DIR entry in the same function.
func execsWithoutRuntimeDir(body *ast.BlockStmt, names map[string]bool) (int, []string) {
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
		id, ok := call.Args[prog].(*ast.Ident)
		if !ok || !names[id.Name] {
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
