package fuzz

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A GAP HOLDS ONE PROPOSAL, AND THE BENCH RULES IT — pinned deterministically. A second proposal while
// the first stands is refused as a conflict and records nothing, so the bench convened on the gap rules
// the one motion the record holds, and its sitting ends with none standing.
func TestAGapHoldsOneProposalAndTheBenchRulesIt(t *testing.T) {
	bin := buildBinary(t)
	runDir := recordtest.TmpRun(t)
	// THE CAST, as setup writes it: the chair's dispatch below is checked against it.
	stageRun, err := record.NewRun(runDir)
	if err != nil {
		t.Fatal(err)
	}
	cast, lanes := record.CastFor(1)
	if _, err := record.Append(record.Identity{Run: stageRun, SeatID: record.HarnessSeat}, &recordpb.Cast{SeatIds: cast, LaneSeatIds: lanes}); err != nil {
		t.Fatal(err)
	}
	r := newRunner(bin, runDir, newLockedRand(7))
	const sentence = "A § fuzz sentence to anchor findings."
	if err := r.stagePart(sentence + "\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.chairPlans(t)
	lens := fuzzLensSeats[0]
	// A DOCUMENT CHECK, so blue's sitting on the gap owes no live answer and ends admitted.
	r.dispatchSits("lens", lens)
	r.mustExec(t, "--json", "--seat-id", lens, "show", "report")
	r.mustExec(t, "--json", "--seat-id", lens, "mint", "--quote", sentence,
		"--class", "false_universal", "--check-kind", "document", "--check", "acc", "--problem", "fuzz problem",
		"--fix", dirIgnore, "--severity", "high", "--impact", "high", "--likelihood", "high", "--reason", "fuzz reason")
	r.dispatchReturns(lens, false)
	open := r.openGaps()
	if len(open) != 1 {
		t.Fatalf("the board holds %v open after one mint, want one gap", open)
	}
	gapID := open[0]
	r.exchangeOn(t)
	if !r.exchanged(gapID) {
		t.Fatalf("%s has had no counted exchange after blue's sitting on it", gapID)
	}

	r.dispatchSits("blue", "blue-respond")
	if _, ok := r.admit(proposeCloseArgs("blue-respond", gapID, "not_a_defect", "fuzz: "+gapID+" is the bench's to settle")...); !ok {
		t.Fatalf("the first proposal on %s was refused: %s", gapID, r.actErr)
	}
	first := r.unruledDocketMotions(gapID)
	if len(first) != 1 {
		t.Fatalf("the first proposal stands %d unruled docket motion(s) on %s, want one: %v", len(first), gapID, first)
	}
	r.refuse(feov.Conflict, proposeCloseArgs("blue-respond", gapID, "defect_accepted", "fuzz: a second proposal on "+gapID)...)
	if r.actErr != "" {
		t.Fatalf("the second proposal while %s stands: %s", first[0], r.actErr)
	}
	if got := r.unruledDocketMotions(gapID); len(got) != 1 {
		t.Fatalf("the refused proposal left %d unruled docket motion(s) on %s, want the first alone: %v", len(got), gapID, got)
	}
	r.dispatchReturns("blue-respond", false)

	// THE CHAIR CONVENES THE BENCH ON THE GAP once its parties have filed their closings on it, so the
	// record holds the dispatch row the bench's docket sitting answers.
	r.partiesClose(t, lens)
	r.dispatchSits("bench", "judge", "docket")
	r.benchDisposes("judge", gapID, "remanded", "--direction", "blue pursuing the stated direction",
		"--reason", "fuzz: carried with a stated direction for "+gapID)
	if r.actErr != "" {
		t.Fatalf("the bench's ruling: %s", r.actErr)
	}
	if left := r.standingOn(gapID); len(left) > 0 {
		t.Errorf("the bench sat for %s and left %v standing", gapID, left)
	}
	r.dispatchReturns("judge", r.envelopeEnds("judge"))
	requireEveryDispatchReturned(t, r)
	if r.actErr != "" {
		t.Errorf("the bench's envelope after its ruling: %s", r.actErr)
	}
}
