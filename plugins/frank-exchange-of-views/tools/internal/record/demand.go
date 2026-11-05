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
// owes no demand), whether the gap view's evidence_answered holds, and the evidence anchors the gap
// reaches — the one it is about, and each read by a reading it carries — whose retire, or whose
// check again by the seat whose check the gap carries, would meet it.
type EvidenceDemand struct {
	CheckKind, AnswerKind string
	// AnswerNoun and AnswerVerb are the owed kind's row's noun and placing verb.
	AnswerNoun, AnswerVerb string
	Answered               bool
	// About is the evidence anchor the gap was minted about, and MintedBy the seat that minted it;
	// About is "" for a gap about anything else.
	About, MintedBy string
	// Readings are the readings the gap carries (the carrying_reading view, the one reader of
	// Verify.carried_by), each its citation and its reader, in record order.
	Readings []CarryingReading
}

// CarryingReading is one reading a gap carries: the citation it read, and the seat that read it.
type CarryingReading struct{ Anchor, Seat string }

// Unmet reports whether the demand is owed and nothing meets it.
func (d EvidenceDemand) Unmet() bool { return d.AnswerKind != "" && !d.Answered }

// EvidenceDemandOf reads gap gapID's demand off the mint, the gap view and the carrying readings. A
// gap not on the record answers the zero demand; a read error is returned, never folded into "no
// demand".
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
// retire of the gap's only answer and the close racing it cannot both land.
func evidenceDemandAt(q recordsql.Querier, gapID string) (EvidenceDemand, error) {
	var d EvidenceDemand
	if gapID == "" || noRecord(q) {
		return d, nil
	}
	var aboutKind, aboutRef string
	found, err := queryRowAt(q, []any{&d.CheckKind, &d.Answered, &aboutKind, &aboutRef, &d.MintedBy},
		`SELECT COALESCE(m."check_kind", ''), COALESCE(g."evidence_answered", 0), COALESCE(m."about_kind", ''), COALESCE(m."about_ref", ''), COALESCE(e."seat_id", '')
		 FROM "mint" m JOIN "gap" g ON g."gap_id" = m."gap_id" JOIN "events" e ON e."id" = m."event_id"
		 WHERE m."gap_id" = ?`, gapID)
	if err != nil || !found {
		return EvidenceDemand{}, err
	}
	d.AnswerKind = AnswerKindOf(d.CheckKind)
	if row, ok := recordsql.EvidenceTableOf(d.AnswerKind); ok {
		d.AnswerNoun, d.AnswerVerb = row.Noun, row.Verb
	}
	if aboutKind == recordpb.Word(recordpb.AboutKind_ABOUT_KIND_ANCHOR) && anchor.Backs(aboutRef) {
		d.About = aboutRef
	}
	rows, err := q.Query(`SELECT "anchor", "seat_id" FROM "carrying_reading" WHERE "gap_id" = ? AND "anchor" IS NOT NULL ORDER BY "event_id"`, gapID)
	if err != nil {
		return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var c CarryingReading
		if err := rows.Scan(&c.Anchor, &c.Seat); err != nil {
			return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
		}
		d.Readings = append(d.Readings, c)
	}
	if err := rows.Err(); err != nil {
		return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
	}
	return d, nil
}

// requireDemandMet refuses the close of gap gapID while its live-answer demand is unmet, asked in the
// write's transaction (judgeAct). The verb asks the same question before its write for the order of
// its refusals; this is the answer that holds, because a retire of the gap's only answer can commit
// between the verb's read and this write, and evidence_answered then reads false.
func requireDemandMet(q recordsql.Querier, gapID string) error {
	d, err := evidenceDemandAt(q, gapID)
	if err != nil || !d.Unmet() {
		return err
	}
	return DemandRefusal(gapID, d)
}

// DemandRefusal is the close's refusal of gap gapID, whose demand d nothing meets: the demand in its
// one text, then what meets it on this gap — blue's anchor placed as its answer, a retire of each
// evidence anchor the gap reaches, and the check again by the seat whose check the gap carries — and
// the bench, for a demand the lens no longer holds.
func DemandRefusal(gapID string, d EvidenceDemand) error {
	ways := []string{fmt.Sprintf("blue's `%s --answers %s`, a %s placed on the sentence as its answer (after a reword, on the reworded sentence)", d.AnswerVerb, gapID, d.AnswerNoun)}
	var retire []string
	if d.About != "" {
		retire = append(retire, fmt.Sprintf("blue's `retire --anchor %s`, the %s it is about", d.About, anchor.Kind(d.About)))
	}
	for _, r := range d.Readings {
		retire = append(retire, fmt.Sprintf("blue's `retire --anchor %s`, the citation whose reading by %s it carries", r.Anchor, r.Seat))
	}
	if len(retire) > 0 {
		ways = append(ways, strings.Join(retire, ", or "))
	}
	// THE CHECK AGAIN, BY ITS OWN SEAT, SINCE THE MINT, of an anchor of the demanded kind: a reading
	// that backs under `source`, a `sound` judgement under `computation`. Another seat's check meets
	// nothing.
	again := func(id, seat string) string {
		if anchor.Kind(id) != d.AnswerKind {
			return ""
		}
		if d.AnswerKind == "computation" {
			return fmt.Sprintf("a `sound` judgement of %s by %s, made since the gap was minted", id, seat)
		}
		return fmt.Sprintf("a reading of %s by %s, made since the gap was minted, that reads the work and backs the sentence", id, seat)
	}
	if d.About != "" {
		if a := again(d.About, d.MintedBy); a != "" {
			ways = append(ways, a)
		}
	}
	for _, r := range d.Readings {
		if a := again(r.Anchor, r.Seat); a != "" {
			ways = append(ways, a)
		}
	}
	return feov.Errorf(feov.Conflict, "gap %s was minted --check-kind %s, and nothing meets its live-answer demand, so this close is refused under every disposition — `moot`, `not_a_defect` and a close to a successor included. Nothing is recorded. "+
		"The demand: %s. What meets it on this gap: %s; another seat's check meets nothing. "+
		"A demand you no longer hold is not closed: file `motion docket file` on the gap, and the bench's closing ruling, which is not held to the demand, disposes of it; a regrade keeps it open",
		gapID, d.CheckKind, recordpb.DemandClause, strings.Join(ways, "; or "))
}
