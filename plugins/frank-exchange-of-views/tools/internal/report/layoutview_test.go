package report

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE LAYOUT SAYS A SECTION SHIPS WHEN ASSEMBLY COMPOSED IT, AND ITS SECTIONS ARE THE DOCUMENT.
//
// A line of the report's own text can open as a section's does: a paragraph that begins
// "**Question:**", or, in a record from before the refusal, a "## Bibliography" typed by hand in a
// run that cites nothing. Read off the document alone, the first made the layout report the
// Question line as shipping and every section after it as not in the set; the second gave the
// tool's Bibliography row the report's own text. The sections that ship, joined in order, are
// report.md's body with nothing between them and nothing left over.
func TestTheLayoutShipsWhatAssemblyComposed(t *testing.T) {
	run := runtest.Open(t, newRun(t))
	id := record.Identity{Run: run, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	blue := "# On 91\n\nA line before any heading.\n\n## TL;DR\n\n**Question:** is 91 prime?\n\x00\n**Outcome:** it is not.\n\n## Appendix\n\n91 / 7 = 13.\n\x00\n## Risk matrix\n\n- a list typed by hand.\n\n## Bibliography\n\n- another.\n\n## Open questions\n\nWhether 13 is prime.\n"
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String(blue)}); err != nil {
		t.Fatal(err)
	}
	got, err := Layout(run)
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]LayoutSection{}
	var texts []string
	for _, s := range got.Sections {
		by[s.Heading] = s
		if s.File == FileReport && s.Ships && s.Heading != "Title" {
			texts = append(texts, s.Text)
		}
	}
	if s := by["Question"]; s.Ships || s.Text != "" {
		t.Errorf("a paragraph of the report's text that opens \"**Question:**\" is reported as the Question line: %+v", s)
	}
	if s := by["Outcome"]; !strings.HasPrefix(s.Text, "**Outcome:** PENDING") {
		t.Errorf("the Outcome line is not the one assembly composed: %q", s.Text)
	}
	// THE SECTIONS THAT OPEN THE REPORT ARE THE BODY'S: the table holds their place in the text,
	// assembly places none of them, and the layout has no entry of their own to read as unfilled.
	for _, h := range []string{"TL;DR", "The Catechism"} {
		if s, ok := by[h]; ok {
			t.Errorf("the layout has an entry for %q, a section assembly does not place: %+v", h, s)
		}
	}
	if s := by[bibliographyHeading]; s.Ships || s.Text != "" {
		t.Errorf("a run that cites nothing ships no bibliography of the tool's, and the layout gives that row the report's own section: %+v", s)
	}
	// THE BODY IS WHAT ASSEMBLY PLACED AS IT, not what a heading search finds: text before the
	// report's first heading opens it, and a section in it under the heading of the first row the
	// tool composes after it — a record from before the refusal holds one — is the body's, with
	// the tool's own row found past it. It is the text as written, up to the section that closes it.
	s := by[bodyHeading]
	if !s.Ships || !slices.Equal(s.AuthoredHeadings, []string{"TL;DR", "Appendix", "Risk matrix", "Bibliography"}) ||
		!strings.HasPrefix(s.Text, "A line before any heading.\n\n## TL;DR\n\n**Question:** is 91 prime?") || !strings.HasSuffix(s.Text, "- another.") || !strings.Contains(s.Text, "**Outcome:** it is not.") {
		t.Errorf("the body does not hold the report's text as written, from the text before the first heading to the section that closes it: %+v", s)
	}
	if md := got.Markdown(); !strings.Contains(md, "\n    as the record stands: TL;DR; Appendix; Risk matrix; Bibliography\n") {
		t.Errorf("the layout a seat reads does not list the report's own sections on the body's entry:\n%s", md)
	}
	if rm := by["Risk matrix"]; !rm.Ships || !strings.HasPrefix(rm.Text, "## Risk matrix\n\n| Risk |") || strings.Contains(rm.Text, "typed by hand") || strings.Contains(rm.Text, "another") {
		t.Errorf("the tool's Risk matrix row is given the report's own section of that name: %+v", rm)
	}
	// THE SECTION THAT CLOSES THE TEXT SHIPS AFTER THE TOOL'S, and is the one part assembly places
	// apart from the body.
	if s := by["Open questions"]; !s.Ships || s.Text != "## Open questions\n\nWhether 13 is prime." {
		t.Errorf("the section that closes the report's text is not the Open questions row's: %+v", s)
	}
	docs, err := AssembleAll(run)
	if err != nil {
		t.Fatal(err)
	}
	// The report's text holds two lines that are a NUL and nothing more: what a mark the layout
	// slices by must not be taken for, and all of it that the set that ships holds.
	if n := strings.Count(docs[0].Body, "\x00"); n != 2 {
		t.Errorf("the set that ships holds %d NUL(s) and the report's text 2: a mark the layout slices the body by is in it", n)
	}
	if joined, body := strings.Join(texts, "\n\n"), strings.TrimSpace(docs[0].Body); joined != body {
		t.Errorf("the layout's sections, joined, are not report.md's body.\n--- sections:\n%s\n--- report.md:\n%s", joined, body)
	}
}

// THE REPORT TEMPLATE'S MAP IS THE TABLE. references/report_template.md lists every section of the
// set for a seat to read, file by file; it is hand-authored prose with a paragraph on each
// section, which is why it is held to the table here rather than generated from it. Four of its
// entries are no row, and each for a stated reason: inquiries.md is `show inquiries` printed whole
// under headings that carry counts, followed by the chair's reviews (inquiryReviews), run.md's Cost
// is folded in by capture after assembly, and README.md is the index.
func TestTheReportTemplateListsTheLayoutsSections(t *testing.T) {
	path, err := repotree.Plugin("skills", "research-protocol", "references", "report_template.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	noRow := map[string]bool{FileInquiries + " (by fate)": true, FileInquiries + " Inquiry reviews": true, FileRun + " Cost": true}
	var listed []string
	file := ""
	for _, ln := range strings.Split(string(b), "\n") {
		entry, _, _ := strings.Cut(ln, "<!--")
		entry = strings.TrimSpace(entry)
		switch {
		case strings.HasPrefix(entry, "# ") && strings.Contains(entry, ".md — "):
			file, _, _ = strings.Cut(strings.TrimPrefix(entry, "# "), " — ")
		case strings.HasPrefix(entry, "## ") && file != "" && !noRow[file+" "+strings.TrimPrefix(entry, "## ")]:
			listed = append(listed, file+" "+strings.TrimPrefix(entry, "## "))
		}
	}
	var rows []string
	for _, s := range layout {
		switch {
		case s.embeds: // no heading ships for the body, and the template writes its entry as a placeholder after the sections that open it
			for _, h := range s.first {
				rows = append(rows, s.File+" "+h)
			}
			rows = append(rows, s.File+" <"+s.Heading+">")
		case s.Line == "":
			rows = append(rows, s.File+" "+s.Heading)
		}
	}
	if !slices.Equal(listed, rows) {
		t.Errorf("the report template's sections and the layout's differ.\n template: %s\n layout:   %s", strings.Join(listed, " | "), strings.Join(rows, " | "))
	}
	for _, s := range layout {
		if s.Line != "" && s.Heading != "Title" && !strings.Contains(string(b), "\n"+strings.TrimSpace(s.Line)) {
			t.Errorf("the report template does not show the %s line (%s)", s.Heading, s.Line)
		}
	}
}

// A FENCE OF FOUR HOLDS A LINE OF THREE, for every reader of the report's lines: read as a close,
// the inner line made the code after it prose, and a "# " line in the code the report's title.
func TestALongerFenceHoldsAShorterOne(t *testing.T) {
	got := part("````\n```\n# Not a title\n## Not a section\n````\n# The title\n\nBody.\n")
	if got.title != "# The title" {
		t.Errorf("the title is %q, want the line after the fence", got.title)
	}
	var fenced []bool
	eachLine("~~~~\n~~~\nin\n~~~~ x\n~~~~\nout", func(l mdLine) { fenced = append(fenced, l.code) })
	if want := []bool{true, true, true, true, true, false}; !slices.Equal(fenced, want) {
		t.Errorf("fenced lines %v, want %v", fenced, want)
	}
}

// WHAT IS CODE, AND WHAT IS A HEADING, IS THE PARSER'S ANSWER for every reader of the report's lines
// (#1580). A line of backticks indented into a code block opens no fence, and neither does a line of
// prose that opens with a code span — a backtick fence's info string holds no backtick — so the
// sections after either are sections; a "## " line indented into code, or carried on as a paragraph's
// continuation, opens none, and a "# " line indented into code is no title.
func TestCodeAndHeadingsAreTheParsersReading(t *testing.T) {
	for _, c := range []struct{ name, text, title, body, close string }{
		{"a fence line indented into code", "    ```\n    x\n\n# Title\n\n## Open questions\n\nq\n", "# Title", "```\n    x", "## Open questions\n\nq"},
		{"prose opening with a code span", "```x``` is code.\n\n# Title\n\n## Open questions\n\nq\n", "# Title", "```x``` is code.", "## Open questions\n\nq"},
		{"a heading indented into code", "# Title\n\n    ## Open questions\n\nbody\n", "# Title", "## Open questions\n\nbody", ""},
		{"a title line indented into code", "Intro.\n\n    # Not a title\n\n# Title\n\nbody\n", "# Title", "    # Not a title\n\n\nbody", ""},
		{"a heading carried on as a paragraph's continuation", "# Title\n\npara\n    ## Open questions\n", "# Title", "para\n    ## Open questions", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := part(c.text)
			if got.title != c.title || !strings.HasSuffix(got.body, c.body) || got.close != c.close {
				t.Errorf("parted %q as title %q, body %q, close %q; want %q, a body ending %q, close %q", c.text, got.title, got.body, got.close, c.title, c.body, c.close)
			}
		})
	}
}

// EVERY ROW OF THE JSON CARRIES ITS TEXT AS A STRING, AND A ROW WITH NONE SAYS WHY.
//
// A row that did not ship used to omit `text`, so a seat's `jq '.sections[] | select(.text |
// test(...))'` stopped on a null four times in one run (universe-m34), and a null reads as data the
// view failed to read. The encoded form is the contract, so it is what this reads.
func TestEveryLayoutRowCarriesTextOrSaysWhyItHasNone(t *testing.T) {
	run := runtest.Open(t, newRun(t))
	id := record.Identity{Run: run, SeatID: "blue-synthesize"}
	if _, _, err := record.RegisterSeat(id, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := record.Append(id, &recordpb.BaseIngest{Text: proto.String("# On 91\n\n## TL;DR\n\nIt is not prime.\n")}); err != nil {
		t.Fatal(err)
	}
	got, err := Layout(run)
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var rows struct {
		Sections []map[string]any `json:"sections"`
	}
	if err := json.Unmarshal(b, &rows); err != nil {
		t.Fatal(err)
	}
	shipping, absent := 0, 0
	for _, r := range rows.Sections {
		text, isString := r["text"].(string)
		if !isString {
			t.Errorf("section %q encodes `text` as %#v, not a string", r["heading"], r["text"])
		}
		reason, hasReason := r["no_text"]
		switch ships, _ := r["ships"].(bool); {
		case ships:
			shipping++
			if hasReason || text == "" {
				t.Errorf("section %q ships with text %q and no_text %#v", r["heading"], text, reason)
			}
		default:
			absent++
			if reason != NotInTheSet || text != "" {
				t.Errorf("section %q does not ship and says %#v of its text %q, want %q", r["heading"], reason, text, NotInTheSet)
			}
		}
	}
	if shipping == 0 || absent == 0 {
		t.Fatalf("the run measured %d shipping and %d absent rows; both kinds are needed for this to test anything", shipping, absent)
	}
}
