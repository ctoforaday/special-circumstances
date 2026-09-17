package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ManifestOwed is the set of gaps blue owed a correctness-manifest row, in the order each first
// became owed. A row is owed for a gap blue REPAIRED while it was open, which the record states as
// three facts together:
//
//  1. blue-respond was DISPATCHED onto the gap;
//  2. the gap was not CLOSED between that dispatch and blue-respond's REGISTER for the sitting;
//  3. a BlueEdit by blue-respond in that sitting ANSWERS the gap.
//
// It is the one predicate the scorecard divides by, the report names missing receipts from, and
// blue's work list offers the receipt from. The blue constitutions have always said "one
// manifest-row event per repaired gap"; this is that sentence read off the record.
//
// A GAP ITS LENS CLOSED FIRST OWES NOTHING (#868). The lenses sit before blue, so a lens can close
// a gap between the dispatch that engaged blue on it and blue's sitting; blue then finds nothing to
// repair and correctly files no row. The register is blue's first act of the sitting (record.go:
// "every seat's first record action"), so a close before it means the gap was closed when blue sat.
// A close after it is a closure of blue's repair, and the row was owed.
//
// A GAP BLUE REBUTTED OWES NOTHING. With no edit answering it, nothing was repaired, and a receipt
// for a repair that did not happen is not a receipt. The rebuttal is on the record as blue's
// position; the manifest is not where it goes.
//
// THE SITTING IS BlueSittings': from blue's register to where sittingCloser ends it, read at when.
// A sitting nothing has closed is read to the end of the record either way, which is what lets a
// live sitting's work list see its own unreceipted edits — so Gaps is the same at either reading,
// and while the run is running Unresolved says those sittings may yet hold more.
// A dispatch blue never sat owes nothing: no register, no sitting, and a missing sitting is the
// sitting-record audit's finding, not the manifest's.
func ManifestOwed(evs []*Event, when ReadWhen) ManifestOwing {
	var o ManifestOwing
	seen := map[string]bool{}
	for _, s := range BlueSittings(evs, when) {
		if s.Unresolved {
			o.Unresolved++
		}
		open := map[string]bool{}
		for _, g := range s.Open {
			open[g] = true
		}
		for _, e := range s.Acts {
			if be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
				if g := be.GetAnswers(); open[g] && !seen[g] {
					seen[g] = true
					o.Gaps = append(o.Gaps, g)
				}
			}
		}
	}
	return o
}

// ManifestOwing is what the record says blue owed the manifest: the gaps, and how many of blue's
// sittings it cannot close — never any after the run, when the end of the record closes them. THE GAPS ARE A FLOOR AND UNRESOLVED IS THE REST OF THE ANSWER — an
// unresolved sitting's edits are read as far as the record reaches, and it may yet hold more, so a
// reader handed Gaps alone would read "owed nothing more" where the record says "not yet known".
type ManifestOwing struct {
	Gaps       []string
	Unresolved int // blue sittings the record cannot close (BlueSitting.Unresolved)
}

// NotMeasured words the unresolved sittings for a reader of the owed set, and it is THE ONE PLACE
// that case is worded; it is empty when the record closes every blue sitting.
func (o ManifestOwing) NotMeasured() string {
	if o.Unresolved == 0 {
		return ""
	}
	return fmt.Sprintf("%d blue sitting(s) NOT MEASURED — the record cannot close them (blue has not registered since and its agent's stop is not on the record), so what they owed is counted only as far as the record reaches", o.Unresolved)
}

// BlueSitting is one sitting blue-respond took for a dispatch: from its register to where
// sittingCloser ends it.
type BlueSitting struct {
	Engaged []string // the gaps the dispatch named
	Open    []string // of those, the ones no close preceded blue's register — what the sitting owed an answer on
	Acts    []*Event // blue-respond's standing events in the sitting, its register first, then any repair of it
	// Unresolved is a sitting the record cannot close while the run is running: neither blue's next
	// register nor its agent's stop is on the record since it began. Acts then runs to the end of the
	// record — a floor, not the sitting — so a missing act in it is not a finding and a present one
	// is. Read after the run, the end of the record closes it and it is never Unresolved.
	Unresolved bool
	opened     int64 // the place, in Live order, of the register that opened it
}

// Owes is what this sitting owed and did not file: a position and a revision, each missing one, for
// a sitting that found a gap it was engaged on still open. It is THE ONE PREDICATE of an owing
// sitting — capture's record-parity audit fails on it, and the register refuses a repair of a
// sitting for which it is empty.
func (s BlueSitting) Owes() []recordpb.EventType {
	if len(s.Open) == 0 {
		return nil
	}
	var position, revision bool
	for _, e := range s.Acts {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_POSITION:
			position = true
		case recordpb.EventType_EVENT_TYPE_REVISION:
			revision = true
		}
	}
	var owes []recordpb.EventType
	if !position {
		owes = append(owes, recordpb.EventType_EVENT_TYPE_POSITION)
	}
	if !revision {
		owes = append(owes, recordpb.EventType_EVENT_TYPE_REVISION)
	}
	return owes
}

// BlueSittings is every blue-respond sitting for a dispatch, in stream order. It is the one
// reading of "what did this blue sitting owe": the manifest counts receipts off it, capture's
// record-parity audit holds each owing sitting to a position and a revision, and blue's work list
// asks it whether this sitting owes a revision. A dispatch blue never sat has no sitting here.
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
func BlueSittings(evs []*Event, when ReadWhen) []BlueSitting {
	// The acts that stand: a corrected close or row is read as its replacement, in its place.
	evs = Live(evs)
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, registers := dispatchLedger(evs, seq)
	closer := sittingCloserOf(evs, seq, registers, when)
	var out []BlueSitting
	for k, d := range ds {
		if d.seat != blueRespondSeat {
			continue
		}
		start, sat := sittingFor(registers[blueRespondSeat], d)
		if !sat || laterBlueDispatchBefore(ds[k+1:], start) {
			continue // not sat, or not this sitting's dispatch: a later row before the register is
		}
		s := BlueSitting{Engaged: d.gaps, opened: start}
		closedFirst := map[string]bool{}
		for _, e := range evs[d.at+1 : start] {
			if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok {
				closedFirst[c.GetGapId()] = true
			}
		}
		for _, g := range d.gaps {
			if !closedFirst[g] {
				s.Open = append(s.Open, g)
			}
		}
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

// ManifestUnreceipted is ManifestOwed less every gap a manifest-row event names, in the same order:
// the repairs nobody audited, including their author. Unresolved carries over: a sitting the record
// cannot close may yet file the row, so its unreceipted gaps are what the record holds so far.
func ManifestUnreceipted(evs []*Event, when ReadWhen) ManifestOwing {
	evs = Live(evs)
	rowed := map[string]bool{}
	for _, e := range evs {
		if mr, ok := recordpb.BodyAs[*recordpb.ManifestRow](e); ok && mr.GetGapId() != "" {
			rowed[mr.GetGapId()] = true
		}
	}
	owed := ManifestOwed(evs, when)
	out := ManifestOwing{Unresolved: owed.Unresolved}
	for _, g := range owed.Gaps {
		if !rowed[g] {
			out.Gaps = append(out.Gaps, g)
		}
	}
	return out
}
