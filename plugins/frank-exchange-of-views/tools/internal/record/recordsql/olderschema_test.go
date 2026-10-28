package recordsql

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// olderRunWants is what the open's refusal of a record behind this binary names: the cause and the
// way out, spelled as the operator types it.
var olderRunWants = []string{"older binary", "`feov-migrate --from <runDir> --to <freshDir>`"}

// notBlamedOnAnOlderRun fails a post-open error that names an older binary or sends the operator
// to migrate: the open already held this record to this binary's epoch.
func notBlamedOnAnOlderRun(t *testing.T, what string, err error) {
	t.Helper()
	for _, not := range []string{"older binary", "migrate"} {
		if err != nil && strings.Contains(err.Error(), not) {
			t.Errorf("%s was blamed on an older run (%q): %v", what, not, err)
		}
	}
}

// A TABLE THAT GOES MISSING UNDER AN OPEN HANDLE IS A PLAIN ERROR. The open holds every record this
// binary did not create to its epoch, and an epoch names its tables, so a missing table has some
// other cause, and the read returns SQLite's error under the table's name rather than sending the
// operator to `migrate`.
func TestATableDroppedUnderAnOpenHandleIsAPlainError(t *testing.T) {
	// tmpRun, not t.TempDir: Open caches the handle, and Windows refuses to remove an open file.
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP TABLE "retire_anchors"`); err != nil {
		t.Fatal(err)
	}
	_, err = scanLists(db, "retire_anchors")
	for _, want := range []string{"reading retire_anchors", "no such table"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("reading a list table dropped under an open handle: %v — want %q", err, want)
		}
	}
	notBlamedOnAnOlderRun(t, "a table dropped under a handle the open already checked", err)
	// Unquoted: SQLite reads a double-quoted unknown identifier as a string literal, not an error.
	_, err = scanTable(db, "retire", []string{`no_such_column`})
	if err == nil {
		t.Error("a read naming a column no table has returned rows")
	}
	notBlamedOnAnOlderRun(t, "a failure on a table that exists", err)
}

// A COLUMN THAT GOES MISSING UNDER AN OPEN HANDLE IS A PLAIN ERROR. The read refuses the column by
// name before the SELECT — SQLite reads the quoted name as a string literal and hands back the
// column's own name as every row's value — and the write returns SQLite's error; neither names an
// older binary.
func TestAColumnMissingUnderAnOpenHandleIsAPlainError(t *testing.T) {
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	// SQLite refuses to drop a column a view reads, so the views go first.
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
	// constraint names it: every column but the one under test.
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
	// THE VIEWS COME BACK, as this binary defines them. They were dropped only so the table could be
	// rebuilt beneath them; the write path reads one of them (`sittings`) for every event, and the
	// write below must meet the missing column alone.
	if _, err := db.Exec(ViewsDDL); err != nil {
		t.Fatal(err)
	}
	_, err = scanTable(db, "blue_edit", []string{`"exact_span"`})
	for _, want := range []string{"reading blue_edit", "the table has no column", `"exact_span"`} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("reading a column the table lacks: %v — want %s named", err, want)
		}
	}
	notBlamedOnAnOlderRun(t, "a read of a column missing under a handle the open already checked", err)
	// A column the binary does not declare reads the same way.
	_, err = scanTable(db, "blue_edit", []string{`no_such_column`})
	if err == nil {
		t.Error("a read naming a column no table has returned rows")
	}
	notBlamedOnAnOlderRun(t, "an undeclared column", err)
	// THE WRITE PATH MEETS THE SAME TABLE.
	_, err = Insert(db, event(t, 2, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{ExactSpan: proto.Bool(true)}))
	for _, want := range []string{"recording a blue_edit", "exact_span"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("writing a column the table lacks: %v — want %s named", err, want)
		}
	}
	notBlamedOnAnOlderRun(t, "a write of a column missing under a handle the open already checked", err)
}

// THE EPOCHS DECIDE WHICH PARTY IS STALE, in both directions, from the record's own field.
//
// Met for real, and it cost a whole sitting: on one run a seat typed a bare tool name and PATH handed
// it the HOST's plugin cache — the same plugin version, an older event-schema epoch — instead of the
// run's own binary. Its write died on a raw SQLite constraint, every later call was refused for having
// not registered, and that sitting made 100 tool calls and recorded nothing.
func TestTheEpochsDecideWhichPartyIsStale(t *testing.T) {
	for _, tc := range []struct {
		name       string
		epoch      int
		wants      []string
		mustNotSay string
	}{
		{
			// The record is AHEAD: this binary is the stale party and migration cannot help, because
			// the replayer is the thing that is behind.
			name:       "record ahead of the binary",
			epoch:      recordpb.EventSchema + 1,
			wants:      []string{"NEWER binary", "THIS BINARY is the stale party", "/.bin/feov-record"},
			mustNotSay: "Migrate it",
		},
		{
			// The record is BEHIND: an older run, and migrate is exactly the remedy.
			name:       "record behind the binary",
			epoch:      recordpb.EventSchema - 1,
			wants:      []string{"older binary", "`feov-migrate --from <runDir> --to <freshDir>`"},
			mustNotSay: "stale party",
		},
		{
			// The record states NO epoch: it predates the field, so it is an older run.
			name:       "record that states no epoch",
			epoch:      0,
			wants:      []string{"states no event-schema epoch", "older binary", "`feov-migrate --from <runDir> --to <freshDir>`"},
			mustNotSay: "stale party",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpRun(t), "record.db")
			db, err := Open(path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
				t.Fatal(err)
			}
			stamp := `DROP TABLE "schema_epoch"`
			args := []any{}
			if tc.epoch != 0 {
				stamp, args = `UPDATE "schema_epoch" SET "epoch" = ? WHERE "id" = 1`, []any{tc.epoch}
			}
			if _, err := db.Exec(stamp, args...); err != nil {
				t.Fatal(err)
			}
			// The handle is cached per path, and a cache hit never reaches the check.
			if err := Close(path); err != nil {
				t.Fatal(err)
			}

			_, err = Open(path)
			if err == nil {
				t.Fatalf("a record at epoch %d opened against a binary at %d", tc.epoch, recordpb.EventSchema)
			}
			// BOTH EPOCHS ARE NAMED. A refusal that says only "wrong epoch" leaves the reader unable to
			// tell which side to change.
			wants := append(tc.wants, fmt.Sprint(recordpb.EventSchema))
			if tc.epoch != 0 {
				wants = append(wants, fmt.Sprint(tc.epoch))
			}
			for _, want := range wants {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal does not name %s: %v", want, err)
				}
			}
			if strings.Contains(err.Error(), tc.mustNotSay) {
				t.Errorf("refusal offers the OTHER direction's remedy (%q): %v", tc.mustNotSay, err)
			}
		})
	}
}

// THE OPEN TRUSTS THE EPOCH. A record at this binary's epoch opens whatever else it holds: every
// schema change bumps the epoch (TestEveryDDLChangeBumpsTheEpoch), so a table this binary does not
// declare, or a view defined otherwise, at the same epoch was put there by hand, and the open does
// not rebuild the declared schema to look for it (#1614).
func TestTheOpenTrustsTheEpoch(t *testing.T) {
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	// "board_counts" is a leaf — no other view selects from it — so replacing it cannot fail for a
	// dependency.
	for _, stmt := range []string{
		`CREATE TABLE "left_over" ("x" TEXT) STRICT`,
		`DROP VIEW "board_counts"`,
		`CREATE VIEW "board_counts" AS SELECT 1 AS "open_gaps", 1 AS "closed_gaps"`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	// The handle is cached per path, and a cache hit never reaches the check.
	if err := Close(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err != nil {
		t.Errorf("a record at this binary's epoch was refused: %v", err)
	}
}
