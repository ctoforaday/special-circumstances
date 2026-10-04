package record

import (
	"fmt"
	"strings"
)

// Convergence is the never-hard-fail predicate (plans/roundless.md §III.B.2.1) over the board as it
// stood at one point of the record: nothing material open, the open mass below the run's fraction of
// its peak board mass, and no fresh MATERIAL mint in the point's epoch (a run minting one fresh
// trifle per sitting must still trip it). MaxSeverityMass is reported beside it and decides nothing.
//
// ONE DEFINITION, TWO READERS. The quantities are the convergence_at view's row for the point; the
// rule over them is atFraction. The FAIL refusal reads the row at the record's newest event, where
// the board as it stood is the board now; the scorecard's detector (ConvergenceVsVerdict) reads the
// row at each verdict afterwards. So the detector flags a FAIL exactly where the rule held when it
// was issued — a FAIL the refusal stood aside for, because another seat held the PASS or a
// migration replayed it — and a later regrade does not move a past verdict's answer.
type Convergence struct {
	Mass, Peak, MaxSeverityMass float64
	MaterialOpen                int
	FreshMaterialMints          int
	Fraction                    float64
	Holds                       bool
}

// convergenceColumns are convergence_at's quantities in the order dest scans them.
const convergenceColumns = `"mass", "peak", "max_severity_mass", "material_open", "fresh_material_mints"`

func (c *Convergence) dest() []any {
	return []any{&c.Mass, &c.Peak, &c.MaxSeverityMass, &c.MaterialOpen, &c.FreshMaterialMints}
}

// atFraction applies the rule at the run's fraction.
func (c *Convergence) atFraction(fraction float64) {
	c.Fraction = fraction
	c.Holds = c.Peak > 0 && c.Mass < fraction*c.Peak && c.MaterialOpen == 0 && c.FreshMaterialMints == 0
}

// convergenceOf is the board at the record's newest event — the point a verdict being written is
// judged at.
func convergenceOf(run Run) (Convergence, error) {
	var c Convergence
	p, err := RunParams(run)
	if err != nil {
		return c, err
	}
	if _, err := queryRow(run, c.dest(), `SELECT `+convergenceColumns+` FROM "convergence_at"
	  WHERE "seq" = (SELECT MAX("id") FROM "events")`); err != nil {
		return c, err
	}
	c.atFraction(p.ConvergenceFraction)
	return c, nil
}

// requireFailIsNotConvergent refuses a FAIL over a board that has converged: nothing material is
// open, its mass is a small fraction of the run's peak, and this epoch minted nothing fresh and
// material. Red must raise something material, or PASS — the report is not held open by gaps that
// change no reader decision.
//
// IT STANDS ONLY WHERE A PASS IS THE CHAIR'S TO RECORD. While a blocker another seat must clear
// holds the PASS — a petition the bench has not ruled, a contradiction no lens has raised — the
// refusal's "or issue `--as PASS`" is false, and refusing the FAIL too left the chair able to record
// no verdict at all (#1202). The chair's own items (the avenue review, the stale-area spot-check)
// do not lift it: the refusal names them, and the PASS is the chair's once it has done them.
func requireFailIsNotConvergent(run Run, blockers []Blocker) error {
	var own []string
	for _, b := range blockers {
		if !b.ChairOwned() {
			return nil
		}
		own = append(own, b.WorkItem())
	}
	c, err := convergenceOf(run)
	if err != nil {
		return err
	}
	if !c.Holds {
		return nil
	}
	pass := "issue `--as PASS`"
	if len(own) > 0 {
		pass = "issue `--as PASS` once your own items are done: " + strings.Join(own, "; ")
	}
	return fmt.Errorf("record: verdict FAIL refused — the board has converged: open mass %.1f is below %.0f%% of this run's peak %.1f, nothing open is material, and this epoch minted no fresh material gap. "+
		"A FAIL here holds the report open on gaps that change no reader decision. Raise something material — a material finding, by its class or else graded medium or above, minted as a gap — or %s; the gaps that are not material stay open on the board",
		c.Mass, c.Fraction*100, c.Peak, pass)
}
