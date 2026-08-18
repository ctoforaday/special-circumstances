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
	// THE SPAN IS THE GAP'S OWN `location`. It was a separate `fix_old` holding the same
	// sentence, matched by a second matcher — a gap's location and the span its proposal
	// replaces were never two facts.
	return g.Mint.Str("location") == old && g.Mint.Str("fix_new") == new, nil
}

// FrictionKindEstoppel marks a friction event as an estoppel REFUSAL rather than a seat's
// own complaint. It is the single symbol the writer and the counter share.
//
// THE DEFECT IT REPLACES (#283, and it was mine). The count was derived by matching the
// prose substring "estoppel —" in the friction text. That reads ZERO both when the guard
// never fires and when someone rewords the message — and zero was printed as "red behaved".
// A measurement whose miss is indistinguishable from its healthy answer is not a
// measurement; see [[facts-are-fields]]. The fact is now a FIELD, written by the tool at the
// moment it refuses.
const FrictionKindEstoppel = "estoppel"

// FrictionKindToolError marks a friction event as a TOOL FAULT — unparseable input, an
// undecodable row, a check that could not run — rather than a seat reporting a capability it
// lacked.
//
// It is a distinct kind for the reason FrictionKindEstoppel is: friction is otherwise a SEAT's
// report, and folding tool faults into that count would move a number an operator reads for a
// reason unrelated to what it measures. Filter by kind; the counts stay honest.
//
// It exists because an error nobody learns about is one nothing improves on. A tool that cannot
// parse its own input either prints somewhere unread or says nothing; this is the third option,
// and it is durable where a hook's reason string is not.
const FrictionKindToolError = "tool_error"

// FrictionKindKey is the payload key carrying the kind.
const FrictionKindKey = "kind"

// EstoppelRejections counts the mints refused because their location was text red itself
// prescribed and blue applied verbatim.
//
// It counts the FIELD, never the message. Rewriting the refusal's wording — which is prose
// aimed at a seat and should stay editable — must not move a number an operator reads as
// evidence about red's behaviour.
func EstoppelRejections(b *Board) int {
	n := 0
	for _, e := range b.Events {
		if e.Type == "friction" && e.Payload.Str(FrictionKindKey) == FrictionKindEstoppel {
			n++
		}
	}
	return n
}

// CheckKindComputation is the acceptance-check kind that prose cannot settle. It was compared
// as a bare literal in the close gate and in two projections, which is three chances to disagree
// about the one rule that decides whether a gap can close at all.
const CheckKindComputation = "computation"

// proofNames is ProofAnswers against a board the caller already has. The projections run it per
// open gap, and re-reading the whole record each time would make a board render quadratic in its
// own size.
func proofNames(b *Board, gapID string) bool {
	if gapID == "" || b == nil {
		return false
	}
	for _, e := range b.Events {
		if e.Type == "proof" && e.Payload.Str("answers") == gapID {
			return true
		}
	}
	return false
}

// ProofAnswers reports whether any recorded proof names this gap in its --answers.
//
// It is the join a `computation` acceptance check closes on. The link is a FIELD checked
// against the board at write time, never a gap id mentioned in prose — the convention
// blue_edit's --answers replaced after it measured 73% reliable, which is reliable enough to
// look like a key and not reliable enough to be one.
func ProofAnswers(runDir, gapID string) bool {
	if gapID == "" {
		return false
	}
	b, err := BoardState(runDir)
	if err != nil {
		return false
	}
	return proofNames(b, gapID)
}

// RemovalVerified / RemovalAsserted say how far a recorded retirement can be trusted, on the
// same footing as fix_basis, proof_basis and verdict_basis.
//
// A retire used to record whatever it was told: nothing confirmed the claim had ever been in
// the report, and nothing confirmed it had left. That mattered beyond tidiness — the
// scorecard's additive-integrity detector computes unrecorded_claim_loss as the drop in
// claim_count MINUS the retire events, so a retirement of a claim that was never there
// subtracts from the accounted side and cancels real loss, blinding the one detector built to
// catch silent deletion.
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
func ClaimAppearsInAnEdit(runDir, claim string) bool {
	if claim == "" {
		return false
	}
	b, err := BoardState(runDir)
	if err != nil {
		return false
	}
	for _, e := range b.Events {
		if e.Type == "blue_edit" && strings.Contains(e.Payload.Str("old"), claim) {
			return true
		}
	}
	return false
}

// GapsAwaitingProof lists the OPEN gaps minted --check-kind computation that no proof answers,
// in board order.
//
// It is the same join the close gate and the board projection use, exposed as a debt a seat can
// be handed at the moment it matters. The gate is at the MERGE's close, one seat and one round
// after blue — the only seat that can discharge it — has finished its sitting.
//
// Measured: projecting check_kind moved `prove` from 0 uses across eighteen sittings to 1 across
// nine. A seat reading "check_kind: computation" learns a property of the gap; it does not learn
// that it owes a program, and only the second changes what the sitting produces.
func GapsAwaitingProof(runDir string) []string {
	b, err := BoardState(runDir)
	if err != nil {
		return nil
	}
	var out []string
	for _, id := range b.GapOrder {
		g := b.Gaps[id]
		if g == nil || !g.Open || g.Mint == nil {
			continue
		}
		if g.Mint.Str("check_kind") == CheckKindComputation && !proofNames(b, id) {
			out = append(out, id)
		}
	}
	return out
}
