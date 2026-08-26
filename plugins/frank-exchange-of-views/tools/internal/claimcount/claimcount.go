// Package claimcount computes a blue report's claim_count deterministically AND
// enumerates where each footnoted claim appears (the claim-index).
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
// THE RULE. A claim is a sentence carrying at least one tool-inserted citation
// anchor ("<!--cite:c-<hex>-->"). The citation axis replaced the hand-typed
// "[^label]" footnote as the claim unit: citations are tool-managed, so what a
// report CITES is exactly what it ANCHORS, and counting the anchor counts the
// backed claim. A finding anchor ("<!--fx:f-<hex>-->") is NOT a claim and never
// counts (the regex is cite-prefix-specific). The claim unit is bounded by
// sentence punctuation (. ! ?) OR a line break, so a cited list emits one claim
// per line and a claim spanning two lines counts once. Excluded, because none is a
// declarative claim: fenced code, footnote-DEFINITION lines ("[^L1]: https://..."),
// and headings. NOTE: this counts the PRE-assembly report, whose citations are
// invisible anchors; the visible [^N] footnotes exist only after assembly weaves
// them, and nothing counts the assembled report.
//
// ONE SCANNER, TWO READINGS. Scan walks the report once and yields the KEPT
// segments (exclusions applied) with their position, heading context, and the
// distinct footnote labels each carries. Count and Index both build on Scan, so
// their exclusion set cannot drift. Count = the number of segments with a marker;
// Index groups the markers into per-label occurrences (the claim-index, used by
// blue to locate every site of a claim it is correcting without re-reading the
// whole report). NOTE the reconciliation: Count counts per SEGMENT (a segment with
// >=1 marker = 1), while Index enumerates per DISTINCT LABEL per segment — so
// sum(occurrences) == Count only under per-sentence-unique authoring (one label per
// segment); a segment carrying two distinct labels makes the index exceed Count by
// design, because that one site states two claims.
package claimcount

import (
	"regexp"
	"strings"
)

var (
	footnoteDef = regexp.MustCompile(`^\s*\[\^[^\]]+\]:`) // "[^L1]: https://..." — a footnote definition line
	fenceLine   = regexp.MustCompile("^\\s*(```|~~~)")
)

// splitClaims splits one line into claim segments at sentence punctuation (. ! ?),
// collapsing runs, exactly as the old `[\n.!?]+` regexp split did — EXCEPT that
// punctuation INSIDE an HTML comment is not a boundary. A citation anchor
// "<!--cite:c-1-->" carries a '!' (in "<!--") that is not a sentence end; treating the
// comment as opaque (the same skip annotationLen uses when matching) keeps the anchor
// whole in one segment so segmentLabels can see it. (Line breaks are handled by Scan
// splitting on "\n" before this runs, so only .!? matter here.)
func splitClaims(line string) []string {
	boundary := make([]bool, len(line))
	for i := 0; i < len(line); {
		if strings.HasPrefix(line[i:], "<!--") {
			if j := strings.Index(line[i:], "-->"); j >= 0 {
				i += j + 3 // skip the whole comment: its punctuation is not a boundary
				continue
			}
		}
		switch line[i] {
		case '.', '!', '?':
			boundary[i] = true
		}
		i++
	}
	var segs []string
	var cur []byte
	inRun := false
	for i := 0; i < len(line); i++ {
		if boundary[i] {
			if !inRun {
				segs = append(segs, string(cur))
				cur = cur[:0]
				inRun = true
			}
			continue
		}
		inRun = false
		cur = append(cur, line[i])
	}
	return append(segs, string(cur))
}

// Segment is one kept unit of the report — a sentence/line-bounded piece that
// survived the exclusions — with the position and heading context a reader needs to
// locate it, and the DISTINCT footnote labels it carries.
type Segment struct {
	Text    string   // the segment's raw text
	Line    int      // 1-based line number in the original report where the segment sits
	Heading string   // nearest preceding markdown heading (stripped of leading # and space)
	Labels  []string // distinct citation labels (c-<hex>) anchored inline in this segment, first-seen order
}

// Scan walks report markdown once and returns the kept segments in reading order.
// Fenced code blocks, footnote-DEFINITION lines (the bibliography), and headings are
// excluded from the claim stream — headings still update the heading context for the
// segments that follow. This is the single source of the exclusion rule; Count and
// Index both consume it so they cannot disagree about what is a claim.
func Scan(md string) []Segment {
	var segs []Segment
	heading := ""
	inFence := false
	for i, ln := range strings.Split(md, "\n") {
		if fenceLine.MatchString(ln) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "#") { // a heading: not a claim, but it sets context
			heading = strings.TrimSpace(strings.TrimLeft(t, "#"))
			continue
		}
		if footnoteDef.MatchString(ln) { // the bibliography: a definition, not a claim
			continue
		}
		for _, seg := range splitClaims(ln) {
			segs = append(segs, Segment{Text: seg, Line: i + 1, Heading: heading, Labels: segmentLabels(seg)})
		}
	}
	return segs
}

// Count returns the number of footnoted declarative claims in a blue report's
// markdown. Reproducible and monotonic over perfect — see the package doc. It is
// exactly the number of Scan segments that carry a marker.
func Count(md string) int {
	n := 0
	for _, s := range Scan(md) {
		if len(s.Labels) > 0 {
			n++
		}
	}
	return n
}

// Occurrence is one site a footnoted claim appears: which section, which line, and a
// content hash of the enclosing segment that survives line-number drift (the durable
// locator; line is the convenience pointer).
type Occurrence struct {
	Heading string `json:"heading"`
	Line    int    `json:"line"`
}

// NO `sentence_hash`. It was an FNV-1a of the normalized sentence, emitted here and described as
// "a locator stable under line-number shift, so blue can match an occurrence after edits move
// it" — and nothing could use it for that. Its only reader was its own test; no production code,
// no prompt, no document referenced the field. An AGENT was the intended consumer, and an agent
// cannot compute FNV-1a by hand to compare one, so the durable locator was durable and unusable.
// Anchors are the locator, and they are visible in the text precisely so a seat can carry them.

// LabelOccurrences is one footnoted claim (by its label) and every site it appears.
type LabelOccurrences struct {
	Label       string       `json:"label"`
	Occurrences []Occurrence `json:"occurrences"`
}

// Index enumerates, per footnote label, every site the claim appears — the
// claim-index. Blue queries it to propagate a correction to ALL sites of a claim
// without re-reading the whole report. A claim at N sites resolves to N occurrences;
// labels appear in first-seen order, occurrences in reading order.
func Index(md string) []LabelOccurrences {
	order := []string{}
	byLabel := map[string][]Occurrence{}
	for _, s := range Scan(md) {
		for _, label := range s.Labels {
			if _, seen := byLabel[label]; !seen {
				order = append(order, label)
			}
			byLabel[label] = append(byLabel[label], Occurrence{
				Heading: s.Heading, Line: s.Line,
			})
		}
	}
	out := make([]LabelOccurrences, 0, len(order))
	for _, label := range order {
		out = append(out, LabelOccurrences{Label: label, Occurrences: byLabel[label]})
	}
	return out
}

// segmentLabels returns the DISTINCT citation labels anchored inline in a segment, in
// first-seen order. A claim is now a sentence carrying a tool-inserted "<!--cite:c-…-->"
// anchor (the citation axis replaced the hand-typed "[^label]" footnote as the claim unit —
// citations are tool-managed, so what a report cites is exactly what it anchors). The
// label comes straight from citationMarkerRe's capture, so Index yields the real c-<hex>
// ids, never a mangled "--cite:c-ab--". A label repeated in one segment is one site.
func segmentLabels(seg string) []string {
	ms := citationMarkerRe.FindAllStringSubmatch(seg, -1)
	if len(ms) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, m := range ms {
		label := m[1] // the c-<hex> id captured by citationMarkerRe
		if !seen[label] {
			seen[label] = true
			out = append(out, label)
		}
	}
	return out
}

// findingMarkerRe matches an invisible finding-anchor token "<!--fx:f-<hex>-->" (slice
// 1b). A finding-marker is an HTML COMMENT, not a footnote, so it never touches Count
// or the claim-index; it is extracted here purely for the tampering detector.
var findingMarkerRe = regexp.MustCompile(`<!--fx:(f-[0-9a-f]+)-->`)

// citationMarkerRe matches an invisible citation-anchor token "<!--cite:c-<hex>-->" (the
// bibliography axis, symmetric to the finding marker). Unlike a finding marker it DOES
// bound a claim — a sentence carrying one counts (see the package doc and inlineMarker) —
// and it is the anchor the assembly weaves into a visible [^N]. Here it feeds the
// citation-id extractor behind the bijection detector and the lockdown class-sweep.
var citationMarkerRe = regexp.MustCompile(`<!--cite:(c-[0-9a-f]+)-->`)

// proofMarkerRe matches an invisible proof-anchor token "<!--proof:p-<hex>-->" (#277): the
// sentence this one sits at is backed by a COMPUTATION whose script and output are cached
// under <run>/proofs/<sha256>, and which the auditor re-RUNS rather than re-reads.
//
// It is a third class on the one immortal-anchor mechanism, not a third mechanism. Like a
// finding marker and unlike a citation it does not itself bound a claim for the counter —
// a proof BACKS a claim the prose already makes, and counting it as one would double-count
// a sentence that also carries a cite.
var proofMarkerRe = regexp.MustCompile(`<!--proof:(p-[0-9a-f]+)-->`)

// FindingAnchorIDs returns the distinct finding-anchor ids PRESENT in the report, in
// first-seen order. This is the immortal-marker detector's PRESENT set: an anchored
// finding_id absent from it is a dropped marker (a hard violation). Pure id membership
// over the report text.
func FindingAnchorIDs(md string) []string { return anchorIDs(findingMarkerRe, md) }

// CitationAnchorIDs returns the distinct citation-anchor ids PRESENT in the report, in
// first-seen order — the citation-axis twin of FindingAnchorIDs. It is the PRESENT set
// behind the unbacked_citations detector and the blue-edit lockdown's cite class-sweep:
// under the cite⟺anchor bijection it equals the set of cite-event labels, and any
// divergence (a hand-typed footnote, a tampered anchor) is a real defect.
func CitationAnchorIDs(md string) []string { return anchorIDs(citationMarkerRe, md) }

// ProofAnchorIDs returns the distinct proof-anchor ids PRESENT in the report. A proof is
// evidence a seat produced by running something, so dropping its anchor silently unbacks a
// claim exactly as dropping a citation's would — it joins the protected union below.
func ProofAnchorIDs(md string) []string { return anchorIDs(proofMarkerRe, md) }

// anchorIDs returns the distinct capture-group-1 ids matched by re in md, first-seen
// order — the shared extractor behind both anchor classes so their membership logic
// cannot drift.
func anchorIDs(re *regexp.Regexp, md string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range re.FindAllStringSubmatch(md, -1) {
		if id := m[1]; !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

// ProtectedAnchorIDs returns the distinct ids of BOTH immortal anchor classes present in
// the report — finding markers (f-…) then citation anchors (c-…), each first-seen. The two
// classes never collide (distinct prefixes), so this is the union both the blue-edit
// lockdown and the PostToolUse backstop sweep: an edit may drop NEITHER a finding nor a
// citation, and keying the guard on this union means a future third anchor class is added
// in ONE place, not patched per call site.
func ProtectedAnchorIDs(md string) []string {
	ids := append(FindingAnchorIDs(md), CitationAnchorIDs(md)...)
	return append(ids, ProofAnchorIDs(md)...)
}

// missingFrom returns the ids in `expected` absent from `present` (in expected's order) —
// the shared EXPECTED⊄PRESENT set-diff behind every dropped-anchor check.
func missingFrom(expected, present []string) []string {
	have := map[string]bool{}
	for _, id := range present {
		have[id] = true
	}
	var missing []string
	for _, id := range expected {
		if !have[id] {
			missing = append(missing, id)
		}
	}
	return missing
}

// MissingAnchorIDs returns the ids in `expected` that do NOT appear as FINDING-markers in
// reportMD (in expected's order). It backs the scorecard's dropped_finding_markers detector —
// a marker red anchored that is gone from the report is blue tampering. (Citations have
// their own EXPECTED⊄PRESENT check, MissingCitationAnchorIDs; the PostToolUse backstop
// sweeps both via MissingProtectedAnchorIDs.)
func MissingAnchorIDs(expected []string, reportMD string) []string {
	return missingFrom(expected, FindingAnchorIDs(reportMD))
}

// MissingCitationAnchorIDs returns the ids in `expected` (cite-event labels) absent from the
// report's citation anchors — the citation-axis twin of MissingAnchorIDs, behind the
// unbacked_citations detector.
func MissingCitationAnchorIDs(expected []string, reportMD string) []string {
	return missingFrom(expected, CitationAnchorIDs(reportMD))
}

// MissingProofAnchorIDs returns the ids in `expected` (proof-event labels) absent from the
// report's proof anchors — the PROOF-axis twin of MissingCitationAnchorIDs.
//
// It is the last mile of a road that was already built. ProofAnchorIDs has existed since proof
// anchors joined the protected union, so the hookgate lockdown already refuses a raw write that
// drops one — but nothing ever READ that set to ask the opposite question: did a computation this
// run actually performed reach the document at all? The protection existed without the
// observation, which is the half-state its two siblings were each added to close.
//
// A proof whose anchor is absent is a claim the report makes on its own authority while the
// evidence for it sits in the record and the cache, unreferenced. That is the shape #591 names:
// a sentence framed as measured that resolves to no proof id.
func MissingProofAnchorIDs(expected []string, reportMD string) []string {
	return missingFrom(expected, ProofAnchorIDs(reportMD))
}

// MissingProtectedAnchorIDs returns the ids in `expected` (findings AND cite labels) absent
// from the report's anchors of EITHER class — the union check the PostToolUse lockdown
// backstop runs, so a raw write that drops a finding marker OR a citation anchor is caught.
func MissingProtectedAnchorIDs(expected []string, reportMD string) []string {
	return missingFrom(expected, ProtectedAnchorIDs(reportMD))
}
