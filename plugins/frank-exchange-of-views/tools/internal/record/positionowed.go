package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
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
// audit fails a closed sitting short of one. The chair's work list holds no item for the position
// itself, so that duty is held by its prompt and by capture. The bench hears the position it files
// (positionheard.go), and the sitting after an overruling owes the verdict instead: that one IS on
// the chair's work list (overruledVerdictItem), and capture fails a closed sitting short of it.
//
// THE TRANSCRIPT PRINTS WHAT WAS FILED AND NOTHING ELSE: a sitting that holds no position has no
// section there, and capture's audit is the reader that reports it.

// PositionRefusal is who may file a position: the chair, and no other seat. The write path refuses
// on it whatever surface mounts the verb, and migrate writes no event for an archived position it
// refuses.
func PositionRefusal(seatID string) error {
	if seatID == chairSeat {
		return nil
	}
	return feov.Errorf(feov.RoleViolation,
		"a position is the chair's argument to the bench on what is about no one gap, and %s is not the chair; nothing was recorded. "+
			"A party's argument on a docketed gap is its closing, and a seat's argument on its own work is in its acts",
		seatID)
}

// PositionState is what the record says of one chair sitting's position. THE STATES ARE DEFINED
// HERE AND NOWHERE ELSE (positionStateOf): a reader that asks "did this sitting file its position"
// gets one of them, never an empty answer it must interpret.
type PositionState string

const (
	// PositionNotOwed: the sitting withheld no PASS its plan permitted — it recorded the PASS, or its
	// plan permitted none, or it stored no plan — and it owes no verdict after an overruled position.
	PositionNotOwed PositionState = "not_owed"
	// PositionFiled: owed, and the sitting's acts hold a position the bench hears (heardPositionsOf).
	PositionFiled PositionState = "filed"
	// PositionMissing: owed, none filed, and the record closes the sitting — the chair's next
	// opening, the stop of the agent that opened it, or, read after the run, the end of the record.
	PositionMissing PositionState = "missing"
	// PositionUnresolved: owed, not discharged so far, and the record cannot close the sitting. It
	// may be in flight; it is neither complete nor short.
	PositionUnresolved PositionState = "unresolved"
	// PositionVerdictMissing: the sitting is the chair's first after the bench overruled its position,
	// its plan permitted the PASS, and it closed with no verdict (gblock, 2026-10-30): the overruled
	// objection is decided, and the sitting owes the verdict the PASS was held back from.
	PositionVerdictMissing PositionState = "verdict_missing"
)

// positionStateOf is the one definition of the states. overruled says the sitting is the chair's
// first after the bench overruled its position, where the duty is the verdict.
func positionStateOf(owed, discharged, closed, overruled bool) PositionState {
	switch {
	case !owed:
		return PositionNotOwed
	case discharged && overruled:
		return PositionNotOwed // the verdict the overruling left owed is recorded
	case discharged:
		return PositionFiled
	case !closed:
		return PositionUnresolved
	case overruled:
		return PositionVerdictMissing
	}
	return PositionMissing
}

// PositionSitting is one chair sitting and what the record says of its position.
type PositionSitting struct {
	// Ordinal counts the chair's sittings from 1, in stream order: "red-chair sitting 3".
	Ordinal int
	// SittingID is the stored sitting the row is.
	SittingID int64
	State     PositionState
	// Overruled is the position the bench overruled before this sitting opened, where the sitting is
	// the chair's first after that ruling; "" for every other sitting.
	Overruled string
}

// PositionSittings is one row per chair sitting, in stream order, read at when: the openings
// dispatchLedger holds for the chair, bounded by sittingCloser — the reading BlueSittings and the
// manifest use, so a sitting is closed for this function exactly when it is closed for them.
//
// WITHHELD IS READ OFF THE SITTING'S OWN ACTS, as heardPositionsOf reads it: the plan the sitting
// stored last before its position — or, with no position heard, the last it stored — says whether a
// PASS was permitted, and a gate recording PASS among its acts says it was recorded. A plan stored
// in another sitting answers nothing here. A position heard by the bench is the duty discharged.
//
// THE SITTING AFTER AN OVERRULED POSITION OWES A VERDICT INSTEAD: where its plan permitted the PASS,
// a verdict among its acts discharges it, and a position there is refused at the write
// (requireNoRepeatedPosition).
func PositionSittings(evs []*Event, win WindowIndex, when ReadWhen) []PositionSitting {
	var out []PositionSitting
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	_, openings := dispatchLedger(evs, seq, win)
	closer := sittingCloserOf(evs, seq, win, openings, when)
	heard := heardPositionsOf(evs, win)
	overruled := overruledFollowers(motionsAt(evs, seq, win), openings[chairSeat])
	for k, start := range openings[chairSeat] {
		spans, end, closed := closer.bounds(chairSeat, start)
		filed, permitted, passed, verdict := false, false, false, false
		for i := start; i < end; i++ {
			e := evs[i]
			if e.GetSeatId() != chairSeat || !holds(spans, i) {
				continue
			}
			switch b := mustBodyOf(e).(type) {
			case *recordpb.Position:
				filed = filed || heard[e]
			case *recordpb.Plan:
				if p, _, err := ReadPlan([]byte(b.GetPrinted())); err == nil && !filed {
					permitted = p.PassPermitted
				}
			case *recordpb.Gate:
				verdict = true
				passed = passed || b.GetVerdict() == recordpb.Verdict_VERDICT_PASS
			}
		}
		row := PositionSitting{Ordinal: k + 1, SittingID: win.Of(evs[start]).SittingID, Overruled: overruled[start]}
		if row.Overruled != "" {
			row.State = positionStateOf(permitted, verdict, closed, true)
		} else {
			row.State = positionStateOf(filed || (permitted && !passed), filed, closed, false)
		}
		out = append(out, row)
	}
	return out
}

// THE BENCH HEARS A WITHHELD PASS (gblock, 2026-10-30). heardPositionsOf is the one definition of
// which positions the bench hears: a chair's position filed in a sitting whose last stored plan
// before it permitted a PASS, and that records no PASS. A position filed anywhere else argues to the
// bench at the exit and readies nobody. The sitting is the stored one (win), and each sitting holds
// one position (oncepersitting.go), so a sitting has at most one heard position. Once one is heard
// the PASS gate refuses the sitting's PASS on it, so "records no PASS" is decided at the position on
// a live record; an archived sitting that recorded a PASS after its position withheld nothing.
//
// A HEARD POSITION IS A MOTION of subject `position` (motionsAt): the position is its filing and its
// id is the position's own, as an inquiry's proposal is an inquiry motion's. Unruled, it holds the
// PASS and readies the bench for a position sitting (benchSittingFor).
func heardPositionsOf(evs []*Event, win WindowIndex) map[*Event]bool {
	// Only a sitting that holds a position asks its plans, so a record with none parses none.
	withPosition := map[int64]bool{}
	for _, e := range evs {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION && e.GetSeatId() == chairSeat {
			withPosition[win.Of(e).SittingID] = true
		}
	}
	heard := map[*Event]bool{}
	if len(withPosition) == 0 {
		return heard
	}
	passed := map[int64]bool{}
	for _, e := range evs {
		if g, ok := recordpb.BodyAs[*recordpb.Gate](e); ok && e.GetSeatId() == chairSeat && g.GetVerdict() == recordpb.Verdict_VERDICT_PASS {
			passed[win.Of(e).SittingID] = true
		}
	}
	permitted := map[int64]bool{}
	for _, e := range evs {
		id := win.Of(e).SittingID
		if e.GetSeatId() != chairSeat || !withPosition[id] || passed[id] {
			continue
		}
		switch b := mustBodyOf(e).(type) {
		case *recordpb.Plan:
			if p, _, err := ReadPlan([]byte(b.GetPrinted())); err == nil {
				permitted[id] = p.PassPermitted
			}
		case *recordpb.Position:
			heard[e] = permitted[id]
		}
	}
	return heard
}

// overruledFollowers maps the opening of each chair sitting that is the chair's first after the
// bench overruled a position to that position's id. A ruling after the chair's last opening — the
// terminal bench's — has no following sitting.
func overruledFollowers(motions []*Motion, chairOpenings []int64) map[int64]string {
	out := map[int64]string{}
	for _, m := range motions {
		if m == nil || m.Subject != positionSubject || m.Ruling != positionOverruled {
			continue
		}
		if next, ok := firstAfter(chairOpenings, m.ruled); ok {
			out[next] = m.ID
		}
	}
	return out
}

// positionSubject and the two ruling words are the schema's spellings, read once.
var (
	positionSubject   = recordpb.Word(recordpb.MotionSubject_MOTION_SUBJECT_POSITION)
	positionUpheld    = recordpb.Word(recordpb.PositionRuling_POSITION_RULING_UPHELD)
	positionOverruled = recordpb.Word(recordpb.PositionRuling_POSITION_RULING_OVERRULED)
)

// mustBodyOf is an event's body, or nil for an event that carries none.
func mustBodyOf(e *Event) any {
	b, _ := recordpb.Body(e)
	return b
}
