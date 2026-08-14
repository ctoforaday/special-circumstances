package verify

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// find returns the named check from a Run result.
func find(t *testing.T, checks []Check, name string) Check {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no check named %q", name)
	return Check{}
}

func p(kv ...string) *record.Payload {
	pl := record.NewPayload()
	for i := 0; i+1 < len(kv); i += 2 {
		pl.Set(kv[i], kv[i+1])
	}
	return pl
}

// A gap closed by a merge carries closure_class; one closed by a bench opinion carries a
// disposition. BOTH are recorded reasons — the check must accept either. This is the
// regression that shipped in the first cut (it looked only for closure_class and flagged all
// eight opinion-closed gaps in the 2026-07-22 run).
func TestGapsDisposedAcceptsBothClosureFields(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"C1", "C2", "OPEN", "TORN"},
		Gaps: map[string]*record.Gap{
			"C1":   {ID: "C1", Open: false, Closure: p("closure_class", "closed")},           // merge close
			"C2":   {ID: "C2", Open: false, Closure: p("disposition", "rebuttal_sustained")}, // bench opinion
			"OPEN": {ID: "OPEN", Open: true},                                                 // ignored
			"TORN": {ID: "TORN", Open: false, Closure: p("something", "else")},               // no reason
		},
	}
	c := find(t, Run(b), "gaps-disposed")
	if c.OK {
		t.Fatalf("expected a torn closure to fail the check:\n%+v", c)
	}
	if len(c.Violations) != 1 || c.Violations[0] != "TORN" {
		t.Errorf("only TORN should be flagged (C1/C2 are validly closed, OPEN is open): %v", c.Violations)
	}
}

func TestFoundByResolves(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"G1", "G2"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", Mint: p().Set("found_by", []string{"L1-F1"})},
			"G2": {ID: "G2", Mint: p().Set("found_by", []string{"GHOST"})},
		},
		Events: []record.Event{{Type: "finding", Payload: p("label", "L1-F1")}},
	}
	c := find(t, Run(b), "found-by-resolves")
	if c.OK || len(c.Violations) != 1 || c.Violations[0] != "G2→GHOST" {
		t.Errorf("a found_by naming a finding nobody recorded must be flagged: %+v", c)
	}
}

func TestDialecticRefsResolve(t *testing.T) {
	b := &record.Board{
		Gaps: map[string]*record.Gap{"R1-1": {ID: "R1-1"}},
		Events: []record.Event{
			{Type: "opinion", SeatID: "judge-r1", Payload: p("gap_id", "R1-1", "disposition", "closed")},
			{Type: "closing", SeatID: "blue-r1", Payload: p("gap_id", "PHANTOM")},
		},
	}
	c := find(t, Run(b), "dialectic-refs-resolve")
	if c.OK || len(c.Violations) != 1 || c.Violations[0] != "blue-r1/closing→PHANTOM" {
		t.Errorf("a closing about a nonexistent gap must be flagged: %+v", c)
	}
}

func TestPassClosesAllGaps(t *testing.T) {
	openUnderPass := &record.Board{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", Open: true}},
		Events:   []record.Event{{Type: "outcome", Payload: p("verdict", "PASS")}},
	}
	if c := find(t, Run(openUnderPass), "pass-closes-all-gaps"); c.OK {
		t.Error("PASS with an open gap must fail the #67 gate")
	}
	// A non-PASS verdict makes the gate inapplicable — an open gap is fine.
	ceiling := &record.Board{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", Open: true}},
		Events:   []record.Event{{Type: "outcome", Payload: p("verdict", "CEILING")}},
	}
	if c := find(t, Run(ceiling), "pass-closes-all-gaps"); !c.OK {
		t.Error("a CEILING verdict must not trip the PASS gate")
	}
}

func TestRegisterBeforeAppend(t *testing.T) {
	b := &record.Board{
		Events: []record.Event{
			{Type: "register", SeatID: "blue-r1"},
			{Type: "position", SeatID: "blue-r1"},
			{Type: "finding", SeatID: "red-lens-r1-L5"}, // never registered
		},
	}
	c := find(t, Run(b), "register-before-append")
	if c.OK || len(c.Violations) != 1 {
		t.Errorf("a seat whose first event is not register must be flagged: %+v", c)
	}
}

func TestComputeStatsReproducesCoverage(t *testing.T) {
	b := &record.Board{
		GapOrder: []string{"G1", "G2"},
		Gaps: map[string]*record.Gap{
			"G1": {ID: "G1", Open: true, Mint: p().Set("found_by", []string{"L5-F1"})},
			"G2": {ID: "G2", Open: false, Closure: p("disposition", "closed")},
		},
		Events: []record.Event{
			{Type: "finding", Payload: p("label", "L5-F1")}, // minted
			{Type: "finding", Payload: p("label", "L5-F2")}, // un-minted
			{Type: "opinion", Payload: p("gap_id", "G2", "disposition", "closed")},
			{Type: "verify", Payload: p("claim", "c1", "reference", "r1")},
			{Type: "verify", Payload: p("claim", "c2", "reference", "r2")},
			{Type: "outcome", Payload: p("verdict", "CEILING")},
		},
	}
	s := Compute(b)
	if s.GapsTotal != 2 || s.GapsOpen != 1 || s.GapsClosed != 1 {
		t.Errorf("gap counts wrong: %+v", s)
	}
	if s.Findings != 2 || s.FindingsMinted != 1 || s.FindingsUnminted != 1 {
		t.Errorf("finding coverage wrong: %+v", s)
	}
	// citations_checked's canonical source: one per cite event the replay carries.
	// Cite events are reference-keyed, so a re-verification is deduped upstream in the
	// board replay before Compute sees it — this is a pure count of what survives.
	if s.Citations != 2 {
		t.Errorf("citations tally wrong: got %d, want 2: %+v", s.Citations, s)
	}
	if s.GapsWithOpinion != 1 || s.GapsWithClosing != 0 {
		t.Errorf("dialectic coverage wrong: %+v", s)
	}
	if s.Verdict != "CEILING" {
		t.Errorf("verdict not read: %q", s.Verdict)
	}
}
