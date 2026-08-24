package recordsql

import (
	"os"
	"path/filepath"
	"testing"
)

// THE DERIVED SCHEMA, WRITTEN DOWN SO A HUMAN CAN READ IT.
//
// # Why a green suite is not enough here
//
// Every other check in this package asks whether the schema BEHAVES: does it refuse a ruling that
// precedes its filing, does a closure need a successor, does `carried` fail a merge close. Each one
// asks about a constraint somebody already thought of — which means the suite is blind in exactly
// the direction that matters most for GENERATED output. A column silently missing, a NOT NULL that
// quietly is not, a foreign key never emitted because the field's kind took a different branch:
// none of those fail a test, because no test was written for a constraint nobody knew was absent.
//
// That blindness is worse for derived DDL than for hand-written DDL, and the reason is the whole
// argument for this file: nobody typed it, so nobody has ever read it. A hand-written schema was
// reviewed once by whoever wrote it. This one has been reviewed by no one.
//
// So the golden is the reading. It is the entire schema as a human sees it, and a diff is a
// deliberate change that has to be justified where it is regenerated.
//
// # What this golden is NOT
//
// It is not a second source of truth, and it must never be hand-edited to make a test pass. The
// schema is derived from the descriptors; this is a photograph of the derivation. If the two
// disagree the code is right and the photograph is stale — a hand-patched golden starts asserting
// a schema the tool does not build, which is the failure this whole migration is about.
//
//	UPDATE_GOLDENS=1 go test ./internal/record/recordsql
//
// the repo's convention (difftest/golden_test.go records why it is an env var and not a -update
// flag: a Go test flag is package-scoped, and one command has to drive every suite).
func TestTheDerivedSchemaMatchesItsGolden(t *testing.T) {
	got, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	got += ViewsDDL

	path := filepath.Join("testdata", "schema.sql")
	if os.Getenv("UPDATE_GOLDENS") == "1" {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Log("golden rewritten — READ THE DIFF before committing it; that reading is the whole point of this file")
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no golden schema: %v\n\nGenerate it with UPDATE_GOLDENS=1 and read what it produced", err)
	}
	if string(want) != got {
		// The diff is left to `git diff` after a regeneration rather than printed here: 400 lines of
		// DDL in test output is unreadable, and unreadable output is how a regeneration gets
		// accepted without anyone looking at it — the exact outcome this file exists to prevent.
		t.Errorf("the derived schema differs from testdata/schema.sql.\n\n"+
			"Regenerate with UPDATE_GOLDENS=1 go test ./internal/record/recordsql and READ the diff. "+
			"A change here changes what the record can physically hold: %d bytes now, %d in the golden",
			len(got), len(want))
	}
}
