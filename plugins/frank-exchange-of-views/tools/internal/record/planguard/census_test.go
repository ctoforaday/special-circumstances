package planguard_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/planguard"
)

// THE CENSUS READS WHAT THE SOURCE CAN ISSUE, AND NOTHING ELSE. A constant is inlined where it is
// used and is not a statement where it is declared; a concatenation and a format are one statement
// each, read whole, with a hole where the text is computed; a verb alone is not a statement; a
// test file's statements are the test's.
func TestTheCensusReadsEveryStatementTheSourceSpells(t *testing.T) {
	got, err := planguard.SourceStatements(filepath.Join("testdata", "censusfixture"))
	if err != nil {
		t.Fatal(err)
	}
	var texts []string
	for _, s := range got {
		texts = append(texts, s.String())
	}
	want := []string{
		`SELECT "a" FROM "t" WHERE "b" = ?`,
		`SELECT "a" FROM "t" …`,
		`INSERT INTO … ("x") VALUES (?)`,
		`UPDATE "t" SET "a" = 1 WHERE "b" = ?`,
	}
	slices.Sort(texts)
	slices.Sort(want)
	if !slices.Equal(texts, want) {
		t.Errorf("census =\n  %q\nwant\n  %q", texts, want)
	}
}

// A STATEMENT NO PLAN MATCHES IS REPORTED, which is the census's only purpose: the cases here are
// the ones that could pass a statement that never ran. A plan must BEGIN with the statement's
// text (a helper may append a clause, so it need not end it), and a computed stretch matches any
// text but the constant stretches around it must still appear, in order.
func TestAStatementIsPlannedOnlyByAPlanOfItsOwnText(t *testing.T) {
	got, err := planguard.SourceStatements(filepath.Join("testdata", "censusfixture"))
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]planguard.SourceStatement{}
	for _, s := range got {
		by[s.String()] = s
	}
	for _, tc := range []struct {
		stmt, plan string
		planned    bool
	}{
		{`SELECT "a" FROM "t" WHERE "b" = ?`, `SELECT "a" FROM "t" WHERE "b" = ?`, true},
		{`SELECT "a" FROM "t" WHERE "b" = ?`, `SELECT "a" FROM "t" WHERE "b" = ? ORDER BY "a"`, true},
		{`SELECT "a" FROM "t" WHERE "b" = ?`, `SELECT "a" FROM "t" WHERE "c" = ?`, false},
		{`SELECT "a" FROM "t" WHERE "b" = ?`, `WITH x AS (SELECT "a" FROM "t" WHERE "b" = ?) SELECT * FROM x`, false},
		{`INSERT INTO … ("x") VALUES (?)`, `INSERT INTO "mint" ("x") VALUES (?)`, true},
		{`INSERT INTO … ("x") VALUES (?)`, `INSERT INTO "mint" ("y") VALUES (?)`, false},
		{`UPDATE "t" SET "a" = 1 WHERE "b" = ?`, `UPDATE "t" SET "a" = 1 WHERE "b" = ?`, true},
	} {
		s, ok := by[tc.stmt]
		if !ok {
			t.Fatalf("the census has no %q", tc.stmt)
		}
		if s.PlannedBy([]string{tc.plan}) != tc.planned {
			t.Errorf("%q planned by %q = %v, want %v", tc.stmt, tc.plan, !tc.planned, tc.planned)
		}
	}
}
