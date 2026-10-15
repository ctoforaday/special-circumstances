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
// judged at. A run with no record has no board and has not converged. A record that HAS events and
// no row at its newest one is an error, never "not converged": that miss would read exactly like a
// board the rule does not hold on, and admit the FAIL the refusal exists to stop.
func convergenceOf(w writeView) (Convergence, error) {
	var c Convergence
	db := w.q
	if noRecord(db) {
		return c, nil
	}
	var events int
	if _, err := queryRowAt(db, []any{&events}, `SELECT count(*) FROM "events"`); err != nil {
		return c, err
	}
	if events == 0 {
		return c, nil
	}
	found, err := queryRowAt(db, c.dest(), `SELECT `+convergenceColumns+` FROM "convergence_at"
	  WHERE "seq" = (SELECT MAX("id") FROM "events")`)
	if err != nil {
		return c, err
	}
	if !found {
		return c, fmt.Errorf("convergence_at has no row at this record's newest event, so whether the board has converged cannot be read: the view does not answer for the record it is over")
	}
	fraction, err := w.files.convergenceFraction()
	if err != nil {
		return c, err
	}
	c.atFraction(fraction)
	return c, nil
}

// requireFailIsNotConvergent refuses a FAIL over a board that has converged: nothing material is
// open, its mass is a small fraction of the run's peak, and this epoch minted nothing fresh and
// material. The report is not held open by gaps that change no reader decision.
//
// IT STANDS ONLY WHERE A PASS IS THE CHAIR'S TO RECORD. While a blocker another seat must clear
// holds the PASS — a petition the bench has not ruled, a contradiction no lens has raised, a
// citation or proof unchecked since its sentence changed — the refusal's "or issue `--as PASS`" is
// false, and refusing the FAIL too left the chair able to record no verdict at all (#1202). The
// chair's own items (the inquiry review, the stale-area spot-check) do not lift it: the refusal names them, and the PASS is the chair's once it has done them.
//
// THE REFUSAL NAMES ONLY THE CHAIR'S OWN ACTS, because the chair is the one seat that reads it. It
// told the chair to "raise something material — a material finding … minted as a gap", which is a
// lens's `finding` and `mint`; the chair has neither, and no route to ready a lens
// (plans/feov-lens-bar.md D6). What the chair CAN do with a defect that still stops it over a
// converged board is record no verdict and say why, so the run ends unverified on the record: in its
// position, which the bench reads in the transcript and the report renders under the epoch's
// no-verdict heading, or in the stale area's spot-check where the defect lies in one.
// TestTheConvergenceRefusalNamesOnlyTheChairsActs walks the rendered text against the chair's tree.
func requireFailIsNotConvergent(w writeView, blockers []Blocker) error {
	var own []string
	for _, b := range blockers {
		if !b.ChairOwned() {
			return nil
		}
		own = append(own, b.WorkItem())
	}
	c, err := convergenceOf(w)
	if err != nil {
		return err
	}
	if !c.Holds {
		return nil
	}
	pass := "Where nothing still stops you, issue `--as PASS`."
	if len(own) > 0 {
		pass = "Where nothing still stops you but your own items, issue `--as PASS` once they are done: " + strings.Join(own, "; ") + "."
	}
	return fmt.Errorf("FAIL refused — the board has converged: open mass %.1f is below %.0f%% of this run's peak %.1f, nothing open is material, and this epoch minted no fresh material gap. "+
		"A FAIL here holds the report open on gaps that change no reader decision. %s "+
		"Where a defect does, record no verdict and state the defect in your `position` — or, where it lies in a stale area, in that area's `spot-check`; the gaps that are not material stay open on the board",
		c.Mass, c.Fraction*100, c.Peak, pass)
}
