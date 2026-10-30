package record

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// RECONSIDERATION IS ON NEW GROUNDS, CITED AND VERIFIED (gblock, 2026-10-30).
//
// A docket motion on a gap the bench has ruled closed asks the bench to rule it again. Nothing
// refused one, so a party that lost could put the same question back before the bench every sitting
// and spend a ruling each time on what the first already settled. The motion names its new grounds
// the way every act names what it is about — the id of an act the seats already read, or a quote of
// the report as it now reads — and the tool checks that the grounds are NEW (recorded after the
// ruling) and about THIS gap, refusing otherwise with which of the two failed.
//
// WHAT COUNTS. An id: a finding about the gap (`--about-kind gap`), a gap minted about it or
// superseding it, a proof answering it, or a citation placed in text an edit answering it wrote —
// each recorded after the ruling. A quote: report text an edit answering the gap wrote after the
// ruling, which the report still carries (the verb's placement guard asks that of the report the
// motion lands on). A ruling marked final still admits new grounds: final is the bench's forecast,
// and new grounds are a fact the bench then rules.
//
// A REMAND CLOSES NOTHING, so a gap whose latest ruling is a remand is back in the debate and needs
// no grounds to come before the bench again: the remand is the new entry, and the dispatch's own
// docket after the remand's exchange is that.

// closedBy is the latest docket ruling on a gap, where it closed the gap.
type closedBy struct {
	at                  int64
	motion, disposition string
	settled, reopensOn  string
	final               bool
}

// latestDocketRulingSQL is a gap's latest docket ruling.
const latestDocketRulingSQL = `SELECT r."event_id", mr."motion_id", r."disposition", r."settled", COALESCE(r."reopens_on", ''), COALESCE(r."final", 0)
     FROM "motion_rule_docket" r JOIN "motion_rule" mr ON mr."event_id" = r."event_id"
     JOIN "motion" mo ON mo."motion_id" = mr."motion_id" JOIN "motion_docket" md ON md."event_id" = mo."event_id"
     WHERE md."gap_id" = ? ORDER BY r."event_id" DESC LIMIT 1`

// closingRulingOn is gap's latest docket ruling where it closed the gap; ok is false where the bench
// has not ruled the gap, or its latest ruling remanded it.
func closingRulingOn(q recordsql.Querier, gap string) (closedBy, bool, error) {
	var r closedBy
	var final int64
	found, err := queryRowAt(q, []any{&r.at, &r.motion, &r.disposition, &r.settled, &r.reopensOn, &final}, latestDocketRulingSQL, gap)
	if err != nil || !found || !slices.Contains(ClosingDispositions(), r.disposition) {
		return closedBy{}, false, err
	}
	r.final = final != 0
	return r, true, nil
}

// words is the ruling as a reconsideration's refusal states it.
func (r closedBy) words(gap string) string {
	reopen := "it is final"
	if !r.final {
		reopen = fmt.Sprintf("it reopens on: %q", r.reopensOn)
	}
	return fmt.Sprintf("the bench ruled gap %s %s on motion %s (event %d), settling %q, and %s", gap, r.disposition, r.motion, r.at, r.settled, reopen)
}

// GroundsKinds is what --grounds takes, in the words a page and a refusal state it.
const GroundsKinds = "the id of a finding about the gap, a gap minted about it or superseding it, a proof answering it, or a citation placed in text an edit answering it wrote"

// requireNewGrounds is the docket arm's check of a reconsideration's grounds.
func requireNewGrounds(q recordsql.Querier, d *recordpb.DocketMotion) error {
	gap := d.GetGapId()
	if gap == "" {
		return nil
	}
	ruling, ruled, err := closingRulingOn(q, gap)
	if err != nil {
		return err
	}
	cited := d.Grounds != nil || d.GroundsQuote != nil
	switch {
	case !ruled && cited:
		return feov.Errorf(feov.Validation,
			"the bench has not ruled gap %s closed, so a docket motion on it reconsiders nothing and takes no new grounds: leave --grounds and --quote off", gap)
	case !ruled:
		return nil
	case !cited:
		return feov.Errorf(feov.Validation,
			"%s. A docket motion on it asks the bench to rule it again, and is admitted on NEW GROUNDS only: pass --grounds with %s, recorded after that ruling — or --quote with report text such an edit wrote after it. Nothing was recorded",
			ruling.words(gap), GroundsKinds)
	case d.Grounds != nil && d.GroundsQuote != nil:
		return feov.Errorf(feov.Validation, "--grounds and --quote each locate the new grounds; pass one")
	case d.GroundsQuote != nil:
		at, err := editWroteAfter(q, gap, ruling.at, d.GetGroundsQuote())
		if err != nil {
			return err
		}
		if at == 0 {
			return feov.Errorf(feov.Validation,
				"--quote %q is not new grounds on gap %s: no edit answering the gap after the ruling (event %d) wrote it, so it is text the ruling already had before it. %s. Nothing was recorded",
				d.GetGroundsQuote(), gap, ruling.at, ruling.words(gap))
		}
		return nil
	}
	return requireGroundsAct(q, gap, ruling, d.GetGrounds())
}

// groundsActSQL is the act that placed id, the earliest first: its place, its kind, whether it is about
// gap ?2, and a citation's location.
const groundsActSQL = `SELECT e, kind, about, loc FROM (
       SELECT "event_id" e, 'finding' kind, ("about_kind" = 'gap' AND "about_ref" = ?2) about, '' loc FROM "finding" WHERE "id" = ?1
       UNION ALL SELECT m."event_id", 'gap', ((m."about_kind" = 'gap' AND m."about_ref" = ?2)
           OR EXISTS (SELECT 1 FROM "mint_supersedes" s WHERE s."event_id" = m."event_id" AND s."value" = ?2)), '' FROM "mint" m WHERE m."gap_id" = ?1
       UNION ALL SELECT "event_id", 'proof', ("answers" = ?2), '' FROM "proof" WHERE "proof_id" = ?1
       UNION ALL SELECT "event_id", 'citation', 0, COALESCE("location", '') FROM "cite" WHERE "label" = ?1)
     ORDER BY e LIMIT 1`

// requireGroundsAct checks an id named as grounds: an act on the record placed it, after the ruling,
// about the gap.
func requireGroundsAct(q recordsql.Querier, gap string, ruling closedBy, id string) error {
	var at int64
	var kind, loc string
	var about sql.NullBool
	found, err := queryRowAt(q, []any{&at, &kind, &about, &loc}, groundsActSQL, id, gap)
	if err != nil {
		return err
	}
	if !found {
		return feov.Errorf(feov.Validation,
			"--grounds names %s, which no act on this record placed: it takes %s — the id inside its %s, %s, %s or %s token in the report or on the board. Nothing was recorded",
			id, GroundsKinds, anchor.Token("F-…"), anchor.Token("G-…"), anchor.Token("P-…"), anchor.Token("C-…"))
	}
	if at <= ruling.at {
		return feov.Errorf(feov.Validation,
			"--grounds names %s %s, recorded at event %d — before the ruling it would reopen (event %d), so the ruling already had it. %s. Nothing was recorded",
			kind, id, at, ruling.at, ruling.words(gap))
	}
	concerns := about.Valid && about.Bool
	if kind == "citation" {
		edit, err := editWroteAfter(q, gap, ruling.at, loc)
		if err != nil {
			return err
		}
		concerns = edit != 0
	}
	if !concerns {
		return feov.Errorf(feov.Validation,
			"--grounds names %s %s, which is not about gap %s: the grounds are %s. Nothing was recorded",
			kind, id, gap, GroundsKinds)
	}
	return nil
}

// editWroteAfter is the place of the latest edit after the ruling at `after` that answers gap and
// wrote quote, 0 where none did.
func editWroteAfter(q recordsql.Querier, gap string, after int64, quote string) (int64, error) {
	if strings.TrimSpace(quote) == "" {
		return 0, nil
	}
	rows, err := q.Query(`SELECT "event_id", COALESCE("new", '') FROM "blue_edit" WHERE "answers" = ? AND "event_id" > ? ORDER BY "event_id" DESC`, gap, after)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var at int64
		var wrote string
		if err := rows.Scan(&at, &wrote); err != nil {
			return 0, err
		}
		if s, _ := anchortext.LocateSpan(wrote, quote); s >= 0 {
			return at, nil
		}
	}
	return 0, rows.Err()
}
