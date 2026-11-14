package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE QUESTION IS SPLIT INTO PARTS ON THE RECORD, AND EACH PART CARRIES ITS COVERAGE (ruled by gblock
// 2026-11-11, E and J). A run passed VERIFIED while it omitted a named quarter of its question, because
// whether the report answered the question lived in prose nothing could refuse. So blue splits the
// question into question parts (`question split`), records each part's coverage — covered, not
// obtained or excluded, with the reason (`question cover`) — and the logic lens checks each part's
// coverage as it checks an inquiry's fate (`question check`). A PASS over a part not covered derives
// QUALIFIED, and report.md states each such part in a sentence the tool composes from the coverage.
//
// EVERY FACT HERE IS THE question_part_state VIEW'S (gblock 2026-11-14: "Is this you doing in go what
// you can just as easily do from views and prepared statements again?"). The view states each part's
// coverage and whether a check stands on it; this file reads its rows and words them, and replays no
// event.
//
// Each part holds one judged state:
//
//	open       no coverage is recorded — blue owes it one
//	wanting    an open material gap is about it — the gap's route clears it
//	unchecked  it has coverage and no check stands — the logic lens owes the check
//	sound      a check stands on it
//
// A record with a cast and no part at all holds the PASS as question_unsplit: a run that never split
// its question has no account of whether it answered it.

// questionPartOwner is the lens that owes every question part's check (ruled by gblock 2026-11-11, E:
// the logic lens checks it, as it checks inquiries). One owner, for the reason an inquiry has one.
const questionPartOwner = inquiryOwner

// QuestionPartOwner is the seat that owes every question part's check.
func QuestionPartOwner() string { return questionPartOwner }

// The judged states of a question part, as the wire prints them.
const (
	PartOpen      = "open"
	PartWanting   = "wanting"
	PartUnchecked = "unchecked"
	PartSound     = "sound"
)

// QuestionPartRow is one row of the question_part_state view: a part with its latest coverage and its
// judged state's inputs.
type QuestionPartRow struct {
	ID         string `json:"part_id"`
	Text       string `json:"text"`
	ProposedBy string `json:"proposed_by"`
	// Coverage is the latest coverage's word, "" while none is recorded; Reason is that record's.
	Coverage  string `json:"coverage,omitempty"`
	Reason    string `json:"reason,omitempty"`
	CoveredBy string `json:"covered_by,omitempty"`
	// CheckedBy is the seat of the check that stands, "" while none does.
	CheckedBy string `json:"checked_by,omitempty"`
	// State is the judged state: open, wanting, unchecked or sound.
	State string `json:"state"`
	// proposed, covered, gapMint and lastCheck are the events.id of the proposal, the latest coverage
	// record, the latest gap minted about the part and its latest check; 0 where there is none.
	proposed, covered, gapMint, lastCheck int64
	wanting                               int
	stands                                bool
}

// questionPartsSQL reads the view, in proposal order.
const questionPartsSQL = `SELECT q."part_id", COALESCE(q."text", ''), COALESCE(q."proposed_by", ''), q."proposed_seq",
       COALESCE(q."coverage", ''), COALESCE(q."reason", ''), COALESCE(q."coverage_seq", 0), COALESCE(q."covered_by", ''),
       COALESCE(q."gap_seq", 0), q."wanting", COALESCE(q."check_seq", 0), COALESCE(q."checked_by", ''), q."check_stands"
  FROM "question_part_state" q ORDER BY q."proposed_seq"`

// questionPartsAt is every question part on the record q reads, in proposal order. Never nil on a
// read that succeeded: an empty list is a record with no part.
func questionPartsAt(q recordsql.Querier) ([]QuestionPartRow, error) {
	out := []QuestionPartRow{}
	if noRecord(q) {
		return out, nil
	}
	rows, err := q.Query(questionPartsSQL)
	if err != nil {
		return nil, fmt.Errorf("asking the record for its question parts: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p QuestionPartRow
		var checkedBy string
		if err := rows.Scan(&p.ID, &p.Text, &p.ProposedBy, &p.proposed, &p.Coverage, &p.Reason, &p.covered, &p.CoveredBy,
			&p.gapMint, &p.wanting, &p.lastCheck, &checkedBy, &p.stands); err != nil {
			return nil, err
		}
		if p.stands {
			p.CheckedBy = checkedBy
		}
		p.State = p.state()
		out = append(out, p)
	}
	return out, rows.Err()
}

// QuestionParts is every question part on the run, in proposal order.
func QuestionParts(run Run) ([]QuestionPartRow, error) {
	var out []QuestionPartRow
	err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		out, err = questionPartsAt(q)
		return err
	})
	return out, err
}

// state is the judged state. WANTING WINS over unchecked, as it does for an inquiry: the check waits
// for the gap's route, and the gap holds the PASS meanwhile.
func (p QuestionPartRow) state() string {
	switch {
	case p.wanting > 0:
		return PartWanting
	case p.Coverage == "":
		return PartOpen
	case p.stands:
		return PartSound
	}
	return PartUnchecked
}

// Covered reports whether the part's latest coverage is `covered`.
func (p QuestionPartRow) Covered() bool {
	return p.Coverage == recordpb.Word(recordpb.PartCoverage_PART_COVERAGE_COVERED)
}

// CheckableFrom reports whether a check from a sitting opened at event sitting judges the part as it
// stands: its latest coverage and the latest gap about it both landed before the sitting opened.
// requireQuestionPartCheck refuses the check otherwise.
func (p QuestionPartRow) CheckableFrom(sitting int64) bool {
	return p.covered <= sitting && p.gapMint <= sitting
}

// partBlockersOf is the PASS gate's three question arms: no part on a record with a cast, a part with
// no coverage, and a part with coverage, no open material gap about it and no check standing on it. A
// wanting part holds the PASS through its gap's material_gap row, and has no row here.
func partBlockersOf(cast bool, parts []QuestionPartRow) []Blocker {
	if !cast {
		return nil
	}
	if len(parts) == 0 {
		return []Blocker{{Kind: BlockerQuestionUnsplit, Owner: blueRespondSeat}}
	}
	var out []Blocker
	for _, p := range parts {
		switch p.State {
		case PartOpen:
			out = append(out, Blocker{Kind: BlockerOpenPart, Subject: p.ID, Owner: blueRespondSeat, Detail: p.Text, since: p.proposed})
		case PartUnchecked:
			out = append(out, Blocker{Kind: BlockerUncheckedPart, Subject: p.ID, Owner: questionPartOwner, Detail: p.Coverage,
				since: p.covered, placed: max(p.covered, p.gapMint), lastCheck: p.lastCheck, gapMint: p.gapMint})
		}
	}
	return out
}

// unsplitFact, openPartFact and uncheckedPartFact are the one sentence every wording of each kind
// states.
func unsplitFact() string {
	return "the question this run was asked is split into no question part on the record, so nothing states which parts of it the report answers"
}

func openPartFact(b Blocker) string {
	return fmt.Sprintf("question part %s (%q) has no coverage recorded", b.Subject, b.Detail)
}

func uncheckedPartFact(b Blocker) string {
	s := fmt.Sprintf("question part %s is recorded %s (its coverage recorded at event %d) and no check of it stands", b.Subject, b.Detail, b.since)
	if b.lastCheck > 0 {
		s += fmt.Sprintf(" — its last check, event %d, came before that coverage or before a gap minted about it", b.lastCheck)
	}
	return s
}

// uncheckedPartAct is how an unchecked part is judged.
func uncheckedPartAct(b Blocker) string {
	return fmt.Sprintf("%s owes the judgement: read its row (`show question`) and the report against it, then `question check --id %s` where its coverage stands as the report's account of that part, or mint the gap about it (`mint --about-kind question --about %s`, class coverage_misstated) where it does not",
		ownerOr(b, questionPartOwner), b.Subject, b.Subject)
}

// blueQuestionItemsOf is a blue seat's blocking items for the question: the frontier owes the split
// while the record holds no part, the synthesizer and blue-respond each part's coverage while it has
// none, and blue-respond the split as well. A lane owes none: it reads the parts.
func blueQuestionItemsOf(seatID string, parts []QuestionPartRow) []Item {
	switch seatID {
	case frontierSeat, synthesizerSeat, blueRespondSeat:
	default:
		return nil
	}
	if len(parts) == 0 {
		if seatID == synthesizerSeat {
			return nil
		}
		return []Item{{Blocks: true, Kind: ItemKind(BlockerQuestionUnsplit), What: unsplitFact() + " — split it before your sitting ends: one `question split` for each part of the question a reader would ask answered, in the subject's terms. A part not split out is a part nothing holds the report to, and the PASS is refused while the question has none"}}
	}
	if seatID == frontierSeat {
		return nil
	}
	var out []Item
	for _, p := range parts {
		if p.State != PartOpen {
			continue
		}
		out = append(out, Item{Blocks: true, Kind: ItemKind(BlockerOpenPart), Subject: about(SubjectQuestionPart, p.ID),
			What: fmt.Sprintf("question part %s (%q) has no coverage recorded — record it before your sitting ends (`question cover --id %s --as covered|not_obtained|excluded --reason \"...\"`): covered where the report answers it, not_obtained where the sources that would answer it could not be obtained, excluded where it lies outside the report's answer. The PASS is refused while it has none, and a part not covered derives QUALIFIED", p.ID, p.Text, p.ID)})
	}
	return out
}

// The blue seats the question's items are for.
const (
	frontierSeat    = "frontier"
	synthesizerSeat = "blue-synthesize"
)

// questionPartWord is a coverage value's word, refused where the schema does not hold it.
func questionPartCoverageOf(word string) (recordpb.PartCoverage, bool) {
	vd, ok := recordpb.BySpelling(recordpb.PartCoverage(0).Descriptor(), word)
	if !ok || vd.Number() == 0 {
		return recordpb.PartCoverage_PART_COVERAGE_UNSPECIFIED, false
	}
	return recordpb.PartCoverage(vd.Number()), true
}

// PartCoverageOf is the coverage word to the schema's value; false where the word is no coverage.
func PartCoverageOf(word string) (recordpb.PartCoverage, bool) { return questionPartCoverageOf(word) }

// requireQuestionPartRef refuses an id no question part event wrote.
func requireQuestionPartRef(q recordsql.Querier, id, flag string) error {
	found, err := recordHasAt(q, `SELECT 1 FROM "question_part" WHERE "part_id" = ? LIMIT 1`, id)
	if err != nil {
		return err
	}
	if !found {
		return feov.Errorf(feov.Validation, "%s names question part %s, which blue never split out — `show question` lists every part with its id", flag, id)
	}
	return nil
}

// QuestionPartExists is a question part id flag's check: the part is on the record of the run in
// runDir.
func QuestionPartExists(runDir, flag, id string) error {
	db, err := recordOfDir(runDir)
	if err != nil {
		return err
	}
	return requireQuestionPartRef(db, id, flag)
}

// requireQuestionPart refuses a question part event the record cannot hold: a proposal with no text,
// or one that carries coverage; a coverage record of a part nobody proposed, with no coverage, or with
// no reason.
func requireQuestionPart(q recordsql.Querier, b *recordpb.QuestionPart) error {
	if b.GetPartId() == "" {
		return fmt.Errorf("a question part requires an id — the tool assigns one on `question split` and --id names it on `question cover`")
	}
	if b.GetSupersedesCoverage() == "" {
		if strings.TrimSpace(b.GetText()) == "" {
			return fmt.Errorf("requires --text (the part of the question this names, in the subject's terms — a part with no text is no part of any question)")
		}
		if b.Coverage != nil || b.GetReason() != "" {
			return fmt.Errorf("a question part is split out with its text alone — its coverage and the reason for it are recorded on it with `question cover`")
		}
		return nil
	}
	if err := requireQuestionPartRef(q, b.GetPartId(), "--id"); err != nil {
		return err
	}
	if b.GetText() != "" {
		return fmt.Errorf("a coverage record carries no text — the part's text is the one it was split out with")
	}
	if b.GetCoverage() == recordpb.PartCoverage_PART_COVERAGE_UNSPECIFIED {
		return fmt.Errorf("requires --as (the part's coverage: covered, not_obtained or excluded)")
	}
	if strings.TrimSpace(b.GetReason()) == "" {
		return fmt.Errorf("a %s coverage requires --reason — what in the report answers the part, which sources could not be obtained and what was tried, or why the part lies outside the report's answer; the logic lens judges the coverage against it", recordpb.Word(b.GetCoverage()))
	}
	return nil
}

// requireQuestionPartCheck refuses a check the record cannot hold as a judgement of the part as it
// stands: any seat but the owner's, a part with no coverage, a wanting part, coverage or a gap about
// the part recorded after the seat's sitting opened, or a check that already stands. Asked under the
// write's lock.
func requireQuestionPartCheck(w writeView, seatID string, b *recordpb.QuestionPartCheck) error {
	id := b.GetPartId()
	if seatID != questionPartOwner {
		return feov.Errorf(feov.RoleViolation, "the check of a question part is %s's, the lens that owns every part's judgement; you are %s. Where you find part %s's coverage misstated, mint the gap about it (`mint --about-kind question --about %s`, class coverage_misstated)", questionPartOwner, seatID, id, id)
	}
	if err := requireQuestionPartRef(w.q, id, "--id"); err != nil {
		return err
	}
	parts, err := questionPartsAt(w.q)
	if err != nil {
		return err
	}
	var p QuestionPartRow
	for _, r := range parts {
		if r.ID == id {
			p = r
		}
	}
	_, win, err := eventsAt(w.q)
	if err != nil {
		return err
	}
	sitting := win.LatestSittingOf(seatID)
	switch {
	case p.wanting > 0:
		return feov.Errorf(feov.Conflict, "question part %s is wanting: an open material gap is about it, and its check waits for the gap's route — its minter closes it once blue's answer holds, or the bench rules it. Nothing is recorded", id)
	case p.Coverage == "":
		return feov.Errorf(feov.Conflict, "question part %s has no coverage recorded, so there is nothing to judge yet: blue owes it one. Nothing is recorded", id)
	case sitting < p.covered:
		return feov.Errorf(feov.Conflict, "question part %s's coverage was recorded at event %d, after your sitting opened at event %d, so this check would judge the coverage as it stood before. Nothing is recorded. The sitting you are readied for next lists it", id, p.covered, sitting)
	case sitting < p.gapMint:
		return feov.Errorf(feov.Conflict, "a gap was minted about question part %s at event %d, after your sitting opened at event %d, so this check would judge its coverage without that gap in hand. Nothing is recorded. The sitting you are readied for next lists it with the gap", id, p.gapMint, sitting)
	case p.stands:
		return feov.Errorf(feov.Conflict, "a check of question part %s stands — event %d, by %s — and neither its coverage nor any gap about it has moved since. Nothing is recorded", id, p.lastCheck, p.CheckedBy)
	}
	return nil
}

// QuestionJSON is the wire form of the question's parts, in proposal order: `show question --json`.
type QuestionJSON struct {
	Parts []QuestionPartRow `json:"parts"`
}

// QuestionJSONOf is the run's question parts, as the projection serves them.
func QuestionJSONOf(run Run) (QuestionJSON, error) {
	parts, err := QuestionParts(run)
	if parts == nil {
		parts = []QuestionPartRow{}
	}
	return QuestionJSON{Parts: parts}, err
}

// QuestionBody is the question projection without its heading: one line per part with its judged
// state, its coverage and the reason for it, and the seat whose check stands — "" where the record
// holds no part.
func QuestionBody(parts []QuestionPartRow) string {
	var b strings.Builder
	for _, p := range parts {
		fmt.Fprintf(&b, "- **%s** (%s) — %s\n", p.ID, p.State, p.Text)
		if p.Coverage == "" {
			b.WriteString("  - coverage: none recorded\n")
		} else {
			fmt.Fprintf(&b, "  - coverage: %s (%s) — %s\n", p.Coverage, p.CoveredBy, p.Reason)
		}
		if p.CheckedBy != "" {
			fmt.Fprintf(&b, "  - checked by %s\n", p.CheckedBy)
		}
	}
	return b.String()
}

// partCoveragesAtSQL counts the question parts proposed at or before a point by their latest coverage
// at that point — "" for a part with none there. A prepared read bounded by events.id, so a
// certification's line counts the parts as the record stood at it.
const partCoveragesAtSQL = `SELECT COALESCE(c."coverage", ''), COUNT(*)
  FROM (SELECT a0."part_id" AS "part_id" FROM "question_part" a0
         WHERE COALESCE(a0."part_id", '') != '' AND a0."supersedes_coverage" IS NULL AND a0."event_id" <= ?1
         GROUP BY a0."part_id") p
  LEFT JOIN "question_part" c ON c."event_id" =
    (SELECT MAX(x."event_id") FROM "question_part" x
      WHERE x."part_id" = p."part_id" AND x."supersedes_coverage" IS NOT NULL AND x."event_id" <= ?1)
  GROUP BY COALESCE(c."coverage", '')`

// PartCoverages is the question parts by coverage at one point, keyed by the coverage's word ("" for a
// part with none recorded there).
type PartCoverages map[string]int

// PartCoveragesAt is the question parts as the run's record stood at the event whose events.id is
// point, counted by their latest coverage there.
func PartCoveragesAt(run Run, point int64) (PartCoverages, error) {
	out := PartCoverages{}
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		rows, err := q.Query(partCoveragesAtSQL, point)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var cov string
			var n int
			if err := rows.Scan(&cov, &n); err != nil {
				return err
			}
			out[cov] = n
		}
		return rows.Err()
	})
	return out, err
}
