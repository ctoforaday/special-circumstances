package setup

import (
	"os"
	"path/filepath"
	"testing"
)

// THE MIRROR MUST NOT REPORT FILES IT DID NOT WRITE.
//
// MEASURED 2026-08-16, by reusing a mirror from the seat probe: nothing created `inputs/`, the
// os.WriteFile error was discarded, and it returned {Written: true, Files: 55} with nothing on
// disk. Real runs were masked because BuildSkeleton makes the directory earlier in run.go — so the
// failure was reachable by any other caller and invisible to all of them.
//
// A count that outruns the disk is the failure — a run with no law is ordinary and says so, while
// one reporting a corpus it does not have looks complete to every reader after it.
func TestTheMirrorDoesNotClaimFilesItDidNotWrite(t *testing.T) {
	corpus := t.TempDir()
	if err := os.WriteFile(filepath.Join(corpus, "a_statute.md"), []byte("---\nname: s\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A run directory with NO inputs/ — the condition every caller outside run.go presents.
	run := t.TempDir()
	r := MirrorLaw(corpus, runOf(t, run))

	staged := filepath.Join(run, "inputs", "law", "a_statute.md")
	_, statErr := os.Stat(staged)
	switch {
	case r.Written && statErr != nil:
		t.Fatalf("reported Written with Files=%d and no file exists (%v) — the law corpus would be absent from the run while setup reported it staged", r.Files, statErr)
	case !r.Written && statErr == nil:
		t.Fatalf("reported NOT written (%q) but the file is there — the honest failure and the silent success are swapped", r.Reason)
	case !r.Written:
		t.Fatalf("the mirror could not stage into a fresh run directory: %q. It must create inputs/ itself; every caller outside run.go arrives without it", r.Reason)
	}
	if r.Files != 1 {
		t.Errorf("Files = %d, want 1 — the count is what a caller believes about delivery", r.Files)
	}
}
