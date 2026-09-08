package record

import (
	"fmt"
)

// Convergence is the corrected never-hard-fail predicate (plans/roundless.md §III.B.2.1), read
// off the record at a FAIL: the board has nothing material on it and red is still failing.
//
// The view's `divergent` was wrong for this purpose on three counts, and the refusal uses the
// corrected predicate: fresh MATERIAL mints in this epoch (a run minting one fresh trifle per
// sitting must trip it); max severity STRICTLY below material (GRADE_MEDIUM is material and would
// otherwise force a PASS over a live gap); the CURRENT severity, which a regrade can move; and the
// mass bound as a fraction of this run's PEAK board mass rather than a magic number.
type Convergence struct {
	Mass, Peak, MaxSeverityMass float64
	FreshMaterialMints          int
	Fraction                    float64
	Holds                       bool
}

func convergenceOf(run Run) (Convergence, error) {
	var c Convergence
	p, err := RunParams(run)
	if err != nil {
		return c, err
	}
	c.Fraction = p.ConvergenceFraction
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return c, err
	}
	// The open board now: its mass and its top severity, on CURRENT grades.
	if _, err := queryRow(run, []any{&c.Mass, &c.MaxSeverityMass}, `
	  SELECT COALESCE(SUM(COALESCE(gl."mass", 0.0) * COALESCE(gi."mass", 0.0)), 0.0),
	         COALESCE(MAX(COALESCE(gs."mass", 0.0)), 0.0)
	  FROM "gap" g
	  LEFT JOIN "enum_grade" gl ON gl."value" = g."current_likelihood"
	  LEFT JOIN "enum_grade" gi ON gi."value" = g."current_impact"
	  LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
	  WHERE g."open"`); err != nil {
		return c, err
	}
	// The peak: the largest open-board mass at any gate so far, or now.
	var peakAtGates float64
	if _, err := queryRow(run, []any{&peakAtGates},
		`SELECT COALESCE(MAX("mass"), 0.0) FROM "convergence_vs_verdict"`); err != nil {
		return c, err
	}
	c.Peak = peakAtGates
	if c.Mass > c.Peak {
		c.Peak = c.Mass
	}
	// Fresh material mints in the CURRENT epoch: minted this epoch, superseding nothing, material now.
	if _, err := queryRow(run, []any{&c.FreshMaterialMints}, `
	  SELECT count(*) FROM "gap" g LEFT JOIN "enum_grade" gs ON gs."value" = g."current_severity"
	  WHERE g."supersedes_count" = 0 AND COALESCE(gs."mass", 0.0) >= ?
	    AND g."minted_epoch" = (SELECT count(*) FROM "events" WHERE "type" = 'register' AND "seat_id" = 'red-chair')`,
		material); err != nil {
		return c, err
	}
	c.Holds = c.Peak > 0 && c.Mass < c.Fraction*c.Peak && c.MaxSeverityMass < material && c.FreshMaterialMints == 0
	return c, nil
}

// requireFailIsNotConvergent refuses a FAIL over a board that has converged: nothing material is
// open, its mass is a small fraction of the run's peak, and this epoch minted nothing fresh and
// material. Red must raise something material, or PASS — the report is not held open by trifles.
func requireFailIsNotConvergent(run Run) error {
	c, err := convergenceOf(run)
	if err != nil {
		return err
	}
	if !c.Holds {
		return nil
	}
	return fmt.Errorf("record: verdict FAIL refused — the board has converged: open mass %.1f is below %.0f%% of this run's peak %.1f, nothing open is material (top severity mass %.1f, material is %.1f), and this epoch minted no fresh material gap. "+
		"A FAIL here holds the report open on trifles. Raise something material — a finding graded medium or above, minted as a gap — or issue `--as PASS`; the sub-material gaps stay on the board and the report lists them as not certified against",
		c.Mass, c.Fraction*100, c.Peak, c.MaxSeverityMass, material)
}
