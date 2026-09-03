package recordsql

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestOrphanedHandlesReportsADatabaseWhoseFileIsGone(t *testing.T) {
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

	// THE OS HANDLE IS DROPPED FIRST, AND THE CACHE ENTRY IS LEFT BEHIND. That is exactly the
	// state OrphanedHandles reports — a cached path with no file — and constructing it this way
	// is the only way to construct it on BOTH platforms.
	//
	// The first draft of this test removed the directory out from under a LIVE handle, which is a
	// POSIX manoeuvre. Windows refuses it, and refusing it is the entire defect this gate exists
	// for — so that test could only ever run where the defect is silent, and the Windows leg
	// failed it with the very message the gate is about. A test of a cross-platform detector must
	// not be built out of the one operation the platforms disagree on.
	openMu.Lock()
	db := openCache[abs]
	openMu.Unlock()
	if db == nil {
		t.Fatalf("Open did not cache a handle for %s", abs)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("closing the handle: %v", err)
	}
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
		t.Fatalf("OrphanedHandles did not report %s after its file was removed; it reported %v", abs, OrphanedHandles())
	}
}

// THE PLATFORM DIFFERENCE ITSELF, asserted rather than assumed.
//
// The whole gate rests on it and it is stated in comments across three packages with nothing
// checking it: on Linux an open file can be unlinked, so the removal SUCCEEDS, the test passes,
// and the leak is invisible; on Windows the same removal FAILS, which is why every one of the
// eleven instances was caught there, a push and a CI round-trip after it was written.
//
// It SKIPS on Windows rather than asserting the failure text — the message belongs to the OS and
// pinning it would be pinning someone else's wording. What is asserted is the half the gate needs:
// where the removal succeeds, the cache is left holding a path with no file behind it.
func TestUnlinkingAnOpenDatabaseIsHowTheLeakGoesUnnoticed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows refuses to remove a directory holding an open file — this test performs the manoeuvre that platform makes impossible, which is precisely why the leak is invisible on Linux and fatal here")
	}
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

	// The handle stays OPEN across this removal. That is the whole experiment.
	if err := os.RemoveAll(filepath.Join(dir, "run")); err != nil {
		t.Fatalf("removing a directory holding an open database failed on %s, which this gate's whole design assumes it does not: %v", runtime.GOOS, err)
	}
	var found bool
	for _, p := range OrphanedHandles() {
		if p == abs {
			found = true
		}
	}
	if !found {
		t.Fatalf("the database was unlinked while open and OrphanedHandles did not report %s; it reported %v", abs, OrphanedHandles())
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
