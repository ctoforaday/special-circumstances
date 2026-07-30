package dashboard

import "testing"

func fmin(min float64) *float64 { v := min * 60000; return &v }

// Ports run-dashboard.test.mjs's four projectCompletion cases 1:1.
func TestProjectCompletionRangesOverSpans(t *testing.T) {
	seats := []Seat{
		{Seat: "red-merge", Label: "red-merge-r1", Done: true, StartedMs: fmin(0), EndedMs: fmin(11)},
		{Seat: "red-merge", Label: "red-merge-r2", Done: true, StartedMs: fmin(20), EndedMs: fmin(37)},
		{Seat: "blue-synthesize", Label: "blue-synthesize", Done: true, StartedMs: fmin(40), EndedMs: fmin(58)},
		{Seat: "red-merge", Label: "red-merge-r3", Done: false, StartedMs: fmin(95)},
	}
	p := projectCompletion(seats, *fmin(100))
	if p.LowMin != 24 {
		t.Errorf("lowMin = %d, want 24", p.LowMin)
	}
	if p.HighMin != 30 {
		t.Errorf("highMin = %d, want 30", p.HighMin)
	}
	if !(p.HighMin > p.LowMin) {
		t.Error("want a range, not a point estimate")
	}
	if p.Basis != "3 completed seat(s) in this run" {
		t.Errorf("basis = %q", p.Basis)
	}
}

func TestProjectCompletionPerRoundCost(t *testing.T) {
	seats := []Seat{
		{Seat: "red-lens", Label: "red-lens-r1", Done: true, StartedMs: fmin(0), EndedMs: fmin(4)},
		{Seat: "red-merge", Label: "red-merge-r1", Done: true, StartedMs: fmin(4), EndedMs: fmin(15)},
		{Seat: "blue-respond", Label: "blue-respond-r1", Done: true, StartedMs: fmin(15), EndedMs: fmin(20)},
	}
	p := projectCompletion(seats, *fmin(20))
	if p.PerRoundLowMin != 20 || p.PerRoundHighMin != 20 {
		t.Errorf("perRound = %d–%d, want 20–20", p.PerRoundLowMin, p.PerRoundHighMin)
	}
}

func TestProjectCompletionNamesUnmeasured(t *testing.T) {
	p := projectCompletion([]Seat{{Seat: "red-lens", Label: "red-lens-r1", Done: false, StartedMs: fmin(0)}}, *fmin(1))
	has := func(x string) bool {
		for _, u := range p.Unmeasured {
			if u == x {
				return true
			}
		}
		return false
	}
	if !has("red-lens-r1") {
		t.Error("a live seat with no completed sibling must be named unmeasured")
	}
	if !has("assembly") {
		t.Error("assembly with no precedent must be named unmeasured")
	}
}

func TestProjectCompletionCompleteOnAssembly(t *testing.T) {
	p := projectCompletion([]Seat{{Seat: "assemble", Label: "assemble", Done: true, StartedMs: fmin(0), EndedMs: fmin(10)}}, *fmin(11))
	if p.State != "complete" {
		t.Errorf("state = %q, want complete", p.State)
	}
	if p.HighMin != 0 {
		t.Errorf("highMin = %d, want 0", p.HighMin)
	}
}
