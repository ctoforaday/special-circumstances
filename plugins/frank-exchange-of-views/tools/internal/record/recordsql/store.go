package recordsql

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	// THE DRIVER BELONGS TO THE PACKAGE THAT OPENS THE DATABASE, not to its tests.
	//
	// It lived in schema_test.go, so database/sql had a registered "sqlite" driver throughout the
	// suite and none in the shipped binary. Every test passed and the first real `chair register`
	// failed with `unknown driver "sqlite"`. A blank import is invisible to the compiler's unused
	// check, which is exactly why the wrong file stayed good enough.
	_ "modernc.org/sqlite"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// Open creates or opens a run's database and applies the derived schema to a new one.
//
// The pragmas are not defaults and each is load-bearing:
//
//   - foreign_keys is OFF by default in SQLite, for backwards compatibility with databases written
//     before it existed. Left off, every foreign key in this schema is decoration — including the
//     one that makes a ruling-before-its-filing unwritable, which is the reason the record is here
//     at all.
//   - journal_mode=WAL lets the readers a run has — the projections, the dashboard watcher — read
//     while a seat writes, which the shard files achieved by giving every seat its own file and
//     paid for with an ordering problem.
//   - busy_timeout keeps a concurrent seat waiting rather than failing. Seats are dispatched in
//     parallel and a lock contended for a few milliseconds is not an error to report to a human.
//
// dsnFor builds the DSN, and it uses net/url rather than escaping the path by hand.
//
// WE ARE THE ONES BUILDING A URI. `_txlock=immediate` is a DRIVER setting — it changes how BEGIN
// is issued and cannot be an Exec'd pragma — so the query-string form is not optional, and having
// chosen a URI, escaping the path into it is this function's obligation. It was `"file:" + path`,
// which is a filesystem path spliced into a grammar that reserves `?` and `#`: a run directory
// containing either was TRUNCATED there, silently, to a path that still opens. Two runs then share
// one database, or a run opens one that is not its own, and every write reports success.
//
// A HAND-ROLLED ESCAPE GOT THE THREE RESERVED CHARACTERS AND STOPPED THERE. url.URL escapes the
// same three and everything else the grammar needs, in the right order, and it also handles the
// one case percent-encoding cannot: a path beginning `//` would parse as an AUTHORITY, and
// url.URL emits an explicit empty authority (`file:////server/share`) so the path survives whole.
// RawQuery is written verbatim, which is what keeps `busy_timeout(5000)`'s parentheses intact.
func dsnFor(path string) string {
	// A WINDOWS PATH IS NOT A URI PATH, AND url.URL CANNOT KNOW THAT. `C:\\Users\\x` has no
	// leading slash, so String() writes `file://` and then the path — putting `C:` where the
	// AUTHORITY goes. SQLite accepts an empty authority or `localhost` and nothing else, so every
	// open on Windows failed with "invalid uri authority" and the store was unwritable: 60 of 60
	// fuzz runs, zero events of every type, round 0.
	//
	// SQLite documents the Windows form as `file:///C:/Users/x` — forward slashes, and the volume
	// behind a third slash so the authority stays empty. Both steps are needed: ToSlash alone still
	// leaves `C:/Users/x` unrooted, and a leading slash alone leaves backslashes that escape to
	// %5C and take the whole path into the authority with them.
	//
	// On Unix both are no-ops: `/runs/x` is already rooted and already slashed.
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{
		Scheme: "file",
		Path:   p,
		// BUSY_TIMEOUT COMES FIRST, AND THE ORDER IS THE POINT. Pragmas apply left to right, and
		// converting a fresh database to WAL takes an EXCLUSIVE lock — so with the timeout set
		// after it, the conversion is the one operation that runs with no timeout in force. A
		// second opener arriving in that window got SQLITE_BUSY immediately instead of waiting:
		// measured on the Windows CI leg, TestConcurrentOpenOnFreshDatabase failing in 0.04s with
		// "database is locked" (#801), where a live 5-second timeout would have waited and then
		// succeeded. The production shape is the fuzz's concurrent lanes (#630), several seat
		// processes creating the schema at once on round 0; `setup` creating the database in
		// advance is the accident that hides it the rest of the time.
		RawQuery: "_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)",
	}
	return u.String()
}

// open handles are cached per absolute path, and CloseAll is the release the cache never had.
//
// WHY THERE WAS NO CLOSE. A seat runs one command per process, so the handle dies with the process
// and the OS reclaims it; the dashboard is the only long-lived reader and it wants the handle held.
// Nothing in production ever needed to close one.
//
// TESTS ARE THE OTHER SHAPE: one process, hundreds of run directories, every handle retained. On
// Linux that is invisible — unlinking an open file succeeds, so `t.TempDir` cleanup passes. On
// WINDOWS an open file cannot be removed and the cleanup fails, which is how this surfaced:
// `TempDir RemoveAll cleanup: unlinkat ...\records\record.db` across ten packages at once. The
// leak was always real; only one of the two platforms was willing to say so.
//
// THE CACHE LIVES HERE RATHER THAN IN record, and that is what makes it releasable. It was one
// layer up, where the fixture package cannot reach it: recordtest cannot import record, because
// record's own tests import recordtest. Owned by the thing that opens the database, both the
// caching and the closing are available to every caller without a cycle.
var (
	openMu    sync.Mutex
	openCache = map[string]*sql.DB{}
)

// Close releases the handle for ONE database, which is what a test fixture wants.
//
// CloseAll IS THE WRONG TOOL FROM A PER-TEST CLEANUP, measured: a subtest that closes everything
// takes its parent's and its siblings' handles with it, and the next act on a run that is still
// live fails with "sql: database is closed". A fixture releases the run it created; nothing else
// is its to release.
func Close(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	openMu.Lock()
	defer openMu.Unlock()
	db, ok := openCache[abs]
	if !ok {
		return nil
	}
	delete(openCache, abs)
	return db.Close()
}

// CloseUnder releases every cached handle whose database lives under dir.
//
// THE PATH CANNOT BE GUESSED FROM THE RUN, which is what the per-path release assumed. A SEPARATED
// run keeps its records outside the run directory entirely (RecordsDir resolves it), so
// `<run>/records/record.db` finds nothing and the real handle stays open — measured on Windows as
// `TestASeparatedRunKeepsNoEventsUnderTheRun` failing its own TempDir cleanup.
//
// A prefix is the honest key for a test: everything this test created lives under its temp
// directory, wherever the resolution put it, and nothing else does.
func CloseUnder(dir string) error {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	prefix := abs + string(filepath.Separator)
	openMu.Lock()
	defer openMu.Unlock()
	var first error
	for k, db := range openCache {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		if err := db.Close(); err != nil && first == nil {
			first = err
		}
		delete(openCache, k)
	}
	return first
}

// OrphanedHandles returns cached databases whose FILE IS GONE — the exact shape of the leak
// recordtest.TmpRun exists to prevent, measured rather than pattern-matched.
//
// WHY THE MISSING FILE IS THE SIGNAL. A test that opens a record under `t.TempDir()` and never
// releases the handle leaves this cache holding a path the test framework then removes. On Linux
// the removal SUCCEEDS — an open file can be unlinked — so the test passes and nothing says a
// word; on Windows the same removal fails and the test is refused. One platform's silence is the
// whole reason the mistake has been made nine times.
//
// A cached path with no file behind it can only mean the directory was removed while this cache
// still held the handle. That is not a heuristic about how the test was WRITTEN — it is the leak
// itself, so there is no pattern to evade and no exemption list to keep.
//
// Reported paths name the test that leaked them: `t.TempDir` builds its directory out of the
// test's own name, so the path IS the attribution. A directory the test made itself
// (`os.MkdirTemp`) gives only its prefix, which is where both of the first two catches came from.
//
// IT REPORTS NOTHING ON WINDOWS, AND THAT IS NOT A HOLE. This is the LINUX-SIDE instrument for a
// defect Windows already refuses natively: there the removal fails and takes the test with it, so
// there is no orphan to find. The point is to fail on the platform the author is actually running,
// rather than a push and a CI round-trip later.
func OrphanedHandles() []string {
	openMu.Lock()
	defer openMu.Unlock()
	var out []string
	for k := range openCache {
		if _, err := os.Stat(k); errors.Is(err, fs.ErrNotExist) {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// Seal folds the write-ahead log back into the database file and empties it, so the record is ONE
// file that carries the whole run.
//
// THE RECORD RUNS journal_mode=WAL AND NOTHING EVER CHECKPOINTS IT. SQLite checkpoints on the last
// connection close; a seat runs one command per process, so the handle dies with the process and
// the log simply grows. A finished run is therefore a 766 KB database beside a 4.1 MB log, and
// run-archive ships both.
//
// WHAT THAT COSTS IS A WRONG ANSWER, NOT A BIG DIRECTORY. A reader that opens the database without
// the log — anyone who copies record.db out of the archive, or opens it `immutable=1`, which tells
// SQLite the file cannot change and to skip the log — gets a SMALLER, INTERNALLY CONSISTENT,
// ENTIRELY PLAUSIBLE run. Measured on the 2026-09-20 run, same file, same moment: 186 events
// against 301, four gaps against five, zero verdicts against one, and NO OUTCOME ROW against
// `verified`. An agent read the first column, concluded that runs do not record whether they
// passed, surveyed five more runs the same way and filed it as a release blocker. The run had
// passed and the record said so.
//
// The truncated read and the honest read are the same bytes, which is why this is sealed at the
// source rather than documented as a caveat for every reader to remember.
func Seal(path string) error {
	db, err := Open(path)
	if err != nil {
		return err
	}
	// TRUNCATE rather than PASSIVE or FULL: PASSIVE gives up rather than wait for a reader, and
	// FULL leaves the log file at its grown size with the frames still in it. TRUNCATE is the only
	// mode that leaves a zero-length log, which is what makes the result CHECKABLE by size alone —
	// see SealedSize, which is how a caller proves the seal took rather than trusting a nil error.
	var busy, logFrames, checkpointed int
	if err := db.QueryRow(`PRAGMA wal_checkpoint(TRUNCATE)`).Scan(&busy, &logFrames, &checkpointed); err != nil {
		return fmt.Errorf("sealing %s: %w", path, err)
	}
	// busy=1 means a reader held the log open and the checkpoint did NOT complete. Returning nil
	// here would be the same silent-partial this function exists to remove.
	if busy != 0 {
		return fmt.Errorf("sealing %s: a concurrent reader held the write-ahead log open, so it was not folded in "+
			"(%d frames left). Seal a run that has ENDED; a live run's log is meant to be open", path, logFrames)
	}
	return nil
}

// SealedSize reports the bytes still sitting in a record's write-ahead log. Zero means sealed.
//
// The caller asks the FILESYSTEM rather than the database, because that is the question a later
// reader's failure actually turns on: not "did a checkpoint run" but "is there a second file here
// that holds events". A missing log is sealed by definition.
func SealedSize(path string) (int64, error) {
	st, err := os.Stat(path + "-wal")
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return st.Size(), nil
}

// CloseAll releases every cached handle. For a process shutting down, not for a test cleanup.
func CloseAll() error {
	openMu.Lock()
	defer openMu.Unlock()
	var first error
	for k, db := range openCache {
		if err := db.Close(); err != nil && first == nil {
			first = err
		}
		delete(openCache, k)
	}
	return first
}

// driverNameOverride is empty in every build and every process EXCEPT the query-plan guard,
// which registers a driver that runs EXPLAIN QUERY PLAN over each statement before executing it
// (internal/record/planguard). The guard cannot ask its question any other way: the record's
// reads are 59 inline statements across seventeen files, so there is nothing to enumerate and
// the plan has to be intercepted where the statement is issued.
//
// A VARIABLE AND NOT A BUILD TAG, because the guard drives the SAME code production runs — a
// tagged build would guard a different program from the one that ships.
var driverNameOverride atomic.Pointer[string]

func driverName() string {
	if s := driverNameOverride.Load(); s != nil {
		return *s
	}
	return "sqlite"
}

// UseDriver points subsequent Open calls at a different database/sql driver and returns the
// function that puts it back. FOR THE QUERY-PLAN GUARD, and nothing else.
//
// A SHIPPED RUN PAYS NOTHING FOR THE GUARD, because the override is nil unless this is called,
// and TestNoShippedCodeInstallsTheQueryPlanGuard fails if non-test code ever calls it. The cost
// it keeps out is small but real: measured at a one-time ~14us on a point lookup and ~35us on a
// full replay per DISTINCT statement, with no measurable per-execution cost after that (the plan
// is memoised — planguard.Recorder.shouldExplain says why that is sound and what would break
// it). Before memoisation it was 2.74x on a point lookup, which is what a run would have paid on
// every read.
//
// It does NOT reach handles already open: Open caches one *sql.DB per path, and a database
// opened before this call keeps the driver it was opened with. A caller wanting guarded reads
// must use a database it opens after installing, which every guard run does by working in a
// fresh directory.
func UseDriver(name string) (restore func()) {
	driverNameOverride.Store(&name)
	return func() { driverNameOverride.Store(nil) }
}

// Open returns the cached handle for this database, opening it on first use.
func Open(path string) (*sql.DB, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path // an unresolvable path still deserves ONE handle rather than none
	}
	openMu.Lock()
	defer openMu.Unlock()
	if db, ok := openCache[abs]; ok {
		return db, nil
	}
	db, err := openUncached(path)
	if err != nil {
		return nil, err
	}
	openCache[abs] = db
	return db, nil
}

func openUncached(path string) (*sql.DB, error) {
	// _txlock=immediate IS NOT A TUNING KNOB. IT IS THE DIFFERENCE BETWEEN WORKING AND NOT.
	//
	// A default `BEGIN` is DEFERRED: it takes no lock, acquires a READ lock at the first SELECT, and
	// tries to UPGRADE to a write lock at the first INSERT. The write path counts the seat's existing
	// events and then inserts, so it is exactly that shape — and SQLite deliberately does NOT apply
	// busy_timeout to an upgrade, because two readers both waiting to upgrade would deadlock. It
	// returns SQLITE_BUSY immediately instead.
	//
	// MEASURED, through the real binary: 8 concurrent seat processes writing 5 events each lost
	// roughly half of them to `database is locked (5) (SQLITE_BUSY)`. Not a slow path — a REFUSED
	// act, with the seat told its record failed. The shard layout had no contention by construction
	// (one file per seat), so this is a hazard the storage change INTRODUCED and had to answer.
	//
	// `immediate` takes the write lock at BEGIN, before any read. busy_timeout then applies, because
	// a writer waiting for a writer is a queue rather than a deadlock. It also fixes the correctness
	// half: a deferred transaction's count could be read from a snapshot another writer has already
	// moved past, which is what SQLITE_BUSY_SNAPSHOT exists to refuse.
	db, err := sql.Open(driverName(), dsnFor(path))
	if err != nil {
		return nil, err
	}
	// ONE CONNECTION. A pool is the wrong shape for a single-writer file database.
	//
	// `_txlock=immediate` takes the write lock at BEGIN, so two goroutines holding two pooled
	// connections collide in SQLITE_BUSY and then poll — on SQLite's default busy schedule, which
	// is a FIXED sequence ({1,2,5,10,15,20,25,25,25,50,50,100}ms, 100ms thereafter) with NO
	// jitter, so contending waiters wake together and retry together. With one connection they
	// queue on a Go mutex instead: no lock contention to resolve, no polling, no herd.
	//
	// It does NOT replace busy_timeout, which is still what handles the case this cannot touch —
	// a DIFFERENT PROCESS holding the write lock. Seats are separate processes; that is the
	// contention the 8-process test exists for.
	//
	// SAFE HERE because nothing queries the pool while a transaction from it is open: `Append`
	// validates (which reads the board) BEFORE insertNumbered begins its transaction, and
	// deriveKey counts through the tx rather than around it. With one connection, a db.Query
	// inside an open tx would deadlock on itself rather than merely contend — so that ordering is
	// now load-bearing, not incidental.
	db.SetMaxOpenConns(1)

	existed := hasEvents(db)
	if err := ensureSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("recordsql: preparing %s: %w", path, err)
	}
	// A DATABASE THIS BINARY DID NOT CREATE is held to this binary's columns before anything reads
	// it. Its tables and views are the creating binary's (ensureSchema applies once), so a view
	// read naming a column added since fails with SQLite's "no such column", which names neither
	// the cause nor the way out. Refused here by CONTENT — what the database lacks — never by a
	// recorded version.
	if existed {
		if err := requireDeclaredSchema(db); err != nil {
			db.Close()
			return nil, err
		}
	}
	return db, nil
}

// declared is this binary's own table/column set, built once per process from the schema applied
// to a scratch in-memory database — the schema itself, not a second copy of it.
var declared struct {
	once   sync.Once
	tables []string
	cols   map[string][]string
	views  []string
	err    error
}

func declaredSchema() ([]string, map[string][]string, error) {
	declared.once.Do(func() {
		schema, err := Schema()
		if err != nil {
			declared.err = err
			return
		}
		mem, err := sql.Open(driverName(), ":memory:")
		if err != nil {
			declared.err = err
			return
		}
		defer mem.Close()
		mem.SetMaxOpenConns(1) // one connection is one in-memory database
		if _, err := mem.Exec(schema); err != nil {
			declared.err = err
			return
		}
		// THE VIEWS GO IN TOO, because they are half of what a reader names. A projection that is a
		// query over a view fails with SQLite's "no such table: change" on a record written before
		// that view existed — the same unnamed failure requireDeclaredSchema exists to replace for a
		// column, in a record whose EVENTS are all present and readable.
		if _, err := mem.Exec(ViewsDDL); err != nil {
			declared.err = err
			return
		}
		rows, err := mem.Query(`SELECT "name" FROM sqlite_master WHERE type = 'table' AND "name" NOT LIKE 'sqlite_%' ORDER BY "name"`)
		if err != nil {
			declared.err = err
			return
		}
		var tables []string
		for rows.Next() {
			var n string
			if err := rows.Scan(&n); err != nil {
				rows.Close()
				declared.err = err
				return
			}
			tables = append(tables, n)
		}
		rows.Close()
		cols := map[string][]string{}
		for _, t := range tables {
			have, err := columnsOf(mem, t)
			if err != nil {
				declared.err = err
				return
			}
			for c := range have {
				cols[t] = append(cols[t], c)
			}
			sort.Strings(cols[t])
		}
		vrows, err := mem.Query(`SELECT "name" FROM sqlite_master WHERE type = 'view' ORDER BY "name"`)
		if err != nil {
			declared.err = err
			return
		}
		var views []string
		for vrows.Next() {
			var n string
			if err := vrows.Scan(&n); err != nil {
				vrows.Close()
				declared.err = err
				return
			}
			views = append(views, n)
		}
		vrows.Close()
		declared.tables, declared.cols, declared.views = tables, cols, views
	})
	return declared.tables, declared.cols, declared.err
}

// declaredViews is every view this binary's ViewsDDL creates, sorted. It shares declaredSchema's
// once, so asking for it applies the same one-time in-memory schema.
func declaredViews() ([]string, error) {
	if _, _, err := declaredSchema(); err != nil {
		return nil, err
	}
	return declared.views, nil
}

// requireDeclaredSchema refuses a database missing a table, a column or a VIEW this binary declares,
// naming each and the way out (`migrate`). Missing columns are named first: they are what a run
// written before a field was added lacks, and the seat needs to read which field it was.
//
// VIEWS ARE CHECKED FOR THE SAME REASON THE COLUMNS ARE, and were not. A run's views are fixed when
// its database is created — ensureSchema applies once, and deliberately does not take a write lock on
// every open — so a binary that added a view reads an older record and gets "no such table", naming
// neither the cause nor the way out. That failure is worse than the column one, because a view is
// DERIVED: every event the projection needs is present, and the only thing missing is the query.
func requireDeclaredSchema(db *sql.DB) error {
	tables, cols, err := declaredSchema()
	if err != nil {
		return fmt.Errorf("recordsql: reading this binary's own schema: %w", err)
	}
	var lacks, absent []string
	for _, t := range tables {
		have, err := columnsOf(db, t)
		if err != nil {
			return err
		}
		if len(have) == 0 {
			absent = append(absent, fmt.Sprintf("%q", t))
			continue
		}
		var missing []string
		for _, c := range cols[t] {
			if !have[c] {
				missing = append(missing, fmt.Sprintf("%q", c))
			}
		}
		if len(missing) > 0 {
			lacks = append(lacks, fmt.Sprintf("this run's %q table has no %s column", t, strings.Join(missing, ", ")))
		}
	}
	if len(absent) > 0 {
		lacks = append(lacks, "this run's record has no "+strings.Join(absent, ", ")+" table")
	}
	views, err := declaredViews()
	if err != nil {
		return fmt.Errorf("recordsql: reading this binary's own views: %w", err)
	}
	var noView []string
	for _, v := range views {
		has, err := hasView(db, v)
		if err != nil {
			return err
		}
		if !has {
			noView = append(noView, fmt.Sprintf("%q", v))
		}
	}
	if len(noView) > 0 {
		lacks = append(lacks, "this run's record has no "+strings.Join(noView, ", ")+" view")
	}
	if len(lacks) == 0 {
		return nil
	}
	return fmt.Errorf("recordsql: %s — %s", strings.Join(lacks, "; "), olderRunAdvice)
}

// ensureSchema applies the schema if this database does not have it, DECIDING INSIDE THE
// TRANSACTION THAT WOULD CREATE IT (#557).
//
// It used to read whether `events` existed and then apply, with nothing between. Two openers of a
// FRESH database both saw zero and both applied; the second failed with "table … already exists".
// Nothing ever hit it, for two reasons that are accidents rather than guarantees: the in-process
// handle cache means one process cannot race itself, and the cross-process test creates the
// schema before spawning its children. Production was covered by the same accident — `setup`
// makes the run directory, and its database, before dispatching a seat.
//
// The fuzz is the harness that removes the accident. It builds its run directory directly rather
// than through `setup`, so the first seat command is what creates the schema — and once the lanes
// run concurrently (#630 phase 3), round 0 opens a fresh database from three processes at once.
//
// WHY THIS AND NOT `CREATE TABLE IF NOT EXISTS`: the two differ on a HALF-CREATED schema, and the
// DDL is ~171 statements applied in one transaction precisely so that a crash mid-apply leaves
// nothing rather than a partial record. IF NOT EXISTS would make each statement independently
// survivable and turn a torn apply into a database that looks complete. Deciding inside the
// transaction keeps the all-or-nothing property and closes the race with it.
//
// The DSN carries _txlock=immediate, so Begin takes the WRITE lock before any read: a second
// opener queues on busy_timeout rather than racing, and re-reads after the first commits. A
// writer waiting for a writer is a queue; that is the same property the seat write path relies on.
func ensureSchema(db *sql.DB) error {
	// The fast path stays lock-free: an existing database is the overwhelmingly common case, and
	// taking a write lock on every open would put every reader behind the writer queue.
	if hasEvents(db) {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after Commit; the whole point on the loser's path
	// RE-READ UNDER THE LOCK. This is the half that makes it a check-then-create no longer: the
	// loser of the race arrives here after the winner committed and sees the table it would have
	// created.
	var n int
	if err := tx.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='events'`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return applySchemaTx(tx)
}

// hasEvents reports whether the record's table is already there. A read error answers "no" and
// lets the transactional path produce the real diagnosis rather than two spellings of one fault.
func hasEvents(db *sql.DB) bool {
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name='events'`).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

// applySchema creates the whole record in ONE transaction.
//
// IT WAS 171 STATEMENTS, EACH ITS OWN IMPLICIT TRANSACTION. SQLite autocommits any statement not
// already inside one, and every commit is an fsync — so creating a run directory paid 171 disk
// syncs to write a schema that is DERIVED and could be regenerated for free. Measured here:
// 499ms per fresh database, against 39ms for the same DDL in one transaction. Thirteen times, and
// the cost is paid by every test that opens a run (208 of them in internal/cli alone) as well as
// by every real run.
//
// NOTHING IS TRADED FOR IT. The obvious alternatives all weaken durability — `synchronous=off`
// measured 46ms and risks corruption, `synchronous=normal` 67ms and can drop recent commits — and
// this is FASTER than either while leaving `synchronous` at its default. One fsync instead of
// 171 is not a relaxed guarantee; it is the same guarantee, asked for once.
//
// SQLite's own forum states the mechanism: an autocommitted statement fsyncs on its own, so
// batching replaces one-fsync-per-statement with one fsync at COMMIT.
func applySchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // no-op after a successful Commit
	return applySchemaTx(tx)
}

// applySchemaTx is the apply itself, on a transaction the CALLER owns — because the existence
// check and the apply have to be one atomic act (#557), and a function that opens its own
// transaction cannot be part of someone else's decision.
func applySchemaTx(tx *sql.Tx) error {
	schema, err := Schema()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(schema); err != nil {
		return err
	}
	if _, err := tx.Exec(ViewsDDL); err != nil {
		return err
	}
	return tx.Commit()
}

// Insert writes one event and its body inside a single transaction.
//
// ATOMIC BECAUSE A BODY WITHOUT ITS ENVELOPE IS NOT AN EVENT. The shard format could not express a
// half-written act — a torn line was simply dropped, and `ReadShard` counted it as an anomaly — so
// the failure mode moved rather than disappeared: an event with a row in `events` and none in its
// body table would replay as an act with no content, which reads as a seat that did nothing.
func Insert(db *sql.DB, ev *recordpb.Event) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // the commit below is what matters; a rollback after it is a no-op
	// THE ID COMES BACK FROM InsertTx, NOT FROM last_insert_rowid().
	//
	// The first split of this function asked SQLite for the last inserted rowid after InsertTx
	// returned — which is the id of the LAST row written, and InsertTx writes the envelope, then
	// the body, then a row per repeated value. For a mint with two `supersedes` entries it handed
	// back the id of a `mint_supersedes` row, and every caller then looked up an event that was
	// not there.
	id, err := InsertTx(tx, ev)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// sittingOf mints this event's id and answers which sitting it belongs to, in the transaction that
// is about to insert it.
//
// THE ID IS MINTED RATHER THAN TAKEN FROM last_insert_rowid() BECAUSE AN OPENING NAMES ITSELF. An
// event that opens a sitting carries its OWN id in sitting_id, and the events table is append-only
// by trigger, so there is no second statement in which to fill the column in. Minting inside the
// writing transaction is safe by the same property the idempotency ordinal already relies on: the
// transaction holds the write lock from its BEGIN, so no other writer can take the id between the
// max() and the insert.
//
// A NON-OPENING JOINS THE SEAT'S LATEST SITTING, or none. "Latest" is by id, which is append order;
// a seat that has never opened one — the harness's own bookkeeping, a verb run before `register` —
// gets NULL, and every reader that counts sittings then counts none rather than inventing one.
func sittingOf(tx *sql.Tx, ev *recordpb.Event) (id int64, sitting any, err error) {
	if err := tx.QueryRow(`SELECT COALESCE(max("id"), 0) + 1 FROM "events"`).Scan(&id); err != nil {
		return 0, nil, fmt.Errorf("recordsql: minting the event's id: %w", err)
	}
	if seat, opens := recordpb.SeatOpeningSitting(ev); opens {
		// A REGISTER UNDER AN AGENT THE HARNESS ALREADY BRACKETED JOINS THAT SITTING RATHER THAN
		// OPENING A SECOND ONE. Both events are openings on their own terms — either can be the only
		// one a dispatch gets — and in the live shape a dispatch has BOTH. The agent id is the join,
		// and it is the harness's id for one subagent invocation, so a pair cannot span two
		// dispatches; a resume carries no bracket and finds nothing here.
		if agent := recordpb.AgentOpening(ev); agent != "" {
			var bracket sql.NullInt64
			if err := tx.QueryRow(
				`SELECT max(e."id") FROM "events" e JOIN "sitting_open" o ON o."event_id" = e."id"
				  WHERE o."agent_id" = ? AND o."seat_id" = ? AND e."sitting_id" = e."id"`,
				agent, seat).Scan(&bracket); err != nil {
				return 0, nil, olderSchema(tx, "sitting_open",
					fmt.Errorf("recordsql: asking whether %s was already bracketed into a sitting: %w", agent, err))
			}
			if bracket.Valid {
				return id, bracket.Int64, nil
			}
		}
		return id, id, nil
	}
	var open sql.NullInt64
	if err := tx.QueryRow(
		`SELECT max("id") FROM "sittings" WHERE "seat_id" = ?`, ev.GetSeatId()).Scan(&open); err != nil {
		return 0, nil, olderSchema(tx, "sittings",
			fmt.Errorf("recordsql: asking which sitting of %s this act joins: %w", ev.GetSeatId(), err))
	}
	if !open.Valid {
		return id, nil, nil
	}
	return id, open.Int64, nil
}

// InsertTx writes one event inside a transaction the CALLER owns.
//
// The write path needs that: an event's seq and its idempotency ordinal are counted from the rows
// already present, and a count in one transaction followed by an insert in another is a
// read-then-write with a gap in it. Here the count, the derivation and the insert are one atomic
// unit, which is a guarantee the shard layout got by giving each seat its own file and lost the
// moment two processes shared one.
func InsertTx(tx *sql.Tx, ev *recordpb.Event) (int64, error) {
	id, sitting, err := sittingOf(tx, ev)
	if err != nil {
		return 0, err
	}
	if _, err := tx.Exec(
		`INSERT INTO events (id, seat_id, ts, type, key, sitting_id) VALUES (?, ?, ?, ?, ?, ?)`,
		id, ev.GetSeatId(), ev.GetTs(),
		recordpb.Word(ev.GetType()), nullable(ev.GetKey()), sitting,
	); err != nil {
		return 0, olderRun(tx, "events", []string{`"sitting_id"`},
			fmt.Errorf("recordsql: recording the event: %w", err))
	}

	body, ok := recordpb.Body(ev)
	if !ok {
		return 0, fmt.Errorf("recordsql: event %d carries no body — every EventType names a body message", id)
	}
	return id, insertBody(tx, id, body.ProtoReflect())
}

// insertBody writes one message into its table, and recurses for the message arms of a oneof.
//
// It walks the SAME descriptors the DDL was derived from, so a field the schema has and the writer
// skips is not possible: both loops are the same loop over the same objects in the same process.
// A generated mapper would be a second derivation of one source, committed at a different moment
// from the schema, and the two could disagree.
func insertBody(tx *sql.Tx, eventID int64, m protoreflect.Message) error {
	md := m.Descriptor()
	table := TableName(md)

	cols := []string{`"event_id"`}
	vals := []any{eventID}
	var lists []protoreflect.FieldDescriptor
	var arms []protoreflect.FieldDescriptor

	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if fd.IsList() {
			if m.Has(fd) {
				lists = append(lists, fd)
			}
			continue
		}
		if fd.Message() != nil {
			if m.Has(fd) {
				arms = append(arms, fd)
			}
			continue
		}
		if !m.Has(fd) {
			// ABSENT AND ZERO ARE DIFFERENT and the column is nullable for that reason. Writing a
			// zero here would record a grade the seat never gave, or a verdict for a run that
			// reached none.
			continue
		}
		cols = append(cols, fmt.Sprintf("%q", fd.Name()))
		vals = append(vals, value(fd, m.Get(fd)))
	}

	// The oneof's discriminator, so a reader knows which arm to look for without probing every
	// child table.
	for _, fd := range arms {
		if od := fd.ContainingOneof(); od != nil && !od.IsSynthetic() {
			cols = append(cols, fmt.Sprintf("%q", string(od.Name())+"_case"))
			vals = append(vals, string(fd.Name()))
		}
	}

	stmt := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table,
		strings.Join(cols, ", "), strings.TrimSuffix(strings.Repeat("?, ", len(vals)), ", "))
	if _, err := tx.Exec(stmt, vals...); err != nil {
		return olderRun(tx, table, cols, fmt.Errorf("recordsql: recording a %s: %w", table, err))
	}

	for _, fd := range lists {
		l := m.Get(fd).List()
		lt := table + "_" + string(fd.Name())
		for i := 0; i < l.Len(); i++ {
			if _, err := tx.Exec(
				fmt.Sprintf("INSERT INTO %q (\"event_id\", \"ord\", \"value\") VALUES (?, ?, ?)", lt),
				eventID, i, listValue(fd, l.Get(i)),
			); err != nil {
				return olderSchema(tx, lt, fmt.Errorf("recordsql: recording %s.%s: %w", table, fd.Name(), err))
			}
		}
	}
	for _, fd := range arms {
		if err := insertArm(tx, eventID, table+"_"+string(fd.Name()), m.Get(fd).Message()); err != nil {
			return err
		}
	}
	return nil
}

// listValue is one list element in its column's type — the inverse of protoValue.
func listValue(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return v.Int()
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.BytesKind:
		return v.Bytes()
	case protoreflect.EnumKind:
		if vd := fd.Enum().Values().ByNumber(v.Enum()); vd != nil {
			return recordpb.Spelling(vd)
		}
	}
	return v.String()
}

func insertArm(tx *sql.Tx, eventID int64, table string, m protoreflect.Message) error {
	md := m.Descriptor()
	cols := []string{`"event_id"`}
	vals := []any{eventID}
	for i := 0; i < md.Fields().Len(); i++ {
		fd := md.Fields().Get(i)
		if !m.Has(fd) || fd.IsList() || fd.Message() != nil {
			continue
		}
		cols = append(cols, fmt.Sprintf("%q", fd.Name()))
		vals = append(vals, value(fd, m.Get(fd)))
	}
	stmt := fmt.Sprintf("INSERT INTO %q (%s) VALUES (%s)", table,
		strings.Join(cols, ", "), strings.TrimSuffix(strings.Repeat("?, ", len(vals)), ", "))
	if _, err := tx.Exec(stmt, vals...); err != nil {
		return olderRun(tx, table, cols, fmt.Errorf("recordsql: recording a %s: %w", table, err))
	}
	return nil
}

// value renders one field for SQL. An enum goes in as its WORD, which is what the vocabulary table
// holds and what a human reading the record with a plain SELECT expects to see.
func value(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.EnumKind:
		return recordpb.Word(enumAt{fd.Enum(), v.Enum()})
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind, protoreflect.Int64Kind:
		return v.Int()
	case protoreflect.BytesKind:
		return v.Bytes()
	}
	return v.String()
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type enumAt struct {
	ed protoreflect.EnumDescriptor
	n  protoreflect.EnumNumber
}

func (e enumAt) Descriptor() protoreflect.EnumDescriptor { return e.ed }
func (e enumAt) Type() protoreflect.EnumType             { return nil }
func (e enumAt) Number() protoreflect.EnumNumber         { return e.n }

var _ = proto.Marshal
