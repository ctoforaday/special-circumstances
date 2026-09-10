package recordsql

import (
	"path/filepath"
	"strings"
	"testing"
)

// A TABLE THIS BINARY HAS AND THE RUN DOES NOT IS NAMED AS AN OLDER RUN. The schema is fixed at a
// run's creation with no migration, so a list table added since is missing from an older run, and
// the reader used to surface SQLite's bare "no such table". A table that exists and fails for
// another reason keeps its own error.
func TestAMissingTableIsNamedAsAnOlderRun(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE "retire_anchors"`); err != nil {
		t.Fatal(err)
	}
	_, err = scanLists(db, "retire_anchors")
	if err == nil || !strings.Contains(err.Error(), "older binary") || !strings.Contains(err.Error(), "retire_anchors") {
		t.Errorf("reading a list table the run lacks: %v — want the older-binary cause named", err)
	}
	// ...and the way out: the repository's migrate command, spelled as the operator types it.
	if err == nil || !strings.Contains(err.Error(), "`--seat-id operator migrate --from <runDir> --to <freshDir>`") {
		t.Errorf("the older-run message does not name the migrate invocation: %v", err)
	}
	// Unquoted: SQLite reads a double-quoted unknown identifier as a string literal, not an error.
	if _, err := scanTable(db, "retire", []string{`no_such_column`}); err == nil || strings.Contains(err.Error(), "older binary") {
		t.Errorf("a failure on a table that exists was blamed on an older run: %v", err)
	}
}
