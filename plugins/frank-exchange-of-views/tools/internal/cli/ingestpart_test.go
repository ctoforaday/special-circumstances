package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// laneRun is a run whose cast seats `lanes` lanes, as setup writes it.
func laneRun(t *testing.T, lanes int) string {
	t.Helper()
	dir := newRun(t)
	seats, laneSeats := record.CastFor(lanes)
	recordtest.Seed(t, dir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats}))
	return dir
}

// partFile is where seat's part waits for ingest: the path the tool builds from the seat.
func partFile(runDir, seat string) string {
	return filepath.Join(runDir, "blue", "candidates", seat+".md")
}

func writePart(t *testing.T, runDir, seat, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(partFile(runDir, seat)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(partFile(runDir, seat), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ingestPart registers seat, where it has not, and runs its ingest.
func ingestPart(t *testing.T, runDir, seat string) (string, error) {
	t.Helper()
	if _, err := run(t, "register", "--run", runDir, "--seat-id", seat); err != nil && !strings.Contains(err.Error(), "already") {
		t.Fatalf("register %s: %v", seat, err)
	}
	return run(t, "ingest", "--run", runDir, "--seat-id", seat)
}

func mustIngestPart(t *testing.T, runDir, seat, text string) {
	t.Helper()
	writePart(t, runDir, seat, text)
	if out, err := ingestPart(t, runDir, seat); err != nil {
		t.Fatalf("%s's ingest: %v\n%s", seat, err, out)
	}
}

// refusedWithNothingRecorded holds a refused ingest to its text, to an unchanged count of parts, and
// to its file standing.
func refusedWithNothingRecorded(t *testing.T, runDir, seat, out string, err error, parts int, want ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s's ingest was recorded; want a refusal:\n%s", seat, out)
	}
	msg := err.Error() + "\n" + out
	for _, w := range want {
		if !strings.Contains(msg, w) {
			t.Errorf("the refusal does not say %q:\n%s", w, msg)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPORT_PART); n != parts {
		t.Errorf("the record holds %d part(s) after the refusal, want %d", n, parts)
	}
	if _, serr := os.Stat(partFile(runDir, seat)); serr != nil {
		t.Errorf("the refused part's file did not stand: %v", serr)
	}
}

// A LANE'S PART IS READ FROM THE PATH THE TOOL BUILDS, RECORDED, VERIFIED AND REMOVED. A lane with no
// file there is told the path (R4) and nothing is recorded; once it writes there, the part is the
// report.
func TestALanesPartIsRecordedFromThePathItsSeatNames(t *testing.T) {
	runDir := laneRun(t, 1)
	out, err := ingestPart(t, runDir, "blue-lane-1")
	if err == nil {
		t.Fatalf("ingest recorded a part with no file:\n%s", out)
	}
	if want := "there is no part at " + partFile(runDir, "blue-lane-1") + " — ingest records that one file, the one named for your seat, so write your part there first"; !strings.Contains(err.Error(), want) {
		t.Errorf("the refusal does not name the path:\n got %v\nwant %s", err, want)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPORT_PART); n != 0 {
		t.Fatalf("a refused ingest recorded %d part(s)", n)
	}

	part := "## The factors\n\n91 is 7 x 13.\n"
	writePart(t, runDir, "blue-lane-1", part)
	out, err = ingestPart(t, runDir, "blue-lane-1")
	if err != nil {
		t.Fatalf("ingest at the built path: %v\n%s", err, out)
	}
	if !strings.Contains(out, partFile(runDir, "blue-lane-1")+" recorded as your part of the report (30 bytes)") {
		t.Errorf("the confirmation does not name the file and its size:\n%s", out)
	}
	if _, serr := os.Stat(partFile(runDir, "blue-lane-1")); !os.IsNotExist(serr) {
		t.Errorf("the recorded part's file was not removed: %v", serr)
	}
	if got := readReport(t, runDir); got != part {
		t.Errorf("the report is not the one part as written:\n got %q\nwant %q", got, part)
	}
}

// R1: A PART IS A LANE'S. The synthesizer and every other seat change the report through its acts.
func TestASeatThatIsNoLaneRecordsNoPart(t *testing.T) {
	runDir := laneRun(t, 2)
	for _, seat := range []string{"blue-synthesize", "blue-respond"} {
		writePart(t, runDir, seat, "## A section\n\nA sentence.\n")
		out, err := ingestPart(t, runDir, seat)
		refusedWithNothingRecorded(t, runDir, seat, out, err, 0,
			seat+" records no part of the report: a part is a lane's, and this run's lanes are blue-lane-1, blue-lane-2. NOTHING IS RECORDED. The report changes through edit, cite, compute and retire")
	}
}

// R2: A LANE RECORDS ONE PART.
func TestALaneRecordsOnePart(t *testing.T) {
	runDir := laneRun(t, 1)
	mustIngestPart(t, runDir, "blue-lane-1", "## First\n\nA sentence.\n")
	writePart(t, runDir, "blue-lane-1", "## Second\n\nAnother sentence.\n")
	out, err := ingestPart(t, runDir, "blue-lane-1")
	refusedWithNothingRecorded(t, runDir, "blue-lane-1", out, err, 1,
		"blue-lane-1 has already recorded its part of the report, and a lane records one. NOTHING IS RECORDED and the file stands. Your part is in the report as recorded; the synthesizer works on it from here, so end your sitting")
}

// R3: NO PART AFTER THE REPORT'S FIRST CHANGE. A part then would land after text the synthesis has
// reorganized, so the lane says in its log what its part holds instead.
func TestNoPartIsRecordedOnceTheReportHasChanged(t *testing.T) {
	runDir := laneRun(t, 2)
	mustIngestPart(t, runDir, "blue-lane-1", "## First\n\nA sentence.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	if out, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-synthesize", "--key", "E1",
		"--quote", "A sentence.", "--new", "A sentence, reworded.", "--reason", "merge"); err != nil {
		t.Fatalf("the synthesizer's edit: %v\n%s", err, out)
	}
	writePart(t, runDir, "blue-lane-2", "## Second\n\nAnother sentence.\n")
	out, err := ingestPart(t, runDir, "blue-lane-2")
	refusedWithNothingRecorded(t, runDir, "blue-lane-2", out, err, 1,
		"the report has already been changed, so a part recorded now would land after text the synthesis has reorganized. NOTHING IS RECORDED and the file stands. Your part is not in the report: say so in your log, with what it holds that the report lacks")
}

// R5: A PART HOLDS ONLY THE REPORT'S OWN SECTIONS. The title, TL;DR, Catechism and Open questions are
// the synthesizer's, so two lanes cannot both claim the title, whichever records first.
func TestAPartHoldsOnlyTheReportsOwnSections(t *testing.T) {
	for _, tc := range []struct{ name, part, line, row string }{
		{"a lane's title", "# Lane 1: Is 91 prime?\n\n## The factors\n\n91 is 7 x 13.\n", "# Lane 1: Is 91 prime?", "the title line (`# …`)"},
		{"a TL;DR", "## TL;DR\n\n91 is composite.\n", "## TL;DR", "`## TL;DR`"},
		{"a Catechism", "## The factors\n\n91 is 7 x 13.\n\n## The Catechism\n\n- Is 91 prime? No.\n", "## The Catechism", "`## The Catechism`"},
		{"open questions", "## The factors\n\n91 is 7 x 13.\n\n## Open questions\n\n- None.\n", "## Open questions", "`## Open questions`"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runDir := laneRun(t, 1)
			writePart(t, runDir, "blue-lane-1", tc.part)
			out, err := ingestPart(t, runDir, "blue-lane-1")
			refusedWithNothingRecorded(t, runDir, "blue-lane-1", out, err, 0,
				"your part would carry \""+tc.line+"\", and "+tc.row+" is the synthesizer's to write, not a lane's: a part holds only the report's own sections, under headings of its own. NOTHING IS RECORDED and the file stands. Take that line and what stands under it out of your part; what it says belongs in a section of your own or to the synthesizer")
		})
	}
}

// R6: NO READING OR CHANGING ACT BEFORE A PART. There is no report until a lane records one.
func TestTheReportIsNothingUntilALaneRecordsItsPart(t *testing.T) {
	runDir := laneRun(t, 1)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", "blue-synthesize")
	if err == nil || !strings.Contains(err.Error()+out, reportproj.NoPart) {
		t.Fatalf("a read before any part did not say there is none: %v\n%s", err, out)
	}
}

// THE PARTS JOIN IN EVENT ORDER: the second lane to record follows the first, after one blank line,
// whatever their lane numbers. Trailing blank lines of the text before it are trimmed.
func TestThePartsJoinInTheOrderTheyAreRecorded(t *testing.T) {
	runDir := laneRun(t, 3)
	mustIngestPart(t, runDir, "blue-lane-3", "## Third\n\nC.\n\n\n")
	mustIngestPart(t, runDir, "blue-lane-1", "## First\n\nA.\n")
	mustIngestPart(t, runDir, "blue-lane-2", "## Second\n\nB.")
	want := "## Third\n\nC.\n\n## First\n\nA.\n\n## Second\n\nB."
	if got := readReport(t, runDir); got != want {
		t.Errorf("the parts did not join in event order:\n got %q\nwant %q", got, want)
	}
}

// LANES, THEN PARTS, THEN THE SYNTHESIZER'S ACTS. Three lanes record their parts; the synthesizer
// reads the joined report, writes the title and TL;DR with an edit that quotes the text beside them,
// merges a restated claim with an edit, and every one of those changes is an op on the record with
// its reason — the merge the archive keeps.
func TestTheSynthesizerWorksOnTheLanesPartsThroughItsActs(t *testing.T) {
	runDir := laneRun(t, 3)
	mustIngestPart(t, runDir, "blue-lane-1", "## The factors\n\n91 is 7 x 13.\n")
	mustIngestPart(t, runDir, "blue-lane-2", "## A check by division\n\n91 divided by 7 is 13, so 91 is 7 x 13.\n")
	mustIngestPart(t, runDir, "blue-lane-3", "## Primality tests\n\nA Fermat test to base 2 rejects 91.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "show", "report", "--run", runDir, "--seat-id", "blue-synthesize")
	if err != nil || !strings.Contains(out, "91 divided by 7 is 13") || !strings.Contains(out, "A Fermat test") {
		t.Fatalf("the synthesizer does not read the joined parts: %v\n%s", err, out)
	}
	for _, e := range []struct{ key, quote, new, reason string }{
		{"E1", "## The factors", "# Is 91 prime?\n\n## TL;DR\n\nNo: 91 is 7 x 13.\n\n## The factors", "write the title and TL;DR"},
		{"E2", "91 divided by 7 is 13, so 91 is 7 x 13.", "Dividing 91 by 7 leaves no remainder.", "merge a restatement of the factors"},
	} {
		if out, err := run(t, "edit", "--run", runDir, "--seat-id", "blue-synthesize", "--key", e.key,
			"--quote", e.quote, "--new", e.new, "--reason", e.reason); err != nil {
			t.Fatalf("edit %s: %v\n%s", e.key, err, out)
		}
	}
	want := "# Is 91 prime?\n\n## TL;DR\n\nNo: 91 is 7 x 13.\n\n## The factors\n\n91 is 7 x 13.\n\n" +
		"## A check by division\n\nDividing 91 by 7 leaves no remainder.\n\n## Primality tests\n\nA Fermat test to base 2 rejects 91.\n"
	if got := readReport(t, runDir); got != want {
		t.Errorf("the synthesized report:\n got %q\nwant %q", got, want)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_REPORT_PART); n != 3 {
		t.Errorf("%d parts on the record, want 3", n)
	}
	changes, err := run(t, "show", "changes", "--run", runDir, "--seat-id", "blue-synthesize")
	if err != nil {
		t.Fatal(err)
	}
	for _, reason := range []string{"write the title and TL;DR", "merge a restatement of the factors"} {
		if !strings.Contains(changes, reason) {
			t.Errorf("the change %q is not on the record's changes:\n%s", reason, changes)
		}
	}
	// A lane that comes late finds the report changed.
	writePart(t, runDir, "blue-lane-1", "## More\n\nText.\n")
	if _, err := ingestPart(t, runDir, "blue-lane-1"); err == nil {
		t.Error("a lane recorded a second part after the synthesis")
	}
}
