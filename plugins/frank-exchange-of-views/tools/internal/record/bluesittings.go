package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// BlueSitting is one sitting blue-respond took for a dispatch: from its register to where
// sittingCloser ends it.
type BlueSitting struct {
	Engaged []string // the gaps the dispatch named
	Open    []string // of those, the ones no close preceded blue's register — what the sitting owed an answer on
	// FoundClosed is the rest of Engaged, in its order: for each gap a closing act of its lens's
	// preceded blue's register on, that act — the latest close, whose seat and class are what blue
	// found when it sat, or the mint that closed it by superseding it.
	FoundClosed []FoundClosure
	Acts        []*Event // blue-respond's standing events in the sitting, its register first, then any repair of it
	// Unresolved is a sitting the record cannot close while the run is running: neither blue's next
	// register nor its agent's stop is on the record since it began. Acts then runs to the end of the
	// record — a floor, not the sitting — so a missing act in it is not a finding and a present one
	// is. Read after the run, the end of the record closes it and it is never Unresolved.
	Unresolved bool
	opened     int64 // the place, in record order, of the register that opened it
}

// BlueSittings is every blue-respond sitting for a dispatch, in stream order. It is the one
// reading of "what did this blue sitting owe": the work list names the gaps it found closed, and
// blue's answered list reads the gaps it found open. A dispatch blue never sat has no sitting here.
//
// ITS BOUNDS ARE THE SHARED ONES. The sitting begins where sittingFor says and ends where
// sittingCloser says at when — blue's own next register, its agent's stop, or after the run the end
// of the record — with the acts of every register repairing it, and nothing the chair writes
// ends it: the chair's next dispatch row naming blue is another seat's act, and a read keyed on it
// depends on a rule enforced at the chair's write path (#1002).
//
// ENGAGED IS THE LAST DISPATCH BEFORE THE REGISTER. A chair that writes the plan twice leaves two
// rows for one sitting, and the later is the dispatch; a close counts as "closed first" only
// between that row and the register.
func BlueSittings(evs []*Event, win WindowIndex, when ReadWhen) []BlueSitting {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, openings := dispatchLedger(evs, seq, win)
	closer := sittingCloserOf(evs, seq, win, openings, when)
	var out []BlueSitting
	for k, d := range ds {
		if d.seat != blueRespondSeat {
			continue
		}
		start, sat := sittingFor(openings[blueRespondSeat], d)
		if !sat || laterBlueDispatchBefore(ds[k+1:], start) {
			continue // not sat, or not this sitting's dispatch: a later row before the register is
		}
		s := engagementOf(evs, d, start)
		s.opened = start
		spans, end, closed := closer.bounds(blueRespondSeat, start)
		s.Unresolved = !closed
		for i := start; i < end; i++ {
			if evs[i].GetSeatId() == blueRespondSeat && holds(spans, i) {
				s.Acts = append(s.Acts, evs[i])
			}
		}
		out = append(out, s)
	}
	return out
}

// engagementOf is what a dispatch row engages blue-respond on, read up to end: its gaps, split by
// whether a close landed on each between the row and end. It is the one split of a dispatch's gaps
// into found-closed and open — BlueSittings reads it up to the register that sat for the row, and
// the work list's `engaged` reads it up to the end of the record for a row not yet sat for
// (engagedOf), so the block blue reads before it registers is the one its sitting then holds.
func engagementOf(evs []*Event, d dispatchRow, end int64) BlueSitting {
	s := BlueSitting{Engaged: d.gaps}
	closedFirst := map[string]*Event{}
	superseded := supersessionsOf(evs[:end])
	for i, e := range evs[d.at+1 : end] {
		switch b := mustBody(e).(type) {
		case *recordpb.Close:
			closedFirst[b.GetGapId()] = e
		case *recordpb.Mint:
			for _, anc := range b.GetSupersedes() {
				if superseded[anc] == int(d.at)+1+i {
					closedFirst[anc] = e
				}
			}
		}
	}
	for _, g := range d.gaps {
		if c := closedFirst[g]; c != nil {
			s.FoundClosed = append(s.FoundClosed, FoundClosure{GapID: g, Act: c})
		} else {
			s.Open = append(s.Open, g)
		}
	}
	return s
}

// FoundClosure is one engaged gap blue found closed, and the lens's act that closed it: a close, or
// the successor's mint that superseded it.
type FoundClosure struct {
	GapID string
	Act   *Event
}

// Fate is the word the act closed the gap with: the close's class, or the supersession's.
func (f FoundClosure) Fate() string {
	if c, ok := recordpb.BodyAs[*recordpb.Close](f.Act); ok {
		return recordpb.Word(c.GetClosureClass())
	}
	return ClosedSuperseded
}

// laterBlueDispatchBefore reports whether a later row naming blue-respond precedes its register.
func laterBlueDispatchBefore(later []dispatchRow, register int64) bool {
	for _, d := range later {
		if d.at >= register {
			return false
		}
		if d.seat == blueRespondSeat {
			return true
		}
	}
	return false
}
