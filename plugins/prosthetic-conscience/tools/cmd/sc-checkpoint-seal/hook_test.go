package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func call(t *testing.T, stdin, projectDir string, now time.Time, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &e, projectDir, now)
	return o.String(), e.String(), code
}

func input(t *testing.T, in hookInput) string {
	t.Helper()
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeNote(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func snapshots(t *testing.T, projectDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(projectDir, ".claude", "checkpoints"))
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.Name() != "CHECKPOINT.md" {
			out = append(out, e.Name())
		}
	}
	return out
}

const sampleNote = "---\nschema: 2\nstatus: in-progress\n---\n" +
	"## Validation loop\n1. go test ./...  · re-armed by: any change under tools/\n" +
	"## Next intended steps\n1. wire the hook (issue #999)\n"

var noon = time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)

// The contract's hard floor: compaction is never blocked, on any path.
func TestNeverBlocksCompaction(t *testing.T) {
	cases := []struct {
		name       string
		projectDir func(t *testing.T) string
		stdin      string
	}{
		{"no project dir", func(*testing.T) string { return "" }, input(t, hookInput{Trigger: "auto"})},
		{"project dir with no note", func(t *testing.T) string { return t.TempDir() }, input(t, hookInput{Trigger: "auto"})},
		{"malformed stdin", func(t *testing.T) string { return t.TempDir() }, "{not json"},
		{"empty stdin", func(t *testing.T) string { return t.TempDir() }, ""},
		{"unwritable project dir", func(t *testing.T) string { return filepath.Join(t.TempDir(), "no", "such", "path") },
			input(t, hookInput{Trigger: "auto"})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, code := call(t, tc.stdin, tc.projectDir(t), noon)
			if code != 0 {
				t.Fatalf("exit %d — this hook must never block compaction", code)
			}
		})
	}
}

// With no note there is nothing the session established, so the hook says nothing.
func TestSilentWhenNoCheckpointExists(t *testing.T) {
	dir := t.TempDir()
	stdout, _, code := call(t, input(t, hookInput{Trigger: "auto"}), dir, noon)
	if code != 0 || strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected silence, got code=%d stdout=%q", code, stdout)
	}
	if s := snapshots(t, dir); len(s) != 0 {
		t.Fatalf("sealed something with no note: %v", s)
	}
}

// The seal: snapshot written, stamped, and the instruction emitted.
func TestSealsTheNoteAndSteersTheSummary(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	stdout, stderr, code := call(t, input(t, hookInput{
		Trigger: "auto", SessionID: "sess-1", AgentID: "seat-a",
	}), dir, noon)

	if code != 0 {
		t.Fatalf("exit %d, stderr=%q", code, stderr)
	}
	snaps := snapshots(t, dir)
	if len(snaps) != 1 || snaps[0] != "20260728T120000Z-auto-seat-a.md" {
		t.Fatalf("snapshot names = %v", snaps)
	}
	sealed, err := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", snaps[0]))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"trigger=auto", "session=sess-1", "agent=seat-a", "## Validation loop"} {
		if !strings.Contains(string(sealed), want) {
			t.Errorf("sealed snapshot missing %q", want)
		}
	}
	for _, want := range []string{"the validation loop", "ordered next actions"} {
		if !strings.Contains(stdout, want) {
			t.Errorf("instruction missing %q; got %q", want, stdout)
		}
	}
	// Reinforce, never introduce: the note's own text must not be pasted.
	if strings.Contains(stdout, "go test ./...") || strings.Contains(stdout, "#999") {
		t.Errorf("instruction echoed note body: %q", stdout)
	}
}

// A run/project workspace owns the note when one is active.
func TestPrefersTheRunWorkspaceNote(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		"## Open threads\n- fallback\n")
	writeNote(t, filepath.Join(dir, "projects", "feov", "CHECKPOINT.md"), sampleNote)

	stdout, _, _ := call(t, input(t, hookInput{Trigger: "manual"}), dir, noon)
	if !strings.Contains(stdout, "the validation loop") {
		t.Fatalf("did not read the run-workspace note: %q", stdout)
	}
	if strings.Contains(stdout, "the open threads") {
		t.Errorf("read the fallback note instead of the run workspace: %q", stdout)
	}
}

// Auto-compaction thrashes: three compactions in three turns were measured.
// Repeated firing must stay bounded and must not corrupt what is already sealed.
func TestRepeatedFiringStaysBounded(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	for i := 0; i < keepSnapshots+5; i++ {
		ts := noon.Add(time.Duration(i) * time.Second)
		if _, _, code := call(t, input(t, hookInput{Trigger: "auto"}), dir, ts); code != 0 {
			t.Fatalf("iteration %d exited %d", i, code)
		}
	}
	snaps := snapshots(t, dir)
	if len(snaps) != keepSnapshots {
		t.Fatalf("unbounded snapshot dir: %d entries, want %d", len(snaps), keepSnapshots)
	}
	// The survivors must be the newest, not an arbitrary subset.
	if snaps[0] != "20260728T120005Z-auto.md" {
		t.Errorf("oldest survivor = %q, want the 6th second", snaps[0])
	}
	// The live note is never consumed by sealing it.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")); err != nil {
		t.Errorf("live note disappeared: %v", err)
	}
}

// A manual /compact carries the human's own instructions; this hook's stdout
// becomes the custom instructions, so clobbering them would lose the human's ask.
func TestManualInstructionsSurvive(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)
	stdout, _, _ := call(t, input(t, hookInput{
		Trigger: "manual", CustomInstructions: "keep the migration decision",
	}), dir, noon)
	if !strings.Contains(stdout, "keep the migration decision") {
		t.Fatalf("human's /compact instructions were dropped: %q", stdout)
	}
	if !strings.Contains(stdout, "the validation loop") {
		t.Fatalf("seal instruction lost: %q", stdout)
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, _, code := call(t, "", t.TempDir(), noon, "-version")
	if code != 0 || !strings.Contains(stdout, "sc-checkpoint-seal") {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}
}
