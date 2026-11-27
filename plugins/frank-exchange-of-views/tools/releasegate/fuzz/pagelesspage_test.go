package fuzz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE PAGELESS PAGE IS DRIVEN ONCE A RUN AND REFUSED: the citation owner's first sitting with a live
// citation passes `verify --page` on it, the tool refuses it as a validation and records nothing, and
// a later sitting drives it no more. Another lens, and the owner before blue cites, drive nothing.
func TestThePagelessPageIsRefusedOnceARun(t *testing.T) {
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1") + "\n" + lanePartText("blue-lane-2") + "\n" + lanePartText("blue-lane-3")); err != nil {
		t.Fatalf("record the lanes' parts: %v", err)
	}
	const blue = "blue-respond"
	r.sit("lens", citationOwner)
	r.pagelessPageRefused(citationOwner)
	if r.pagelessPageDriven || r.actErr != "" {
		t.Fatalf("with no citation on the record the drive reads driven %v (%s)", r.pagelessPageDriven, r.actErr)
	}
	r.sit("blue", blue)
	r.citeAs(t, blue, "The cost is rising over time", "/pageless")
	r.sit("lens", "red-lens-logic")
	r.pagelessPageRefused("red-lens-logic")
	if r.pagelessPageDriven {
		t.Fatal("a lens that is not the citation owner drove the pageless page")
	}
	r.sit("lens", citationOwner)
	r.pagelessPageRefused(citationOwner)
	if !r.pagelessPageDriven || r.actErr != "" {
		t.Fatalf("the owner's sitting over a live citation: driven %v (%s)", r.pagelessPageDriven, r.actErr)
	}
	b, ok := r.board()
	if !ok {
		t.Fatal("the board did not read")
	}
	for _, e := range b.Events {
		if _, isV := recordpb.BodyAs[*recordpb.Verify](e); isV {
			t.Fatalf("the refused --page recorded a reading by %s", e.GetSeatId())
		}
	}
	before := lensVerifies()
	r.pagelessPageRefused(citationOwner)
	if n := lensVerifies() - before; n != 0 {
		t.Fatalf("a run that has driven the pageless page drove `lens verify` %d more time(s)", n)
	}
}

// THE SITTING CALLS THE ARM ON NO DRAW. The guarantee that every part passes `lens verify --page` is
// that the arm runs at every lens sitting whatever the run's draws: so envelopeFor calls it, and no
// `if` around the call, nor a closure it sits in, tests a coin, a maybe, the run's rng or drawsReading.
func TestThePagelessPageArmIsCalledOnNoDraw(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "fuzz_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	drawn := map[string]bool{"coin": true, "maybe": true, "rng": true, "drawsReading": true}
	mentionsDraw := func(n ast.Node) string {
		found := ""
		ast.Inspect(n, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if drawn[x.Name] && found == "" {
					found = x.Name
				}
			case *ast.FuncLit:
				return false // a closure in a condition is not the condition
			}
			return true
		})
		return found
	}
	calls := 0
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "envelopeFor" {
			continue
		}
		var stack []ast.Node
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "pagelessPageRefused" {
					calls++
					for i, a := range stack {
						switch x := a.(type) {
						case *ast.IfStmt:
							if d := mentionsDraw(x.Cond); d != "" {
								t.Errorf("the call to pagelessPageRefused sits under an if testing %s", d)
							}
						case *ast.FuncLit:
							if i > 0 {
								if outer, ok := stack[i-1].(*ast.CallExpr); ok {
									if d := mentionsDraw(outer.Fun); d != "" {
										t.Errorf("the call to pagelessPageRefused sits in a closure passed to %s", d)
									}
								}
							}
						}
					}
				}
			}
			stack = append(stack, n)
			return true
		})
	}
	if calls == 0 {
		t.Fatal("envelopeFor never calls pagelessPageRefused, so no sitting drives `lens verify --page` on a draw-free path")
	}
}

// lensVerifies is the process's `lens verify` invocations so far.
func lensVerifies() int {
	execMu.Lock()
	defer execMu.Unlock()
	return execRuns["lens verify"]
}
