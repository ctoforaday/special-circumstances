package catalogue

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func store(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "c.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// CLOSURE INGESTS BEFORE IT DECIDES. A session's final turn is never hook-ingested — there is no
// later Stop — but the transcript does receive it (406 of 409 completed transcripts hold it). A
// sweep that only marked sessions shut would leave that text on disk unread.
func TestClosureIngestsBeforeMarkingShut(t *testing.T) {
	root := corpus(t)
	db := store(t)
	files, _ := TranscriptFiles(root)
	// A session enters the queue only once the catalogue holds an offset for it — a session it
	// has never seen is `backfill`'s job, which is the bound that keeps closure from facing the
	// whole 202 MB of non-live history on its first run. Seed one file, then APPEND: closure must
	// pick up the appended turn, which is the "final turn no Stop will ingest" case.
	seed := sessionFile(t, root)
	if _, err := IngestFile(db, seed); err != nil {
		t.Fatal(err)
	}
	extra := `{"uuid":"fin","parentUuid":"u1","timestamp":"2026-09-08T10:05:00Z","message":{"role":"assistant","content":[{"type":"text","text":"the final turn"}]}}` + "\n"
	fh, _ := os.OpenFile(seed.Path, os.O_APPEND|os.O_WRONLY, 0o600)
	fh.WriteString(extra)
	fh.Close()

	st, err := Closure(db, files, map[string]bool{}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if st.SessionsShut != 1 {
		t.Fatalf("sessions shut = %d, want 1", st.SessionsShut)
	}
	var n int
	db.QueryRow(`SELECT count(*) FROM word WHERE text='the final turn'`).Scan(&n)
	if n != 1 {
		t.Errorf("the final turn is stored %d times, want once — closure marked the session shut "+
			"without first reading the bytes no Stop will ever ingest", n)
	}
	var closed sql.NullInt64
	db.QueryRow(`SELECT closed_at FROM session WHERE session_id='S'`).Scan(&closed)
	if !closed.Valid {
		t.Error("closed_at was not set, so the session stays in the queue forever")
	}
}

// A LIVE session is never closed: its bytes are still arriving.
func TestClosureSkipsLiveSessions(t *testing.T) {
	root := corpus(t)
	db := store(t)
	files, _ := TranscriptFiles(root)
	for _, f := range files { // give it an offset record so it enters the queue at all
		IngestFile(db, f)
	}
	st, err := Closure(db, files, map[string]bool{"S": true}, DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if st.SessionsShut != 0 {
		t.Errorf("closed %d live sessions, want 0", st.SessionsShut)
	}
}

// THE CAP IS ON FILES AND BYTES, not sessions: offsets are per file, and the eight largest
// non-live sessions on the measured corpus hold 94 files, so a per-session cap bounds nothing.
// The remainder must be DEFERRED and reported, never dropped.
func TestSweepCapsFilesAndDefersTheRest(t *testing.T) {
	root := corpus(t)
	db := store(t)
	files, _ := TranscriptFiles(root)
	for _, f := range files {
		IngestFile(db, f) // give every file an offset so the session is in the queue
	}
	lim := DefaultLimits()
	lim.MaxFiles = 1 // one file's worth per invocation

	st, err := Closure(db, files, map[string]bool{}, lim)
	if err != nil {
		t.Fatal(err)
	}
	if st.FilesRead > 3 {
		t.Errorf("read %d files under a 1-file cap", st.FilesRead)
	}
	if st.SessionsShut != 1 {
		t.Errorf("sessions shut = %d; the fixture holds one session", st.SessionsShut)
	}
	// A second invocation must be a no-op rather than re-reading.
	st2, _ := Closure(db, files, map[string]bool{}, lim)
	if st2.SessionsShut != 0 {
		t.Errorf("a closed session was revisited: %d shut on the second pass", st2.SessionsShut)
	}
}

// Retention deletes past the window and leaves everything inside it.
func TestRetentionDeletesOnlyPastTheWindow(t *testing.T) {
	db := store(t)
	now := time.Now()
	old := now.Add(-40 * 24 * time.Hour).Unix()
	fresh := now.Add(-1 * time.Hour).Unix()
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('OLD','/p',?,?)`, old, old)
	db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES('NEW','/p',?,?)`, fresh, fresh)
	db.Exec(`INSERT INTO act(session_id,seq,ts,tool,outcome) VALUES('OLD',0,?,'Bash','ok')`, old)
	db.Exec(`INSERT INTO act(session_id,seq,ts,tool,outcome) VALUES('NEW',0,?,'Bash','ok')`, fresh)

	if _, err := Retain(db, DefaultLimits(), now); err != nil {
		t.Fatal(err)
	}
	var sessions, acts int
	db.QueryRow(`SELECT count(*) FROM session`).Scan(&sessions)
	db.QueryRow(`SELECT count(*) FROM act`).Scan(&acts)
	if sessions != 1 || acts != 1 {
		t.Errorf("after retention: %d sessions / %d acts, want 1/1 — only the expired row goes", sessions, acts)
	}
	var left string
	db.QueryRow(`SELECT session_id FROM act`).Scan(&left)
	if left != "NEW" {
		t.Errorf("retention kept %q, want NEW", left)
	}
}

// The switchboard's first question, end to end.
func TestAgentsReportsLivenessAndActivity(t *testing.T) {
	root := corpus(t)
	db := store(t)
	files, _ := TranscriptFiles(root)
	for _, f := range files {
		IngestFile(db, f)
	}
	sessions := t.TempDir()
	writeSessionFile(t, sessions, "1.json", `{"pid":999999,"sessionId":"S","cwd":"/w","procStart":"1","pidDomain":"`+LocalPidDomain()+`"}`)

	got, err := Agents(context.Background(), db, sessions)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("agents = %d, want 1", len(got))
	}
	if got[0].Acts == 0 {
		t.Error("the agent row carries no activity, so 'who is working on what' cannot be answered")
	}
	// pid 999999 will not exist, and its procStart cannot match: Ended, never Live.
	if got[0].Liveness == Live {
		t.Error("a nonexistent pid reported as live")
	}
}

// "Is anyone else editing this file" — matched on a suffix so a caller need not know the
// absolute path every agent used.
func TestTouchedFindsBySuffix(t *testing.T) {
	db := store(t)
	db.Exec(`INSERT INTO act(session_id,agent_id,seq,ts,tool,target,outcome)
	         VALUES('S','',0,100,'Edit','/abs/repo/pkg/thing.go','ok')`)
	got, err := Touched(context.Background(), db, "pkg/thing.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].SessionID != "S" {
		t.Fatalf("touched = %+v, want the one session", got)
	}
}

// NULL must render as the word, not as an empty string: they are different answers.
func TestRawQueryRendersNullDistinctly(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.db")
	db, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	db.Exec(`INSERT INTO word(session_id,ts,role,text,source,provisional) VALUES('S',1,'assistant','hi','transcript',0)`)
	db.Close()

	ro, err := OpenRead(p)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	cols, rows, err := RawQuery(context.Background(), ro, `SELECT prompt_id, text FROM v_word`, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 2 || len(rows) != 1 {
		t.Fatalf("cols=%v rows=%v", cols, rows)
	}
	if rows[0][0] != "NULL" {
		t.Errorf("a NULL prompt_id rendered as %q; empty and NULL must not read alike", rows[0][0])
	}
}

func writeSessionFile(t *testing.T, dir, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// manySessions builds N single-file sessions, which is what the cap actually has to bound: one
// session cannot exercise it, because the loop runs once whatever the limit says.
func manySessions(t *testing.T, n int) string {
	t.Helper()
	root := t.TempDir()
	proj := filepath.Join(root, "-home-u-work")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		p := filepath.Join(proj, "S"+string(rune('a'+i))+".jsonl")
		if err := os.WriteFile(p, []byte(lines(userRec, bashOK, bashRes)), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// THE CAP BOUNDS ONE INVOCATION AND DEFERS THE REST, and the deferral drains over successive
// runs rather than being dropped. A single-session fixture cannot see this: a mutation disabling
// the cap entirely survived TestSweepCapsFilesAndDefersTheRest, because with one session the loop
// runs once either way.
func TestCapBoundsAnInvocationAndTheBacklogDrains(t *testing.T) {
	const sessions = 10
	root := manySessions(t, sessions)
	db := store(t)
	files, _ := TranscriptFiles(root)
	if len(files) != sessions {
		t.Fatalf("fixture built %d files, want %d", len(files), sessions)
	}
	for _, f := range files {
		if _, err := IngestFile(db, f); err != nil {
			t.Fatal(err)
		}
	}

	lim := DefaultLimits()
	lim.MaxFiles = 3

	st, err := Closure(db, files, map[string]bool{}, lim)
	if err != nil {
		t.Fatal(err)
	}
	if st.SessionsShut > 4 {
		t.Errorf("one invocation shut %d sessions under a 3-file cap — the cap is not bounding it", st.SessionsShut)
	}
	if st.Deferred == 0 {
		t.Error("nothing was reported deferred, so a caller cannot tell the backlog from an empty queue")
	}

	// The remainder must DRAIN over successive invocations, never be dropped.
	runs := 1
	for {
		s, err := Closure(db, files, map[string]bool{}, lim)
		if err != nil {
			t.Fatal(err)
		}
		if s.SessionsShut == 0 {
			break
		}
		runs++
		if runs > sessions+2 {
			t.Fatal("the queue is not draining")
		}
	}
	var open int
	db.QueryRow(`SELECT count(*) FROM session WHERE closed_at IS NULL`).Scan(&open)
	if open != 0 {
		t.Errorf("%d sessions never closed after %d invocations — deferred became dropped", open, runs)
	}
	if runs < 2 {
		t.Errorf("the whole backlog cleared in one invocation; the cap did not bind")
	}
}
