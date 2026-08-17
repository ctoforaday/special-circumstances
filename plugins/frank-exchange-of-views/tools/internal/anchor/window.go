package anchor

import (
	"fmt"
	"strings"
)

// READING THE REPORT AROUND AN ANCHOR, RATHER THAN QUOTING A SENTENCE SOMEBODY TYPED ONCE.
//
// # Why the anchor is the address and the line number is not
//
// The report is edited every round. Blue inserting a paragraph shifts every line below it, so a
// line number captured in round 2 points at different text in round 3 — an address derived from a
// rendering, which is the defect class this suite keeps finding in its own surfaces.
//
// The anchor is the stable identity and it is already enforced as one: `blue edit` refuses any edit
// that drops, duplicates or invents an anchor token, so `<!--fx:f-a1b2c3-->` survives rewrites,
// insertions and reflows. Blue MAY rewrite the sentence carrying it — that is transit, not
// authorship — which is exactly the case a frozen quote gets wrong and a live window gets right.
//
// So: the anchor addresses, and line numbers are OUTPUT. A seat can say "I read lines 118–134" and
// a human can follow it, without anything storing a line number as a fact.
//
// # Why lines, and why the heading comes too
//
// In prose markdown a line is a paragraph — the wrapping belongs to the viewer, not the file. So a
// window of N either side is "this paragraph, its neighbours", which is the unit a reader actually
// wants. Characters cut mid-word; sentences need a splitter that is wrong on every abbreviation and
// every citation.
//
// N COUNTS PARAGRAPHS, NOT RAW LINES, and the first cut of this got it wrong. Paragraphs in a
// markdown file are separated by BLANK lines, so a raw window of 3 reaches about one and a half
// paragraphs either way — half the budget spent on empty strings. The window walks outward counting
// non-blank lines and carries the blanks it crossed, so N=3 means three paragraphs of content and
// the rendering still shows the document's real line numbers. Caught by its own test asserting a
// heading three "lines" up was in range when it was not.
//
// The nearest preceding `#`-heading is included even when it falls outside the window, because
// "which section is this in" is the first question a reader has and the answer is usually further
// up than N paragraphs. It is reported separately rather than by widening the window, so N keeps
// meaning what it says.

// Window is a live read of the report around one anchor.
type Window struct {
	// Anchor is the id that addressed this window.
	Anchor string
	// Heading is the nearest preceding markdown heading, "" when the anchor sits above the first
	// one. HeadingLine is its 1-based line number.
	Heading     string
	HeadingLine int
	// First and Last are the 1-based, inclusive line range Lines covers.
	First, Last int
	// Lines is the window itself, in document order.
	Lines []string
	// AnchorLine is the 1-based line the anchor token sits on.
	AnchorLine int
}

// DefaultWindow is how many lines either side of the anchor a read carries when none is asked for.
//
// Three, because a markdown paragraph is a line: the anchored paragraph, what leads into it and
// what follows. Small enough that a board carrying one per gap stays readable, large enough that a
// claim is not read stripped of the sentences that qualify it.
const DefaultWindow = 3

// ReadAround returns the live text around an anchor, or an error naming what it could not find.
//
// IT REFUSES RATHER THAN RETURNING AN EMPTY WINDOW. An anchor that is not in the report is either a
// stale reference or an id from another run, and both are things the caller must be told — an empty
// window would read as "the report says nothing here", which is a different fact and a false one.
func ReadAround(report, anchorID string, n int) (Window, error) {
	if n < 0 {
		return Window{}, fmt.Errorf("anchor: a window of %d lines is not a window", n)
	}
	token := Token(anchorID)
	lines := strings.Split(report, "\n")

	at := -1
	for i, l := range lines {
		if strings.Contains(l, token) {
			if at >= 0 {
				// TWO OCCURRENCES IS NOT A WINDOW, IT IS A BUG UPSTREAM. An anchor is minted once
				// and `blue edit` refuses a duplicate, so a second one means the invariant broke;
				// picking either would hide that behind a plausible answer.
				return Window{}, fmt.Errorf("anchor: anchor %s appears on lines %d AND %d — an anchor is unique by construction "+
					"(`blue edit` refuses one that duplicates), so two occurrences is a broken invariant rather than an ambiguous read",
					anchorID, at+1, i+1)
			}
			at = i
		}
	}
	if at < 0 {
		return Window{}, fmt.Errorf("anchor: anchor %s (%s) is not in the report — it is a stale reference or belongs to another run. "+
			"`show findings` lists every finding anchor; `show evidence` resolves citation and proof anchors",
			anchorID, token)
	}

	// Walk outward counting CONTENT lines; blanks are carried but not charged against N.
	first, last := at, at
	for seen := 0; first > 0 && seen < n; {
		first--
		if strings.TrimSpace(lines[first]) != "" {
			seen++
		}
	}
	for seen := 0; last < len(lines)-1 && seen < n; {
		last++
		if strings.TrimSpace(lines[last]) != "" {
			seen++
		}
	}

	w := Window{
		Anchor:     anchorID,
		First:      first + 1,
		Last:       last + 1,
		AnchorLine: at + 1,
		Lines:      append([]string(nil), lines[first:last+1]...),
	}
	for i := at; i >= 0; i-- {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			w.Heading, w.HeadingLine = strings.TrimSpace(lines[i]), i+1
			break
		}
	}
	return w, nil
}

// Render writes the window the way a seat reads it: the section it lives in, then the lines with
// their numbers, with the anchored line marked.
//
// The line numbers are for SAYING WHAT YOU READ, not for addressing it again later — see the note
// at the top of this file.
func (w Window) Render() string {
	var b strings.Builder
	if w.Heading != "" {
		fmt.Fprintf(&b, "in section: %s   (line %d)\n", w.Heading, w.HeadingLine)
	} else {
		b.WriteString("in section: (none — this is above the first heading)\n")
	}
	fmt.Fprintf(&b, "anchor %s at line %d · showing lines %d-%d\n\n", w.Anchor, w.AnchorLine, w.First, w.Last)
	for i, l := range w.Lines {
		n := w.First + i
		mark := "  "
		if n == w.AnchorLine {
			mark = "> "
		}
		fmt.Fprintf(&b, "%s%4d  %s\n", mark, n, l)
	}
	return b.String()
}
