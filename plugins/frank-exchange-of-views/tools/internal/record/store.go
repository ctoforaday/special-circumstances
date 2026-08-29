package record

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE RUN DIRECTORY IS THE DATABASE, and this is the one place that says where.
//
// There is no migration story and no version negotiation because a run directory is created by
// `setup` and never outlives the schema that made it. That is the property that makes deriving the
// schema from the descriptors safe: the DDL is rebuilt on open, and a run recorded under different
// descriptors is a different run, in a different directory.
const dbName = "record.db"

// openRun opens the run's record, creating it on first use.
//
// # Why a handle cache and not a handle per call
//
// `MergedEvents` has 28 call sites inside this package and several fire per verb. Opening a
// connection per call would re-apply the derived DDL every time — cheap but not free, and it
// serialises behind the same file lock every other seat is using. The cache is keyed on the
// RESOLVED records directory, so two run paths that resolve to one record share a handle, which is
// the property the second-blackboard incident (#358) turned on.
var ()

func openRun(run Run) (*sql.DB, error) {
	// RESOLVED ALREADY, AND ABSOLUTE ALREADY. Both openers began with RecordsDir followed by
	// filepath.Abs — the exact pair record.Run performs once at construction — so every read and
	// every write repeated it, and the two copies could disagree about the same run.
	abs := run.Records()
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	// THE CACHE MOVED DOWN A LAYER. It was here, keyed on the records directory, and there was no
	// way to release it: recordtest cannot import this package (record's own tests import
	// recordtest), so the fixture that creates run directories could not close their handles.
	// recordsql.Open caches and recordsql.CloseAll releases, and both are reachable from every
	// caller without a cycle.
	return recordsql.Open(filepath.Join(abs, dbName))
}

// openRunForRead is openRun for the read path, where a run that has recorded NOTHING is a legal
// state and an absent database is exactly that.
//
// The distinction is the one RecordsDir already draws and this must not flatten: a resolution
// FAILURE is not an empty run. RecordsDir refuses on the former; this returns (nil, nil) only for
// the latter, and every caller reads a nil handle as an empty record rather than as an error it
// can ignore.
//
// A THIRD STATE EXISTS, and flattening it into the second is how six archived runs came to audit
// clean over nothing. `dbName` absent does not mean "recorded nothing" when the directory is full
// of a FORMER format's shards: the run recorded plenty, this binary just cannot read it. Read as
// an empty run it produced `verify` output where every invariant said [ok] — vacuously true over
// zero events — with a zero board and exit 0, which is the same bytes as a clean board. That is
// the defect shape this package names in recordroot.go and builds the separated-record marker to
// prevent ("turn a lost pointer into an error instead of an empty board"); legacy shards are the
// same fault with a different cause, and get the same answer.
func openRunForRead(run Run) (*sql.DB, error) {
	// RESOLVED ALREADY, AND ABSOLUTE ALREADY. Both openers began with RecordsDir followed by
	// filepath.Abs — the exact pair record.Run performs once at construction — so every read and
	// every write repeated it, and the two copies could disagree about the same run.
	abs := run.Records()
	// A SEPARATED RUN WHOSE ROOT IS GONE IS NOT AN EMPTY RUN, and this check has to be here
	// rather than only at resolution. The handle resolved once; the root can vanish afterwards,
	// and `dashboard --watch` reads through the same handle every fifteen seconds. Without this
	// the Stat below takes its IsNotExist arm and returns the honest zero for a record that
	// exists and is simply unreachable.
	if run.Separated() {
		if _, err := os.Stat(abs); err != nil {
			return nil, fmt.Errorf("record: run %s keeps its record at %s, which is no longer there (%v) — "+
				"the events are gone or the directory moved. This is NOT an empty run and will not be "+
				"reported as one; re-declare the root with %s=<path> if it moved", run.Dir(), abs, err, RecordRootEnv)
		}
	}
	if _, err := os.Stat(filepath.Join(abs, dbName)); os.IsNotExist(err) {
		if shards := legacyShards(abs); len(shards) > 0 {
			return nil, fmt.Errorf("record: %s holds %d event shard(s) of a FORMER record format "+
				"(%s …) and no %s — this run recorded events that THIS BINARY CANNOT READ, which is "+
				"not the same as a run that recorded nothing, and reporting it as an empty board "+
				"would make every invariant pass over zero events. Read it with a binary of the "+
				"event-schema epoch it was written under (runs set up after this landed record it "+
				"as eventSchema in inputs/run-config.json; older ones predate the field, and this "+
				"binary writes epoch %d), or re-run the research under this binary; a run "+
				"directory is created by `setup` and does not outlive the schema that made it",
				abs, len(shards), filepath.Base(shards[0]), dbName, EventSchema)
		}
		return nil, nil
	}
	return openRun(run)
}

// legacyShards names the per-seat JSONL event files the record kept before it became a database.
// It is deliberately a SHAPE check and not a version negotiation: this binary owns exactly one
// format, and everything else is something to refuse by name rather than to parse.
//
// The shape is the FORMER format's own naming rule — events-<seat>-<8 hex nonce>.jsonl — and not
// the looser events-*.jsonl, because that rule is load-bearing in both directions. A file whose
// nonce is absent or not hex was never a shard even then, and TestMergedEventsOnAnEmptyOrAbsentRun
// has pinned since the JSONL era that such a file is ignored rather than parsed. Widening the
// match here would refuse a run over a stray file that never carried an event.
var legacyShardName = regexp.MustCompile(`^events-.+-[0-9a-f]{8}\.jsonl$`)

func legacyShards(recordsDir string) []string {
	entries, err := os.ReadDir(recordsDir)
	if err != nil {
		return nil
	}
	var shards []string
	for _, e := range entries {
		if !e.IsDir() && legacyShardName.MatchString(e.Name()) {
			shards = append(shards, filepath.Join(recordsDir, e.Name()))
		}
	}
	sort.Strings(shards)
	return shards
}
