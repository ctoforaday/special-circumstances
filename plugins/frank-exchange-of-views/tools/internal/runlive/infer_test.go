package runlive

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	if got := InferRunDir(proj); got != run {
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
	if got := InferRunDir(deep); got != run {
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
			if got := InferRunDir(proj); got != "" {
				t.Fatalf("got %q, want empty", got)
			}
		})
	}
}

func TestInferRunDirIsEmptyWithNoMarkerAnywhere(t *testing.T) {
	if got := InferRunDir(t.TempDir()); got != "" {
		t.Fatalf("got %q, want empty when no run is live", got)
	}
}

// An absolute runDir in the marker is used as-is rather than joined onto the project.
func TestInferRunDirHonoursAnAbsoluteRunDir(t *testing.T) {
	proj := t.TempDir()
	elsewhere := t.TempDir()
	mustMkdir(t, filepath.Join(proj, ".claude"))
	writeMarker(t, proj, `{"runs":[{"runDir":`+quote(elsewhere)+`}]}`)
	if got := InferRunDir(proj); got != elsewhere {
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
