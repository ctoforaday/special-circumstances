package catalogue

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func newStore(t *testing.T) (string, *sql.DB) {
	t.Helper()
	p := filepath.Join(t.TempDir(), "catalogue.db")
	db, err := Open(p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return p, db
}

// THE VIEWS ARE THE CONTRACT. Agents write their own SQL against these, so a renamed column
// breaks work outside this repository that no sweep can reach. This is the check that makes the
// rename fail HERE — #819's lesson, where 0 of 7 archives became unreadable because words were
// retired.
func TestViewColumnsAreTheContract(t *testing.T) {
	_, db := newStore(t)
	for view, want := range ViewColumns {
		rows, err := db.Query(`SELECT * FROM ` + view + ` LIMIT 0`)
		if err != nil {
			t.Errorf("%s: %v", view, err)
			continue
		}
		got, err := rows.Columns()
		rows.Close()
		if err != nil {
			t.Errorf("%s: %v", view, err)
			continue
		}
		if len(got) != len(want) {
			t.Errorf("%s columns = %v, want %v", view, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s columns = %v, want %v", view, got, want)
				break
			}
		}
	}
	// And the roster itself is pinned: a NEW view that nobody declared is as much a contract
	// change as a renamed column.
	rows, _ := db.Query(`SELECT name FROM sqlite_master WHERE type='view' ORDER BY name`)
	var names []string
	for rows.Next() {
		var n string
		rows.Scan(&n)
		names = append(names, n)
	}
	rows.Close()
	var want []string
	for k := range ViewColumns {
		want = append(want, k)
	}
	sort.Strings(want)
	if len(names) != len(want) {
		t.Errorf("views on the contract = %v, declared = %v", names, want)
	}
}

// Writes are refused on the query path, by `mode=ro` — which reaches SQLite only because OpenRead
// opens readURI's `file:` URI; appended to a bare path the driver cut it off — and by
// query_only(1). query_only is turned OFF first here, so what refuses below is `mode=ro` alone:
// before readURI, this exact sequence wrote.
func TestWritesAreRefusedOnTheQueryPath(t *testing.T) {
	p, db := newStore(t)
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('S','/p',1,1)`)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if _, err := ro.Exec(`PRAGMA query_only=0`); err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('X','/p',1,1)`,
		`UPDATE session SET project_dir='zzz'`,
		`DELETE FROM session`,
		`DROP TABLE session`,
		`CREATE TABLE evil(x)`,
	} {
		if _, err := ro.Exec(q); err == nil {
			t.Errorf("read-only connection ALLOWED %q", q)
		}
	}
	var n int
	if err := ro.QueryRow(`SELECT count(*) FROM v_session`).Scan(&n); err != nil || n != 1 {
		t.Errorf("a plain SELECT must still work: n=%d err=%v", n, err)
	}
}

// ATTACH is closed ON THE CONNECTION THE VERB USES, and this is the case that regresses
// silently: sqlite.Limit binds ONE connection, so a query issued through the pool can land on a
// connection the limit was never applied to. A test that only checks the pinned connection
// passes against an implementation that then queries through the pool.
func TestAttachIsRefusedOnThePinnedConnection(t *testing.T) {
	p, db := newStore(t)
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('S','/p',1,1)`)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ctx := context.Background()
	c, err := PinnedConn(ctx, ro)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.ExecContext(ctx, `ATTACH DATABASE '`+p+`' AS other`); err == nil {
		t.Error("ATTACH was allowed on the pinned connection — the limit did not bind")
	}
	// And an ordinary read still works through it.
	var n int
	if err := c.QueryRowContext(ctx, `SELECT count(*) FROM v_session`).Scan(&n); err != nil || n != 1 {
		t.Errorf("SELECT through the pinned connection: n=%d err=%v", n, err)
	}
}

// PRAGMA writable_schema=1 is a silent no-op under _defensive=1 — and unlike the ATTACH limit,
// this one holds pool-wide because it is applied at connection open.
func TestWritableSchemaIsANoOp(t *testing.T) {
	p, _ := newStore(t)
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ro.Exec(`PRAGMA writable_schema=1`)
	var ws int
	if err := ro.QueryRow(`PRAGMA writable_schema`).Scan(&ws); err != nil {
		t.Fatal(err)
	}
	if ws != 0 {
		t.Errorf("writable_schema reads %d after being set; _defensive=1 did not hold", ws)
	}
}

// A runaway query is cancellable, so a bad SELECT cannot wedge the box.
func TestRunawayQueryIsCancelled(t *testing.T) {
	p, db := newStore(t)
	tx, _ := db.Begin()
	st, _ := tx.Prepare(`INSERT INTO act(session_id,seq,ts,tool,outcome) VALUES('S',?,0,'Bash','ok')`)
	for i := 0; i < 20000; i++ {
		st.Exec(i)
	}
	tx.Commit()
	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := ro.QueryContext(ctx, `SELECT count(*) FROM act a, act b, act c`); err == nil {
		t.Error("a three-way cross join was not cancelled")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("cancellation took %v — the deadline is not being honoured", d)
	}
}

// A store written by a NEWER binary is refused, not read as this shape. The store is derived, so
// refusing costs a reprojection; guessing costs a store that answers wrongly while looking well.
func TestANewerShapeIsRefused(t *testing.T) {
	p, db := newStore(t)
	if _, err := db.Exec(`PRAGMA user_version = 99`); err != nil {
		t.Fatal(err)
	}
	db.Close()
	_, err := Open(p, io.Discard)
	if err == nil {
		t.Fatal("a store from a newer shape was opened rather than refused")
	}
	// The skew text, and never the old advice: the usual cause is a session still running the
	// previous binary, and deleting the store then destroys a healthy catalogue.
	if !strings.Contains(err.Error(), "is stamped 99 by a newer gray-area") || strings.Contains(err.Error(), "delete it") {
		t.Errorf("newer-shape refusal = %q", err)
	}
}

// An absent store is an ERROR with a reason, never an empty result. "No rows" and "no store"
// are the same answer to a caller that cannot tell them apart.
func TestAnAbsentStoreIsRefusedNotEmpty(t *testing.T) {
	_, err := OpenRead(filepath.Join(t.TempDir(), "nothing.db"))
	if err == nil {
		t.Fatal("an absent store opened successfully; a caller would read zero rows as a clean board")
	}
}

// ---- fixtures -------------------------------------------------------------------------------

// rawOpen is the FIXTURE writer: a plain read-write connection with no pragmas, so a fixture's
// journal mode is whatever the fixture sets. It goes through uriFor like every other connection,
// so a temp dir carrying `#` or `%` from a test name still names the right file.
func rawOpen(t testing.TB, p string) *sql.DB {
	t.Helper()
	u, err := uriFor(p, "")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", u)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1) // journal_mode and wal_autocheckpoint are per connection
	return db
}

func mustExec(t testing.TB, db *sql.DB, stmts ...string) {
	t.Helper()
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
}

// shape1 is the DDL a catalogue carried before #867, kept once in telecli's testdata.
func shape1(t testing.TB) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "telecli", "testdata", "shape1.sql"))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// shape1Store is a catalogue as a pre-#867 binary left it — WAL, as every real one is — stamped
// `stamp`: 1 as that binary wrote it, 2 as a post-#867 binary re-stamped it (this box's state),
// 0 as an interrupted first open left it.
func shape1Store(t testing.TB, dir string, stamp int) string {
	t.Helper()
	p := filepath.Join(dir, "catalogue.db")
	db := rawOpen(t, p)
	mustExec(t, db, `PRAGMA journal_mode=WAL`, shape1(t),
		`INSERT INTO session(session_id,project_dir,first_seen,last_seen) VALUES('OLD','/p',1,1)`,
		fmt.Sprintf(`PRAGMA user_version = %d`, stamp))
	db.Close()
	return p
}

// fixture builds a file from raw statements in a fresh dir.
func fixture(t testing.TB, stmts ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "catalogue.db")
	db := rawOpen(t, p)
	mustExec(t, db, stmts...)
	db.Close()
	return p
}

// currentAt is a store Open created, closed, then re-stamped v.
func currentAt(t testing.TB, v int) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "catalogue.db")
	db, err := Open(p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
	raw := rawOpen(t, p)
	mustExec(t, raw, fmt.Sprintf(`PRAGMA user_version = %d`, v))
	raw.Close()
	return p
}

// classify runs recognise on a readURI connection, as Open's pre-classification does.
func classify(t testing.TB, p string) (int, Class, []string) {
	t.Helper()
	u, err := readURI(p)
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", u)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	v, c, tables, err := recognise(db)
	if err != nil {
		t.Fatalf("recognise(%s): %v", p, err)
	}
	return v, c, tables
}

// The shape-1 `session` table on its own: what a first open killed after one statement left.
const shape1Session = `CREATE TABLE session (session_id TEXT PRIMARY KEY, project_dir TEXT NOT NULL,
    cwd TEXT NOT NULL DEFAULT '', first_seen INTEGER NOT NULL, last_seen INTEGER NOT NULL,
    closed_at INTEGER, capture_build TEXT NOT NULL DEFAULT '')`

// ---- URIs -----------------------------------------------------------------------------------

// EVERY CONNECTION OPENS THE PATH IT WAS GIVEN, so every input goes through BOTH constructors and
// each is held to its own exact query. The path is compared DECODED (url.Parse(got).Path), so a
// working directory carrying a space, `#` or `%` still compares equal; the escaped form is
// asserted separately, because an unescaped `%41` would decode to `A` and still parse.
func TestURIsNameThePathGiven(t *testing.T) {
	const wantRead = "mode=ro&_pragma=query_only(1)&_defensive=1&_pragma=busy_timeout(2000)"
	const wantWrite = "_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(1)"
	rooted := func(p string) string {
		s := filepath.ToSlash(p)
		if !strings.HasPrefix(s, "/") {
			s = "/" + s
		}
		return s
	}
	base := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	type row struct {
		name, in, path   string // path: what url.Parse(got).Path must be
		prefix, contains string
	}
	rows := []row{
		// NEVER file:///rel.db — that is the root of the filesystem.
		{name: "relative", in: "rel.db", path: rooted(filepath.Join(cwd, "rel.db"))},
		{name: "percent", in: filepath.Join(base, "a%41b", "c.db"), path: rooted(filepath.Join(base, "a%41b", "c.db")), contains: "/a%2541b/c.db"},
		{name: "hash", in: filepath.Join(base, "a#b", "c.db"), path: rooted(filepath.Join(base, "a#b", "c.db")), contains: "/a%23b/c.db"},
	}
	if runtime.GOOS == "windows" {
		rows = append(rows,
			row{name: "volume", in: `C:\Users\x.db`, path: "/C:/Users/x.db", prefix: "file:///C:/Users/x.db?"},
			row{name: "unc", in: `\\server\share\x.db`, path: "//server/share/x.db", prefix: "file:////server/share/x.db?"},
		)
	} else {
		rows = append(rows,
			row{name: "question", in: filepath.Join(base, "a?b", "c.db"), path: rooted(filepath.Join(base, "a?b", "c.db")), contains: "/a%3Fb/c.db"},
			// Abs cleans the doubled slash: on Unix this is an ordinary rooted path, not a share.
			row{name: "double-slash", in: "//srv/share/x.db", path: "/srv/share/x.db", prefix: "file:///srv/share/x.db?"},
		)
	}
	for _, r := range rows {
		for _, c := range []struct {
			ctor  string
			fn    func(string) (string, error)
			query string
		}{{"readURI", readURI, wantRead}, {"writeURI", writeURI, wantWrite}} {
			t.Run(r.name+"/"+c.ctor, func(t *testing.T) {
				got, err := c.fn(r.in)
				if err != nil {
					t.Fatal(err)
				}
				u, err := url.Parse(got)
				if err != nil {
					t.Fatalf("%s does not parse: %v", got, err)
				}
				if u.Scheme != "file" || u.Host != "" {
					t.Errorf("%s: scheme %q host %q, want file and an empty authority", got, u.Scheme, u.Host)
				}
				if u.Path != r.path {
					t.Errorf("%s: path %q, want %q", got, u.Path, r.path)
				}
				if u.RawQuery != c.query {
					t.Errorf("%s: query %q, want %q", got, u.RawQuery, c.query)
				}
				if r.prefix != "" && !strings.HasPrefix(got, r.prefix) {
					t.Errorf("%s does not begin %q", got, r.prefix)
				}
				if r.contains != "" && !strings.Contains(got, r.contains) {
					t.Errorf("%s does not carry the escaped %q", got, r.contains)
				}
			})
		}
	}
}

// onlyTheStore fails unless the only files anywhere under root are dir/catalogue.db and, at most,
// its -wal and -shm. A truncated or mis-anchored connection creates or reads ANOTHER file —
// `<root>/a` for a path cut at `?`, `<root>/aAb` for an unescaped `%41` — and this is where it shows.
func onlyTheStore(t *testing.T, root, dir string) {
	t.Helper()
	store := filepath.Join(root, dir, "catalogue.db")
	allowed := map[string]bool{filepath.Join(root, dir): true, store: true, store + "-wal": true, store + "-shm": true}
	filepath.WalkDir(root, func(p string, _ fs.DirEntry, err error) error {
		if err != nil {
			t.Errorf("walking %s: %v", p, err)
			return nil
		}
		if p != root && !allowed[p] {
			t.Errorf("%s exists: a connection opened a file other than the one it was given", p)
		}
		return nil
	})
	if _, err := os.Stat(store); err != nil {
		t.Errorf("the named store does not exist: %v", err)
	}
}

// THE NAMED FILE, AND ONLY IT — on the write side (Open creating, then Open pre-classifying an
// existing file through readURI) and on the read side (OpenRead).
func TestOnlyTheNamedFileIsOpened(t *testing.T) {
	for _, dir := range []string{"a#b", "a%41b", "a?b", "relative"} {
		t.Run(dir, func(t *testing.T) {
			if dir == "a?b" && runtime.GOOS == "windows" {
				t.Skip("Windows reserves `?` in file and directory names, so this path cannot exist there")
			}
			root := t.TempDir()
			sub, p := dir, filepath.Join(root, dir, "catalogue.db")
			if dir == "relative" {
				t.Chdir(root)
				sub, p = "a#b", filepath.Join("a#b", "catalogue.db")
			}

			db, err := Open(p, io.Discard) // (1) creates
			if err != nil {
				t.Fatal(err)
			}
			db.Close()
			onlyTheStore(t, root, sub)

			var notice bytes.Buffer // (2) the file exists: this pre-classifies through readURI
			db, err = Open(p, &notice)
			if err != nil {
				t.Fatal(err)
			}
			var rebuilt int
			if err := db.QueryRow(`SELECT count(*) FROM meta WHERE key = ?`, MetaRebuiltAt).Scan(&rebuilt); err != nil {
				t.Fatal(err)
			}
			db.Close()
			if notice.Len() != 0 || rebuilt != 0 {
				t.Errorf("reopening a current store at %s rebuilt it (notice %q, rebuilt_at rows %d) — "+
					"the pre-classification read some other file", p, notice.String(), rebuilt)
			}
			onlyTheStore(t, root, sub)

			ro, err := OpenRead(p) // (3) refuses every class but Current
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ro.Exec(`PRAGMA query_only=0`); err != nil {
				t.Fatal(err)
			}
			_, err = ro.Exec(`INSERT INTO meta(key,value) VALUES('x','y')`)
			var se *sqlite.Error
			if !errors.As(err, &se) || se.Code()&0xff != sqlite3.SQLITE_READONLY {
				t.Errorf("an INSERT after query_only=0 on the read path returned %v, want a read-only error", err)
			}
			ro.Close()
			onlyTheStore(t, root, sub)
			if dir == "relative" {
				if _, err := os.Stat(filepath.Join(string(filepath.Separator), "a#b")); err == nil {
					t.Error("a relative path was anchored at the filesystem root")
				}
			}
		})
	}
}

// ---- recognition ----------------------------------------------------------------------------

func TestRecognition(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "catalogue.db")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		p    string
		want Class
	}{
		{"fresh: no tables", empty, Fresh},
		{"current", currentAt(t, UserVersion), Current},
		{"older: shape 1 stamped 1", shape1Store(t, t.TempDir(), 1), Older},
		{"older: shape 1 stamped 2 (this box)", shape1Store(t, t.TempDir(), 2), Older},
		{"older: complete shape 1 at stamp 0", shape1Store(t, t.TempDir(), 0), Older},
		{"older: stamp 0 holding only the shape-1 session", fixture(t, shape1Session), Older},
		{"newer: a stamp above ours, today's tables", currentAt(t, UserVersion+1), Newer},
		{"newer: a stamp above ours with an unknown table", func() string {
			p := currentAt(t, UserVersion+1)
			db := rawOpen(t, p)
			mustExec(t, db, `CREATE TABLE future_thing(x)`)
			db.Close()
			return p
		}(), Newer},
		{"newer: a stamp above ours with a session column renamed", func() string {
			p := currentAt(t, UserVersion+1)
			db := rawOpen(t, p)
			mustExec(t, db, `DROP VIEW v_session`, `ALTER TABLE session RENAME COLUMN cwd TO working_dir`)
			db.Close()
			return p
		}(), Newer},
		{"foreign: {act} alone at stamp 0", fixture(t, `CREATE TABLE act(id INTEGER)`), Foreign},
		{"foreign: {meta} alone at stamp 0", fixture(t, `CREATE TABLE meta(key TEXT PRIMARY KEY, value TEXT)`), Foreign},
		{"foreign: an unrelated table at stamp 1", fixture(t, `CREATE TABLE ledger(id INTEGER)`, `PRAGMA user_version = 1`), Foreign},
		{"foreign: session lacking capture_build", fixture(t,
			`CREATE TABLE session (session_id TEXT, project_dir TEXT, cwd TEXT, first_seen INTEGER, last_seen INTEGER, closed_at INTEGER)`,
			`CREATE TABLE file_offset(path TEXT)`, `PRAGMA user_version = 1`), Foreign},
		{"foreign: a stamp above ours without file_offset", fixture(t, shape1Session, `CREATE TABLE act(id INTEGER)`, fmt.Sprintf(`PRAGMA user_version = %d`, UserVersion+1)), Foreign},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, got, tables := classify(t, tc.p)
			if got != tc.want {
				t.Errorf("classified %s (stamp %d, tables %v), want %s", got, v, tables, tc.want)
			}
		})
	}
}

// ---- live stores ----------------------------------------------------------------------------

// A LIVE STORE IS JUDGED BY ITS TRUE STATE. The read path sees a writer's uncheckpointed frames,
// so a healthy store opened while hooks are writing it is Current — and a store rebuilt a moment
// ago is Current too, with the rebuild markers readable, rather than refused as the older shape
// its checkpointed bytes still hold.
func TestLiveStoresClassifyByTheirTrueState(t *testing.T) {
	t.Run("current, writer open", func(t *testing.T) {
		p, db := newStore(t) // the writer stays open until cleanup
		mustExec(t, db, `INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('LIVE','/p',1,1)`)
		ro, err := OpenRead(p)
		if err != nil {
			t.Fatal(err)
		}
		defer ro.Close()
		var n int
		if err := ro.QueryRow(`SELECT count(*) FROM v_session WHERE session_id='LIVE'`).Scan(&n); err != nil || n != 1 {
			t.Errorf("the read path did not see the writer's frames: n=%d err=%v", n, err)
		}
	})
	t.Run("rebuilt, writer open", func(t *testing.T) {
		p := shape1Store(t, t.TempDir(), 2)
		db, err := Open(p, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		ro, err := OpenRead(p)
		if err != nil {
			t.Fatalf("a store rebuilt with its writer still open was refused: %v", err)
		}
		defer ro.Close()
		_, from, pending, err := RebuildPending(ro)
		if err != nil || !pending || from != 2 {
			t.Errorf("RebuildPending = from %d pending %v err %v, want pending from 2", from, pending, err)
		}
	})
}

// ---- no write on refusal --------------------------------------------------------------------

func sha(t *testing.T, p string) string {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func header(t *testing.T, p string) [2]byte {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil || len(b) < 20 {
		t.Fatalf("reading %s's header: %v (%d bytes)", p, err, len(b))
	}
	return [2]byte{b[18], b[19]}
}

func listDir(t *testing.T, dir string) map[string]int64 {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]int64{}
	for _, e := range es {
		fi, err := e.Info()
		if err != nil {
			t.Fatal(err)
		}
		out[e.Name()] = fi.Size()
	}
	return out
}

func copyFile(t *testing.T, from, to string) {
	t.Helper()
	b, err := os.ReadFile(from)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// walSnapshot copies a live WAL database and its -wal, with the writer still open, into a fresh
// dir: a WAL file with PENDING FRAMES and no live writer. build runs on the source connection;
// after it, nothing may checkpoint.
func walSnapshot(t *testing.T, build func(db *sql.DB)) string {
	t.Helper()
	src := filepath.Join(t.TempDir(), "src.db")
	db := rawOpen(t, src)
	defer db.Close()
	build(db)
	dst := filepath.Join(t.TempDir(), "catalogue.db")
	copyFile(t, src, dst)
	copyFile(t, src+"-wal", dst+"-wal")
	return dst
}

// A FILE NOBODY RECOGNISED IS NEVER WRITTEN — by Open or by OpenRead. Its database bytes do not
// change; an existing -wal is neither checkpointed nor rewritten; a rollback-journal file gains
// nothing at all; and the only footprint permitted is what any SQLite reader leaves beside a WAL
// file that had no sidecars: an empty -wal and a -shm.
func TestARefusedFileIsNeverWritten(t *testing.T) {
	const ledger = `CREATE TABLE ledger(id INTEGER PRIMARY KEY, amount INTEGER)`
	for _, tc := range []struct {
		name   string
		build  func(t *testing.T) string
		refuse string
		mayAdd map[string]bool // new files allowed beside the store, by suffix
		empty  map[string]bool // ...and which of them must be empty
	}{
		{
			name: "rollback journal",
			build: func(t *testing.T) string {
				p := fixture(t, ledger, `INSERT INTO ledger(amount) VALUES(1)`)
				if h := header(t, p); h != [2]byte{1, 1} {
					t.Fatalf("fixture header %v, want 01 01 (a rollback-journal file)", h)
				}
				return p
			},
			refuse: "is not a gray-area catalogue (tables: ledger)",
		},
		{
			name: "cleanly closed WAL",
			build: func(t *testing.T) string {
				p := fixture(t, `PRAGMA journal_mode=WAL`, ledger, `INSERT INTO ledger(amount) VALUES(1)`)
				if h := header(t, p); h != [2]byte{2, 2} {
					t.Fatalf("fixture header %v, want 02 02 (WAL)", h)
				}
				return p
			},
			refuse: "is not a gray-area catalogue (tables: ledger)",
			mayAdd: map[string]bool{"-wal": true, "-shm": true},
			empty:  map[string]bool{"-wal": true},
		},
		{
			name: "WAL with pending frames",
			build: func(t *testing.T) string {
				return walSnapshot(t, func(db *sql.DB) {
					mustExec(t, db, `PRAGMA journal_mode=WAL`, ledger, `PRAGMA wal_checkpoint(TRUNCATE)`,
						`PRAGMA wal_autocheckpoint=0`, `INSERT INTO ledger(amount) VALUES(1)`, `INSERT INTO ledger(amount) VALUES(2)`)
				})
			},
			refuse: "is not a gray-area catalogue (tables: ledger)",
			mayAdd: map[string]bool{"-shm": true},
		},
		{
			name: "WAL whose tables exist only in its -wal",
			build: func(t *testing.T) string {
				p := walSnapshot(t, func(db *sql.DB) {
					mustExec(t, db, `PRAGMA journal_mode=WAL`, `CREATE TABLE scratch(x)`, `PRAGMA wal_checkpoint(TRUNCATE)`,
						`DROP TABLE scratch`, `PRAGMA wal_checkpoint(TRUNCATE)`,
						`PRAGMA wal_autocheckpoint=0`, ledger, `INSERT INTO ledger(amount) VALUES(1)`)
				})
				// The fixture proves itself: without its -wal the same database has no tables.
				bare := filepath.Join(t.TempDir(), "bare.db")
				copyFile(t, p, bare)
				if _, c, tables := classify(t, bare); c != Fresh {
					t.Fatalf("the database without its -wal is %s (%v); the schema is not only in the -wal", c, tables)
				}
				return p
			},
			refuse: "is not a gray-area catalogue (tables: ledger)",
			mayAdd: map[string]bool{"-shm": true},
		},
		{
			name:   "newer",
			build:  func(t *testing.T) string { return currentAt(t, UserVersion+1) },
			refuse: fmt.Sprintf("is stamped %d by a newer gray-area", UserVersion+1),
			mayAdd: map[string]bool{"-wal": true, "-shm": true},
			empty:  map[string]bool{"-wal": true},
		},
	} {
		for _, opener := range []struct {
			name string
			open func(string) (*sql.DB, error)
		}{
			{"Open", func(p string) (*sql.DB, error) { return Open(p, io.Discard) }},
			{"OpenRead", OpenRead},
		} {
			t.Run(tc.name+"/"+opener.name, func(t *testing.T) {
				p := tc.build(t)
				dir, base := filepath.Dir(p), filepath.Base(p)
				before := listDir(t, dir)
				dbSum := sha(t, p)
				walSum := ""
				if _, ok := before[base+"-wal"]; ok {
					walSum = sha(t, p+"-wal")
				}
				db, err := opener.open(p)
				if err == nil {
					db.Close()
					t.Fatalf("%s opened it", opener.name)
				}
				if !strings.Contains(err.Error(), tc.refuse) {
					t.Errorf("refusal = %q, want it to contain %q", err, tc.refuse)
				}
				if sha(t, p) != dbSum {
					t.Error("the database bytes changed")
				}
				if walSum != "" {
					if _, err := os.Stat(p + "-wal"); err != nil {
						t.Errorf("the existing -wal was removed: %v", err)
					} else if sha(t, p+"-wal") != walSum {
						t.Error("the existing -wal was rewritten or checkpointed")
					}
				}
				for name, size := range listDir(t, dir) {
					if _, had := before[name]; had {
						continue
					}
					suffix := strings.TrimPrefix(name, base)
					if name == suffix || !tc.mayAdd[suffix] {
						t.Errorf("%s appeared beside the refused file", name)
					} else if tc.empty[suffix] && size != 0 {
						t.Errorf("%s appeared with %d bytes, want empty", name, size)
					}
				}
			})
		}
	}
}

// ---- behaviour ------------------------------------------------------------------------------

// lockedBuffer is a notice sink several goroutines can share.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// AN OLDER CATALOGUE IS REBUILT ONCE, AND SAYS SO — including a stamp-2 store with shape-1 columns,
// which is what every store created before #867 now is, and a partial one left at stamp 0.
func TestAnOlderCatalogueIsRebuilt(t *testing.T) {
	for _, tc := range []struct {
		name string
		p    func(t *testing.T) string
		from int
	}{
		{"shape 1 stamped 1", func(t *testing.T) string { return shape1Store(t, t.TempDir(), 1) }, 1},
		{"shape 1 stamped 2", func(t *testing.T) string { return shape1Store(t, t.TempDir(), 2) }, 2},
		{"shape 1 at stamp 0", func(t *testing.T) string { return shape1Store(t, t.TempDir(), 0) }, 0},
		{"only session at stamp 0", func(t *testing.T) string { return fixture(t, shape1Session) }, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.p(t)
			var notice bytes.Buffer
			db, err := Open(p, &notice)
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("catalogue: %s was stamped %d, rebuilt empty at shape %d — run `telepathy backfill` to re-read the corpus\n",
				p, tc.from, UserVersion)
			if notice.String() != want {
				t.Errorf("notice = %q, want %q", notice.String(), want)
			}
			var v, n int
			db.QueryRow(`PRAGMA user_version`).Scan(&v)
			if v != UserVersion {
				t.Errorf("stamped %d, want %d", v, UserVersion)
			}
			if err := db.QueryRow(`SELECT count(*) FROM v_session`).Scan(&n); err != nil || n != 0 {
				t.Errorf("v_session after the rebuild: n=%d err=%v (want 0 rows, queryable)", n, err)
			}
			if _, from, pending, err := RebuildPending(db); err != nil || !pending || from != tc.from {
				t.Errorf("RebuildPending = from %d pending %v err %v", from, pending, err)
			}
			tables, _ := names(db, `SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
			if got := strings.Join(tables, ","); got != "act,file_offset,meta,session,skip,thought,word" {
				t.Errorf("tables after the rebuild = %s", got)
			}
			db.Close()

			notice.Reset()
			db, err = Open(p, &notice)
			if err != nil {
				t.Fatal(err)
			}
			db.Close()
			if notice.Len() != 0 {
				t.Errorf("a second open rebuilt again: %q", notice.String())
			}
		})
	}
}

// A SHAPE-4 STORE IS RECOGNISED OLDER AND REBUILT AT 5, WITH THE CAPTURE COLUMN. #894's mechanism
// carries the bump unchanged: bridge_session_id is not in sessionCore, so a shape-4 `session`
// still reads as a catalogue, and the rebuild applies today's DDL.
func TestAShape4StoreIsRebuiltWithTheCaptureColumn(t *testing.T) {
	p := currentAt(t, 4)
	raw := rawOpen(t, p)
	// What the shape-4 binary left: no capture column on `session`, none on v_session.
	mustExec(t, raw, `DROP VIEW v_session`, `ALTER TABLE session DROP COLUMN bridge_session_id`,
		`CREATE VIEW v_session AS SELECT s.session_id, s.project_dir, s.cwd, NULL AS first_act, NULL AS last_act,
		     s.ingested_first, s.ingested_last, s.closed_at, s.capture_build FROM session s`,
		`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last,closed_at) VALUES('OLD','/p',1,1,1)`)
	raw.Close()
	if v, c, _ := classify(t, p); v != 4 || c != Older {
		t.Fatalf("a shape-4 store classified %v at stamp %d, want older at 4", c, v)
	}
	var notice bytes.Buffer
	db, err := Open(p, &notice)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if want := fmt.Sprintf("was stamped 4, rebuilt empty at shape %d", UserVersion); !strings.Contains(notice.String(), want) {
		t.Errorf("notice = %q, want it to say %q", notice.String(), want)
	}
	cols, err := names(db, `SELECT name FROM pragma_table_info('session')`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(","+strings.Join(cols, ",")+",", ",bridge_session_id,") {
		t.Errorf("session after the rebuild = %v, want bridge_session_id among them", cols)
	}
	rows, err := db.Query(`SELECT * FROM v_session LIMIT 0`)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := rows.Columns()
	rows.Close()
	if strings.Join(got, ",") != strings.Join(ViewColumns["v_session"], ",") {
		t.Errorf("v_session after the rebuild = %v, want %v", got, ViewColumns["v_session"])
	}
	var n int
	if db.QueryRow(`SELECT count(*) FROM session`).Scan(&n); n != 0 {
		t.Errorf("%d sessions survived the rebuild, want 0", n)
	}
}

func TestACurrentStoreKeepsItsRows(t *testing.T) {
	p, db := newStore(t)
	mustExec(t, db, `INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('KEEP','/p',1,1)`)
	db.Close()
	var notice bytes.Buffer
	db2, err := Open(p, &notice)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	var n int
	if err := db2.QueryRow(`SELECT count(*) FROM session WHERE session_id='KEEP'`).Scan(&n); err != nil || n != 1 {
		t.Errorf("a current store lost its rows: n=%d err=%v", n, err)
	}
	if notice.Len() != 0 {
		t.Errorf("a current store printed a notice: %q", notice.String())
	}
}

// EXACTLY ONE REBUILD UNDER CONCURRENT OPENERS, because the decision is re-taken inside one
// IMMEDIATE transaction. A second rebuild would wipe the rows the first opener has already
// written — so the first to return writes one, and it must survive every other opener.
func TestConcurrentOpenersRebuildOnce(t *testing.T) {
	const n = 8
	t.Run("older", func(t *testing.T) {
		p := shape1Store(t, t.TempDir(), 2)
		var notice lockedBuffer
		var first sync.Once
		var wg sync.WaitGroup
		errs := make(chan error, 2*n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db, err := Open(p, &notice)
				if err != nil {
					errs <- err
					return
				}
				defer db.Close()
				first.Do(func() {
					if _, err := db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('FIRST','/p',1,1)`); err != nil {
						errs <- err
					}
				})
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Error(err)
		}
		if got := strings.Count(notice.String(), "rebuilt empty"); got != 1 {
			t.Errorf("%d rebuild notices under %d openers, want 1:\n%s", got, n, notice.String())
		}
		ro, err := OpenRead(p)
		if err != nil {
			t.Fatal(err)
		}
		defer ro.Close()
		var rows int
		if err := ro.QueryRow(`SELECT count(*) FROM session WHERE session_id='FIRST'`).Scan(&rows); err != nil || rows != 1 {
			t.Errorf("the row written after the first return did not survive: n=%d err=%v", rows, err)
		}
	})
	t.Run("fresh", func(t *testing.T) {
		p := filepath.Join(t.TempDir(), "catalogue.db")
		var notice lockedBuffer
		var wg sync.WaitGroup
		errs := make(chan error, n)
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				db, err := Open(p, &notice)
				if err != nil {
					errs <- err
					return
				}
				db.Close()
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			t.Error(err)
		}
		if notice.String() != "" {
			t.Errorf("a fresh path printed a rebuild notice: %q", notice.String())
		}
		ro, err := OpenRead(p)
		if err != nil {
			t.Fatal(err)
		}
		defer ro.Close()
		if _, _, pending, err := RebuildPending(ro); err != nil || pending {
			t.Errorf("a fresh store reads as rebuilt: pending %v err %v", pending, err)
		}
		var marks int
		if err := ro.QueryRow(`SELECT count(*) FROM meta WHERE key = ?`, MetaRebuiltAt).Scan(&marks); err != nil || marks != 0 {
			t.Errorf("rebuilt_at rows on a fresh store = %d (err %v)", marks, err)
		}
	})
}

// OpenRead REFUSES IN WORDS, per class, and never with a SQL error from querying a shape it does
// not know — the old read path ran `v_session` against shape-1 columns and printed
// "no such column: s.ingested_first".
func TestOpenReadRefusesEachClassInWords(t *testing.T) {
	empty := filepath.Join(t.TempDir(), "catalogue.db")
	if err := os.WriteFile(empty, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, p, want string
	}{
		{"older", shape1Store(t, t.TempDir(), 2), fmt.Sprintf("is stamped 2; this binary reads shape %d. The next session start or turn end", UserVersion)},
		{"newer", currentAt(t, UserVersion+1), fmt.Sprintf("is stamped %d by a newer gray-area (this binary writes %d)", UserVersion+1, UserVersion)},
		{"foreign", fixture(t, `CREATE TABLE ledger(id INTEGER)`), "is not a gray-area catalogue (tables: ledger) — refusing to touch it; check `--store`"},
		{"fresh", empty, "holds no catalogue yet — let a session run or run `telepathy backfill`"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, err := OpenRead(tc.p)
			if err == nil {
				db.Close()
				t.Fatal("opened")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("refusal = %q, want it to contain %q", err, tc.want)
			}
		})
	}
}

// BenchmarkOpen reports what Open costs on a current store — the read-only pre-classification,
// then the IMMEDIATE transaction — which every capture hook pays. Reported, not asserted.
func BenchmarkOpen(b *testing.B) {
	p := filepath.Join(b.TempDir(), "catalogue.db")
	db, err := Open(p, io.Discard)
	if err != nil {
		b.Fatal(err)
	}
	db.Close()
	for b.Loop() {
		db, err := Open(p, io.Discard)
		if err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}
