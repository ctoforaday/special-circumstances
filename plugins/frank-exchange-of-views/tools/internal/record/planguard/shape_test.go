package planguard_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/planguard"
)

// tinySchema is a purpose-built schema small enough to read whole: two growing tables, a view
// over one of them, and one table named as an enum vocabulary.
const tinySchema = `
CREATE TABLE a(id INTEGER PRIMARY KEY, k TEXT, v INT);
CREATE TABLE b(id INTEGER PRIMARY KEY, k TEXT, w INT);
CREATE TABLE enum_tiny(value INTEGER PRIMARY KEY, label TEXT);
CREATE VIEW v AS SELECT k, count(*) AS n FROM b GROUP BY k;
`

// shapesOf plans query against tinySchema plus extra and returns its findings and its rendered
// plan, for the failure message.
func shapesOf(t *testing.T, extra, query string) ([]planguard.Shape, string) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1) // one :memory: database, not one per connection
	if _, err := db.Exec(tinySchema + extra); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	views, err := planguard.Views(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	an, err := planguard.NewAnalyzer(views)
	if err != nil {
		t.Fatal(err)
	}
	p, err := planguard.PlanQuery(ctx, db, query)
	if err != nil {
		t.Fatal(err)
	}
	return an.Shapes(query, p), p.Render()
}

func has(shapes []planguard.Shape, kind planguard.ShapeKind, name string) bool {
	for _, s := range shapes {
		if s.Kind == kind && s.Name == name {
			return true
		}
	}
	return false
}

func hasKind(shapes []planguard.Shape, kind planguard.ShapeKind) bool {
	for _, s := range shapes {
		if s.Kind == kind {
			return true
		}
	}
	return false
}

// Each kind is shown on a statement that provokes it and on the same statement once the cause is
// fixed. The fixed half is what keeps a rule from passing by reporting everything.
func TestEachShapeIsCaughtAndItsFixIsNot(t *testing.T) {
	for _, tc := range []struct {
		name       string
		kind       planguard.ShapeKind
		want       string // the finding's name on the broken statement
		query      string
		fix, fixed string // the schema fix, or a rewritten statement, that removes it
	}{
		{
			name:  "a join on an unindexed column builds an automatic index",
			kind:  planguard.AutomaticIndex,
			want:  "b (k=?)",
			query: `SELECT a.id FROM a JOIN b ON b.k = a.k`,
			fix:   `CREATE INDEX b_k ON b(k);`,
		},
		{
			name:  "a correlated max() over an unindexed column walks the table per row",
			kind:  planguard.CorrelatedScan,
			want:  "b [min/max]",
			query: `SELECT (SELECT max(x.w) FROM b x WHERE x.k = a.k) FROM a`,
			fix:   `CREATE INDEX b_k ON b(k);`,
		},
		{
			name:  "a correlated EXISTS over an unindexed column walks the table per row",
			kind:  planguard.CorrelatedScan,
			want:  "b",
			query: `SELECT id FROM a WHERE EXISTS (SELECT 1 FROM b WHERE b.k = a.k ORDER BY w LIMIT 1)`,
			fix:   `CREATE INDEX b_k ON b(k);`,
		},
		{
			name:  "a correlated IN over an unindexed column walks the table per row",
			kind:  planguard.CorrelatedScan,
			want:  "b",
			query: `SELECT id FROM a WHERE a.v IN (SELECT w FROM b WHERE b.k = a.k)`,
			fix:   `CREATE INDEX b_k ON b(k);`,
		},
		{
			name:  "one view evaluated by two subqueries is evaluated twice",
			kind:  planguard.RepeatedSubplan,
			want:  "v",
			query: `SELECT id FROM a WHERE a.k IN (SELECT k FROM v) AND a.v IN (SELECT n FROM v)`,
			fixed: `WITH c AS MATERIALIZED (SELECT k, n FROM v) SELECT id FROM a WHERE a.k IN (SELECT k FROM c) AND a.v IN (SELECT n FROM c)`,
		},
		{
			name:  "a correlated ORDER BY on an unindexed column sorts per row",
			kind:  planguard.CorrelatedSort,
			want:  "FOR ORDER BY over b",
			query: `SELECT (SELECT w FROM b WHERE b.k = a.k ORDER BY w LIMIT 1) FROM a`,
			fix:   `CREATE INDEX b_kw ON b(k, w);`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, plan := shapesOf(t, "", tc.query)
			if !has(got, tc.kind, tc.want) {
				t.Errorf("not reported as %s %q\n  sql:  %s\n  plan:\n%s  found: %v", tc.kind, tc.want, tc.query, plan, got)
			}
			fixedQuery := tc.query
			if tc.fixed != "" {
				fixedQuery = tc.fixed
			}
			got, plan = shapesOf(t, tc.fix, fixedQuery)
			if hasKind(got, tc.kind) {
				t.Errorf("still reported as %s after the fix\n  fix:  %s\n  sql:  %s\n  plan:\n%s  found: %v", tc.kind, tc.fix, fixedQuery, plan, got)
			}
		})
	}
}

// A BOUNDED VOCABULARY IS CORRECTLY SCANNED, EVEN PER ROW. Reporting it would teach the reader that
// a correlated-scan finding can be ignored.
func TestACorrelatedScanOfAnEnumVocabularyIsNotReported(t *testing.T) {
	const q = `SELECT (SELECT max(t.value) FROM enum_tiny t WHERE t.label = a.k) FROM a`
	got, plan := shapesOf(t, "", q)
	if hasKind(got, planguard.CorrelatedScan) {
		t.Errorf("a scan of an enum vocabulary was reported\n  plan:\n%s  found: %v", plan, got)
	}
	// The same statement over a growing table is: the exemption is the table, not the shape.
	got, plan = shapesOf(t, "", `SELECT (SELECT max(t.w) FROM b t WHERE t.k = a.k) FROM a`)
	if !has(got, planguard.CorrelatedScan, "b [min/max]") {
		t.Errorf("the same shape over a growing table was not reported\n  plan:\n%s  found: %v", plan, got)
	}
}

// A MATERIALIZE IS COMPUTED ONCE, so a scan inside one is not a per-row scan even when the
// materialization sits under a correlated subquery. The scan of the materialized rows is.
func TestAScanInsideAMaterializationRunsOnce(t *testing.T) {
	const q = `WITH c AS MATERIALIZED (SELECT k, count(*) AS n FROM b GROUP BY k)
		SELECT (SELECT max(c.n) FROM c WHERE c.n > a.v) FROM a`
	got, plan := shapesOf(t, "", q)
	if has(got, planguard.CorrelatedScan, "b") || has(got, planguard.CorrelatedScan, "b [min/max]") {
		t.Errorf("the scan of b inside the materialization was reported as per-row\n  plan:\n%s  found: %v", plan, got)
	}
	if !has(got, planguard.CorrelatedScan, "c [min/max]") {
		t.Errorf("the per-row scan of the materialized CTE was not reported\n  plan:\n%s  found: %v", plan, got)
	}
}

// AN ALIAS IS READ FROM ITS FROM ITEM, NOT FROM TEXT THAT LOOKS LIKE ONE. A comment and a string
// literal that spell a binding must not create one.
func TestScopeBindsAliasesOnlyFromFromItems(t *testing.T) {
	s := planguard.NewScope(`SELECT 'FROM b p' /* FROM v p */ FROM "a" AS p -- FROM b p
		JOIN (SELECT 1) q ON 1, b r`, nil)
	for word, want := range map[string]string{"p": "a", "q": "q", "r": "b", "zz": "zz"} {
		if got := s.Name(word); got != want {
			t.Errorf("Name(%q) = %q, want %q", word, got, want)
		}
	}
	// A view the statement names is read too, so an alias declared inside it resolves.
	s = planguard.NewScope(`SELECT * FROM "outer_v"`, []planguard.View{{Name: "outer_v", SQL: `CREATE VIEW outer_v AS SELECT * FROM b inner_alias`}})
	if got := s.Name("inner_alias"); got != "b" {
		t.Errorf("an alias declared in a named view resolved to %q, want b", got)
	}
}

// THE GOLDEN MUST NOT CHURN ON SUBQUERY NUMBERING, which shifts with any upstream edit.
func TestNormalizeDetailDropsSubqueryNumbering(t *testing.T) {
	for in, want := range map[string]string{
		"CORRELATED SCALAR SUBQUERY 17": "CORRELATED SCALAR SUBQUERY",
		"LIST SUBQUERY 3":               "LIST SUBQUERY",
		"SCAN (subquery-52)":            "SCAN (subquery)",
		"CO-ROUTINE (subquery-8)":       "CO-ROUTINE (subquery)",
		"SCAN (join-4)":                 "SCAN (join)",
		"SEARCH ss USING AUTOMATIC COVERING INDEX (id=?) LEFT-JOIN": "SEARCH ss USING AUTOMATIC COVERING INDEX (id=?) LEFT-JOIN",
	} {
		if got := planguard.NormalizeDetail(in); got != want {
			t.Errorf("NormalizeDetail(%q) = %q, want %q", in, got, want)
		}
	}
}

// A PLAN THAT CANNOT BE ASSEMBLED IS REFUSED, not re-rooted: a node moved to the top loses the
// correlated ancestor every per-row finding is read from.
func TestBuildPlanRefusesAnOrphanAndAnEmptyPlan(t *testing.T) {
	if _, err := planguard.BuildPlan([]planguard.PlanRow{{ID: 2, Parent: 0, Detail: "SCAN a"}, {ID: 5, Parent: 9, Detail: "SCAN b"}}); err == nil {
		t.Error("a row naming a missing parent was accepted")
	}
	if _, err := planguard.BuildPlan(nil); err == nil {
		t.Error("an empty plan was accepted")
	}
}

// A MIN/MAX OVER A VIEW IS JUDGED BY WHAT THE VIEW COSTS, NOT BY THE ONE ROW IT RETURNS.
//
// `events` is named here because constrained-scan is about the record's growing tables and reads
// that set from the real schema. Three forms of `max(id)` over a view, under a constraint:
//
//   - a view SQLite flattens into its table is the bare "SEARCH e" min/max loop over `events`,
//     which walks back from the end until a row matches — the sittings shape;
//   - a view SQLite co-routines, built from a covering-index walk, shows "SEARCH v" over the
//     subplan and no heap scan at all: every row of the walk is produced, and none of it is a
//     scan heapScan counts, so only the min/max loop over the subplan can carry the finding;
//   - a co-routined view over a bounded vocabulary, or of rowid lookups, is not a finding: the
//     exemption is what the subplan reads, not the shape.
func TestAMinMaxOverAViewIsJudgedByTheView(t *testing.T) {
	const events = `
CREATE TABLE events(id INTEGER PRIMARY KEY, type TEXT, seat TEXT, sitting INT);
CREATE INDEX events_type ON events(type);
CREATE VIEW flat AS SELECT e.id AS id, e.seat AS seat FROM events e WHERE e.sitting = e.id;
CREATE VIEW per_type AS SELECT type, max(id) AS id FROM events GROUP BY type;
CREATE VIEW per_label AS SELECT label, max(value) AS id FROM enum_tiny GROUP BY label;
CREATE VIEW two AS SELECT id FROM events WHERE id = 5 UNION ALL SELECT id FROM events WHERE id = 6;
`
	for _, tc := range []struct {
		name, query, want string
	}{
		{"a flattened view walks its table", `SELECT o.id FROM events o WHERE o.id = (SELECT max(id) FROM flat WHERE seat = 'x')`, "events [min/max]"},
		{"a co-routined view built from an index walk", `SELECT o.id FROM events o WHERE o.id = (SELECT max(id) FROM per_type)`, "per_type [min/max]"},
		{"a co-routined view over a vocabulary", `SELECT a.id FROM a WHERE a.v = (SELECT max(id) FROM per_label)`, ""},
		{"a co-routined view of rowid lookups", `SELECT a.id FROM a WHERE a.v = (SELECT max(id) FROM two)`, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, plan := shapesOf(t, events, tc.query)
			if tc.want == "" {
				if hasKind(got, planguard.ConstrainedScan) {
					t.Errorf("reported as %s\n  sql:  %s\n  plan:\n%s  found: %v", planguard.ConstrainedScan, tc.query, plan, got)
				}
				return
			}
			if !has(got, planguard.ConstrainedScan, tc.want) {
				t.Errorf("not reported as %s %q\n  sql:  %s\n  plan:\n%s  found: %v", planguard.ConstrainedScan, tc.want, tc.query, plan, got)
			}
		})
	}
}
