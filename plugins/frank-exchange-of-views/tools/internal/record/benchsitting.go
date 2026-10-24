package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// docketHeld is what holds the bench's current sitting open: every docket motion standing unruled
// on a gap that sitting is convened on (gblock's ruling, 2026-10-24). gaps is those gaps, in the
// order the dispatch rows name them; motions is the unruled-motion blockers (unruledMotionBlockers)
// on them, in filing order. Both are empty when the bench's latest sitting is not a docket sitting.
//
// THE RECORD ADMITS A SECOND DOCKET MOTION ON A GAP WHILE THE FIRST STANDS, and the plan keys the
// bench's sitting for a gap on its NEWEST unruled filing (benchSatFor). A bench that rules one of two
// has therefore sat for the other: no plan readies it again, and the run stalls with nobody ready
// until the terminal sitting (#1557). So the sitting that sat for both owes both, and its envelope is
// refused while either stands (BenchEnvelopeRefusal).
//
// THE GAPS ARE THE ROWS THIS SITTING ANSWERS: each bench dispatch row convening a docket whose first
// docket sitting after it is this one (benchRowSitting's rule). A motion on a gap no such row names
// is not this sitting's — a docket on a gap the bench was not convened on waits for the plan that
// readies it, and a petition for its petition sitting; both stay on the list as holding the outcome.
//
// The stream's positions stand in for events.id, as in owedSitting: evs is in id order.
func docketHeld(evs []*Event, ids []int64, win WindowIndex) (gaps []string, motions []Blocker) {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	// A bench that has never sat reads place 0, the cast's, which is no docket sitting's opening.
	opening := placesOf(evs, seq, win)[win.LatestSittingOf(benchSeat)]
	// A sitting that is not a docket sitting is no row's first docket sitting, so it holds no gaps.
	docket := benchOccasionSittings(evs, seq, win)[recordpb.Occasion_OCCASION_DOCKET]
	ds, _ := dispatchLedger(evs, seq, win)
	for _, d := range ds {
		if d.seat != benchSeat {
			continue
		}
		if r, sat := firstAfter(docket, d.at); !sat || r != opening {
			continue
		}
		for _, g := range d.gaps {
			if !slices.Contains(gaps, g) {
				gaps = append(gaps, g)
			}
		}
	}
	// A motion About a gap is a docket or a grade motion, and a grade motion's gavel is the chair's.
	for _, b := range unruledMotionBlockers(evs, ids, win) {
		if b.Owner == benchSeat && slices.Contains(gaps, b.About) {
			motions = append(motions, b)
		}
	}
	return gaps, motions
}

// heldMotionsSentence names each motion holding the sitting with the gap it is on.
func heldMotionsSentence(motions []Blocker) string {
	named := make([]string, 0, len(motions))
	for _, b := range motions {
		named = append(named, b.Subject+" (on "+b.About+")")
	}
	return strings.Join(named, ", ")
}

// BenchEnvelopeRefusal is the refusal of the bench's envelope while a docket motion stands unruled
// on a gap its current sitting is convened on (docketHeld), naming each one; "" when nothing holds
// the sitting, and for every other seat. The PreToolUse hook gives it to the seat as the denial of
// the call that returns its envelope, so the sitting cannot end until each is ruled.
func BenchEnvelopeRefusal(run Run, seatID string) (string, error) {
	if seatID != benchSeat {
		return "", nil
	}
	var refusal string
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		evs, win, err := eventsAt(q)
		if err != nil {
			return err
		}
		gaps, motions := docketHeld(evs, win.IDs(evs), win)
		if len(motions) == 0 {
			return nil
		}
		refusal = fmt.Sprintf("your envelope is refused: this sitting is convened on %s, and %d docket motion(s) on it stand unruled — %s. "+
			"The sitting cannot end while any stands: a motion you leave is one no later plan convenes you for. "+
			"Rule each with `motion docket rule`, then return your envelope",
			strings.Join(gaps, ", "), len(motions), heldMotionsSentence(motions))
		return nil
	})
	return refusal, err
}
