// Package claimcount computes a blue report's claim_count deterministically.
//
// THE PROBLEM THIS SOLVES. claim_count — "the number of FOOTNOTED declarative
// claims" — sizes red's per-sitting citation dispatch (the first sitting ceil(claims/40),
// later sittings on the delta) and arms the capture-side retire-vs-drop detector,
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
// cited claim lowers the count by exactly the citations that left with it, the
// property the retire-vs-drop detector rests on.
//
// THE RULE. The count is the number of tool-inserted citation anchors
// ("<!--cite:c-<hex>-->") ATTACHED to prose — some prose before them in their
// sentence, since every inserter places the anchor after the sentence it backs. An
// anchor with nothing before it — what an edit leaves when it cuts a cited sentence
// away, because an edit may carry an anchor but never drop one — is BARE, not a
// claim: counting it would let the sentence leave while the count stood still, so
// a gutted claim read as no loss and a retire of it cancelled some other, real one
// (see BareAnchorIDs for why "bare" cannot mean "alone in its sentence", and for
// markdown around the terminator). The citation axis replaced the
// hand-typed "[^label]" footnote as the claim unit: citations are tool-managed, so
// what a report CITES is exactly what it ANCHORS, and counting the anchor counts the
// backed claim.
//
// PER CITATION, NOT PER SENTENCE. The count used to be the number of SEGMENTS carrying
// an anchor, so a sentence citing two sources counted once. That made the unit the
// detector compares against retires the wrong one twice over: merging two cited
// sentences into one — carrying both anchors, which an edit must — lowered the count
// and read as unrecorded loss that nothing could clear, though no citation left and
// retire exists so prose may be merged freely; and a retire that took two cited
// sentences out at once named two citation anchors against a drop that might be one
// or two. Counted per attached citation, a merge moves nothing, and every citation
// anchor a retire takes out is exactly one unit of the fall it explains.
//
// Only a kind whose row in the anchor kinds table says so counts — a citation. The claim
// unit is a sentence as anchor.Sentences splits one: a list item is its own block, so a cited
// list emits one claim per item, and a sentence soft-wrapped over two lines is one claim. Excluded,
// because none is a declarative claim: fenced code, footnote definitions ("[^L1]: https://..."),
// and headings. NOTE: this counts the PRE-assembly report, whose citations are
// invisible anchors; the visible [^N] footnotes exist only after assembly weaves
// them, and nothing counts the assembled report.
//
// ONE SCANNER. Scan walks the report once and yields the KEPT segments (exclusions
// applied), each with its text and the distinct citation labels it carries. Count and
// BareAnchorIDs both build on Scan, so their exclusion set cannot drift. Count = the
// number of attached labels across segments; BareAnchorIDs = the anchors no segment
// attaches to prose.
package claimcount

import (
	"unicode"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// Segment is one kept sentence of the report — one that survived the exclusions — with the
// DISTINCT citation labels it carries.
type Segment struct {
	Text   string   // the sentence's raw text, which may span lines
	Labels []string // distinct citation ids ATTACHED in this sentence — after some prose — first-seen order
}

// Scan walks report markdown once and returns the kept sentences in reading order. Fenced
// code, footnote definitions (the bibliography) and headings are excluded from the claim
// stream. This is the single source of the exclusion rule; Count and BareAnchorIDs both
// consume it so they cannot disagree about what is a claim.
func Scan(md string) []Segment {
	var segs []Segment
	for _, b := range anchor.Blocks(md) {
		switch b.Kind {
		case anchor.Fence, anchor.FootnoteDef, anchor.Heading:
			continue
		}
		for _, sp := range b.Sentences {
			text := md[sp[0]:sp[1]]
			segs = append(segs, Segment{Text: text, Labels: segmentLabels(text)})
		}
	}
	return segs
}

// Count returns the number of cited claims in a blue report's markdown: the citation
// anchors attached to prose. Reproducible and monotonic over perfect — see the package
// doc. It is exactly the total of every Scan segment's Labels.
func Count(md string) int {
	n := 0
	for _, s := range Scan(md) {
		n += len(s.Labels)
	}
	return n
}

// HasProse reports whether a segment says anything besides its anchors: a letter or a digit
// once every anchor token, of every kind in the table, is removed. Whitespace, list markers and
// punctuation are not prose — "- <!--fx:f-1-->?" is an emptied bullet, not a sentence.
func HasProse(seg string) bool {
	for _, r := range StripAnchors(seg) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// StripAnchors removes every anchor token, of every kind in the table, leaving the prose around
// them.
func StripAnchors(s string) string {
	return anchor.Replace(s, func(string, string) string { return "" })
}

// AN ANCHOR BACKS THE PROSE BEFORE IT. Every inserter places its token flush after the last
// content character of the sentence it anchors, so an anchor with no prose ahead of it in its
// sentence backs nothing — it is BARE. That is the shape an edit leaves when it cuts an anchored
// sentence away (an edit may carry an anchor but never drop one), and it is NOT always alone in
// its sentence: the splice tidy removes the cut sentence's orphaned terminator, so "X<c>. Y." cut
// down to its anchor renders "<c> Y." — the bare anchor now sits at the head of the NEXT
// sentence. Reading "the sentence has prose" would count Y as cited by the claim that just left.
//
// Markdown around the terminator needs no rule here: the closers after a terminator, and an
// anchor flush after them, belong to the sentence they end ("**Water is wet.**<c>"), while an
// anchor after whitespace or flush after a bare terminator opens the next one ("One. <c>",
// "One.<c> Two."), and a list marker lies outside every sentence ("1) <c>").

// BareAnchorIDs returns the distinct anchor ids, of every kind in the table, that back NO prose
// anywhere they stand: no occurrence has prose before it in its sentence. These are what
// `blue retire` may take out with a claim — an anchor still attached to any prose is never
// bare, and an anchor outside the claim stream (a heading, a fence) is never reported, so the
// answer errs toward keeping an anchor rather than removing one.
func BareAnchorIDs(md string) []string {
	bare := map[string]bool{}
	var order []string
	for _, s := range Scan(md) {
		anchor.Each(s.Text, func(start, _ int, id string) {
			was, seen := bare[id]
			if !seen {
				order = append(order, id)
				was = true
			}
			bare[id] = was && !HasProse(s.Text[:start])
		})
	}
	var out []string
	for _, id := range order {
		if bare[id] {
			out = append(out, id)
		}
	}
	return out
}

// segmentLabels returns the DISTINCT citation labels anchored inline in a sentence, in
// first-seen order. A claim is a sentence carrying a tool-inserted anchor of a kind that counts
// as a claim — a citation (the citation axis replaced the hand-typed "[^label]" footnote as the
// claim unit — citations are tool-managed, so what a report cites is exactly what it anchors).
// The label is the token's id as anchor.Each reads it. A label repeated in one sentence is one
// claim. A BARE label — no prose before it in the sentence — backs
// nothing and is not returned.
func segmentLabels(text string) []string {
	seen := map[string]bool{}
	var out []string
	anchor.Each(text, func(start, _ int, id string) {
		if HasProse(text[:start]) && anchor.CountsAsClaim(id) && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	})
	return out
}

// ProtectedAnchorIDs returns the distinct ids of every anchor present in the report, walking the
// kinds table in its order and each kind first-seen. It is the set the blue-edit lockdown, the
// dropped-marker backstop, the reopened set and the retire tidy all read: an edit may drop no
// anchor of any kind.
func ProtectedAnchorIDs(md string) []string {
	ids := anchor.IDs(md)
	var out []string
	for _, kind := range anchor.Kinds() {
		for _, id := range ids {
			if anchor.Kind(id) == kind {
				out = append(out, id)
			}
		}
	}
	return out
}

// The MissingAnchorIDs / MissingCitationAnchorIDs / MissingProofAnchorIDs /
// MissingProtectedAnchorIDs EXPECTED⊄PRESENT checks lived here — they backed the scorecard's
// dropped_finding_markers and unbacked_citations detectors, the proof-backing audit, and the
// PostToolUse lockdown backstop. Under report-as-record (#709) the report is replayed from the
// record and places only recorded markers, so EXPECTED⊄PRESENT is 0 by construction; every
// consumer was removed, and these went with them.
