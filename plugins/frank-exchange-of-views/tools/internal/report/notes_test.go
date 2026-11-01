package report

import (
	"html"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT A SEAT TYPED OPENS NO FOOTNOTE. Both rows that head this table shipped in one run's set: a
// finding naming the report's footnote, printed in the board document, and a log entry quoting a
// regex, printed in the run document.
func TestEscapeNotesMakesEveryNoteOpenerText(t *testing.T) {
	for _, c := range []struct{ name, in, want string }{
		{"a label quoted from another document", "the one footnote [^K5] cites, still types the constants", `the one footnote \[^K5] cites, still types the constants`},
		{"a regex character class", "the form grep -o '[^<>]{0,120}term[^<>]{0,120}', fails", `the form grep -o '\[^<>]{0,120}term\[^<>]{0,120}', fails`},
		{"a definition typed at a line's start", "first\n[^1]: a source nobody cited", "first\n\\[^1]: a source nobody cited"},
		{"two in a row", "[^1][^2]", `\[^1]\[^2]`},
		{"a code span is left alone", "run `grep '[^<>]'` and read [^3]", "run `grep '[^<>]'` and read \\[^3]"},
		{"a fence is left alone", "```\n[^K1] in output\n```\nand [^K1] in prose", "```\n[^K1] in output\n```\nand \\[^K1] in prose"},
		{"one already escaped is left alone", `a \[^K5] and a [^K6]`, `a \[^K5] and a \[^K6]`},
		{"a backslash that is itself escaped does not escape", `a \\[^K5]`, `a \\\[^K5]`},
		{"nothing to escape", "no note here [x] ^", "no note here [x] ^"},
	} {
		got := stdReaders().escapeNotes(c.in)
		if got != c.want {
			t.Errorf("%s:\n  in   %q\n  got  %q\n  want %q", c.name, c.in, got, c.want)
		}
		if again := stdReaders().escapeNotes(got); again != got {
			t.Errorf("%s: escaping twice changes it again: %q", c.name, again)
		}
	}
}

// THE HTML TIER READS THE ESCAPE: the seat's five characters, and no citation link. Unescaped, the
// same words are a reference — which is what the weave's own notes are.
func TestAnEscapedNoteIsTextInTheHTMLTier(t *testing.T) {
	got := inline(`the one footnote \[^K5] cites`, map[string]string{"K5": "K5"}, FileReport)
	if !strings.Contains(got, "the one footnote [^K5] cites") || strings.Contains(got, "fnref") || strings.Contains(got, `\`) {
		t.Errorf("an escaped note did not render as the text the seat typed: %s", got)
	}
	if woven := inline("the norms multiply[^K5].", map[string]string{"K5": "K5"}, FileReport); !strings.Contains(woven, `<sup class="fnref"><a class="idref" data-doc="report.md" href="#fn-K5">K5</a></sup>`) {
		t.Errorf("the weave's own reference no longer renders as one: %s", woven)
	}
}

// A TOOL-COMPOSED BLOCK IS ESCAPED WHERE IT IS COMPOSED, and the report's own text is not: a
// footnote under a label that is not the weave's is the report's to write.
func TestOnlyWhatAssemblyComposesIsEscaped(t *testing.T) {
	says := func(*assembly, Section) string { return "a seat wrote [^K5] here" }
	tool := Section{Heading: "The board", Composer: Tool, compose: says}.block(&assembly{rd: stdReaders()})
	if !strings.Contains(tool, `a seat wrote \[^K5] here`) {
		t.Errorf("a tool-composed block ships a seat's words as a footnote reference:\n%s", tool)
	}
	blue := Section{Heading: "Body", Composer: Blue, compose: says}.block(&assembly{rd: stdReaders()})
	if blue != "a seat wrote [^K5] here" {
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

// A CITATION'S URL IS A LINK'S DESTINATION, NEVER PROSE, AND ITS TITLE IS ESCAPED IN PLACE. A URL and
// a copy's URL are a seat's text as much as a title is, and not prose: written as one, a `[^1]` in it
// is a footnote reference only the weave may write, a `*` or `_` opens emphasis, and the extended
// autolink drops a last `.` or `!` from the link. So each that names a scheme report.html links is
// written `[text](destination)`: the text is the URL as recorded with markdown's bytes escaped, and the
// destination holds the URL's own bytes, only a byte no URL may hold raw where it stands
// percent-encoded. A sub-delimiter (`;`, `*`, `'`, `(`, `)`, `!`, `&`), an underscore, a last `.` and
// an IP literal's brackets stay as written: each names its resource raw, and encoded could name another.
// A backtick is a character reference in the text and encoded in the destination, since one would
// close a span the title opens, so the title's `[^2]` is escaped where it stands. A URL whose scheme
// report.html does not link ships as its text. Each want is bytes cmark-gfm 0.29.0.gfm.13 (`--unsafe
// -e autolink -e table -e strikethrough -e tasklist -e footnotes`) reads with no footnote reference and
// no HTML in the line, linking the URL to the target written here with the URL as recorded for its
// text; report.html is read here, every link of the line's URLs in the note and in the Bibliography.
func TestACitationsLineIsEscapedWhole(t *testing.T) {
	for _, c := range []struct {
		src  record.Source
		line string
		// links are each link report.html makes of the line, as href then text, read as a browser
		// reads the attribute and the text.
		links [][2]string
	}{
		{record.Source{URL: "doi:10.1/a[^1]b", Title: "T", AccessDate: "2026-10-12",
			CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED, CopyURL: "https://copy.invalid/c[^1]d"},
			"T. doi:10.1/a\\[^1\\]b (accessed 2026-10-12) **[READ FROM A RECOVERED COPY: [https://copy.invalid/c\\[^1\\]d](https://copy.invalid/c%5B%5E1%5Dd)]**",
			[][2]string{{"https://copy.invalid/c%5B%5E1%5Dd", "https://copy.invalid/c[^1]d"}}},
		{record.Source{URL: "https://x.co/a?q[^1]=b", Title: "x `[^2]"},
			"x `\\[^2]. [https://x.co/a?q\\[^1\\]=b](https://x.co/a?q%5B%5E1%5D=b)", [][2]string{{"https://x.co/a?q%5B%5E1%5D=b", "https://x.co/a?q[^1]=b"}}},
		{record.Source{URL: "https://example.invalid/a`b", Title: "x `[^2]"},
			"x `\\[^2]. [https://example.invalid/a&#96;b](https://example.invalid/a%60b)", [][2]string{{"https://example.invalid/a%60b", "https://example.invalid/a`b"}}},
		{record.Source{URL: "https://x.co/a_b_*c*~d~'e&amp;f;g!h<i> j.", Title: "T"},
			"T. [https://x.co/a\\_b\\_\\*c\\*\\~d\\~'e&amp;amp;f;g\\!h\\<i\\> j.](https://x.co/a_b_*c*~d~'e&amp;amp;f;g!h%3Ci%3E%20j.)",
			[][2]string{{"https://x.co/a_b_*c*~d~'e&amp;f;g!h%3Ci%3E%20j.", "https://x.co/a_b_*c*~d~'e&amp;f;g!h<i> j."}}},
		{record.Source{URL: "https://x.co/ß(a))", Title: "T"}, "T. [https://x.co/ß(a))](https://x.co/%C3%9F\\(a\\)\\))", [][2]string{{"https://x.co/%C3%9F(a))", "https://x.co/ß(a))"}}},
		{record.Source{URL: "https://x.co/a#b#c", Title: "T"}, "T. [https://x.co/a#b#c](https://x.co/a#b%23c)", [][2]string{{"https://x.co/a#b%23c", "https://x.co/a#b#c"}}},
		{record.Source{URL: "http://[2001:db8::1]/paper", Title: "T"}, "T. [http://\\[2001:db8::1\\]/paper](http://[2001:db8::1]/paper)",
			[][2]string{{"http://[2001:db8::1]/paper", "http://[2001:db8::1]/paper"}}},
		{record.Source{URL: "https://w.invalid/wiki/foo_(bar)", Title: "T",
			CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED, CopyURL: "https://web.archive.org/web/2020id_/https://w.invalid/a;b=c"},
			"T. [https://w.invalid/wiki/foo\\_(bar)](https://w.invalid/wiki/foo_\\(bar\\)) **[READ FROM A RECOVERED COPY: [https://web.archive.org/web/2020id\\_/https://w.invalid/a;b=c](https://web.archive.org/web/2020id_/https://w.invalid/a;b=c)]**",
			[][2]string{{"https://w.invalid/wiki/foo_(bar)", "https://w.invalid/wiki/foo_(bar)"}, {"https://web.archive.org/web/2020id_/https://w.invalid/a;b=c", "https://web.archive.org/web/2020id_/https://w.invalid/a;b=c"}}},
		{record.Source{URL: "https://x.co/a\\b", Title: "T"}, "T. [https://x.co/a\\\\b](https://x.co/a\\\\b)", [][2]string{{"https://x.co/a%5Cb", "https://x.co/a\\b"}}},
	} {
		c.src.Label = "C-00000001"
		got := weaveCitations("Claim<!--cite:C-00000001-->.\n", []record.Source{c.src})
		for _, want := range []string{"[^1]: " + c.line + "\n", "- " + c.line + "\n"} {
			if !strings.Contains(got, want) {
				t.Errorf("%q: a citation's line is not its URLs as links and its title escaped, want %q:\n%s", c.src.URL, want, got)
			}
		}
		if strings.Contains(ship(got), "\\[^1]") {
			t.Errorf("%q: a URL took an escape:\n%s", c.src.URL, ship(got))
		}
		var links [][2]string
		for _, m := range htmlLink.FindAllStringSubmatch(renderOne(got, FileReport, anchors{}, map[string]int{}, nil).html, -1) {
			if !strings.HasPrefix(m[1], "#") {
				links = append(links, [2]string{html.UnescapeString(m[1]), html.UnescapeString(m[2])})
			}
		}
		if want := append(slices.Clone(c.links), c.links...); !slices.Equal(links, want) {
			t.Errorf("%q: report.html links %q, want %q, the note's then the Bibliography's", c.src.URL, links, want)
		}
	}
}

// htmlLink is a link report.html writes: its href and its text.
var htmlLink = regexp.MustCompile(`<a href="([^"]*)"[^>]*>(.*?)</a>`)

// WHAT IS CODE TO ESCAPENOTES IS WHAT IS CODE TO BOTH SHIPPED READINGS (#1580). A reader with the HTML
// grammar ends a comment's block at its line and takes the indented line after it as code; report.html,
// with none, and the shipped .md, which escapes the comment, take it as the paragraph's continuation —
// prose, where an unescaped `[^a]` opens a footnote.
func TestEscapeNotesReadsCodeAsTheShippedReadingsDo(t *testing.T) {
	got := stdReaders().escapeNotes("x\n\n<!-- c -->\n    see [^a]\n")
	if !strings.Contains(got, `see \[^a]`) {
		t.Errorf("a `[^` both shipped readings take as prose is not escaped:\n%s", got)
	}
}

// A FRAGMENT IS READ WHERE IT STANDS. A citation's title and a computation's reason are spliced into the
// middle of a note's line, where they are inline text: read alone as a document, "```[^9] spec" opens
// a fence whose info string holds the `[^9]`, and it was left as code; in place it is prose and opened
// a footnote. A code span the fragment holds whole is code in place too, and is left as it is.
func TestAFragmentIsEscapedAsItReadsInPlace(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"```[^9] spec", "```\\[^9] spec"},
		{"`[^9]`", "`[^9]`"},
		{"a `[^9]` and a [^9]", "a `[^9]` and a \\[^9]"},
		{"    [^9] indented", "    \\[^9] indented"},
		{"# [^9] a heading's shape", "# \\[^9] a heading's shape"},
	} {
		if got := escapeNotesInLine(c.in); got != c.want {
			t.Errorf("escapeNotesInLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
	// Through the two weaves that splice one: the note line holds the escape.
	cited := weaveCitations("A claim<!--cite:C-00000001-->.\n", []record.Source{{Label: "C-00000001", URL: "https://example.invalid/a", Title: "```[^9] spec"}})
	if !strings.Contains(cited, "[^1]: ```\\[^9] spec. [https://example.invalid/a](https://example.invalid/a)") {
		t.Errorf("the citation's note does not escape the title's [^9]:\n%s", cited)
	}
	proved, _ := weaveComputationRefs("A claim<!--computation:K-00000001-->.\n", []record.Computation{{Label: "K-00000001", Reason: "```[^9] spec"}})
	if !strings.Contains(proved, "[^K1]: ```\\[^9] spec — ") {
		t.Errorf("the computation's note does not escape the reason's [^9] as it reads in place:\n%s", proved)
	}
}

// A `[^` IS LEFT ONLY WHERE BOTH TIERS READ CODE. GitHub's extended autolink takes the backticks
// after "www.a.com", so in the shipped .md the `[^1]` that report.html holds in a span is prose, and
// cmark-gfm 0.29.0.gfm.13 (`--unsafe -e autolink -e table -e strikethrough -e tasklist -e footnotes`)
// read the unescaped text with a live footnote reference. Each want is bytes it reads with none, and
// report.html reads the escape in its code span as a backslash and text.
func TestANoteTheTiersReadDifferentlyIsEscaped(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"See www.a.com`` and [^1]``.", "See www.a.com`` and \\[^1]``."},
		{"see www.x.com`` [^1]``", "see www.x.com`` \\[^1]``"},
		{"a `[^9]` and http://x.co/a` [^9]`", "a `[^9]` and http://x.co/a` \\[^9]`"},
	} {
		if got := stdReaders().escapeNotes(c.in); got != c.want {
			t.Errorf("escapeNotes(%q) = %q, want %q", c.in, got, c.want)
		}
		if got := escapeNotesInLine(c.in); got != c.want {
			t.Errorf("escapeNotesInLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// THE REFUSAL READS A LABEL AS BOTH TIERS DO. The report's text is not escaped, so a label the shipped
// .md reads as prose — the extended autolink took the backticks that would have held it — ships as a
// live note colliding with the weave's, and is refused at the write.
func TestTheRefusalReadsALabelAsBothTiersDo(t *testing.T) {
	for _, doc := range []string{"see www.x.com`` [^1]``\n", "See www.a.com`` and [^1]``.\n"} {
		if l, _, found := WovenNoteIn(ParseText(""), ParseText(doc)); !found || l != "[^1]" {
			t.Errorf("WovenNoteIn(ParseText(%q)) = %q, %v; the shipped .md reads a live [^1]", doc, l, found)
		}
	}
}
