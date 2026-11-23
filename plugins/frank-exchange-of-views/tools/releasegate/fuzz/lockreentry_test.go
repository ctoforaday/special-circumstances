package fuzz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestNoRunnerMethodThatTakesItsLockIsCalledUnderIt fails a call, made while the runner's mu is held, to
// a runner method that takes mu itself — directly or through another runner method. sync.Mutex is not
// reentrant, so such a call blocks its drive forever: the sweep then runs to the test binary's timeout,
// every part panics, and the tally has nothing to judge. Plain `go test` never reaches runOne's end, so
// this reads the package's source instead of waiting for a sweep to hang.
//
// Held is tracked through each function body in order: `<x>.mu.Lock()` holds it, `<x>.mu.Unlock()`
// releases it, `defer <x>.mu.Unlock()` holds it to the end; a nested block inherits the state and hands
// back the state it ends in. A function literal starts unheld (it runs later, or on another goroutine).
// <x> is a runner method's receiver, or `r` in a function that is no runner method.
func TestNoRunnerMethodThatTakesItsLockIsCalledUnderIt(t *testing.T) {
	files, err := filepath.Glob("*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	type fn struct {
		recv string // the runner receiver's name; "r" for a function that is no runner method
		body *ast.BlockStmt
		name string
		pos  token.Pos
	}
	var funcs []fn
	methods := map[string]*ast.FuncDecl{}
	for _, f := range files {
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range file.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			recv := "r"
			if r := runnerRecv(fd); r != "" {
				recv = r
				methods[fd.Name.Name] = fd
			} else if fd.Recv != nil {
				continue
			}
			funcs = append(funcs, fn{recv: recv, body: fd.Body, name: fd.Name.Name, pos: fd.Pos()})
		}
	}
	if len(methods) == 0 || len(funcs) == 0 {
		t.Fatalf("read %d runner methods and %d functions — the check found nothing to check", len(methods), len(funcs))
	}
	// locking: each runner method that takes mu, closed over the runner methods it calls on its receiver.
	locking := map[string]bool{}
	for name, fd := range methods {
		recv := runnerRecv(fd)
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if _, lit := n.(*ast.FuncLit); lit {
				return false
			}
			if muCall(n, recv) == "Lock" {
				locking[name] = true
			}
			return true
		})
	}
	for changed := true; changed; {
		changed = false
		for name, fd := range methods {
			if locking[name] {
				continue
			}
			recv := runnerRecv(fd)
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if _, lit := n.(*ast.FuncLit); lit {
					return false
				}
				if m := recvCall(n, recv); m != "" && locking[m] {
					locking[name], changed = true, true
				}
				return true
			})
		}
	}
	for _, must := range []string{"noteActErr", "docketTally", "lockedAxes", "countGate"} {
		if !locking[must] {
			t.Fatalf("%s is not read as taking the runner's lock — the check's reading of the source is wrong, so its silence would mean nothing", must)
		}
	}
	var found []string
	for _, f := range funcs {
		var walk func(stmts []ast.Stmt, held bool) bool
		var visit func(n ast.Node, held bool)
		visit = func(n ast.Node, held bool) {
			ast.Inspect(n, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.FuncLit:
					walk(x.Body.List, false)
					return false
				case *ast.BlockStmt:
					walk(x.List, held)
					return false
				}
				if m := recvCall(n, f.recv); m != "" && held && locking[m] {
					found = append(found, fset.Position(n.Pos()).String()+": "+f.name+" calls "+m+" with "+f.recv+".mu held")
				}
				return true
			})
		}
		walk = func(stmts []ast.Stmt, held bool) bool {
			for _, s := range stmts {
				if es, ok := s.(*ast.ExprStmt); ok {
					switch muCall(es.X, f.recv) {
					case "Lock":
						held = true
						continue
					case "Unlock":
						held = false
						continue
					}
				}
				if ds, ok := s.(*ast.DeferStmt); ok && muCall(ds.Call, f.recv) == "Unlock" {
					continue
				}
				if bs, ok := s.(*ast.BlockStmt); ok {
					held = walk(bs.List, held)
					continue
				}
				if is, ok := s.(*ast.IfStmt); ok {
					if is.Init != nil {
						visit(is.Init, held)
					}
					visit(is.Cond, held)
					after := walk(is.Body.List, held)
					if is.Else != nil {
						visit(is.Else, held)
					}
					held = after
					continue
				}
				visit(s, held)
			}
			return held
		}
		walk(f.body.List, false)
	}
	sort.Strings(found)
	if len(found) > 0 {
		t.Fatalf("a runner method that takes the runner's lock is called while it is held — the drive blocks forever:\n  %s", strings.Join(found, "\n  "))
	}
}

// runnerRecv is fd's receiver name when fd is a method on *runner, else "".
func runnerRecv(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) != 1 || len(fd.Recv.List[0].Names) != 1 {
		return ""
	}
	star, ok := fd.Recv.List[0].Type.(*ast.StarExpr)
	if !ok {
		return ""
	}
	if id, ok := star.X.(*ast.Ident); !ok || id.Name != "runner" {
		return ""
	}
	return fd.Recv.List[0].Names[0].Name
}

// muCall is "Lock" or "Unlock" when n is `<recv>.mu.Lock()` or `<recv>.mu.Unlock()`, else "".
func muCall(n ast.Node, recv string) string {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || (sel.Sel.Name != "Lock" && sel.Sel.Name != "Unlock") {
		return ""
	}
	mu, ok := sel.X.(*ast.SelectorExpr)
	if !ok || mu.Sel.Name != "mu" {
		return ""
	}
	if id, ok := mu.X.(*ast.Ident); ok && id.Name == recv {
		return sel.Sel.Name
	}
	return ""
}

// recvCall is the method name when n is a call `<recv>.<method>(…)`, else "".
func recvCall(n ast.Node, recv string) string {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if id, ok := sel.X.(*ast.Ident); ok && id.Name == recv {
		return sel.Sel.Name
	}
	return ""
}
