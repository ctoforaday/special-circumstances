package capture

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// ARCHIVING READS THE RECORD'S BYTES WITHOUT TAKING THE PROCESS'S LOCKS WITH THEM.
//
// Capture keeps reading the record after it archives it — the seat-turn ingest, the audits. A read
// of a record file while this process's connection is open closes a descriptor on it and with it
// every POSIX lock the process holds on that file; the connection keeps answering and never
// re-takes a lock it believes it holds. The -shm locks are POSIX on every platform, so the
// dead-man-switch lock there is the one such a read drops; the SHARED lock on record.db is an OFD
// lock on Linux, which no stray close drops, and POSIX elsewhere. So the act after the archive is
// read through the record path and the kernel's table is asked whether that read holds both.
//
// And the archived database is the whole record on its own: the rows the connection wrote, with
// no log beside it.
func TestArchivingTheRecordLeavesTheNextReadHoldingItsLock(t *testing.T) {
	repo := t.TempDir()
	run := filepath.Join(repo, "research", "2026-10-25_locks")
	recs := filepath.Join(run, "records")
	if err := os.MkdirAll(recs, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(recs, "record.db")
	db, err := recordsql.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = recordsql.CloseUnder(run) })
	if _, err := db.Exec(`CREATE TABLE ev (id INTEGER PRIMARY KEY, body TEXT)`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 400; i++ {
		if _, err := db.Exec(`INSERT INTO ev (body) VALUES (?)`, "padding so the log grows past one page"); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := ArchiveRecord(runtest.Open(t, run), repo); err != nil {
		t.Fatal(err)
	}

	after, err := recordsql.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	var live int
	if err := after.QueryRow(`SELECT COUNT(*) FROM ev`).Scan(&live); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{dbPath, dbPath + "-shm"} {
		if n, measured := recordtest.LocksHeld(t, f); measured && n == 0 {
			t.Errorf("a read through the record after archiving holds no lock on %s — the archive closed a "+
				"descriptor on it beside this process's open connection", filepath.Base(f))
		}
	}

	got := archivedRows(t, filepath.Join(repo, "run-archive", "2026-10-25_locks.tar.gz"))
	if got != live || live != 400 {
		t.Errorf("the archived database holds %d rows on its own; the live record holds %d (wrote 400)", got, live)
	}
}

// archivedRows extracts records/record.db ALONE from the archive and counts its rows.
func archivedRows(t *testing.T, archive string) int {
	t.Helper()
	f, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "record.db")
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			t.Fatalf("the archive carries no records/record.db: %v", err)
		}
		if h.Name != "records/record.db" {
			continue
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(out, b, 0o644); err != nil {
			t.Fatal(err)
		}
		break
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(out)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM ev`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
