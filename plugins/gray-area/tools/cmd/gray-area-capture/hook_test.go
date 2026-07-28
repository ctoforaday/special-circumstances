package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
	if !strings.Contains(stderr, "CLAUDE_PROJECT_DIR") {
		t.Errorf("stderr should name the missing variable, got %q", stderr)
	}
}

func TestVersionFlag(t *testing.T) {
	stdout, _, code := call(t, "", t.TempDir(), okStat(0), "-version")
	if code != 0 || !strings.Contains(stdout, "gray-area-capture") {
		t.Fatalf("code=%d stdout=%q", code, stdout)
	}
}
