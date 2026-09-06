package recordsql

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// CONCURRENT SEATS MUST NOT LOSE ACTS, AND THIS IS THE HAZARD THE STORAGE CHANGE INTRODUCED.
//
// # Why the shard layout never needed this test
//
// One seat owned one file, so two writers never touched the same bytes. Contention was designed
// out by the naming scheme. One database per run trades that for real concurrency control, and the
// trade is only sound if the new mechanism is exercised — a green suite of single-writer tests says
// nothing about eight seats mid-round.
//
// # What it caught
//
// A default `BEGIN` is DEFERRED: no lock at BEGIN, a READ lock at the first SELECT, then an
// attempted UPGRADE at the INSERT. The write path counts a seat's existing events before inserting,
// so it is exactly that shape — and SQLite does NOT apply busy_timeout to an upgrade, because two
// readers both waiting to upgrade would deadlock. It returns SQLITE_BUSY at once. Measured through
// the real binary: 8 seat processes writing 5 events each lost about HALF to `database is locked
// (5)`. Not a slow path — a refused act, with the seat told its record failed.
//
// # IT MUST BE PROCESSES, and the first version of this test was wrong about that
//
// The first attempt used goroutines with separate *sql.DB handles, on the reasoning that separate
// handles take the same locks separate processes do. It passed WITH THE FIX REVERTED — so it was
// asserting nothing, and would have shipped as evidence for a property it never tested. What makes
// the difference is process boundaries: SQLite tracks locks per process through shared inode state,
// so two handles inside one process do not collide the way two processes do.
//
// So the test re-executes its own binary. A child writes and exits; the parent counts. That is
// slower and it is the only version that fails when the fix is removed — which is the only property
// that makes a regression test worth having.
const concChildEnv = "RECORDSQL_CONCURRENCY_CHILD"

func TestConcurrentSeatsDoNotLoseEvents(t *testing.T) {
	if seat := os.Getenv(concChildEnv); seat != "" {
		writeAsChild(t, os.Getenv("RECORDSQL_CONCURRENCY_DB"), seat)
		return
	}

	path := filepath.Join(tmpRun(t), "record.db")
	seed, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	// NOT seed.Close(): Open CACHES per database, so closing here poisons the cache and the next
	// Open returns the same closed handle. The contention this test measures is between separate
	// PROCESSES (concChildEnv), which the in-process cache does not touch.
	_ = seed

	const seats = 8
	var wg sync.WaitGroup
	fails := make(chan string, seats)
	for s := 0; s < seats; s++ {
		wg.Add(1)
		go func(s int) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=^TestConcurrentSeatsDoNotLoseEvents$")
			cmd.Env = append(os.Environ(),
				concChildEnv+"=red-merge-r"+strconv.Itoa(s+1),
				"RECORDSQL_CONCURRENCY_DB="+path)
			if out, err := cmd.CombinedOutput(); err != nil {
				fails <- string(out)
			}
		}(s)
	}
	wg.Wait()
	close(fails)
	for out := range fails {
		t.Errorf("a seat's act was REFUSED under concurrency:\n%s\nThe seat is told its record failed; "+
			"in a run that is a lost act, not a retry", out)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = CloseAll() })
	// NO ERRORS IS NOT NO LOSS, so the rows are counted rather than inferred from silence.
	var got int
	if err := db.QueryRow(`SELECT count(*) FROM "friction_none"`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if want := seats * concEach; got != want {
		t.Errorf("%d acts recorded, %d written — the difference vanished with nothing reporting it", got, want)
	}
	// Every body must have its envelope: a half-written act replays as a seat that did nothing.
	var orphans int
	if err := db.QueryRow(`SELECT count(*) FROM "friction_none" f LEFT JOIN "events" e ON e."id" = f."event_id" WHERE e."id" IS NULL`).Scan(&orphans); err != nil {
		t.Fatal(err)
	}
	if orphans != 0 {
		t.Errorf("%d bodies have no envelope — the insert transaction is not atomic under contention", orphans)
	}
}

const concEach = 5

func writeAsChild(t *testing.T, path, seat string) {
	db, err := Open(path)
	if err != nil {
		t.Fatalf("child %s could not open the record: %v", seat, err)
	}
	t.Cleanup(func() { _ = CloseAll() })
	for i := 0; i < concEach; i++ {
		ev := &recordpb.Event{}
		if _, err := recordpb.SetBody(ev, &recordpb.FrictionNone{
			Text: proto.String("nothing blocked this sitting"),
		}); err != nil {
			t.Fatal(err)
		}
		ev.SeatId = proto.String(seat)
		ev.Round = proto.Int32(1)
		ev.Ts = proto.String("2026-01-01T00:00:00Z")

		// THE SHAPE IS THE POINT: COUNT, THEN WRITE, IN ONE TRANSACTION.
		//
		// A plain Insert does not reproduce the hazard, because it never reads before it writes —
		// its BEGIN goes straight to a write lock whether or not the transaction is deferred. The
		// real write path (record.insertNumbered) counts a seat's existing events to number the new
		// one, and it is that READ-then-UPGRADE that SQLite refuses to apply busy_timeout to. A
		// test that skipped the count passed with the fix reverted, which is how this version came
		// to exist.
		if err := func() error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback() //nolint:errcheck // the commit is what matters
			var prior int
			if err := tx.QueryRow(`SELECT count(*) FROM "events" WHERE "seat_id" = ? AND "type" = ?`,
				seat, "friction_none").Scan(&prior); err != nil {
				return err
			}
			ev.Key = proto.String(seat + ":friction_none:#" + strconv.Itoa(prior+1))
			if _, err := InsertTx(tx, ev); err != nil {
				return err
			}
			return tx.Commit()
		}(); err != nil {
			t.Fatalf("child %s act %d refused: %v", seat, i, err)
		}
	}
}
