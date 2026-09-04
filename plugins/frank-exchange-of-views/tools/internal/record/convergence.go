package record

import "fmt"

// ConvergenceRound is one round's answer from the convergence_vs_verdict view.
//
// The inputs travel with the verdict deliberately. A bare count of divergent rounds is the shape
// that failed before — `0` with nothing behind it reads as a clean board — so a reader gets the
// mass, the top severity and the fresh-mint count that produced the answer and can check it.
type ConvergenceRound struct {
	Round           int
	Verdict         string
	Mass            float64
	MaxSeverityMass float64
	FreshMints      int
	Divergent       bool
}

// ConvergenceVsVerdict asks the record which rounds diverged.
//
// IT RETURNS AN ERROR RATHER THAN AN EMPTY SLICE when it cannot ask, and the distinction is the
// whole point of this function existing. The metric it serves spent seven runs reporting 0 because
// a map key was missing — absence rendered as a measurement. Here a run with no verdicts yields
// no rows (a real zero: nothing has been adjudicated), and a run that cannot be read yields an
// error the caller must say something about.
func ConvergenceVsVerdict(run Run) ([]ConvergenceRound, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}
	rows, err := db.Query(`SELECT "round", "verdict", "mass", "max_severity_mass", "fresh_mints", "divergent" ` +
		`FROM "convergence_vs_verdict" ORDER BY "round"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for convergence-vs-verdict: %w", err)
	}
	defer rows.Close()
	var out []ConvergenceRound
	for rows.Next() {
		var c ConvergenceRound
		if err := rows.Scan(&c.Round, &c.Verdict, &c.Mass, &c.MaxSeverityMass, &c.FreshMints, &c.Divergent); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
