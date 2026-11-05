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
}

// unindexed is each walk of a growing table that no index can serve, keyed by its shapes.golden
// line, with the reason the index cannot exist. A statement's own walk of a growing table is a
// missing index, and planguard refuses it: the index lands with the statement. An entry here is
// the exception argued in review; one whose finding is gone fails the suite.
var unindexed = map[string]string{}
