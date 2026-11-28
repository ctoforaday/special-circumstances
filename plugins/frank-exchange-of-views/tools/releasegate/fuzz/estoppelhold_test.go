package fuzz

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE ESTOPPEL RUN DRIVES ITS ESTOPPEL WHATEVER ORDER ITS SITTINGS TAKE. Two lenses prescribe the same
// swap of the cost sentence. Blue answers both in one sitting and makes a verbatim apply its last change
// to the sentence (verbatimLast), and so again at a sitting beside the prescribing lens, and that lens's
// probe is refused by the tool's estoppel guard. Without the order, blue's counter-edit of the second
// gap (its span taken by the first apply) lists the first reopened and ends its estoppel before its lens
// sits, and the sweep drives no estoppel (CI run 46094143286, seed 2).
func TestTheEstoppelRunsApplyStandsUntilItsLensProbes(t *testing.T) {
	r := stagedRunner(t)
	r.forceEstoppel = true
	if err := r.stagePart("A § fuzz sentence to anchor findings.\n\nThe cost is rising over time across the whole sampled corpus.\n"); err != nil {
		t.Fatalf("record the lane's part: %v", err)
	}
	r.chairPlans(t)
	var gaps []string
	for _, lens := range fuzzLensSeats[:2] {
		r.dispatchSits("lens", lens)
		g := r.mint(lens)
		if g == "" {
			t.Fatalf("%s's mint landed nothing: %s", lens, r.actErr)
		}
		r.dispatchReturns(lens, false)
		gaps = append(gaps, g)
	}
	sitBlue := func() {
		r.chairPlans(t)
		r.dispatchSits("blue", "blue-respond")
		r.blueRespondTo("blue-respond", r.openGaps())
		r.dispatchReturns("blue-respond", false)
	}
	sitBlue()
	if answered := r.answeredGaps(); len(answered) != len(gaps) {
		t.Fatalf("blue's sitting answered %v of %v, want every gap: %s", answered, gaps, r.actErr)
	}
	var g string
	for _, id := range r.appliedVerbatimGaps() {
		if r.stillEstops(id) {
			g = id
		}
	}
	if g == "" {
		t.Fatalf("after blue's sitting no verbatim apply of %v still estops (applied %v): %s", gaps, r.appliedVerbatimGaps(), r.actErr)
	}
	// BLUE SITS AGAIN BEFORE THE PROBE: its last change is a verbatim apply again.
	sitBlue()
	g = ""
	for _, id := range r.appliedVerbatimGaps() {
		if r.stillEstops(id) {
			g = id
		}
	}
	if g == "" {
		t.Fatalf("blue's sitting before the probe left no verbatim apply that estops")
	}
	lens := r.minterOf(g)
	r.dispatchSits("lens", lens)
	r.mintEstopped(lens)
	r.dispatchReturns(lens, false)
	if n := r.toolEstoppels(); n == 0 {
		t.Fatalf("%s's probe over its applied prescription on %s wrote no estoppel refusal (misses %v): %s", lens, g, r.estoppelMisses, r.actErr)
	}
}

// answeredGaps is each gap a blue edit on the record answers.
func (r *runner) answeredGaps() []string {
	b, ok := r.board()
	if !ok {
		return nil
	}
	var out []string
	for _, e := range b.Events {
		if ed, isE := recordpb.BodyAs[*recordpb.BlueEdit](e); isE && ed.GetAnswers() != "" && !slices.Contains(out, ed.GetAnswers()) {
			out = append(out, ed.GetAnswers())
		}
	}
	return out
}

// toolEstoppels is the estoppel refusals the tool logged.
func (r *runner) toolEstoppels() int {
	b, ok := r.board()
	if !ok {
		return 0
	}
	n := 0
	for _, e := range b.Events {
		if l, isL := recordpb.BodyAs[*recordpb.Log](e); isL && l.GetType() == recordpb.LogType_LOG_TYPE_ESTOPPEL && l.GetSource() == recordpb.LogSource_LOG_SOURCE_TOOL {
			n++
		}
	}
	return n
}
