package recordsql

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// TestConcurrentOpenOnFreshDatabase is the gate #557 said had to arrive with its fix.
//
// `Open` reads whether `events` exists and applies the schema if it does not. With nothing
// between the two, two openers of a FRESH database both saw zero and both applied; the second
// failed with "table … already exists". The issue records why nothing ever hit it, and both
// reasons are accidents rather than guarantees:
//
//   - the in-process handle cache means one process cannot race itself, so this test calls
//     openUncached — the cache is exactly what has to be bypassed to ask the question;
//   - the cross-process test seeds the schema before spawning its children, so the racing
//     branch is unreachable there by construction.
//
// Production was covered by the same accident: `setup` creates the run directory and its
// database before any seat is dispatched. The fuzz removes it — it builds its run directory
// directly, so the first seat command creates the schema, and with concurrent lanes (#630)
// that is several processes at once on round 0 of every run.
func TestConcurrentOpenOnFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")

	// SIX, not two. Two openers can pass by luck of scheduling; the failure this guards is a
	// window, and more racers is more of the window sampled per run.
	const openers = 6
	var wg sync.WaitGroup
	errs := make([]error, openers)
	start := make(chan struct{})
	for i := range openers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start // released together, so they contend rather than queue politely
			db, err := openUncached(path)
			errs[i] = err
			if db != nil {
				_ = db.Close()
			}
		}(i)
	}
	close(start)
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("opener %d failed on a fresh database: %v", i, err)
		}
	}

	// AND THE SCHEMA IS USABLE, not merely present. A losing opener that swallowed its error
	// would satisfy the loop above while handing back a database nothing can write.
	db, err := openUncached(path)
	if err != nil {
		t.Fatalf("reopening after the race: %v", err)
	}
	defer db.Close()
	if has, err := hasEvents(db); err != nil || !has {
		t.Fatalf("no events table after six concurrent opens (%v) — every opener thought another had made it", err)
	}
}

// TestOpenWaitsOutALockedFreshDatabase pins the connect-time wait deterministically, where the
// race above only samples it: another connection holds the fresh file's write lock, in rollback
// mode, for longer than the loser used to take to fail. The open must come back with the schema,
// not with SQLITE_BUSY.
func TestOpenWaitsOutALockedFreshDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")
	tx := holdFresh(t, path)

	const held = 300 * time.Millisecond
	type opened struct {
		at  time.Time // when the open RETURNED, read in the goroutine
		err error
	}
	done := make(chan opened, 1)
	go func() {
		db, err := openUncached(path)
		if db != nil {
			_ = db.Close()
		}
		done <- opened{time.Now(), err}
	}()
	time.Sleep(held)
	if _, err := tx.Exec(`DROP TABLE placeholder`); err != nil {
		t.Fatal(err)
	}
	// The lock is released inside Commit, so an open that waited returns no earlier than the
	// moment Commit was entered. An open that did not wait returned during the Sleep above. The
	// clock is read before Commit rather than after it, because the waiter can return between
	// the release inside Commit and a reading taken on Commit's return.
	committing := time.Now()
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("open against a briefly locked fresh database: %v", got.err)
	}
	if got.at.Before(committing) {
		t.Fatalf("open returned %v before the lock was released, which was held for %v — it did not wait", committing.Sub(got.at), held)
	}
}

// holdFresh opens the fresh database at path on a plain rollback-journal connection and leaves a
// write transaction open on it. The first write takes a RESERVED lock, which is what blocks another
// connection's `journal_mode(WAL)` conversion — the pragma needs the write lock and RESERVED
// already denies it; EXCLUSIVE is taken only at this transaction's commit. The caller commits or
// rolls back to release it.
func holdFresh(t *testing.T, path string) *sql.Tx {
	t.Helper()
	holder, err := sql.Open(driverName(), dsnWithQuery(path, ""))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = holder.Close() })
	holder.SetMaxOpenConns(1)
	tx, err := holder.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(`CREATE TABLE placeholder (x)`); err != nil {
		t.Fatal(err)
	}
	return tx
}

// TestOpenGivesUpWhenTheLockOutlivesTheBudget is the deadline half of connect's contract: a lock
// held past openBusyBudget comes back as the busy error, never as a hang. The budget is shortened so
// the test proves the expiry rather than waiting five seconds for it.
func TestOpenGivesUpWhenTheLockOutlivesTheBudget(t *testing.T) {
	was := openBusyBudget
	openBusyBudget = 150 * time.Millisecond
	t.Cleanup(func() { openBusyBudget = was })

	path := filepath.Join(t.TempDir(), "fresh.db")
	tx := holdFresh(t, path)
	defer func() { _ = tx.Rollback() }()

	started := time.Now()
	db, err := openUncached(path)
	if db != nil {
		_ = db.Close()
	}
	took := time.Since(started)
	if err == nil {
		t.Fatal("opened against a lock that was never released")
	}
	if !isBusy(err) {
		t.Fatalf("the budget expired with an error that is not busy: %v", err)
	}
	// An exhausted wait names its budget, so it reads differently from an immediate refusal.
	if want := "still busy after " + openBusyBudget.String(); !strings.Contains(err.Error(), want) {
		t.Errorf("the exhausted wait does not say %q: %v", want, err)
	}
	if took < openBusyBudget {
		t.Fatalf("gave up after %v, before the %v budget", took, openBusyBudget)
	}
	if took > was {
		t.Fatalf("gave up after %v — the shortened %v budget was not the one waited with", took, openBusyBudget)
	}
}

// olderBinaryCreatesFirst is the stub for a different binary winning the first create: it writes
// this binary's whole schema and stamps it one epoch OLDER, inside a transaction it leaves open so
// the caller decides when the win lands.
func olderBinaryCreatesFirst(t *testing.T, path string) *sql.Tx {
	t.Helper()
	other, err := sql.Open(driverName(), dsnWithQuery(path, "_txlock=immediate"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = other.Close() })
	other.SetMaxOpenConns(1)
	tx, err := other.Begin()
	if err != nil {
		t.Fatal(err)
	}
	schema, err := Schema()
	if err != nil {
		t.Fatal(err)
	}
	for _, ddl := range []string{schema, ViewsDDL} {
		if _, err := tx.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := tx.Exec(`INSERT INTO "schema_epoch" ("id", "epoch") VALUES (1, ?)`, recordpb.EventSchema-1); err != nil {
		t.Fatal(err)
	}
	return tx
}

// TestTheLoserOfAFirstCreateIsHeldToTheDeclaredSchema: whether this open CREATED the record is
// decided under the create's own lock, so the opener that queued behind another binary's create
// refuses that binary's record by its declared schema — never by a bare SQLite error later. A
// lock-free read taken before the create answered "did not exist" to the loser and let it through.
func TestTheLoserOfAFirstCreateIsHeldToTheDeclaredSchema(t *testing.T) {
	path := filepath.Join(t.TempDir(), "fresh.db")
	// The loser connects and sees a fresh file — the lock-free read answers "no table" here.
	loser, err := sql.Open(driverName(), dsnFor(path))
	if err != nil {
		t.Fatal(err)
	}
	defer loser.Close()
	loser.SetMaxOpenConns(1)
	if err := connect(loser); err != nil {
		t.Fatal(err)
	}
	if has, err := hasEvents(loser); err != nil || has {
		t.Fatalf("fresh database: hasEvents = (%v, %v)", has, err)
	}
	// The other binary wins the create while the loser is on its way to the lock.
	win := olderBinaryCreatesFirst(t, path)
	done := make(chan struct {
		created bool
		err     error
	}, 1)
	go func() {
		created, err := ensureSchema(loser) // queues on Begin behind the winner's write lock
		done <- struct {
			created bool
			err     error
		}{created, err}
	}()
	time.Sleep(200 * time.Millisecond)
	if err := win.Commit(); err != nil {
		t.Fatal(err)
	}
	got := <-done
	if got.err != nil {
		t.Fatalf("ensureSchema behind another binary's create: %v", got.err)
	}
	if got.created {
		t.Fatal("ensureSchema reports that it created a record another binary had already written")
	}
	// And what the open does with that answer: the full path, against the record the other
	// binary left, refuses with the declared-schema message rather than a SQLite error.
	db, err := openUncached(path)
	if db != nil {
		_ = db.Close()
	}
	if err == nil {
		t.Fatal("opened another binary's older-epoch record without refusing it")
	}
	if !strings.Contains(err.Error(), "event-schema epoch") || !strings.Contains(err.Error(), "older binary") {
		t.Fatalf("the refusal does not name the declared-schema cause: %v", err)
	}
}

// TestOpenFailsAtOnceOnAnUnreadableDatabase is the other half of connect's contract: only
// SQLITE_BUSY is waited for. A file that is not a database is refused immediately, not after the
// busy timeout.
func TestOpenFailsAtOnceOnAnUnreadableDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "junk.db")
	if err := os.WriteFile(path, []byte("not a database, and long enough to be read as one"), 0o644); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	db, err := openUncached(path)
	if db != nil {
		_ = db.Close()
	}
	if err == nil {
		t.Fatal("a file that is not a database opened")
	}
	if isBusy(err) {
		t.Fatalf("refused as busy rather than unreadable: %v", err)
	}
	if took := time.Since(started); took >= openBusyBudget {
		t.Fatalf("an unreadable database took %v to refuse — it was waited for as if busy", took)
	}
}
