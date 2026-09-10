package catalogue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math/rand/v2"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// DefaultDir is the store's home: XDG state, NOT ~/.claude. Writing our store into the vendor's
// directory is squatting in a namespace we do not own — the same principle that stops
// ResolveSession globbing ~/.claude/projects/ to guess which transcript is ours.
func DefaultDir() (string, error) {
	if x := os.Getenv("XDG_STATE_HOME"); x != "" {
		return filepath.Join(x, "special-circumstances", "catalogue"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("catalogue: no home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state", "special-circumstances", "catalogue"), nil
}

// writeDSN carries the pragmas in the order that matters. busy_timeout is set BEFORE
// journal_mode(WAL), because converting to WAL takes an exclusive lock and a concurrent opener
// with no timeout yet fails instantly rather than waiting — the defect fixed in #814, on the same
// engine, which this store would otherwise reproduce.
//
// It reaches SQLite only as the query of writeURI's `file:` URI, never appended to a bare path:
// see uriFor for why.
const writeDSN = "?_txlock=immediate" +
	"&_pragma=busy_timeout(5000)" +
	"&_pragma=journal_mode(WAL)" +
	"&_pragma=synchronous(NORMAL)" +
	"&_pragma=foreign_keys(1)"

// readDSN is the query path's own. `_defensive=1` makes `PRAGMA writable_schema=1` a silent
// no-op — measured: it reads back 0 with the flag and 1 without, on the same connection — and
// unlike the ATTACH limit below it holds on every pooled connection, because it is applied at
// connection open.
//
// `mode=ro` IS A URI PARAMETER, AND FOR A LONG TIME IT NEVER ARRIVED. modernc (v1.57.0,
// conn.go newConn) cuts everything from the first `?` off any DSN not prefixed `file:`, opens the
// remainder READWRITE|CREATE, and applies only the `_pragma`s it parsed from the cut-off query.
// Appended to a bare path, `mode=ro` was dropped and `query_only(1)` — which a statement can turn
// off — was the only thing between a read verb and a write; one open and close even checkpointed
// a pending WAL and removed its `-wal`. Through readURI it reaches SQLite, and the connection
// cannot write, checkpoint or roll back a hot journal.
const readDSN = "?mode=ro&_pragma=query_only(1)&_defensive=1&_pragma=busy_timeout(2000)"

// uriFor renders path as the `file:` URI every connection this package opens is given, with
// rawQuery written verbatim (which is what keeps `busy_timeout(5000)`'s parentheses intact).
//
// A BARE PATH NAMES A DIFFERENT FILE. modernc truncates a non-`file:` DSN at its first `?`, so a
// store at `…/a?b/catalogue.db` opened `…/a` — created it, 8192 bytes — and a `#` or `%` in a URI
// path that was not escaped would be read as a fragment or decoded as an escape. So the path is
// escaped by net/url, as frank-exchange-of-views' recordsql.dsnFor does (restated, not imported:
// gray-area is its own module), and for the same reasons: url.URL escapes every reserved
// character, and a path beginning `//` gets an explicit empty authority (`file:////server/share`)
// so it is not parsed as a host.
//
// ONE CORRECTION TO dsnFor: filepath.Abs FIRST. dsnFor prefixes `/` to any path that is not
// rooted, which anchors a RELATIVE path at the filesystem root; its only caller passes an absolute
// path, so it never shows. `telepathy --store` is taken as typed, so here a relative path would
// have opened `/catalogue.db`. After Abs, the only unrooted shape left is a Windows volume path
// (`C:/…`), which SQLite documents as `file:///C:/…` — the leading slash keeps the volume out of
// the authority. On Unix both steps after Abs are no-ops.
func uriFor(path, rawQuery string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("catalogue: resolving %s: %w", path, err)
	}
	p := filepath.ToSlash(abs)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p, RawQuery: rawQuery}
	return u.String(), nil
}

// readURI is the read-only connection: classification, and every read verb.
func readURI(path string) (string, error) {
	return uriFor(path, strings.TrimPrefix(readDSN, "?"))
}

// writeURI is Open's connection, and the only one that can write.
func writeURI(path string) (string, error) {
	return uriFor(path, strings.TrimPrefix(writeDSN, "?"))
}

// Class is what recognise makes of a file.
type Class int

const (
	Fresh   Class = iota // no tables at all: a new or empty file
	Current              // a catalogue at exactly UserVersion
	Older                // a catalogue of an earlier shape, or one left unstamped: Open rebuilds it
	Newer                // a catalogue stamped by a newer binary: refused, never read
	Foreign              // not recognised as a catalogue: never written
)

func (c Class) String() string {
	switch c {
	case Fresh:
		return "fresh"
	case Current:
		return "current"
	case Older:
		return "older"
	case Newer:
		return "newer"
	case Foreign:
		return "foreign"
	}
	return fmt.Sprintf("Class(%d)", int(c))
}

// catalogueTables is every table name a catalogue has ever carried (e111c79a, bd478a05, 28ad35e5
// and today). A file holding any other table is not one of ours.
var catalogueTables = map[string]bool{
	"act": true, "file_offset": true, "meta": true, "provisional_skip": true,
	"session": true, "skip": true, "thought": true, "word": true,
}

// sessionCore is the five `session` columns every shape has carried; #867 renamed only the other
// two. A file whose `session` table lacks any of them is not a catalogue of a known shape.
var sessionCore = []string{"session_id", "project_dir", "cwd", "closed_at", "capture_build"}

// querier is what recognise reads through: the read-only *sql.DB, or Open's *sql.Tx.
type querier interface {
	QueryRow(query string, args ...any) *sql.Row
	Query(query string, args ...any) (*sql.Rows, error)
}

// recognise classifies a file from its stamp, its table names and `session`'s columns. The rules
// run IN THIS ORDER, and the order is the design:
//
//  1. no user tables → Fresh;
//  2. a stamp above UserVersion → Newer if `session` and `file_offset` both exist, judged on
//     table NAMES ALONE (a future shape may add tables or rename `session`'s columns — the
//     recognition contract in schema.go), else Foreign;
//  3. no `session`, or one lacking any of sessionCore → Foreign;
//  4. any table a catalogue has never had → Foreign;
//  5. stamp 0 → Older. STAMP 0 IS NOT PROOF OF A FRESH FILE: before the rebuild, Open ran the
//     schema and the stamp as separate autocommit statements, so a first open killed part-way
//     left a prefix of the tables — `session` always first — unstamped;
//  6. a stamp below UserVersion → Older if `file_offset` exists, else Foreign;
//  7. exactly UserVersion → Current if `file_offset` exists, else Foreign.
//
// The residual it accepts: a foreign file with a `session` table carrying all five columns, only
// catalogue table names and a stamp of 2 or less would be rebuilt. A negative stamp matches no
// rule and is Foreign.
func recognise(q querier) (v int, class Class, tables []string, err error) {
	if err := q.QueryRow(`PRAGMA user_version`).Scan(&v); err != nil {
		return 0, 0, nil, fmt.Errorf("reading user_version: %w", err)
	}
	tables, err = names(q, `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite\_%' ESCAPE '\' ORDER BY name`)
	if err != nil {
		return v, 0, nil, fmt.Errorf("reading the table list: %w", err)
	}
	has := map[string]bool{}
	for _, t := range tables {
		has[t] = true
	}
	switch {
	case len(tables) == 0:
		return v, Fresh, tables, nil
	case v > UserVersion:
		if has["session"] && has["file_offset"] {
			return v, Newer, tables, nil
		}
		return v, Foreign, tables, nil
	case !has["session"]:
		return v, Foreign, tables, nil
	}
	cols, err := names(q, `SELECT name FROM pragma_table_info('session')`)
	if err != nil {
		return v, 0, tables, fmt.Errorf("reading session's columns: %w", err)
	}
	hasCol := map[string]bool{}
	for _, c := range cols {
		hasCol[c] = true
	}
	for _, c := range sessionCore {
		if !hasCol[c] {
			return v, Foreign, tables, nil
		}
	}
	for _, t := range tables {
		if !catalogueTables[t] {
			return v, Foreign, tables, nil
		}
	}
	switch {
	case v == 0:
		return v, Older, tables, nil
	case v > 0 && v < UserVersion:
		if has["file_offset"] {
			return v, Older, tables, nil
		}
	case v == UserVersion:
		if has["file_offset"] {
			return v, Current, tables, nil
		}
	}
	return v, Foreign, tables, nil
}

func names(q querier, query string) ([]string, error) {
	rows, err := q.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// foreignRefusal and newerRefusal are shared by Open and OpenRead, so the two can never describe
// one file in different words.
func foreignRefusal(path string, tables []string) error {
	return fmt.Errorf("catalogue: %s is not a gray-area catalogue (tables: %s) — refusing to touch it; check `--store`",
		path, strings.Join(tables, ", "))
}

// newerRefusal no longer tells anyone to delete the store. The usual cause is version skew — a
// session begun before a plugin update still running the old binary against a store the new one
// has already written — and deleting the store then destroys a healthy catalogue to silence a
// message that a restart resolves on its own.
func newerRefusal(path string, v int) error {
	return fmt.Errorf("catalogue: %s is stamped %d by a newer gray-area (this binary writes %d) — "+
		"usually a session started after a plugin update while this one still runs the older binary; "+
		"it resolves when this session restarts. Do not delete the store.", path, v, UserVersion)
}

// Open opens the store for WRITING, creating it if absent, and rebuilds an older catalogue.
// notice receives the one line that says a rebuild happened; it may be io.Discard.
//
// # Classify read-only first, decide inside the write transaction
//
// A WRITE CONNECTION MUST NEVER MEET A FILE NOBODY RECOGNISED. The write URI's
// `journal_mode(WAL)` is applied at connection open and is PERSISTENT: merely opening a
// mistyped `--store` for writing converts somebody's rollback-journal database to WAL, before a
// single statement of ours has run. So an existing file is first classified on a readURI
// connection, which cannot write, checkpoint or roll back — and which SEES a writer's pending WAL
// frames, so a live store, a store rebuilt but not yet checkpointed, and a foreign file whose
// schema lives only in its `-wal` are all judged by their true state. Foreign and Newer are
// refused there, and no write connection is opened for them.
//
// The one footprint that pre-classification leaves is any SQLite reader's: a WAL-mode file that
// had no sidecars gains an empty `-wal` and a `-shm`. Its database bytes do not change, and a
// rollback-journal file gains nothing.
//
// Only Fresh, Older or Current reach the write URI, and there the answer is taken AGAIN inside a
// `BEGIN IMMEDIATE` transaction — the authoritative one, since a concurrent opener may have
// rebuilt the file in between; that is what makes N concurrent openers of one old catalogue
// produce exactly one rebuild. The schema, the rebuild markers and the stamp commit together or
// not at all, so an interrupted open can no longer leave a prefix of the tables behind.
//
// The window between the two classifications is the only place a foreign file could meet the
// write connection: a catalogue replaced by a foreign file between two opens of one path. The
// in-transaction check still refuses it before any DDL, but the WAL pragma will have run.
func Open(path string, notice io.Writer) (*sql.DB, error) {
	if notice == nil {
		notice = io.Discard
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("catalogue: %w", err)
	}
	// BUSY IS RETRIED HERE, BECAUSE busy_timeout DOES NOT COVER THE CASE THAT HITS.
	//
	// SQLite deliberately does not apply busy_timeout to a lock UPGRADE — two connections each
	// holding a read lock and waiting to write would deadlock — and returns SQLITE_BUSY at once
	// instead (frank-exchange-of-views' recordsql.openUncached documents the same rule). Opening
	// the write URI is such an upgrade on a file not yet in WAL mode: the `journal_mode(WAL)`
	// pragma reads the header under a shared lock, then needs an exclusive one to convert. N
	// independent openers of a fresh store — N hook processes at a session start — do exactly
	// that at once. MEASURED on the Windows CI leg: TestConcurrentOpenersRebuildOnce/fresh failed
	// in 0.17s with `database is locked (5) (SQLITE_BUSY)`, the 5-second busy_timeout never
	// engaged. recordsql escapes it by caching one connection per path in-process; hooks are
	// separate processes, so here the whole classify-and-write phase is retried instead. Once one
	// opener has converted the file, the others' pragma is a no-op and their IMMEDIATE BEGIN
	// queues normally. The budget matches busy_timeout; the backoff is jittered so contending
	// openers do not wake and collide in lockstep.
	deadline := time.Now().Add(openBusyBudget)
	for attempt := 0; ; attempt++ {
		db, from, rebuilt, err := openOnce(path)
		if err == nil {
			if rebuilt {
				fmt.Fprintf(notice, "catalogue: %s was stamped %d, rebuilt empty at shape %d — run `telepathy backfill` to re-read the corpus\n",
					path, from, UserVersion)
			}
			return db, nil
		}
		if !isBusy(err) || !time.Now().Before(deadline) {
			return nil, err
		}
		time.Sleep(busyBackoff(attempt))
	}
}

// openBusyBudget is how long Open keeps retrying SQLITE_BUSY: the same 5s as the write DSN's
// busy_timeout. A variable so a test can shorten it.
var openBusyBudget = 5 * time.Second

// openOnce is one attempt at Open's classify-and-write phase. It closes whatever it opened on
// failure, so a retry starts clean.
func openOnce(path string) (*sql.DB, int, bool, error) {
	switch _, err := os.Stat(path); {
	case err == nil:
		if err := preclassify(path); err != nil {
			return nil, 0, false, err
		}
	case !errors.Is(err, fs.ErrNotExist):
		return nil, 0, false, fmt.Errorf("catalogue: %w", err)
	}
	uri, err := writeURI(path)
	if err != nil {
		return nil, 0, false, err
	}
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, 0, false, fmt.Errorf("catalogue: open %s: %w", path, err)
	}
	from, rebuilt, err := settle(db, path)
	if err != nil {
		db.Close()
		return nil, 0, false, err
	}
	return db, from, rebuilt, nil
}

// isBusy reports whether err is SQLite's SQLITE_BUSY, including its extended forms
// (BUSY_RECOVERY, BUSY_SNAPSHOT, …), whose low byte is the primary code.
func isBusy(err error) bool {
	var se *sqlite.Error
	return errors.As(err, &se) && se.Code()&0xff == sqlite3.SQLITE_BUSY
}

// busyBackoff is 5ms doubling to a 160ms ceiling, plus up to as much again at random.
func busyBackoff(attempt int) time.Duration {
	base := 5 * time.Millisecond << min(attempt, 5)
	return base + rand.N(base)
}

// preclassify is Open's read-only look at an existing file. It returns nil only for a class Open
// may write: Fresh, Older or Current.
func preclassify(path string) error {
	uri, err := readURI(path)
	if err != nil {
		return err
	}
	ro, err := sql.Open("sqlite", uri)
	if err != nil {
		return fmt.Errorf("catalogue: open read-only %s: %w", path, err)
	}
	defer ro.Close()
	v, class, tables, err := recognise(ro)
	if err != nil {
		return fmt.Errorf("catalogue: classifying %s: %w", path, err)
	}
	switch class {
	case Foreign:
		return foreignRefusal(path, tables)
	case Newer:
		return newerRefusal(path, v)
	}
	return nil
}

// settle takes the authoritative classification inside one IMMEDIATE transaction and brings the
// file to UserVersion: rebuilding an older catalogue, applying the schema to a fresh or current
// one. It reports the stamp it rebuilt from.
func settle(db *sql.DB, path string) (from int, rebuilt bool, err error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, false, fmt.Errorf("catalogue: opening %s: %w", path, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	v, class, tables, err := recognise(tx)
	if err != nil {
		return 0, false, fmt.Errorf("catalogue: classifying %s: %w", path, err)
	}
	switch class {
	case Foreign:
		return 0, false, foreignRefusal(path, tables)
	case Newer:
		return 0, false, newerRefusal(path, v)
	case Older:
		if err := dropAll(tx); err != nil {
			return 0, false, fmt.Errorf("catalogue: rebuilding %s: %w", path, err)
		}
	}
	if _, err := tx.Exec(Schema); err != nil {
		return 0, false, fmt.Errorf("catalogue: schema: %w", err)
	}
	if class == Older {
		if err := MarkRebuilt(tx, v, time.Now()); err != nil {
			return 0, false, fmt.Errorf("catalogue: recording the rebuild of %s: %w", path, err)
		}
	}
	if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, UserVersion)); err != nil {
		return 0, false, fmt.Errorf("catalogue: setting user_version: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, false, fmt.Errorf("catalogue: committing %s: %w", path, err)
	}
	return v, class == Older, nil
}

// dropAll removes every view, index and table that is not SQLite's own, views first so none is
// left naming a table that has gone.
func dropAll(tx *sql.Tx) error {
	for _, kind := range []string{"view", "index", "table"} {
		got, err := names(tx, `SELECT name FROM sqlite_master WHERE type = '`+kind+`' AND name NOT LIKE 'sqlite\_%' ESCAPE '\'`)
		if err != nil {
			return err
		}
		sort.Strings(got)
		for _, n := range got {
			if _, err := tx.Exec(`DROP ` + strings.ToUpper(kind) + ` IF EXISTS "` + strings.ReplaceAll(n, `"`, `""`) + `"`); err != nil {
				return fmt.Errorf("dropping %s %s: %w", kind, n, err)
			}
		}
	}
	return nil
}

// OpenRead opens the store for every read verb, READ-ONLY in fact: through readURI, so `mode=ro`
// reaches SQLite and a statement that turns `query_only` off still cannot write.
//
// It CLASSIFIES the file before handing it over, with the same recognise Open uses on the same
// kind of connection, and refuses every class but Current — so it never runs a query against a
// shape it does not write, and its refusal of an older store (which promises a rebuild) is true
// by construction: Open would classify that file Older too.
//
// It returns a *sql.DB whose pool is capped at one connection, and callers must obtain that
// connection through PinnedConn — never db.Query. See PinnedConn for why.
func OpenRead(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("catalogue: no store at %s — nothing has been captured yet, "+
			"which is not the same as a store that is empty: run `telepathy backfill` or let a session run", path)
	}
	uri, err := readURI(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("catalogue: open read-only %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)
	v, class, tables, err := recognise(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("catalogue: classifying %s: %w", path, err)
	}
	switch class {
	case Current:
		return db, nil
	case Older:
		err = fmt.Errorf("catalogue: %s is stamped %d; this binary reads shape %d. The next session start or turn end "+
			"(gray-area's `SessionStart`, `Stop` and `SessionEnd` hooks) rebuilds it, or run `telepathy backfill`.",
			path, v, UserVersion)
	case Newer:
		err = newerRefusal(path, v)
	case Foreign:
		err = foreignRefusal(path, tables)
	default: // Fresh
		err = fmt.Errorf("catalogue: %s holds no catalogue yet — let a session run or run `telepathy backfill`.", path)
	}
	db.Close()
	return nil, err
}

// PinnedConn hands back a connection carrying SQLITE_LIMIT_ATTACHED = 0.
//
// # Why a pinned connection and not the pool
//
// `sqlite.Limit` binds ONE connection, not the *sql.DB. Measured on v1.57.0: after
// Limit(c1, ATTACHED, 0) the first connection refuses ATTACH while a second from the same pool
// reads the limit back as 10 and attaches successfully, as does db.Exec straight through the
// pool. So a query issued through the pool can land on a connection the limit was never applied
// to, and the hole reopens under concurrency while every single-goroutine test still passes.
//
// Writes are already refused on this connection — by `mode=ro`, which reaches SQLite only because
// OpenRead opens a `file:` URI (uriFor: a bare path had the parameter cut off by the driver), and
// by `query_only`. This closes ATTACH, whose only power is reading another SQLite file. On this
// box that is not an escalation — every caller already has Read and Bash — but it becomes one the
// moment the surface is exposed to a caller less privileged than the local user, and it costs one
// call to shut.
func PinnedConn(ctx context.Context, db *sql.DB) (*sql.Conn, error) {
	c, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("catalogue: acquiring connection: %w", err)
	}
	if _, err := sqlite.Limit(c, sqlite3.SQLITE_LIMIT_ATTACHED, 0); err != nil {
		c.Close()
		return nil, fmt.Errorf("catalogue: closing ATTACH on the query connection: %w", err)
	}
	return c, nil
}
