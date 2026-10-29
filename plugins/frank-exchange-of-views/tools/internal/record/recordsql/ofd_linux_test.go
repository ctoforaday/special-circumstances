package recordsql_test

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// These tests stand a second PROCESS beside a lock, because a lock's owner is what is under test:
// POSIX locks belong to a process and OFD locks to a descriptor, and two connections inside one
// process share SQLite's per-inode lock state, so they never ask the kernel about each other. The
// test binary re-executes itself as a lock holder or a prober (TestLockChild); each child is one
// short process, so the whole table costs about a second.

const (
	lockChildEnv = "RECORDSQL_LOCK_CHILD" // "hold" or "probe"
	lockKindEnv  = "RECORDSQL_LOCK_KIND"  // "posix" or "ofd": the lock kind the child takes
	lockShapeEnv = "RECORDSQL_LOCK_SHAPE" // a lockShapes key
	lockDBEnv    = "RECORDSQL_LOCK_DB"
	childMark    = "LOCKCHILD"
)

// lockShapes are the locks a record's connections take, each as the holder's open and act and the
// prober's open and act. The prober runs with busy_timeout 0, so a conflict answers SQLITE_BUSY at
// once instead of waiting it out.
var lockShapes = map[string]struct {
	journal         string // the journal_mode both sides open with
	exclusiveProbe  bool   // the prober opens in locking_mode EXCLUSIVE
	holderInWriteTx bool   // the holder sits inside an immediate transaction rather than idle
	proberAct       string // the statement the prober issues
}{
	// The record's write lock. In WAL mode it is a lock in -shm, POSIX in every binary.
	"wal-write": {journal: "WAL", holderInWriteTx: true, proberAct: "BEGIN IMMEDIATE"},
	// The last-closer check. An idle WAL connection holds SHARED on the database file; a connection
	// deciding it is the last one, or opening in exclusive mode, needs EXCLUSIVE on it. This is the
	// lock whose loss lets another process reset a log this one still writes to.
	"wal-exclusive": {journal: "WAL", exclusiveProbe: true, proberAct: "SELECT count(*) FROM sqlite_master"},
	// A write lock on the database file itself (RESERVED), which rollback-journal mode takes.
	"rollback-write": {journal: "DELETE", holderInWriteTx: true, proberAct: "BEGIN IMMEDIATE"},
}

func lockDSN(path, journal string, busyMs int, exclusive bool) string {
	q := fmt.Sprintf("_txlock=immediate&_pragma=busy_timeout(%d)", busyMs)
	if exclusive {
		// Exclusive mode must be in force before the connection's first read: that read opens the
		// log, and an exclusive-mode connection opening a log takes EXCLUSIVE on the database file.
		// A journal_mode pragma would read first, so the database's own persistent mode stands.
		q += "&_pragma=locking_mode(EXCLUSIVE)"
	} else {
		q += fmt.Sprintf("&_pragma=journal_mode(%s)", journal)
	}
	return (&url.URL{Scheme: "file", Path: path, RawQuery: q}).String()
}

// TestLockChild is the child half of the tests below; run directly it does nothing.
func TestLockChild(t *testing.T) {
	role := os.Getenv(lockChildEnv)
	if role == "" {
		return
	}
	kind, path := os.Getenv(lockKindEnv), os.Getenv(lockDBEnv)
	shape, ok := lockShapes[os.Getenv(lockShapeEnv)]
	if !ok {
		t.Fatalf("unknown lock shape %q", os.Getenv(lockShapeEnv))
	}
	// The kind is set before this process's first lock, which is the only moment it can be.
	if _, err := sqlite.OFDLocking(kind == "ofd"); err != nil {
		t.Fatalf("setting %s locks: %v", kind, err)
	}
	took := func() string {
		if sqlite.OFDLockingEnabled() {
			return "ofd"
		}
		return "posix"
	}
	switch role {
	case "hold":
		db, err := sql.Open("sqlite", lockDSN(path, shape.journal, 5000, false))
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS t (x)`); err != nil {
			t.Fatal(err)
		}
		if shape.holderInWriteTx {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(`INSERT INTO t VALUES (1)`); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = tx.Rollback() }()
		}
		fmt.Println(childMark, "READY", took())
		_, _ = io.Copy(io.Discard, os.Stdin) // held until the parent closes stdin
	case "probe":
		db, err := sql.Open("sqlite", lockDSN(path, shape.journal, 0, shape.exclusiveProbe))
		if err != nil {
			t.Fatal(err)
		}
		db.SetMaxOpenConns(1)
		_, err = db.Exec(shape.proberAct)
		var se *sqlite.Error
		switch {
		case err == nil:
			fmt.Println(childMark, "GOT", took())
		case errors.As(err, &se) && se.Code()&0xff == sqlite3.SQLITE_BUSY:
			fmt.Println(childMark, "BUSY", took())
		default:
			fmt.Println(childMark, "ERR", took(), err)
		}
		_ = db.Close()
	}
}

func lockChild(role, kind, shape, path string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], "-test.run=^TestLockChild$", "-test.count=1")
	cmd.Env = append(os.Environ(), lockChildEnv+"="+role, lockKindEnv+"="+kind, lockShapeEnv+"="+shape, lockDBEnv+"="+path)
	return cmd
}

// childLine is the child's one result line: its verdict and the lock kind it actually took.
func childLine(t *testing.T, out string) (verdict, kind string) {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if f := strings.Fields(line); len(f) >= 3 && f[0] == childMark {
			return f[1], f[2]
		}
	}
	t.Fatalf("the child printed no result line:\n%s", out)
	return "", ""
}

// probe runs a prober of kind against path and returns its verdict, failing the test when the
// prober did not take the kind it was asked to.
func probe(t *testing.T, kind, shape, path string) string {
	t.Helper()
	out, err := lockChild("probe", kind, shape, path).CombinedOutput()
	if err != nil {
		t.Fatalf("prober (%s, %s): %v\n%s", kind, shape, err, out)
	}
	verdict, took := childLine(t, string(out))
	if took != kind {
		t.Fatalf("the prober asked for %s locks and took %s", kind, took)
	}
	if verdict == "ERR" {
		t.Fatalf("the prober (%s, %s) failed for another reason:\n%s", kind, shape, out)
	}
	return verdict
}

// hold starts a holder of kind on path and returns once its lock is held; the cleanup releases it.
func hold(t *testing.T, kind, shape, path string) {
	t.Helper()
	cmd := lockChild("hold", kind, shape, path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = cmd.Wait()
	})
	ready := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(stdout)
		for sc.Scan() {
			if f := strings.Fields(sc.Text()); len(f) >= 3 && f[0] == childMark {
				ready <- f[2]
				break
			}
		}
		_, _ = io.Copy(io.Discard, stdout)
		close(ready)
	}()
	select {
	case took, ok := <-ready:
		if !ok {
			t.Fatalf("the holder (%s, %s) exited before holding its lock", kind, shape)
		}
		if took != kind {
			t.Fatalf("the holder asked for %s locks and took %s", kind, took)
		}
	case <-time.After(60 * time.Second):
		t.Fatalf("the holder (%s, %s) never reported its lock held", kind, shape)
	}
}

// requireOFD skips where this kernel or filesystem rejects OFD locks: the library falls back to
// POSIX locks there, and there is no OFD lock to test. It probes in a child, because the probe's
// first lock fixes the mode of the process that takes it.
func requireOFD(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "probe.db")
	out, err := lockChild("probe", "ofd", "wal-write", path).CombinedOutput()
	if err != nil {
		t.Fatalf("OFD probe: %v\n%s", err, out)
	}
	if _, took := childLine(t, string(out)); took != "ofd" {
		t.Skipf("this kernel or filesystem rejects OFD locks; SQLite locks POSIX here")
	}
}

// A BINARY THAT LOCKS POSIX AND ONE THAT LOCKS OFD EXCLUDE EACH OTHER ON ONE RECORD. During an
// upgrade an old binary's seat and a new binary's seat share a record; the kernel resolves a POSIX
// and an OFD lock on one file range against each other, so neither kind can take a lock the other
// holds. Each case is first run with no holder, so a BUSY below is the holder's and not a probe
// that cannot succeed.
func TestPosixAndOFDLocksExcludeEachOther(t *testing.T) {
	requireOFD(t, t.TempDir())
	for shape := range lockShapes {
		for _, pair := range [][2]string{{"posix", "ofd"}, {"ofd", "posix"}} {
			holder, prober := pair[0], pair[1]
			t.Run(fmt.Sprintf("%s/%s-holds-%s-probes", shape, holder, prober), func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "record.db")
				seed, err := sql.Open("sqlite", lockDSN(path, lockShapes[shape].journal, 5000, false))
				if err != nil {
					t.Fatal(err)
				}
				if _, err := seed.Exec(`CREATE TABLE t (x)`); err != nil {
					t.Fatal(err)
				}
				_ = seed.Close()
				if got := probe(t, prober, shape, path); got != "GOT" {
					t.Fatalf("with no holder the %s prober answered %s — the probe cannot succeed, so a BUSY proves nothing", prober, got)
				}
				hold(t, holder, shape, path)
				if got := probe(t, prober, shape, path); got != "BUSY" {
					t.Errorf("a %s prober took the %s lock a %s holder holds (%s): the two kinds do not exclude each other",
						prober, shape, holder, got)
				}
			})
		}
	}
}

// A STRAY DESCRIPTOR ON THE RECORD LEAVES ITS LOCK HELD. The record's connection holds SHARED on
// record.db while it is idle; an os.Open and Close of record.db in the same process is the act that
// drops a POSIX lock. With the database file locked OFD, the lock survives it: this process still
// holds it, and another process — of either kind, the new binary or an old one — still cannot take
// the EXCLUSIVE lock that would let it reset the log this connection writes to.
func TestAStrayCloseOfTheRecordLeavesItLocked(t *testing.T) {
	dir := t.TempDir()
	requireOFD(t, dir)
	path := filepath.Join(dir, "record.db")
	db, err := recordsql.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.CloseUnder(dir) })
	if _, err := sqlite.OFDLocking(true); err != nil {
		t.Errorf("after the record opened, this process's lock mode is not OFD: %v", err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master`).Scan(&n); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Read(make([]byte, 4096)); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if held, measured := recordtest.LocksHeld(t, path); measured && held == 0 {
		t.Errorf("after a stray close of record.db this process holds no lock on it while its connection is open")
	}
	for _, kind := range []string{"posix", "ofd"} {
		if got := probe(t, kind, "wal-exclusive", path); got != "BUSY" {
			t.Errorf("after a stray close of record.db a %s process took EXCLUSIVE on it (%s) while this process's "+
				"connection is open — it could reset the log this connection writes to", kind, got)
		}
	}

	// The control: with this process's connection closed the same probe succeeds, so the BUSY above
	// was the connection's lock.
	if err := recordsql.CloseUnder(dir); err != nil {
		t.Fatal(err)
	}
	if got := probe(t, "posix", "wal-exclusive", path); got != "GOT" {
		t.Errorf("with the record closed the probe answered %s — it cannot succeed, so the BUSY above proves nothing", got)
	}
}
