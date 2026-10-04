package surface

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE QUERY-PLAN GUARD COSTS NOTHING BECAUSE NOTHING SHIPS IT, AND THAT IS NOW A GATE.
//
// recordsql.UseDriver swaps in a driver that runs EXPLAIN QUERY PLAN before each statement.
// Measured: a one-time ~14us (point lookup) to ~35us (full replay) per DISTINCT statement, and
// no measurable per-execution cost after that, because the plan for a statement is memoised —
// see planguard.Recorder.shouldExplain for why memoising is sound here and what would break it.
//
// Small, but it is not zero, and the reason a production run pays none of it is that the override
// is nil unless something calls UseDriver. That is a convention, and a convention is exactly the
// thing that holds until someone reaches for a handy diagnostic in a verb. This makes it a
// property: the ONLY callers are tests.
//
// It fails towards noticing. A new caller in shipped code fails here by file and line, rather
// than quietly adding a second SQL statement to every read a seat performs.
func TestNoShippedCodeInstallsTheQueryPlanGuard(t *testing.T) {
	sources, err := repotree.GoSources("plugins", "frank-exchange-of-views", "tools")
	if err != nil {
		t.Fatalf("locating the tool's sources: %v", err)
	}
	if len(sources) == 0 {
		// AN EMPTY SWEEP IS NOT A CLEAN ONE. repotree refuses an empty set for this reason; the
		// check is here too because a gate that measured nothing must not report agreement.
		t.Fatal("no Go sources found — this gate measured nothing")
	}
	var callers []string
	for _, path := range sources {
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Errorf("reading %s: %v", path, rerr)
			continue
		}
		if strings.Contains(string(b), "UseDriver(") && !strings.HasSuffix(path, "store.go") {
			callers = append(callers, path)
		}
	}
	if len(callers) > 0 {
		t.Errorf("shipped (non-test) code calls recordsql.UseDriver: %s\n\n"+
			"That installs the EXPLAIN QUERY PLAN interceptor, which adds a planning pass per "+
			"distinct statement. It is a test instrument (internal/record/planguard) and a run "+
			"must not pay for it. If a diagnostic really needs plans, it wants its own command, "+
			"not the driver every seat's reads go through.", strings.Join(callers, ", "))
	}
}

// A RECORD HANDLE OUTSIDE THE CACHE IS A TEST INSTRUMENT. recordsql.OpenSeparate opens a second
// handle on a path, which is how a test stands another process's writer beside a reader; in a
// shipped binary it is two connections contending for the write lock in one process, the polling
// herd recordsql.Open's one connection exists to prevent. The ONLY callers are tests.
func TestNoShippedCodeOpensARecordOutsideTheCache(t *testing.T) {
	sources, err := repotree.GoSources("plugins", "frank-exchange-of-views", "tools")
	if err != nil {
		t.Fatalf("locating the tool's sources: %v", err)
	}
	if len(sources) == 0 {
		t.Fatal("no Go sources found — this gate measured nothing")
	}
	var callers []string
	for _, path := range sources {
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			t.Errorf("reading %s: %v", path, rerr)
			continue
		}
		if strings.Contains(string(b), "OpenSeparate(") && !strings.HasSuffix(filepath.ToSlash(path), "recordsql/store.go") {
			callers = append(callers, path)
		}
	}
	if len(callers) > 0 {
		t.Errorf("shipped (non-test) code calls recordsql.OpenSeparate: %s\n\n"+
			"That opens a second handle on a record the process already reaches through recordsql.Open's "+
			"one connection, and the two contend for the write lock on SQLite's fixed busy schedule. "+
			"Read and write through recordsql.Open.", strings.Join(callers, ", "))
	}
}
