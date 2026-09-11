package claimcount

import "strings"

// Paragraphs counts the PROSE PARAGRAPHS in report markdown — the unit a lens reading the whole
// report scales its mint budget by (record's per-area table).
//
// THE RULE. A paragraph is a block of lines between blank lines (a line that is empty once
// whitespace is trimmed). A block counts when at least one of its lines carries prose: a letter or
// a digit once every anchor token is removed (HasProse). Lines that never carry prose, whatever
// they contain:
//
//   - headings ("#" at the head of the trimmed line) — a title is not a paragraph;
//   - anchor-only lines ("<!--cite:c-1-->", "- <!--fx:f-2-->") — what a cut leaves, not text;
//   - footnote definitions ("[^L1]: https://...") — the bibliography, not the report;
//   - fenced code, fence lines included — a blank line INSIDE a fence does not end the block, so a
//     code block is part of the paragraph it sits in and never one on its own.
//
// A list or a table is one block, and counts once, when it has no blank lines between its items;
// a loose list counts per item, because each item is then a block. Same exclusions as Scan,
// read per block rather than per sentence.
func Paragraphs(md string) int {
	n := 0
	inFence, prose := false, false
	end := func() {
		if prose {
			n++
		}
		prose = false
	}
	for _, ln := range strings.Split(md, "\n") {
		if fenceLine.MatchString(ln) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		t := strings.TrimSpace(ln)
		switch {
		case t == "":
			end()
		case strings.HasPrefix(t, "#"), footnoteDef.MatchString(ln):
		case HasProse(t):
			prose = true
		}
	}
	end()
	return n
}
