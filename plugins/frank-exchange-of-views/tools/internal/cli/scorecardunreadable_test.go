package cli

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// openRunDB opens a run's record directly, for a test that has to build a state the write path
// refuses on purpose.
func openRunDB(t *testing.T, runDir string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(runDir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// A SCORECARD OVER A RECORD THIS BINARY CANNOT READ IS A REFUSAL, NOT A PAGE OF ZEROS.
//
// scorecard discarded FamilyOf's error on the reasoning that "a run with no readable record"
// should render the record-derived rows as "not computed". That collapses two different runs:
// one that recorded nothing, and one whose events this binary cannot decode. Measured against
// run-archive/2026-09-02_quadratic-formula — which holds 40 findings — every row rendered
// `_not computed_ — no findings on the record yet` and the command exited 0.
//
// It is the sharper case of the plausible zero every other read verb on this surface refuses,
// because a scorecard is HARVESTED into feov-memory (#743): the empty answer outlives the run
// that could not produce it and becomes a cross-run memory row asserting nothing was found.
func TestAScorecardRefusesARecordItCannotRead(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-chair",
		"--key", "G1", "--class", "scope-creep", "--quote", "Seven is prime.",
		"--problem", "p", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	// APPEND an event typed with a word this schema does not declare — the shape a record
	// written before a rename has. It is appended rather than retyped because the record's own
	// trigger refuses an edit ("an event cannot be edited after it is written"), which is the
	// guarantee working; the foreign key into the type vocabulary is what refuses the word on
	// the write path, and it comes off to build the state a FORMER binary legitimately left.
	db := openRunDB(t, runDir)
	if _, err := db.Exec(`PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id","ts","type","key")
	                      VALUES ('red-lens-logic', '2026-09-02T00:00:00Z', 'friction', 'former-epoch')`); err != nil {
		t.Fatal(err)
	}

	_, err := run(t, "scorecard", "--run", runDir, "--seat-id", "operator", "--chair", "red")
	if err == nil {
		t.Fatal("a scorecard over an unreadable record was rendered instead of refused — " +
			"every row reads `not computed` and the command exits 0, which is a page of zeros " +
			"about a run nobody could read")
	}
	if !strings.Contains(err.Error(), "does not declare") {
		t.Errorf("the refusal does not name the cause:\n%v", err)
	}
}

// AN EMPTY RUN IS STILL SCORED. This is the half that keeps the refusal from breaking every
// mid-run scorecard: a run that has recorded nothing renders its rows as "not computed" and
// exits 0, which is the honest answer and the one the command exists to give.
func TestAScorecardOverAnEmptyRunIsStillRendered(t *testing.T) {
	runDir := newRun(t)
	out, err := run(t, "scorecard", "--run", runDir, "--seat-id", "operator", "--chair", "red")
	if err != nil {
		t.Fatalf("a run that has recorded nothing was refused: %v", err)
	}
	if !strings.Contains(out, "not computed") {
		t.Errorf("an empty run should render its rows as not computed:\n%s", out)
	}
}
