package record

import (
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// refused while either stands (benchHeld, EnvelopeRefusal).
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

// benchHeld is the unruled motions whose gavel is the bench's that its CURRENT sitting is convened
// to rule, by the occasion that sitting's register states — the bench's blocking items, and so what
// refuses its envelope (EnvelopeRefusal). Every other unruled bench motion waits for the sitting
// convened for it, and is listed without blocking: the outcome refusal (requireBenchMotionsRuled)
// still holds the run on it.
//
//   - docket: every docket motion on a gap the sitting is convened on (docketHeld).
//   - petition: every petition filed before the sitting opened. One petition sitting hears a filing
//     (benchReadiness), so a petition this sitting leaves is one no later plan convenes it for. A
//     petition filed during the sitting waits for its own.
//   - terminal and assemble: every one. The run's outcome is refused while any stands.
//   - a bench that has not sat, or a sitting with no occasion, holds none.
func benchHeld(evs []*Event, ids []int64, win WindowIndex) map[string]bool {
	held := map[string]bool{}
	opening := win.LatestSittingOf(benchSeat)
	if opening == 0 {
		return held
	}
	occasion := recordpb.Occasion_OCCASION_UNSPECIFIED
	for o, at := range benchOccasionSittings(evs, ids, win) {
		if slices.Contains(at, opening) {
			occasion = o
		}
	}
	switch occasion {
	case recordpb.Occasion_OCCASION_DOCKET:
		_, motions := docketHeld(evs, ids, win)
		for _, b := range motions {
			held[b.Subject] = true
		}
	case recordpb.Occasion_OCCASION_PETITION:
		for _, b := range unruledMotionBlockers(evs, ids, win) {
			subj, _ := MotionSubjectEnum(b.motionSubject)
			if b.Owner == benchSeat && subj == recordpb.MotionSubject_MOTION_SUBJECT_PETITION && b.since < opening {
				held[b.Subject] = true
			}
		}
	case recordpb.Occasion_OCCASION_TERMINAL, recordpb.Occasion_OCCASION_ASSEMBLE:
		for _, b := range unruledMotionBlockers(evs, ids, win) {
			if b.Owner == benchSeat {
				held[b.Subject] = true
			}
		}
	}
	return held
}
