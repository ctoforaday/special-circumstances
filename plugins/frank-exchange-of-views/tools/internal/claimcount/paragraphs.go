package claimcount

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// Paragraphs counts the PROSE PARAGRAPHS in report markdown — the unit a lens reading the whole
// report scales its mint budget by (record's per-area table).
//
// THE RULE. A paragraph is a block of lines between blank lines (a line that is empty once
// whitespace is trimmed). A block counts when at least one of its lines carries prose: a letter or
// a digit once every anchor token is removed (HasProse). Lines that never carry prose, whatever
// they contain:
//
//   - headings ("## Section") — a title is not a paragraph;
//   - anchor-only lines ("<!--cite:c-1-->", "- <!--fx:f-2-->") — what a cut leaves, not text;
//   - footnote definitions ("[^L1]: https://...") — the bibliography, not the report;
//   - fenced code, fence lines included — a blank line INSIDE a fence does not end the block, so a
//     code block is part of the paragraph it sits in and never one on its own.
//
// A list or a table is one paragraph, and counts once, when it has no blank lines between its
// items; a loose list counts per item. The blocks are anchor.Blocks', grouped by blank line —
// the same block reader Scan reads, so the two cannot disagree about what is a fence, a heading or
// a footnote definition.
func Paragraphs(md string) int {
	n, prose, last := 0, false, -1
	for _, b := range anchor.Blocks(md) {
		if last >= 0 && strings.Count(md[last:b.Start], "\n") > 1 {
			if prose {
				n++
			}
			prose = false
		}
		last = b.End
		switch b.Kind {
		case anchor.Heading, anchor.FootnoteDef, anchor.Fence:
		default:
			prose = prose || HasProse(md[b.Start:b.End])
		}
	}
	if prose {
		n++
	}
	return n
}
