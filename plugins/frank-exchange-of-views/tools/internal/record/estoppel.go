package record

import (
	"database/sql"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// ESTOPPEL — red is bound by the fix it prescribed.
//
// THE PATHOLOGY, MEASURED. In the 2026-08-04 smoke blue made 26 edits; round 1 was nine
// edits, every one additive, +2,856 characters, each answering a gap red had raised. Epoch
// 2's gaps then targeted that new text: 3 of 3 were about text blue added in round 1 AT
// RED'S INSTRUCTION. Blue's discipline was not the problem — 19 of 26 edits named the gap
// they answered. It did what it was told, carefully, and was penalised for it.
//
// The rule this file enforces: a mint whose quote is text a lens PRESCRIBED and the author applied
// VERBATIM is the prescribing lens re-raising its own words, refused to that lens. It may still be
// unsatisfied — but it argues that on the original gap, where its own prescription is on the record
// next to it, rather than opening a new one with a clean slate. Any other lens audits the applied
// text as blue's.
//
// It is bounded to VERBATIM application on purpose. If blue counter-edited, the text is
// blue's authorship and red never proposed it, so nothing estops red from auditing it
// normally. Estoppel attaches to red's own words, and to nothing else.
//
// EVERY `--quote` MINT MEETS THE GUARD, a found source's claim_contradicts_source included (gblock,
// 2026-11-06, F4 (a)).
//
// THE BAR IS A WRITTEN ROW (#1695 §III.B6): the verbatim edit's transaction holds an ESTOPPEL
// tombstone on the gap whose statement it applied, and the author's next change of that gap's subject
// ends it (tombstone.go). A later verbatim application places it again. A fix spanning several
// sentences carries the token in one, so a reword of another of its sentences leaves the bar standing.

// minEstoppelOverlap is the shortest prescribed text that may trigger the guard, in
// characters after whitespace collapse.
//
// A short prescription ("7 is prime", "38%") can legitimately occur in text red never wrote,
// and refusing a real gap is far worse than missing an estoppel — the guard exists to
// stop a systematic pathology, not to win every edge case. Below this length the mint is
// allowed through and red's own judgment (which the prompt now names explicitly) is the only
// check. Stated so nobody reads the guard as exhaustive.
const minEstoppelOverlap = 40

// appliedVerbatim maps each gap id whose concrete proposal blue applied EXACTLY to the text
// red prescribed, at any edit: what DeclineStatsOf counts as applied — an edit blue took is taken
// whether or not it later reworded the text.
//
// The fact is recorded at edit time by the tool, never by a seat asserting it — TWO WAYS, and
// the difference is one of kind. On an ordinary `blue edit` it is the outcome of comparing the
// bytes blue typed against the ones red prescribed. On `blue edit --accept` the tool supplied
// those bytes from the mint, so it holds by construction and no transcription slip can cost it.
// `BlueEdit.accepted` tells the two apart where that matters; estoppel does not care which.
func appliedVerbatim(evs []*Event, gaps map[string]*Gap) map[string]string {
	out := map[string]string{}
	for _, e := range evs {
		ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](e)
		// The bench's application of red's closing text on a for-raiser answer is no offer blue took.
		if !ok || !ed.GetAppliedVerbatim() || seatclass.RoleOfSeat(e.GetSeatId()) == "bench" {
			continue
		}
		if g, ok := gaps[ed.GetAnswers()]; ok && g.Mint != nil && g.Mint.GetFixNew() != "" {
			out[ed.GetAnswers()] = g.Mint.GetFixNew()
		}
	}
	return out
}

// Estoppel is a prescription a mint's quote relitigates: the gap whose statement the author applied,
// its minter — the lens that prescribed it — and the text the placing act wrote.
type Estoppel struct{ Gap, PrescribedBy, Text string }

// EstoppelConflict reports the applied prescription seat's `quote` relitigates, and whether there is
// one: a gap seat minted that an ESTOPPEL tombstone stands on at the end of f.Events, whose placing
// act's applied text the quote meets.
//
// ESTOPPEL BINDS ONLY THE PRESCRIBING LENS (#1695 X11, picked 2026-11-23): its own words, applied by
// the author, are what it may not re-raise. Any other lens audits that text as it audits any of
// blue's, and mints an ordinary gap there.
//
// THE TEXT IS THE PLACING ACT'S: the author's application wrote it (the accepting edit's new span),
// whether the statement it took was the mint's prescription or a pair a later closing of the raiser's
// stated, so a closing's pair bars as a mint's does.
//
// Containment either way, on whitespace-collapsed text: the new mint may quote a fragment
// of the applied text or a span that swallows it whole, and both are the same act.
func EstoppelConflict(f Family, seat, quote string) (Estoppel, bool) {
	// VISIBLE, NOT MERELY COLLAPSED. The applied text carries the gap's token and every anchor the
	// edit kept, while `quote` is taken from the report as a reader sees it; on raw bytes this check
	// goes blind at the instant it becomes relevant.
	q := anchortext.Visible(quote)
	if len(q) == 0 {
		return Estoppel{}, false
	}
	minters := mintersOf(f.Events)
	applied := estoppelTextsIn(f.Events)
	for _, id := range slices.Sorted(maps.Keys(applied)) {
		if minters[id] != seat {
			continue
		}
		text := applied[id]
		p := anchortext.Visible(text)
		if len(p) < minEstoppelOverlap && len(q) < minEstoppelOverlap {
			continue
		}
		if strings.Contains(p, q) || strings.Contains(q, p) {
			return Estoppel{Gap: id, PrescribedBy: seat, Text: p}, true
		}
	}
	return Estoppel{}, false
}

// estoppelTextsIn is each gap an estoppel tombstone stands on at the end of evs, with the text its
// placing act wrote: the edit the placing row follows in its transaction (writeTombstones writes the
// act, then its rows), its new span.
func estoppelTextsIn(evs []*Event) map[string]string {
	out := map[string]string{}
	var act *Event
	for _, e := range evs {
		t, ok := recordpb.BodyAs[*recordpb.Tombstone](e)
		if !ok {
			act = e
			continue
		}
		if t.GetKind() != recordpb.TombstoneKind_TOMBSTONE_KIND_ESTOPPEL {
			continue
		}
		if t.GetEnds() {
			delete(out, t.GetGapId())
			continue
		}
		out[t.GetGapId()] = ""
		if ed, ok := recordpb.BodyAs[*recordpb.BlueEdit](act); ok && ed.GetAnswers() == t.GetGapId() {
			out[t.GetGapId()] = ed.GetNew()
		}
	}
	return out
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
// does not back its sentence — the gap anchors — placed at its content end, where Attach
// places an anchor on any quote ending there; the run's citations and computations are the edit's to carry
// or refuse. Otherwise the recorded pair.
//
// A GAP'S ANCHOR RIDES WITH THE GAP. A gap's anchor stands on its sentence, and left to the edit,
// that anchor was dropped by red's text and every such accept refused as stale on a report nothing
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

// redTextAppliedAt reports whether an act already did what red's pair road on gapID does: a recorded
// edit replaced exactly the words the road quotes with exactly the words it states, anchors aside, and
// the report still holds those words at the gap's place — an occurrence of fix whose span or abutting
// anchor run holds the gap's token. The freeze admits such an edit, and the answer for red then applies
// nothing more (gblock 2026-11-23, as for a retire road). An edit that rewords the span otherwise, or
// adds words to red's, matches neither half and is held.
func redTextAppliedAt(q recordsql.Querier, report, gapID, loc, fix string) (bool, error) {
	if strings.TrimSpace(fix) == "" {
		return false, nil
	}
	tok := anchor.Token(gapID)
	stands := false
	for _, sp := range anchortext.LocateAll(report, fix) {
		after := strings.TrimLeft(report[sp[1]:], anchortext.TrailingPunct)
		if strings.Contains(report[sp[0]:sp[1]], tok) || strings.Contains(after[:anchor.SkipRun(after, 0)], tok) {
			stands = true
			break
		}
	}
	if !stands {
		return false, nil
	}
	spans, err := editSpansAt(q)
	if err != nil {
		return false, err
	}
	words := func(s string) string { return strings.TrimSpace(anchor.StripWritten(s, nil)) }
	for _, e := range spans {
		if words(e.Old) == words(loc) && words(e.New) == words(fix) {
			return true, nil
		}
	}
	return false, nil
}

// RedTextApplied is redTextAppliedAt asked of run's record as it renders now.
func RedTextApplied(run Run, gapID, loc, fix string) (bool, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return false, err
	}
	report, err := RenderAt(db)
	if err != nil {
		return false, err
	}
	return redTextAppliedAt(db, report, gapID, loc, fix)
}

// RedTextApplied is record.RedTextApplied asked of the landing record, over report as it renders there.
func (l *Landing) RedTextApplied(report, gapID, loc, fix string) (bool, error) {
	return redTextAppliedAt(l.w.q, report, gapID, loc, fix)
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

// AnchorRemoval is the act that took an anchor out of the report (anchor_removal's first row for it):
// its event, By — "retire" for a retire, which keeps the anchor's sentence, or "edit" for a cut, which
// took the sentence with it — the act's reason, and for a retire the sentence it kept. A cut's
// Sentence is empty: the sentence left, and no field stores it.
type AnchorRemoval struct {
	Event            int64
	By               string
	Reason, Sentence string
}

// RemovedByRetire and RemovedByEdit are anchor_removal's "by": a retire, which keeps the anchor's
// sentence, and a cut, whose edit took the anchor with its sentence.
const (
	RemovedByRetire = "retire"
	RemovedByEdit   = "edit"
)

// AnchorRemovedAt answers where an anchor went when it is no longer in the report: the act that took
// it out, a retire off a sentence that stays or a cut that took the sentence. found is false for an
// anchor no act took out — then it is a stale reference or another run's id, and the caller says so.
func AnchorRemovedAt(run Run, id string) (r AnchorRemoval, found bool, err error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return r, false, err
	}
	return anchorRemovedAt(db, id)
}

// anchorRemovedAt is AnchorRemovedAt asked of q — on the write path the writing transaction, so a
// check that refuses an act on an anchor out of the report and the act racing it cannot both land.
func anchorRemovedAt(q recordsql.Querier, id string) (r AnchorRemoval, found bool, err error) {
	var why, sentence sql.NullString
	found, err = queryRowAt(q, []any{&r.Event, &r.By, &why, &sentence},
		`SELECT ar."event_id", ar."by", COALESCE(r."reason", b."text"), r."claim"
		   FROM "anchor_removal" ar
		   LEFT JOIN "retire" r ON ar."by" = 'retire' AND r."event_id" = ar."event_id"
		   LEFT JOIN "blue_edit" b ON ar."by" = 'edit' AND b."event_id" = ar."event_id"
		  WHERE ar."anchor_id" = ? ORDER BY ar."event_id" LIMIT 1`, id)
	r.Reason, r.Sentence = why.String, sentence.String
	return r, found, err
}

// Said is the removal as a refusal names it: the act, its event and its reason, and for a retire the
// sentence it kept.
func (r AnchorRemoval) Said() string {
	if r.By == RemovedByEdit {
		return fmt.Sprintf("blue's edit at event %d cut its sentence, and the anchor left with it (%s)", r.Event, reasonOf(r.Reason))
	}
	return fmt.Sprintf("blue retired it at event %d (%s; its sentence %s kept)", r.Event, reasonOf(r.Reason), quoted(r.Sentence))
}

// editSpan is one recorded blue_edit's (old, new) pair, in record order.
type editSpan struct{ Old, New string }

// editSpansAt returns every recorded blue_edit's (old, new) on q, in record order — what
// redTextAppliedAt reads to learn whether an edit already applied red's pair.
func editSpansAt(q recordsql.Querier) ([]editSpan, error) {
	if noRecord(q) {
		return nil, nil
	}
	rows, err := q.Query(`SELECT "old", "new" FROM "blue_edit" ORDER BY "event_id"`)
	if err != nil {
		return nil, fmt.Errorf("reading the recorded edits: %w", err)
	}
	defer rows.Close()
	var out []editSpan
	for rows.Next() {
		var o, n sql.NullString
		if err := rows.Scan(&o, &n); err != nil {
			return nil, err
		}
		out = append(out, editSpan{Old: o.String, New: n.String})
	}
	return out, rows.Err()
}
