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

// reopened releases the cached handle and opens the record again: a cache hit never reaches the
// open's schema check, and a second process is what meets a record it did not create.
func reopened(t *testing.T, path string) error {
	t.Helper()
	if err := Close(path); err != nil {
		t.Fatal(err)
	}
	_, err := Open(path)
	return err
}

// notBlamedOnAnOlderRun fails a post-open error that names an older binary or sends the operator
// to migrate: the open already held this record to the declared schema.
func notBlamedOnAnOlderRun(t *testing.T, what string, err error) {
	t.Helper()
	for _, not := range []string{"older binary", "migrate"} {
		if err != nil && strings.Contains(err.Error(), not) {
			t.Errorf("%s was blamed on an older run (%q): %v", what, not, err)
		}
	}
}

// A TABLE THIS BINARY DECLARES AND THE RECORD LACKS IS AN OLDER RUN AT THE OPEN, AND ONLY THERE. The
// open holds every record this binary did not create to its declared schema, so a table that goes
// missing under a handle that passed that check has some other cause, and the read returns SQLite's
// error under the table's name rather than sending the operator to `migrate`.
func TestAMissingTableIsAnOlderRunAtTheOpenAndAPlainErrorAfterIt(t *testing.T) {
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

	err = reopened(t, path)
	for _, want := range append([]string{`"retire_anchors"`, "has no", "table"}, olderRunWants...) {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("opening a record that lacks a declared table: %v — want %s named", err, want)
		}
	}
}

// A COLUMN THIS BINARY DECLARES AND THE RECORD'S TABLE LACKS IS AN OLDER RUN AT THE OPEN, AND ONLY
// THERE. After it, the read still refuses the column by name before the SELECT — SQLite reads the
// quoted name as a string literal and hands back the column's own name as every row's value — and
// the write returns SQLite's error; neither names an older binary.
func TestAMissingDeclaredColumnIsAnOlderRunAtTheOpenAndAPlainErrorAfterIt(t *testing.T) {
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
	// open below must meet the missing column alone.
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

	err = reopened(t, path)
	for _, want := range append([]string{`"blue_edit"`, `"exact_span"`, "has no", "column"}, olderRunWants...) {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("opening a record whose table lacks a declared column: %v — want %s named", err, want)
		}
	}
}

// A VIEW THIS BINARY DECLARES AND THE RUN LACKS IS AN OLDER RUN TOO — the same class as a missing
// column, and worse to walk into. A run's views are fixed when its database is created (ensureSchema
// applies once, and deliberately takes no write lock on later opens), so a binary that adds a view
// meets an older record with SQLite's "no such table: change" — about a name that is not a table, in
// a record where every event the projection needs is present and only the QUERY is gone.
//
// The check runs at Open, on the `existed` path, so this fixture has to look like a record an older
// binary wrote: events in it, the view gone, and a fresh handle.
func TestAMissingDeclaredViewIsNamedAsAnOlderRun(t *testing.T) {
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// `existed` is decided by whether the record holds any event, so an empty database takes the
	// create path and is never checked.
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP VIEW "change"`); err != nil {
		t.Fatal(err)
	}
	// The handle is cached per path, so the second Open is a cache hit until this releases it —
	// and a cache hit never reaches the check.
	if err := Close(path); err != nil {
		t.Fatal(err)
	}

	_, err = Open(path)
	// "has no" rather than just the word "view": the sibling case (present and DIFFERENT) also names
	// the view and the way out, so a looser assertion passes on either arm and neither test then pins
	// which one fired. Proven by mutation — disabling the absent arm sends a missing view down the
	// stale arm, and every weaker assertion here still matched.
	for _, want := range []string{"older binary", `"change"`, "has no", "view", "`feov-migrate --from <runDir> --to <freshDir>`"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("opening a run whose record lacks a declared view: %v — want %s named", err, want)
		}
	}
	if err != nil && strings.Contains(err.Error(), "defines it differently") {
		t.Errorf("an ABSENT view was reported as one defined differently: %v", err)
	}
	// AND THE VIEW SET IS READ FROM THE DDL, not from a list kept beside it: every view ViewsDDL
	// creates is checked, so adding one cannot silently go unwatched. A hand-kept list here would
	// reproduce, one level up, the drift this check exists to catch.
	views, err := declaredViews()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(views), strings.Count(ViewsDDL, "CREATE VIEW "); got != want {
		t.Errorf("declaredViews reports %d view(s) and ViewsDDL creates %d — the check covers less than the DDL writes", got, want)
	}
}

// A VIEW THAT IS PRESENT AND DIFFERENT IS THE QUIETER HALF, and a name check cannot see it. The view
// exists, the query succeeds, and it answers by the definition its CREATING binary held — so a view
// whose rule was FIXED goes on returning the answer it was fixed for, on every record made before the
// fix, with nothing missing to raise.
//
// Met for real: events_w's two correlated subqueries were replaced by MATERIALIZED CTEs, and a record
// created one commit earlier kept the quadratic definition while passing a name check.
func TestAViewDefinedDifferentlyIsNamedAsAnOlderRun(t *testing.T) {
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	// Replaced, not dropped: the name stays and only the rule moves, which is the state under test.
	// "board_counts" is a leaf — no other view selects from it — so this cannot fail for a dependency.
	for _, stmt := range []string{
		`DROP VIEW "board_counts"`,
		`CREATE VIEW "board_counts" AS SELECT 1 AS "open_gaps", 1 AS "closed_gaps"`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	if err := Close(path); err != nil {
		t.Fatal(err)
	}

	_, err = Open(path)
	for _, want := range []string{"older binary", `"board_counts"`, "defines it differently", "`feov-migrate --from <runDir> --to <freshDir>`"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("opening a run whose view is defined differently: %v — want %s named", err, want)
		}
	}
	// AND IT IS NOT REPORTED AS MISSING. The two are different repairs to describe even though both
	// end at migrate, and a seat told a view is absent when it is present reads the record as damaged.
	if err != nil && strings.Contains(err.Error(), "has no") {
		t.Errorf("a view that is present and different was reported as absent: %v", err)
	}
}

// THE EPOCHS DECIDE WHICH PARTY IS STALE, in both directions, from the record's own field.
//
// This asked the question by SHAPE — did the record carry a table or view this binary does not
// declare — and shape cannot tell a stale binary from a view the current binary deliberately DROPPED.
// That made removing a view a two-step change and could name the wrong party. The record states its
// epoch now, so one integer answers it.
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
			if _, err := db.Exec(`UPDATE "schema_epoch" SET "epoch" = ? WHERE "id" = 1`, tc.epoch); err != nil {
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
			for _, want := range append(tc.wants, fmt.Sprint(tc.epoch), fmt.Sprint(recordpb.EventSchema)) {
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

// AN EXTRA TABLE IS NO LONGER A REFUSAL, and that is the point of moving to the epoch: a view or table
// the binary does not declare may simply have been DROPPED by it, which is an ordinary forward change.
// Shape said "your binary is stale" to both cases and could not tell them apart.
func TestAnUndeclaredTableAloneIsNotAStaleBinary(t *testing.T) {
	path := filepath.Join(tmpRun(t), "record.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Insert(db, event(t, 1, recordpb.EventType_EVENT_TYPE_BLUE_EDIT, &recordpb.BlueEdit{})); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE "left_over" ("x" TEXT) STRICT`); err != nil {
		t.Fatal(err)
	}
	if err := Close(path); err != nil {
		t.Fatal(err)
	}
	if _, err := Open(path); err != nil {
		t.Errorf("a record carrying a table this binary does not declare was refused, though its epoch agrees: %v", err)
	}
}
