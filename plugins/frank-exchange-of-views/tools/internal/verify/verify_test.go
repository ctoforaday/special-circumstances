package verify

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
		Events: []*record.Event{{Type: "finding", Payload: p("label", "L1-F1")}},
	}
	c := find(t, Run(b), "found-by-resolves")
	if c.OK || len(c.Violations) != 1 || c.Violations[0] != "G2→GHOST" {
		t.Errorf("a found_by naming a finding nobody recorded must be flagged: %+v", c)
	}
}

func TestDialecticRefsResolve(t *testing.T) {
	b := &record.Board{
		Gaps: map[string]*record.Gap{"R1-1": {ID: "R1-1"}},
		Events: []*record.Event{
			{Type: "opinion", SeatID: "judge-r1", Payload: p("gap_id", "R1-1", "disposition", "closed")},
			{Type: "closing", SeatID: "blue-r1", Payload: p("gap_id", "PHANTOM")},
		},
	}
	c := find(t, Run(b), "dialectic-refs-resolve")
	if c.OK || len(c.Violations) != 1 || c.Violations[0] != "blue-r1/closing→PHANTOM" {
		t.Errorf("a closing about a nonexistent gap must be flagged: %+v", c)
	}
}

// CORRECTED: this test used to build `{Type: "outcome", verdict: "PASS"}` and assert the gate
// fired. It passed — against an implementation that read `outcome` events — and BOTH were wrong
// about the tree they describe.
//
// An `outcome` event's vocabulary is VERIFIED|CEILING|HALTED|UNVERIFIED, validated at the write.
// It can never carry "PASS"; PASS lives on a `verdict` event. So the board this test constructed
// was one the tool cannot produce, the implementation was reading a field that never held the
// value it compared against, and the two agreed with each other while the gate ran on nothing.
//
// A hand-built board is the right tool for testing an after-the-fact verifier — that is the
// whole point of one — but building it past a validation the real writer enforces turns the test
// into a mirror of the implementation. The fixture has to be a state the WRITER could reach.
func TestPassClosesAllGaps(t *testing.T) {
	openUnderPass := &record.Board{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", Open: true}},
		Events:   []*record.Event{{Type: "verdict", Payload: p("verdict", "PASS")}},
	}
	if c := find(t, Run(openUnderPass), "pass-closes-all-gaps"); c.OK {
		t.Error("PASS with an open gap must fail the #67 gate")
	}
	// A FAIL verdict makes the gate inapplicable — an open gap is expected there.
	failed := &record.Board{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", Open: true}},
		Events:   []*record.Event{{Type: "verdict", Payload: p("verdict", "FAIL")}},
	}
	if c := find(t, Run(failed), "pass-closes-all-gaps"); !c.OK {
		t.Error("a FAIL verdict must not trip the PASS gate")
	}
	// And a terminal CEILING outcome, which is a different event and a different question.
	// Whether outcome=VERIFIED with open gaps should ALSO be a violation is a live design
	// question, not something this fix decided — see the issue at passClosesAllGaps.
	ceiling := &record.Board{
		GapOrder: []string{"G1"},
		Gaps:     map[string]*record.Gap{"G1": {ID: "G1", Open: true}},
		Events:   []*record.Event{{Type: "outcome", Payload: p("verdict", "CEILING")}},
	}
	if c := find(t, Run(ceiling), "pass-closes-all-gaps"); !c.OK {
		t.Error("a CEILING outcome must not trip the PASS gate")
	}
}

func TestRegisterBeforeAppend(t *testing.T) {
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "blue-r1", 0, &recordpb.Register{}),
			recordtest.Event(t, "blue-r1", 0, &recordpb.Position{}),
			recordtest.Event(t, "red-lens-r1-L5", 0, &recordpb.Finding{}), // never registered
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
		Events: []*record.Event{
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

// THE AFTER-THE-FACT GATE, which had never run.
//
// passClosesAllGaps scanned `outcome` events for verdict "PASS". An outcome's vocabulary is
// VERIFIED|CEILING|HALTED|UNVERIFIED, validated at the write, so the comparison could not match
// and the check reported "gate not applicable" on every run ever recorded.
//
// The board is built DIRECTLY here rather than through the CLI, and that is the point: the live
// gate in record.Append refuses `merge verdict --as PASS` while a gap is open, so the
// contradiction cannot be produced the normal way. A record assembled some other way — a
// hand-edited shard, a legacy run, a regressed live gate — is the only thing this check is for,
// and it is what a test of it has to construct.
func TestPassClosesAllGapsFiresOnAPassWithAnOpenGap(t *testing.T) {
	b := &record.Board{
		Events:   []*record.Event{recordtest.Event(t, "", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
	}
	got := passClosesAllGaps(b)
	if got.OK {
		t.Fatalf("a PASS verdict with R1-1 still open must FAIL the gate; got ok with detail %q", got.Detail)
	}
	if len(got.Violations) != 1 || got.Violations[0] != "R1-1" {
		t.Errorf("the violation must name the open gap, got %v", got.Violations)
	}
}

func TestPassClosesAllGapsPassesWhenPassClosedEverything(t *testing.T) {
	b := &record.Board{
		Events:   []*record.Event{recordtest.Event(t, "", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: false}},
	}
	if got := passClosesAllGaps(b); !got.OK {
		t.Errorf("a PASS with every gap closed must pass; got %q %v", got.Detail, got.Violations)
	}
}

// A FAIL verdict is not the gate's business, and neither is a terminal `outcome` — reading the
// latter is what made the check dead. Both must leave it inapplicable rather than firing.
func TestPassClosesAllGapsIsNotApplicableWithoutAPassVerdict(t *testing.T) {
	for _, ev := range []*record.Event{
		recordtest.Event(t, "", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
		recordtest.Event(t, "", 0, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_VERIFIED)}),
	} {
		b := &record.Board{
			Events:   []*record.Event{ev},
			GapOrder: []string{"R1-1"},
			Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
		}
		if got := passClosesAllGaps(b); !got.OK {
			t.Errorf("%s/%s must leave the gate inapplicable, not fire it: %q",
				ev.Type, ev.Payload.Str("verdict"), got.Detail)
		}
	}
}

// THE CLASS, not the instance.
//
// `verdict` and `outcome` events both carry their word under the payload key "verdict", and
// their vocabularies are disjoint — PASS|FAIL against VERIFIED|CEILING|HALTED|UNVERIFIED. So
// reading the wrong event type is a type error the compiler cannot see, and it does not fail
// loudly: the comparison simply never matches and the gate reports "not applicable" forever.
// A check that can only ever be inapplicable is indistinguishable from a check that holds.
//
// This asserts the PAIR against enums.go, which is where the vocabulary is actually declared.
// Re-point the gate at `outcome`, or rename PASS, and this fails — instead of the gate going
// quietly dark the way it already did once.
func TestPassVerdictIsAWordItsEventTypeCanActuallyCarry(t *testing.T) {
	declared, found := record.Enum(passVerdictType, "verdict")
	if !found {
		t.Fatalf("no declared vocabulary for (%s, verdict) — the gate switches on a word "+
			"nothing validates at the write", passVerdictType)
	}
	if !declared.Allows(passVerdictWord) {
		var have []string
		for _, v := range declared.Values {
			have = append(have, v.Name)
		}
		t.Errorf("%s events cannot carry %q (they carry %v), so the gate's comparison can "+
			"NEVER match and it reports 'not applicable' on every run",
			passVerdictType, passVerdictWord, have)
	}
	// And the event type it was wrongly pointed at must still be the wrong one — if the two
	// vocabularies ever converge, the confusion this guards against stops being detectable
	// and this test needs rewriting rather than silently continuing to pass.
	if other, ok := record.Enum("outcome", "verdict"); ok && other.Allows(passVerdictWord) {
		t.Errorf("`outcome` now also allows %q; the two vocabularies have converged and this "+
			"guard no longer distinguishes the event types", passVerdictWord)
	}
}

// THE THIRD STATE (#411). `pass-closes-all-gaps` was inapplicable on every run ever recorded
// and printed as `[ok  ]`, which reads as a check that held. These pin the distinction.
func TestAnInapplicableCheckIsMarkedNAAndIsNotAFailure(t *testing.T) {
	// A REGISTERED seat, because this test also asserts the whole run has no failures — and a
	// bare event with no seat id trips register-before-append, which would make the assertion
	// pass or fail for a reason that has nothing to do with the third state.
	b := &record.Board{
		Events: []*record.Event{
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.Register{}),
			recordtest.Event(t, "red-merge-r1", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}),
		},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
	}
	got := find(t, Run(b), "pass-closes-all-gaps")
	if !got.NA {
		t.Errorf("a FAIL verdict leaves the PASS gate inapplicable; NA must say so: %+v", got)
	}
	if got.Status() != "n/a" {
		t.Errorf("Status() = %q, want n/a", got.Status())
	}
	// The exit code must not move. An inapplicable check is not a violation — a legitimately
	// halted run leaves this gate inapplicable and that is correct.
	if !got.OK {
		t.Error("an inapplicable check must not read as failed; the exit code is for violations")
	}
	if len(Failed(Run(b))) != 0 {
		t.Error("an inapplicable check must not appear in Failed()")
	}
	if na := NotApplicable(Run(b)); len(na) != 1 || na[0].Name != "pass-closes-all-gaps" {
		t.Errorf("NotApplicable() = %+v, want exactly the PASS gate", na)
	}
	if got.Detail == "" {
		t.Error("an unexplained 'did not apply' is the unreadable zero this state exists to remove")
	}
}

// A check that HELD must not be marked n/a — the distinction has to cut both ways or it is
// decoration.
func TestAHeldCheckIsNotMarkedNA(t *testing.T) {
	b := &record.Board{
		Events:   []*record.Event{recordtest.Event(t, "", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})},
		GapOrder: []string{"R1-1"},
		Gaps: map[string]*record.Gap{
			"R1-1": {ID: "R1-1", Open: false, Closure: p("closure_class", "closed")},
		},
	}
	got := find(t, Run(b), "pass-closes-all-gaps")
	if got.NA {
		t.Errorf("a PASS that closed everything HELD; it did not fail to apply: %+v", got)
	}
	if got.Status() != "ok" {
		t.Errorf("Status() = %q, want ok", got.Status())
	}
}

// A violated check reports FAIL, and is neither n/a nor ok.
func TestAViolatedCheckReportsFail(t *testing.T) {
	b := &record.Board{
		Events:   []*record.Event{recordtest.Event(t, "", 0, &recordpb.RoundVerdict{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)})},
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
	}
	got := find(t, Run(b), "pass-closes-all-gaps")
	if got.OK || got.NA || got.Status() != "FAIL" {
		t.Errorf("a PASS over an open gap must be FAIL, got %+v (status %q)", got, got.Status())
	}
}
