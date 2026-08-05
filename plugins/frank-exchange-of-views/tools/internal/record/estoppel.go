package record

import "strings"

// ESTOPPEL — red is bound by the fix it prescribed.
//
// THE PATHOLOGY, MEASURED. In the 2026-08-04 smoke blue made 26 edits; round 1 was nine
// edits, every one additive, +2,856 characters, each answering a gap red had raised. Round
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

// AppliedVerbatim maps each gap id whose concrete proposal blue applied EXACTLY to the text
// red prescribed. Only these gaps estop red.
//
// The fact is recorded at edit time by the tool comparing bytes (`blue edit` sets
// applied_verbatim), never by a seat asserting it.
func AppliedVerbatim(b *Board) map[string]string {
	out := map[string]string{}
	for _, e := range b.Events {
		if e.Type != "blue_edit" || !e.Payload.Bool("applied_verbatim") {
			continue
		}
		id := e.Payload.Str("answers")
		g, ok := b.Gaps[id]
		if !ok || g.Mint == nil {
			continue
		}
		if fixNew := g.Mint.Str("fix_new"); fixNew != "" {
			out[id] = fixNew
		}
	}
	return out
}

// EstoppelConflict reports the gap whose prescribed text `quote` is relitigating, or "".
//
// Containment either way, on whitespace-collapsed text: the new finding may quote a fragment
// of the prescribed sentence or a span that swallows it whole, and both are the same act.
func EstoppelConflict(b *Board, quote string) (gapID, prescribed string) {
	q := collapse(quote)
	if len(q) == 0 {
		return "", ""
	}
	for id, fixNew := range AppliedVerbatim(b) {
		p := collapse(fixNew)
		if len(p) < minEstoppelOverlap && len(q) < minEstoppelOverlap {
			continue
		}
		if strings.Contains(p, q) || strings.Contains(q, p) {
			return id, fixNew
		}
	}
	return "", ""
}

// DeclineStats counts how often blue exercised its right to disagree with a concrete
// proposal — the measurement that falsifies this whole design if it comes back zero.
//
// offered: gaps red gave concrete text for. applied: those blue took verbatim. declined:
// those blue answered with its OWN edit instead, or contested.
//
// IF BLUE NEVER DECLINES, THE MECHANISM IS MANUFACTURING AGREEMENT RATHER THAN EARNING IT.
// Applying is instant and free; a counter-edit costs a round and invites re-audit; a dispute
// needs red's agreement. The cheapest path is always compliance, so a 100% apply rate is the
// same pathology this file exists to fix, inverted — and it is only visible if counted.
//
// A gap red offered text for and blue has not answered at all is neither applied nor
// declined: it is unanswered, and it is left out of both rather than scored as agreement.
func DeclineStats(b *Board) (offered, applied, declined int) {
	verbatim := AppliedVerbatim(b)
	answered := map[string]bool{}
	for _, e := range b.Events {
		switch e.Type {
		case "blue_edit":
			if id := e.Payload.Str("answers"); id != "" {
				answered[id] = true
			}
		case "dispute":
			if id := e.Payload.Str("gap_id"); id != "" {
				answered[id] = true
			}
		}
	}
	for id, g := range b.Gaps {
		if g.Mint == nil || g.Mint.Str("fix_basis") != "verified" {
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

// collapse normalizes runs of whitespace to single spaces so a quote and the prescribed text
// compare on their words, not on how either was wrapped.
func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

// ProposalAppliedVerbatim reports whether this edit is EXACTLY the fix the named gap
// proposed — byte-for-byte on both halves.
//
// Exactness is the whole point. A near-application ("I applied red's fix, roughly") is blue
// authoring, and it must be audited as blue's text; only an identical pair estops red from
// re-arguing what it wrote itself. Whitespace is NOT normalized here for that reason: the
// looser the match, the more of blue's own writing red is barred from auditing.
func ProposalAppliedVerbatim(runDir, gapID, old, new string) (bool, error) {
	b, err := BoardState(runDir)
	if err != nil {
		return false, err
	}
	g, ok := b.Gaps[gapID]
	if !ok || g.Mint == nil {
		return false, nil
	}
	return g.Mint.Str("fix_old") == old && g.Mint.Str("fix_new") == new, nil
}
