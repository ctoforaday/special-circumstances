package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// INGEST READS ONE FILE, AND A REPORT ANYWHERE ELSE IS A REFUSAL THAT SAYS SO.
//
// B7: setup stubbed both blue/report.md and the run-root report.md. The synthesizer copied its
// 14,560-byte draft onto the run-root stub, and ingest froze the untouched blue/report.md — 40
// bytes of heading — as the report's base, then reported "the file removed" without naming it.
// With no stub, the same slip is an ingest that finds nothing, names where it looked, points at
// the stray file, and records nothing, so the author can move the file and ingest once.
func TestIngestRefusesAReportWrittenElsewhere(t *testing.T) {
	runDir := newRun(t)
	body := "# On 91\n\n91 is 7 x 13.\n"
	if err := os.WriteFile(filepath.Join(runDir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatalf("register: %v", err)
	}

	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	if err == nil {
		t.Fatalf("ingest froze a base with nothing at blue/report.md:\n%s", out)
	}
	msg := err.Error() + "\n" + out
	for _, want := range []string{filepath.Join("blue", "report.md"), "A report.md exists at the run root"} {
		if !strings.Contains(msg, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, msg)
		}
	}

	// Nothing reached the record, so the one ingest a run gets is still there to spend.
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(runDir, "report.md"), filepath.Join(runDir, "blue", "report.md")); err != nil {
		t.Fatal(err)
	}
	out, err = run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	if err != nil {
		t.Fatalf("ingest at the right path after a refusal: %v\n%s", err, out)
	}
	if !strings.Contains(out, filepath.Join("blue", "report.md")+" frozen into the record ("+"23 bytes)") {
		t.Errorf("the confirmation does not name the file it froze and its size:\n%s", out)
	}
}
