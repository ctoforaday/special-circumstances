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

// A PATH A SEAT NAMES IS NEVER THE RECORD. prove reads its --script and ingest its blue/report.md
// inside the process whose cached connection holds the record's locks; a read of record.db there
// closes a descriptor on it and releases them, and the act then writes through a connection that
// believes it still holds them. Every spelling of the record is refused before the read — the
// path itself, its log, a symlink, a hard link — and the record keeps its lock.
func TestAPathASeatNamesIsRefusedWhenItIsTheRecord(t *testing.T) {
	const sentence = "Nine is composite, so the protocol rejects a false claim."
	cases := map[string]func(t *testing.T, runDir, db string) string{
		"the database by path": func(t *testing.T, runDir, db string) string { return db },
		"its write-ahead log":  func(t *testing.T, runDir, db string) string { return db + "-wal" },
		"a symlink to it": func(t *testing.T, runDir, db string) string {
			if err := os.Symlink(db, filepath.Join(runDir, "nine.py")); err != nil {
				t.Skipf("no symlinks here: %v", err)
			}
			return "nine.py"
		},
		"a hard link to it": func(t *testing.T, runDir, db string) string {
			if err := os.Link(db, filepath.Join(runDir, "nine.js")); err != nil {
				t.Skipf("no hard links here: %v", err)
			}
			return "nine.js"
		},
	}
	for name, path := range cases {
		t.Run(name, func(t *testing.T) {
			runDir := newRun(t)
			seat := proveSeat(t, runDir, "# H\n\n"+sentence+"\n")
			db := filepath.Join(runtest.Open(t, runDir).Records(), "record.db")
			if _, err := os.Stat(db + "-wal"); err != nil {
				t.Fatalf("the open record has no write-ahead log to name: %v", err)
			}
			script := path(t, runDir, db)
			before := countType(t, runDir, recordpb.EventType_EVENT_TYPE_PROOF)

			out, err := run(t, "prove", "--run", runDir, "--seat-id", seat, "--quote", sentence,
				"--script", script, "--reason", "trial division settles it")
			if err == nil || !strings.Contains(err.Error(), "this run's record") {
				t.Errorf("prove --script %s was not refused as the record: err=%v\n%s", script, err, out)
			}
			if n, measured := recordtest.PosixLocksHeld(t, db); measured && n == 0 {
				t.Errorf("after the refused prove, this process holds no POSIX lock on record.db — the script was read")
			}
			if after := countType(t, runDir, recordpb.EventType_EVENT_TYPE_PROOF); after != before {
				t.Errorf("a refused prove recorded %d proof(s)", after-before)
			}
		})
	}
}

func TestIngestRefusesAReportThatIsTheRecord(t *testing.T) {
	runDir := newRun(t)
	db := filepath.Join(runtest.Open(t, runDir).Records(), "record.db")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-synthesize"); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(db, filepath.Join(runDir, "blue", "report.md")); err != nil {
		t.Skipf("no symlinks here: %v", err)
	}
	out, err := run(t, "ingest", "--run", runDir, "--seat-id", "blue-synthesize")
	if err == nil || !strings.Contains(err.Error(), "this run's record") {
		t.Errorf("ingest of a report linked to the record was not refused as the record: err=%v\n%s", err, out)
	}
	if n, measured := recordtest.PosixLocksHeld(t, db); measured && n == 0 {
		t.Errorf("after the refused ingest, this process holds no POSIX lock on record.db — the report was read")
	}
}
