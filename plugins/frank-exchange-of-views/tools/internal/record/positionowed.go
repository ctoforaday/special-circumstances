package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A POSITION IS THE CHAIR'S, AND THIS FILE IS WHERE WHEN IT IS OWED IS SAID (gblock, 2026-10-20).
//
// A party's argument to the bench on one gap is its closing (closingowed.go). The position is the
// chair's argument on what is about no one gap: why it withheld a verdict the board permitted, why a
// step the record offered it — a gap not put before the bench, a closure not restated — was left
// untaken. The chair is the one seat that holds the verb; a lens and blue argue on a gap in their
// closings and in their acts.
//
// IT IS OWED WHERE THE CHAIR WITHHELD A PASS THE PLAN PERMITTED: a chair sitting whose last stored
// plan permitted a PASS and that recorded none. A sitting that recorded the PASS argues nothing it
// withheld, and one whose plan permitted none withheld nothing the board offered — the plan's
// blockers already say what holds the gate. The chair may file a position at any sitting; this is
// when one is owed.
//
// ONE PREDICATE, AND EVERY SURFACE THAT AUDITS OR RENDERS THE DUTY CALLS IT: capture's record-parity
// audit fails a closed sitting short of one. The chair's work list holds no item for it — its
// blocking items are the PASS gate's blockers and nothing else — so the duty is held by its prompt
// and by capture.
//
// THE TRANSCRIPT PRINTS WHAT WAS FILED AND NOTHING ELSE: a sitting that holds no position has no
// section there, and capture's audit is the reader that reports it.

// PositionState is what the record says of one chair sitting's position. THE FOUR STATES ARE
// DEFINED HERE AND NOWHERE ELSE (positionStateOf): a reader that asks "did this sitting file its
// position" gets one of them, never an empty answer it must interpret.
type PositionState string

const (
	// PositionNotOwed: the sitting withheld no PASS its plan permitted — it recorded the PASS, or its
	// plan permitted none, or it stored no plan.
	PositionNotOwed PositionState = "not_owed"
	// PositionFiled: owed, and the sitting's acts hold a position.
	PositionFiled PositionState = "filed"
	// PositionMissing: owed, none filed, and the record closes the sitting — the chair's next
	// opening, the stop of the agent that opened it, or, read after the run, the end of the record.
	PositionMissing PositionState = "missing"
	// PositionUnresolved: owed, none filed so far, and the record cannot close the sitting. It may
	// be in flight; it is neither complete nor short.
	PositionUnresolved PositionState = "unresolved"
)

// positionStateOf is the one definition of the four states.
func positionStateOf(owed, filed, closed bool) PositionState {
	switch {
	case !owed:
		return PositionNotOwed
	case filed:
		return PositionFiled
	case closed:
		return PositionMissing
	}
	return PositionUnresolved
}

// PositionSitting is one chair sitting and what the record says of its position.
type PositionSitting struct {
	// Ordinal counts the chair's sittings from 1, in stream order: "red-chair sitting 3".
	Ordinal int
	// SittingID is the stored sitting the row is.
	SittingID int64
	State     PositionState
}

// PositionSittings is one row per chair sitting, in stream order, read at when: the openings
// dispatchLedger holds for the chair, bounded by sittingCloser — the reading BlueSittings and the
// manifest use, so a sitting is closed for this function exactly when it is closed for them.
//
// WITHHELD IS READ OFF THE SITTING'S OWN ACTS. Its last stored plan (StoredPlans' reading of the
// printed plan) says whether a PASS was permitted; a gate recording PASS among its acts says it was
// recorded. A plan stored in another sitting answers nothing here.
func PositionSittings(evs []*Event, win WindowIndex, when ReadWhen) []PositionSitting {
	var out []PositionSitting
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	_, openings := dispatchLedger(evs, seq, win)
	closer := sittingCloserOf(evs, seq, win, openings, when)
	for k, start := range openings[chairSeat] {
		spans, end, closed := closer.bounds(chairSeat, start)
		filed, permitted, passed := false, false, false
		for i := start; i < end; i++ {
			e := evs[i]
			if e.GetSeatId() != chairSeat || !holds(spans, i) {
				continue
			}
			switch b := mustBodyOf(e).(type) {
			case *recordpb.Position:
				filed = true
			case *recordpb.Plan:
				if p, _, err := ReadPlan([]byte(b.GetPrinted())); err == nil {
					permitted = p.PassPermitted
				}
			case *recordpb.Gate:
				passed = passed || b.GetVerdict() == recordpb.Verdict_VERDICT_PASS
			}
		}
		out = append(out, PositionSitting{Ordinal: k + 1, SittingID: win.Of(evs[start]).SittingID,
			State: positionStateOf(permitted && !passed, filed, closed)})
	}
	return out
}

// mustBodyOf is an event's body, or nil for an event that carries none.
func mustBodyOf(e *Event) any {
	b, _ := recordpb.Body(e)
	return b
}
