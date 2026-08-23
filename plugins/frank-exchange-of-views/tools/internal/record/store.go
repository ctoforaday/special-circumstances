package record

import (
	"database/sql"
	"os"
	"path/filepath"
	"sync"

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
var (
	dbsMu sync.Mutex
	dbs   = map[string]*sql.DB{}
)

func openRun(runDir string) (*sql.DB, error) {
	dir, err := RecordsDir(runDir)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	dbsMu.Lock()
	defer dbsMu.Unlock()
	if db, ok := dbs[abs]; ok {
		return db, nil
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, err
	}
	db, err := recordsql.Open(filepath.Join(abs, dbName))
	if err != nil {
		return nil, err
	}
	dbs[abs] = db
	return db, nil
}

// openRunForRead is openRun for the read path, where a run that has recorded NOTHING is a legal
// state and an absent database is exactly that.
//
// The distinction is the one RecordsDir already draws and this must not flatten: a resolution
// FAILURE is not an empty run. RecordsDir refuses on the former; this returns (nil, nil) only for
// the latter, and every caller reads a nil handle as an empty record rather than as an error it
// can ignore.
func openRunForRead(runDir string) (*sql.DB, error) {
	dir, err := RecordsDir(runDir)
	if err != nil {
		return nil, err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filepath.Join(abs, dbName)); os.IsNotExist(err) {
		return nil, nil
	}
	return openRun(runDir)
}
