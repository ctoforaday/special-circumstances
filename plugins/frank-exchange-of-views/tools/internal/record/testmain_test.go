package record

import (
	"fmt"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/planguard"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — the leak that is invisible on Linux and fails the Windows
// leg. See recordtest.Main (#666).
//
// AND EVERY STATEMENT THE RECORD ISSUES IS PLANNED. The whole suite runs under the query-plan
// guard: each statement this package or recordsql issues is planned, every statement their source
// spells must have been, and the plans and their shape findings are held to testdata/plans.golden
// and testdata/shapes.golden (planguard.Suite).
func TestMain(m *testing.M) {
	suite, err := planguard.StartSuite(planguard.SuiteConfig{
		Dirs:     []string{".", "recordsql"},
		Excluded: censusExcluded,
		Route:    recordsql.UseDriver,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "planguard:", err)
		os.Exit(1)
	}
	recordtest.MainThen(m, suite.Finish)
}

// censusExcluded is each statement the source spells that no run of the guard can plan.
var censusExcluded = map[string]string{
	// The enum vocabularies' rows, written into the schema script. The script runs as one
	// multi-statement Exec beginning CREATE, which the guard does not plan, and an insert of literal
	// rows reads no table, so there is no plan to take.
	`INSERT INTO %q (%s) VALUES (%s);`: "a line of the schema script, not a statement of its own",
	// The arms recordsql generates from EvidenceTables into a view body: each is planned with the
	// view it is part of (view evidence_anchor, view gap), never issued alone.
	`SELECT x." … " AS "anchor_id", ' … ' AS "kind", ' … ' AS "check_kind", x."answers" AS "answers", x."event_id" AS "placed_event" FROM " … " x WHERE x." … " IS NOT NULL`:                                                                                                            "an arm of the evidence_anchor view's body, planned as that view",
	`SELECT x." … " AS "anchor", ce."seat_id" AS "seat", ' … ' AS "check_kind", x."event_id" AS "event_id", ( … ) AS "holds", row_number() OVER (PARTITION BY x." … ", ce."seat_id" ORDER BY x."event_id" DESC) AS "n" FROM " … " x JOIN "events" ce ON ce."id" = x."event_id" WHERE …`: "an arm of the gap view's latest_check, planned as that view",
}
