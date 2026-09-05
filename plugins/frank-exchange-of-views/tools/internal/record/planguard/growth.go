// Package planguard answers one question about the record's SQL: does any statement the code
// runs SCAN a table that GROWS with the run?
//
// # Why a scan is the thing to watch, and why not every scan
//
// `internal/secrets` reported 100% statement coverage while two of its eight patterns could be
// deleted with the suite green — coverage cannot ask whether a test would NOTICE. The same hole
// exists for query plans: every read path is tested, every test passes, and a dropped index
// changes nothing except how long the run takes. Nothing fails, so nothing is reported. The
// record is SQLite, indexed, and the read path was audited clean once (#684 F8); the value here
// is that it STAYS clean, automatically, rather than being re-audited by hand after the next
// schema change.
//
// A blanket "no SCAN anywhere" detector would be noise. SQLite correctly scans the bounded enum
// tables — an eight-row grade table is cheaper read whole than through an index — and those
// scans are right at any run length. What matters is a scan over a table whose row count rises
// with the run, because that is a cost that looks fine on a fixture and is quadratic on a real
// one.
//
// # The growth set is DERIVED, never listed
//
// A hand-kept allowlist of "tables that grow" is the defect this guard exists to catch,
// reproduced one level up: it would silently miss the next table someone adds and report a clean
// board for the run that added it. The set comes from the same descriptors the schema itself is
// generated from — `events`, one table per arm of the Event `body` oneof, and one child table
// per repeated field on those arms. A new event type joins the guarded set on the day its
// message is added, by nobody.
package planguard

import (
	"fmt"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EventsTable is the one growth table not derived from a body descriptor: every event lands in
// it whatever its type.
const EventsTable = "events"

// GrowthTables is every table whose row count rises with the length of a run, derived from the
// proto descriptors the schema is generated from.
//
// It returns an error rather than a partial set for the reason the whole package exists: a guard
// that quietly guards less than it claims is worse than no guard, because the run it stops
// covering still reports a clean board.
func GrowthTables() (map[string]bool, error) {
	bodies, err := recordsql.Bodies()
	if err != nil {
		return nil, fmt.Errorf("planguard: deriving the growth set: %w", err)
	}
	out := map[string]bool{EventsTable: true}
	for _, md := range bodies {
		table := recordsql.TableName(md)
		out[table] = true
		// A REPEATED FIELD IS ITS OWN TABLE, and it grows with the run exactly as its parent
		// does — `supersedes` gets a row per superseded gap, per mint. Missing these would leave
		// the child tables unguarded while the parents looked covered.
		fields := md.Fields()
		for i := 0; i < fields.Len(); i++ {
			f := fields.Get(i)
			if f.IsList() {
				out[table+"_"+string(f.Name())] = true
			}
		}
	}
	if len(out) <= 1 {
		return nil, fmt.Errorf("planguard: derived only %q as a growth table — the descriptors "+
			"answered nothing, and a guard over one table would pass everything else in silence", EventsTable)
	}
	return out, nil
}

// GrowthTableNames is GrowthTables sorted, for a message a human reads.
func GrowthTableNames() ([]string, error) {
	set, err := GrowthTables()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	return names, nil
}
