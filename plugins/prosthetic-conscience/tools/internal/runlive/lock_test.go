package runlive

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
)

// A READER WAITS OUT A WRITER. frank-exchange-of-views's writers hold the marker's lock exclusive
// across their rename, because on Windows an open landing inside a rename finds no file and a guard
// reads that as no live run. Read takes the same lock shared: while a writer holds it, Read does not
// return; once it is released, Read sees the marker.
func TestReadWaitsForAWriterHoldingTheMarkerLock(t *testing.T) {
	proj := t.TempDir()
	dir := filepath.Join(proj, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run-live.json"), []byte(`{"runs":[{"runDir":"research/x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// The writer's lock, taken exactly as frank-exchange-of-views's lockMarker takes it.
	writer := flock.New(filepath.Join(dir, markerLock), flock.SetFlag(os.O_CREATE|os.O_RDWR), flock.SetPermissions(0o644))
	if err := writer.Lock(); err != nil {
		t.Fatal(err)
	}
	done := make(chan State, 1)
	go func() { done <- Read(proj) }()
	select {
	case <-done:
		writer.Close()
		t.Fatal("Read returned while a writer held the marker lock")
	case <-time.After(200 * time.Millisecond):
	}
	writer.Close()
	select {
	case s := <-done:
		if len(s.Runs) != 1 {
			t.Fatalf("after the writer released, Read saw %+v", s)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Read did not return after the writer released the lock")
	}
}

// A READER NEVER CREATES THE LOCK FILE. No lock file means no writer has written the marker, so
// there is no rename to wait out: Read reads without it and leaves the directory as it found it.
func TestReadCreatesNoLockFile(t *testing.T) {
	proj := t.TempDir()
	dir := filepath.Join(proj, ".claude")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run-live.json"), []byte(`{"runs":[{"runDir":"research/x"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if s := Read(proj); len(s.Runs) != 1 {
		t.Fatalf("with no lock file, Read saw %+v", s)
	}
	if _, err := os.Stat(filepath.Join(dir, markerLock)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Read created the marker's lock file: %v", err)
	}
}
