package report

import (
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
	blue := "# On 91\n\n## TL;DR\n\n**Question:** is 91 prime?\n\n**Outcome:** it is not.\n\n## Appendix\n\n91 / 7 = 13.\n\n## Bibliography\n\n- a list typed by hand.\n"
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
	if s := by["Outcome"]; !strings.HasPrefix(s.Text, "**Outcome:** NONE") {
		t.Errorf("the Outcome line is not the one assembly composed: %q", s.Text)
	}
	if s := by["TL;DR"]; !s.Ships || !strings.Contains(s.Text, "**Question:** is 91 prime?") || !strings.Contains(s.Text, "**Outcome:** it is not.") {
		t.Errorf("TL;DR does not ship with the report's own lines: %+v", s)
	}
	for _, h := range []string{"The Catechism", "Analysis", "Risk matrix", "Research areas", "Open questions", "Additional findings"} {
		if !by[h].Ships {
			t.Errorf("%q is in the assembled report.md and the layout says it is not in the set", h)
		}
	}
	if s := by[bibliographyHeading]; s.Ships || s.Text != "" {
		t.Errorf("a run that cites nothing ships no bibliography of the tool's, and the layout gives that row the report's own section: %+v", s)
	}
	if s := by["Additional findings"]; !slices.Equal(s.AuthoredHeadings, []string{"Appendix", "Bibliography"}) || !strings.Contains(s.Text, "a list typed by hand") {
		t.Errorf("Additional findings does not hold the report's own sections: %+v", s)
	}
	docs, err := AssembleAll(run)
	if err != nil {
		t.Fatal(err)
	}
	if joined, body := strings.Join(texts, "\n\n"), strings.TrimSpace(docs[0].Body); joined != body {
		t.Errorf("the layout's sections, joined, are not report.md's body.\n--- sections:\n%s\n--- report.md:\n%s", joined, body)
	}
}

// THE REPORT TEMPLATE'S MAP IS THE TABLE. references/report_template.md lists every section of the
// set for a seat to read, file by file; it is hand-authored prose with a paragraph on each
// section, which is why it is held to the table here rather than generated from it. Three of its
// entries are no row, and each for a stated reason: avenues.md is `show avenues` printed whole
// under headings that carry counts, run.md's Cost is folded in by capture after assembly, and
// README.md is the index.
func TestTheReportTemplateListsTheLayoutsSections(t *testing.T) {
	path, err := repotree.Plugin("skills", "research-protocol", "references", "report_template.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	noRow := map[string]bool{FileAvenues + " (by fate)": true, FileRun + " Cost": true}
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
		if s.Line == "" {
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
