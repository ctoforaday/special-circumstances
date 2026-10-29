package cli

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE VERDICT'S CHECKPOINT LEAVES THIS PROCESS'S HOLD ON THE RECORD WHERE IT WAS.
//
// The verb records the verdict through the cached connection and then mirrors the record. Under
// POSIX fcntl locking a descriptor opened on record.db or its -shm and closed again — a file copy — drops every
// POSIX lock the process holds on that file, and the connection carries on believing it still
// holds them. Measured from the kernel's own table, because a lost lock changes no answer the
// connection gives.
//
// AND THE MIRROR IS THE RECORD: one database file, openable alone, carrying the verdict the act
// just committed — not a database, log and shared memory copied at three different moments.
func TestTheVerdictCheckpointKeepsTheRecordLockedAndMirrorsOneConsistentDatabase(t *testing.T) {
	runDir := castRun(t)
	chairSitsAgain(t, runDir)
	// A FILE IN THE MIRROR THAT IS NOT THE SNAPSHOT is removed: a log beside a fresh database
	// is replayed into it by the next open.
	want, err := record.MirrorDir(runDir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(want, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(want, "record.db-wal"), []byte("not this run's log"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := run(t, "verdict", "--run", runDir, "--seat-id", "red-chair", "--as", "FAIL")
	if err != nil {
		t.Fatalf("verdict: %v\n%s", err, out)
	}

	live := filepath.Join(runDir, "records", "record.db")
	for _, f := range []string{live, live + "-shm"} {
		if n, measured := recordtest.LocksHeld(t, f); measured && n == 0 {
			t.Errorf("after the verdict, this process holds no lock on %s while its connection is still open — "+
				"something opened and closed a descriptor on the live record beside it", filepath.Base(f))
		}
	}

	_, after, ok := strings.Cut(out, "checkpointed to ")
	if !ok {
		t.Fatalf("the verdict names no checkpoint:\n%s", out)
	}
	mirror, _, _ := strings.Cut(after, "\n")
	if mirror != want {
		t.Fatalf("the verdict checkpointed to %s; record.MirrorDir names %s", mirror, want)
	}
	entries, err := os.ReadDir(mirror)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if len(names) != 1 || names[0] != "record.db" {
		t.Fatalf("the mirror holds %v, want exactly [record.db]", names)
	}

	liveDB, err := recordsql.Open(live)
	if err != nil {
		t.Fatal(err)
	}
	snap, err := sql.Open("sqlite", "file:"+filepath.ToSlash(filepath.Join(mirror, "record.db"))+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Close()
	const q = `SELECT COUNT(*), MAX("id"), (SELECT COUNT(*) FROM "events" WHERE "type" = 'verdict') FROM "events"`
	var wantN, wantMax, wantVerdicts, gotN, gotMax, gotVerdicts int64
	if err := liveDB.QueryRow(q).Scan(&wantN, &wantMax, &wantVerdicts); err != nil {
		t.Fatal(err)
	}
	if err := snap.QueryRow(q).Scan(&gotN, &gotMax, &gotVerdicts); err != nil {
		t.Fatalf("the mirror does not open as a record on its own: %v", err)
	}
	if gotN != wantN || gotMax != wantMax || gotVerdicts != wantVerdicts || wantVerdicts != 1 {
		t.Errorf("mirror holds %d events (max id %d, %d verdicts); the live record holds %d (max id %d, %d verdicts)",
			gotN, gotMax, gotVerdicts, wantN, wantMax, wantVerdicts)
	}
}
