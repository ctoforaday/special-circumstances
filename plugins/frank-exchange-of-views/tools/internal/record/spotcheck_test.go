package record

import "testing"

// board builds a Board with the gaps and events a spot-check audit reads.
func spotBoard(gaps map[string]*Gap, evs ...Event) *Board {
	return &Board{Gaps: gaps, Events: evs}
}

func closedIn(round int) *Gap { return &Gap{HasClosed: true, ClosedRound: round, Open: false} }

// THE FLOOR IS OWED ONLY WHERE THERE WAS SOMETHING TO SAMPLE. W1.8 keyed it on the archive's
// state at round START precisely because a round-NUMBER rule degenerated: run 5's round-2 merge
// entered with an empty archive, so "from round 2" became the seat attesting blocks it was about
// to write itself.
func TestSpotCheckDebtOnlyWhereTheArchiveWasNonEmpty(t *testing.T) {
	// r1 closes a gap; r2 enters with it archived and samples nothing.
	b := spotBoard(map[string]*Gap{"R1-1": closedIn(1)},
		Event{Round: 1, Type: "close", SeatID: "red-merge-r1"},
		Event{Round: 2, Type: "position", SeatID: "red-merge-r2"},
	)
	_, debt, _ := SpotCheckAudit(b)
	if len(debt) != 1 || debt[0] != 2 {
		t.Fatalf("round 2 entered with an archived closure and sampled none — expected debt [2], got %v", debt)
	}

	// The SAME shape with the closure landing IN round 2 owes nothing: the archive was empty
	// when round 2 started, which is the whole point of keying on round start.
	b = spotBoard(map[string]*Gap{"R2-1": closedIn(2)},
		Event{Round: 2, Type: "close", SeatID: "red-merge-r2"},
	)
	if _, debt, _ := SpotCheckAudit(b); len(debt) != 0 {
		t.Errorf("a closure made DURING the round was not in the archive at its start; no debt is owed: %v", debt)
	}

	// A round the merge never sat in owes nothing — demanding a sample from an absent seat is
	// the round-number keying W1.8 replaced, in a new spelling.
	b = spotBoard(map[string]*Gap{"R1-1": closedIn(1)},
		Event{Round: 1, Type: "close", SeatID: "red-merge-r1"},
		Event{Round: 2, Type: "position", SeatID: "blue-respond-r2"},
	)
	if _, debt, _ := SpotCheckAudit(b); len(debt) != 0 {
		t.Errorf("the merge did not sit in round 2; no duty was skipped: %v", debt)
	}
}

// A discharge clears the debt.
func TestSpotCheckDischargeClearsTheDebt(t *testing.T) {
	b := spotBoard(map[string]*Gap{"R1-1": closedIn(1)},
		Event{Round: 1, Type: "close", SeatID: "red-merge-r1"},
		Event{Round: 2, Type: "spot-check", SeatID: "red-merge-r2",
			Payload: NewPayload().Set("ids", []string{"R1-1"}).Set("notes", "the anchor still resolves")},
	)
	checks, debt, falseEmpty := SpotCheckAudit(b)
	if len(debt) != 0 || len(falseEmpty) != 0 {
		t.Fatalf("a real sample discharges the duty: debt=%v falseEmpty=%v", debt, falseEmpty)
	}
	if len(checks) != 1 || checks[0].Archived != 1 {
		t.Fatalf("the discharge must carry the board's count of what was available: %+v", checks)
	}
	if got := checks[0].Describe(); got == "" {
		t.Error("a discharge must render for the reader")
	}
}

// THE DIRECT HEIR OF THE RUN-5 DEGENERACY. "There was nothing to sample" is a claim, and the
// board can now refuse it. Every repair before this one asked the seat for the number it was
// being checked against.
func TestAFalseEmptyClaimIsCaught(t *testing.T) {
	b := spotBoard(map[string]*Gap{"R1-1": closedIn(1)},
		Event{Round: 1, Type: "close", SeatID: "red-merge-r1"},
		Event{Round: 2, Type: "spot-check", SeatID: "red-merge-r2",
			Payload: NewPayload().Set("none", true).Set("reason", "nothing archived")},
	)
	_, debt, falseEmpty := SpotCheckAudit(b)
	if len(falseEmpty) != 1 {
		t.Fatalf("a --none claim against a non-empty archive must be caught: %v", falseEmpty)
	}
	if falseEmpty[0].Archived != 1 {
		t.Errorf("the contradiction must carry the board's count: %+v", falseEmpty[0])
	}
	// It is a FALSE CLAIM, not a missing one — reporting it as debt too would double-count the
	// same round under two different accusations.
	if len(debt) != 0 {
		t.Errorf("a round that recorded something is not a round that recorded nothing: %v", debt)
	}

	// An HONEST --none, against an archive the board agrees was empty, is a discharge.
	b = spotBoard(map[string]*Gap{},
		Event{Round: 1, Type: "spot-check", SeatID: "red-merge-r1",
			Payload: NewPayload().Set("none", true).Set("reason", "nothing archived")},
	)
	if _, _, fe := SpotCheckAudit(b); len(fe) != 0 {
		t.Errorf("an honestly-empty round is a discharge, not a violation: %v", fe)
	}
}

func TestSpotCheckAuditHandlesANilBoard(t *testing.T) {
	if c, d, f := SpotCheckAudit(nil); c != nil || d != nil || f != nil {
		t.Errorf("a nil board must not panic or invent violations: %v %v %v", c, d, f)
	}
}
