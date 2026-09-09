package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/gray-area/tools/internal/catalogue"
)

// fixtureHome builds a fake ~/.claude with one session's transcript in it, and redirects every
// variable the code under test resolves a home directory through.
//
// EVERY variable, and USERPROFILE is the one that matters: os.UserHomeDir reads $HOME on Unix and
// %USERPROFILE% on Windows, so setting HOME alone leaves this test reading the real profile and
// writing the store into the developer's actual state directory. It failed loudly in CI, which was
// lucky — on a developer's own Windows box it would have passed while quietly ingesting their real
// transcripts. A test that writes to the machine's own state is worse than no test.
func fixtureHome(t *testing.T, sessionID string, tools ...string) (home string, storePath string) {
	t.Helper()
	home = t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_STATE_HOME", filepath.Join(home, "state"))
	storePath = filepath.Join(home, "state", "special-circumstances", "catalogue", "catalogue.db")

	var blocks []string
	for i, tl := range tools {
		blocks = append(blocks, `{"type":"tool_use","id":"t`+strconv.Itoa(i)+`","name":"`+tl+
			`","input":{"file_path":"/w/f`+strconv.Itoa(i)+`.go"}}`)
	}
	lines := []string{
		`{"uuid":"u1","promptId":"p1","sessionId":"` + sessionID + `","cwd":"/w","type":"user",` +
			`"timestamp":"2026-09-09T12:00:00Z","message":{"role":"user","content":[{"type":"text","text":"go"}]}}`,
		`{"uuid":"a1","parentUuid":"u1","sessionId":"` + sessionID + `","cwd":"/w","type":"assistant",` +
			`"timestamp":"2026-09-09T12:00:01Z","message":{"role":"assistant","content":[` +
			strings.Join(blocks, ",") + `]}}`,
	}
	p := filepath.Join(home, ".claude", "projects", "-w", sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return home, storePath
}

func actCount(t *testing.T, storePath, sessionID string) int {
	t.Helper()
	db, err := catalogue.OpenRead(storePath)
	if err != nil {
		return -1 // no store at all, which is a different answer from zero acts
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM v_action WHERE session_id = ?`, sessionID).Scan(&n); err != nil {
		t.Fatalf("counting acts: %v", err)
	}
	return n
}

// THE HOOK ACTUALLY FILLS THE STORE.
//
// Every other test here asserts what the hook does NOT do — it does not block, it does not write a
// manifest row. Those all pass equally well against a hook that does nothing at all, which is
// exactly the shape of failure this plugin exists to refuse: the manifest stays correct, the
// session is never slowed, and the catalogue is empty for a reason nobody notices until they query
// it and read the zero as "that session did nothing".
func TestStopIngestsTheSessionIntoTheStore(t *testing.T) {
	sid := "sess-ingest-1"
	_, store := fixtureHome(t, sid, "Read", "Edit", "Bash")

	if n := actCount(t, store, sid); n != -1 {
		t.Fatalf("a store exists before the hook ran (%d acts) — the fixture is not isolated", n)
	}
	dir := t.TempDir()
	in := `{"session_id":"` + sid + `","cwd":` + strconv.Quote(dir) + `}`
	if _, stderr, code := call(t, in, dir, okStat(9), "-event", "Stop"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if n := actCount(t, store, sid); n != 3 {
		t.Fatalf("the store holds %d acts after Stop, want 3 — the hook is wired but not ingesting", n)
	}

	// And a second Stop does not double-count. Every turn fires this hook, so a projection that
	// re-read from zero each time would multiply every session's act count by its turn count.
	if _, _, code := call(t, in, dir, okStat(9), "-event", "Stop"); code != 0 {
		t.Fatalf("second Stop exited %d", code)
	}
	if n := actCount(t, store, sid); n != 3 {
		t.Errorf("a second Stop took the store to %d acts — the stored offset is not being honoured", n)
	}
}

// SessionEnd ingests too, and it is the LAST chance to: after it the session may be archived or
// deleted, and gray-area's data is meant to outlive the transcript by design.
func TestSessionEndIngests(t *testing.T) {
	sid := "sess-ingest-2"
	_, store := fixtureHome(t, sid, "Grep", "Write")
	dir := t.TempDir()
	in := `{"session_id":"` + sid + `","cwd":` + strconv.Quote(dir) + `}`
	if _, stderr, code := call(t, in, dir, okStat(9), "-event", "SessionEnd"); code != 0 {
		t.Fatalf("exit %d: %s", code, stderr)
	}
	if n := actCount(t, store, sid); n != 2 {
		t.Fatalf("the store holds %d acts after SessionEnd, want 2", n)
	}
}

// A hook fires for ONE session and must ingest only that one. Ingesting the whole corpus on every
// turn is the cold read the per-invocation cap exists to avoid, and it would arrive as a pause on
// somebody's Stop.
func TestStopIngestsOnlyItsOwnSession(t *testing.T) {
	sid := "sess-mine"
	home, store := fixtureHome(t, sid, "Read")
	other := filepath.Join(home, ".claude", "projects", "-w", "sess-theirs.jsonl")
	if err := os.WriteFile(other,
		[]byte(`{"uuid":"x1","sessionId":"sess-theirs","cwd":"/w","type":"assistant",`+
			`"timestamp":"2026-09-09T12:00:00Z","message":{"role":"assistant","content":`+
			`[{"type":"tool_use","id":"z","name":"Bash","input":{"command":"ls"}}]}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	in := `{"session_id":"` + sid + `","cwd":` + strconv.Quote(dir) + `}`
	if _, _, code := call(t, in, dir, okStat(9), "-event", "Stop"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if n := actCount(t, store, sid); n != 1 {
		t.Errorf("own session: %d acts, want 1", n)
	}
	if n := actCount(t, store, "sess-theirs"); n != 0 {
		t.Errorf("a Stop for one session ingested another's transcript (%d acts)", n)
	}
}

// SessionStart runs the sweep, and the retention marker must be set exactly once a day. The sweep
// is best-effort, so the assertion is on its EFFECT rather than on an exit code every path returns
// 0 from anyway.
func TestSessionStartRunsTheSweepAndMarksRetention(t *testing.T) {
	sid := "sess-sweep"
	_, store := fixtureHome(t, sid, "Read")
	dir := t.TempDir()
	in := `{"session_id":"` + sid + `","transcript_path":"/t/s.jsonl","cwd":` + strconv.Quote(dir) + `}`
	if _, _, code := call(t, in, dir, okStat(9), "-event", "SessionStart"); code != 0 {
		t.Fatalf("exit %d", code)
	}
	db, err := sql.Open("sqlite", store)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var marked string
	if err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, catalogue.MetaRetainedOn).Scan(&marked); err != nil {
		t.Fatalf("SessionStart did not record that retention ran: %v", err)
	}
	if len(marked) != len("2006-01-02") {
		t.Errorf("the retention marker is %q, which is not a UTC date", marked)
	}
}

// A BROKEN STORE MUST NOT COST A SESSION ITS TURN. This is the contract every other failure path
// in this binary holds, and the catalogue is the newest and largest way to break it.
func TestACorruptStoreDoesNotBlockTheSession(t *testing.T) {
	sid := "sess-corrupt"
	_, store := fixtureHome(t, sid, "Read")
	if err := os.MkdirAll(filepath.Dir(store), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(store, []byte("this is not a SQLite database"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	in := `{"session_id":"` + sid + `","transcript_path":"/t/s.jsonl","cwd":` + strconv.Quote(dir) + `}`
	for _, ev := range []string{"Stop", "SessionEnd", "SessionStart"} {
		_, stderr, code := call(t, in, dir, okStat(9), "-event", ev)
		if code != 0 {
			t.Errorf("%s exited %d against an unreadable store — the hook must degrade, not block", ev, code)
		}
		if !strings.Contains(stderr, "capture unaffected") && !strings.Contains(stderr, "catalogue") {
			t.Errorf("%s failed silently against an unreadable store: %q", ev, stderr)
		}
	}
}
