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
		Dirs:      []string{".", "recordsql"},
		Excluded:  censusExcluded,
		Unindexed: unindexed,
		Route:     recordsql.UseDriver,
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
	// view it is part of (view evidence_anchor, view gap_refresh), never issued alone.
	`SELECT x." … " AS "anchor_id", ' … ' AS "kind", ' … ' AS "check_kind", x."answers" AS "answers", x."event_id" AS "placed_event" FROM " … " x WHERE x." … " IS NOT NULL`:                                                                "an arm of the evidence_anchor view's body, planned as that view",
	`SELECT x." … " AS "anchor", ce."seat_id" AS "seat", ' … ' AS "check_kind", x."event_id" AS "event_id", ( … ) AS "holds" FROM " … " x JOIN "events" ce ON ce."id" = x."event_id" WHERE x." … " IN (SELECT "anchor" FROM "reach") AND …`: "an arm of gap_refresh's checks, planned as that view",
	// demand_reach's body (recordsql.demandReachSQL), rendered into the view and into gap_refresh:
	// planned as those views.
	`SELECT l."gap_id" AS "gap_id", l."from_gap" AS "from_gap", h."check_kind" AS "check_kind", h."event_id" AS "minted", a."about_ref" AS "anchor", e."seat_id" AS "seat", h."event_id" AS "reached" FROM … JOIN "mint" h ON h."gap_id" = l."gap_id" JOIN "mint" a ON a."gap_id" = l."from_gap" JOIN "events" e ON e."id" = a."event_id" WHERE a."about_kind" = 'anchor'`: "demand_reach's body, planned as view demand_reach and view gap_refresh",
	// The gap-state marker triggers' bodies (recordsql.gapMarkers): each reads NEW, the row its
	// trigger fires on, and runs only inside that trigger, so it has no plan of its own.
	`SELECT "gap_id" FROM "demand_reach" WHERE "anchor" = …`: "a gap-state marker trigger's body, over NEW",
	`SELECT NEW."gap_id"`: "a gap-state marker trigger's body, over NEW",
	`SELECT NEW."value" UNION SELECT "gap_id" FROM "mint" WHERE "event_id" = NEW."event_id"`:                                              "a gap-state marker trigger's body, over NEW",
	`SELECT md."gap_id" FROM "motion" mo JOIN "motion_docket" md ON md."event_id" = mo."event_id" WHERE mo."motion_id" = NEW."motion_id"`: "a gap-state marker trigger's body, over NEW",
	`SELECT NEW."value"`: "a gap-state marker trigger's body, over NEW",
	`SELECT "answers" FROM "evidence_anchor" WHERE "anchor_id" = NEW."value" AND "answers" IS NOT NULL UNION …`: "a gap-state marker trigger's body, over NEW",
	`SELECT NEW."answers" WHERE NEW."answers" IS NOT NULL UNION …`:                                              "a gap-state marker trigger's body, over NEW",
}

// unindexed is each walk of a growing table that no index can serve, keyed by its shapes.golden
// line, with the reason the index cannot exist. A statement's own walk of a growing table is a
// missing index, and planguard refuses it: the index lands with the statement. An entry here is
// the exception argued in review; one whose finding is gone fails the suite.
var unindexed = map[string]string{}
