package runlive

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
	lf, err := os.OpenFile(filepath.Join(dir, markerLock), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if err := lockFile(lf, true); err != nil {
		t.Fatal(err)
	}
	done := make(chan State, 1)
	go func() { done <- Read(proj) }()
	select {
	case <-done:
		lf.Close()
		t.Fatal("Read returned while a writer held the marker lock")
	case <-time.After(200 * time.Millisecond):
	}
	lf.Close()
	select {
	case s := <-done:
		if len(s.Runs) != 1 {
			t.Fatalf("after the writer released, Read saw %+v", s)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Read did not return after the writer released the lock")
	}
}
