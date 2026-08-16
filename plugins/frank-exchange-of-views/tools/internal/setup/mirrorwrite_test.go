package setup

import (
	"os"
	"path/filepath"
	"testing"
)

// THE MIRROR MUST NOT REPORT FILES IT DID NOT WRITE.
//
// MEASURED 2026-08-16, by reusing MirrorGapPatterns from the seat probe. `out` is
// <runDir>/inputs/red-gap-patterns.md, nothing here created `inputs/`, the os.WriteFile error was
// discarded, and the function returned {Written: true, Files: 55} with nothing on disk. Real runs
// were masked because BuildSkeleton makes the directory earlier in run.go — so the failure was
// reachable by any other caller and invisible to all of them.
//
// Red's ENTIRE accumulated memory ships through this one write. A caller that trusts Written has
// no way to discover the corpus never arrived, which is the same shape as everything in the corpus.
func TestTheMirrorDoesNotClaimFilesItDidNotWrite(t *testing.T) {
	corpus := t.TempDir()
	if err := os.WriteFile(filepath.Join(corpus, "a_pattern.md"), []byte("---\nname: p\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// A run directory with NO inputs/ — the condition every caller outside run.go presents.
	run := t.TempDir()
	r := MirrorGapPatterns([]string{corpus}, run)

	staged := filepath.Join(run, "inputs", "red-gap-patterns.md")
	_, statErr := os.Stat(staged)
	switch {
	case r.Written && statErr != nil:
		t.Fatalf("reported Written with Files=%d and no file exists (%v) — red's memory would be absent from the run while setup reported it staged", r.Files, statErr)
	case !r.Written && statErr == nil:
		t.Fatalf("reported NOT written (%q) but the file is there — the honest failure and the silent success are swapped", r.Reason)
	case !r.Written:
		t.Fatalf("the mirror could not stage into a fresh run directory: %q. It must create inputs/ itself; every caller outside run.go arrives without it", r.Reason)
	}
	if r.Files != 1 {
		t.Errorf("Files = %d, want 1 — the count is what a caller believes about delivery", r.Files)
	}
}
