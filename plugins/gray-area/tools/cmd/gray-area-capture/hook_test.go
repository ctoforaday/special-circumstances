package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func call(t *testing.T, stdin, projectDir string, st statFunc, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &e, projectDir, noon, st)
	return o.String(), e.String(), code
}

func readManifest(t *testing.T, projectDir, session string) []manifestRow {
	t.Helper()
	b, err := os.ReadFile(manifestPath(projectDir, session))
	if err != nil {
		return nil
	}
	var out []manifestRow
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if l == "" {
			continue
		}
		var r manifestRow
		if err := json.Unmarshal([]byte(l), &r); err != nil {
			t.Fatalf("manifest line is not valid JSON: %v (%q)", err, l)
		}
		out = append(out, r)
	}
	return out
}

// The hard floor: a subagent never loses its turn to this hook.
func TestNeverBlocksTheSubagent(t *testing.T) {
	cases := []struct {
		name       string
		projectDir func(t *testing.T) string
		stdin      string
	}{
		{"no project dir", func(*testing.T) string { return "" }, `{"agent_id":"a"}`},
		{"malformed stdin", func(t *testing.T) string { return t.TempDir() }, "{not json"},
		{"empty stdin", func(t *testing.T) string { return t.TempDir() }, ""},
		{"payload with no agent fields", func(t *testing.T) string { return t.TempDir() }, `{"session_id":"s"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, code := call(t, tc.stdin, tc.projectDir(t), okStat(0)); code != 0 {
				t.Fatalf("exit %d — this hook must never block the subagent", code)
			}
		})
	}
}

// One row per seat, and the path it names is the seat's own file.
func TestCapturesOneRowPerSeat(t *testing.T) {
	dir := t.TempDir()
	seat := filepath.Join(dir, "agent-a1.jsonl")
	if err := os.WriteFile(seat, []byte(`{"type":"assistant"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, s := range []struct{ id, typ, path string }{
		{"a1", "red-auditor", seat},
		{"a2", "blue-lane-1", filepath.Join(dir, "missing.jsonl")},
	} {
		in, _ := json.Marshal(hookInput{
			SessionID: "run-1", AgentID: s.id, AgentType: s.typ,
			TranscriptPath: filepath.Join(dir, "parent.jsonl"), AgentTranscriptPath: s.path,
		})
		if _, _, code := call(t, string(in), dir, statSize); code != 0 {
			t.Fatalf("seat %s exited %d", s.id, code)
		}
	}

	rows := readManifest(t, dir, "run-1")
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if !rows[0].Resolved || rows[0].SizeBytes == 0 || rows[0].AgentType != "red-auditor" {
		t.Errorf("resolved seat row wrong: %+v", rows[0])
	}
	// The unresolved seat is still recorded — visibly, with a reason.
	if rows[1].Resolved || rows[1].CaptureError == "" {
		t.Errorf("unresolved seat should be recorded with a reason: %+v", rows[1])
	}
	if rows[1].AgentID != "a2" {
		t.Errorf("lost the seat identity: %+v", rows[1])
	}
}

// Concurrent runs write to separate manifests keyed by session.
func TestConcurrentRunsDoNotInterleave(t *testing.T) {
	dir := t.TempDir()
	for _, sess := range []string{"run-a", "run-b", "run-a"} {
		in, _ := json.Marshal(hookInput{SessionID: sess, AgentID: "x", AgentTranscriptPath: "/nope"})
		call(t, string(in), dir, statSize)
	}
	if got := len(readManifest(t, dir, "run-a")); got != 2 {
		t.Errorf("run-a rows = %d, want 2", got)
	}
	if got := len(readManifest(t, dir, "run-b")); got != 1 {
		t.Errorf("run-b rows = %d, want 1", got)
	}
}

// A missing project dir is reported once, not silently swallowed — the operator
// should be able to tell capture off from capture broken.
func TestNoProjectDirSaysSo(t *testing.T) {
	_, stderr, code := call(t, `{"agent_id":"a"}`, "", okStat(0))
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, want := range []string{"CLAUDE_PROJECT_DIR", "cwd"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr should name %s, so a reader can tell WHICH input was missing: %q", want, stderr)
		}
	}
}

// ...but "missing" means missing from BOTH places the harness states it. The payload
// carries `cwd`, this hook already parsed it, and it gave up anyway — losing a session's
// whole trajectory for want of a value it had been handed.
func TestProjectDirFallsBackToThePayloadCWD(t *testing.T) {
	dir := t.TempDir()
	_, stderr, code := call(t, `{"session_id":"s1","agent_id":"a","cwd":`+strconv.Quote(dir)+`}`, "", okStat(0))
	if code != 0 {
		t.Fatalf("exit %d (%s)", code, stderr)
	}
	if rows := readManifest(t, dir, "s1"); len(rows) != 1 {
		t.Fatalf("rows = %d, want 1 — the payload said where the project was and capture ignored it: %s", len(rows), stderr)
	}
	// The environment still WINS when both are present: it is the harness's explicit
	// statement, and cwd is only where the process happens to have been started.
	envDir := t.TempDir()
	if _, _, code := call(t, `{"session_id":"s2","agent_id":"a","cwd":`+strconv.Quote(dir)+`}`, envDir, okStat(0)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if rows := readManifest(t, envDir, "s2"); len(rows) != 1 {
		t.Errorf("the payload cwd overrode CLAUDE_PROJECT_DIR — the fallback became the primary")
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, _, code := call(t, "", t.TempDir(), okStat(0), "-version")
	if code != 0 || !strings.Contains(stdout, "gray-area-capture") {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}
}

// SessionStart records where THIS session's transcript is — the thing SubagentStop
// never captures, and the reason a command could not find it without globbing.
func TestSessionStartRecordsTheSessionTranscript(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "s.jsonl")
	if err := os.WriteFile(f, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in := `{"session_id":"S1","transcript_path":` + strconv.Quote(f) + `,"cwd":` + strconv.Quote(dir) + `}`
	if _, stderr, code := call(t, in, dir, okStat(12), "-event", "SessionStart"); code != 0 {
		t.Fatalf("exit %d (%s)", code, stderr)
	}
	rows := readManifest(t, dir, "S1")
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].Kind != "session" {
		t.Errorf("kind = %q, want \"session\" — a consumer asking for the session's transcript must not be handed a seat's", rows[0].Kind)
	}
	if rows[0].TranscriptPath != f || !rows[0].Resolved {
		t.Errorf("row did not record a resolvable transcript: %+v", rows[0])
	}
}

// THE ALARM (plan §11.3). hook-surface-spike.md §3 claims every event carries
// transcript_path, but that was never re-measured for SessionStart and a running
// session cannot fire one to check. If the claim is wrong, this hook must write
// NOTHING and say so — a row with an empty path resolves to nothing later,
// silently, which is the failure this plugin exists to avoid.
func TestSessionStartWithNoTranscriptPathWritesNoRow(t *testing.T) {
	dir := t.TempDir()
	in := `{"session_id":"S1","cwd":` + strconv.Quote(dir) + `}`
	_, stderr, code := call(t, in, dir, okStat(0), "-event", "SessionStart")
	if code != 0 {
		t.Fatalf("a missing field must never cost the session its turn: exit %d", code)
	}
	if _, err := os.Stat(manifestPath(dir, "S1")); err == nil {
		t.Error("a row was written for a payload with no transcript_path — it would resolve to nothing, silently")
	}
	if !strings.Contains(stderr, "transcript_path") {
		t.Errorf("the alarm does not name the missing field: %q", stderr)
	}
}

// The event comes from the wiring, not from guessing at which fields are present.
// Default stays SubagentStop so existing hooks.json entries keep working.
func TestTheEventComesFromTheFlagNotThePayload(t *testing.T) {
	dir := t.TempDir()
	// A payload with BOTH a transcript_path and an agent_transcript_path. Which row
	// gets written must depend on the flag alone.
	in := `{"session_id":"S1","transcript_path":"/t/s.jsonl","agent_transcript_path":"/t/a.jsonl","agent_type":"red","cwd":` + strconv.Quote(dir) + `}`
	if _, _, code := call(t, in, dir, okStat(9)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if rows := readManifest(t, dir, "S1"); len(rows) != 1 || rows[0].Kind != "seat" {
		t.Fatalf("default did not produce a seat row: %+v", rows)
	}
}
