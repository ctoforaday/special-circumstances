package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// blockedPlan is the plan the chair's `dispatch next` stores over a board whose open material gap G1
// holds the PASS: nothing permits it, and the lens that minted G1 clears it.
func (b *stage) blockedPlan() *stage {
	p := emptyPlan()
	p.Head = b.head()
	p.Blockers = []PlanBlocker{{Kind: BlockerMaterialGap, Subject: "G1", Owner: evLens}}
	return b.add(chairSeat, &recordpb.Plan{Printed: proto.String(p.WithKey().Printed())})
}

// verdictItems is every verdict item on a work list, blocking or not.
func verdictItems(s SittingJSON) []Item {
	var out []Item
	for _, it := range s.Open {
		if it.Kind == ItemVerdictOwed {
			out = append(out, it)
		}
	}
	return out
}

// THE VERDICT ITEM IS LISTED ONLY WHERE THE PASS IS THE CHAIR'S TO RECORD IN THIS SITTING. The smoke
// that found it (ev902): a chair sitting whose plan permitted no PASS — a material gap open, its
// minter's to close — was told it owed its terminal act, because no verdict had been recorded in
// the run. A sitting whose plan permits the PASS owes the PASS or a position, and either clears it.
func TestTheVerdictItemIsListedOnlyWhereThePlanPermitsThePass(t *testing.T) {
	blocked := func(t *testing.T) *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).mint(evLens, "G1", "medium").
			register("red-chair")
	}
	for _, c := range []struct {
		name  string
		board func(t *testing.T) *stage
		owed  bool
	}{
		{"a plan that permits no PASS, nothing recorded in the run", func(t *testing.T) *stage { return blocked(t).blockedPlan() }, false},
		{"a plan that permits no PASS, a position filed", func(t *testing.T) *stage { return blocked(t).blockedPlan().position(chairSeat) }, false},
		{"a plan that permits the PASS, nothing recorded", func(t *testing.T) *stage { return permittingBoard(t).permittedPlan() }, true},
		{"a plan that permits the PASS, the PASS recorded", func(t *testing.T) *stage { return permittingBoard(t).permittedPlan().add(chairSeat, &recordpb.Gate{}) }, false},
		{"a plan that permits the PASS, a position withholding it", func(t *testing.T) *stage { return permittingBoard(t).withhold(heardID) }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := sittingOfRunT(t, c.board(t).seed(), "chair", "red-chair")
			items := verdictItems(s)
			switch {
			case c.owed && len(items) != 1:
				t.Fatalf("the chair's list holds %d verdict item(s), want the one PASS-or-position item:\n%+v", len(items), s.Open)
			case c.owed && items[0].What != "your plan permits the PASS: record it, or withhold it with a position":
				t.Errorf("the verdict item reads %q", items[0].What)
			case !c.owed && len(items) != 0:
				t.Errorf("the chair's list holds a verdict item where none is owed:\n%+v", items)
			}
		})
	}
}
