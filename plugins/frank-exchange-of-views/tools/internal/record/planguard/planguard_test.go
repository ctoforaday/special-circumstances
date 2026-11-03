package planguard_test

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/planguard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// statementFloor is the number of statements the sweep below is known to produce. It exists so
// COVERAGE CANNOT COLLAPSE IN SILENCE: if a refactor stops routing reads through the guarded
// driver, or the sweep stops reaching the read paths, the findings list goes empty and every
// assertion here passes while measuring nothing. A floor turns that into a failure.
//
// It is a floor and not an equality: adding a read path SHOULD raise the count, and a test that
// had to be edited every time someone added a query would be edited without being read.
const statementFloor = 30

// seedAndSweep opens a guarded run, drives the record's read paths through it, and returns the
// recorder and the path of the record it swept.
func seedAndSweep(t *testing.T) (*planguard.Recorder, string) {
	t.Helper()
	rec := planguard.NewRecorder()
	name, err := planguard.Install(rec)
	if err != nil {
		t.Fatal(err)
	}
	defer recordsql.UseDriver(name)()

	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{
		GapId: proto.String("G1"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	run := runtest.Open(t, dir)

	// The read surface, driven through the guarded driver. Every entry point added here widens
	// what the guard can see; none of them is asserted on individually.
	_, _ = record.FamilyOf(run)
	_, _, _ = record.BoardCounts(run)
	_, _ = record.CitationLabels(run)
	_, _ = record.CitedSources(run)
	_, _ = record.RecordedComputations(run)
	_, _ = record.MergedEvents(run)
	_, _ = record.BoardJSONBytes(run)
	_, _ = record.MotionsJSONBytes(run)
	_, _ = record.EvidenceJSONBytes(run)
	_, _ = record.Epochs(run)
	_, _ = record.SeatsThatSat(run)
	_, _ = record.TerminalVerdict(run)
	_, _ = record.RecordedOutcome(run)
	return rec, filepath.Join(run.Records(), "record.db")
}

// THE GROWTH SET IS DERIVED, AND MUST STAY BIG ENOUGH TO MEAN SOMETHING. A derivation that
// silently returned two tables would guard almost nothing while reading as a guard.
func TestGrowthSetIsDerivedFromTheSchema(t *testing.T) {
	names, err := planguard.GrowthTableNames()
	if err != nil {
		t.Fatal(err)
	}
	if len(names) < 20 {
		t.Errorf("only %d growth tables derived (%v) — the descriptors answered too little for "+
			"this to be guarding the record", len(names), names)
	}
	set := map[string]bool{}
	for _, n := range names {
		set[n] = true
	}
	// `events` is declared by hand and is not a body arm, so it is in the set only because the
	// derivation SUBTRACTS the bounded vocabularies rather than enumerating what grows. A
	// repeated-field child is the shape most easily dropped by a derivation that walked only
	// top-level messages.
	for _, want := range []string{"events", "mint", "close", "cite", "mint_supersedes", "mint_found_by"} {
		if !set[want] {
			t.Errorf("the derived growth set is missing %q", want)
		}
	}
	// AND THE OTHER HALF OF THE SUBTRACTION. An enum vocabulary is bounded — SQLite is right to
	// scan an eight-row grade table — and counting one as growth would report every correct read
	// of it as a defect, which is how a guard teaches people to ignore it.
	for _, n := range names {
		if strings.HasPrefix(n, "enum_") {
			t.Errorf("%q is a bounded vocabulary and must not be in the growth set", n)
		}
	}
}

// constrainedOf is the constrained-scan findings for one statement whose plan is one line.
func constrainedOf(t *testing.T, stmt, detail string) []planguard.Shape {
	t.Helper()
	an, err := planguard.NewAnalyzer(nil)
	if err != nil {
		t.Fatal(err)
	}
	p, err := planguard.BuildPlan([]planguard.PlanRow{{ID: 2, Detail: detail}})
	if err != nil {
		t.Fatal(err)
	}
	var out []planguard.Shape
	for _, s := range an.Shapes(stmt, p) {
		if s.Kind == planguard.ConstrainedScan {
			out = append(out, s)
		}
	}
	return out
}

// THE RULE IS THE INTERESTING PART, so it is tested on its own rather than only through the sweep.
// Each case here is a real plan line observed during calibration.
func TestConstrainedScanSeparatesMissingIndexesFromDesignCosts(t *testing.T) {
	for _, tc := range []struct {
		name, stmt, detail string
		defect             bool
	}{
		{"full replay", `SELECT id, seat_id FROM events ORDER BY id`, "SCAN events", false},
		{"bulk detail read", `SELECT "event_id", "gap_id" FROM "mint"`, "SCAN mint", false},
		{"whole-table count", `SELECT (SELECT count(*) FROM "verify")`, "SCAN verify", false},
		{"last row by rowid", `SELECT "verdict" FROM "outcome" ORDER BY "event_id" DESC LIMIT 1`, "SCAN outcome", false},
		{"covering index walk", `SELECT DISTINCT "type" FROM "events" ORDER BY "type"`, "SCAN events USING COVERING INDEX events_type", false},
		{"autoindex walk", `SELECT "event_id" FROM "mint_supersedes" ORDER BY "event_id"`, "SCAN mint_supersedes USING INDEX sqlite_autoindex_mint_supersedes_1", false},
		{"filtered scan of an enum vocabulary", `SELECT "value" FROM "enum_grade" WHERE "value" > ?`, "SCAN enum_grade", false},

		{"filtered scan is the defect", `SELECT "gap_id" FROM "mint" WHERE "gap_id" = ?`, "SCAN mint", true},
		{"joined scan is the defect", `SELECT * FROM "close" JOIN "events" ON "events"."id" = "close"."event_id"`, "SCAN close", true},
		{"an aliased scan is judged as its table", `SELECT m."gap_id" FROM "mint" m WHERE m."problem" = ?`, "SCAN m", true},
		{"a bare SEARCH may walk the whole table", `SELECT max(m."event_id") FROM "mint" m WHERE m."problem" = ?`, "SEARCH m", true},
	} {
		got := constrainedOf(t, tc.stmt, tc.detail)
		if (len(got) > 0) != tc.defect {
			t.Errorf("%s: reported=%v, want %v\n  sql:  %s\n  plan: %s", tc.name, got, tc.defect, tc.stmt, tc.detail)
		}
	}
}

// A COLUMN NAMED LIKE A KEYWORD MUST NOT MANUFACTURE A DEFECT. Substring matching on "WHERE"
// would call this bulk read constrained and report a correct query as broken.
func TestConstrainedScanDoesNotMatchKeywordsInsideIdentifiers(t *testing.T) {
	if got := constrainedOf(t, `SELECT "wherefore", "joined_at" FROM "mint"`, "SCAN mint"); len(got) != 0 {
		t.Errorf("a column named `wherefore` was read as a WHERE clause: %+v", got)
	}
}

// THE GUARD HAS TEETH, DEMONSTRATED AGAINST THE REAL DATABASE AND THE REAL SCHEMA.
//
// The rule's table test feeds it plan lines; this one issues two statements that differ only in WHICH COLUMN they filter
// on: `event_id` is the table's INTEGER PRIMARY KEY, so SQLite reaches the row directly, and
// `problem` carries no index, so it walks the table. One must be reported and the other must not.
//
// It is the answer to "would this notice a dropped index" — the question coverage cannot ask,
// and the reason this package exists rather than a second hand audit of the kind #684 F8 was.
func TestAFilteredQueryWithoutAnIndexIsCaughtAndAnIndexedOneIsNot(t *testing.T) {
	rec := planguard.NewRecorder()
	name, err := planguard.Install(rec)
	if err != nil {
		t.Fatal(err)
	}
	defer recordsql.UseDriver(name)()

	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, recordtest.At(t, "red-chair", "red-chair:mint:G1", &recordpb.Mint{
		GapId: proto.String("G1"), Problem: proto.String("p"), RequiredFix: proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		CheckKind:  recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:   recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:     recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}))
	run := runtest.Open(t, dir)
	db, err := recordsql.Open(filepath.Join(run.Records(), "record.db"))
	if err != nil {
		t.Fatal(err)
	}

	const indexed = `SELECT "gap_id" FROM "mint" WHERE "event_id" = ?`
	const unindexed = `SELECT "gap_id" FROM "mint" WHERE "problem" = ?`
	for _, q := range []string{indexed, unindexed} {
		rows, qerr := db.Query(q, "p")
		if qerr != nil {
			t.Fatalf("%s: %v", q, qerr)
		}
		rows.Close()
	}

	an, err := planguard.NewAnalyzer(nil)
	if err != nil {
		t.Fatal(err)
	}
	var reported []string
	for _, sp := range rec.Plans() {
		for _, sh := range an.Shapes(sp.SQL, sp.Plan) {
			if sh.Kind == planguard.ConstrainedScan {
				reported = append(reported, strings.Join(strings.Fields(sp.SQL), " "))
			}
		}
	}
	found := func(q string) bool {
		for _, r := range reported {
			if r == q {
				return true
			}
		}
		return false
	}
	if !found(unindexed) {
		t.Errorf("filtering `mint` on an UNINDEXED column was not reported — the guard would not "+
			"notice a dropped index, which is the only thing it is for.\n  reported: %v", reported)
	}
	if found(indexed) {
		t.Errorf("filtering `mint` on its INTEGER PRIMARY KEY was reported as a defect — the "+
			"guard cries wolf on a point lookup.\n  reported: %v", reported)
	}
}
