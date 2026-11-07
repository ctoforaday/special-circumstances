package runtest

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A REFUSED DROP RELEASES THE HANDLE TOO: the seam opens the record through the process's one cached
// handle, and a read that fails part-way still leaves it closed, so nothing holds the file open after
// the run directory goes (recordsql.OrphanedHandles).
func TestDropReopenedReleasesTheHandleOnARefusal(t *testing.T) {
	dir := t.TempDir()
	run := New(t, dir)
	if err := os.MkdirAll(run.Records(), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(run.Records(), "record.db")
	db, err := recordsql.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// The read the seam makes fails: the table it reads is gone.
	if _, err := db.Exec(`DROP TABLE "blue_edit_reopened"`); err != nil {
		t.Fatal(err)
	}
	if _, err := DropReopened(run, "K-0000000a"); err == nil {
		t.Fatal("DropReopened read a table that is not there and reported no error")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("the run directory cannot be removed after the refusal: %v", err)
	}
	if slices.Contains(recordsql.OrphanedHandles(), abs) {
		_ = recordsql.Close(path)
		t.Errorf("DropReopened's refusal left the handle on %s open", abs)
	}
}
