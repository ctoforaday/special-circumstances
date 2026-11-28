package fuzz

import "testing"

// THE MINTER ANSWERS A GAP ON ITS GAP IN THE STALE DRIVE (§III.B0's gap row): its amend of G restates
// G's wording and names G2, and its regrade of G names G2; both land, and the raiser's regrade of G2 is
// still refused stale and recorded again. The sweep passes amend's --answers, --problem and
// --required-fix and regrade's --answers here and nowhere else.
func TestTheStaleDriveAnswersTheGapOnTheGap(t *testing.T) {
	r := stagedSynthesizedRunner(t)
	minter, raiser := "red-lens-logic", "red-lens-voice"
	r.register("lens", minter)
	g, why := r.lifecycleMint(minter, "document", "A § fuzz sentence to anchor findings.", "", "medium")
	if g == "" {
		t.Fatal(why)
	}
	r.register("lens", raiser)
	r.staleOnASubject(raiser)
	if r.actErr != "" || !r.staleOnSubject {
		t.Fatalf("the stale drive: stale %v, %s", r.staleOnSubject, r.actErr)
	}
	b, ok := r.board()
	if !ok {
		t.Fatal("the board")
	}
	var amends, regrades int
	for _, e := range b.Events {
		if a := e.GetAmend(); a.GetGapId() == g && a.GetAnswers() != "" && a.GetProblem() != "" && a.GetRequiredFix() != "" {
			amends++
		}
		if rg := e.GetRegrade(); rg.GetGapId() == g && rg.GetAnswers() != "" {
			regrades++
		}
	}
	if amends != 1 || regrades != 1 {
		t.Errorf("the minter's answers on the gap on %s: %d amend(s) and %d regrade(s) naming it, want one each", g, amends, regrades)
	}
}
