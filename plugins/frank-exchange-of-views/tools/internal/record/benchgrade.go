package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A RULING ON A GRADE LOCKS THE AXES IT DECIDED (gblock 2026-10-30; #1695 L35′, "A ruling re-locks",
// 2026-11-21).
//
// A docket ruling decides a gap's grades where its question is a grade: a regrade proposal's ruling,
// either side, decides the axes the proposal names, and a grade-movement docketing's ruling — --keep or
// --restore — every axis the window moved. The ruling locks each axis it decided at the grade its
// transaction leaves: the stored or the restored grade where the tool wrote the bench's regrade before
// it, the grade standing where it wrote none. The bench writes no grade of its own; its regrade is the
// tool's, in a ruling's transaction, and sets no lock of its own — benchSetAxesOf folds the regrade and
// the ruling after it as one step. While an axis is locked:
//
//   - a lens's regrade that MOVES it is refused; one that keeps it is the lens saying its gap stands,
//     and is admitted;
//   - a mint superseding the gap whose grade differs on it is refused: the lock follows the lineage,
//     so a successor carries the ruled grade and the lens cannot escape it by re-minting.
//
// No move lifts a lock. A later ruling on the axis re-locks it at its own grade; the lock ends with the
// ruling's tombstone (§III.B14), whose end rows nothing writes yet, so a ruled grade holds for the run.

// gradeAxes is the four axes in the order the folds below index them.
var gradeAxes = [4]recordpb.GradeDimension{
	recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY,
	recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD,
	recordpb.GradeDimension_GRADE_DIMENSION_IMPACT,
	recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY,
}

// regradeAxes is r's grades in gradeAxes' order, nil where it names none.
func regradeAxes(r *recordpb.Regrade) [4]*recordpb.Grade {
	return [4]*recordpb.Grade{r.Severity, r.Likelihood, r.Impact, r.ComplexityCost}
}

// regradeResultAxes is the grades r stored as the gap's once it stands, in gradeAxes' order; the
// zero where the gap is graded on no such axis.
func regradeResultAxes(r *recordpb.Regrade) [4]recordpb.Grade {
	return [4]recordpb.Grade{r.GetResultSeverity(), r.GetResultLikelihood(), r.GetResultImpact(), r.GetResultComplexityCost()}
}

// mintAxes is m's grades in gradeAxes' order.
func mintAxes(m *recordpb.Mint) [4]recordpb.Grade {
	return [4]recordpb.Grade{m.GetSeverity(), m.GetLikelihood(), m.GetImpact(), m.GetComplexityCost()}
}

// axisGrade is where one axis of one gap stands: its grade, and — where a ruling locked it, or the gap
// inherits the lock from an ancestor — the place of the ruling that locked it.
type axisGrade struct {
	grade recordpb.Grade
	// benchAt is the place of the ruling that locked the axis (an ancestor's, for a successor that
	// inherits it), 0 while it is unlocked.
	benchAt int64
}

// benchSetAxesOf folds evs into every gap's four axes. ids is evs's places, aligned with it.
func benchSetAxesOf(evs []*Event, ids []int64) map[string][4]axisGrade {
	out := map[string][4]axisGrade{}
	// THE AXES A RULING DECIDES. A grade-movement docketing's ruling decides every axis the window
	// moved — a party's regrade that moved it since the axis's last ruling (gradeMovementOf's window,
	// TestTheLockedAxesAreTheWindowsMovedAxes holds the two level); a regrade proposal's ruling the
	// axes the proposal names.
	moved := map[string][4]bool{}
	named := map[string][4]bool{} // a regrade proposal's motion id -> the axes it names
	type docketed struct {
		gap, motion string
		cause       recordpb.DocketingCause
	}
	docketings := map[string]docketed{}
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
			// THE LOCK FOLLOWS THE LINEAGE: an axis a ruling holds locked on an ancestor is locked on its
			// successor, which the write admits only at the ancestor's grade.
			if a := b.GetSupersedes(); a != "" {
				for k, ax := range out[a] {
					if ax.benchAt != 0 && cur[k].benchAt == 0 {
						cur[k].benchAt = ax.benchAt
					}
				}
			}
			out[b.GetGapId()] = cur
		case *recordpb.Motion:
			if r := b.GetRegrade(); r != nil {
				var n [4]bool
				for k, g := range regradeAxes(r) {
					n[k] = g != nil
				}
				named[b.GetMotionId()] = n
			}
		case *recordpb.Docketing:
			docketings[b.GetDocketingId()] = docketed{gap: b.GetGapId(), motion: b.GetMotionId(), cause: b.GetCause()}
		case *recordpb.Regrade:
			g := b.GetGapId()
			cur := out[g]
			bench := seatclass.RoleOfSeat(e.GetSeatId()) == "bench"
			// THE GRADES ARE THE ONES THE REGRADE STORED (recordsql.stampRegradeResult, the one
			// statement of the overlay); this fold re-derives none. An axis whose stored grade differs
			// from the one before is an axis this regrade moved. The bench's regrade is a ruling's
			// statement, and the ruling after it locks; a party's move is movement in the window, and on
			// a locked axis — which the write refuses — leaves it unlocked.
			m := moved[g]
			for k, now := range regradeResultAxes(b) {
				if now == cur[k].grade {
					continue
				}
				cur[k].grade = now
				if !bench {
					cur[k].benchAt = 0
					m[k] = k < 3
				}
			}
			moved[g] = m
			out[g] = cur
		case *recordpb.MotionRule:
			r := b.GetDocket()
			k, ok := docketings[r.GetDocketingId()]
			if r == nil || !ok || r.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED {
				continue
			}
			var decided [4]bool
			switch {
			case k.cause == recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT:
				decided = moved[k.gap]
			case k.motion != "":
				decided = named[k.motion]
			}
			cur := out[k.gap]
			m := moved[k.gap]
			for ax, d := range decided {
				if d {
					cur[ax].benchAt = ids[i]
					m[ax] = false
				}
			}
			moved[k.gap] = m
			out[k.gap] = cur
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
	return fmt.Sprintf("gap %s's %s is %s, which the bench's ruling at event %d decided", gap, recordpb.Word(gradeAxes[k]), recordpb.Word(ax.grade), ax.benchAt)
}

// requireRegradeWithinBenchLock is the regrade arm of the lock. A lens's regrade may move no axis a
// ruling holds locked. The bench's regrade is a ruling's statement, written by the tool before the
// ruling: it needs a docketing standing on the gap, unruled, and must move an axis; it may move a locked
// axis, which the ruling after it re-locks at its own grade.
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
		docketing, _, err := standingOnGap(w.q, r.GetGapId())
		if err != nil {
			return err
		}
		if docketing == "" {
			return feov.Errorf(feov.Conflict,
				"no docketing on gap %s stands unruled: the bench's regrade is the statement a ruling applies, written in that ruling's transaction. Nothing was recorded", r.GetGapId())
		}
		for k, g := range regradeAxes(r) {
			if g != nil && *g != cur[k].grade {
				return nil
			}
		}
		return feov.Errorf(feov.Validation,
			"this regrade moves no axis of gap %s: the grades it states are the ones the gap holds, so no ruling applies it — the bench answers for the other side. Nothing was recorded", r.GetGapId())
	}
	for k, g := range regradeAxes(r) {
		if g == nil || *g == cur[k].grade || cur[k].benchAt == 0 {
			continue
		}
		return feov.Errorf(feov.Conflict,
			"%s, and a ruled grade is not moved back. A regrade that keeps it — naming it at %s, or naming no grade — is yours to file. Nothing was recorded",
			benchLockWords(r.GetGapId(), k, cur[k]), recordpb.Word(cur[k].grade))
	}
	return nil
}

// requireSuccessorKeepsBenchGrades refuses a mint superseding a gap whose grade differs from the
// ancestor's on an axis a ruling holds locked: the successor carries the ruled grade.
func requireSuccessorKeepsBenchGrades(w writeView, m *recordpb.Mint) error {
	a := m.GetSupersedes()
	if a == "" {
		return nil
	}
	axes, err := benchSetAxesAt(w)
	if err != nil {
		return err
	}
	mine := mintAxes(m)
	for k, ax := range axes[a] {
		if ax.benchAt == 0 || mine[k] == ax.grade {
			continue
		}
		return feov.Errorf(feov.Conflict,
			"--supersedes names %s: %s, and a successor carries the ruled grade, so mint it at --%s %s. Nothing was recorded",
			a, benchLockWords(a, k, ax), recordpb.Word(gradeAxes[k]), recordpb.Word(ax.grade))
	}
	return nil
}

// requireNoRegradeResult refuses a regrade that arrives with result grades. The insert stamps them
// from the gap's grades and the axes the regrade names (recordsql.stampRegradeResult); a writer that
// could set them could give a gap any grade without naming the axis it moved.
func requireNoRegradeResult(r *recordpb.Regrade) error {
	if recordsql.HasRegradeResult(r) {
		return fmt.Errorf("a regrade's result grades are stamped by the tool from the gap's grades and the axes the regrade names, never supplied by a seat — regrade without them")
	}
	return nil
}
