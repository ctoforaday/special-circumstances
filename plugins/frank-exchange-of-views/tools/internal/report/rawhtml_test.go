package report

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	xhtml "golang.org/x/net/html"
	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// WHAT BECOMES HTML IS DECIDED ONCE, ON THE WHOLE SHIPPED DOCUMENT (#1569).
//
// Assembly composes a document from units — the report's text, each section a seat's words fill —
// and a unit escaped alone is parsed alone, while every reader of the shipped document parses it
// whole. A fence one unit leaves open flips the next unit's code and prose; a reference definition
// in one unit makes an image of a reference in another; an indented line one unit wrote as code is
// HTML inside the list the unit before it ended with. Each case below joins its units the way
// assembly does, and the shipped document — report.html, and the .md a markdown viewer renders —
// must carry none of the units' HTML and every section after it.
var seatHTMLCases = []struct {
	name  string
	units []string
	// after is text from the section that follows the HTML, which must still ship.
	after string
	// heading, when set, is the id of a heading after the HTML that must still be a heading.
	heading string
}{
	{name: "a fence one section leaves open flips the next section's code and prose",
		units: []string{
			"## A\n\nFinding one says:\n```\nleft open",
			"## B\n\nFinding two quotes code:\n```\n<script>alert(1)</script>\n```\n",
			"## C\n\nthe last section",
		},
		after: "the last section"},
	{name: "an HTML block opened by a div carries a script and an iframe",
		units: []string{
			"## A\n\npara",
			"<div>\n<script>alert(1)</script>\n<iframe src=\"https://evil.example\"></iframe>",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
	{name: "a comment line that goes on into an external image and a script",
		units: []string{
			"## A\n\npara",
			"<!-- t --><img src=\"https://evil.example/p.png\"><script>x()</script>",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
	{name: "an unclosed comment would hide every later section",
		units: []string{
			"## A\n\npara",
			"<!--\nnever closed",
			"## Next\n\nhidden ![x](https://evil.example/y.png)",
		},
		after: "hidden", heading: "next"},
	{name: "a paragraph of nothing but inline HTML",
		units: []string{
			"## A\n\n<b>bold</b><script>x()</script>",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
	{name: "a line one unit indents as code is HTML in the list the unit before it ends with",
		units: []string{
			"## A\n\n- an authored item",
			"    <script>x()</script>",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
	{name: "a reference definition in one unit makes an image of a reference in another",
		units: []string{
			"## A\n\nBlue's paragraph.\n\n[r]: https://evil.example/p.png\n[j]: javascript:alert(1)",
			"## B\n\nA finding quotes ![x][r] and [y][j] here.",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
	{name: "an anchor in the tool's shape with an attribute of the seat's",
		units: []string{
			"## A\n\n<a id=\"P-deadbeef\" onmouseover=\"x()\"></a>",
			"## After\n\nmore",
		},
		after: "more", heading: "after"},
}

// joinUnits is a document as assembly joins one: each unit escaped where it is composed
// (Section.block), the units joined by a blank line.
func joinUnits(units []string) string {
	var esc []string
	for _, u := range units {
		esc = append(esc, stdReaders().escapeNotes(u))
	}
	return "# R\n\n" + strings.Join(esc, "\n\n") + "\n"
}

// commonmarkUnsafe renders markdown the way a CommonMark viewer that renders raw HTML does: the
// grammar a shipped .md is read in (viewerReader), with raw HTML passed through.
func commonmarkUnsafe(t *testing.T, md string) string {
	t.Helper()
	var b bytes.Buffer
	viewer := goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote), goldmark.WithRendererOptions(html.WithUnsafe()))
	if err := viewer.Convert([]byte(md), &b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// viewerReader is a markdown viewer's grammar: CommonMark with its HTML grammar, GitHub's extensions and
// footnotes.
var viewerReader = goldmark.New(goldmark.WithExtensions(extension.GFM, extension.Footnote))

// parseMD is a parse of a shipped .md as a markdown viewer reads it (viewerReader).
func parseMD(src []byte) ast.Node { return viewerReader.Parser().Parse(text.NewReader(src)) }

// htmlNodes is the source of every HTML node a CommonMark reader finds in md: each HTML block and
// each inline tag or comment.
func htmlNodes(md string) []string {
	src := []byte(md)
	var out []string
	_ = ast.Walk(parseMD(src), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch x := n.(type) {
		case *ast.HTMLBlock:
			var b strings.Builder
			for i := 0; i < x.Lines().Len(); i++ {
				seg := x.Lines().At(i)
				b.Write(seg.Value(src))
			}
			if x.HasClosure() {
				b.Write(x.ClosureLine.Value(src))
			}
			out = append(out, b.String())
		case *ast.RawHTML:
			out = append(out, rawText(x, src))
		}
		return ast.WalkContinue, nil
	})
	return out
}

// THE MARKDOWN TIER TAKES THE DECISION ON THE WHOLE DOCUMENT. A CommonMark reader of a shipped .md
// finds no HTML in it the tool did not write, and the section after the seat's HTML still ships.
func TestShippedMarkdownCarriesNoSeatHTML(t *testing.T) {
	for _, c := range seatHTMLCases {
		t.Run(c.name, func(t *testing.T) {
			d := Doc{File: FileReport, Nav: "Report", Body: joinUnits(c.units)}
			md := stdReaders().shipSet("# T <script>t()</script>", []Doc{d}).md[0]
			if found := htmlNodes(md); len(found) > 0 {
				t.Errorf("a CommonMark reader finds HTML in the shipped .md: %q\n%s", found, md)
			}
			if got := commonmarkUnsafe(t, md); !strings.Contains(got, c.after) {
				t.Errorf("the section after the HTML is gone (want %q):\n%s", c.after, got)
			}
		})
	}
}

// IN EVERY CONTAINER, AND SHOWN AS TYPED. A seat's HTML inside a list, a quote, a footnote or a table
// cell, mid-paragraph or opening a line, is no HTML to a CommonMark reader of the .md, and both a
// viewer of the .md and report.html show it as the seat typed it.
func TestSeatHTMLInEveryContainerShipsAsTyped(t *testing.T) {
	for _, c := range []struct{ name, md, typed string }{
		{"a script", "Before.\n\n<script>alert(1)</script>\n\nAfter.\n", "<script>alert(1)</script>"},
		{"an image with a handler", "Text <img src=x onerror=alert(1)> more.\n", "<img src=x onerror=alert(1)>"},
		{"a comment", "Text <!-- hidden --> more.\n", "<!-- hidden -->"},
		{"an unclosed comment", "Text.\n\n<!--\n\n## Later\n\nstill here.\n", "<!--"},
		{"an HTML block at a line's start", "<div onclick=x()>\nblock text\n</div>\n", "<div onclick=x()>"},
		{"inline HTML mid-paragraph", "A <b>bold</b> word.\n", "<b>bold</b>"},
		{"in a list", "- item <b>x</b>\n- <script>y()</script>\n", "<script>y()</script>"},
		{"in a quote", "> <script>q()</script>\n> <!-- c -->\n", "<script>q()</script>"},
		{"in a footnote", "A claim[^n].\n\n[^n]: <script>f()</script> the note.\n", "<script>f()</script> the note."},
		{"in a table cell", "| a | b |\n|---|---|\n| <img src=x onerror=y()> | <!-- c --> |\n", "<img src=x onerror=y()>"},
	} {
		t.Run(c.name, func(t *testing.T) {
			set := stdReaders().shipSet("# T", []Doc{{File: FileReport, Nav: "Report", Body: c.md}})
			md := set.md[0]
			if found := htmlNodes(md); len(found) > 0 {
				t.Errorf("a CommonMark reader finds HTML in the shipped .md: %q\n%s", found, md)
			}
			viewed := commonmarkUnsafe(t, md)
			if bad := seatElements(viewed); len(bad) > 0 {
				t.Errorf("a viewer of the .md builds %q:\n%s", bad, viewed)
			}
			page := mainOf(RenderSite(set, record.Family{}))
			for tier, shown := range map[string]string{"the .md in a viewer": textOf(viewed), "report.html": textOf(page)} {
				if !strings.Contains(shown, c.typed) {
					t.Errorf("%s does not show %q as typed:\n%s", tier, c.typed, shown)
				}
			}
		})
	}
}

// seatElements is every element, handler and comment of the seats' markup in page, as a browser's
// tokenizer reads it: the cases above write only these, and a viewer's own markup holds none of them.
func seatElements(page string) []string {
	var bad []string
	z := xhtml.NewTokenizer(strings.NewReader(page))
	for {
		switch z.Next() {
		case xhtml.ErrorToken:
			return bad
		case xhtml.CommentToken:
			bad = append(bad, "<!--"+string(z.Text())+"-->")
		case xhtml.StartTagToken, xhtml.SelfClosingTagToken:
			tok := z.Token()
			if slices.Contains([]string{"script", "img", "b", "iframe"}, tok.Data) {
				bad = append(bad, "<"+tok.Data+">")
			}
			for _, a := range tok.Attr {
				if strings.HasPrefix(a.Key, "on") {
					bad = append(bad, "<"+tok.Data+" "+a.Key+">")
				}
			}
		}
	}
}

// THE .md KEEPS THE TOOL'S ANCHOR AND ESCAPES A SEAT'S COPY. evidence.md's anchor per proof, which the
// other documents link by label, ships as HTML; the same tag typed by a seat, and an anchor token a seat
// typed (assembly weaves or strips every token the tool placed), are text.
func TestTheMarkdownKeepsOnlyTheToolsAnchor(t *testing.T) {
	md := "# R\n\n" + proofAnchorMark("P-deadbeef") + "\n\n### P1 — the check\n\n" +
		"<a id=\"P-deadbeef\"></a>\n\n<!--cite:C-00000001-->\n\nA claim<!--gap:G-0000abcd--> stands.\n"
	got := stdReaders().shipSet("# T", []Doc{{File: FileEvidence, Nav: "Evidence", Body: md}}).md[0]
	if found := htmlNodes(got); !slices.Equal(found, []string{`<a id="P-deadbeef">`, "</a>"}) {
		t.Errorf("a CommonMark reader finds HTML other than the proof's anchor: %q\n%s", found, got)
	}
	for _, want := range []string{"\n" + proofAnchor("P-deadbeef") + "\n", `&lt;a id="P-deadbeef">&lt;/a>`, `&lt;!--cite:C-00000001-->`, `A claim&lt;!--gap:G-0000abcd--> stands.`} {
		if !strings.Contains(got, want) {
			t.Errorf("the shipped .md lacks %q:\n%s", want, got)
		}
	}
}

// EVERY `<` OUTSIDE CODE SHIPS AS `&lt;`, AND CODE SHIPS BYTE FOR BYTE. Nothing is predicted about
// which `<` a reader takes for HTML: a `<` that opens no tag is written as the entity too. In every
// container a seat's code can stand in — a fence and its info string, a span, an indented block, a
// block an enclosing HTML node held, a footnote, a list item, a table cell — the code is the same bytes
// and the same code to a viewer of the .md as in the text, and no `<` is left outside it.
func TestEveryProseAngleIsAnEntityAndCodeIsByteForByte(t *testing.T) {
	for _, c := range []struct{ name, md string }{
		{"a fence and its info string", "a <b and 3<4\\<x\n\n```<lang>\\<\n<script>x()</script>\n```\n\n~~~ <x>\n<!-- c\n~~~\n"},
		{"a fence inside a fence", "````\n```\n<td>42</td>\n````\n<p>\n"},
		{"spans", "Prose `<b>code</b>` and ``a `<i>` b`` and `a\\`<u>.\n"},
		{"an indented block", "<p>\n\n    <div>indented</div>\n    \\<x>\n"},
		{"inside a dissolved HTML block", "<div>\n`<b>` is a tag\n\n```\n<b>\n```\n</div>\n"},
		{"after an unclosed comment", "<!--\n`<b>`\n\n    <i>\n"},
		{"in a footnote", "A claim[^n] <x>.\n\n[^n]: <script>f()</script> and `<c>`\n\n    ```\n    <d>\n    ```\n"},
		{"in a list", "- item <b>x</b> `<c>`\n- <s>\n\n      <code block>\n"},
		{"in a table", "| a | `<b>` |\n|---|---|\n| <img src=x onerror=y()> | `<c>` <d> |\n"},
		{"in a quote", "> <q> `<c>`\n>\n>     <i>\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := ship(c.md)
			in, out := []byte(c.md), []byte(got)
			// The text's code is report.html's reading of it: a viewer of the text as typed reads the
			// lines an HTML node encloses as HTML.
			wantCode := codeText(codeIn(parsePage(in), in), in)
			haveCode := codeText(codeIn(parseMD(out), out), out)
			if !slices.Equal(haveCode, wantCode) {
				t.Errorf("a viewer of the .md reads the code as %q, want %q:\n%s", haveCode, wantCode, got)
			}
			inCode := 0
			for _, s := range haveCode {
				inCode += strings.Count(s, "<")
			}
			if n := strings.Count(got, "<"); n != inCode {
				t.Errorf("%d `<` outside code ship as typed:\n%s", n-inCode, got)
			}
			if a, b := strings.Count(c.md, "<"), strings.Count(got, "<")+strings.Count(got, ltEntity); a != b {
				t.Errorf("the text's %d `<` ship as %d `<` and entities:\n%s", a, b, got)
			}
			if found := htmlNodes(got); len(found) > 0 {
				t.Errorf("a CommonMark reader finds HTML in the shipped .md: %q\n%s", found, got)
			}
		})
	}
	// A backslash that escaped a `<` goes with it: the entity is the escape.
	if got, want := ship("an escaped \\<u> and \\\\<v>\n"), "an escaped &lt;u> and \\\\&lt;v>\n"; got != want {
		t.Errorf("shipped %q, want %q", got, want)
	}
	// A seat's autolink and angle-bracket destination are text in the .md: their `<` is outside code.
	for md, want := range map[string]string{
		"<https://example.invalid/a> and <someone@example.invalid>\n": "&lt;https://example.invalid/a> and &lt;someone@example.invalid>\n",
		"See [a](<foo bar.md>) and ![i](<x.png>).\n":                  "See [a](&lt;foo bar.md>) and ![i](&lt;x.png>).\n",
		"[r]\n\n[r]: <https://x.invalid/r r>\n":                       "[r]\n\n[r]: &lt;https://x.invalid/r r>\n",
	} {
		if got := ship(md); got != want {
			t.Errorf("shipped %q, want %q", got, want)
		}
	}
	// And the decision is taken once: the shipped document shipped again is the same document.
	once := ship(joinUnits(seatHTMLCases[1].units))
	if twice := ship(once); twice != once {
		t.Errorf("shipping a shipped document changed it:\n--- once\n%s\n--- twice\n%s", once, twice)
	}
}

// A PASTED HTML TABLE COSTS ONE PARSE, AND NOT THE CODE ELSEWHERE. Every `<` is decided in the one
// reading, however many rows there are.
func TestAPastedHTMLTableLeavesCodeAlone(t *testing.T) {
	md := "Run `make <target>` first.\n\n<table>\n" + strings.Repeat("<tr><td>x</td></tr>\n", 200) + "</table>\n\n```\n<td>kept</td>\n```\n"
	rd, all, _ := countParses("")
	got := rd.ship(md)
	if *all != 1 {
		t.Errorf("a 200-row table parsed the text %d times, want 1", *all)
	}
	for _, code := range []string{"`make <target>`", "```\n<td>kept</td>\n```"} {
		if !strings.Contains(got, code) {
			t.Errorf("the code %q was edited:\n%s", code, got)
		}
	}
	if found := htmlNodes(got); len(found) > 0 {
		t.Errorf("a CommonMark reader finds HTML in the shipped text: %q\n%s", found, got)
	}
}

// CODE AN ENCLOSING HTML NODE HID SHIPS BYTE FOR BYTE (#1580). Inside `<div>` … `</div>`, or after an
// unclosed `<!--`, a reader with the HTML grammar holds the lines as HTML; report.html, with none, reads
// the code span and fence in them. The shipped .md is that reading: the enclosing nodes' `<` and the
// autolink's are escaped, and the code is the same bytes — and the same code — to a CommonMark reader of
// the .md as to report.html.
func TestCodeInsideDissolvedHTMLShipsByteForByte(t *testing.T) {
	inner := "`<b>` is a tag\n\n```\n<b>\n```\n\nsee <https://example.invalid/a>\n"
	for _, c := range []struct {
		name, md string
		escapes  int // the `<` of each HTML node — the div's open and close tags, or the comment — and the autolink's
	}{
		{"in a div", "<div>\n" + inner + "</div>\n", 3},
		{"after an unclosed comment", "<!--\n" + inner, 2},
		{"in a list item", "- <div>\n" + prefixLines(inner, "  ") + "  </div>\n", 3},
		{"in a quote", "> <!--\n" + prefixLines(inner, "> "), 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			md := c.md
			got := ship(md)
			if found := htmlNodes(got); len(found) > 0 {
				t.Errorf("a CommonMark reader finds HTML in the shipped text: %q\n%s", found, got)
			}
			if n := strings.Count(got, ltEntity); n != c.escapes || strings.Contains(got, `\`) {
				t.Errorf("shipped %d entities, want %d, one for each `<` outside code, and no backslash:\n%s", n, c.escapes, got)
			}
			src := []byte(md)
			want := codeText(codeIn(parsePage(src), src), src)
			if have := codeText(codeIn(parseMD([]byte(got)), []byte(got)), []byte(got)); !slices.Equal(have, want) {
				t.Errorf("a reader of the .md reads the code as %q, report.html as %q:\n%s", have, want, got)
			}
		})
	}
}

// prefixLines is s with prefix before each of its lines.
func prefixLines(s, prefix string) string {
	return prefix + strings.ReplaceAll(strings.TrimSuffix(s, "\n"), "\n", "\n"+prefix) + "\n"
}

// codeText is the text of each stretch of src in code.
func codeText(code [][2]int, src []byte) []string {
	var out []string
	for _, c := range code {
		out = append(out, string(src[c[0]:c[1]]))
	}
	return out
}

// A FOOTNOTE OPENING WITH A COMMENT STAYS A NOTE. Whatever the label's width, the definition's text is
// one paragraph of the note: nothing moves the text after the comment, so nothing can indent it into
// code.
func TestAFootnoteOpeningWithACommentStaysANote(t *testing.T) {
	for _, label := range []string{"a", "abc", "abcdefghij"} {
		t.Run(label, func(t *testing.T) {
			md := ship("A claim[^" + label + "].\n\n[^" + label + "]: <!--cite:C-deadbeef--> the note says <b>x</b>.\n")
			if found := htmlNodes(md); len(found) > 0 {
				t.Errorf("a CommonMark reader finds HTML in the note: %q\n%s", found, md)
			}
			var note ast.Node
			_ = ast.Walk(parseMD([]byte(md)), func(n ast.Node, entering bool) (ast.WalkStatus, error) {
				if _, ok := n.(*east.Footnote); ok && entering {
					note = n
				}
				return ast.WalkContinue, nil
			})
			if note == nil || note.ChildCount() != 1 || note.FirstChild().Kind() != ast.KindParagraph {
				t.Errorf("the note is not one paragraph:\n%s", md)
			}
		})
	}
}

// THROUGH REAL ASSEMBLY: the report's text and a gap a lens minted both carry HTML, and neither
// reaches report.html or the shipped markdown.
func TestAssembledSetCarriesNoSeatHTML(t *testing.T) {
	runDir := newRun(t)
	blue := strings.Join([]string{
		"# Whether the cache is coherent — research report",
		"",
		"## TL;DR",
		"The cache is coherent.",
		"",
		"<div>",
		"<script>alert(1)</script>",
		"<iframe src=\"https://evil.example\"></iframe>",
		"",
		"## Analysis",
		"Single-writer invalidation holds.",
		"",
		"<!--",
		"never closed",
		"",
		"## Open questions",
		"Whether eviction races the reader.",
		"",
	}, "\n")
	seen := map[string]bool{}
	add := func(seatID string, body proto.Message) {
		t.Helper()
		id := record.Identity{Run: runtest.Open(t, runDir), SeatID: seatID}
		if !seen[seatID] {
			occasion := ""
			if record.SeatOwesOccasion(seatID) {
				occasion = "assemble"
			}
			if _, _, err := record.RegisterSeat(id, "", occasion); err != nil {
				t.Fatalf("register %s: %v", seatID, err)
			}
			seen[seatID] = true
		}
		if _, err := record.Append(id, body); err != nil {
			t.Fatalf("append %s/%T: %v", seatID, body, err)
		}
	}
	add("blue-synthesize", &recordpb.BaseIngest{Text: proto.String(blue)})
	if _, err := Assemble(runtest.Open(t, runDir)); err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	page, err := os.ReadFile(filepath.Join(runDir, "report.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := mainOf(string(page))
	if bad := outsideAllowlist(body); len(bad) > 0 {
		t.Errorf("report.html carries markup the renderer does not write, from the report's text: %q\n%s", bad, body)
	}
	if !strings.Contains(body, `id="open-questions"`) {
		t.Errorf("a section after the unclosed comment is no longer a section:\n%s", body)
	}
	md, err := os.ReadFile(filepath.Join(runDir, FileReport))
	if err != nil {
		t.Fatal(err)
	}
	if found := htmlNodes(string(md)); len(found) > 0 {
		t.Errorf("a CommonMark reader finds HTML in report.md: %q\n%s", found, md)
	}
}

// countParses is readers that count the parses made with them, report.html's and ship's, and of them
// the ones of a text that holds marker.
func countParses(marker string) (rd readers, all, marked *int) {
	all, marked = new(int), new(int)
	seen := func(src []byte) {
		*all++
		if marker != "" && bytes.Contains(src, []byte(marker)) {
			*marked++
		}
	}
	std := stdReaders()
	return readers{
		page: func(src []byte) ast.Node { seen(src); return std.page(src) },
		code: func(src []byte) ast.Node { seen(src); return std.code(src) },
	}, all, marked
}

// THE PARSES ARE BOUNDED: one, whatever the text. Each `<` is decided in the reading the shipped text
// has, so escaping one exposes no other: a backtick a tag held, freed, pairs with the next line's in
// that reading already. The parses are counted, not timed: the box is shared.
func TestShippingIsBounded(t *testing.T) {
	for name, md := range map[string]string{
		"lines that each open a block":  strings.Repeat("<script>\n", 5000),
		"lines that each hold a tag":    strings.Repeat("text <b>x</b> and <!-- c -->\n\n", 5000),
		"lines that each free the next": "<a t=\"`\">" + strings.Repeat("`<b t='`'>\n", 5000),
		"a pasted table":                "<table>\n" + strings.Repeat("<tr><td>x</td></tr>\n", 5000) + "</table>\n",
	} {
		t.Run(name, func(t *testing.T) {
			rd, all, _ := countParses("")
			got := rd.ship(md)
			if *all != 1 {
				t.Errorf("shipping 5,000 lines of HTML parsed the text %d times, want 1", *all)
			}
			if found := htmlNodes(got); len(found) > 0 {
				t.Errorf("a CommonMark reader finds %d HTML node(s) in the shipped text", len(found))
			}
		})
	}
}

// TEXT WITH NOTHING TO ESCAPE IS NOT PARSED. Most documents hold no `<` but a proof's anchor, and most
// sections no `[^`: neither costs a parse.
func TestTextWithNothingToEscapeIsNotParsed(t *testing.T) {
	rd, all, _ := countParses("")
	md := "## Proofs\n\n" + proofAnchorMark("P-deadbeef") + "\n\nn > 3, and `code`.\n"
	if got := rd.ship(md); got != resolveMarks(md) {
		t.Errorf("shipped %q, want the text with its anchor in place", got)
	}
	if got := rd.escapeNotes("a [note] and `[x]`\n"); got != "a [note] and `[x]`\n" {
		t.Errorf("escapeNotes edited text with no `[^`: %q", got)
	}
	if *all != 0 {
		t.Errorf("parsed %d time(s) a text with nothing to escape", *all)
	}
}

// THE SHIPPED READING DECIDES WHAT IS CODE. Escaping a `<` changes how a reader reads the text around
// it: a comment or a tag that held a backtick is text, and the backtick opens a span; a comment that
// held an indented line no longer does, and the line is a paragraph's. Each `<` is decided in that
// reading, and only a `<` it holds as code ships as typed.
func TestTheShippedReadingDecidesWhatIsCode(t *testing.T) {
	for _, c := range []struct{ md, want string }{
		{"<!-- ` -->\nfoo ` <b>bad</b> `\n", "&lt;!-- ` -->\nfoo ` &lt;b>bad&lt;/b> `\n"},
		{"<a title=\"`\">`<b>\n", "&lt;a title=\"`\">`&lt;b>\n"},
		{"<xy:a`b> `<b>`\n", "&lt;xy:a`b> `&lt;b>`\n"},
		{"[t](<` x>) `<b>`\n", "[t](&lt;` x>) `&lt;b>`\n"},
		{"<div>\n- x\n\n    <script>y()</script>\n", "&lt;div>\n- x\n\n    &lt;script>y()&lt;/script>\n"},
		{"<!--\nfoo\n-->\n    <b>x</b>\n", "&lt;!--\nfoo\n-->\n    &lt;b>x&lt;/b>\n"},
	} {
		if got := ship(c.md); got != c.want {
			t.Errorf("%q ships as %q, want %q", c.md, got, c.want)
		}
		if found := htmlNodes(c.want); len(found) > 0 {
			t.Errorf("a CommonMark reader finds HTML in %q: %q", c.want, found)
		}
	}
}

// EACH TIER READS EACH DOCUMENT ONCE, IN SHIPPING. shipSet parses each document's text once for report.html
// and once for the .md — a marker byte in it costs no second shipping.
// RenderSite and Write take its result, and hold no reader to parse a body with.
func TestEachDocumentIsShippedOnce(t *testing.T) {
	const body = "BODYTEXT"
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: "## A\n\n" + body + " one <b>x</b>.\n"},
		{File: FileDebate, Nav: "Debate", Body: "## B\n\n" + body + " two \x01.\n"},
	}
	rd, _, marked := countParses(body)
	rd.shipSet("# T", docs)
	if *marked > 2*len(docs) {
		t.Errorf("shipSet parsed the bodies %d times, want at most %d", *marked, 2*len(docs))
	}
}

// THE SEAT READS THE RECORD'S TEXT. `show report --assembled` and its --quote and --match selectors
// read the layout, and a seat copies a sentence from it into `edit`: an entity the shipped file
// carries would make the copy a misquote and a quote of the text match nothing.
func TestTheLayoutShowsTheRecordsTextUnescaped(t *testing.T) {
	run := runtest.Open(t, newRun(t))
	id := record.Identity{Run: run, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	blue := "# On lists\n\n## TL;DR\n\nA List<String> holds <b>names</b>.\n\n## Types of List<String>\n\nMore.\n\n## Open questions\n\nNone.\n"
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
		t.Fatal(err)
	}
	got, err := Layout(run)
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	var headings []string
	for _, s := range got.Sections {
		all.WriteString(s.Text + "\n")
		headings = append(headings, s.AuthoredHeadings...)
	}
	for _, want := range []string{"A List<String> holds <b>names</b>.", "## Types of List<String>"} {
		if !strings.Contains(all.String(), want) {
			t.Errorf("the assembled layout does not hold %q as the record has it:\n%s", want, all.String())
		}
	}
	if strings.Contains(all.String(), ltEntity) {
		t.Errorf("the assembled layout carries the shipped file's escapes:\n%s", all.String())
	}
	if !slices.Contains(headings, "Types of List<String>") {
		t.Errorf("the layout's authored headings are not the record's: %q", headings)
	}
}

// A PROOF'S ANCHOR IS THE TOOL'S ONLY WHERE THE TOOL WROTE IT. A seat that types a proof's anchor into
// the report's text would give the page two elements of that id, and the report's tab comes first: every
// footnote linking to the proof would land on the seat's copy.
func TestASeatsCopyOfAProofAnchorIsText(t *testing.T) {
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: "## A\n\n<a id=\"P-deadbeef\"></a>\n\nA claim[^P1].\n\n[^P1]: [the computation](evidence.md#P-deadbeef)\n"},
		{File: FileEvidence, Nav: "Evidence", Body: "## Proofs\n\n" + proofAnchorMark("P-deadbeef") + "\n\n### P1 — the check\n"},
	}
	set := stdReaders().shipSet("# T", docs)
	page := RenderSite(set, record.Family{})
	if n := strings.Count(page, `id="P-deadbeef"`); n != 1 {
		t.Fatalf("report.html holds %d element(s) with the proof's id, want 1", n)
	}
	if strings.Index(page, `id="P-deadbeef"`) < strings.Index(page, `<article id="`+docElement(FileEvidence)) {
		t.Errorf("the proof's id is not evidence.md's element")
	}
	if md := set.md[0]; !strings.Contains(md, `&lt;a id="P-deadbeef">&lt;/a>`) {
		t.Errorf("report.md ships the seat's copy of the anchor as HTML:\n%s", md)
	}
	if md := set.md[1]; !strings.Contains(md, "\n"+proofAnchor("P-deadbeef")+"\n") {
		t.Errorf("evidence.md does not ship the proof's anchor:\n%s", md)
	}
}

// A PROOF SHOWS THE EXACT BYTES THAT RAN. A fence of three backticks closed at an output's own "```"
// line, and the rest shipped as prose — escaped, so the evidence no longer showed what ran.
func TestAProofsCodeShipsByteForByte(t *testing.T) {
	runDir := newRun(t)
	const sha = "abc123def4567890"
	script := "print('```')\nprint('<td>42</td>')"
	output := "```\n<td>42</td>\n````"
	seedProof(t, runDir, sha, script, output+"\n")
	recorded := "`<x>` at the edge"
	proofs := []record.Proof{{Label: "P-deadbeef", SHA: sha, Basis: "reproducible", Script: "s.py",
		Rerun: &record.ProofRerun{Rerun: recordpb.Rerun_RERUN_DIFFERS, Recorded: recorded, Observed: "a ` b"}}}
	ev := evidenceDoc(runtest.Open(t, runDir), proofs, map[string]bool{"P-deadbeef": true})
	set := stdReaders().shipSet("# T", []Doc{{File: FileEvidence, Nav: "Evidence", Body: ev}})
	blocks := codeBlocks(t, mainOf(RenderSite(set, record.Family{})))
	for _, want := range []string{script, output} {
		if !slices.Contains(blocks, want) {
			t.Errorf("report.html does not show %q as one code block: %q", want, blocks)
		}
	}
	md := set.md[0]
	viewed := commonmarkUnsafe(t, md)
	for _, want := range []string{script, output} {
		if !strings.Contains(md, want) {
			t.Errorf("evidence.md does not hold the bytes as they ran, %q:\n%s", want, md)
		}
		if !strings.Contains(viewed, fmt.Sprintf(">%s\n</code>", escapeForTest(want))) {
			t.Errorf("a viewer does not show %q as one code block:\n%s", want, viewed)
		}
	}
	for _, want := range []string{"<code>" + escapeForTest(recorded) + "</code>", "<code>a ` b</code>"} {
		if !strings.Contains(viewed, want) {
			t.Errorf("a re-run's output is not one code span, want %q:\n%s", want, viewed)
		}
	}
}

// escapeForTest is s as goldmark's HTML renderer writes text.
func escapeForTest(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}

// rawText is an inline HTML node's source.
func rawText(x *ast.RawHTML, src []byte) string {
	var b bytes.Buffer
	for i := 0; i < x.Segments.Len(); i++ {
		seg := x.Segments.At(i)
		b.Write(seg.Value(src))
	}
	return b.String()
}

// A SECTION THE STRIP LEAVES AS IT WAS IS READ ONCE FOR BOTH. The strip reads where a tool-composed
// section's code is; where it took no token out the text is the same, and the note escape reads its
// code from that reading rather than parsing again. Where the strip changed the text, what is code may
// have changed with it, and the escape reads the new text.
func TestASectionTheStripLeavesIsReadOnce(t *testing.T) {
	for _, c := range []struct {
		says string
		want int
	}{
		{"a `<!--gap:G-00000001-->` and [^x]\n", 1},
		{"a <!--gap:G-00000001--> and [^x]\n", 2},
	} {
		rd, all, _ := countParses("")
		says := func(*assembly, Section) string { return c.says }
		got := Section{Heading: "The board", Composer: Tool, compose: says}.block(&assembly{rd: rd})
		if !strings.Contains(got, `\[^x]`) {
			t.Errorf("%q: the note opener is not escaped:\n%s", c.says, got)
		}
		if pages := *all / 2; pages != c.want {
			t.Errorf("%q: read %d time(s) (%d parses), want %d", c.says, pages, *all, c.want)
		}
	}
}
