package fuzz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

// answerCallees are the calls that return the tool's answer as (output, error). Their callers read
// both, or go through admit / refuse / either (acts_test.go), which read them.
var answerCallees = map[string]bool{"exec": true, "execAs": true, "result": true, "drive": true}

// setupWrites are the calls that build a run's files; their error fails the test (writeSetup).
var setupWrites = map[string]bool{"WriteFile": true, "MkdirAll": true}

// TestNoActDiscardsTheToolsAnswer fails on a discarded answer anywhere in this package: a call
// to exec, result or drive whose error is a blank or whose results are dropped as a statement,
// `_ = err`, and a setup write whose error is a blank. It reads the syntax tree, so a site is found
// by what it calls, however it is spelled or wrapped across lines.
func TestNoActDiscardsTheToolsAnswer(t *testing.T) {
	files, err := filepath.Glob("*_test.go")
	if err != nil || len(files) == 0 {
		t.Fatalf("list the package's files: %v (found %d)", err, len(files))
	}
	fset := token.NewFileSet()
	var bad []string
	for _, f := range files {
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", f, err)
		}
		bad = append(bad, discards(fset, file)...)
	}
	if len(bad) > 0 {
		t.Errorf("%d site(s) discard the tool's answer. Issue the act through admit, refuse or either "+
			"(acts_test.go), or read both results; a setup write fails the test on its error:\n  %s",
			len(bad), strings.Join(bad, "\n  "))
	}
}

// discards is every discarding site in one file.
func discards(fset *token.FileSet, file *ast.File) []string {
	var bad []string
	at := func(n ast.Node, what string) {
		bad = append(bad, fset.Position(n.Pos()).String()+": "+what)
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.ExprStmt:
			if name := calleeName(s.X); answerCallees[name] {
				at(s, name+"(...) as a statement drops its error")
			}
		case *ast.AssignStmt:
			if len(s.Rhs) != 1 || len(s.Lhs) == 0 {
				return true
			}
			last := s.Lhs[len(s.Lhs)-1]
			if !isBlank(last) {
				return true
			}
			name := calleeName(s.Rhs[0])
			switch {
			case answerCallees[name]:
				at(s, name+"(...) assigns its error to _")
			case setupWrites[name] && isPkg(s.Rhs[0], "os"):
				at(s, "os."+name+"(...) assigns its error to _")
			case len(s.Lhs) == 1:
				if id, ok := s.Rhs[0].(*ast.Ident); ok && (id.Name == "err" || strings.HasSuffix(id.Name, "Err") || strings.HasSuffix(id.Name, "err")) {
					at(s, "_ = "+id.Name+" discards an error")
				}
			}
		}
		return true
	})
	return bad
}

// calleeName is the function or method name a call expression calls, "" for anything else.
func calleeName(e ast.Expr) string {
	call, ok := e.(*ast.CallExpr)
	if !ok {
		return ""
	}
	switch f := call.Fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// isPkg reports whether the call's receiver is the package identifier pkg.
func isPkg(e ast.Expr, pkg string) bool {
	sel, ok := e.(*ast.CallExpr).Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == pkg
}

func isBlank(e ast.Expr) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == "_"
}

// TestTheDiscardCheckFindsEachShape holds the check to every shape it names, so a check that has
// stopped seeing one fails here rather than passing the package.
func TestTheDiscardCheckFindsEachShape(t *testing.T) {
	src := `package p
func f() {
	_, _ = r.exec("a")
	r.exec("a")
	out, _ := r.exec("a")
	_, _ = c.result()
	_, _ = drive(bin, "a")
	_ = err
	_ = os.WriteFile("p", nil, 0)
	_ = os.MkdirAll("p", 0)
	_ = out
	r.admit("a")
	_ = os.RemoveAll("p")
	o, err := r.exec("a")
	_, _ = o, err
}`
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	got := discards(fset, file)
	if len(got) != 8 {
		t.Errorf("the check found %d of the 8 discarding shapes:\n  %s", len(got), strings.Join(got, "\n  "))
	}
}
