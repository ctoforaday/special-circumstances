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
