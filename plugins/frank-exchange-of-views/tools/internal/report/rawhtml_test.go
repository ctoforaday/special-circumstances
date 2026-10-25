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
			page := RenderSite("# T", Ship([]Doc{{File: FileReport, Nav: "Report", Body: joinUnits(c.units)}}), record.Family{})
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
// GitHub's extensions and raw HTML passed through. It is goldmark, the parser ship reads code with,
// so it checks the .md against THAT parser's reading of it and cannot find a place where another
// viewer reads code differently. What makes the .md safe in other viewers is the escape itself:
// every `<` outside code is escaped (TestEveryAngleOutsideCodeIsEscaped), so a viewer sees HTML only
// where it reads as prose what goldmark reads as code.
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
			md := Ship([]Doc{d}).file(0, "# T <script>t()</script>")
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
	md := "# R\n\n" + proofAnchorMark("P-deadbeef") + "\n\n### P1 — the check\n\n" +
		"<!--cite:C-00000001-->\n\nA claim<!--gap:G-0000abcd--> stands.\n\n<!-- a seat's comment -->\n"
	page := RenderSite("# T", Ship([]Doc{{File: FileEvidence, Nav: "Evidence", Body: md}}), record.Family{})
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
	got := Ship([]Doc{d}).file(0, "# T")
	for _, want := range []string{proofAnchor("P-deadbeef"), "<!--cite:C-00000001-->", "A claim<!--gap:G-0000abcd--> stands.", `\<!-- a seat's comment -->`} {
		if !strings.Contains(got, want) {
			t.Errorf("the shipped .md lacks %q:\n%s", want, got)
		}
	}
}

// Code is left as written in the .md: inside a fence or a code span a backslash is a character,
// and an escape there would print. An autolink is markdown, not HTML, and stays a link.
func TestEscapingLeavesCodeAndAutolinksAlone(t *testing.T) {
	md := "Prose with `<b>code</b>` and <https://example.invalid/a> and <someone@example.invalid>.\n\n```\n<script>x()</script>\n```\n"
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

// countParses counts the parses ship makes while the test runs, and of them the ones of a text that
// holds marker.
func countParses(t *testing.T, marker string) (all, marked *int) {
	t.Helper()
	all, marked = new(int), new(int)
	was := parseMD
	parseMD = func(src []byte) ast.Node {
		*all++
		if marker != "" && bytes.Contains(src, []byte(marker)) {
			*marked++
		}
		return was(src)
	}
	t.Cleanup(func() { parseMD = was })
	return all, marked
}

// THE ESCAPE IS LINEAR. Escaping one opener used to expose the next — each `<script>` line opens a
// block running to the end once the line above it is a paragraph — and a parse per opener made a seat's
// thousand such lines a million lines parsed, on every assembly and every `show report --assembled`.
// The parses are counted, not timed: the box is shared.
func TestEscapingIsLinear(t *testing.T) {
	all, _ := countParses(t, "")
	md := strings.Repeat("<script>\n", 5000)
	got := string(ship(md).src)
	if *all > 4 {
		t.Errorf("shipping 5,000 lines that each open a block parsed the text %d times, want at most 4", *all)
	}
	if strings.Count(got, `\<script>`) != 5000 {
		t.Errorf("not every opener was escaped:\n%.200s", got)
	}
}

// ONE DECISION PER DOCUMENT, READ BY BOTH TIERS. Ship parses each document's text at most twice on
// the common path, and RenderSite and Write take its result: neither parses a body again.
func TestEachDocumentIsShippedOnce(t *testing.T) {
	const body = "BODYTEXT"
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: "## A\n\n" + body + " one <b>x</b>.\n"},
		{File: FileDebate, Nav: "Debate", Body: "## B\n\n" + body + " two.\n"},
	}
	all, marked := countParses(t, body)
	set := Ship(docs)
	if *marked > 2*len(docs) {
		t.Errorf("Ship parsed the bodies %d times, want at most %d", *marked, 2*len(docs))
	}
	*all, *marked = 0, 0
	RenderSite("# T", set, record.Family{})
	if *all != 0 {
		t.Errorf("RenderSite parsed %d text(s); it renders the trees Ship parsed", *all)
	}
	*marked = 0
	if _, err := Write(runtest.Open(t, newRun(t)), "# T", set, "index"); err != nil {
		t.Fatal(err)
	}
	if *marked != 0 {
		t.Errorf("Write parsed a body %d time(s); it writes the text Ship escaped", *marked)
	}
}

// THE SEAT READS THE RECORD'S TEXT. `show report --assembled` and its --quote and --match selectors
// read the layout, and a seat copies a sentence from it into `edit`: a backslash the shipped file
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
	if strings.Contains(all.String(), `\<`) {
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
	set := Ship(docs)
	page := RenderSite("# T", set, record.Family{})
	if n := strings.Count(page, `id="P-deadbeef"`); n != 1 {
		t.Fatalf("report.html holds %d element(s) with the proof's id, want 1", n)
	}
	if strings.Index(page, `id="P-deadbeef"`) < strings.Index(page, `<article id="doc-`+slugFile(FileEvidence)) {
		t.Errorf("the proof's id is not evidence.md's element")
	}
	if md := set.file(0, "# T"); !strings.Contains(md, `\<a id="P-deadbeef">\</a>`) {
		t.Errorf("report.md ships the seat's copy of the anchor as HTML:\n%s", md)
	}
	if md := set.file(1, "# T"); !strings.Contains(md, "\n"+proofAnchor("P-deadbeef")+"\n") {
		t.Errorf("evidence.md does not ship the proof's anchor:\n%s", md)
	}
}

// THE TOOL'S TOKENS ARE NEVER SHOWN AS TYPED, wherever they stand on a line. A token opening a line
// opens an HTML block, which takes the rest of the line raw — where a backslash escapes nothing — so
// the text after it moves to a line of its own, in the containers it stood in.
func TestAnchorTokensAreNeverShownAsTyped(t *testing.T) {
	for _, md := range []string{
		"## A\n\n<!--cite:C-00000001--> The next sentence <b>x</b>.\n",
		"## A\n\nA paragraph.\n<!--cite:C-00000001--> The next sentence <b>x</b>.\n",
		"## A\n\n<!--cite:C-00000001--> <!--gap:G-0000abcd-->\n\nThe next sentence <b>x</b>.\n",
		"## A\n\n- <!--cite:C-00000001--> The next sentence <b>x</b>.\n",
		"## A\n\n> <!--cite:C-00000001--> The next sentence <b>x</b>.\n",
		"## A\n\n[^1]: <!--cite:C-00000001--> The next sentence <b>x</b>.\n\nA claim[^1].\n",
	} {
		t.Run(md, func(t *testing.T) {
			set := Ship([]Doc{{File: FileReport, Nav: "Report", Body: md}})
			page := RenderSite("# T", set, record.Family{})
			body := page[strings.Index(page, `<main`):]
			body = body[:strings.Index(body, "</main>")]
			shipped := set.file(0, "# T")
			viewed := commonmarkUnsafe(t, shipped)
			for tier, got := range map[string]string{"report.html": body, "the .md in a viewer": viewed} {
				if strings.Contains(got, "&lt;!--") {
					t.Errorf("%s shows a token as typed:\n%s", tier, got)
				}
				if strings.Contains(got, "<b>") {
					t.Errorf("%s renders the text after a token as HTML:\n%s", tier, got)
				}
				if !strings.Contains(got, "The next sentence") {
					t.Errorf("%s lost the text after the token:\n%s", tier, got)
				}
			}
			if strings.Contains(shipped, `\<!--`) {
				t.Errorf("the .md escapes a token:\n%s", shipped)
			}
		})
	}
}

// EVERY `<` OUTSIDE CODE IS ESCAPED, whether or not a parser reads HTML there, so the .md holds no
// HTML for any viewer to find — and where escaping turns code into prose (a comment escaped into a
// paragraph takes the next line, or a backtick, with it), the parse of the escaped text says so and
// every `<` is escaped, code or not.
func TestEveryAngleOutsideCodeIsEscaped(t *testing.T) {
	for _, c := range []struct{ md, want string }{
		{"a <b and c\n", `a \<b and c`},
		{"x <notatag here\n", `x \<notatag`},
		{"<!-- ` -->\nfoo ` <b>bad</b> `\n", `\<b>bad`},
		{"<div>\n- x\n\n    <script>y()</script>\n", `\<script>`},
		{"<!--\nfoo\n-->\n    <b>x</b>\n", `\<b>x`},
	} {
		t.Run(c.md, func(t *testing.T) {
			got := string(ship(c.md).src)
			if !strings.Contains(got, c.want) {
				t.Errorf("shipped:\n%s\nwant it to hold %q", got, c.want)
			}
			viewed := commonmarkUnsafe(t, got)
			for _, bad := range []string{"<b>", "<script", "<div>", "<notatag"} {
				if strings.Contains(viewed, bad) {
					t.Errorf("the shipped text renders %q:\n--- shipped\n%s\n--- viewed\n%s", bad, got, viewed)
				}
			}
		})
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
	md := Ship([]Doc{{File: FileEvidence, Nav: "Evidence", Body: ev}}).file(0, "# T")
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
