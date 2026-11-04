package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// ESTOPPEL — red is bound by the fix it prescribed.
//
// THE PATHOLOGY, MEASURED. In the 2026-08-04 smoke blue made 26 edits; round 1 was nine
// edits, every one additive, +2,856 characters, each answering a gap red had raised. Epoch
// 2's gaps then targeted that new text: 3 of 3 were about text blue added in round 1 AT
// RED'S INSTRUCTION. Blue's discipline was not the problem — 19 of 26 edits named the gap
// they answered. It did what it was told, carefully, and was penalised for it.
//
// The rule this file enforces: a finding whose location is text red PRESCRIBED and blue
// applied VERBATIM is an amendment to the gap that prescribed it, not a fresh gap. Red may
// still be unsatisfied — but it argues that on the original gap, where its own prescription
// is on the record next to it, rather than opening a new one with a clean slate.
//
// It is bounded to VERBATIM application on purpose. If blue counter-edited, the text is
// blue's authorship and red never proposed it, so nothing estops red from auditing it
// normally. Estoppel attaches to red's own words, and to nothing else.
//
// A READING'S GAP IS EXEMPT (gblock, 2026-10-31: "Exempt, stated"). A lens's reading of a source
// that finds against the sentence mints its gap in the reading's own act (AppendReading), and no
// estoppel guard is attached there: the source's refutation is new evidence the prescription never
// weighed, not red relitigating its own words. So a reading's gap opens also on text blue applied
// verbatim from a red --new, nothing is logged, and EstoppelRejectionsOf counts lens mints alone.

// minEstoppelOverlap is the shortest prescribed text that may trigger the guard, in
// characters after whitespace collapse.
//
// A short prescription ("7 is prime", "38%") can legitimately occur in text red never wrote,
// and refusing a real finding is far worse than missing an estoppel — the guard exists to
// stop a systematic pathology, not to win every edge case. Below this length the finding is
// allowed through and red's own judgment (which the prompt now names explicitly) is the only
// check. Stated so nobody reads the guard as exhaustive.
const minEstoppelOverlap = 40

// appliedVerbatim maps each gap id whose concrete proposal blue applied EXACTLY to the text
// red prescribed. Only these gaps estop red.
//
// The fact is recorded at edit time by the tool, never by a seat asserting it — TWO WAYS, and
// the difference is one of kind. On an ordinary `blue edit` it is the outcome of comparing the
// bytes blue typed against the ones red prescribed. On `blue edit --accept` the tool supplied
// those bytes from the mint, so it holds by construction and no transcription slip can cost it.
// `BlueEdit.accepted` tells the two apart where that matters; estoppel does not care which.
func appliedVerbatim(evs []*Event, gaps map[string]*Gap) map[string]string {
	out := map[string]string{}
	for _, e := range evs {
		// THE BODY IS THE FILTER. "not a blue_edit" is the same set the `e.Type != "blue_edit"`
		// test selected, and it cannot go stale against the enum the way a second reading of the
		// type could.
		ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		if !ok || !ed.GetAppliedVerbatim() {
			continue
		}
		id := ed.GetAnswers()
		g, ok := gaps[id]
		if !ok || g.Mint == nil {
			continue
		}
		if fixNew := g.Mint.GetFixNew(); fixNew != "" {
			out[id] = fixNew
		}
	}
	return out
}

// Estoppel is a prescription a mint's quote relitigates: the gap whose fix it is, the seat that minted
// that gap — the author of the prescription — and the prescribed text.
type Estoppel struct{ Gap, PrescribedBy, Text string }

// EstoppelConflict reports the applied prescription `quote` is relitigating, and whether there is one.
//
// RED IS THE PARTY BOUND, NOT THE LENS THAT WROTE IT. The guard was written when one red seat minted
// every gap, so "red's prescription" and "the minting seat's prescription" were the same set; with the
// lenses minting, a lens's mint against another lens's applied fix is refused too, and the refusal
// names the author from the record rather than telling the minting lens the words were its own.
//
// Containment either way, on whitespace-collapsed text: the new finding may quote a fragment
// of the prescribed sentence or a span that swallows it whole, and both are the same act.
func EstoppelConflict(f Family, quote string) (Estoppel, bool) {
	// VISIBLE, NOT MERELY COLLAPSED. Collapsing whitespace alone leaves the anchors, and the
	// two sides of this comparison differ by exactly that: `fix_new` is red's prose, written before
	// any anchor existed, while `quote` is taken from the report, which anchors the sentence the
	// moment the fix lands. On raw bytes this check goes blind at the instant it becomes relevant.
	q := anchortext.Visible(quote)
	if len(q) == 0 {
		return Estoppel{}, false
	}
	for id, fixNew := range appliedVerbatim(f.Events, GapsByID(f.Gaps)) {
		p := anchortext.Visible(fixNew)
		if len(p) < minEstoppelOverlap && len(q) < minEstoppelOverlap {
			continue
		}
		if strings.Contains(p, q) || strings.Contains(q, p) {
			return Estoppel{Gap: id, PrescribedBy: minterOf(f.Events, id), Text: fixNew}, true
		}
	}
	return Estoppel{}, false
}

// minterOf is the seat whose mint created gapID: the event's seat, which the record stamps at the write.
func minterOf(evs []*Event, gapID string) string {
	for _, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && m.GetGapId() == gapID {
			return e.GetSeatId()
		}
	}
	return ""
}

// DeclineStatsOf is DeclineStats over the family pieces themselves (events + the gap index),
// for the run-shaped renders.
func DeclineStatsOf(evs []*Event, gaps map[string]*Gap) (offered, applied, declined int) {
	verbatim := appliedVerbatim(evs, gaps)
	answered := map[string]bool{}
	for _, e := range evs {
		if ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e); ok {
			if id := ed.GetAnswers(); id != "" {
				answered[id] = true
			}
		}
		if id := retireAnswers(e); id != "" {
			answered[id] = true
		}
	}
	for id, g := range gaps {
		if g.Mint == nil || g.Mint.GetFixBasis() != "verified" {
			continue
		}
		offered++
		switch {
		case verbatim[id] != "":
			applied++
		case answered[id]:
			declined++
		}
	}
	return offered, applied, declined
}

// Proposal is the pair a gap's mint prescribes, as an edit of report applies it — the pair the board
// serves, `blue edit --accept` sends, and a typed application is compared with, so the three cannot
// disagree. found=false means no mint for that gap. An empty fixNew with found=true is a real state —
// red raised the gap without prescribing concrete text — and is the caller's to refuse, because the
// refusal it wants to give names the verb it was called from.
func Proposal(run Run, gapID, report string) (old, fixNew string, found bool, err error) {
	db, err := openRunForRead(run)
	if err != nil {
		return "", "", false, err
	}
	return proposalAt(db, gapID, report)
}

// proposalAt is Proposal asked of q.
func proposalAt(q recordsql.Querier, gapID, report string) (old, fixNew string, found bool, err error) {
	loc, fix, found, err := prescriptionAt(q, gapID)
	if err != nil || !found {
		return "", "", false, err
	}
	old, fixNew = ProposalOver(report, gapID, loc, fix)
	return old, fixNew, true, nil
}

// prescriptionAt is the location and text gapID's mint recorded, as q holds them.
func prescriptionAt(q recordsql.Querier, gapID string) (loc, fix string, found bool, err error) {
	var l, f sql.NullString
	found, err = queryRowAt(q, []any{&l, &f}, `SELECT "location", "fix_new" FROM "mint" WHERE "gap_id" = ?`, gapID)
	return l.String, f.String, found, err
}

// ProposalOver is a gap's prescribed pair over report. The location is LOCATED BY THE GAP'S TOKEN,
// never by its text alone: among the location's occurrences, the one followed by an anchor run holding
// the gap's own anchor — after any trailing punctuation, where an edit's abutting-anchor rule reads
// it — is the gap's, and the report holds that token once, so a sentence whose words the report holds
// twice yields the occurrence the gap stands at. old is the location as the report holds it there,
// through that run, then the location's own trailing punctuation: the edit then replaces the sentence with its
// anchors as `show report` prints them. new is fix with each anchor of the run it does not carry that
// does not back its sentence — the gap and finding anchors — placed at its content end, where Attach
// places an anchor on any quote ending there; the run's citations and computations are the edit's to carry
// or refuse. Otherwise the recorded pair.
//
// A FINDING RIDES WITH THE GAP. A gap is minted from a finding on its sentence, and left to the edit,
// that finding was dropped by red's text and every such accept refused as stale on a report nothing
// had changed (m25 ev512, ev644: 0 of 3 accepts applied). A citation or computation is the sentence's
// evidence, and carried onto red's rewording it would back text its source was never read against, so
// it comes along only where AutoPlace finds its sentence word for word. `lens mint` judges its pair
// through this function too (blue.PlanAccept), so a pair the accept would refuse is refused there. No
// occurrence whose run holds the gap's token — the token gone, or a run that does not hold it — is
// the recorded pair.
func ProposalOver(report, gapID, loc, fix string) (string, string) {
	if fix == "" {
		return loc, fix
	}
	tok := anchor.Token(gapID)
	for _, sp := range anchortext.LocateAll(report, loc) {
		start, end := sp[0], sp[1]
		tail := report[end:]
		after := strings.TrimLeft(tail, anchortext.TrailingPunct)
		run := after[:anchor.SkipRun(after, 0)]
		if !strings.Contains(run, tok) {
			continue
		}
		trimmed := strings.TrimSpace(loc)
		old := report[start:end+len(tail)-len(after)+len(run)] + trimmed[len(strings.TrimRight(trimmed, anchortext.TrailingPunct)):]
		at := anchortext.ContentEnd(fix)
		var carried strings.Builder
		for _, id := range anchor.IDs(run) {
			if !anchor.Backs(id) && !strings.Contains(fix, anchor.Token(id)) {
				carried.WriteString(anchor.Token(id))
			}
		}
		return old, fix[:at] + carried.String() + fix[at:]
	}
	return loc, fix
}

// ProposalAppliedVerbatim reports whether the pair a seat typed is EXACTLY the fix the named gap
// proposed over the report it edited — byte-for-byte on both halves, against Proposal's pair.
//
// Exactness is the whole point. A near-application ("I applied red's fix, roughly") is blue
// authoring, and it must be audited as blue's text; only an identical pair estops red from
// re-arguing what it wrote itself. Whitespace is NOT normalized here for that reason: the
// looser the match, the more of blue's own writing red is barred from auditing.
func ProposalAppliedVerbatim(run Run, gapID, report, old, new string) (bool, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return false, err
	}
	return proposalAppliedVerbatimAt(db, gapID, report, old, new)
}

// proposalAppliedVerbatimAt is ProposalAppliedVerbatim asked of q.
func proposalAppliedVerbatimAt(q recordsql.Querier, gapID, report, old, new string) (bool, error) {
	pOld, pNew, found, err := proposalAt(q, gapID, report)
	if err != nil || !found || pNew == "" {
		return false, err
	}
	return pOld == old && pNew == new, nil
}

// THE THREE FRICTION-KIND WORDS ARE GONE, and the deletion is the point rather than a tidy-up.
// `FrictionKindEstoppel`, `FrictionKindToolError` and `FrictionKindKey` were the PRE-SCHEMA
// carrier: two words and the payload key that held them. `Friction.kind` is now a generated enum,
// so the words were a SECOND SPELLING of values the schema already owns — and the key named a
// slot in a payload that no longer exists, which made its own doc comment false.
//
// They were not harmless while they stood. A constant still in scope is an invitation to convert
// an enum compare back into a string round-trip, which is exactly the drift this migration
// removes: the estoppel count was once derived by matching the prose substring "estoppel —",
// reading ZERO both when the guard never fired and when someone reworded the message (#283).
// Write `recordpb.FrictionKind_FRICTION_KIND_ESTOPPEL` / `_TOOL_ERROR`; read the field.

// EstoppelRejectionsOf is EstoppelRejections over the events themselves.
func EstoppelRejectionsOf(evs []*Event) int {
	n := 0
	for _, e := range evs {
		// THE FIELD IS NOW THE SCHEMA'S, and the count reads the enum rather than a word. The
		// type is required at the write, so an entry that is not an estoppel refusal carries a
		// different type rather than an absent one, and still does not count.
		if fr, ok := recordpb.BodyAs[*recordpb.Log](e); ok &&
			fr.GetType() == recordpb.LogType_LOG_TYPE_ESTOPPEL {
			n++
		}
	}
	return n
}

// `CheckKindComputation` IS GONE for the same reason. It was the acceptance-check kind prose
// cannot settle, compared as a bare literal at three sites in three packages — three chances to
// disagree about the one rule that decides whether a gap can close at all. The schema owns the
// values, recordsql.EvidenceTables says which check kinds an evidence anchor answers, and the
// QUESTION has one home: `Gap.AnswerKind()` in replay.go, which every one of those sites calls.
// Where a lowercase spelling is genuinely needed for a projection it is `recordpb.Spelling`, never
// a literal.

// EvidenceAnswers reports whether the gap's live-answer demand is met: the gap view's
// evidence_answered (recordpb.DemandClause) — a live evidence anchor of its check kind placed as its
// answer, the anchor it is about or whose reading it carries retired, or its own check of that
// anchor made again since the mint and holding.
//
// The answer is a FIELD checked against the board at write time (Cite.answers,
// Computation.answers), never a gap id mentioned in prose — the convention blue_edit's --answers
// replaced after it measured 73% reliable, which is reliable enough to look like a key and not
// reliable enough to be one.
//
// A read error is returned, never folded into false: "not answered" and "could not look" are two
// answers, and the caller refuses on the second naming it.
func EvidenceAnswers(run Run, gapID string) (bool, error) {
	if gapID == "" {
		return false, nil
	}
	return recordHas(run, `SELECT 1 FROM "gap" WHERE "gap_id" = ? AND "evidence_answered" LIMIT 1`, gapID)
}

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

// EvidenceDemandOf reads gap gapID's demand off the mint, the gap view and the carrying readings. A
// gap not on the record answers the zero demand; a read error is returned, never folded into "no
// demand".
func EvidenceDemandOf(run Run, gapID string) (EvidenceDemand, error) {
	var d EvidenceDemand
	if gapID == "" {
		return d, nil
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return d, err
	}
	var aboutKind, aboutRef string
	found, err := queryRowAt(db, []any{&d.CheckKind, &d.Answered, &aboutKind, &aboutRef, &d.MintedBy},
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
	rows, err := db.Query(`SELECT "anchor", "seat_id" FROM "carrying_reading" WHERE "gap_id" = ? AND "anchor" IS NOT NULL ORDER BY "event_id"`, gapID)
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

// RemovalVerified / RemovalAsserted say how far a recorded retirement can be trusted, on the
// same footing as fix_basis, a computation's basis and verdict_basis.
//
// A retire used to record whatever it was told: nothing confirmed the claim had ever been in
// the report, and nothing confirmed it had left. That mattered beyond tidiness — the changelog
// lists every retirement as a claim that left the report, so a retirement of a claim that was
// never there is a removal the reader is told of and the report never had.
const (
	// RemovalVerified: the claim is absent from the report now AND appears in the old span of
	// a recorded edit, so the record can SHOW it leaving.
	RemovalVerified = "verified"
	// RemovalAsserted: absent now, but nothing on the record shows it was ever present. The
	// round-0 author writes the report directly rather than through edits, so a claim it never
	// wrote — or wrote and rewrote in the same sitting — lands here honestly.
	RemovalAsserted = "asserted"
)

// EditSpan is one recorded blue_edit's (old, new) pair, in record order.
type EditSpan struct{ Old, New string }

// editSpansAt returns every recorded blue_edit's (old, new) on q, in record order — what `blue
// retire` reads (Landing.EditSpans) to learn which edit took a claim out and what that edit left
// behind in its place.
func editSpansAt(q recordsql.Querier) ([]EditSpan, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(`SELECT "old", "new" FROM "blue_edit" ORDER BY "event_id"`)
	if err != nil {
		return nil, fmt.Errorf("reading the recorded edits: %w", err)
	}
	defer rows.Close()
	var out []EditSpan
	for rows.Next() {
		var o, n sql.NullString
		if err := rows.Scan(&o, &n); err != nil {
			return nil, err
		}
		out = append(out, EditSpan{Old: o.String, New: n.String})
	}
	return out, rows.Err()
}

// AnchorRetire is the retire that took an anchor out of the report: its event, the claim and reason
// it recorded, and whether it kept that claim — kept true for an anchor retired alone off a sentence
// that stays.
type AnchorRetire struct {
	Event         int64
	Claim, Reason string
	Kept          bool
}

// AnchorRetiredAt answers where an anchor went when it is no longer in the report: the retire
// event that took it out — with its claim, or, kept true, off a claim that stayed. found is false
// for an anchor no retire named — then it is a stale reference or another run's id, and the caller
// says so.
func AnchorRetiredAt(run Run, id string) (r AnchorRetire, found bool, err error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return r, false, err
	}
	return anchorRetiredAt(db, id)
}

// anchorRetiredAt is AnchorRetiredAt asked of q — on the write path the writing transaction, so a
// check that refuses an act on a retired anchor and the retire racing it cannot both land.
func anchorRetiredAt(q recordsql.Querier, id string) (r AnchorRetire, found bool, err error) {
	var c, why sql.NullString
	var k sql.NullBool
	found, err = queryRowAt(q, []any{&r.Event, &c, &why, &k},
		`SELECT r."event_id", r."claim", r."reason", r."claim_kept" FROM "retire_anchors" ra JOIN "retire" r ON r."event_id" = ra."event_id"
		 WHERE ra."value" = ? ORDER BY r."event_id" LIMIT 1`, id)
	r.Claim, r.Reason, r.Kept = c.String, why.String, k.Bool
	return r, found, err
}

// Said is the retire as a refusal names it: its event, its reason, and whether the sentence stayed.
func (r AnchorRetire) Said() string {
	kept := "it left the report with its claim " + quoted(r.Claim)
	if r.Kept {
		kept = "its sentence " + quoted(r.Claim) + " kept"
	}
	return fmt.Sprintf("blue retired it at event %d (%s; %s)", r.Event, reasonOf(r.Reason), kept)
}
