package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// A SECTION THE TOOL COMPOSES IS REFUSED AT THE WRITE, through the real verbs, AND THE REFUSAL
// SAYS HOW CONTENT REACHES IT.
//
// Assembly composes the "## Alternatives considered" section from the inquiries, so a section of
// that name in blue's report never ships. The author reads the refusal at ingest and at edit, the
// two acts that put text in the report; each names the heading, what assembly composes the
// section from, and the way in.

const composedSection = "## Alternatives considered\n\n- A probable-prime test was weighed and rejected.\n"

func refusesTheSection(t *testing.T, out string, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("report text carrying a section the tool composes was recorded:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, w := range append(want, "NOTHING IS RECORDED", "HOW CONTENT REACHES IT", "WHAT TO DO: take that line and what stands under it out", "`show report --assembled`") {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
}

var inquiryWay = []string{`"## Alternatives considered"`, "report.md", "the inquiries on the record that are declined or abandoned",
	"`inquiry propose`", "`inquiry record`", "`declined` or `abandoned`"}

func TestIngestRefusesASectionTheToolComposes(t *testing.T) {
	runDir := newRun(t)
	path := filepath.Join(runDir, "blue", "report.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n"+composedSection), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("register: %v", err)
	}
	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	refusesTheSection(t, out, err, inquiryWay...)
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BASE_INGEST); n != 0 {
		t.Errorf("the refused ingest recorded %d base(s)", n)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the refused ingest did not leave the file for its author to fix: %v", err)
	}

	// A section the refused seat has no way into says so, and whose act fills it.
	if err := os.WriteFile(path, []byte("# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Risk Matrix\n\n| a risk |\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	refusesTheSection(t, out, err, `"## Risk Matrix"`, "the open gaps on the board", "No act of blue's adds a row", "A lens's `mint`")

	// The file stands and the one ingest a run gets is unspent: the heading out, it freezes.
	if err := os.WriteFile(path, []byte("# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("ingest after the heading was taken out: %v\n%s", err, out)
	}
}

func TestEditRefusesASectionTheToolComposes(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Other approaches\n\n- A probable-prime test was weighed and rejected.\n")
	registerBlue(t, runDir)

	// The heading whole, in --new.
	out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E1",
		"--quote", "91 is 7 x 13.", "--new", "91 is 7 x 13.\n\n"+strings.TrimSpace(composedSection), "--reason", "list what was weighed")
	refusesTheSection(t, out, err, inquiryWay...)
	// And a heading the edit COMPLETES: no part of it is in --new, so only the planned document shows it.
	out, err = run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E2",
		"--quote", "Other approaches", "--new", "Alternatives considered", "--reason", "name the section for what it lists")
	refusesTheSection(t, out, err, inquiryWay...)
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT); n != 0 {
		t.Errorf("the refused edits recorded %d edit(s)", n)
	}

	// The same words as a deeper heading under the report's own section are the report's.
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E3",
		"--quote", "Other approaches", "--new", "Other approaches\n\n### Alternatives considered", "--reason", "group the list"); err != nil {
		t.Fatalf("a `### ` heading under the report's own section was refused: %v\n%s", err, out)
	}
}

// EVERY ACT A ROW'S PATH NAMES IS A COMMAND SOME SEAT HAS. The Path is what a refusal hands the
// seat as the way in; a verb renamed under it would send the seat to a command that is not there.
func TestEveryActTheLayoutNamesIsACommand(t *testing.T) {
	paths := map[string]bool{}
	for _, seatID := range []string{"blue-respond", "blue-synthesize", "red-lens-evidence", "red-chair", "judge"} {
		walk(NewRootFor(seatID), func(_ *cobra.Command, path []string) {
			paths[strings.Join(path, " ")] = true
		})
	}
	ticked := regexp.MustCompile("`([a-z][a-z -]*)`")
	fates := map[string]bool{}
	for _, f := range record.InquiryStatusNames() {
		fates[f] = true
	}
	acts := 0
	for _, s := range report.Sections() {
		for _, m := range ticked.FindAllStringSubmatch(s.Path, -1) {
			if fates[m[1]] {
				continue
			}
			acts++
			if !paths[m[1]] {
				t.Errorf("the row %q says content reaches it by `%s`, and no seat has that command", s.Heading, m[1])
			}
		}
	}
	if acts < 20 {
		t.Errorf("read %d act(s) off the layout's paths; the rows name more than that, so this check is not reading them", acts)
	}
}

// THE ASSEMBLED LAYOUT IS WHAT ASSEMBLY SHIPS, read before a report exists and after.
func TestShowReportAssembledIsTheSetAssemblyShips(t *testing.T) {
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	// Before any report is ingested — when the seat about to write one asks.
	out, err := run(t, "show", "report", "--assembled", "--run", runDir, "--seat-id", "blue-synthesize")
	if err != nil {
		t.Fatalf("the assembled layout before any report: %v\n%s", err, out)
	}
	order := []string{"report.md", "Title (no heading", "Outcome (no heading",
		"The report's own sections (no heading of the tool's: each section ships under the heading its author gave it) — NOT IN THE SET", "## Risk matrix", "## Research areas",
		"## Future research directions", "## Alternatives considered", "## Open questions — NOT IN THE SET", "## Bibliography — NOT IN THE SET",
		"docket.md", "## The board", "run.md", "## How this run was conducted"}
	at := 0
	for _, h := range order {
		i := strings.Index(out[at:], h)
		if i < 0 {
			t.Fatalf("the layout does not list %q after %q:\n%s", h, out[:at], out)
		}
		at += i
	}
	for _, want := range []string{"composed by: tool, from the open gaps on the board", "how content reaches it: No act of blue's adds a row",
		"composed by: blue, from the section of this name in the report's text, which closes it", "`inquiry record`",
		// A section the report does not have ships nothing, and this is where a seat reads that.
		"a report with no section of this name ships without one.",
		// The order the write holds is on the entries of the sections it places.
		"The write holds `## TL;DR` and `## The Catechism` first among them, in that order, where the text carries them.", "The write holds it last in the text",
		"composed by: blue, from the report's text up to the section that closes it, whole and in the order written", "assembly writes none above it"} {
		if !strings.Contains(out, want) {
			t.Errorf("the layout does not say %q:\n%s", want, out)
		}
	}

	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n\n## Appendix: the divisions\n\n91 / 7 = 13.\n")
	registerBlue(t, runDir)
	if out, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", blueSeat, "--question", "whether trial division to 9 settles it",
		"--hypothesis", "a factor below 10 exists"); err != nil {
		t.Fatalf("inquiry propose: %v\n%s", err, out)
	}
	out, err = run(t, "show", "report", "--assembled", "--json", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("--assembled --json: %v\n%s", err, out)
	}
	var got report.LayoutJSON
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("--assembled --json is not the layout: %v\n%s", err, out)
	}
	docs, err := shippedDocs(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	shipped := ""
	for _, d := range docs {
		if d.File == report.FileReport {
			shipped = d.Body
		}
	}
	by := map[string]report.LayoutSection{}
	for _, s := range got.Sections {
		by[s.Heading] = s
		// What the view calls a section's text is in the document assembly made, byte for byte.
		if s.File == report.FileReport && s.Ships && s.Heading != "Title" && !strings.Contains(shipped, s.Text) {
			t.Errorf("the layout's text for %q is not in the assembled report.md:\n%s", s.Heading, s.Text)
		}
	}
	if s := by["Research areas"]; !s.Ships || !strings.Contains(s.Text, "whether trial division to 9 settles it") || s.Composer != report.Tool {
		t.Errorf("Research areas does not carry the inquiry just proposed: %+v", s)
	}
	if s := by["The report's own sections"]; !s.Ships || !slices.Equal(s.AuthoredHeadings, []string{"TL;DR", "Appendix: the divisions"}) || !strings.HasPrefix(s.Text, "## TL;DR\n") || !strings.HasSuffix(s.Text, "91 / 7 = 13.") {
		t.Errorf("the body does not carry the report's own section under its author's heading: %+v", s)
	}
	for _, h := range []string{"Technical foundations", "Analysis", "Additional findings"} {
		if _, row := by[h]; row || strings.Contains(shipped, h) {
			t.Errorf("the layout holds a row %q, or report.md names it: no heading is required of a report and none is written above its sections", h)
		}
	}
	if strings.Contains(shipped, "not authored") {
		t.Errorf("report.md carries a line about a section the report does not have:\n%s", shipped)
	}
	if s := by["Bibliography"]; s.Ships || s.Text != "" || s.Path == "" {
		t.Errorf("a run that cites nothing ships no bibliography, and the row still says how content reaches it: %+v", s)
	}

	// One window is not a set of sections.
	if out, err := run(t, "show", "report", "--assembled", "--anchor", "F-00000aaa", "--run", runDir, "--seat-id", blueSeat); err == nil {
		t.Errorf("--assembled with --anchor was answered:\n%s", out)
	}
	// A selector picks sections from the structured form, and prints that form with or without
	// --json: the bare listing has no text to select on.
	jsonSel, err := run(t, "show", "report", "--assembled", "--json", "--match", "trial division", "--run", runDir, "--seat-id", blueSeat)
	if err != nil || !strings.Contains(jsonSel, `"heading": "Research areas"`) || strings.Contains(jsonSel, `"heading": "Risk matrix"`) {
		t.Errorf("--assembled --json --match does not select the section that matches: %v\n%s", err, jsonSel)
	}
	if out, err := run(t, "show", "report", "--assembled", "--match", "trial division", "--run", runDir, "--seat-id", blueSeat); err != nil || out != jsonSel {
		t.Errorf("--assembled --match in the bare form is not the --json form's selection: err %v\n%s", err, out)
	}
	// A section that ships as one line is that line.
	if s := by["Title"]; s.Text != "# On 91" {
		t.Errorf("the Title's text is %q, want the report's `# ` line alone", s.Text)
	}
}

// M20 IS THE CASE: THIS BINARY REFUSES ITS INGEST, AND MIGRATES AND ASSEMBLES ITS RECORD WITH
// NOTHING LOST.
//
// universe-m20 is the first archive whose report carries a heading the tool composes. The refusal
// is the verbs'; migrate replays the record's events, so the archive comes forward with nothing
// refused and the heading still in its report. Assembly keeps no list of headings to leave out:
// the text blue filed there ships in the body with every anchor it carries, and
// the proof footnotes run P1 to P9 with none missing (the shipped m20 report ran P1 to P7, P9).
func TestM20sIngestIsRefusedLiveAndItsArchiveShipsWhole(t *testing.T) {
	to := recordtest.TmpRun(t)
	res, err := migrate.Migrate(recordtest.RequireArchive(t, "2026-10-11_is-91-prime-m20.tar.gz"), to, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate m20: %v", err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("%d event(s) refused migrating m20; the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	dst := runtest.Open(t, to)
	base, have, _, err := record.ReportProjection(dst)
	if err != nil || !have {
		t.Fatalf("the migrated m20 holds no ingested report: %v", err)
	}

	// LIVE: the text m20's synthesizer froze, handed to this binary's ingest.
	runDir := newRun(t)
	path := filepath.Join(runDir, "blue", "report.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(base), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	refusesTheSection(t, out, err, inquiryWay...)

	// MIGRATED: the record holds the heading, and everything under it ships.
	blue, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		t.Fatal(err)
	}
	if line, s, found := report.ComposedIn(report.ParseText(""), report.ParseText(blue)); !found || line != "## Alternatives considered" || s.Heading != "Alternatives considered" {
		t.Fatalf("the migrated m20's report carries %q (found %v), want \"## Alternatives considered\" — this archive is the case", line, found)
	}
	proofs := 0
	for _, id := range anchor.IDs(blue) {
		if anchor.Kind(id) == "proof" {
			proofs++
		}
	}
	if proofs != 9 {
		t.Fatalf("the migrated m20's report carries %d proof anchors, want 9", proofs)
	}
	docs, err := shippedDocs(dst)
	if err != nil {
		t.Fatal(err)
	}
	var shipped string
	for _, d := range docs {
		if d.File == report.FileReport {
			shipped = d.Body
		}
	}
	// The two sections a pass over the whole document writes ship here, and the layout says so.
	view, err := report.Layout(dst)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range view.Sections {
		if (s.Heading == "Bibliography" || s.Heading == "Proofs") && (!s.Ships || !strings.HasPrefix(s.Text, "## "+s.Heading+"\n")) {
			t.Errorf("m20 cites and proves, and the layout's %s row is ships=%v, text %.60q", s.Heading, s.Ships, s.Text)
		}
		// The body stands before the tool's sections and holds a section under one of their names.
		// The body's row has the report's; the tool's row has the tool's.
		switch s.Heading {
		case "The report's own sections":
			if !slices.Contains(s.AuthoredHeadings, "Alternatives considered") || !strings.Contains(s.Text, "### Weighed and rejected") {
				t.Errorf("the body's row does not hold the report's own `## Alternatives considered`: %v", s.AuthoredHeadings)
			}
		case "Alternatives considered":
			if !s.Ships || !strings.HasPrefix(s.Text, "## Alternatives considered\n") || strings.Contains(s.Text, "### Weighed and rejected") {
				t.Errorf("the tool's Alternatives considered row is given the report's own section of that name: %.200q", s.Text)
			}
		}
	}
	// THE ORDER A READER MEETS IT IN: the argument before the record's account of it.
	if body, tool := strings.Index(shipped, "### Weighed and rejected"), strings.LastIndex(shipped, "## Alternatives considered\n"); body < 0 || body > tool || strings.Contains(shipped, "## Additional findings") {
		t.Errorf("m20's own sections do not ship before the sections the tool composes, or ship under a heading of the tool's")
	}
	for n := 1; n <= proofs; n++ {
		if ref := "[^P" + string(rune('0'+n)) + "]"; !strings.Contains(shipped, ref) {
			t.Errorf("report.md carries no %s: a proof anchored in the report did not ship", ref)
		}
	}
	for _, want := range []string{"which other approaches were weighed, and what became of each", "### Weighed and rejected"} {
		if n := strings.Count(shipped, want); n != 1 {
			t.Errorf("report.md carries %q %d time(s), want once: text the report holds under a heading the tool composes ships, and ships once", want, n)
		}
	}
}

// THE BARE LAYOUT CARRIES NO SECTION TEXT, AND SAYS SO WHERE A PIPE CANNOT DROP IT. A search piped
// over it finds no sentence of the report and reads as a clean one: on universe-m22 a gap's
// acceptance check was the bare form through a count, twice, and zero was the answer whatever the
// report held. The statement is on standard error, which reaches the seat beside the count; the
// form that carries the text says nothing there.
func TestTheBareAssembledLayoutSaysItCarriesNoText(t *testing.T) {
	runDir := newRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	writeReport(t, runDir, "# On 91\n\n## TL;DR\n\n91 is 7 x 13.\n")
	registerBlue(t, runDir)
	const sentence = "91 is 7 x 13."
	out, seen, err := runSeen(t, "show", "report", "--assembled", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("the bare layout: %v\n%s", err, out)
	}
	if strings.Contains(out, sentence) {
		t.Fatalf("the bare layout prints the report's text, so this test's premise is gone:\n%s", out)
	}
	if !strings.Contains(seen, report.LayoutHasNoText) {
		t.Errorf("a search of the bare layout finds nothing and nothing says why; standard error carries:\n%q", seen)
	}
	out, seen, err = runSeen(t, "show", "report", "--assembled", "--json", "--run", runDir, "--seat-id", blueSeat)
	if err != nil {
		t.Fatalf("--assembled --json: %v\n%s", err, out)
	}
	if !strings.Contains(out, sentence) {
		t.Errorf("the form the statement names does not carry the report's text:\n%s", out)
	}
	if strings.Contains(seen, report.LayoutHasNoText) {
		t.Errorf("the form that carries the text says it carries none:\n%q", seen)
	}
}

// shippedDocs is the run's report set as it ships: the run assembled, and each document it wrote — its
// name, and its .md, masthead and all, in Body — in reading order.
func shippedDocs(run record.Run) ([]report.Doc, error) {
	if _, err := report.Assemble(run); err != nil {
		return nil, err
	}
	var docs []report.Doc
	for _, f := range report.Files() {
		b, err := os.ReadFile(filepath.Join(run.Dir(), f))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		docs = append(docs, report.Doc{File: f, Body: string(b)})
	}
	return docs, nil
}
