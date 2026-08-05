package record

import "testing"

// board assembles a Board directly from events, so these tests exercise the estoppel logic
// rather than the shard reader (which record_test.go already covers).
func board(t *testing.T, gapFix map[string]string, evs []Event) *Board {
	t.Helper()
	b := &Board{Gaps: map[string]*Gap{}}
	for id, fixNew := range gapFix {
		m := NewPayload().Set("gap_id", id)
		if fixNew != "" {
			m.Set("fix_basis", "verified").Set("fix_old", "OLD:"+id).Set("fix_new", fixNew)
		} else {
			m.Set("fix_basis", "proposed")
		}
		b.Gaps[id] = &Gap{ID: id, Open: true, Mint: m}
	}
	b.Events = evs
	return b
}

const prescribed = "Five verification approaches agree, all sharing one definition of primality."

func edit(gapID string, verbatim bool) Event {
	p := NewPayload().Set("answers", gapID).Set("old", "x").Set("new", "y").Set("text", "r")
	if verbatim {
		p.Set("applied_verbatim", true)
	}
	return Event{Type: "blue_edit", SeatID: "blue-respond-r1", Round: 1, Payload: p}
}

// The core rule: a finding located in text red prescribed and blue applied VERBATIM names
// the gap that prescribed it.
func TestEstoppelCatchesRelitigationOfRedsOwnPrescription(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []Event{edit("R1-1", true)})

	id, got := EstoppelConflict(b, "Five verification approaches agree, all sharing one definition of primality.")
	if id != "R1-1" {
		t.Fatalf("EstoppelConflict = %q, want R1-1 — red re-raising its own prescribed text went undetected", id)
	}
	if got != prescribed {
		t.Errorf("the conflict must return the prescribed text so the refusal can quote it, got %q", got)
	}
}

// A fragment of the prescribed sentence is the same act as quoting the whole of it.
func TestEstoppelMatchesAFragmentOfThePrescribedText(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []Event{edit("R1-1", true)})
	if id, _ := EstoppelConflict(b, "agree, all sharing one definition of primality"); id != "R1-1" {
		t.Errorf("a quoted FRAGMENT of red's own prescription escaped the guard (got %q)", id)
	}
}

// ESTOPPEL ATTACHES TO RED'S OWN WORDS AND NOTHING ELSE. If blue counter-edited, the text is
// blue's authorship and red audits it normally — that is the right to disagree staying real.
func TestNoEstoppelWhenBlueCounterEditedInstead(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []Event{edit("R1-1", false)})
	if id, _ := EstoppelConflict(b, prescribed); id != "" {
		t.Errorf("red was estopped from auditing text BLUE authored (gap %q) — a counter-edit is not red's prescription", id)
	}
}

// Text red never prescribed is auditable, obviously — the guard must not become a general
// shield over the report.
func TestNoEstoppelForUnrelatedText(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []Event{edit("R1-1", true)})
	if id, _ := EstoppelConflict(b, "An entirely different sentence about sieve performance and its costs."); id != "" {
		t.Errorf("an unrelated finding was estopped by gap %q — the guard is over-broad", id)
	}
}

// The overlap floor exists so a short prescription cannot shield half the report. Refusing a
// real finding is worse than missing an estoppel, so the guard declines to fire here.
func TestShortPrescriptionsDoNotEstop(t *testing.T) {
	b := board(t, map[string]string{"R1-1": "7 is prime."}, []Event{edit("R1-1", true)})
	if id, _ := EstoppelConflict(b, "7 is prime."); id != "" {
		t.Errorf("a %d-character prescription estopped a finding (gap %q); the floor is %d",
			len("7 is prime."), id, minEstoppelOverlap)
	}
}

// The decline rate is the measurement that falsifies the design if it comes back zero.
func TestDeclineStatsSeparatesAppliedFromDeclinedFromUnanswered(t *testing.T) {
	b := board(t, map[string]string{
		"R1-1": prescribed,          // applied verbatim
		"R1-2": prescribed + " Two", // blue counter-edited
		"R1-3": prescribed + " Three",
		"R1-4": "", // prose only: not an offer, must not be counted
	}, []Event{
		edit("R1-1", true),
		edit("R1-2", false),
		// R1-3 offered and never answered.
		{Type: "blue_edit", SeatID: "blue-respond-r1", Round: 1, Payload: NewPayload().Set("answers", "R1-4")},
	})

	offered, applied, declined := DeclineStats(b)
	if offered != 3 {
		t.Errorf("offered = %d, want 3 — only gaps carrying a concrete proposal are offers", offered)
	}
	if applied != 1 {
		t.Errorf("applied = %d, want 1", applied)
	}
	if declined != 1 {
		t.Errorf("declined = %d, want 1 — a counter-edit IS a decline", declined)
	}
	// The unanswered one is neither: scoring silence as agreement is how a decline rate lies.
	if unanswered := offered - applied - declined; unanswered != 1 {
		t.Errorf("unanswered = %d, want 1 — an unanswered offer must not be scored as agreement", unanswered)
	}
}

// A dispute is a decline too: blue contesting the grade is exercising the same right.
func TestDisputeCountsAsADecline(t *testing.T) {
	b := board(t, map[string]string{"R1-1": prescribed}, []Event{
		{Type: "dispute", SeatID: "blue-respond-r1", Round: 1, Payload: NewPayload().Set("gap_id", "R1-1")},
	})
	if _, applied, declined := DeclineStats(b); applied != 0 || declined != 1 {
		t.Errorf("applied=%d declined=%d, want 0/1 — a dispute is blue declining", applied, declined)
	}
}
