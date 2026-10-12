package catalogue

import (
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func store(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
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

// RETENTION RUNS ON TRANSCRIPT TIME, session rows included. One session per way the rule can be
// got wrong; each row's name says what it is.
//
//	OLD         records past the window, read long ago            goes
//	NEW         records inside it                                  stays
//	BACKFILLED  records past the window, READ A MOMENT AGO         goes — keyed on when the store
//	            read it, this row lived a month with nothing under it
//	SPOKE       no acts, one word inside the window                stays
//	REGISTERED  no records at all, registered inside the window    stays
//	FORGOTTEN   no records at all, registered past it              goes
//	RESUMED     records past the window, and LIVE                  stays — registration reopened it
//	            a moment before this pass, and deleting it loses the cloud id
func TestRetentionRunsOnTranscriptTime(t *testing.T) {
	db := store(t)
	now := time.Now()
	old := now.Add(-40 * 24 * time.Hour).Unix()
	fresh := now.Add(-1 * time.Hour).Unix()
	session := func(id string, ingested int64) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last) VALUES(?,'/p',?,?)`, id, ingested, ingested); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO file_offset(path,session_id,agent_id,offset,size,head_sha) VALUES(?,?,'',10,10,'x')`, "/p/"+id+".jsonl", id); err != nil {
			t.Fatal(err)
		}
	}
	act := func(id string, ts int64) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO act(session_id,seq,ts,tool,outcome) VALUES(?,0,?,'Bash','ok')`, id, ts); err != nil {
			t.Fatal(err)
		}
	}
	session("OLD", old)
	act("OLD", old)
	session("NEW", fresh)
	act("NEW", fresh)
	session("BACKFILLED", fresh)
	act("BACKFILLED", old)
	session("SPOKE", old)
	if _, err := db.Exec(`INSERT INTO word(session_id,ts,role,text,source) VALUES('SPOKE',?,'user','hello','transcript')`, fresh); err != nil {
		t.Fatal(err)
	}
	session("REGISTERED", fresh)
	session("FORGOTTEN", old)
	session("RESUMED", old)
	act("RESUMED", old)

	if _, err := Retain(db, DefaultLimits(), now, map[string]bool{"RESUMED": true}); err != nil {
		t.Fatal(err)
	}
	ids := func(q string) string {
		t.Helper()
		rows, err := db.Query(q)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var id string
			rows.Scan(&id)
			out = append(out, id)
		}
		return strings.Join(out, ",")
	}
	const want = "NEW,REGISTERED,RESUMED,SPOKE"
	if got := ids(`SELECT session_id FROM session ORDER BY 1`); got != want {
		t.Errorf("sessions kept = %s, want %s", got, want)
	}
	// An offset row goes with its session row and never before it: left behind it queues the
	// expired session in Closure for ever, and `find` counts its path for ever.
	if got := ids(`SELECT session_id FROM file_offset ORDER BY 1`); got != want {
		t.Errorf("offset rows kept = %s, want %s", got, want)
	}
	// Records go by their own time whatever became of the row: RESUMED keeps its row, not its act.
	if got := ids(`SELECT session_id FROM act ORDER BY 1`); got != "NEW" {
		t.Errorf("acts kept = %s, want NEW", got)
	}
}

// MarkGone RECORDS A TRANSCRIPT THE DISK NO LONGER HOLDS, once, and clears it when the file is
// back. The rows read from the file are untouched: they are what is left of it.
func TestMarkGoneRecordsWhatTheDiskNoLongerHolds(t *testing.T) {
	root := corpus(t)
	db := store(t)
	files, _ := TranscriptFiles(root)
	if len(files) < 2 {
		t.Fatalf("the fixture holds %d transcripts, want at least 2", len(files))
	}
	mustIngest(t, db, files...)
	goneAt := func(path string) sql.NullInt64 {
		t.Helper()
		var at sql.NullInt64
		if err := db.QueryRow(`SELECT gone_at FROM v_transcript WHERE path = ?`, path).Scan(&at); err != nil {
			t.Fatal(err)
		}
		return at
	}
	first := time.Unix(1_800_000_000, 0)
	if n, err := MarkGone(db, first); err != nil || n != 0 {
		t.Fatalf("every transcript on disk: %d gone, err %v — want 0", n, err)
	}

	victim := files[0].Path
	body, err := os.ReadFile(victim)
	if err != nil {
		t.Fatal(err)
	}
	var acts int
	db.QueryRow(`SELECT count(*) FROM act`).Scan(&acts)
	if err := os.Remove(victim); err != nil {
		t.Fatal(err)
	}
	if n, err := MarkGone(db, first); err != nil || n != 1 {
		t.Fatalf("one transcript removed: %d gone, err %v — want 1", n, err)
	}
	if at := goneAt(victim); !at.Valid || at.Int64 != first.Unix() {
		t.Errorf("gone_at = %+v, want %d", at, first.Unix())
	}
	if at := goneAt(files[1].Path); at.Valid {
		t.Errorf("a transcript still on disk is marked gone at %d", at.Int64)
	}
	// A LATER LOOK KEEPS THE FIRST NOTICE: gone_at is when the store found out, not the last
	// time it checked.
	if n, err := MarkGone(db, first.Add(24*time.Hour)); err != nil || n != 1 {
		t.Fatalf("second look: %d gone, err %v — want 1", n, err)
	}
	if at := goneAt(victim); at.Int64 != first.Unix() {
		t.Errorf("a second look moved gone_at to %d, want %d kept", at.Int64, first.Unix())
	}
	var after int
	db.QueryRow(`SELECT count(*) FROM act`).Scan(&after)
	if after != acts {
		t.Errorf("marking a transcript gone changed the act count from %d to %d", acts, after)
	}

	// The file is back — a corpus restored from a copy — and the mark goes.
	if err := os.WriteFile(victim, body, 0o644); err != nil {
		t.Fatal(err)
	}
	if n, err := MarkGone(db, first.Add(48*time.Hour)); err != nil || n != 0 {
		t.Fatalf("file restored: %d gone, err %v — want 0", n, err)
	}
	if at := goneAt(victim); at.Valid {
		t.Errorf("a restored transcript is still marked gone at %d", at.Int64)
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
	db, err := Open(p, io.Discard)
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
