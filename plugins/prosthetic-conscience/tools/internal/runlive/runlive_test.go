package runlive

import (
	"os"
	"path/filepath"
	"testing"
)

// writeMarker plants .claude/run-live.json under a fresh project dir and returns it.
func writeMarker(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	claude := filepath.Join(dir, ".claude")
	if err := os.MkdirAll(claude, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(claude, "run-live.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestReadParsesLiveMarker(t *testing.T) {
	dir := writeMarker(t, `{"runDir":"research/2026-07-18_x","pinnedPaths":["ideas/backlog.md","research/old"],"started":"2026-07-18T09:00:00Z"}`)
	m := Read(dir)
	if m == nil {
		t.Fatal("a well-formed marker must read as LIVE")
	}
	if m.RunDir != "research/2026-07-18_x" {
		t.Errorf("runDir = %q", m.RunDir)
	}
	if m.Started != "2026-07-18T09:00:00Z" {
		t.Errorf("started = %q", m.Started)
	}
	if len(m.PinnedPaths) != 2 || m.PinnedPaths[0] != "ideas/backlog.md" || m.PinnedPaths[1] != "research/old" {
		t.Errorf("pinnedPaths lost order or content: %#v", m.PinnedPaths)
	}
}

// The package contract: "missing or malformed marker both mean not live — guards
// fail open". Each case below is one way the marker can be unreadable; every one
// of them MUST yield nil rather than a panic or a half-populated marker.
func TestReadFailsOpen(t *testing.T) {
	cases := []struct {
		name string
		dir  func(t *testing.T) string
	}{
		{"empty projectDir never reads a relative path", func(*testing.T) string { return "" }},
		{"no .claude directory at all", func(t *testing.T) string { return t.TempDir() }},
		{"malformed json", func(t *testing.T) string { return writeMarker(t, `{"runDir":`) }},
		{"truncated json", func(t *testing.T) string { return writeMarker(t, `{"runDir":"research/x"`) }},
		{"empty file", func(t *testing.T) string { return writeMarker(t, ``) }},
		{"json array, not an object", func(t *testing.T) string { return writeMarker(t, `["research/x"]`) }},
		{"wrong field type", func(t *testing.T) string { return writeMarker(t, `{"pinnedPaths":"not-a-list"}`) }},
		{"marker is a directory, not a file", func(t *testing.T) string {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, ".claude", "run-live.json"), 0o755); err != nil {
				t.Fatal(err)
			}
			return dir
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if m := Read(c.dir(t)); m != nil {
				t.Fatalf("want nil (not live), got %+v", m)
			}
		})
	}
}

// The empty projectDir guard is load-bearing on its own: without it, filepath.Join
// would build the RELATIVE path ".claude/run-live.json" and the guard's verdict
// would depend on the caller's working directory. Planting exactly that file in
// the process's cwd must still read as "not live" when projectDir is empty.
func TestReadWithEmptyDirIgnoresCwdMarker(t *testing.T) {
	dir := writeMarker(t, `{"runDir":"research/cwd-trap"}`)
	t.Chdir(dir)
	if m := Read(""); m != nil {
		t.Fatalf("empty projectDir must not resolve against cwd, got %+v", m)
	}
	// Sanity: the same marker IS live when the project dir is named explicitly,
	// proving the case above failed for the right reason.
	if m := Read(dir); m == nil || m.RunDir != "research/cwd-trap" {
		t.Fatalf("explicit projectDir should have read the marker, got %+v", m)
	}
}

// A JSON object that parses but carries no fields still reads as LIVE. That is the
// documented "any readable marker means a run is live" direction — recorded here so
// the fail-open/fail-closed choice can never change silently.
func TestReadEmptyObjectIsStillLive(t *testing.T) {
	m := Read(writeMarker(t, `{}`))
	if m == nil {
		t.Fatal("an empty JSON object is a readable marker and must read as live")
	}
	if m.RunDir != "" || m.Started != "" || m.PinnedPaths != nil {
		t.Fatalf("empty object must yield a zero Marker, got %+v", m)
	}
}

// Unknown keys are forward-compatibility, not corruption: setup-research-run.mjs may
// record more than this struct models, and that must not disarm the guards.
func TestReadIgnoresUnknownFields(t *testing.T) {
	m := Read(writeMarker(t, `{"runDir":"research/x","round":3,"seats":{"blue":"opus"},"started":"t"}`))
	if m == nil || m.RunDir != "research/x" || m.Started != "t" {
		t.Fatalf("unknown fields must be ignored, got %+v", m)
	}
}
