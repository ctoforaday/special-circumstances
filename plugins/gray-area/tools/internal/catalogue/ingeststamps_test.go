package catalogue

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// THE INGEST BOOKKEEPING, TESTED WHERE IT IS WRITTEN (#949).
//
// ingested_first is when the catalogue first read a session, and a later read never moves it;
// ingested_last is the latest read. Nothing asserted either until now — the only watcher was a
// telecli golden that compared them for EQUALITY, which is not a property of the store (IngestFile
// stamps the wall clock once per file, and a big session spans seconds), and which flipped whenever a
// loaded machine let one session's files straddle a second.
//
// No sleep and no injected clock: the stamps are set into the past first, so "the re-read moved
// last and left first alone" is visible however fast the machine is.
func TestAReReadKeepsTheFirstStampAndMovesTheLast(t *testing.T) {
	root := corpus(t)
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	target := sessionFile(t, root)
	if _, err := IngestFile(db, target); err != nil {
		t.Fatal(err)
	}
	const past = 100 // 1970: before any clock this test can run on
	if _, err := db.Exec(`UPDATE session SET ingested_first = ?, ingested_last = ? WHERE session_id = ?`, past, past, target.SessionID); err != nil {
		t.Fatal(err)
	}

	extra := `{"uuid":"a8","parentUuid":"u1","timestamp":"2026-09-08T10:02:00Z","message":{"role":"assistant","content":[{"type":"text","text":"more"}]}}` + "\n"
	f, err := os.OpenFile(target.Path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(extra); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, err := IngestFile(db, target); err != nil {
		t.Fatal(err)
	}

	var first, last int64
	if err := db.QueryRow(`SELECT ingested_first, ingested_last FROM session WHERE session_id = ?`, target.SessionID).Scan(&first, &last); err != nil {
		t.Fatal(err)
	}
	if first != past {
		t.Errorf("ingested_first = %d, want %d: a later read moved the first-read stamp", first, past)
	}
	if last <= past {
		t.Errorf("ingested_last = %d: a read of new bytes did not move the last-read stamp", last)
	}
}

// Across every file of a session, read in one pass, the first stamp is never after the last —
// the invariant the golden now asserts, pinned here against the real multi-file fixture.
func TestTheFirstStampIsNeverAfterTheLast(t *testing.T) {
	root := corpus(t)
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files, err := TranscriptFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 2 {
		t.Fatalf("setup: the fixture should hold a session and its subagents, got %d file(s)", len(files))
	}
	for _, f := range files {
		if _, err := IngestFile(db, f); err != nil {
			t.Fatal(err)
		}
	}
	var bad int
	if err := db.QueryRow(`SELECT count(*) FROM session WHERE ingested_first > ingested_last`).Scan(&bad); err != nil {
		t.Fatal(err)
	}
	if bad != 0 {
		t.Errorf("%d session(s) have a first read after their last", bad)
	}
}
