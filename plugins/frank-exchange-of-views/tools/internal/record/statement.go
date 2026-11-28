package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Statement is what a gap's raiser states its subject should become, the offer its author's accept
// (--accept G) takes: a pair (Location, FixNew) — the mint's prescription, or the pair a later
// closing of the raiser's states — or a retire of the anchor its closing states. A closing that
// states no text (--no-text) withdraws every earlier offer, and leaves Withdrawn.
type Statement struct {
	Location, FixNew string
	Retire           bool
	Withdrawn        bool
	// Wording and Grades are a closing's statement on a gap about a gap: the wording and grades the
	// gap it is about should carry (Closing.problem… and severity…); nil where it states neither.
	Wording *recordpb.Amend
	Grades  *recordpb.Regrade
	// AboutRef is what the gap is about where it is not a sentence: the anchor a retire takes.
	AboutRef string
}

// statementAt is gap g's standing statement on q's record: the raiser's latest closing that states
// one, else the mint's prescription; found false for an unknown gap.
func statementAt(q recordsql.Querier, g string) (Statement, bool, error) {
	loc, fix, found, err := prescriptionAt(q, g)
	if err != nil || !found {
		return Statement{}, found, err
	}
	var ref string
	if _, err := queryRowAt(q, []any{&ref}, `SELECT COALESCE("about_ref", '') FROM "mint" WHERE "gap_id" = ?`, g); err != nil {
		return Statement{}, true, err
	}
	st := Statement{Location: loc, FixNew: fix}
	evs, _, err := eventsAt(q)
	if err != nil {
		return st, true, err
	}
	raiser := ""
	for _, e := range evs {
		body, ok := recordpb.Body(e)
		if !ok {
			continue
		}
		switch b := body.(type) {
		case *recordpb.Mint:
			if b.GetGapId() == g {
				raiser = e.GetSeatId()
			}
		case *recordpb.Closing:
			if b.GetGapId() != g || e.GetSeatId() != raiser || raiser == "" {
				continue
			}
			switch {
			case b.GetNoText():
				st = Statement{Withdrawn: true}
			case b.GetRetire():
				st = Statement{Retire: true}
			case b.GetFixNew() != "":
				st = Statement{Location: b.GetLocation(), FixNew: b.GetFixNew()}
			case closingWording(b) || closingGrades(b):
				st = Statement{}
				if closingWording(b) {
					st.Wording = &recordpb.Amend{Problem: b.Problem, RequiredFix: b.RequiredFix, AcceptanceCheck: b.AcceptanceCheck}
				}
				if closingGrades(b) {
					st.Grades = &recordpb.Regrade{Severity: b.Severity, Likelihood: b.Likelihood, Impact: b.Impact, ComplexityCost: b.ComplexityCost}
				}
			}
		}
	}
	st.AboutRef = ref
	return st, true, nil
}

// StatementOf is gap g's standing statement on run's record (statementAt).
func StatementOf(run Run, g string) (st Statement, found bool, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		st, found, err = statementAt(q, g)
		return err
	})
	return st, found, err
}

// Statement is statementAt asked of the landing record.
func (l *Landing) Statement(g string) (Statement, bool, error) { return statementAt(l.w.q, g) }
