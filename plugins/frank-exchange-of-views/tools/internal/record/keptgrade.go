package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A REGRADE THAT KEEPS THE GRADE IS THE LENS SAYING ITS GAP STANDS (gblock, 2026-10-18).
//
// A lens that re-audits its gap after blue answered it, and finds that its own acceptance check
// still fails, has one thing to say: the gap stands, and why. The work list asked it to say so and
// named no act, so the lens that held a gap with a careful reason recorded a null turn, and the
// reason reached neither blue nor the bench (universe-m28 sitting 381, m29 442).
//
// THE ACT IS `regrade`, and the record already reads it this way. A regrade that names no grade —
// or names only the grades the gap already has — moves nothing: the exchange fold counts it as an
// act and not as movement, so it stalls the gap toward impasse exactly as silence would, and an
// accepted grade motion it follows is answered on the record, in its basis (regradesAfforded).
// What it adds is the reason, where the two seats that act on it read:
//
//   - blue, on its work list beside the gap, until blue answers the gap again (keptGradeItems);
//   - the bench, in the debate transcript it reads before it rules (KeptGrades, which debate.md, the
//     debate view and the assembled transcript all render from).
//
// A KEEP IS JUDGED AGAINST THE GRADES AT ITS PLACE: the mint's, overlaid by every earlier regrade of
// the gap, on all four axes. A regrade that moves complexity alone moved a grade, so it is not a
// keep, though the exchange fold does not count it as movement either.
type gradeKeep struct {
	gap, seat, reason string
	at                int64  // the regrade's events.id
	ev                *Event // the regrade
}

// gradeKeepsOf is every regrade on evs that keeps the grade, in record order.
func gradeKeepsOf(evs []*Event, win WindowIndex) []gradeKeep {
	grades := map[string][4]recordpb.Grade{}
	var out []gradeKeep
	for _, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			if _, seen := grades[b.GetGapId()]; !seen {
				grades[b.GetGapId()] = [4]recordpb.Grade{b.GetSeverity(), b.GetLikelihood(), b.GetImpact(), b.GetComplexityCost()}
			}
		case *recordpb.Regrade:
			cur := grades[b.GetGapId()]
			moved := false
			for k, g := range []*recordpb.Grade{b.Severity, b.Likelihood, b.Impact, b.ComplexityCost} {
				if g != nil && *g != cur[k] {
					cur[k], moved = *g, true
				}
			}
			grades[b.GetGapId()] = cur
			if !moved {
				out = append(out, gradeKeep{gap: b.GetGapId(), seat: e.GetSeatId(), reason: b.GetBasis(), at: win.Of(e).ID, ev: e})
			}
		}
	}
	return out
}

// KeptGrades is the set of regrades on evs that keep the grade — the one reading every renderer of
// the debate transcript asks, so a keep reads as one in all of them. evs must hold the gaps' mints
// and every regrade before the ones asked about.
func KeptGrades(evs []*Event, win WindowIndex) map[*Event]bool {
	out := map[*Event]bool{}
	for _, k := range gradeKeepsOf(evs, win) {
		out[k.ev] = true
	}
	return out
}

// KeptGradeHeading is the transcript's heading for a regrade that keeps the grade, in every
// renderer of it.
func KeptGradeHeading(e *Event) string {
	r := e.GetRegrade()
	return fmt.Sprintf("### RED KEEPS THE GRADE — %s (%s)", r.GetGapId(), e.GetSeatId())
}

// keptGradeItems is blue's item for each open gap of gaps whose minting lens kept its grade after
// blue last answered it (blueAnswersOf): the lens's reason, which is what its check still finds
// failing. A keep blue has answered since is answered, and drops off.
func keptGradeItems(evs []*Event, win WindowIndex, gaps []WorkGapState) []string {
	open := map[string]bool{}
	for _, g := range gaps {
		if g.Open {
			open[g.ID] = true
		}
	}
	answers := blueAnswersOf(evs, win, open, motionsAt(evs, win.IDs(evs), win))
	latest := map[string]gradeKeep{}
	var order []string
	for _, k := range gradeKeepsOf(evs, win) {
		if !open[k.gap] || k.at <= answers[k.gap].at {
			continue
		}
		if _, seen := latest[k.gap]; !seen {
			order = append(order, k.gap)
		}
		latest[k.gap] = k
	}
	var out []string
	for _, g := range order {
		k := latest[g]
		out = append(out, fmt.Sprintf("gap %s: its lens re-audited it and KEEPS ITS GRADE — what its check still finds failing: %s", g, k.reason))
	}
	return out
}
