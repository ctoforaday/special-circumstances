package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// EstoppelConflict reports the gap whose prescribed text `quote` is relitigating, or "".
//
// Containment either way, on whitespace-collapsed text: the new finding may quote a fragment
// of the prescribed sentence or a span that swallows it whole, and both are the same act.
func EstoppelConflict(f Family, quote string) (gapID, prescribed string) {
	// VISIBLE, NOT MERELY COLLAPSED. Collapsing whitespace alone leaves the anchors, and the
	// two sides of this comparison differ by exactly that: `fix_new` is red's prose, written before
	// any anchor existed, while `quote` is taken from the report, which anchors the sentence the
	// moment the fix lands. On raw bytes this check goes blind at the instant it becomes relevant.
	q := anchortext.Visible(quote)
	if len(q) == 0 {
		return "", ""
	}
	for id, fixNew := range appliedVerbatim(f.Events, GapsByID(f.Gaps)) {
		p := anchortext.Visible(fixNew)
		if len(p) < minEstoppelOverlap && len(q) < minEstoppelOverlap {
			continue
		}
		if strings.Contains(p, q) || strings.Contains(q, p) {
			return id, fixNew
		}
	}
	return "", ""
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
	var loc, fix sql.NullString
	found, err = queryRow(run, []any{&loc, &fix}, `SELECT "location", "fix_new" FROM "mint" WHERE "gap_id" = ?`, gapID)
	if err != nil || !found {
		return "", "", false, err
	}
	old, fixNew = proposalOver(report, gapID, loc.String, fix.String)
	return old, fixNew, true, nil
}

// proposalOver is a gap's prescribed pair over report. Where the location, located as an edit locates
// it, is followed by an anchor run holding the gap's own anchor — after any trailing punctuation,
// where an edit's abutting-anchor rule reads it — old is the location as the report holds it, through
// that run, then the location's own trailing punctuation: the edit then replaces the sentence with its
// anchors as `show report` prints them. new is fix with each gap anchor of the run it does not carry
// placed at its content end, where Attach places an anchor on any quote ending there; the run's other
// anchors are the edit's to carry or refuse. Otherwise the recorded pair.
func proposalOver(report, gapID, loc, fix string) (string, string) {
	start, end, err := anchortext.LocateOnce(report, loc, anchortext.CrossParagraphs)
	if fix == "" || err != nil {
		return loc, fix
	}
	tail := report[end:]
	after := strings.TrimLeft(tail, anchortext.TrailingPunct)
	run := after[:anchor.SkipRun(after, 0)]
	if !strings.Contains(run, anchor.Token(gapID)) {
		return loc, fix
	}
	trimmed := strings.TrimSpace(loc)
	old := report[start:end+len(tail)-len(after)+len(run)] + trimmed[len(strings.TrimRight(trimmed, anchortext.TrailingPunct)):]
	at := anchortext.ContentEnd(fix)
	var carried strings.Builder
	for _, id := range anchor.IDs(run) {
		if anchor.Kind(id) == anchor.Kind(gapID) && !strings.Contains(fix, anchor.Token(id)) {
			carried.WriteString(anchor.Token(id))
		}
	}
	return old, fix[:at] + carried.String() + fix[at:]
}

// ProposalAppliedVerbatim reports whether the pair a seat typed is EXACTLY the fix the named gap
// proposed over the report it edited — byte-for-byte on both halves, against Proposal's pair.
//
// Exactness is the whole point. A near-application ("I applied red's fix, roughly") is blue
// authoring, and it must be audited as blue's text; only an identical pair estops red from
// re-arguing what it wrote itself. Whitespace is NOT normalized here for that reason: the
// looser the match, the more of blue's own writing red is barred from auditing.
func ProposalAppliedVerbatim(run Run, gapID, report, old, new string) (bool, error) {
	pOld, pNew, found, err := Proposal(run, gapID, report)
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
// value (`recordpb.CheckKind_CHECK_KIND_COMPUTATION`) and the QUESTION now has one home:
// `Gap.NeedsComputation()` in replay.go, which every one of those sites calls. Where a lowercase
// spelling is genuinely needed for a projection it is `recordpb.Spelling`, never a literal.

// ProofAnswers reports whether any recorded proof names this gap in its --answers.
//
// It is the join a `computation` acceptance check closes on. The link is a FIELD checked
// against the board at write time, never a gap id mentioned in prose — the convention
// blue_edit's --answers replaced after it measured 73% reliable, which is reliable enough to
// look like a key and not reliable enough to be one.
func ProofAnswers(run Run, gapID string) bool {
	if gapID == "" {
		return false
	}
	// A read error folds into false, as the board-read error did: estoppel never blocks on an
	// unreadable record.
	found, err := recordHas(run, `SELECT 1 FROM "proof" WHERE "answers" = ? LIMIT 1`, gapID)
	return err == nil && found
}

// RemovalVerified / RemovalAsserted say how far a recorded retirement can be trusted, on the
// same footing as fix_basis, proof_basis and verdict_basis.
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

// ClaimAppearsInAnEdit reports whether a claim's text appears in the OLD span of any recorded
// blue_edit — the record's evidence that the text was in the report and was taken out.
func ClaimAppearsInAnEdit(run Run, claim string) bool {
	if claim == "" {
		return false
	}
	// instr answers presence exactly as strings.Contains did (a NULL old span yields NULL,
	// which is not > 0 — the same miss as Contains on an empty string). A read error folds
	// into false, as the board-read error did.
	found, err := recordHas(run,
		`SELECT 1 FROM "blue_edit" WHERE instr("old", ?) > 0 LIMIT 1`, claim)
	return err == nil && found
}

// EditSpan is one recorded blue_edit's (old, new) pair, in record order.
type EditSpan struct{ Old, New string }

// EditSpans returns every recorded blue_edit's (old, new), in record order — what `blue retire`
// reads to learn which edit took a claim out and what that edit left behind in its place.
func EditSpans(run Run) ([]EditSpan, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "old", "new" FROM "blue_edit" ORDER BY "event_id"`)
	if err != nil {
		return nil, fmt.Errorf("record: reading the recorded edits: %w", err)
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

// AnchorRetiredAt answers where an anchor went when it is no longer in the report: the retire
// event that took it out with its claim. found is false for an anchor no retire named — then it
// is a stale reference or another run's id, and the caller says so.
func AnchorRetiredAt(run Run, id string) (event int64, claim string, found bool, err error) {
	var c sql.NullString
	found, err = queryRow(run, []any{&event, &c},
		`SELECT r."event_id", r."claim" FROM "retire_anchors" ra JOIN "retire" r ON r."event_id" = ra."event_id"
		 WHERE ra."value" = ? ORDER BY r."event_id" LIMIT 1`, id)
	return event, c.String, found, err
}
