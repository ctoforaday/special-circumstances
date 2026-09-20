package record

import (
	"strings"
	"testing"
)

const passageDoc = `# A report

Some preamble before any section.

## Trial division

91 divided by 7 is 13, so 91 is composite. The check is exhaustive to the square root.

### A deeper heading

This subsection belongs to trial division.

## Formal tests

Fermat's little theorem gives a probabilistic answer.

# Another top level

Nothing here belongs to the sections above.
`

// THE PASSAGE IS THE SECTION, NOT THE SENTENCE — that is the whole point. An auditor handed only
// the quoted sentence has to render the report to learn whether the dispute survives its
// surroundings, which is the 34%-of-all-bytes-read this change removes.
func TestAPassageIsTheWholeSectionAroundTheQuote(t *testing.T) {
	p := PassageAround(passageDoc, "91 divided by 7 is 13")
	if !strings.HasPrefix(p, "## Trial division") {
		t.Errorf("the passage does not start at its heading, so the auditor cannot see which section it is:\n%s", p)
	}
	for _, want := range []string{"exhaustive to the square root", "### A deeper heading", "belongs to trial division"} {
		if !strings.Contains(p, want) {
			t.Errorf("the passage drops %q — a deeper heading is INSIDE the section, and the sentence after the\nquote is exactly the context a quoted sentence alone misleads about:\n%s", want, p)
		}
	}
	// AND IT STOPS AT THE NEXT SECTION OF THE SAME LEVEL. A passage that runs on is the report
	// under another name, which would make this change cost bytes instead of saving them.
	if strings.Contains(p, "Fermat") || strings.Contains(p, "Another top level") {
		t.Errorf("the passage runs past its own section into the next:\n%s", p)
	}
}

// A QUOTE THE REPORT DOES NOT HOLD YIELDS NOTHING, never a neighbourhood guessed at. A gap can be
// anchored to something that is not report text, its quote can have been edited away, and a run can
// have no ingested report at all — in each case the auditor reads the report as it does today.
func TestAnUnfindableQuoteHasNoPassage(t *testing.T) {
	for _, c := range []struct{ name, report, quote string }{
		{"quote not present", passageDoc, "a sentence nobody wrote"},
		{"empty quote", passageDoc, "   "},
		{"no report yet", "", "91 divided by 7"},
	} {
		if got := PassageAround(c.report, c.quote); got != "" {
			t.Errorf("%s: want no passage, got %q", c.name, got)
		}
	}
}

// A QUOTE UNDER THE DOCUMENT'S OWN TITLE IS GOVERNED BY IT, and its section is everything the
// title governs. That is correct and it is also the degenerate case: a level-1 heading runs to the
// next level-1 heading, so a claim written directly under the report's title has the whole document
// as its surrounding section, and only the bound keeps it from being one.
//
// It is rare rather than theoretical — measured on the 2026-09-20 report, every one of the 21
// sections a claim actually sits in is level 2 — and the honest behaviour when it does happen is a
// passage that says it was shortened, not a smaller passage that pretends to be whole.
func TestAQuoteUnderTheTitleIsGovernedByTheTitle(t *testing.T) {
	p := PassageAround(passageDoc, "Some preamble")
	if !strings.Contains(p, "Some preamble") {
		t.Fatalf("the preamble quote found no passage:\n%s", p)
	}
	if !strings.HasPrefix(p, "# A report") {
		t.Errorf("the passage does not start at the heading that governs the quote:\n%s", p)
	}
	// It stops at the next heading of the SAME level, which is what bounds it at all.
	if strings.Contains(p, "Another top level") {
		t.Errorf("the passage ran past the next level-1 heading:\n%s", p)
	}
}

// TEXT BEFORE ANY HEADING AT ALL still gets a passage rather than nothing.
func TestAQuoteBeforeAnyHeadingStillGetsAPassage(t *testing.T) {
	doc := "A claim with no heading above it at all.\n\n## Later\n\nElsewhere.\n"
	p := PassageAround(doc, "A claim with no heading")
	if !strings.Contains(p, "A claim with no heading") {
		t.Fatalf("a quote before any heading found no passage:\n%s", p)
	}
	if strings.Contains(p, "Elsewhere") {
		t.Errorf("it ran into the first real section:\n%s", p)
	}
}

// A SECTION PAST THE BOUND SAYS IT WAS SHORTENED. A silently trimmed passage is the
// decontextualized hunk the adversary constitution refuses — an auditor would read it as complete
// and mint against text the rest of the section already answers.
func TestAnOversizedSectionIsMarkedAsShownInPart(t *testing.T) {
	big := "## Huge\n\n" + strings.Repeat("filler sentence that goes on. ", 400) + "THE QUOTED BIT. " +
		strings.Repeat("more filler after it. ", 400)
	p := PassageAround(big, "THE QUOTED BIT.")
	if len(p) > passageLimit+400 {
		t.Errorf("the bound did not hold: passage is %d characters", len(p))
	}
	if !strings.Contains(p, "shown in part") {
		t.Errorf("an oversized passage does not say it was shortened, so it reads as complete:\n%s", p[:200])
	}
	if !strings.Contains(p, "THE QUOTED BIT.") {
		t.Error("the trimmed passage does not contain the quote it was built around")
	}
	if !strings.HasPrefix(p, "## Huge") {
		t.Error("the trimmed passage lost its heading, so the auditor cannot tell which section it is")
	}
}

// A '#' IN PROSE IS NOT A HEADING. `#1091` and a colour `#fff` both appear in research prose, and
// treating either as a section boundary would cut a passage in half at a word.
func TestAHashInProseIsNotASectionBoundary(t *testing.T) {
	doc := "## Real heading\n\nSee #1091 for the argument. The claim stands.\n\n## Next\n\nElsewhere.\n"
	p := PassageAround(doc, "The claim stands.")
	if !strings.Contains(p, "See #1091") {
		t.Errorf("a '#' in prose was read as a heading and truncated the passage:\n%s", p)
	}
}
