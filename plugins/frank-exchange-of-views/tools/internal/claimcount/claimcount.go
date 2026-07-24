// Package claimcount computes a blue report's claim_count deterministically.
//
// THE PROBLEM THIS SOLVES. claim_count — "the number of FOOTNOTED declarative
// claims" — sizes red's per-round citation dispatch (round 1 ceil(claims/40),
// later rounds on the delta) and arms the capture-side retire-vs-drop detector,
// where an unaccounted FALL in the count against the retire events is the whole
// enforcement. Until now it was hand-counted by the blue LLM against a prose rule
// in two prompts and typed into the envelope; two honest merges diverged 2x on the
// same report. A wrong count mis-scales the one audit dimension measured to work
// and can mask a dropped claim.
//
// So the count moves to a pure function of report.md, invoked through the
// count-claims root command. The point is a RELIABLE count: a subtly wrong
// deterministic one is no better than the divergent human one, so the rule is
// stated precisely and pinned by tests, including MONOTONICITY — removing one
// footnoted claim lowers the count by exactly one, the property the retire-vs-drop
// detector rests on.
//
// THE RULE. A claim is a sentence carrying at least one inline footnote marker
// ([^label]). The claim unit is bounded by sentence punctuation (. ! ?) OR a line
// break, so a footnoted list emits one claim per line and a claim spanning two
// lines counts once. Excluded, because none is a declarative claim: fenced code,
// footnote-DEFINITION lines (the bibliography, "[^L1]: https://..."), and headings.
package claimcount

import (
	"regexp"
	"strings"
)

var (
	footnoteDef   = regexp.MustCompile(`^\s*\[\^[^\]]+\]:`) // "[^L1]: https://..." — the bibliography
	inlineMarker  = regexp.MustCompile(`\[\^[^\]]+\]`)      // "[^L1]" used inline
	claimBoundary = regexp.MustCompile(`[\n.!?]+`)          // sentence punctuation OR a line break
	fenceLine     = regexp.MustCompile("^\\s*(```|~~~)")
)

// Count returns the number of footnoted declarative claims in a blue report's
// markdown. Reproducible and monotonic over perfect — see the package doc.
func Count(md string) int {
	var kept []string
	for _, ln := range strings.Split(stripFences(md), "\n") {
		t := strings.TrimSpace(ln)
		if footnoteDef.MatchString(ln) || strings.HasPrefix(t, "#") { // bibliography + headings
			continue
		}
		kept = append(kept, ln)
	}
	n := 0
	for _, seg := range claimBoundary.Split(strings.Join(kept, "\n"), -1) {
		if inlineMarker.MatchString(seg) {
			n++
		}
	}
	return n
}

// stripFences removes fenced code blocks: a footnote-looking marker inside a code
// sample is a literal, not a claim, and would inflate the count.
func stripFences(md string) string {
	var out []string
	fence := false
	for _, ln := range strings.Split(md, "\n") {
		if fenceLine.MatchString(ln) {
			fence = !fence
			continue
		}
		if !fence {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}
