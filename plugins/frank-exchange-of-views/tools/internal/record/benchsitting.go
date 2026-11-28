package record

import (
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// docketHeld is what holds the bench's current docket sitting open: the docketing standing on each gap
// a bench dispatch row convening that sitting names (gblock's ruling, 2026-10-24; §III.B5). gaps is
// those gaps, in the order the dispatch rows name them; docketings is the docketed blockers
// (docketedBlockers) on them. Both are empty when the bench's latest sitting is not a docket sitting.
//
// ONE DOCKETING STANDS PER GAP, AND IT HOLDS THE ONE PROPOSAL THAT CAN STAND THERE. A docketing freezes
// its gap: no proposal is filed on it while it stands (L38), and its assignee's accept is refused (L27),
// so no assignee window is owed and none is waited for. The proposal a PROPOSAL docketing holds is its
// own motion_id, answered by the docketing's ruling.
//
// THE GAPS ARE THE ROWS THIS SITTING ANSWERS: each bench dispatch row convening a docket whose first
// docket sitting after it is this one (benchRowSitting's rule). A docketing on a gap no such row names
// is not this sitting's — it waits for the plan that readies the bench for it, and stays on the list as
// holding the outcome.
//
// The stream's positions stand in for events.id, as in owedSitting: evs is in id order.
func docketHeld(evs []*Event, ids []int64, win WindowIndex) (gaps []string, docketings []Blocker) {
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
	for _, b := range docketedBlockers(evs, ids) {
		if slices.Contains(gaps, b.About) {
			docketings = append(docketings, b)
		}
	}
	return gaps, docketings
}

// docketFirst is the docket's first: the standing docketing written earliest, "" where none stands.
// THE BENCH RULES THE RUN'S DOCKETINGS ONE AT A TIME, IN THE ORDER WRITTEN (#1695 §III.B4, the docket
// order; gblock 2026-11-26, "One run-wide queue"): each ruling sees the previous one's result, so a
// docketing behind the first waits for it — its ruling, a party's concession, or an amend ending it —
// and the first moves on within the sitting. The order is the rows' (docketingsAt), no field and no
// index.
func docketFirst(evs []*Event, ids []int64) string {
	for _, d := range docketingsAt(evs, ids) {
		if d.Standing() {
			return d.ID
		}
	}
	return ""
}

// benchHeld is what the bench's CURRENT sitting is convened to rule, by the occasion that sitting's
// register states — the bench's blocking items, and so what refuses its envelope (EnvelopeRefusal). It
// is keyed by what a ruling names: a docketing by its docketing id, a petition or a position by its
// motion id. Everything else the bench rules waits for the sitting convened for it, and is listed
// without blocking: the outcome refusal (requireBenchMotionsRuled) still holds the run on it.
//
//   - docket: the docket's first (docketFirst), where the sitting is convened on its gap (docketHeld).
//   - petition: every petition filed before the sitting opened. One petition sitting hears a filing
//     (benchOwedOf), so a petition this sitting leaves is one the bench is not convened for again. A
//     petition filed during the sitting waits for its own.
//   - position: every heard position filed before the sitting opened, on the same terms.
//   - terminal and assemble: the docket's first — the register dockets every proposal no docketing
//     holds at the docket's tail (L29), and the sitting rules the docket in order — and every petition
//     and position. The run's outcome is refused while any stands.
//   - a bench that has not sat, or a sitting with no occasion, holds none.
func benchHeld(evs []*Event, ids []int64, win WindowIndex) map[string]bool {
	held := benchConvened(evs, ids, win)
	first := docketFirst(evs, ids)
	for _, b := range docketedBlockers(evs, ids) {
		if b.Detail != first {
			delete(held, b.Detail)
		}
	}
	return held
}

// benchConvened is everything the bench's current sitting is convened on, by its occasion, before the
// docket order holds it to the docket's first: each standing docketing on a gap a docket sitting is
// convened on, every standing docketing at the terminal sitting and the assembly, and the motions
// benchHeld names.
func benchConvened(evs []*Event, ids []int64, win WindowIndex) map[string]bool {
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
		_, docketings := docketHeld(evs, ids, win)
		for _, b := range docketings {
			held[b.Detail] = true
		}
	case recordpb.Occasion_OCCASION_PETITION, recordpb.Occasion_OCCASION_POSITION:
		for _, b := range unruledMotionBlockers(evs, ids, win) {
			subj, _ := MotionSubjectEnum(b.motionSubject)
			if b.Owner == benchSeat && benchSittingFor[subj] == occasion && b.since < opening {
				held[b.Subject] = true
			}
		}
	case recordpb.Occasion_OCCASION_TERMINAL, recordpb.Occasion_OCCASION_ASSEMBLE:
		for _, b := range docketedBlockers(evs, ids) {
			held[b.Detail] = true
		}
		for _, b := range unruledMotionBlockers(evs, ids, win) {
			if b.Owner == benchSeat {
				held[b.Subject] = true
			}
		}
	}
	return held
}
