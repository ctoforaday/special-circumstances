package runlive

import (
	"os"
	"path/filepath"
	"testing"
)

// A READER NEVER BLOCKS THE WRITER'S REMOVAL. The run-live files are polled while capture removes
// them; on Windows a handle opened without FILE_SHARE_DELETE turns that removal into a sharing
// violation the writer drops, and the run stays live after its capture (#1484).
func TestAnOpenReaderDoesNotBlockRemoval(t *testing.T) {
	p := filepath.Join(t.TempDir(), "run-live.json")
	if err := os.WriteFile(p, []byte(`{"runs":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := openSharingDelete(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := os.Remove(p); err != nil {
		t.Fatalf("removing the marker while a reader holds it: %v", err)
	}
}

// A READER NEVER BLOCKS THE WRITER'S REPLACEMENT. Every marker write renames a complete file over
// the old one (replaceFile); on Windows that rename succeeds over a file a reader holds only because
// the reader shares delete (#1552).
func TestAnOpenReaderDoesNotBlockReplacement(t *testing.T) {
	p := filepath.Join(t.TempDir(), "run-live.json")
	if err := os.WriteFile(p, []byte(`{"runs":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := openSharingDelete(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := replaceFile(p, []byte(`{"runs":[{"runDir":"x"}]}`)); err != nil {
		t.Fatalf("replacing the marker while a reader holds it: %v", err)
	}
	if b, err := readSharingDelete(p); err != nil || string(b) != `{"runs":[{"runDir":"x"}]}` {
		t.Fatalf("after the replacement the path reads %q, %v", b, err)
	}
}
