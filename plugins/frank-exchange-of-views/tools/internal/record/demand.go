package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EvidenceDemand is a gap's live-answer demand as the close verb states it: the anchor kind owed
// (the EvidenceTables row whose check is the gap's check kind; "" for a gap no row answers, which
// owes no demand), whether the gap view's evidence_answered holds, and the evidence anchors the
// demand reaches (the demand_reach view) — whose retire, or whose check again by the seat whose
// check it is, would meet it.
type EvidenceDemand struct {
	CheckKind, AnswerKind string
	// AnswerNoun and AnswerVerb are the owed kind's row's noun and placing verb.
	AnswerNoun, AnswerVerb string
	Answered               bool
	// Awaiting is the gap view's awaiting_evidence: the gap is open, and its demand owed and unmet.
	Awaiting bool
	// Reach is each evidence anchor the demand reaches, the gap's own first, read only while the
	// demand is unmet: the one case a refusal names them.
	Reach []Reach
}

// Reach is one evidence anchor a gap's demand reaches: the anchor From is about, checked by its
// minter Seat, or the anchor of a reading From carries, read by Seat. From is the gap itself, or an
// ancestor it carries the demand from.
type Reach struct {
	Anchor, Seat, From string
	About              bool
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
// retire of the gap's only answer and the close racing it cannot both land. The gap view is read
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
	rows, err := q.Query(`SELECT "anchor", "seat", "from_gap", "via" = 'about' FROM "demand_reach"
		 WHERE "gap_id" = ? AND "anchor" IS NOT NULL ORDER BY "from_gap" <> "gap_id", "via", "reached"`, gapID)
	if err != nil {
		return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r Reach
		if err := rows.Scan(&r.Anchor, &r.Seat, &r.From, &r.About); err != nil {
			return EvidenceDemand{}, fmt.Errorf("asking the record: %w", err)
		}
		// A gap about an anchor of a kind that backs nothing reaches no evidence anchor.
		if r.About && !anchor.Backs(r.Anchor) {
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
// evidence anchor the demand reaches, and the check again by the seat whose check it is — and the
// bench, for a demand the lens no longer holds.
func DemandRefusal(gapID string, d EvidenceDemand) error {
	ways := []string{fmt.Sprintf("blue's `%s --answers %s`, a %s placed on the sentence as its answer (after a reword, on the reworded sentence)", d.AnswerVerb, gapID, d.AnswerNoun)}
	var retire, again []string
	for _, r := range d.Reach {
		var whose string
		switch {
		case r.About && r.From == gapID:
			whose = fmt.Sprintf("the %s it is about", anchor.Kind(r.Anchor))
		case r.About:
			whose = fmt.Sprintf("the %s %s was about, whose demand this gap carries", anchor.Kind(r.Anchor), r.From)
		case r.From == gapID:
			whose = fmt.Sprintf("the citation whose reading by %s it carries", r.Seat)
		default:
			whose = fmt.Sprintf("the citation whose reading by %s %s carries, whose demand this gap carries", r.Seat, r.From)
		}
		retire = appendNew(retire, fmt.Sprintf("blue's `retire --anchor %s`, %s", r.Anchor, whose))
		// THE CHECK AGAIN, BY ITS OWN SEAT, SINCE THE MINT, of an anchor of the demanded kind: a reading
		// that backs under `source`, a `sound` judgement under `computation`. Another seat's check meets
		// nothing.
		switch {
		case anchor.Kind(r.Anchor) != d.AnswerKind:
		case d.AnswerKind == "computation":
			again = appendNew(again, fmt.Sprintf("a `sound` judgement of %s by %s, made since the gap was minted", r.Anchor, r.Seat))
		default:
			again = appendNew(again, fmt.Sprintf("a reading of %s by %s, made since the gap was minted, that reads the work and backs the sentence", r.Anchor, r.Seat))
		}
	}
	if len(retire) > 0 {
		ways = append(ways, strings.Join(retire, ", or "))
	}
	ways = append(ways, again...)
	return feov.Errorf(feov.Conflict, "gap %s was minted --check-kind %s, and nothing meets its live-answer demand, so this close is refused under every disposition — `moot`, `not_a_defect` and a close to a successor included. Nothing is recorded. "+
		"The demand: %s. What meets it on this gap: %s; another seat's check meets nothing. "+
		"A demand you no longer hold is not closed: file `motion docket file` on the gap, and the bench's closing ruling, which is not held to the demand, disposes of it; a regrade keeps it open",
		gapID, d.CheckKind, recordpb.DemandClause, strings.Join(ways, "; or "))
}

// appendNew appends s to list unless list holds it: an anchor the gap and an ancestor both reach is
// one remedy.
func appendNew(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

// THE DEMAND MOVES TO THE SUCCESSOR (ruled by gblock 2026-11-05). A mint superseding an open gap
// closes it in the same act (supersession.go), and that closing goes through neither the close verb
// nor judgeAct's Close case, so a `source` or `computation` gap whose demand nothing meets would close
// unanswered by being superseded. The successor carries the ancestor's check kind instead, and with it
// the demand, which it cannot close without its own live answer. An ancestor whose demand is met, whose
// check kind owes none, or that is already closed (the supersede is lineage only) is unaffected.

// SupersessionAdmitted asks of run, before the mint's write, what the write holds a mint's
// `supersedes` to, in the order a seat is told it: only an open gap's minter supersedes it
// (requireSupersedable), and a successor of an open gap nothing answers yet carries its check kind
// (requireSuccessorCarriesDemand). It refuses before the record's placement and shape checks; the
// write asks both again in its transaction, where the answer holds. A read error is returned, never
// folded into "admitted".
func SupersessionAdmitted(run Run, seatID string, m *recordpb.Mint) error {
	if len(m.GetSupersedes()) == 0 {
		return nil
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return err
	}
	if err := requireSupersedable(db, seatID, m); err != nil {
		return err
	}
	return requireSuccessorCarriesDemand(db, m)
}

// awaitingAncestor is one gap a mint's `supersedes` names that awaits a live answer, and its demand.
type awaitingAncestor struct {
	id     string
	demand EvidenceDemand
}

// requireSuccessorCarriesDemand refuses a mint naming in `supersedes` an open gap that awaits a live
// answer (the gap view's awaiting_evidence: open, its check kind one EvidenceTables answers, and its
// demand unmet) unless the mint carries that gap's check kind, naming every such gap. Replay never
// asks it, so an archived supersede replays as it was recorded.
func requireSuccessorCarriesDemand(q recordsql.Querier, m *recordpb.Mint) error {
	kind := recordpb.Word(m.GetCheckKind())
	var awaiting []awaitingAncestor
	dropped := false
	for _, anc := range m.GetSupersedes() {
		d, err := evidenceDemandAt(q, anc)
		if err != nil {
			return err
		}
		if !d.Awaiting {
			continue
		}
		awaiting = append(awaiting, awaitingAncestor{anc, d})
		dropped = dropped || d.CheckKind != kind
	}
	if !dropped {
		return nil
	}
	return successorDemandRefusal(awaiting, kind)
}

// successorDemandRefusal is the mint's refusal of a successor minted --check-kind kind over the open
// gaps awaiting, which nothing answers yet, one at least at another check kind: each demand passes to
// the successor, so the successor carries the ancestors' check kind — one kind, so ancestors of
// different kinds are superseded in separate mints; the remedies are that kind, each demand met
// first, or the bench.
func successorDemandRefusal(awaiting []awaitingAncestor, kind string) error {
	var named, met, ids []string
	kinds := map[string]bool{}
	for _, a := range awaiting {
		named = append(named, fmt.Sprintf("%s, an open gap minted --check-kind %s", a.id, a.demand.CheckKind))
		met = append(met, fmt.Sprintf("blue's `%s --answers %s`, a %s placed as its answer", a.demand.AnswerVerb, a.id, a.demand.AnswerNoun))
		ids = append(ids, a.id)
		kinds[a.demand.CheckKind] = true
	}
	it, them := "it", "it"
	if len(awaiting) > 1 {
		it, them = "each", "them"
	}
	var carry string
	if first := awaiting[0].demand; len(kinds) == 1 {
		carry = fmt.Sprintf("Mint the successor --check-kind %s, and it owes the %s %s owed", first.CheckKind, first.AnswerNoun, strings.Join(ids, " and "))
	} else {
		carry = "They owe different check kinds and a successor carries one, so no one mint supersedes them all: supersede them in separate mints, each successor minted --check-kind its ancestor's"
	}
	return feov.Errorf(feov.Conflict, "--supersedes names %s, and nothing answers %s yet; a mint superseding an open gap closes it: its live-answer demand passes to the successor, which carries the same check kind, and this mint is --check-kind %s. Nothing was recorded. "+
		"The demand: %s. %s; or meet the demand on %s first — %s, is one way, and its close refusal names the rest — and supersede %s then; "+
		"or, for a demand you no longer hold, file `motion docket file` on %s, and the bench's closing ruling, which is not held to the demand, disposes of it",
		strings.Join(named, ", and "), them, kind, recordpb.DemandClause, carry, it, strings.Join(met, "; "), them, it)
}
