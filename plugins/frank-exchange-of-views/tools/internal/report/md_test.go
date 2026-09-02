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
		// The FIRST citation of an id carries the return anchor its References entry links
		// back to; the entry itself closes the loop with a .fnback arrow.
		`<sup class="fnref" id="fnref-1"><a href="#fn-1">1</a></sup>`,
		`<li id="fn-1">`,
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
	docket := mdToHTML("### R4-3 — the gap\n\nsomething\n", FileDocket, anchor)
	report := mdToHTML("The board still carries R4-3 at close.\n", FileReport, anchor)

	if anchor["R4-3"] == "" {
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
	anchor := anchors{"R1-1": "docket.md#r1-1"}
	got := linkIDs("<p><code>grep R1-1</code> and <a href=\"x\">R1-1</a></p>", anchor, FileReport)
	if strings.Count(got, "idref") != 0 {
		t.Errorf("an id was rewritten inside code or an existing link:\n%s", got)
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
