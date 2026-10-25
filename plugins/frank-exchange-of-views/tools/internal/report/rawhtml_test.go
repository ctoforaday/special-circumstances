package report

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
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
		esc = append(esc, EscapeNotes(u))
	}
	return "# R\n\n" + strings.Join(esc, "\n\n") + "\n"
}

// seatMarkup is what no shipped document may carry as HTML: a script, a frame, an event handler,
// and the elements the seats' units opened. Escaped, each shows as text, which these do not match.
var seatMarkup = []string{"<script", "<iframe", "<div>", "<b>", `onmouseover="`}

// report.html loads no image and links nothing to script: an image is shown as a link to it, and a
// destination that would run script links nowhere. In the .md, an image or a link written as
// markdown is markdown, and a viewer reads it as it always has; a raw <img> tag must not render.
var (
	pageMarkup = append([]string{"<img", `href="javascript`}, seatMarkup...)
	mdMarkup   = append([]string{`<img src="https://evil.example/p.png">`}, seatMarkup...)
)

func TestReportHTMLCarriesNoSeatHTML(t *testing.T) {
	for _, c := range seatHTMLCases {
		t.Run(c.name, func(t *testing.T) {
			page := RenderSite("# T", []Doc{{File: FileReport, Nav: "Report", Body: joinUnits(c.units)}}, record.Family{})
			body := page[strings.Index(page, `<main`):]
			body = body[:strings.Index(body, "</main>")]
			for _, bad := range pageMarkup {
				if strings.Contains(body, bad) {
					t.Errorf("report.html carries %q from a seat's unit:\n%s", bad, body)
				}
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

// commonmarkUnsafe renders markdown the way a viewer that renders raw HTML does: CommonMark with
// GitHub's extensions and raw HTML passed through. It is a reader independent of this package's.
func commonmarkUnsafe(t *testing.T, md string) string {
	t.Helper()
	var b bytes.Buffer
	if err := goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(html.WithUnsafe())).Convert([]byte(md), &b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// THE MARKDOWN TIER TAKES THE SAME DECISION. A markdown viewer, an artifact pane, GitHub before
// its sanitiser: each renders raw HTML from the .md, so the .md that ships carries no seat's HTML
// either — every `<` outside code and the tool's own shapes is a character.
func TestShippedMarkdownCarriesNoSeatHTML(t *testing.T) {
	for _, c := range seatHTMLCases {
		t.Run(c.name, func(t *testing.T) {
			d := Doc{File: FileReport, Nav: "Report", Body: joinUnits(c.units)}
			md := render(d, "# T <script>t()</script>", []Doc{d})
			got := commonmarkUnsafe(t, md)
			for _, bad := range mdMarkup {
				if strings.Contains(got, bad) {
					t.Errorf("the shipped .md renders %q from a seat's unit:\n--- md\n%s\n--- rendered\n%s", bad, md, got)
				}
			}
			if !strings.Contains(got, c.after) {
				t.Errorf("the section after the HTML is gone (want %q):\n%s", c.after, got)
			}
		})
	}
}

// THE TOOL'S OWN SHAPES STILL RENDER: evidence.md's anchor per proof, which six documents link by
// label, and an anchor token, which renders as nothing. An HTML comment that is not exactly a token
// is a seat's, and shows as typed.
func TestTheToolsOwnHTMLStillRenders(t *testing.T) {
	md := "# R\n\n" + proofAnchor("P-deadbeef") + "\n\n### P1 — the check\n\n" +
		"<!--cite:C-00000001-->\n\nA claim<!--gap:G-0000abcd--> stands.\n\n<!-- a seat's comment -->\n"
	page := RenderSite("# T", []Doc{{File: FileEvidence, Nav: "Evidence", Body: md}}, record.Family{})
	for _, want := range []string{
		`<a id="P-deadbeef"></a>`,
		"<!--cite:C-00000001-->",
		"A claim<!--gap:G-0000abcd--> stands.",
		"&lt;!-- a seat&#39;s comment --&gt;",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("report.html lacks %q", want)
		}
	}
	d := Doc{File: FileEvidence, Nav: "Evidence", Body: md}
	got := render(d, "# T", []Doc{d})
	for _, want := range []string{proofAnchor("P-deadbeef"), "<!--cite:C-00000001-->", "A claim<!--gap:G-0000abcd--> stands.", `\<!-- a seat's comment -->`} {
		if !strings.Contains(got, want) {
			t.Errorf("the shipped .md lacks %q:\n%s", want, got)
		}
	}
}

// Code is left as written in the .md: inside a fence or a code span a backslash is a character,
// and an escape there would print. An autolink is markdown, not HTML, and stays a link.
func TestEscapingLeavesCodeAndAutolinksAlone(t *testing.T) {
	md := "Prose with `<b>code</b>` and <https://example.invalid/a>.\n\n```\n<script>x()</script>\n```\n"
	if got := escapeRawHTML(md); got != md {
		t.Errorf("code or an autolink was escaped:\n--- in\n%s\n--- out\n%s", md, got)
	}
	// And the escape is taken once: the shipped document escaped again is the same document.
	once := escapeRawHTML(joinUnits(seatHTMLCases[1].units))
	if twice := escapeRawHTML(once); twice != once {
		t.Errorf("escaping an escaped document changed it:\n--- once\n%s\n--- twice\n%s", once, twice)
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
	body := string(page)
	body = body[strings.Index(body, `<main`):]
	body = body[:strings.Index(body, "</main>")]
	for _, bad := range pageMarkup {
		if strings.Contains(body, bad) {
			t.Errorf("report.html carries %q from the report's text:\n%s", bad, body)
		}
	}
	if !strings.Contains(body, `id="open-questions"`) {
		t.Errorf("a section after the unclosed comment is no longer a section:\n%s", body)
	}
	md, err := os.ReadFile(filepath.Join(runDir, FileReport))
	if err != nil {
		t.Fatal(err)
	}
	got := commonmarkUnsafe(t, string(md))
	for _, bad := range mdMarkup {
		if strings.Contains(got, bad) {
			t.Errorf("report.md renders %q from the report's text:\n%s", bad, got)
		}
	}
}
