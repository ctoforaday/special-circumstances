package recordsql

import (
	"fmt"
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
	// THE VIEWS COME BACK, because an older run HAS them — its own, built from its own schema. They
	// were dropped only so the table could be rebuilt beneath them, and the write path reads one of
	// them (`sittings`) for every event: without this the envelope fails first and the body column
	// this test is about is never reached.
	if _, err := db.Exec(ViewsDDL); err != nil {
		t.Fatal(err)
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
	for _, want := range []string{"older binary", `"change"`, "has no", "view", "`--seat-id operator migrate --from <runDir> --to <freshDir>`"} {
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
	for _, want := range []string{"older binary", `"board_counts"`, "defines it differently", "`--seat-id operator migrate --from <runDir> --to <freshDir>`"} {
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
			wants:      []string{"older binary", "`--seat-id operator migrate --from <runDir> --to <freshDir>`"},
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
