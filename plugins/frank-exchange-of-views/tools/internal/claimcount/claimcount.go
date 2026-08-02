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
// THE RULE. A claim is a sentence carrying at least one inline footnote marker
// ([^label]). The claim unit is bounded by sentence punctuation (. ! ?) OR a line
// break, so a footnoted list emits one claim per line and a claim spanning two
// lines counts once. Excluded, because none is a declarative claim: fenced code,
// footnote-DEFINITION lines (the bibliography, "[^L1]: https://..."), and headings.
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
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

var (
	footnoteDef   = regexp.MustCompile(`^\s*\[\^[^\]]+\]:`) // "[^L1]: https://..." — the bibliography
	inlineMarker  = regexp.MustCompile(`\[\^[^\]]+\]`)      // "[^L1]" used inline
	claimBoundary = regexp.MustCompile(`[\n.!?]+`)          // sentence punctuation OR a line break
	fenceLine     = regexp.MustCompile("^\\s*(```|~~~)")
)

// Segment is one kept unit of the report — a sentence/line-bounded piece that
// survived the exclusions — with the position and heading context a reader needs to
// locate it, and the DISTINCT footnote labels it carries.
//
// FINDING-ANCHOR NAMESPACE (slice 1b). A footnote whose label begins "f-" is NOT a
// blue claim footnote — it is a tool-inserted FINDING-MARKER (the label is the
// finding_id, e.g. "[^f-abc]"). It is routed to Anchors, NOT Labels, so Count (which
// counts Labels) is provably unchanged by markers — the monotonicity the retire-vs-
// drop detector rests on. The reserved rule: a blue claim label must never begin "f-".
type Segment struct {
	Text    string   // the segment's raw text
	Line    int      // 1-based line number in the original report where the segment sits
	Heading string   // nearest preceding markdown heading (stripped of leading # and space)
	Labels  []string // distinct CLAIM footnote labels used inline (never an "f-" finding-anchor)
	Anchors []string // distinct FINDING-ANCHOR ids ("f-…") used inline; a locator, not a claim unit
}

// findingAnchorPrefix marks the reserved namespace: a footnote label beginning with
// it is a finding-marker, not a claim.
const findingAnchorPrefix = "f-"

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
		for _, seg := range claimBoundary.Split(ln, -1) {
			labels, anchors := segmentMarkers(seg)
			segs = append(segs, Segment{Text: seg, Line: i + 1, Heading: heading, Labels: labels, Anchors: anchors})
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
	Heading      string `json:"heading"`
	Line         int    `json:"line"`
	SentenceHash string `json:"sentence_hash"`
}

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
				Heading: s.Heading, Line: s.Line, SentenceHash: sentenceHash(s.Text),
			})
		}
	}
	out := make([]LabelOccurrences, 0, len(order))
	for _, label := range order {
		out = append(out, LabelOccurrences{Label: label, Occurrences: byLabel[label]})
	}
	return out
}

// segmentMarkers returns the DISTINCT inline footnote markers in a segment, PARTITIONED
// by namespace: claim labels (labels) and finding-anchors (anchors, "f-…"). First-seen
// order within each; a marker repeated in one segment is one site, not two. The `f-`
// split is what keeps Count (over labels) unchanged by finding-markers.
func segmentMarkers(seg string) (labels, anchors []string) {
	ms := inlineMarker.FindAllString(seg, -1)
	if len(ms) == 0 {
		return nil, nil
	}
	seen := map[string]bool{}
	for _, m := range ms {
		label := m[2 : len(m)-1] // strip the "[^" prefix and "]" suffix
		if seen[label] {
			continue
		}
		seen[label] = true
		if strings.HasPrefix(label, findingAnchorPrefix) {
			anchors = append(anchors, label)
		} else {
			labels = append(labels, label)
		}
	}
	return labels, anchors
}

// FindingAnchorIDs returns the distinct finding-anchor ids ("f-…") PRESENT in the
// report, in first-seen order. This is the immortal-marker detector's PRESENT set: an
// anchored finding_id absent from it is a dropped marker (a hard violation). Pure id
// membership over the report text — no claim, no Count interaction.
func FindingAnchorIDs(md string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range Scan(md) {
		for _, id := range s.Anchors {
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// sentenceHash is the FNV-1a hash of the whitespace-normalized segment: a locator
// stable under line-number shift, so blue can match an occurrence after edits move
// it. strings.Fields collapses whitespace runs and trims.
func sentenceHash(seg string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.Join(strings.Fields(seg), " ")))
	return fmt.Sprintf("%08x", h.Sum32())
}
