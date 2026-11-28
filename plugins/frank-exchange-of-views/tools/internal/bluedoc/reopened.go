package bluedoc

import (
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
)

// Reopened is what an edit of before into after, replacing before's span [start, end), reopens: the
// anchors whose sentence it changed (ReopenedAnchors over the paragraphs it touched), then each section
// gap's token whose section the span meets (SectionsMet). It is the one derivation every writer of an
// edit's `reopened` stamps, the live edit and the archived one alike.
func Reopened(before, after string, start, end int) []string {
	oldParas, newParas := TouchedParagraphs(before, after, start, end)
	out := ReopenedAnchors(oldParas, newParas)
	for _, id := range SectionsMet(before, after, start, end) {
		if !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}

// SectionsMet is every section gap's token whose section an edit of before into after meets (§III.B0
// "Sections"): a gap token on a heading of before — the one anchor a heading holds — whose section,
// from that heading to the next at its level or above (anchortext.Heading.End), the replaced span
// [start, end) meets, the heading's own rewrite included, and which after still holds. An edit in a
// subsection meets every section enclosing it. A pure insertion meets the section it lands in: one at
// a heading's start lands before that heading, in the section above, and one at a section's end lands
// in that section. A token after no longer holds is not met: a cut took it (removed_anchors), or the
// edit dropped it, which the caller has refused.
func SectionsMet(before, after string, start, end int) []string {
	var out []string
	for _, h := range anchortext.Headings(before) {
		stop := h.End(before)
		met := start < stop && end > h.Start
		if start == end {
			met = h.Start < start && start <= stop
		}
		if !met {
			continue
		}
		for _, id := range anchor.IDs(before[h.Start:h.Stop]) {
			if anchor.Kind(id) == "gap" && strings.Contains(after, anchor.Token(id)) && !slices.Contains(out, id) {
				out = append(out, id)
			}
		}
	}
	return out
}
