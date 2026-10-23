package record

import (
	"fmt"
	"math"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// GRADE MOVEMENT IN TOTAL GOES BEFORE THE BENCH (gblock, 2026-10-20).
//
// Each regrade is the minting lens's own act, and each grade motion is ruled one at a time by the
// chair. Nothing reviewed the movement IN TOTAL, and the total is what games the gate: a gap regraded
// below medium stops holding a PASS, and gaps regraded up keep a converged board out of the FAIL
// refusal. The round-loop engine docketed the cumulative magnitude of a round's accepted grade
// changes past 2 mass units for the judge (debate.js ACCEPTED_DELTA_DOCKET_THRESHOLD, removed with the
// round loop in ac2be0d1, and nothing replaced it). Here the record computes it.
//
// WHAT COUNTS. Every regrade, not only one that follows an accepted grade motion: on the archived
// m28, m29 and m30 runs no grade motion was filed and every grade that moved was a lens's own
// regrade, so a guard over motions alone measured nothing. Both directions, as the old guard did:
// deflation clears a PASS, inflation escapes the convergence refusal. The axes the gate reads —
// severity (materiality) and likelihood and impact (mass) — each weighed |mass(new) - mass(old)| on
// the Grade enum's mass annotation; complexity_cost is optional on a mint and no gate reads it.
// Summed per axis per regrade, so a grade moved down and back up is two movements, as two accepted
// changes were.
//
// THE WINDOW. Movement on an open gap the bench has not ruled on: a regrade is ruled on once a docket
// ruling on its gap lands after it, because that ruling decided the gap's fate with its grades in
// view. Cumulative across sittings — the old per-round window let a change be sliced across rounds.
// A closed gap's grade holds nothing at the gate, so its movement drops out.

// GradeMovementThreshold is the total grade movement, in mass, that the bench rules on: past it
// (strictly), every open gap carrying movement the bench has not ruled on is docketed for it. The
// round-loop engine's value, in the same unit — the mass table is the one Grade annotation.
const GradeMovementThreshold = 2.0

// GradeMovementRule is the rule as every page that needs it states it — the regrade a lens files,
// the dispatch that dockets, the docket ruling that rules on it.
var GradeMovementRule = fmt.Sprintf("GRADE MOVEMENT IN TOTAL GOES BEFORE THE BENCH. Each axis a regrade moves — severity, likelihood or impact — counts the difference in mass between its two grades, up or down. "+
	"Once the open gaps' movement that no docket ruling on its gap has followed totals more than %.1f, the dispatch dockets every open gap carrying such movement for the bench, whatever its grade, and PASS is refused until the bench rules each: that ruling rules on the gap's movement", GradeMovementThreshold)

// gradeMove is one axis one regrade moved.
type gradeMove struct {
	at       int64 // the regrade's place
	seat     string
	axis     string
	from, to string
	mass     float64 // |mass(to) - mass(from)|
}

func (m gradeMove) words() string {
	return fmt.Sprintf("%s %s → %s at event %d by %s", m.axis, wordOrUngraded(m.from), m.to, m.at, m.seat)
}

func wordOrUngraded(w string) string {
	if w == "" {
		return "ungraded"
	}
	return w
}

// gradeMovement is the open board's grade movement the bench has not ruled on.
type gradeMovement struct {
	// total is every open gap's movement, an unruled docket motion's gap included: that movement is
	// not yet ruled on either, and it is the board's.
	total float64
	// moves is each open gap's movement the bench has not ruled on, in the record's order.
	moves map[string][]gradeMove
	// order is the gaps of moves by their first movement.
	order []string
	// pending is the open gaps a docket motion stands unruled on: their ruling rules on the movement.
	pending map[string]bool
}

// gradeMovementOf folds the stream. open is the open gaps; ids is evs's places, aligned with it.
func gradeMovementOf(evs []*Event, ids []int64, win WindowIndex, open map[string]bool) gradeMovement {
	out := gradeMovement{moves: map[string][]gradeMove{}, pending: map[string]bool{}}
	ruled := map[string]int64{} // gap -> the place of its latest docket ruling
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != motionSubjectWord(subjectDocket) || m.GapID == "" {
			continue
		}
		if !m.Ruled() {
			out.pending[m.GapID] = open[m.GapID]
			continue
		}
		ruled[m.GapID] = max(ruled[m.GapID], m.ruled)
	}
	axes := []recordpb.GradeDimension{recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY, recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD, recordpb.GradeDimension_GRADE_DIMENSION_IMPACT}
	grades := map[string][3]string{}
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			if _, seen := grades[b.GetGapId()]; !seen {
				grades[b.GetGapId()] = [3]string{recordpb.Word(b.GetSeverity()), recordpb.Word(b.GetLikelihood()), recordpb.Word(b.GetImpact())}
			}
		case *recordpb.Regrade:
			g := b.GetGapId()
			cur := grades[g]
			for k, w := range []string{recordpb.Word(b.GetSeverity()), recordpb.Word(b.GetLikelihood()), recordpb.Word(b.GetImpact())} {
				if w == "" || w == cur[k] {
					continue
				}
				mv := gradeMove{at: ids[i], seat: e.GetSeatId(), axis: recordpb.Word(axes[k]), from: cur[k], to: w, mass: math.Abs(MASS[w] - MASS[cur[k]])}
				cur[k] = w
				if !open[g] || ruled[g] > mv.at || mv.mass == 0 {
					continue
				}
				if len(out.moves[g]) == 0 {
					out.order = append(out.order, g)
				}
				out.moves[g] = append(out.moves[g], mv)
				out.total += mv.mass
			}
			grades[g] = cur
		}
	}
	return out
}

// openSet is the ids of the gate list's open gaps.
func openSet(gaps []blockerGap) map[string]bool {
	out := make(map[string]bool, len(gaps))
	for _, g := range gaps {
		out[g.id] = true
	}
	return out
}

// past reports whether the total is past the threshold the bench rules at.
func (m gradeMovement) past() bool { return m.total > GradeMovementThreshold }

// docketed reports whether the dispatch dockets gap g for its movement: the total is past the
// threshold, g carries movement the bench has not ruled on, and no docket motion on g stands — one
// that does is ruled with the movement in view.
func (m gradeMovement) docketed(g string) bool {
	return m.past() && len(m.moves[g]) > 0 && !m.pending[g]
}

// gapMass is g's movement the bench has not ruled on.
func (m gradeMovement) gapMass(g string) float64 {
	s := 0.0
	for _, mv := range m.moves[g] {
		s += mv.mass
	}
	return s
}

// blockers is the gate list's grade-movement arm: one per gap the dispatch dockets, the bench's.
func (m gradeMovement) blockers() []Blocker {
	var out []Blocker
	for _, g := range m.order {
		if !m.docketed(g) {
			continue
		}
		var said []string
		for _, mv := range m.moves[g] {
			said = append(said, mv.words())
		}
		out = append(out, Blocker{Kind: BlockerGradeMovement, Subject: g, Owner: benchSeat, Detail: strings.Join(said, "; "),
			moved: m.gapMass(g), totalMoved: m.total})
	}
	return out
}

// gradeMovementFact is the sentence every wording of a grade-movement blocker states: the gap's
// movement the bench has not ruled on, each axis as the record holds it, and the board's total
// against the threshold.
func gradeMovementFact(b Blocker) string {
	return fmt.Sprintf("gap %s's grades moved %.1f mass the bench has not ruled on (%s), and the open gaps' grade movement the bench has not ruled on totals %.1f, past the %.1f at which the bench rules on it",
		b.Subject, b.moved, b.Detail, b.totalMoved, GradeMovementThreshold)
}

// gradeMovementBasis is the docket motion the dispatch files for a gap's movement: the ask, in
// the words the bench rules on.
func gradeMovementBasis(b Blocker) string {
	return gradeMovementFact(b) + ". The bench owes a ruling on whether the gap's grades as they now stand are the ones the leaf supports; its ruling on this motion rules on that movement"
}
