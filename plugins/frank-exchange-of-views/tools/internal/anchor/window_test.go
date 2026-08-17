package anchor

import (
	"strings"
	"testing"
)

const windowDoc = `# Corpus report

intro paragraph, above every heading.

## Findings

the first finding paragraph.

the corpus holds 340 records.<!--fx:f-a1b2c3-->

a paragraph after it.

one more after that.

## Method

method text.
`

func TestReadAroundCarriesTheLiveTextAndItsSection(t *testing.T) {
	w, err := ReadAround(windowDoc, "f-a1b2c3", DefaultWindow)
	if err != nil {
		t.Fatal(err)
	}
	if w.Heading != "## Findings" {
		t.Errorf("heading = %q, want %q — a reader's first question is which section this is in, and the "+
			"answer is usually further up than the window reaches", w.Heading, "## Findings")
	}
	got := strings.Join(w.Lines, "\n")
	if !strings.Contains(got, "the corpus holds 340 records.") {
		t.Errorf("the anchored line is not in the window:\n%s", got)
	}
	if !strings.Contains(got, "the first finding paragraph.") || !strings.Contains(got, "a paragraph after it.") {
		t.Errorf("the window does not reach both sides of the anchor:\n%s", got)
	}
	if !strings.Contains(got, "## Findings") {
		t.Errorf("this window happens to reach its heading and should carry it inline too:\n%s", got)
	}
	// The heading is reported SEPARATELY as well, so N keeps meaning what it says rather than
	// silently widening to reach a heading.
	if w.AnchorLine < w.First || w.AnchorLine > w.Last {
		t.Errorf("the anchor line %d is outside the reported range %d-%d", w.AnchorLine, w.First, w.Last)
	}
}

// THE HEADING IS FOUND EVEN WHEN IT IS OUTSIDE THE WINDOW — that is the whole reason it is a
// separate field rather than "whatever the window happened to include".
func TestTheSectionIsFoundBeyondTheWindow(t *testing.T) {
	w, err := ReadAround(windowDoc, "f-a1b2c3", 1)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(w.Lines, "\n"), "## Findings") {
		t.Fatal("a window of 1 should not reach the heading — the fixture no longer tests what this test is for")
	}
	if w.Heading != "## Findings" {
		t.Errorf("heading = %q, want %q from OUTSIDE the window", w.Heading, "## Findings")
	}
}

// A MISSING ANCHOR REFUSES; IT DOES NOT RETURN AN EMPTY WINDOW.
//
// An empty window reads as "the report says nothing here", which is a different fact and a false
// one — the same plausible zero this suite keeps finding.
func TestAMissingAnchorRefusesRatherThanReadingEmpty(t *testing.T) {
	_, err := ReadAround(windowDoc, "f-deadbeef", DefaultWindow)
	if err == nil {
		t.Fatal("a missing anchor produced a window — an empty read is indistinguishable from a section with nothing in it")
	}
	if !strings.Contains(err.Error(), "stale reference") {
		t.Errorf("the refusal does not say what a missing anchor MEANS:\n%v", err)
	}
}

// A DUPLICATE ANCHOR IS A BROKEN INVARIANT, NOT AN AMBIGUOUS READ. `blue edit` refuses a
// duplicate, so two occurrences means the guarantee failed upstream — picking either would hide
// that behind a plausible answer.
func TestADuplicateAnchorRefuses(t *testing.T) {
	doc := windowDoc + "\nand again.<!--fx:f-a1b2c3-->\n"
	_, err := ReadAround(doc, "f-a1b2c3", DefaultWindow)
	if err == nil {
		t.Fatal("a duplicated anchor was read as though one of them were the right one")
	}
	if !strings.Contains(err.Error(), "broken invariant") {
		t.Errorf("the refusal does not name the invariant that broke:\n%v", err)
	}
}

// Citation and proof anchors resolve through the same door — AnchorToken already knows all three
// classes, so nothing here needs to.
func TestEveryAnchorClassResolves(t *testing.T) {
	doc := "## S\n\ncited.<!--cite:c-11-->\n\nproved.<!--proof:p-22-->\n"
	for _, id := range []string{"c-11", "p-22"} {
		if _, err := ReadAround(doc, id, 1); err != nil {
			t.Errorf("anchor class %q does not resolve: %v", id, err)
		}
	}
}
