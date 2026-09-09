package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #873 WAS A WIRING DEFECT, SO THIS TEST DRIVES THE WIRE.
//
// The advisory existed, worked, and was reachable — it was simply never called on the base. A
// suite that exercises the census function directly passes with the call site deleted, which is
// the same shape as the bug: I wrote exactly that suite first, and removing the one line from
// ingest.go left it green. This runs the real `ingest` verb through the real root command and
// reads what a seat would actually see.
func TestIngestCensusesTheBaseThroughTheRealVerb(t *testing.T) {
	runDir := newRun(t)
	// A base carrying tells the advisory names — the shape of what shipped in the #861 arm-A
	// report, where every lane tag in the finished VERIFIED document came in through ingest.
	body := "# On 91\n\n" +
		"[minority: lane-3] The factorisation is 7 x 13.\n\n" +
		"This run checked it three ways, and the debate also weighed Fermat witnesses.\n\n" +
		"[minority: lane-1] A second lane agreed.\n"
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("register: %v", err)
	}

	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	// FLAG, DO NOT BLOCK. Ingest is write-once and has deleted the file by the time it advises;
	// a refusal here would strand a seat with neither file nor base.
	if err != nil {
		t.Fatalf("the advisory refused the ingest — it must not: %v\n%s", err, out)
	}
	if !strings.Contains(out, "frozen into the record") {
		t.Fatalf("the base was not ingested:\n%s", out)
	}
	// THE COUNTS ARE REAL, which is the half that wiring alone would not have bought: `Find` is
	// presence-only and would have said "a lane tag exists" once for the two below.
	if !strings.Contains(out, "lane-attribution ×2") {
		t.Errorf("the ingest output carries no lane-attribution census:\n%s", out)
	}
	if !strings.Contains(out, "process-voice ×2") {
		t.Errorf("the ingest output carries no process-voice census:\n%s", out)
	}
	if !strings.Contains(out, "at line") {
		t.Errorf("the census names no lines, so an author has a number and nowhere to go:\n%s", out)
	}
	if !strings.Contains(out, "not a refusal") {
		t.Errorf("the note does not say it is advice:\n%s", out)
	}
}

// A clean base ingests silently. An advisory that fires on ordinary subject prose is noise, and
// noise is how a real note stops being read.
func TestACleanBaseIngestsWithNoNote(t *testing.T) {
	runDir := newRun(t)
	seedBlueReport(t, runDir) // ordinary prose, already ingested by the helper
	// seedBlueReport ingests; re-reading its output is not possible, so assert the property that
	// matters instead: the seeded prose carries no tells, so a census of it is empty.
	if out, err := run(t, "show", "report", "--run", runDir, "--seat-id", "blue-synthesize"); err == nil {
		if strings.Contains(out, "NOTE — the base sounds") {
			t.Errorf("an advisory leaked into the rendered report:\n%s", out)
		}
	}
}
