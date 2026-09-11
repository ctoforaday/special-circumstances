package recordsql

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A TABLE THIS BINARY HAS AND THE RUN DOES NOT IS NAMED AS AN OLDER RUN. The schema is fixed at a
// run's creation with no migration, so a list table added since is missing from an older run, and
// the reader used to surface SQLite's bare "no such table". A table that exists and fails for
// another reason keeps its own error.
func TestAMissingTableIsNamedAsAnOlderRun(t *testing.T) {
	// tmpRun, not t.TempDir: Open caches the handle, and Windows refuses to remove an open file.
	db, err := Open(filepath.Join(tmpRun(t), "record.db"))
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

// A COLUMN THIS BINARY DECLARES AND THE RUN'S TABLE LACKS IS AN OLDER RUN TOO. `blue_edit` exists
// in every run; `exact_span` was added to it later, so a run created before that reads and writes
// through a table one column short, and SQLite said only "no such column".
func TestAMissingDeclaredColumnIsNamedAsAnOlderRun(t *testing.T) {
	db, err := Open(filepath.Join(tmpRun(t), "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	// SQLite refuses to drop a column a view reads, so the views go first — this database only
	// has to be the older table's shape, not a whole older run.
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'view'`)
	if err != nil {
		t.Fatal(err)
	}
	var views []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		views = append(views, v)
	}
	rows.Close()
	for _, v := range views {
		if _, err := db.Exec(`DROP VIEW "` + v + `"`); err != nil {
			t.Fatal(err)
		}
	}
	// Rebuilt without the column rather than DROP COLUMN, which SQLite refuses while a CHECK
	// constraint names it — exactly the table an older binary created: every column but the new one.
	cols, err := columnsOf(db, "blue_edit")
	if err != nil {
		t.Fatal(err)
	}
	var keep []string
	for c := range cols {
		if c != "exact_span" {
			keep = append(keep, `"`+c+`"`)
		}
	}
	for _, stmt := range []string{
		`CREATE TABLE "blue_edit_older" AS SELECT ` + strings.Join(keep, ", ") + ` FROM "blue_edit"`,
		`DROP TABLE "blue_edit"`,
		`ALTER TABLE "blue_edit_older" RENAME TO "blue_edit"`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	_, err = scanTable(db, "blue_edit", []string{`"exact_span"`})
	for _, want := range []string{"older binary", `"exact_span"`, `"blue_edit"`, "`--seat-id operator migrate --from <runDir> --to <freshDir>`"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("reading a declared column the run lacks: %v — want %s named", err, want)
		}
	}
	// A column the binary does NOT declare is a defect in the read, not an older run.
	if _, err := scanTable(db, "blue_edit", []string{`no_such_column`}); err == nil || strings.Contains(err.Error(), "older binary") {
		t.Errorf("an undeclared column was blamed on an older run: %v", err)
	}
	// THE WRITE PATH MEETS THE SAME TABLE: an edit that records exact_span into an older run.
	_, err = Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{ExactSpan: proto.Bool(true)}))
	if err == nil || !strings.Contains(err.Error(), "older binary") || !strings.Contains(err.Error(), `"exact_span"`) {
		t.Errorf("writing a declared column the run lacks: %v — want the older-binary cause named", err)
	}
}
