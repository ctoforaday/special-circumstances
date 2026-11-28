package anchortext

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// A SECTION GAP'S TOKEN LANDS WHERE REPLAY PUTS IT (T-B52's placement): the write places it at the end of
// the one heading its reference keys to, and replay, placing the location the write returns at its
// first match, puts it at the same byte. A reference keying no heading, or two, is refused, and so is a
// heading whose line occurs earlier in the report, where replay would put the token.
func TestASectionTokenLandsOnItsHeadingWhereReplayPutsIt(t *testing.T) {
	tok := anchor.Token("G-0000000a")
	doc := "# Title\n\n## Method\n\nWe sampled.\n\n### Method detail\n\nMore.\n\n## Open questions\n\nWhat next.\n"
	placed, loc, err := AttachToHeading(doc, "G-0000000a", "open Questions")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(placed, "## Open questions"+tok+"\n") {
		t.Errorf("the token is not at the heading's end:\n%s", placed)
	}
	replayed, err := InsertAnchor([]byte(doc), loc, tok)
	if err != nil || string(replayed) != placed {
		t.Errorf("replay of location %q places otherwise (%v):\n%s", loc, err, replayed)
	}
	if _, _, err := AttachToHeading(doc, "G-0000000a", "Results"); !errors.Is(err, ErrNoHeading) {
		t.Errorf("a reference keying no heading: %v", err)
	}
	twice := doc + "\n## Method\n\nAgain.\n"
	if _, _, err := AttachToHeading(twice, "G-0000000a", "Method"); !errors.Is(err, ErrHeadingTwice) {
		t.Errorf("a reference keying two headings: %v", err)
	}
	// A DEEPER HEADING WHOSE LINE HOLDS THE SHALLOWER ONE'S, earlier in the report: "### Open questions"
	// holds "## Open questions" as text, and replay's first match is inside it.
	earlier := "# Title\n\n### Open questions about scope\n\nx.\n\n## Open questions\n\ny.\n"
	if _, _, err := AttachToHeading(earlier, "G-0000000a", "Open questions"); !errors.Is(err, ErrHeadingEarlier) {
		t.Errorf("a heading whose line occurs earlier: %v", err)
	}
	// AND WHERE THE HEADING'S WORDS OCCUR EARLIER IN PROSE, the location is the heading's line, `#` run
	// included, so prose never matches it.
	prose := "# Title\n\nOpen questions remain.\n\n## Open questions\n\ny.\n"
	placed, loc, err = AttachToHeading(prose, "G-0000000a", "Open questions")
	if err != nil {
		t.Fatal(err)
	}
	if replayed, _ := InsertAnchor([]byte(prose), loc, tok); string(replayed) != placed || !strings.Contains(placed, "## Open questions"+tok) {
		t.Errorf("the heading's words in earlier prose moved the token:\n%s", replayed)
	}
}

// A SECTION RUNS FROM ITS HEADING TO THE NEXT HEADING AT ITS LEVEL OR ABOVE.
func TestASectionEndsAtTheNextHeadingAtItsLevel(t *testing.T) {
	doc := "## A\n\na.\n\n### A1\n\nb.\n\n## B\n\nc.\n"
	hs := Headings(doc)
	if len(hs) != 3 {
		t.Fatalf("headings %+v", hs)
	}
	if end := hs[0].End(doc); end != strings.Index(doc, "## B") {
		t.Errorf("A's section ends at %d", end)
	}
	if end := hs[1].End(doc); end != strings.Index(doc, "## B") {
		t.Errorf("A1's section ends at %d", end)
	}
	if end := hs[2].End(doc); end != len(doc) {
		t.Errorf("B's section ends at %d", end)
	}
}
