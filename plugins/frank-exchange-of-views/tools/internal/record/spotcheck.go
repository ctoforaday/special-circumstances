package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE SPOT-CHECKS ON THE RECORD, JOINED TO THE BOARD.
//
// A spot-check is owed before a PASS and at no other time, and the write enforces it: a PASS is
// refused while a stale area stands that no spot-check in the chair's sitting names
// (BlockerStaleArea, passblockers.go). Nothing here is a second statement of that rule. This file
// is the reader's account of the spot-checks the chair filed — which closures each re-read, and
// what the board says the archive held at that point, a number no seat can author.
//
// One claim in a spot-check is checkable against that number: "there was nothing to sample". A
// spot-check that says so while the board holds archived closures is reported to the reader as a
// claim the record contradicts.

// SpotCheck is one spot-check the chair filed, joined to what the board says was available to
// sample.
type SpotCheck struct {
	Epoch   int
	Sitting int
	SeatID  string
	// Sampled are the archived closures the chair says it re-verified.
	Sampled []string
	// Prose is what the sample found, or why there was nothing to sample.
	Prose string
	// None is a claim that there was nothing to sample.
	None       bool
	NoneReason string
	// Archived is the board's count of closures available at the START of this epoch — what the
	// claim is read against.
	Archived int
}

// SpotCheckAudit joins each recorded spot-check to the board: the spot-checks in order, and those
// whose emptiness claim the board contradicts.
func SpotCheckAudit(f Family) (checks []SpotCheck, falseEmpty []SpotCheck) {
	if len(f.Gaps) == 0 && len(f.Events) == 0 {
		return nil, nil
	}
	// The archive at the START of epoch R: every gap closed in an epoch strictly before R.
	// Replayed state, not a reported count.
	//
	// The epoch is the closing event's, derived by the record from the chair's stored sittings
	// (events_w), so no seat-id shape stands between a closure and its epoch.
	archivedBefore := func(epoch int) int {
		n := 0
		for _, g := range f.Gaps {
			if g != nil && g.HasClosed && g.ClosedEpoch < epoch {
				n++
			}
		}
		return n
	}

	// THE STORED EPOCH, because the number goes on a recorded ACT (the spot-check the chair filed),
	// and the record holds which of the chair's sittings each act is in.
	for _, e := range f.Live() {
		// The BODY is the type test. Named `body` because the projection struct this loop fills
		// is ALSO called SpotCheck — record.SpotCheck is the audit row, recordpb.SpotCheck the
		// event body.
		body, ok := recordpb.BodyAs[*recordpb.SpotCheck](e)
		if !ok {
			continue
		}
		w := f.At.Of(e)
		sc := SpotCheck{
			Epoch: w.Epoch, Sitting: w.Sitting, SeatID: e.GetSeatId(),
			Sampled: body.GetIds(), Prose: body.GetReason(),
			Archived: archivedBefore(w.Epoch),
		}
		// An absent `none` is false here, and a recorded `none: false` is a seat NOT claiming an
		// empty archive — the same answer.
		if body.GetNone() {
			sc.None = true
			sc.NoneReason = body.GetReason()
		}
		checks = append(checks, sc)
		// THE CLAIM THE BOARD CAN CONTRADICT. "There was nothing to sample" is checkable.
		if sc.None && sc.Archived > 0 {
			falseEmpty = append(falseEmpty, sc)
		}
	}
	return checks, falseEmpty
}

// Describe renders one spot-check as a reader-facing line.
func (s SpotCheck) Describe() string {
	switch {
	case s.None:
		return fmt.Sprintf("#%d (%s): **nothing to sample** — %s", s.Sitting, s.SeatID, s.NoneReason)
	case len(s.Sampled) == 0:
		return fmt.Sprintf("#%d (%s): recorded, naming no closures", s.Sitting, s.SeatID)
	default:
		line := fmt.Sprintf("#%d (%s): re-verified %s of %d archived closure(s)",
			s.Sitting, s.SeatID, strings.Join(s.Sampled, ", "), s.Archived)
		if s.Prose != "" {
			line += " — " + s.Prose
		}
		return line
	}
}
