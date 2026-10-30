package report

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// REPORT.HTML SHOWS THE TEXT'S HTML AS TYPED (#1569). The renderer passes no HTML of the text
// through, so nothing a seat wrote is an element; the one element the tool places that markdown
// cannot say, a computation's anchor, is a node only evidenceDoc's mark produces; and an allowlist of what
// the renderer writes holds every body after it. These tests read the page with an HTML tokenizer —
// what a browser builds — never with the markdown parser the renderer reads.

// viewerAllows is every element and attribute a document body may hold, written down here rather
// than read from bodyPolicy, so a policy widened by mistake is caught rather than agreed with.
var viewerAllows = map[string][]string{
	"p": nil, "br": nil, "hr": nil, "pre": nil, "blockquote": nil, "ul": nil, "em": nil, "strong": nil, "del": nil,
	"table": nil, "thead": nil, "tbody": nil, "tr": nil, "th": nil, "td": nil,
	"h1": {"id"}, "h2": {"id"}, "h3": {"id"}, "h4": {"id"}, "h5": {"id"}, "h6": {"id"}, "li": {"id"},
	"ol": {"start", "class"}, "code": {"class"}, "div": {"class"}, "sup": {"class"}, "section": {"class"},
	"span": {"class", "data-nolink"}, "a": {"href", "id", "class", "title", "data-doc", "aria-label"},
}

// outsideAllowlist is every element, attribute, comment and link scheme in body that viewerAllows does
// not hold, as a browser's tokenizer reads them.
func outsideAllowlist(body string) []string {
	var bad []string
	z := html.NewTokenizer(strings.NewReader(body))
	for {
		switch z.Next() {
		case html.ErrorToken:
			return bad
		case html.CommentToken:
			bad = append(bad, "<!--"+string(z.Text())+"-->")
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			attrs, ok := viewerAllows[tok.Data]
			if !ok {
				bad = append(bad, "<"+tok.Data+">")
				continue
			}
			for _, a := range tok.Attr {
				if !slices.Contains(attrs, a.Key) {
					bad = append(bad, "<"+tok.Data+" "+a.Key+">")
				}
				if a.Key == "href" {
					if u, err := url.Parse(a.Val); err != nil || u.Scheme != "" && !slices.Contains([]string{"http", "https", "mailto"}, u.Scheme) {
						bad = append(bad, "<a href="+a.Val+">")
					}
				}
			}
		}
	}
}

// mainOf is the documents' bodies on a page: what the renderer wrote, each article's frame and
// blurb taken off.
func mainOf(page string) string {
	var b strings.Builder
	for _, art := range strings.Split(page, `<p class="blurb">`)[1:] {
		art = art[strings.Index(art, "</p>\n")+len("</p>\n"):]
		b.WriteString(art[:strings.Index(art, "</article>")])
	}
	return b.String()
}

// codeBlocks is the text of every <pre><code> in body, as a browser shows it.
func codeBlocks(t *testing.T, body string) []string {
	t.Helper()
	var out []string
	z := html.NewTokenizer(strings.NewReader(body))
	inPre, in := false, false
	var b strings.Builder
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				t.Fatal(z.Err())
			}
			return out
		case html.StartTagToken:
			switch tok := z.Token(); tok.Data {
			case "pre":
				inPre = true
			case "code":
				in = inPre
			}
		case html.EndTagToken:
			if tok := z.Token(); tok.Data == "code" && in {
				out = append(out, b.String())
				b.Reset()
				in = false
			} else if tok.Data == "pre" {
				inPre = false
			}
		case html.TextToken:
			if in {
				b.WriteString(z.Token().Data)
			}
		}
	}
}

// textOf is what a browser shows of body: its text, every tag gone.
func textOf(body string) string {
	var b strings.Builder
	z := html.NewTokenizer(strings.NewReader(body))
	for z.Next() != html.ErrorToken {
		if tok := z.Token(); tok.Type == html.TextToken {
			b.WriteString(tok.Data)
		}
	}
	return b.String()
}

// Each case's HTML stays text in what the renderer writes, before the allowlist sees it, and on the
// page after it; the section after the HTML still renders.
func TestReportHTMLCarriesNoSeatHTML(t *testing.T) {
	for _, c := range seatHTMLCases {
		t.Run(c.name, func(t *testing.T) {
			md := joinUnits(c.units)
			if bad := outsideAllowlist(renderOne(md, FileReport, anchors{}, map[string]int{}, nil).html); len(bad) > 0 {
				t.Errorf("the renderer wrote markup of a seat's unit: %q", bad)
			}
			body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: md}}), record.Family{}))
			if bad := outsideAllowlist(body); len(bad) > 0 {
				t.Errorf("report.html carries markup of a seat's unit: %q\n%s", bad, body)
			}
			if !strings.Contains(body, c.after) {
				t.Errorf("the section after the HTML is gone (want %q):\n%s", c.after, body)
			}
			if c.heading != "" && !strings.Contains(body, `id="`+c.heading+`"`) {
				t.Errorf("the heading after the HTML is no longer a heading (want id %q):\n%s", c.heading, body)
			}
		})
	}
}

// THE ALLOWLIST IS ON THE PAGE, not only in its own tests: a renderer that wrote a seat's script would
// still ship none.
func TestTheAllowlistHoldsWhatTheRendererLetThrough(t *testing.T) {
	rr := stdRenderer()
	rr.body = func(ParsedText, string, anchors, map[string]int, map[string]bool) rendered {
		return rendered{html: `<p>kept</p><script>alert(1)</script><img src="x" onerror="y()"><a href="javascript:z()" onclick="w()">l</a><!-- c -->`}
	}
	body := mainOf(renderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: "x\n"}}), record.Family{}, rr))
	if bad := outsideAllowlist(body); len(bad) > 0 {
		t.Errorf("report.html carries what the allowlist exists to drop: %q\n%s", bad, body)
	}
	if !strings.Contains(body, "<p>kept</p>") {
		t.Errorf("the allowlist dropped the renderer's own paragraph:\n%s", body)
	}
}

// tokens is body as a browser's tokenizer reads it: each tag with its attributes, and text decoded.
func tokens(body string) []html.Token {
	var out []html.Token
	z := html.NewTokenizer(strings.NewReader(body))
	for z.Next() != html.ErrorToken {
		out = append(out, z.Token())
	}
	return out
}

// THE ALLOWLIST IS EXACTLY WHAT THE RENDERER WRITES. On honest documents — every construct the
// renderer has an element for, the citation layer, a computation's anchor, links across the set and record
// ids — the policy changes nothing a browser reads. If it does, the policy is wrong, not the page.
func TestTheAllowlistChangesNothingTheRendererWrites(t *testing.T) {
	report := strings.Join([]string{
		"# Title", "",
		"A paragraph with **bold**, _italic_, ~~struck~~, `code`, a [link](docket.md), a [tab](evidence.md#K-deadbeef),",
		"an <https://example.invalid/a> autolink, a mail <someone@example.invalid>, https://example.invalid/b and ![an image](https://example.invalid/i.png).",
		"Line one  ", "line two.", "",
		"> a quotation", "",
		"> a quotation", ">", "> of two paragraphs", "",
		"| a | b |", "|---|:-:|", "| 1 | 2 |", "",
		"- G-00000001 opens the item", "- [x] a task done", "- [ ] a task open", "  - nested", "",
		"3. three", "4. four", "",
		"```bash", "echo '<x>' && echo \"y\"", "```", "",
		"    indented code", "",
		"---", "",
		"## Computations", "",
		"A claim[^1] and another[^K1], about G-00000001.", "",
		"[^1]: The source, which cites[^2]. https://example.invalid/a",
		"[^2]: The second source.",
		"[^K1]: a computation — in [evidence.md](evidence.md#K-deadbeef).",
	}, "\n")
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: report},
		{File: FileDocket, Nav: "Board", Body: "## The board\n\n### G-00000001 — the gap\n\nIt is open; see [the report](report.md).\n"},
		{File: FileEvidence, Nav: "Evidence", Body: "## Computations\n\n" + computationAnchorMark("K-deadbeef") + "\n\n### K1 — the check\n\n```python\nprint(91 % 7)\n```\n"},
	}
	set := stdReaders().shipSet("# T", docs)
	anchor, seen, files := anchors{}, map[string]int{}, map[string]bool{}
	for _, d := range docs {
		files[d.File] = true
	}
	var bodies []rendered
	for i, d := range docs {
		bodies = append(bodies, renderDoc(set.html[i], d.File, anchor, seen, files))
	}
	for i, d := range docs {
		body := linkIDs(bodies[i], anchor, d.File)
		got, want := tokens(bodyPolicy.Sanitize(body)), tokens(body)
		if !slices.EqualFunc(got, want, func(a, b html.Token) bool {
			return a.Type == b.Type && a.Data == b.Data && slices.Equal(a.Attr, b.Attr)
		}) {
			for k := range min(len(got), len(want)) {
				if got[k].String() != want[k].String() {
					t.Errorf("%s: the allowlist changed %q into %q", d.File, want[k].String(), got[k].String())
					break
				}
			}
			t.Errorf("%s: the allowlist changed the renderer's output (%d tokens → %d)", d.File, len(want), len(got))
		}
	}
}

// THE TOOL'S TOKENS ARE NEVER SHOWN AS TYPED, wherever they stand on a line: assembly weaves or strips
// every one before either tier reads the document. Through real assembly, from the report's text.
func TestAnchorTokensAreNeverShownAsTyped(t *testing.T) {
	runDir := newRun(t)
	blue := strings.Join([]string{
		"# On 91 — research report", "",
		"## TL;DR", "",
		"<!--gap:G-0000abcd--> The next sentence <b>x</b>.", "",
		"- <!--computation:K-deadbeef--> A listed sentence <b>y</b>.", "",
		"> <!--fx:F-0000beef--> A quoted sentence.", "",
		"[^n]: <!--gap:G-0000abcd--> A note's sentence.", "",
		"A claim[^n].", "",
		"## Open questions", "",
		"None.", "",
	}, "\n")
	id := record.Identity{Run: runtest.Open(t, runDir), SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
		t.Fatal(err)
	}
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(runDir, "report.html"))
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(filepath.Join(runDir, FileReport))
	if err != nil {
		t.Fatal(err)
	}
	for tier, shown := range map[string]string{"report.html": textOf(mainOf(string(page))), "report.md in a viewer": textOf(commonmarkUnsafe(t, string(md)))} {
		if strings.Contains(shown, "<!--") {
			t.Errorf("%s shows a token as typed:\n%s", tier, shown)
		}
		for _, want := range []string{"The next sentence <b>x</b>.", "A listed sentence <b>y</b>.", "A quoted sentence.", "A note's sentence."} {
			if !strings.Contains(shown, want) {
				t.Errorf("%s does not show %q as typed:\n%s", tier, want, shown)
			}
		}
	}
}

// A `[^` IN A FENCE OF FOUR IS CODE, past a line of three inside it. Read by a scan of fence lines,
// the inner line closed the fence: the label in the code was escaped, so it showed a backslash, and a
// label in the prose after the fence was left open, so it became a reference to a note nothing defines.
func TestALabelInsideALongerFenceStaysCode(t *testing.T) {
	md := "Output:\n\n````\nline\n```\n[^x] inside\n````\n\nProse quotes [^y] after.\n"
	got := stdReaders().escapeNotes(md)
	if want := "\n[^x] inside\n"; !strings.Contains(got, want) {
		t.Errorf("a label in the code was escaped:\n%s", got)
	}
	if want := `quotes \[^y] after`; !strings.Contains(got, want) {
		t.Errorf("a label in the prose after the fence was not escaped:\n%s", got)
	}
	body := renderOne(got, FileReport, anchors{}, map[string]int{}, nil).html
	if !slices.Contains(codeBlocks(t, body), "line\n```\n[^x] inside") {
		t.Errorf("report.html does not show the fence's content as written:\n%s", body)
	}
	if strings.Contains(body, `class="cite"`) {
		t.Errorf("report.html cites a note from a seat's typed label:\n%s", body)
	}
}

// AN AUTOLINK-SHAPED TAG IS NO ELEMENT. A browser's tokenizer takes a `:` in a tag name, so
// `<svg:x/onfocus=…>` is an element to it however markdown reads it.
func TestAnAutolinkShapedTagIsNoElement(t *testing.T) {
	md := "A <svg:x/autofocus/tabindex=1/onfocus=alert(1)> and <https://x.invalid/onfocus=alert(2)> and <javascript:alert(3)>.\n"
	if bad := outsideAllowlist(renderOne(md, FileReport, anchors{}, map[string]int{}, nil).html); len(bad) > 0 {
		t.Errorf("the renderer links an autolink-shaped tag to a scheme no link may name: %q", bad)
	}
	body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: md}}), record.Family{}))
	if bad := outsideAllowlist(body); len(bad) > 0 {
		t.Errorf("report.html makes an element or a script link of an autolink-shaped tag: %q\n%s", bad, body)
	}
	for _, want := range []string{"svg:x/autofocus/tabindex=1/onfocus=alert(1)", "javascript:alert(3)"} {
		if !strings.Contains(textOf(body), want) {
			t.Errorf("report.html does not show %q:\n%s", want, body)
		}
	}
}

// A TOKEN OPENING A FOOTNOTE'S DEFINITION LEAVES THE NOTE A NOTE. Moving the text after it to a line
// of its own indented that text by the label's width, which a long label makes an indented code block.
func TestATokenOpeningAFootnoteKeepsTheNote(t *testing.T) {
	md := "A claim[^scope].\n\n[^scope]: <!--cite:C-deadbeef--> the note says <b>x</b>.\n"
	body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileEvidence, Nav: "Evidence", Body: md}}), record.Family{}))
	if strings.Contains(body, "<pre>") {
		t.Errorf("the note's text renders as code:\n%s", body)
	}
	if bad := outsideAllowlist(body); len(bad) > 0 {
		t.Errorf("report.html carries markup of the note: %q\n%s", bad, body)
	}
	if !strings.Contains(textOf(body), "the note says <b>x</b>.") || !strings.Contains(body, `class="cite"`) {
		t.Errorf("the note is not a cited note showing its text as typed:\n%s", body)
	}
}

// AN ANGLE-BRACKET DESTINATION IS A LINK. Escaping its `<` made `[a](<foo bar>)` literal text.
func TestAngleBracketDestinationsStayLinks(t *testing.T) {
	md := "See [a](<foo bar.md>) and [r] and ![i](<https://x.invalid/a b.png>).\n\n[r]: <https://x.invalid/r r>\n"
	body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: md}}), record.Family{}))
	for _, want := range []string{`<a href="foo%20bar.md">a</a>`, `<a href="https://x.invalid/r%20r">r</a>`, `<a href="https://x.invalid/a%20b.png">i</a>`} {
		if !strings.Contains(body, want) {
			t.Errorf("report.html lacks the link %s:\n%s", want, body)
		}
	}
}

// A COMPUTATION'S LABEL ANCHORS ONCE, IN BOTH TIERS. Two computations under one label gave the second a literal
// tag, and a label that is no computation id panicked the assembly.
func TestAComputationLabelAnchorsOnce(t *testing.T) {
	md := "## Computations\n\n" + computationAnchorMark("K-deadbeef") + "\n\n### K1\n\n" + computationAnchorMark("K-deadbeef") + "\n\n### K2\n\n" + computationAnchorMark("not-a-label") + "\n\n### K3\n"
	set := stdReaders().shipSet("# T", []Doc{{File: FileEvidence, Nav: "Evidence", Body: md}})
	page := mainOf(RenderSite(set, record.Family{}))
	shipped := set.md[0]
	for tier, got := range map[string]string{"report.html": page, "evidence.md": shipped} {
		if n := strings.Count(got, computationAnchor("K-deadbeef")); n != 1 {
			t.Errorf("%s holds %d anchors for the label, want 1:\n%s", tier, n, got)
		}
		if strings.Contains(got, "not-a-label") || strings.Contains(got, "computation-anchor") {
			t.Errorf("%s shows a mark:\n%s", tier, got)
		}
	}
}

// A FENCE'S INFO STRING IS CODE. Escaping its `<` gave the code a class of `language-\<lang>`.
func TestAFenceInfoStringIsNotEscaped(t *testing.T) {
	md := "```<lang>\ncode\n```\n"
	if got := ship(md); got != md {
		t.Errorf("the .md escaped a fence's info string:\n%s", got)
	}
	body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: md}}), record.Family{}))
	z := html.NewTokenizer(strings.NewReader(body))
	for z.Next() != html.ErrorToken {
		if tok := z.Token(); tok.Data == "code" && tok.Type == html.StartTagToken {
			if len(tok.Attr) != 1 || tok.Attr[0].Val != "language-<lang>" {
				t.Errorf("the code's class is %v, want language-<lang>", tok.Attr)
			}
			return
		}
	}
	t.Errorf("report.html holds no code:\n%s", body)
}

// AN AUTOLINK'S URL IS ITS CONTENTS, LITERALLY: CommonMark reads no backslash escape and no character
// reference inside `<…>`, and its own example links `<https://example.com/\[\>` to
// `https://example.com/%5C%5B%5C`. A link's destination is decoded (linkURL), an autolink's is not.
func TestAnAutolinkTakesItsURLLiterally(t *testing.T) {
	md := "See <https://example.com/\\[\\> and <https://x.invalid/a&amp;b> and [d](https://x.invalid/\\[).\n"
	body := mainOf(RenderSite(stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: md}}), record.Family{}))
	var hrefs []string
	z := html.NewTokenizer(strings.NewReader(body))
	for z.Next() != html.ErrorToken {
		if tok := z.Token(); tok.Data == "a" && tok.Type == html.StartTagToken {
			for _, a := range tok.Attr {
				if a.Key == "href" {
					hrefs = append(hrefs, a.Val)
				}
			}
		}
	}
	want := []string{"https://example.com/%5C%5B%5C", "https://x.invalid/a&amp;b", "https://x.invalid/%5B"}
	if !slices.Equal(hrefs, want) {
		t.Errorf("the links go to %q, want %q:\n%s", hrefs, want, body)
	}
}

// EVERY FOOTNOTE REFERENCE LINKS ITS OWN DOCUMENT'S ENTRY. A note's id takes a count when the page
// already gave its name out — to an earlier tab's note of that label, or to a heading "Fn 1" — so a
// reference that spells `#fn-<label>` itself links another element. Each reference resolves through
// its document's own names: a body's cite, the scent its title carries, and the page title's
// reference, which is the report's and opens the report's tab at its entry. A label the report lists
// no entry for links nowhere.
func TestEveryNoteReferenceLinksItsOwnDocumentsEntry(t *testing.T) {
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: "## Fn 1\n\nA claim[^1].\n\n[^1]: The report's source.\n"},
		{File: FileDebate, Nav: "Debate", Body: "## Its own[^1]\n\nA position[^1].\n\n[^1]: The debate's source, after[^2].\n[^2]: A second.\n"},
	}
	page := RenderSite(stdReaders().shipSet("# T[^1] and [^9]", docs), record.Family{})
	root, err := html.Parse(strings.NewReader(page))
	if err != nil {
		t.Fatal(err)
	}
	attr := func(n *html.Node, key string) string {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val
			}
		}
		return ""
	}
	var walk func(n *html.Node, f func(*html.Node))
	walk = func(n *html.Node, f func(*html.Node)) {
		f(n)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c, f)
		}
	}
	elems := map[string]map[string]string{} // document → id → element
	var masthead *html.Node
	walk(root, func(n *html.Node) {
		if n.Type != html.ElementNode {
			return
		}
		if n.Data == "article" {
			ids := map[string]string{}
			walk(n, func(c *html.Node) {
				if id := attr(c, "id"); c.Type == html.ElementNode && id != "" {
					ids[id] = c.Data
				}
			})
			elems[attr(n, "data-doc")] = ids
			doc := attr(n, "data-doc")
			cites := 0
			walk(n, func(c *html.Node) {
				if c.Type != html.ElementNode || c.Data != "a" || !strings.HasPrefix(attr(c, "href"), "#fn-") {
					return
				}
				cites++
				if ids[strings.TrimPrefix(attr(c, "href"), "#")] != "li" {
					t.Errorf("%s links %s, which is no entry of that tab", doc, attr(c, "href"))
				}
			})
			if cites == 0 {
				t.Errorf("%s holds no reference to an entry", doc)
			}
		}
		if n.Data == "h1" && masthead == nil {
			masthead = n
		}
	})
	if !strings.Contains(page, `title="1 — The debate&#39;s source, after2."`) {
		t.Errorf("the debate's cite does not carry its own note's scent:\n%s", page)
	}
	var links []*html.Node
	walk(masthead, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			links = append(links, n)
		}
	})
	if len(links) != 1 {
		t.Fatalf("the title holds %d link(s), want 1, to the report's entry for 1", len(links))
	}
	if id := strings.TrimPrefix(attr(links[0], "href"), "#"); elems[FileReport][id] != "li" || attr(links[0], "data-doc") != FileReport {
		t.Errorf("the title's reference links #%s in %q, want the report's entry", id, attr(links[0], "data-doc"))
	}
}
