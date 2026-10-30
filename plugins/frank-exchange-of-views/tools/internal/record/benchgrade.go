package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE BENCH SETS A GRADE BY FILING A REGRADE (gblock, 2026-10-30).
//
// A docket ruling decides a gap's fate with its grades in view, and the bench had no act that set
// one: a ruling that held "this is high, not medium" left the minting lens free to regrade it to
// medium in its next sitting. The bench files an ordinary `regrade`, through the fold every reader
// of a grade already reads, while a docket motion on the gap stands before it unruled — so the
// regrade precedes the ruling that rules on it, and the grade-movement window (gradeMovementOf)
// closes over it as it closes over every regrade a ruling follows.
//
// ONE PREDICATE, THREE REFUSALS. A bench regrade that moves an axis the bench does not hold LOCKS
// it; one that moves an axis it holds LIFTS the lock, and the axis stands at the bench's new grade
// for the lens to move from there (gblock, 2026-10-30: "Lock returns to lens"). benchSetAxesOf folds
// that toggle, and while an axis is locked:
//
//   - a lens's regrade that MOVES it is refused; one that keeps it is the lens saying its gap stands,
//     and is admitted;
//   - a grade motion on it is refused at filing: the chair could accept it, and the lens would then
//     owe a regrade it cannot file;
//   - a mint superseding the gap whose grade differs on it is refused: the lock follows the lineage,
//     so a successor carries the bench's grade and the lens cannot escape it by re-minting.
//
// Only the bench moves it while it stands, at a docket motion on the gap standing unruled; that move
// releases it to the lens, and a later bench move locks it again. A bench regrade that moves nothing
// is refused: keeping a grade is the lens's act (keptgrade.go), and the bench sets a grade by moving
// it.

// gradeAxes is the four axes in the order the folds below index them.
var gradeAxes = [4]recordpb.GradeDimension{
	recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY,
	recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD,
	recordpb.GradeDimension_GRADE_DIMENSION_IMPACT,
	recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY,
}

// axisOf is gradeAxes' index of d, -1 for none.
func axisOf(d recordpb.GradeDimension) int {
	for k, a := range gradeAxes {
		if a == d {
			return k
		}
	}
	return -1
}

// regradeAxes is r's grades in gradeAxes' order, nil where it names none.
func regradeAxes(r *recordpb.Regrade) [4]*recordpb.Grade {
	return [4]*recordpb.Grade{r.Severity, r.Likelihood, r.Impact, r.ComplexityCost}
}

// mintAxes is m's grades in gradeAxes' order.
func mintAxes(m *recordpb.Mint) [4]recordpb.Grade {
	return [4]recordpb.Grade{m.GetSeverity(), m.GetLikelihood(), m.GetImpact(), m.GetComplexityCost()}
}

// axisGrade is where one axis of one gap stands: its grade, and — where the bench locked it, or the
// gap inherits the lock from an ancestor — the place of the bench regrade that locked it.
type axisGrade struct {
	grade recordpb.Grade
	// benchAt is the place of the bench regrade that locked the axis (an ancestor's, for a successor
	// that inherits it), 0 while it is unlocked.
	benchAt int64
}

// benchSetAxesOf folds evs into every gap's four axes. ids is evs's places, aligned with it.
func benchSetAxesOf(evs []*Event, ids []int64) map[string][4]axisGrade {
	out := map[string][4]axisGrade{}
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Mint:
			if _, seen := out[b.GetGapId()]; seen {
				continue
			}
			var cur [4]axisGrade
			for k, g := range mintAxes(b) {
				cur[k].grade = g
			}
			// THE LOCK FOLLOWS THE LINEAGE: an axis the bench holds locked on an ancestor is locked on its
			// successor, which the write admits only at the ancestor's grade.
			for _, a := range b.GetSupersedes() {
				for k, ax := range out[a] {
					if ax.benchAt != 0 && cur[k].benchAt == 0 {
						cur[k].benchAt = ax.benchAt
					}
				}
			}
			out[b.GetGapId()] = cur
		case *recordpb.Regrade:
			g := b.GetGapId()
			cur := out[g]
			bench := seatclass.RoleOfSeat(e.GetSeatId()) == "bench"
			for k, w := range regradeAxes(b) {
				if w == nil || *w == cur[k].grade {
					continue
				}
				cur[k].grade = *w
				// THE BENCH'S MOVE TOGGLES THE LOCK: it locks an unlocked axis and lifts its own lock,
				// releasing the axis to the lens at the new grade. Any other seat's move leaves it
				// unlocked — the write refuses one on a locked axis.
				if bench && cur[k].benchAt == 0 {
					cur[k].benchAt = ids[i]
				} else {
					cur[k].benchAt = 0
				}
			}
			out[g] = cur
		}
	}
	return out
}

// benchSetAxesAt is benchSetAxesOf over the record q's write is landing on.
func benchSetAxesAt(w writeView) (map[string][4]axisGrade, error) {
	evs, win, err := eventsAt(w.q)
	if err != nil {
		return nil, err
	}
	return benchSetAxesOf(evs, win.IDs(evs)), nil
}

// benchLockWords is the refusal's account of one locked axis.
func benchLockWords(gap string, k int, ax axisGrade) string {
	return fmt.Sprintf("gap %s's %s is %s, which the bench set at event %d", gap, recordpb.Word(gradeAxes[k]), recordpb.Word(ax.grade), ax.benchAt)
}

// requireRegradeWithinBenchLock is the regrade arm of the lock. A lens's regrade may move no axis the
// bench holds locked. The bench's regrade needs a docket motion on the gap standing unruled before it, and
// must move an axis.
func requireRegradeWithinBenchLock(w writeView, seatID string, r *recordpb.Regrade) error {
	if r.GetGapId() == "" {
		return nil
	}
	axes, err := benchSetAxesAt(w)
	if err != nil {
		return err
	}
	cur := axes[r.GetGapId()]
	if seatclass.RoleOfSeat(seatID) == "bench" {
		motion, err := unruledDocketMotionOn(w, r.GetGapId())
		if err != nil {
			return err
		}
		if motion == "" {
			return feov.Errorf(feov.Conflict,
				"no docket motion on gap %s stands before you unruled: the bench sets a grade while it hears the gap, and its ruling then rules on that grade. Nothing was recorded", r.GetGapId())
		}
		for k, g := range regradeAxes(r) {
			if g != nil && *g != cur[k].grade {
				return nil
			}
		}
		return feov.Errorf(feov.Validation,
			"this regrade moves no axis of gap %s: the bench sets a grade by moving it, and keeping a grade is the minting lens's act. Name the axis you set at the grade you rule it carries. Nothing was recorded", r.GetGapId())
	}
	for k, g := range regradeAxes(r) {
		if g == nil || *g == cur[k].grade || cur[k].benchAt == 0 {
			continue
		}
		return feov.Errorf(feov.Conflict,
			"%s, and only the bench moves it while it stands; the bench moving it again releases it to you. A regrade that keeps it — naming it at %s, or naming no grade — is yours to file. Nothing was recorded",
			benchLockWords(r.GetGapId(), k, cur[k]), recordpb.Word(cur[k].grade))
	}
	return nil
}

// requireGradeMotionOffBenchLock refuses a grade motion on an axis the bench holds locked: the chair could
// accept it, and the minting lens would owe a regrade the lock refuses.
func requireGradeMotionOffBenchLock(w writeView, m *recordpb.GradeMotion) error {
	k := axisOf(m.GetDimension())
	if k < 0 || m.GetGapId() == "" {
		return nil
	}
	axes, err := benchSetAxesAt(w)
	if err != nil {
		return err
	}
	ax := axes[m.GetGapId()][k]
	if ax.benchAt == 0 {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"%s, and a grade motion on it asks the chair for a grade no lens may then file. A seat that disputes it files `motion docket file` on the gap, which puts it before the bench — on new grounds, where the bench has ruled the gap closed. Nothing was recorded",
		benchLockWords(m.GetGapId(), k, ax))
}

// requireSuccessorKeepsBenchGrades refuses a mint superseding a gap whose grade differs from the
// ancestor's on an axis the bench holds locked: the successor carries the bench's grade.
func requireSuccessorKeepsBenchGrades(w writeView, m *recordpb.Mint) error {
	if len(m.GetSupersedes()) == 0 {
		return nil
	}
	axes, err := benchSetAxesAt(w)
	if err != nil {
		return err
	}
	mine := mintAxes(m)
	for _, a := range m.GetSupersedes() {
		for k, ax := range axes[a] {
			if ax.benchAt == 0 || mine[k] == ax.grade {
				continue
			}
			return feov.Errorf(feov.Conflict,
				"--supersedes names %s: %s, and a successor carries the bench's grade, so mint it at --%s %s. Nothing was recorded",
				a, benchLockWords(a, k, ax), recordpb.Word(gradeAxes[k]), recordpb.Word(ax.grade))
		}
	}
	return nil
}

// unruledDocketMotionSQL is a docket motion on a gap that no ruling answers, the earliest first.
const unruledDocketMotionSQL = `SELECT mo."motion_id" FROM "motion" mo JOIN "motion_docket" md ON md."event_id" = mo."event_id"
     WHERE md."gap_id" = ? AND NOT EXISTS (SELECT 1 FROM "motion_rule" r WHERE r."motion_id" = mo."motion_id")
     ORDER BY mo."event_id" LIMIT 1`

// unruledDocketMotionOn is a docket motion on gap standing unruled, "" where none does.
func unruledDocketMotionOn(w writeView, gap string) (string, error) {
	var id string
	if _, err := queryRowAt(w.q, []any{&id}, unruledDocketMotionSQL, gap); err != nil {
		return "", err
	}
	return id, nil
}
