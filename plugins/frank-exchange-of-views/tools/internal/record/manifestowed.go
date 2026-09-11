package record

import "github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"

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
// THE SITTING RUNS FROM THE REGISTER TO THE NEXT blue-respond DISPATCH, or to the end of the record,
// which is what lets a live sitting's work list see its own unreceipted edits. A dispatch blue
// never sat owes nothing: no register, no sitting, and a missing sitting is the sitting-record
// audit's finding, not the manifest's. A second register inside a sitting (the sitting-record
// re-prompt) changes nothing.
func ManifestOwed(evs []*Event) []string {
	var owed []string
	seen := map[string]bool{}
	for _, s := range BlueSittings(evs) {
		open := map[string]bool{}
		for _, g := range s.Open {
			open[g] = true
		}
		for _, e := range s.Acts {
			if be, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
				if g := be.GetAnswers(); open[g] && !seen[g] {
					seen[g] = true
					owed = append(owed, g)
				}
			}
		}
	}
	return owed
}

// BlueSitting is one sitting blue-respond took for a dispatch, as ManifestOwed delimits it: from
// its register to the next blue-respond dispatch or the end of the record.
type BlueSitting struct {
	Engaged []string // the gaps the dispatch named
	Open    []string // of those, the ones no close preceded blue's register — what the sitting owed an answer on
	Acts    []*Event // blue-respond's standing events in the sitting, its register first
}

// BlueSittings is every blue-respond sitting for a dispatch, in stream order. It is the one
// reading of "what did this blue sitting owe": the manifest counts receipts off it, and capture's
// record-parity audit holds each owing sitting to a position and a revision. A dispatch blue never
// sat has no sitting here.
func BlueSittings(evs []*Event) []BlueSitting {
	// The acts that stand: a corrected close or row is read as its replacement, in its place.
	evs = Live(evs)
	var out []BlueSitting
	var engaged []string
	closedFirst := map[string]bool{}
	waiting := false
	cur := -1 // the sitting blue's acts belong to; -1 outside one
	for _, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_DISPATCH:
			if d, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok && d.GetSeatId() == "blue-respond" {
				engaged, closedFirst, waiting, cur = d.GetGapIds(), map[string]bool{}, true, -1
			}
		case recordpb.EventType_EVENT_TYPE_CLOSE:
			if c, ok := recordpb.BodyAs[*recordpb.Close](e); ok && waiting {
				closedFirst[c.GetGapId()] = true
			}
		case recordpb.EventType_EVENT_TYPE_REGISTER:
			if e.GetSeatId() == "blue-respond" && waiting {
				s := BlueSitting{Engaged: engaged}
				for _, g := range engaged {
					if !closedFirst[g] {
						s.Open = append(s.Open, g)
					}
				}
				out = append(out, s)
				cur, waiting = len(out)-1, false
			}
		}
		if cur >= 0 && e.GetSeatId() == "blue-respond" {
			out[cur].Acts = append(out[cur].Acts, e)
		}
	}
	return out
}

// ManifestUnreceipted is ManifestOwed less every gap a manifest-row event names, in the same order:
// the repairs nobody audited, including their author.
func ManifestUnreceipted(evs []*Event) []string {
	evs = Live(evs)
	rowed := map[string]bool{}
	for _, e := range evs {
		if mr, ok := recordpb.BodyAs[*recordpb.ManifestRow](e); ok && mr.GetGapId() != "" {
			rowed[mr.GetGapId()] = true
		}
	}
	var out []string
	for _, g := range ManifestOwed(evs) {
		if !rowed[g] {
			out = append(out, g)
		}
	}
	return out
}
