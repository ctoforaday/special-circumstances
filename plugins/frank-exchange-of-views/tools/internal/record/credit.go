package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// requireCreditAnswers refuses a credit that would not answer a contradiction its seat raised
// (gblock, 2026-10-24: the lens answers the contradiction by naming the open gap, and no duplicate
// is minted). Four refusals, each a fact the record holds:
//
//   - the gap is not open: a closed gap answers nothing now, and what it closed on was not this
//     contradiction;
//   - the finding is another seat's: the credit is the raiser's answer to its own raise, as the
//     answering mint is the raiser's (requireMintWithinBudget's exemption);
//   - the finding raised no contradiction;
//   - the contradiction it raised is already answered, by a gap minted from the finding or credited
//     with it — the refusal names that gap.
//
// The contradiction is read from the fold every other reader reads (contradictionsOf), on the
// write's own transaction, so a raise answered by another write a moment earlier is refused here.
func requireCreditAnswers(q recordsql.Querier, seatID string, b *recordpb.Credit) error {
	gapID, finding := b.GetGapId(), b.GetFoundBy()
	if err := requireGap(q, gapID, "--id"); err != nil {
		return err
	}
	if err := requireFindings(q, []string{finding}, "--found-by"); err != nil {
		return err
	}
	var open bool
	if _, err := queryRowAt(q, []any{&open}, `SELECT "open" FROM "gap" WHERE "gap_id" = ?`, gapID); err != nil {
		return err
	}
	if !open {
		return feov.Errorf(feov.Validation,
			"%s is closed, so it answers nothing now: a contradiction is answered by an OPEN gap that is its defect, or by the gap you mint from the finding that raised it", gapID)
	}
	var filer string
	if _, err := queryRowAt(q, []any{&filer},
		`SELECT e."seat_id" FROM "finding" f JOIN "events" e ON e."id" = f."event_id" WHERE f."id" = ?`, finding); err != nil {
		return err
	}
	if filer != seatID {
		return feov.Errorf(feov.Validation,
			"finding %s is %s's, not yours: a lens credits a gap with its OWN finding that raised a contradiction; the contradiction is that finding's raiser's to answer", finding, filer)
	}
	evs, win, err := eventsOfAt(q, contradictionActs...)
	if err != nil {
		return err
	}
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.raisedBy != finding {
			continue
		}
		if c.answeredBy != "" {
			return feov.Errorf(feov.Validation,
				"contradiction %s, which finding %s raised, is already answered, by gap %s: one gap answers it", c.idWord(), finding, c.answeredBy)
		}
		return nil
	}
	return feov.Errorf(feov.Validation,
		"finding %s raised no contradiction: a finding raises one by naming its id (`finding --about-kind contradiction --about <id>`), and the credit answers what it raised", finding)
}

// CreditAnswers is the contradiction a credit answers, by its id: the one its finding raised. It is
// read after the write, for the verb's result line; "" where the record holds no such raise.
func CreditAnswers(run Run, finding string) (string, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return "", err
	}
	evs, win, err := eventsOfAt(db, contradictionActs...)
	if err != nil {
		return "", err
	}
	for _, c := range contradictionsOf(evs, win.IDs(evs)) {
		if c.raisedBy == finding {
			return c.idWord(), nil
		}
	}
	return "", nil
}

// creditAct is the crediting act as a refusal names it to the seat that would take it.
func creditAct(gapID, finding string) string {
	return fmt.Sprintf("`credit --id %s --found-by %s`", gapID, finding)
}
