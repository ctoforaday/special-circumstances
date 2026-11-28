package bluedoc

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// AN EDIT MEETING A SECTION REOPENS ITS TOKEN (T-B45, T-B51's section rows): the section runs from the
// heading holding a section gap's token to the next heading at its level or above, and an edit whose
// span meets it — a subsection's text, the heading's own rewrite, an insertion at its end — lists the
// token; one a byte outside it, or an insertion at its heading's start (which lands in the section
// above), does not.
func TestAnEditMeetingASectionReopensItsToken(t *testing.T) {
	tok := anchor.Token("G-0000000a")
	doc := "# T\n\nIntro.\n\n## A" + tok + "\n\nAlpha one.\n\n### A1\n\nDeep text.\n\n## B\n\nBeta.\n"
	at := func(s string) int { return strings.Index(doc, s) }
	edit := func(start, end int, new string) []string {
		return Reopened(doc, doc[:start]+new+doc[end:], start, end)
	}
	met := func(got []string) bool { return slices.Contains(got, "G-0000000a") }
	cases := []struct {
		name       string
		start, end int
		new        string
		want       bool
	}{
		{"its own text", at("Alpha"), at("Alpha") + len("Alpha"), "Gamma", true},
		{"a subsection's text", at("Deep"), at("Deep") + len("Deep"), "Shallow", true},
		{"the heading's rename", at("## A") + 3, at("## A") + 4, "Aa", true},
		{"an insertion at its end", at("## B"), at("## B"), "Tail.\n\n", true},
		{"the next section's text", at("Beta"), at("Beta") + len("Beta"), "Gamma", false},
		{"the section above", at("Intro"), at("Intro") + len("Intro"), "Opening", false},
		{"an insertion at its heading's start", at("## A"), at("## A"), "Lead.\n\n", false},
	}
	for _, c := range cases {
		if got := edit(c.start, c.end, c.new); met(got) != c.want {
			t.Errorf("%s: reopened %v, want the section token listed = %v", c.name, got, c.want)
		}
	}
	// A TOKEN THE EDIT TAKES is not met: the cut lists it in removed_anchors.
	h := at("## A")
	if got := Reopened(doc, doc[:h]+doc[at("## B"):], h, at("## B")); met(got) {
		t.Errorf("a cut of the whole section lists its token as reopened: %v", got)
	}
	// ONLY A GAP'S TOKEN on a heading is a section's.
	cite := strings.Replace(doc, tok, anchor.Token("C-0000000b"), 1)
	if got := Reopened(cite, strings.Replace(cite, "Alpha", "Gamma", 1), at("Alpha"), at("Alpha")+5); slices.Contains(got, "C-0000000b") {
		t.Errorf("a citation on a heading read as a section's token: %v", got)
	}
}

// A HEADING REWRITE KEEPS ITS SECTION'S TOKEN (T-B52's rename): AutoPlace puts a token that stood on the
// heading back on the rewritten heading, and puts none that stood on prose onto a heading.
func TestAHeadingRewriteKeepsItsSectionsToken(t *testing.T) {
	tok := anchor.Token("G-0000000a")
	if got := AutoPlace("## Open questions"+tok, "## Open questions and limits"); got != "## Open questions"+tok+" and limits" {
		t.Errorf("the section token was not carried onto the rewritten heading: %q", got)
	}
	if got := AutoPlace("Open questions"+tok+" remain.", "## Open questions remain."); strings.Contains(got, tok) {
		t.Errorf("a prose token was placed on a heading: %q", got)
	}
}

// A `#` LINE IN A CODE FENCE IS NO HEADING: it neither ends a section nor holds a section's token, so
// an edit past a fenced `# comment` still meets the section the fence stands in.
func TestAFencedHashLineEndsNoSection(t *testing.T) {
	tok := anchor.Token("G-0000000a")
	doc := "## A" + tok + "\n\n```\n# not a heading\n```\n\nAfter the fence.\n\n## B\n\nBeta.\n"
	at := strings.Index(doc, "After")
	if got := Reopened(doc, strings.Replace(doc, "After", "Past", 1), at, at+len("After")); !slices.Contains(got, "G-0000000a") {
		t.Errorf("an edit after a fenced # line did not meet its section: %v", got)
	}
}
