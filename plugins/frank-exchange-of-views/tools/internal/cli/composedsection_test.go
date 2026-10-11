package cli

import (
	"os"
	"path/filepath"
	"regexp"
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
// universe-m20: blue's report carried a "## Alternatives considered" section. Assembly composes
// that section from the avenues, so blue's never shipped — and nothing said so while a proof, two
// gaps and thirteen findings were anchored in it and lenses audited it for five sittings. The
// author reads the refusal at ingest and at edit, the two acts that put text in the report; each
// names the heading, what assembly composes the section from, and the way in.

const composedSection = "## Alternatives considered\n\n- A probable-prime test was weighed and rejected.\n"

func refusesTheSection(t *testing.T, out string, err error, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("report text carrying a section the tool composes was recorded:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, w := range append(want, "NOTHING IS RECORDED", "HOW CONTENT REACHES IT") {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
}

var avenueWay = []string{`"## Alternatives considered"`, "report.md", "the avenues on the record that are declined or abandoned",
	"`avenue propose`", "`avenue move`", "`declined` or `abandoned`"}

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
	refusesTheSection(t, out, err, avenueWay...)
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
	refusesTheSection(t, out, err, avenueWay...)
	// And a heading the edit COMPLETES: no part of it is in --new, so only the planned document shows it.
	out, err = run(t, "edit", "--run", runDir, "--seat-id", blueSeat, "--key", "E2",
		"--quote", "Other approaches", "--new", "Alternatives considered", "--reason", "name the section for what it lists")
	refusesTheSection(t, out, err, avenueWay...)
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
	for _, f := range record.AvenueStatusNames() {
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

// M20 IS THE CASE: THIS BINARY REFUSES ITS INGEST, AND MIGRATES AND ASSEMBLES ITS RECORD WITH
// NOTHING LOST.
//
// universe-m20 is the first archive whose report carries a heading the tool composes. The refusal
// is the verbs'; migrate replays the record's events, so the archive comes forward with nothing
// refused and the heading still in its report. Assembly keeps no list of headings to leave out:
// the text blue filed there ships under "Additional findings" with every anchor it carries, and
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
	refusesTheSection(t, out, err, avenueWay...)

	// MIGRATED: the record holds the heading, and everything under it ships.
	blue, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		t.Fatal(err)
	}
	if line, s, found := report.ComposedIn(blue); !found || line != "## Alternatives considered" || s.Heading != "Alternatives considered" {
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
	docs, err := report.AssembleAll(dst)
	if err != nil {
		t.Fatal(err)
	}
	var shipped string
	for _, d := range docs {
		if d.File == report.FileReport {
			shipped = d.Body
		}
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
