package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EvidenceDemand is a gap's live-answer demand as the close verb states it: the anchor kind owed
// (the EvidenceTables row whose check is the gap's check kind; "" for a gap no row answers, which
// owes no demand), whether the gap view's evidence_answered holds, and the evidence anchors the
// demand reaches (the demand_reach view) — the anchor the gap is about — whose retire, or the cut of
// whose sentence, or whose check again by the minter, would meet it.
type EvidenceDemand struct {
	CheckKind, AnswerKind string
	// AnswerNoun and AnswerVerb are the owed kind's row's noun and placing verb.
	AnswerNoun, AnswerVerb string
	Answered               bool
	// Awaiting is the gap view's awaiting_evidence: the gap is open, and its demand owed and unmet.
	Awaiting bool
	// Reach is the evidence anchor the demand reaches, read only while the demand is unmet: the one
	// case a refusal names it.
	Reach []Reach
}

// Reach is one evidence anchor a gap's demand reaches: the anchor the gap is about, checked by its
// minter Seat.
type Reach struct {
	Anchor, Seat string
}

// Unmet reports whether the demand is owed and nothing meets it.
func (d EvidenceDemand) Unmet() bool { return d.AnswerKind != "" && !d.Answered }

// EvidenceDemandOf reads gap gapID's demand off the mint, the gap view and the demand's reach. A
// gap not on the record answers the zero demand; a read error, or a check kind this binary does not
// declare, is returned, never folded into "no demand".
func EvidenceDemandOf(run Run, gapID string) (EvidenceDemand, error) {
	if gapID == "" {
		return EvidenceDemand{}, nil
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return EvidenceDemand{}, err
	}
	return evidenceDemandAt(db, gapID)
}

// evidenceDemandAt is EvidenceDemandOf asked of q — on the write path the writing transaction, so a
// retire, or a cut, of the gap's only answer and the close racing it cannot both land. The gap view is read
// once: the demand, whether it is met, and whether the gap awaits it.
func evidenceDemandAt(q recordsql.Querier, gapID string) (EvidenceDemand, error) {
	var d EvidenceDemand
	if gapID == "" || noRecord(q) {
		return d, nil
	}
	found, err := queryRowAt(q, []any{&d.CheckKind, &d.Answered, &d.Awaiting},
		`SELECT m."check_kind", COALESCE(g."evidence_answered", 0), COALESCE(g."awaiting_evidence", 0)
		 FROM "mint" m JOIN "gap" g ON g."gap_id" = m."gap_id"
		 WHERE m."gap_id" = ?`, gapID)
	if err != nil || !found {
		return EvidenceDemand{}, err
	}
	// A CHECK KIND THIS BINARY DOES NOT DECLARE IS NO DEMAND IT CAN STATE. AnswerKindOf answers ""
	// for it, which reads exactly like a `document` gap that owes nothing, so the miss is an error.
	if _, ok := CheckKindOf(d.CheckKind); !ok {
		return EvidenceDemand{}, fmt.Errorf("gap %s was minted --check-kind %q, a check kind this binary does not declare, so its live-answer demand cannot be read", gapID, d.CheckKind)
	}
	d.AnswerKind = AnswerKindOf(d.CheckKind)
	if row, ok := recordsql.EvidenceTableOf(d.AnswerKind); ok {
		d.AnswerNoun, d.AnswerVerb = row.Noun, row.Verb
	}
	if !d.Unmet() {
		return d, nil
	}
	rows, err := q.Query(`SELECT "anchor", "seat" FROM "demand_reach"
		 WHERE "gap_id" = ? AND "anchor" IS NOT NULL`, gapID)
	if err != nil {
		return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r Reach
		if err := rows.Scan(&r.Anchor, &r.Seat); err != nil {
			return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
		}
		// A gap about an anchor of a kind that backs nothing reaches no evidence anchor.
		if !anchor.Backs(r.Anchor) {
			continue
		}
		d.Reach = append(d.Reach, r)
	}
	if err := rows.Err(); err != nil {
		return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
	}
	return d, nil
}

// requireDemandMet refuses the close of gap gapID while its live-answer demand is unmet, asked in the
// write's transaction (judgeAct). The verb asks the same question before its write for the order of
// its refusals; this is the answer that holds, because a retire, or a cut, of the gap's only
// answer can commit between the verb's read and this write, and evidence_answered then reads false.
func requireDemandMet(q recordsql.Querier, gapID string) error {
	d, err := evidenceDemandAt(q, gapID)
	if err != nil || !d.Unmet() {
		return err
	}
	return DemandRefusal(gapID, d)
}

// DemandRefusal is the close's refusal of gap gapID, whose demand d nothing meets: the demand in its
// one text, then what meets it on this gap — blue's anchor placed as its answer, a retire of each
// evidence anchor the demand reaches, or a cut of its sentence, and the check again by the seat whose check it is — and the
// bench, for a demand the lens no longer holds.
func DemandRefusal(gapID string, d EvidenceDemand) error {
	ways := []string{fmt.Sprintf("blue's `%s --answers %s`, a %s placed on the sentence as its answer (after a reword, on the reworded sentence)", d.AnswerVerb, gapID, d.AnswerNoun)}
	var retire, again []string
	for _, r := range d.Reach {
		retire = append(retire, fmt.Sprintf("blue's `retire --anchor %s`, or an edit that cuts its sentence, of the %s it is about", r.Anchor, anchor.Kind(r.Anchor)))
		// THE CHECK AGAIN, BY ITS OWN SEAT, SINCE THE MINT, of an anchor of the demanded kind: a reading
		// that backs under `source`, a `sound` judgement under `computation`. Another seat's check meets
		// nothing.
		switch {
		case anchor.Kind(r.Anchor) != d.AnswerKind:
		case d.AnswerKind == "computation":
			again = append(again, fmt.Sprintf("a `sound` judgement of %s by %s, made since the gap was minted", r.Anchor, r.Seat))
		default:
			again = append(again, fmt.Sprintf("a reading of %s by %s, made since the gap was minted, that reads the work and backs the sentence", r.Anchor, r.Seat))
		}
	}
	if len(retire) > 0 {
		ways = append(ways, strings.Join(retire, ", or "))
	}
	ways = append(ways, again...)
	return feov.Errorf(feov.Conflict, "gap %s was minted --check-kind %s, and nothing meets its live-answer demand while its subject stands in the report, so this close is refused under every disposition — `moot` and `not_a_defect` included. Nothing is recorded. "+
		"The demand: %s. What meets it on this gap: %s; another seat's check meets nothing. "+
		"A demand you no longer hold is not yours to end: another seat proposes ending the gap in a docket motion, or the dispatch puts it before the bench at impasse, and the bench's closing ruling, which is not held to the demand, disposes of it; a regrade keeps it open",
		gapID, d.CheckKind, recordpb.DemandClause, strings.Join(ways, "; or "))
}
