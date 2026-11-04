package recordsql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// stampRegradeResult writes onto r the gap's grades once r stands — the ONE statement of the regrade
// overlay. On each axis: r's grade where r names one, else the grade the gap held before r, which is
// the result stored on the gap's latest earlier regrade, or the mint's grade where there is none. An
// axis nothing ever graded stays absent.
//
// IT RUNS IN InsertTx, beside the sitting and epoch stamps, so every write of a regrade carries it —
// a seat's act, a migration's replay and a seeded fixture alike — and nothing reads a regrade row
// without its result. Each earlier regrade carries its own, so the grade before r is one row, not a
// fold over the gap's history. Both inputs are fixed once r is written — the record is append-only —
// so the stamped grades are the gap's grades from r until its next regrade, and gap_grade reads them
// as they are.
//
// WHATEVER RESULT A REGRADE ARRIVES WITH IS OVERWRITTEN, never kept: a kept value is a grade set
// nothing re-derived, stored with nothing to stop it being stale or wrong. A seat cannot supply one —
// the write path refuses it first (record.requireNoRegradeResult) — and a row copied from a record
// that already stamped it (a fixture seeded from an archived run's events) is stamped again here, to
// the same grades when it lands in the same order. A migration's replay is stamped the same way.
func stampRegradeResult(tx *sql.Tx, r *recordpb.Regrade) error {
	r.ResultSeverity, r.ResultLikelihood, r.ResultImpact, r.ResultComplexityCost = nil, nil, nil, nil
	if r.GetGapId() == "" {
		return nil // a gap_id is required, and the required-field refusal names --id
	}
	var before [4]sql.NullString
	err := tx.QueryRow(`SELECT "result_severity", "result_likelihood", "result_impact", "result_complexity_cost"
	  FROM "regrade" WHERE "gap_id" = ? ORDER BY "event_id" DESC LIMIT 1`, r.GetGapId()).
		Scan(&before[0], &before[1], &before[2], &before[3])
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRow(`SELECT "severity", "likelihood", "impact", "complexity_cost"
		  FROM "mint" WHERE "gap_id" = ?`, r.GetGapId()).
			Scan(&before[0], &before[1], &before[2], &before[3])
		if errors.Is(err, sql.ErrNoRows) {
			return nil // the regrade's foreign key refuses a gap no mint created
		}
	}
	if err != nil {
		return fmt.Errorf("reading gap %s's grades before this regrade: %w", r.GetGapId(), err)
	}
	named := [4]*recordpb.Grade{r.Severity, r.Likelihood, r.Impact, r.ComplexityCost}
	result := [4]**recordpb.Grade{&r.ResultSeverity, &r.ResultLikelihood, &r.ResultImpact, &r.ResultComplexityCost}
	for k := range named {
		if named[k] != nil {
			*result[k] = named[k].Enum()
			continue
		}
		if !before[k].Valid {
			continue
		}
		vd, ok := recordpb.BySpelling(recordpb.Grade(0).Descriptor(), before[k].String)
		if !ok {
			return fmt.Errorf("gap %s held the grade %q before this regrade, which is not a grade", r.GetGapId(), before[k].String)
		}
		*result[k] = recordpb.Grade(vd.Number()).Enum()
	}
	return nil
}

// HasRegradeResult answers whether r carries a result grade on any axis.
func HasRegradeResult(r *recordpb.Regrade) bool {
	return r.ResultSeverity != nil || r.ResultLikelihood != nil || r.ResultImpact != nil || r.ResultComplexityCost != nil
}
