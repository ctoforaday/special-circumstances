package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// blue edit is the ONLY write path to blue/report.md for a response seat. These drive the
// real command tree so the lockdown's write path is pinned where it is enforced: an exact
// span replace that PRESERVES red's invisible finding-markers, rejects a marker-spanning
// edit, records an append-only diff-stack op, and reconciles idempotently after a crash.

func writeReport(t *testing.T, runDir, body string) {
	t.Helper()
	dir := filepath.Join(runDir, "blue")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readReport(t *testing.T, runDir string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(runDir, "blue", "report.md"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func countType(t *testing.T, runDir, typ string) int {
	t.Helper()
	m, err := record.MergedEvents(runDir)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range m.Events {
		if e.Type == typ {
			n++
		}
	}
	return n
}

const blueSeat = "blue-respond-r2"

func registerBlue(t *testing.T, runDir string) {
	t.Helper()
	if _, err := run(t, "blue", "register", "--run", runDir, "--seat-id", blueSeat); err != nil {
		t.Fatal(err)
	}
}

func TestBlueEditReplacesSpanPreservingMarker(t *testing.T) {
	runDir := t.TempDir()
	// A trailing finding-marker after "time"; a footnote after "grows".
	writeReport(t, runDir, "# Findings\n\nThe cost is high and rising over time<!--fx:f-abc123-->. Volume grows[^v] steadily.\n")
	registerBlue(t, runDir)

	out, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising over time", "--new", "climbing sharply", "--reason", "sharper phrasing")
	if err != nil {
		t.Fatalf("blue edit: %v (out %q)", err, out)
	}
	rep := readReport(t, runDir)
	if !strings.Contains(rep, "climbing sharply") || strings.Contains(rep, "rising over time") {
		t.Errorf("span not replaced: %q", rep)
	}
	if !strings.Contains(rep, "<!--fx:f-abc123-->") {
		t.Errorf("finding-marker was dropped: %q", rep)
	}
	ev := lastOfType(t, runDir, "blue_edit")
	if ev.Payload.Str("old") != "rising over time" || ev.Payload.Str("new") != "climbing sharply" {
		t.Errorf("blue_edit op payload wrong: old=%q new=%q", ev.Payload.Str("old"), ev.Payload.Str("new"))
	}
	if ev.Payload.Str("text") != "sharper phrasing" {
		t.Errorf("reason not recorded: %q", ev.Payload.Str("text"))
	}
}

func TestBlueEditRejectsMarkerSpanningEdit(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe value is context<!--fx:f-1-->: important here.\n")
	registerBlue(t, runDir)

	out, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "context: important", "--new", "context: vital", "--reason", "x")
	if err == nil {
		t.Fatalf("expected a reject; out %q", out)
	}
	if !strings.Contains(err.Error(), "f-1") || !strings.Contains(strings.ToLower(err.Error()), "around") {
		t.Errorf("reject message should name the marker and say edit around it: %v", err)
	}
	if strings.Contains(readReport(t, runDir), "vital") {
		t.Error("report was mutated on a rejected marker-spanning edit")
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a rejected edit recorded %d stack ops, want 0", n)
	}
}

func TestBlueEditRejectsAbsentOld(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe scheduler is preemptive.\n")
	registerBlue(t, runDir)

	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "the scheduler is cooperative", "--new", "x", "--reason", "y")
	if err == nil {
		t.Fatal("a mis-quote must be rejected")
	}
	if n := countType(t, runDir, "blue_edit"); n != 0 {
		t.Errorf("a mis-quote recorded %d stack ops, want 0 (no phantom event)", n)
	}
}

func TestBlueEditIdempotentRetry(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe cost is rising fast.\n")
	registerBlue(t, runDir)
	args := []string{"blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising fast", "--new", "climbing fast", "--reason", "r"}
	if _, err := run(t, args...); err != nil {
		t.Fatal(err)
	}
	// Same --key again: old is now gone, new present. Reconcile → no-op, no second op.
	out, err := run(t, args...)
	if err != nil {
		t.Fatalf("idempotent retry errored: %v (%q)", err, out)
	}
	if !strings.Contains(out, "idempotent") {
		t.Errorf("retry should report idempotent: %q", out)
	}
	if n := countType(t, runDir, "blue_edit"); n != 1 {
		t.Errorf("retry recorded %d ops, want exactly 1", n)
	}
	if strings.Count(readReport(t, runDir), "climbing fast") != 1 {
		t.Error("retry double-applied the edit")
	}
}

// Crash after the event append, before the write lands (event-first ordering): the op is
// on the stack but the report still holds `old`. A retry under the same key reconciles
// FORWARD — applies the write — and appends no second op. No wedge.
func TestBlueEditReconcilesEventWithoutWrite(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nThe cost is rising fast.\n")
	registerBlue(t, runDir)
	// Simulate the crash window: the intent event exists, the write never happened.
	p := record.NewPayload()
	p.Set("edit_key", "E1")
	p.Set("old", "rising fast")
	p.Set("new", "climbing fast")
	p.Set("text", "r")
	if _, err := record.Append(runDir, blueSeat, "blue_edit", p); err != nil {
		t.Fatal(err)
	}
	// Retry with the same key → reconcile forward.
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "rising fast", "--new", "climbing fast", "--reason", "r"); err != nil {
		t.Fatalf("reconcile retry errored: %v", err)
	}
	if !strings.Contains(readReport(t, runDir), "climbing fast") {
		t.Error("reconcile did not apply the pending write")
	}
	if n := countType(t, runDir, "blue_edit"); n != 1 {
		t.Errorf("reconcile appended a second op (%d), want 1", n)
	}
}
