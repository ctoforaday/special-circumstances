package recordsql

import (
	"os"
	"path/filepath"
	"testing"
)

// The detector behind the #666 gate, tested both ways.
//
// A leak check that reports everything is as useless as one that reports nothing, and both pass a
// suite where no test happens to leak. So the honest zero is asserted as hard as the catch: a live
// database whose file is still there must NOT be reported, or every guarded package fails for the
// wrong reason and the gate is turned off within a day.
//
// These tests reach into the cache directly because they are in the package that owns it. They also
// clean up after themselves — a test of the leak detector that leaks would trip this package's own
// TestMain, which would be a funny way to find out.

func TestOrphanedHandlesReportsADatabaseWhoseDirectoryIsGone(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "run", "records")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatalf("making the run: %v", err)
	}
	path := filepath.Join(inner, "record.db")
	if _, err := Open(path); err != nil {
		t.Fatalf("opening: %v", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolving: %v", err)
	}
	t.Cleanup(func() { _ = CloseUnder(dir) })

	// Nothing is orphaned while the file is there — the honest zero, asserted BEFORE the removal
	// so a detector that reports every cached handle fails here rather than passing below.
	for _, p := range OrphanedHandles() {
		if p == abs {
			t.Fatalf("%s was reported orphaned while its file still exists", abs)
		}
	}

	// This is the Linux half of the trap, performed deliberately: removing a directory out from
	// under an open handle SUCCEEDS, and nothing about the removal complains.
	if err := os.RemoveAll(filepath.Join(dir, "run")); err != nil {
		t.Fatalf("removing the run directory: %v", err)
	}

	var found bool
	for _, p := range OrphanedHandles() {
		if p == abs {
			found = true
		}
	}
	if !found {
		t.Fatalf("OrphanedHandles did not report %s after its directory was removed; it reported %v", abs, OrphanedHandles())
	}
}

func TestOrphanedHandlesReportsNothingForALiveDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "record.db")
	if _, err := Open(path); err != nil {
		t.Fatalf("opening: %v", err)
	}
	t.Cleanup(func() { _ = CloseUnder(dir) })
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolving: %v", err)
	}
	for _, p := range OrphanedHandles() {
		if p == abs {
			t.Fatalf("a live database was reported orphaned: %s", abs)
		}
	}
}

func TestReleasingAHandleTakesItOutOfTheOrphanReport(t *testing.T) {
	dir := t.TempDir()
	inner := filepath.Join(dir, "records")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatalf("making the records dir: %v", err)
	}
	path := filepath.Join(inner, "record.db")
	if _, err := Open(path); err != nil {
		t.Fatalf("opening: %v", err)
	}
	// The release is what recordtest.TmpRun does in its cleanup; after it, removing the directory
	// leaves nothing cached to orphan. This is the fix's side of the same experiment.
	if err := CloseUnder(dir); err != nil {
		t.Fatalf("releasing: %v", err)
	}
	if err := os.RemoveAll(inner); err != nil {
		t.Fatalf("removing: %v", err)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("resolving: %v", err)
	}
	for _, p := range OrphanedHandles() {
		if p == abs {
			t.Fatalf("%s was still cached after CloseUnder released it", abs)
		}
	}
}
