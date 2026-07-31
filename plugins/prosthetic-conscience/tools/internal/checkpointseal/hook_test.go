package checkpointseal

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
		// Not "unwritable": MkdirAll creates the parents, so this exercises a
		// project dir that does not exist yet. A genuinely unwritable directory is
		// not portably testable — CI runs as root, where a 0500 mode is no barrier.
		{"project dir that does not exist yet", func(t *testing.T) string { return filepath.Join(t.TempDir(), "no", "such", "path") },
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
	}), dir, noon, "-event", "PreCompact")

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
	for _, want := range []string{"event=PreCompact", "occasion=auto", "session=sess-1", "agent=seat-a", "## Validation loop"} {
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

	stdout, _, _ := call(t, input(t, hookInput{Trigger: "manual"}), dir, noon, "-event", "PreCompact")
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
	}), dir, noon, "-event", "PreCompact")
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

// A session that ends WITHOUT ever compacting used to leave nothing at all.
// SessionEnd.reason is "other" for a headless `claude -p` run — measured, not
// assumed — so a seal matching only the interactive reasons never fires in
// exactly the sessions with no human watching. Tested as a class, not a list.
func TestSessionEndSealsOnEveryReasonIncludingOther(t *testing.T) {
	for _, reason := range []string{"other", "clear", "logout", "prompt_input_exit", ""} {
		t.Run("reason="+reason, func(t *testing.T) {
			dir := t.TempDir()
			writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

			stdout, _, code := call(t, input(t, hookInput{
				Reason: reason, SessionID: "sess-end",
			}), dir, noon, "-event", "SessionEnd")

			if code != 0 {
				t.Fatalf("exit %d", code)
			}
			snaps := snapshots(t, dir)
			if len(snaps) != 1 {
				t.Fatalf("reason %q sealed %v, want 1 snapshot", reason, snaps)
			}
			if !strings.HasPrefix(snaps[0], "20260728T120000Z-end") {
				t.Errorf("seal not labelled as a session end: %q", snaps[0])
			}
			// Nothing left to steer, and stdout on SessionEnd addresses nobody.
			if strings.TrimSpace(stdout) != "" {
				t.Errorf("SessionEnd emitted an instruction: %q", stdout)
			}
		})
	}
}

// A seat finishing is the same seam one level down, and SubagentStop is the only
// point where the seat's agent_id and its trajectory are both in hand.
func TestSubagentStopSealsPerSeat(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	stdout, _, code := call(t, input(t, hookInput{
		AgentID: "aeaae1e2e57179ff5", AgentType: "general-purpose", SessionID: "shared",
	}), dir, noon, "-event", "SubagentStop")

	if code != 0 {
		t.Fatalf("exit %d — a seal must never block a seat", code)
	}
	snaps := snapshots(t, dir)
	if len(snaps) != 1 {
		t.Fatalf("snapshots = %v", snaps)
	}
	if !strings.Contains(snaps[0], "aeaae1e2e57179ff5") {
		t.Errorf("seat seal not keyed by agent_id: %q — every subagent shares the parent's session_id", snaps[0])
	}
	if !strings.Contains(snaps[0], "seat-general-purpose") {
		t.Errorf("seal does not say which kind of seat produced it: %q", snaps[0])
	}
	// THE constraint from sc-checkpoint-restore's measurement: stdout here reaches
	// a seat still working, so an instruction is a directive it never established.
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("SubagentStop emitted an instruction into a live seat: %q", stdout)
	}
}

// Concurrent seats finishing in the same second must not overwrite each other's
// evidence. Two seats share session_id and can share a timestamp.
func TestConcurrentSeatsInTheSameSecondBothSurvive(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	for _, id := range []string{"seat-red", "seat-blue"} {
		call(t, input(t, hookInput{AgentID: id, AgentType: "auditor"}), dir, noon, "-event", "SubagentStop")
	}
	if snaps := snapshots(t, dir); len(snaps) != 2 {
		t.Fatalf("snapshots = %v, want one per seat", snaps)
	}
}

// Same second, different events: SessionEnd can land on the same tick as a final
// SubagentStop. Neither may clobber the other.
func TestSameSecondDifferentEventsBothSurvive(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	call(t, input(t, hookInput{}), dir, noon, "-event", "SubagentStop")
	call(t, input(t, hookInput{}), dir, noon, "-event", "SessionEnd")
	call(t, input(t, hookInput{Trigger: "auto"}), dir, noon, "-event", "PreCompact")

	if snaps := snapshots(t, dir); len(snaps) != 3 {
		t.Fatalf("snapshots = %v, want 3 distinct seals", snaps)
	}
}

// Identical event, identical second — the pure collision case.
func TestIdenticalSealsInTheSameSecondDoNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	call(t, input(t, hookInput{Trigger: "auto"}), dir, noon, "-event", "PreCompact")
	call(t, input(t, hookInput{Trigger: "auto"}), dir, noon, "-event", "PreCompact")

	if snaps := snapshots(t, dir); len(snaps) != 2 {
		t.Fatalf("snapshots = %v — a seal silently overwrote another's evidence", snaps)
	}
}

// The version-skew default, pinned so it is a decision rather than an accident.
//
// The ONLY way to reach an unflagged invocation is an older hooks.json, which
// registers PreCompact alone — so the cost is one compaction without steering.
// The alternative default (assume PreCompact) would emit summarizer instructions
// from whatever event a hand-written config pointed at this binary, including
// into a live seat. Degradation beats a directive in the wrong channel.
func TestUnlabelledInvocationSealsButStaysSilent(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	stdout, _, code := call(t, input(t, hookInput{Trigger: "auto"}), dir, noon)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if len(snapshots(t, dir)) != 1 {
		t.Error("an unlabelled invocation must still seal — that half is always safe")
	}
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("emitted an instruction for an unknown event: %q", stdout)
	}
}

// The payload field is a fallback for that skew window, never a dependency: the
// spike never verified hook_event_name exists in hook input.
func TestPayloadEventNameIsUsedWhenTheFlagIsAbsent(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	stdout, _, _ := call(t, input(t, hookInput{
		Trigger: "auto", HookEventName: "PreCompact",
	}), dir, noon)
	if !strings.Contains(stdout, "the validation loop") {
		t.Errorf("payload event name ignored; got %q", stdout)
	}
}

// The flag is the authority when both are present.
func TestFlagBeatsThePayloadField(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	stdout, _, _ := call(t, input(t, hookInput{
		HookEventName: "PreCompact", Trigger: "auto",
	}), dir, noon, "-event", "SubagentStop")
	if strings.TrimSpace(stdout) != "" {
		t.Errorf("payload field overrode the flag: %q", stdout)
	}
}

// A value arriving with a path separator must not escape the snapshot directory.
func TestSnapshotNamesCannotEscapeTheDirectory(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), sampleNote)

	call(t, input(t, hookInput{
		AgentID: "../../escape", AgentType: "a/b",
	}), dir, noon, "-event", "SubagentStop")

	snaps := snapshots(t, dir)
	if len(snaps) != 1 {
		t.Fatalf("snapshots = %v", snaps)
	}
	if strings.ContainsAny(snaps[0], `/\`) || strings.Contains(snaps[0], "..") {
		t.Errorf("path traversal survived into the filename: %q", snaps[0])
	}
}
