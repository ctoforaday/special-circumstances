package record

import (
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A POSITION IS OWED BY TWO SEATS, AND THIS FILE IS WHERE THAT IS SAID.
//
// The position is a seat's argument to the bench for one sitting, and the transcript is its only
// carrier: the bench and the chair hold the projection that renders it, and no other seat does. So
// the seats that owe one are the two that argue before the bench every sitting they sit — the chair,
// and blue's responding seat. A lane, the frontier and the synthesizer argue in what they write:
// a draft, the avenues, the report.
//
// ONE PREDICATE, AND EVERY SURFACE THAT ENFORCES, LISTS, AUDITS OR RENDERS THE DUTY CALLS IT: the
// verb refuses a seat it is false for, the work list blocks blue-respond on it, and capture's
// record-parity audit fails a closed sitting short of one. A second computation of "does this seat
// owe a position" is how a seat comes to be told to do something its surface refuses.
//
// THE TRANSCRIPT PRINTS WHAT WAS FILED AND NOTHING ELSE: a sitting that holds no position has no
// section there, and capture's audit is the reader that reports it.

// positionSeats is every seat that owes a position, in the order a refusal names them.
var positionSeats = []string{blueRespondSeat, chairSeat}

// SeatOwesPosition reports whether a seat owes a position at its sittings.
func SeatOwesPosition(seatID string) bool {
	for _, s := range positionSeats {
		if s == seatID {
			return true
		}
	}
	return false
}

// PositionSeats is the seats SeatOwesPosition is true of, joined for a sentence — what a refusal
// names, so the seats it sends a reader to are the seats the write path admits.
func PositionSeats() string { return strings.Join(positionSeats, " and ") }

// PositionState is what the record says of one sitting's position. THE FOUR STATES ARE DEFINED
// HERE AND NOWHERE ELSE (positionStateOf): a reader that asks "did this sitting file its position"
// gets one of them, never an empty answer it must interpret.
type PositionState string

const (
	// PositionNotOwed: a blue-respond sitting that found every gap it was engaged on closed when it
	// sat. It has nothing to argue, and its absence of a position is not a silence.
	PositionNotOwed PositionState = "not_owed"
	// PositionFiled: owed, and the sitting's acts hold a position.
	PositionFiled PositionState = "filed"
	// PositionMissing: owed, none filed, and the record closes the sitting — its seat's next
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

// holdsPosition reports whether a sitting's acts hold a position.
func holdsPosition(acts []*Event) bool {
	for _, e := range acts {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
			return true
		}
	}
	return false
}

// PositionSitting is one sitting of a seat that owes a position, and what the record says of it.
type PositionSitting struct {
	Seat string
	// Ordinal counts the seat's rows from 1, in stream order: "red-chair sitting 3".
	Ordinal int
	// SittingID is the stored sitting the row is — what the work list holds against the seat's
	// latest sitting.
	SittingID int64
	// Blue is, for blue-respond, the sitting as BlueSittings reads it: the gaps it found open are
	// what a finding about it names. Nil for the chair, whose every sitting owes a position.
	Blue   *BlueSitting
	State  PositionState
	opened int64
}

// PositionSittings is one row per sitting of each seat SeatOwesPosition names, in stream order,
// read at when. A sitting of any other seat has no row.
//
// ITS SITTINGS ARE THE SHARED ONES. blue-respond's rows are BlueSittings' — a sitting for a
// dispatch, owing exactly when it found an engaged gap open. The chair's are the openings
// dispatchLedger holds for it, bounded by sittingCloser, the two readings BlueSittings uses: a
// sitting is closed for this function exactly when it is closed for the manifest.
//
// THE READERS DIFFER ONLY IN WHAT THEY DO WITH A STATE. The work list treats missing and unresolved
// alike, because the seat reading its own list can still file. Capture fails missing and never
// fails unresolved.
func PositionSittings(evs []*Event, win WindowIndex, when ReadWhen) []PositionSitting {
	var out []PositionSitting
	live := Live(evs)
	if SeatOwesPosition(blueRespondSeat) {
		for k, s := range BlueSittings(evs, win, when) {
			out = append(out, PositionSitting{Seat: blueRespondSeat, Ordinal: k + 1, Blue: &s, opened: s.opened,
				SittingID: win.Of(live[s.opened]).SittingID,
				State:     positionStateOf(len(s.Open) > 0, holdsPosition(s.Acts), !s.Unresolved)})
		}
	}
	if SeatOwesPosition(chairSeat) {
		seq := make([]int64, len(live))
		for i := range seq {
			seq[i] = int64(i)
		}
		_, registers := dispatchLedger(live, seq, win)
		closer := sittingCloserOf(live, seq, win, registers, when)
		for k, start := range registers[chairSeat] {
			spans, end, closed := closer.bounds(chairSeat, start)
			filed := false
			for i := start; i < end && !filed; i++ {
				filed = live[i].GetSeatId() == chairSeat && holds(spans, i) &&
					live[i].GetType() == recordpb.EventType_EVENT_TYPE_POSITION
			}
			out = append(out, PositionSitting{Seat: chairSeat, Ordinal: k + 1, opened: start,
				SittingID: win.Of(live[start]).SittingID, State: positionStateOf(true, filed, closed)})
		}
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].opened < out[b].opened })
	return out
}

// positionSittingNow is the seat's row for the sitting it is in — the latest sitting the record
// holds for it — read while the run is running. ok is false for a seat that owes no position, and
// for one whose latest sitting has no row: a blue-respond sitting no dispatch answers to.
func positionSittingNow(evs []*Event, win WindowIndex, seatID string) (PositionSitting, bool) {
	if !SeatOwesPosition(seatID) {
		return PositionSitting{}, false
	}
	latest := win.LatestSittingOf(seatID)
	rows := PositionSittings(evs, win, WhileRunning)
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i].Seat == seatID {
			return rows[i], rows[i].SittingID == latest
		}
	}
	return PositionSitting{}, false
}
