package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A PATH A SEAT NAMES IS NEVER THE RECORD. compute reads its --dir bundle and ingest its blue/report.md
// inside the process whose cached connection holds the record's locks; a read of a record file
// there closes a descriptor on it and releases the POSIX locks the process holds on that file, and
// the act then writes through a connection that believes it still holds them. The -shm locks are
// POSIX on every platform, so its dead-man-switch lock is the one a read of -shm drops; the
// database file's locks are OFD on Linux and survive a read of record.db, and are POSIX elsewhere.
// Every spelling of the record is refused before the read — the path itself, its log, its shared
// memory, a hard link in the bundle — and the record keeps its locks on both files. A symlink in a
// bundle is refused as a symlink before anything is read, whatever it names.
func TestAPathASeatNamesIsRefusedWhenItIsTheRecord(t *testing.T) {
	const sentence = "Nine is composite, so the protocol rejects a false claim."
	const record = "this run's record"
	// inBundle makes a bundle directory under the run's scratch/ and lays name in it with link.
	inBundle := func(t *testing.T, runDir, db, name string, link func(oldname, newname string) error) string {
		dir := filepath.Join(runDir, "scratch", "nine")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := link(db, filepath.Join(dir, name)); err != nil {
			t.Skipf("no such link here: %v", err)
		}
		return "scratch/nine"
	}
	cases := map[string]struct {
		dir     func(t *testing.T, runDir, db string) string
		command string
		refusal string
	}{
		"the database by path": {func(t *testing.T, runDir, db string) string { return db }, "nine.js", record},
		"its write-ahead log":  {func(t *testing.T, runDir, db string) string { return db + "-wal" }, "nine.js", record},
		"its shared memory":    {func(t *testing.T, runDir, db string) string { return db + "-shm" }, "nine.js", record},
		"a symlink to it in the bundle": {func(t *testing.T, runDir, db string) string {
			return inBundle(t, runDir, db, "nine.py", os.Symlink)
		}, "nine.py", "symlink"},
		"a hard link to it in the bundle": {func(t *testing.T, runDir, db string) string {
			return inBundle(t, runDir, db, "nine.js", os.Link)
		}, "nine.js", record},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			runDir := newRun(t)
			seat := computeSeat(t, runDir, "# H\n\n"+sentence+"\n")
			db := filepath.Join(runtest.Open(t, runDir).Records(), "record.db")
			if _, err := os.Stat(db + "-wal"); err != nil {
				t.Fatalf("the open record has no write-ahead log to name: %v", err)
			}
			dir := c.dir(t, runDir, db)
			interpreter := map[string]string{".js": "node", ".py": "python3"}[filepath.Ext(c.command)]
			before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION)

			out, err := run(t, "compute", "--run", runDir, "--seat-id", seat, "--quote", sentence,
				"--dir", dir, "--reason", "trial division settles it", "--", interpreter, c.command)
			if err == nil || !strings.Contains(err.Error(), c.refusal) {
				t.Errorf("compute --dir %s was not refused as %q: err=%v\n%s", dir, c.refusal, err, out)
			}
			requireRecordLocked(t, db, "the refused compute")
			if after := countType(t, runDir, recordpb.EventType_EVENT_TYPE_COMPUTATION); after != before {
				t.Errorf("a refused compute recorded %d computation(s)", after-before)
			}
		})
	}
}

func TestIngestRefusesAReportThatIsTheRecord(t *testing.T) {
	for _, target := range []string{"record.db", "record.db-shm"} {
		t.Run(target, func(t *testing.T) {
			runDir := newRun(t)
			db := filepath.Join(runtest.Open(t, runDir).Records(), "record.db")
			if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(filepath.Dir(db), target), filepath.Join(runDir, "blue", "report.md")); err != nil {
				t.Skipf("no symlinks here: %v", err)
			}
			out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
			if err == nil || !strings.Contains(err.Error(), "this run's record") {
				t.Errorf("ingest of a report linked to %s was not refused as the record: err=%v\n%s", target, err, out)
			}
			requireRecordLocked(t, db, "the refused ingest")
		})
	}
}

// requireRecordLocked fails when this process, its connection still open, holds no lock on
// record.db or on its -shm. The -shm count is the one a read of a record file can still empty: its
// locks are POSIX. The record.db count holds the database file's own lock — OFD on Linux, which no
// stray close drops, so it empties only if the connection closed or the lock mode is not OFD.
func requireRecordLocked(t *testing.T, db, after string) {
	t.Helper()
	for _, f := range []string{db, db + "-shm"} {
		if n, measured := recordtest.LocksHeld(t, f); measured && n == 0 {
			t.Errorf("after %s, this process holds no lock on %s while its connection is open — the "+
				"named file was read beside the connection", after, filepath.Base(f))
		}
	}
}
