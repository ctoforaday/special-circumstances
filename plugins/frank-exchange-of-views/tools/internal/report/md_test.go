package report

import (
	"strings"
	"testing"
)

// The subset is a closed set, and every member of it is here. An unsupported construct must
// degrade to visible text — never to a silent drop, which is the failure mode a renderer
// nobody tests has.
func TestMarkdownSubsetRenders(t *testing.T) {
	md := strings.Join([]string{
		"# Title",
		"",
		"A paragraph with **bold**, _italic_, `code` and a [link](docket.md).",
		"",
		"> a quotation",
		"",
		"| a | b |",
		"|---|---|",
		"| 1 | 2 |",
		"",
		"- first",
		"- second",
		"",
		"1. one",
		"2. two",
		"",
		"```bash",
		"## not a heading",
		"echo 'x' > y",
		"```",
		"",
		"---",
		"",
		"A claim[^1].",
		"",
		"[^1]: the source. https://example.invalid/a",
	}, "\n")

	got := mdToHTML(md, FileReport, anchors{})
	for _, want := range []string{
		`<h1 id="title">Title</h1>`,
		"<strong>bold</strong>", "<em>italic</em>", "<code>code</code>",
		`<a href="docket.md">link</a>`,
		"<blockquote>a quotation</blockquote>",
		"<table><thead><tr><th>a</th><th>b</th>", "<td>1</td>",
		"<ul><li>first</li>", "<ol><li>one</li>",
		`<pre><code class="language-bash">## not a heading`,
		"echo &#39;x&#39; &gt; y", // escaped, and NOT parsed as a heading or a quote
		"<hr>",
		// Citations render NUMBERED, in first-citation order, as bracketed links — and the
		// FIRST citation of an id carries the return anchor its References entry links back
		// to; the entry itself closes the loop with a .fnback arrow and keeps the markdown
		// tier's slug as a muted tag. The title is the pre-tap scent the number lost: the
		// slug plus the entry's opening text.
		`<a class="cite" id="fnref-1" href="#fn-1" title="1 — the source. https://example.invalid/a">[1]</a>`,
		`<li id="fn-1">`,
		`<span class="fnlabel" data-nolink>1</span>`,
		`<a class="fnback" href="#fnref-1"`,
		`<h2>References</h2>`,
		`<a href="https://example.invalid/a">`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

// HTML IN THE SOURCE IS ESCAPED, not executed. A seat's prose is model output and a finding's
// text is quoted from a repository; either can contain a tag, and a report that renders it is
// a report that can be written into.
func TestMarkupInProseIsEscaped(t *testing.T) {
	got := mdToHTML("A finding about <script>alert(1)</script> in the parser.\n", FileReport, anchors{})
	if strings.Contains(got, "<script>") {
		t.Errorf("a tag in prose reached the document unescaped:\n%s", got)
	}
	if !strings.Contains(got, "&lt;script&gt;") {
		t.Errorf("the tag was dropped instead of shown:\n%s", got)
	}
}

// A heading that INTRODUCES a record id owns it; every other mention across the set becomes a
// link to it. This is the join a seven-file markdown set cannot express.
func TestRecordIdsLinkToTheirDefinition(t *testing.T) {
	anchor := anchors{}
	docket := mdToHTML("### G-00000001 — the gap\n\nsomething\n", FileDocket, anchor)
	report := mdToHTML("The board still carries G-00000001 at close.\n", FileReport, anchor)

	if anchor["G-00000001"] == "" {
		t.Fatalf("the defining heading did not register an anchor: %v", anchor)
	}
	linkedReport := linkIDs(report, anchor, FileReport)
	if !strings.Contains(linkedReport, `data-doc="docket.md"`) {
		t.Errorf("a cross-document id was not linked:\n%s", linkedReport)
	}
	// AND NOT AT ITS OWN DEFINITION. A heading that links to itself makes every definition
	// read as a reference to somewhere else.
	linkedDocket := linkIDs(docket, anchor, FileDocket)
	if strings.Contains(linkedDocket, `<a class="idref"`) {
		t.Errorf("the id was linked at the site that defines it:\n%s", linkedDocket)
	}
}

// An id inside a code span or an existing link is left ALONE — a blind replace over rendered
// HTML rewrites hrefs and produces a document that links every id to itself.
func TestIdsInsideCodeAndLinksAreNotRewritten(t *testing.T) {
	anchor := anchors{"G-00000001": "docket.md#r1-1"}
	got := linkIDs("<p><code>grep G-00000001</code> and <a href=\"x\">G-00000001</a></p>", anchor, FileReport)
	if strings.Count(got, "idref") != 0 {
		t.Errorf("an id was rewritten inside code or an existing link:\n%s", got)
	}
	// THE SAME ID IN PLAIN TEXT IS LINKED, so the zero above is the code span and the link holding
	// it back and not an id the pattern does not read.
	if got := linkIDs("<p>see G-00000001</p>", anchor, FileReport); strings.Count(got, "idref") != 1 {
		t.Errorf("the id in plain text was not linked:\n%s", got)
	}
}

// Hard breaks survive. The composers use them to keep a gap's four facts on four lines; a
// renderer that drops them re-fuses exactly what they were added to separate.
func TestHardBreaksSurvive(t *testing.T) {
	got := mdToHTML("cache.go:88  \nseverity high  \nrequired_fix: lock\n", FileDocket, anchors{})
	if strings.Count(got, "<br>") != 2 {
		t.Errorf("hard breaks were dropped, so four facts render as one sentence:\n%s", got)
	}
}

// The reader of the rendered tier sees arXiv-style numbers assigned by FIRST CITATION, and the
// References list is reordered to match — entry [1] is the first source the text cites, even
// when the author defined it last. The slugs survive underneath: anchors and tags, both tiers.
func TestCitationsAreNumberedByFirstUse(t *testing.T) {
	md := "First claim[^zulu]. Second claim[^alpha], and zulu again[^zulu].\n\n" +
		"[^alpha]: A, defined first. https://example.invalid/a\n" +
		"[^zulu]: Z, defined second. https://example.invalid/z\n"
	got := mdToHTML(md, FileReport, anchors{})
	for _, want := range []string{
		`<a class="cite" id="fnref-zulu" href="#fn-zulu" title="zulu — Z, defined second. https://example.invalid/z">[1]</a>`,
		`<a class="cite" id="fnref-alpha" href="#fn-alpha" title="alpha — A, defined first. https://example.invalid/a">[2]</a>`,
		// the repeat: same number and scent, no second anchor
		`<a class="cite" href="#fn-zulu" title="zulu — Z, defined second. https://example.invalid/z">[1]</a>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	if zi, ai := strings.Index(got, `<li id="fn-zulu">`), strings.Index(got, `<li id="fn-alpha">`); zi < 0 || ai < 0 || zi > ai {
		t.Errorf("References not reordered to citation order (zulu at %d, alpha at %d):\n%s", zi, ai, got)
	}
}

// EVERY ID THE DOCUMENTS QUOTE LINKS TO ITS DEFINITION: a gap's, a motion's and a finding's.
//
// An id idToken does not read is silent in exactly the way that matters: an unlinked id looks the
// same as an id nobody else mentioned, so the report renders clean while every join to the docket is
// gone. Both halves are asserted because they fail separately — `define` must register the anchor,
// and `linkText` must find it — in a heading and in a list item, which define by separate paths.
// The finding is defined as the documents print it, with its area after the id.
func TestEveryQuotedIDLinksToItsDefinition(t *testing.T) {
	for _, tc := range []struct{ id, definedAs string }{
		{"G-5e10a3c2", "G-5e10a3c2 — eviction races the reader"},
		{"M-0000aaa1", "M-0000aaa1 — a petition to reopen"},
		{"F-7cdcd115", "F-7cdcd115 (dark-side) — what the lens found"},
	} {
		for form, md := range map[string]string{"heading": "### " + tc.definedAs + "\n\nsomething\n", "list item": "- " + tc.definedAs + "\n"} {
			t.Run(tc.id+" in a "+form, func(t *testing.T) {
				anchor := anchors{}
				mdToHTML(md, FileDocket, anchor)
				report := mdToHTML("Blue answered "+tc.id+" in the same round.\n", FileReport, anchor)

				if !strings.HasPrefix(anchor[tc.id], FileDocket+"#") {
					t.Fatalf("%s defined no anchor — the report will not link it anywhere: %v", tc.id, anchor)
				}
				if got := linkIDs(report, anchor, FileReport); !strings.Contains(got, `data-doc="docket.md">`+tc.id+`</a>`) {
					t.Errorf("%s was not linked across documents:\n%s", tc.id, got)
				}
			})
		}
	}
	// AN ID OF A KIND THE ASSEMBLY WEAVES, OR OF ANOTHER SHAPE, DEFINES NOTHING: a citation's id
	// stands in an anchor, and a word that begins like an id is a word.
	for _, text := range []string{"C-0000aaa1 — a source", "G20 — the summit", "G-5e10a3c — seven", "G-5e10a3c2f — nine"} {
		anchor := anchors{}
		mdToHTML("### "+text+"\n", FileDocket, anchor)
		if len(anchor) != 0 {
			t.Errorf("%q defined an anchor: %v", text, anchor)
		}
	}
}
