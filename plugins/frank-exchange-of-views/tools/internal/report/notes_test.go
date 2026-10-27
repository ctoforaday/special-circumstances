package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// WHAT A SEAT TYPED OPENS NO FOOTNOTE. Both rows that head this table shipped in one run's set: a
// finding naming the report's footnote, printed in the board document, and a log entry quoting a
// regex, printed in the run document.
func TestEscapeNotesMakesEveryNoteOpenerText(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"a label quoted from another document", "the one footnote [^P5] cites, still types the constants", `the one footnote \[^P5] cites, still types the constants`},
		{"a regex character class", "the form grep -o '[^<>]{0,120}term[^<>]{0,120}', fails", `the form grep -o '\[^<>]{0,120}term\[^<>]{0,120}', fails`},
		{"a definition typed at a line's start", "first\n[^1]: a source nobody cited", "first\n\\[^1]: a source nobody cited"},
		{"two in a row", "[^1][^2]", `\[^1]\[^2]`},
		{"a code span is left alone", "run `grep '[^<>]'` and read [^3]", "run `grep '[^<>]'` and read \\[^3]"},
		{"a fence is left alone", "```\n[^P1] in output\n```\nand [^P1] in prose", "```\n[^P1] in output\n```\nand \\[^P1] in prose"},
		{"one already escaped is left alone", `a \[^P5] and a [^P6]`, `a \[^P5] and a \[^P6]`},
		{"a backslash that is itself escaped does not escape", `a \\[^P5]`, `a \\\[^P5]`},
		{"nothing to escape", "no note here [x] ^", "no note here [x] ^"},
	} {
		got := EscapeNotes(c.in)
		if got != c.want {
			t.Errorf("%s:\n  in   %q\n  got  %q\n  want %q", c.name, c.in, got, c.want)
		}
		if again := EscapeNotes(got); again != got {
			t.Errorf("%s: escaping twice changes it again: %q", c.name, again)
		}
	}
}

// THE HTML TIER READS THE ESCAPE: the seat's five characters, and no citation link. Unescaped, the
// same words are a reference — which is what the weave's own notes are.
func TestAnEscapedNoteIsTextInTheHTMLTier(t *testing.T) {
	got := inline(`the one footnote \[^P5] cites`)
	if !strings.Contains(got, "the one footnote [^P5] cites") || strings.Contains(got, "fnref") || strings.Contains(got, `\`) {
		t.Errorf("an escaped note did not render as the text the seat typed: %s", got)
	}
	if woven := inline("the norms multiply[^P5]."); !strings.Contains(woven, `<sup class="fnref"><a href="#fn-P5">P5</a></sup>`) {
		t.Errorf("the weave's own reference no longer renders as one: %s", woven)
	}
}

// A TOOL-COMPOSED BLOCK IS ESCAPED WHERE IT IS COMPOSED, and the report's own text is not: a
// footnote under a label that is not the weave's is the report's to write.
func TestOnlyWhatAssemblyComposesIsEscaped(t *testing.T) {
	says := func(*assembly, Section) string { return "a seat wrote [^P5] here" }
	tool := Section{Heading: "The board", Composer: Tool, compose: says}.block(nil)
	if !strings.Contains(tool, `a seat wrote \[^P5] here`) {
		t.Errorf("a tool-composed block ships a seat's words as a footnote reference:\n%s", tool)
	}
	blue := Section{Heading: "Body", Composer: Blue, compose: says}.block(nil)
	if blue != "a seat wrote [^P5] here" {
		t.Errorf("the report's own text was changed on its way into the set:\n%s", blue)
	}
}

// A NOTE THE WEAVE WRITES CARRIES A SEAT'S WORDS AFTER THE TOOL'S LABEL — a source's title — and
// those words open no second note inside it.
func TestTheWeavesNoteEscapesTheTitleItCarries(t *testing.T) {
	got := weaveCitations("Claim<!--cite:C-00000001-->.\n", []record.Source{
		{Label: "C-00000001", URL: "https://example.invalid/re", Title: "Negated classes: [^a-z] and friends", AccessDate: "2026-10-12"},
	})
	if !strings.Contains(got, "Claim[^1].") || !strings.Contains(got, `[^1]: Negated classes: \[^a-z] and friends`) {
		t.Errorf("the note does not carry its own label unescaped and the title's bracket escaped:\n%s", got)
	}
}

// WHAT IS CODE TO ESCAPENOTES IS WHAT IS CODE TO BOTH SHIPPED READINGS (#1580). A reader with the HTML
// grammar ends a comment's block at its line and takes the indented line after it as code; report.html,
// with none, and the shipped .md, which escapes the comment, take it as the paragraph's continuation —
// prose, where an unescaped `[^a]` opens a footnote.
func TestEscapeNotesReadsCodeAsTheShippedReadingsDo(t *testing.T) {
	got := EscapeNotes("x\n\n<!-- c -->\n    see [^a]\n")
	if !strings.Contains(got, `see \[^a]`) {
		t.Errorf("a `[^` both shipped readings take as prose is not escaped:\n%s", got)
	}
}
