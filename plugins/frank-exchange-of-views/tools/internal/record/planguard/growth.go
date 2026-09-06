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
	"regexp"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EventsTable is the envelope every event lands in, whatever its type.
const EventsTable = "events"

// createTable finds the tables the schema declares. The DDL is GENERATED — from the proto
// descriptors and from EnvelopeDDL — so reading it back is reading the generator's own output,
// not a second hand-written statement of what exists.
var createTable = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+"?([A-Za-z_][A-Za-z0-9_]*)"?`)

// GrowthTables is every table whose row count rises with the length of a run: EVERY TABLE THE
// SCHEMA DECLARES, MINUS THE BOUNDED ENUM VOCABULARIES.
//
// # Why subtraction, and not a list of what grows
//
// The first version derived the set from the event-body descriptors — `events`, one table per arm
// of the `body` oneof, one child per repeated field. It was right about every table that existed
// when it was written and structurally unable to see the next one, because a table added by hand
// to EnvelopeDDL-style schema text is not a body arm. That is the same defect this guard exists
// to catch, one level up: the guard would have reported a clean board for exactly the run that
// added an unguarded growing table.
//
// AND IT WAS ALREADY WRONG WHEN IT WAS WRITTEN, which is the part worth keeping. Measured against
// this derivation, the descriptor walk missed `motion_direction`, `motion_grade` and
// `motion_petition` — nested message tables that are not arms of the `body` oneof and so were
// never visited. Three growing tables unguarded, on a guard whose whole claim was that it derived
// its own scope.
//
// Subtraction inverts the failure. A NEW TABLE IS GUARDED BY DEFAULT and someone has to argue it
// out, rather than being unguarded by default and needing someone to remember to argue it in. The
// only things taken out are the enum vocabularies, and they are identified by the prefix
// EnumTableName itself builds — checked below against the generator rather than spelled here — so
// the exclusion cannot drift from the naming it depends on.
func GrowthTables() (map[string]bool, error) {
	ddl, err := recordsql.Schema()
	if err != nil {
		return nil, fmt.Errorf("planguard: deriving the growth set: %w", err)
	}
	prefix, err := enumPrefix()
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, m := range createTable.FindAllStringSubmatch(ddl, -1) {
		name := m[1]
		if strings.HasPrefix(name, prefix) {
			continue // a bounded vocabulary: correctly scanned at any run length
		}
		out[name] = true
	}
	if !out[EventsTable] {
		return nil, fmt.Errorf("planguard: the schema declares no %q table — the derivation read "+
			"something other than the record's DDL, and a growth set without the event log "+
			"guards nothing that matters", EventsTable)
	}
	if len(out) <= 1 {
		return nil, fmt.Errorf("planguard: derived only %d growth table(s) — the schema answered "+
			"too little for this to be guarding the record", len(out))
	}
	return out, nil
}

// enumPrefix asks the generator what it names an enum table, rather than repeating "enum_" here.
// A rename there would otherwise silently reclassify every vocabulary as a growth table, turning
// every correct scan of an eight-row table into a reported defect.
func enumPrefix() (string, error) {
	ed := recordpb.File_record_proto.Enums().Get(0)
	name := recordsql.EnumTableName(ed)
	i := strings.Index(name, "_")
	if i <= 0 {
		return "", fmt.Errorf("planguard: cannot read the enum-table prefix out of %q", name)
	}
	return name[:i+1], nil
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
