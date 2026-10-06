package record

import "fmt"

// VerdictConvergence is one verdict's answer from the convergence_at view: the board as it stood when
// the verdict was issued, judged by the same rule the FAIL refusal applies at the write path.
//
// The inputs travel with the verdict deliberately. A bare count of divergent epochs is the shape
// that failed before — `0` with nothing behind it reads as a clean board — so a reader gets the
// mass, the peak, the top severity and the fresh-mint count that produced the answer and can check
// it.
type VerdictConvergence struct {
	Epoch   int
	Verdict string
	Convergence
	// Divergent is a FAIL issued over a board the convergence rule holds on.
	Divergent bool
}

// ConvergenceVsVerdict asks the record which verdicts diverged, each on the board as it stood at it
// and against the run's own fraction.
//
// IT RETURNS AN ERROR RATHER THAN AN EMPTY SLICE when it cannot ask, and the distinction is the
// whole point of this function existing. The metric it serves spent seven runs reporting 0 because
// a map key was missing — absence rendered as a measurement. Here a run with no verdicts yields
// no rows (a real zero: nothing has been adjudicated), and a run that cannot be read yields an
// error the caller must say something about. The run's terms are read after the record, and only
// the fraction: a run with no record has nothing to judge whatever its config says, and a term the
// rule does not use cannot refuse it.
func ConvergenceVsVerdict(run Run) ([]VerdictConvergence, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}
	fraction, err := runConvergenceFraction(run)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "epoch", "verdict", ` + convergenceColumns +
		` FROM "convergence_at" WHERE "verdict" IS NOT NULL ORDER BY "seq"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for convergence-vs-verdict: %w", err)
	}
	defer rows.Close()
	var out []VerdictConvergence
	for rows.Next() {
		var c VerdictConvergence
		if err := rows.Scan(append([]any{&c.Epoch, &c.Verdict}, c.Convergence.dest()...)...); err != nil {
			return nil, err
		}
		c.atFraction(fraction)
		c.Divergent = c.Verdict == "fail" && c.Holds
		out = append(out, c)
	}
	return out, rows.Err()
}
