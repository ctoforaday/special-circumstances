package runlive

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// t.TempDir() AND NOT recordtest.TmpRun, deliberately. TmpRun is TempDir plus a cleanup that
// closes a cached database handle before the directory goes — required of any test that OPENS a
// record, and the trap recordtest.Main exists to catch. These tests open none: InferRunDir stats
// directories and reads one JSON marker. Reaching for TmpRun here would put `record`, and with it
// a SQLite driver, into this leaf package's test graph for a cleanup with nothing to clean.

// Ten of the first live run's 55 tool-call errors were a missing --run. The seat
// copies the engine's register line, then improvises later verbs and drops the flag.
// It cannot be carried in the environment: shell state does not persist between tool
// calls, and every subagent shares the parent's CLAUDE_CODE_SESSION_ID, so there is no
// per-seat identity to key on. The marker on disk already knows the answer.
func TestInferRunDirReadsTheLiveMarker(t *testing.T) {
	proj := t.TempDir()
	run := filepath.Join(proj, "research", "2026-07-18_topic")
	mustMkdir(t, run)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"research/2026-07-18_topic"}]}`)

	if got := InferRunDir(proj).Dir; got != run {
		t.Fatalf("InferRunDir(%q) = %q, want %q", proj, got, run)
	}
}

// A seat's cwd is usually deeper than the project root.
func TestInferRunDirWalksUpFromASubdirectory(t *testing.T) {
	proj := t.TempDir()
	run := filepath.Join(proj, "research", "r")
	mustMkdir(t, run)
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":"research/r"}]}`)

	deep := filepath.Join(proj, "plugins", "x", "y")
	mustMkdir(t, deep)
	if got := InferRunDir(deep).Dir; got != run {
		t.Fatalf("from %q got %q, want %q", deep, got, run)
	}
}

// Attaching a seat's events to the WRONG run is worse than the error it replaces, so
// an unusable marker yields nothing and the caller's --run error stands.
func TestInferRunDirRefusesAnUnusableMarker(t *testing.T) {
	for _, tc := range []struct{ name, body string }{
		{"run directory does not exist", `{"runs":[{"runDir":"research/gone"}]}`},
		{"no runDir field", `{"started":"2026-07-19T00:00:00Z"}`},
		{"not json at all", `this is not json`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proj := t.TempDir()
			mustMkdir(t, filepath.Join(proj, ".claude"))
			writeMarker(t, proj, tc.body)
			if got := InferRunDir(proj).Dir; got != "" {
				t.Fatalf("got %q, want empty", got)
			}
		})
	}
}

func TestInferRunDirIsEmptyWithNoMarkerAnywhere(t *testing.T) {
	if got := InferRunDir(t.TempDir()).Dir; got != "" {
		t.Fatalf("got %q, want empty when no run is live", got)
	}
}

// An absolute runDir in the marker is used as-is rather than joined onto the project.
func TestInferRunDirHonoursAnAbsoluteRunDir(t *testing.T) {
	proj := t.TempDir()
	elsewhere := t.TempDir()
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":`+quote(elsewhere)+`}]}`)
	if got := InferRunDir(proj).Dir; got != elsewhere {
		t.Fatalf("got %q, want %q", got, elsewhere)
	}
}

func mustMkdir(t *testing.T, d string) {
	t.Helper()
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeMarker(t *testing.T, proj, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(proj, ".claude", "run-live.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func quote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// EVERY REASON IS DISTINCT, AND ONLY TWO ARE FAULTS. A bare "" used to answer all six; a caller that
// must tell a human about a broken marker cannot do it from a zero value the healthy cases share —
// and must not raise an alarm for two runs legitimately open at once.
func TestInferRunDirSaysWhy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		marker *string // nil: no marker at all
		dirs   []string
		why    Inference
		fault  bool
		hasDir bool
	}{
		{name: "no marker", why: NoMarker},
		{name: "resolved", marker: ptr(`{"runs":[{"runDir":"run"}]}`), dirs: []string{"run"}, why: Resolved, hasDir: true},
		{name: "no run open", marker: ptr(`{"runs":[]}`), why: NoRunOpen},
		{name: "two runs open", marker: ptr(`{"runs":[{"runDir":"a"},{"runDir":"b"}]}`), dirs: []string{"a", "b"}, why: Ambiguous},
		{name: "unreadable", marker: ptr("{ not json"), why: Unreadable, fault: true},
		{name: "stale", marker: ptr(`{"runs":[{"runDir":"gone"}]}`), why: Stale, fault: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			proj := t.TempDir()
			for _, d := range tc.dirs {
				mustMkdir(t, filepath.Join(proj, d))
			}
			if tc.marker != nil {
				mustMkdir(t, filepath.Join(proj, ".claude"))
				writeMarker(t, proj, *tc.marker)
			}
			got := InferRunDir(proj)
			if got.Why != tc.why {
				t.Fatalf("Why = %v, want %v", got.Why, tc.why)
			}
			if got.Why.Fault() != tc.fault {
				t.Errorf("Fault() = %v, want %v", got.Why.Fault(), tc.fault)
			}
			if (got.Dir != "") != tc.hasDir {
				t.Errorf("Dir = %q", got.Dir)
			}
			if tc.marker != nil && got.MarkerDir != proj {
				t.Errorf("MarkerDir = %q, want %q — a fault is scoped to the project its marker is in", got.MarkerDir, proj)
			}
		})
	}
}

func ptr(s string) *string { return &s }

// NOWHERE TO START: no cwd, no CLAUDE_PROJECT_DIR, and a working directory that no longer exists.
// That fell through to NoMarker — "not in a run" — the one answer it cannot honestly give.
func TestNowhereToStartIsAFault(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a process's working directory cannot be deleted from under it on Windows")
	}
	t.Setenv("CLAUDE_PROJECT_DIR", "")
	gone := filepath.Join(t.TempDir(), "gone")
	mustMkdir(t, gone)
	t.Chdir(gone)
	if err := os.Remove(gone); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Getwd(); err == nil {
		t.Skip("Getwd still succeeds with its directory removed on this platform — the premise never took")
	}
	got := InferRunDir("")
	if got.Why != Unlocatable || !got.Why.Fault() {
		t.Fatalf("Why = %v, fault %v; want Unlocatable, a fault", got.Why, got.Why.Fault())
	}
	if FaultDetail(got) == "" {
		t.Error("a fault with no words for its fix")
	}
}
