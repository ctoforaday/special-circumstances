package catalogue

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// lostT0 anchors every fixture time here; the boot is lostT0+7h, 16:00 UTC.
var (
	lostT0   = time.Date(2026, 9, 10, 9, 0, 0, 0, time.UTC)
	lostBoot = lostT0.Add(7 * time.Hour)
)

func jsonRec(t *testing.T, m map[string]any) string {
	t.Helper()
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// promptAt is a human prompt — one word row at ts — carrying mode when mode is not "".
func promptAt(t *testing.T, uuid, cwd string, ts time.Time, mode string) string {
	m := map[string]any{"uuid": uuid, "type": "user", "cwd": cwd, "timestamp": ts.UTC().Format(time.RFC3339),
		"origin": map[string]any{"kind": "human"}, "message": map[string]any{"role": "user", "content": "go on"}}
	if mode != "" {
		m["permissionMode"] = mode
	}
	return jsonRec(t, m)
}

// toolResultAt is a `user` record that is a tool's result, which never carries a mode.
func toolResultAt(t *testing.T, uuid, cwd string, ts time.Time) string {
	return jsonRec(t, map[string]any{"uuid": uuid, "type": "user", "cwd": cwd, "timestamp": ts.UTC().Format(time.RFC3339),
		"message": map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "x"}}}})
}

// writeSession writes <root>/<folder>/<sid>.jsonl.
func writeSession(t *testing.T, root, folder, sid string, recs ...string) TranscriptFile {
	t.Helper()
	p := filepath.Join(root, folder, sid+".jsonl")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(lines(recs...)), 0o600); err != nil {
		t.Fatal(err)
	}
	return TranscriptFile{Path: p, SessionID: sid}
}

func mustIngest(t *testing.T, db *sql.DB, tfs ...TranscriptFile) {
	t.Helper()
	for _, tf := range tfs {
		if _, err := IngestFile(db, tf); err != nil {
			t.Fatal(err)
		}
	}
}

func lostIDs(rep LostReport) []string {
	var out []string
	for _, l := range rep.Lost {
		out = append(out, l.SessionID)
	}
	return out
}

// THE SELECTION RULE, one session per clause, so deleting any clause fails here. Every time is
// transcript time; the boot is 16:00.
func TestFindLostSelectsByTranscriptTimeAndEveryClauseBinds(t *testing.T) {
	root := t.TempDir()
	db := store(t)
	h := func(d time.Duration) time.Time { return lostT0.Add(d) }
	const cwd = "/home/u/work"
	folder := projectKey(cwd)

	a := writeSession(t, root, folder, "A", promptAt(t, "a1", cwd, h(6*time.Hour), "")) // 15:00, the newest pre-boot
	b := writeSession(t, root, folder, "B", promptAt(t, "b1", cwd, h(5*time.Hour), "")) // 14:00, closed 14:30
	c := writeSession(t, root, folder, "C", promptAt(t, "c1", cwd, h(4*time.Hour), "")) // 13:00, closed, then grew
	d := writeSession(t, root, folder, "D", promptAt(t, "d1", cwd, h(8*time.Hour), "")) // 17:00, after the boot
	e := writeSession(t, root, folder, "E", promptAt(t, "e1", cwd, h(6*time.Hour-72*time.Hour), ""))
	f := writeSession(t, root, "-elsewhere", "F", promptAt(t, "f1", cwd, h(3*time.Hour), "")) // closed AFTER the boot
	hh := writeSession(t, root, folder, "H", promptAt(t, "h1", cwd, h(5*time.Hour+30*time.Minute), ""))
	mustIngest(t, db, a, b, c, d, e, f, hh)

	mustExecDB := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatal(err)
		}
	}
	mustExecDB(`UPDATE session SET closed_at = ? WHERE session_id = 'B'`, h(5*time.Hour+30*time.Minute).Unix())
	mustExecDB(`UPDATE session SET closed_at = ? WHERE session_id = 'C'`, h(4*time.Hour+30*time.Minute).Unix())
	mustExecDB(`UPDATE session SET closed_at = ? WHERE session_id = 'F'`, lostBoot.Add(time.Minute).Unix())
	// C RESUMES AFTER ITS CLOSURE: new bytes, read by a later pass, reopen it.
	fh, err := os.OpenFile(c.Path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	fh.WriteString(promptAt(t, "c2", cwd, h(4*time.Hour+45*time.Minute), "") + "\n")
	fh.Close()
	mustIngest(t, db, c)
	// I holds a row NEWER than its closure — a close that preceded its last act did not end it.
	mustExecDB(`INSERT INTO session(session_id,project_dir,ingested_first,ingested_last,closed_at) VALUES('I','',1,1,?)`,
		h(3*time.Hour+30*time.Minute).Unix())
	mustExecDB(`INSERT INTO word(session_id,ts,role,text,source) VALUES('I',?,'user','x','transcript')`, h(3*time.Hour+45*time.Minute).Unix())
	// G registered with no rows yet: its sign of life is when it registered.
	if err := RegisterSession(db, "G", filepath.Join(root, folder), "session_01G", h(2*time.Hour)); err != nil {
		t.Fatal(err)
	}

	var aLast int64
	db.QueryRow(`SELECT ingested_last FROM session WHERE session_id='A'`).Scan(&aLast)
	if aLast <= lostBoot.Unix() {
		t.Fatalf("fixture premise: A was read at %d, which must be AFTER the boot %d — a pre-boot tail read after the boot", aLast, lostBoot.Unix())
	}

	rep, err := FindLost(context.Background(), db, lostBoot.Unix(), 48*time.Hour, map[string]bool{"H": true})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(lostIDs(rep), ","), "A,C,I,F,G"; got != want {
		t.Errorf("lost = %s, want %s (newest first)\n"+
			"  A pre-boot, unclosed, read after the boot; B closed after its act and before the boot;\n"+
			"  C closed then reopened by new bytes; D after the boot; E outside the window; F closed after\n"+
			"  the boot; G registered with no rows; H live; I closed before its last act", got, want)
	}
	if rep.NewestSession != "A" || rep.NewestPreBoot != h(6*time.Hour).Unix() || rep.Sessions != 9 {
		t.Errorf("newest pre-boot = %s at %d over %d sessions, want A at %d over 9",
			rep.NewestSession, rep.NewestPreBoot, rep.Sessions, h(6*time.Hour).Unix())
	}
	by := map[string]LostSession{}
	for _, l := range rep.Lost {
		by[l.SessionID] = l
	}
	if l := by["A"]; !l.Verified || l.CWD != cwd || l.Registered || l.Transcript != a.Path {
		t.Errorf("A = %+v, want a verified %s, not registered, transcript %s", l, cwd, a.Path)
	}
	if l := by["F"]; l.Verified || l.CWD != cwd {
		t.Errorf("F's cwd does not encode to its folder, so it must read unverified: %+v", l)
	}
	if l := by["G"]; !l.Registered || l.LastAt != h(2*time.Hour).Unix() || l.BridgeSessionID != "session_01G" ||
		l.Transcript != filepath.Join(root, folder, "G.jsonl") {
		t.Errorf("G = %+v, want registered at %d with its cloud id and the transcript it will have", l, h(2*time.Hour).Unix())
	}

	// The window is honoured: widened, it reaches E.
	rep, err = FindLost(context.Background(), db, lostBoot.Unix(), 96*time.Hour, map[string]bool{"H": true})
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(lostIDs(rep), ","); got != "A,C,I,F,G,E" {
		t.Errorf("with a 96h window, lost = %s, want E added last", got)
	}
}

// NOTHING BEFORE THE BOOT IS ITS OWN ANSWER, and not a report of every session.
func TestFindLostWithNothingBeforeTheBoot(t *testing.T) {
	root := t.TempDir()
	db := store(t)
	mustIngest(t, db, writeSession(t, root, "-w", "A", promptAt(t, "a1", "/w", lostBoot.Add(time.Hour), "")))
	rep, err := FindLost(context.Background(), db, lostBoot.Unix(), 48*time.Hour, nil)
	if err != nil {
		t.Fatal(err)
	}
	if rep.NewestSession != "" || rep.NewestPreBoot != 0 || len(rep.Lost) != 0 || rep.Sessions != 1 {
		t.Errorf("report = %+v, want no pre-boot activity, nothing lost, one session", rep)
	}
}

// A STORE REBUILT 4 → 5 AND BACKFILLED LISTS A SUPERSET. The rebuild loses closed_at (Accepted
// costs), so a session closure had settled before the boot comes back as a candidate — never the
// other way round.
func TestARebuiltStoreListsASupersetOfTheOriginal(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(t.TempDir(), "catalogue.db")
	db, err := Open(p, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	a := writeSession(t, root, "-w", "A", promptAt(t, "a1", "/w", lostT0.Add(6*time.Hour), ""))
	b := writeSession(t, root, "-w", "B", promptAt(t, "b1", "/w", lostT0.Add(5*time.Hour), ""))
	mustIngest(t, db, a, b)
	if _, err := db.Exec(`UPDATE session SET closed_at = ? WHERE session_id = 'B'`, lostT0.Add(5*time.Hour+30*time.Minute).Unix()); err != nil {
		t.Fatal(err)
	}
	before, err := FindLost(context.Background(), db, lostBoot.Unix(), 48*time.Hour, nil)
	if err != nil {
		t.Fatal(err)
	}
	db.Close()

	raw := rawOpen(t, p)
	mustExec(t, raw, `PRAGMA user_version = 4`)
	raw.Close()
	var notice bytes.Buffer
	db, err = Open(p, &notice)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if !strings.Contains(notice.String(), "was stamped 4, rebuilt empty at shape 5") {
		t.Fatalf("the shape-4 store was not rebuilt: notice %q", notice.String())
	}
	mustIngest(t, db, a, b) // the backfill
	after, err := FindLost(context.Background(), db, lostBoot.Unix(), 48*time.Hour, nil)
	if err != nil {
		t.Fatal(err)
	}
	in := map[string]bool{}
	for _, id := range lostIDs(after) {
		in[id] = true
	}
	for _, id := range lostIDs(before) {
		if !in[id] {
			t.Errorf("%s was lost before the rebuild and not after it", id)
		}
	}
	if got := strings.Join(lostIDs(before), ","); got != "A" {
		t.Errorf("before the rebuild: %s, want A (B was closed before the boot)", got)
	}
	if got := strings.Join(lostIDs(after), ","); got != "A,B" {
		t.Errorf("after the rebuild: %s, want A,B — the rebuild lost B's closure", got)
	}
}

// THE MODE IS THE NEWEST `user` RECORD THAT CARRIES IT — not the newest `user` record, which is
// usually a tool result and carries nothing.
func TestPermissionModeIsTheNewestUserRecordThatCarriesIt(t *testing.T) {
	root := t.TempDir()
	at := func(m int) time.Time { return lostT0.Add(time.Duration(m) * time.Minute) }
	assistant := jsonRec(t, map[string]any{"type": "assistant", "permissionMode": "bypassPermissions", "timestamp": at(9).Format(time.RFC3339),
		"message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "ok"}}}})
	for _, tc := range []struct {
		name string
		recs []string
		want string
	}{
		{"auto then plan", []string{promptAt(t, "1", "/w", at(1), "auto"), promptAt(t, "2", "/w", at(2), "plan")}, "plan"},
		{"a tool result after the prompt", []string{promptAt(t, "1", "/w", at(1), "plan"), toolResultAt(t, "2", "/w", at(2))}, "plan"},
		{"no record carries it", []string{promptAt(t, "1", "/w", at(1), ""), toolResultAt(t, "2", "/w", at(2))}, ""},
		{"an assistant record is not a user record", []string{promptAt(t, "1", "/w", at(1), "auto"), assistant}, "auto"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tf := writeSession(t, root, "-w", strings.ReplaceAll(tc.name, " ", "_"), tc.recs...)
			if got := PermissionMode(tf.Path); got != tc.want {
				t.Errorf("PermissionMode = %q, want %q", got, tc.want)
			}
		})
	}
	if got := PermissionMode(filepath.Join(root, "absent.jsonl")); got != "" {
		t.Errorf("an absent transcript read as mode %q", got)
	}
}

// THE COMMANDS ARE PASTED, so they are asserted as exact strings.
func TestRecoveryCommandsAreExact(t *testing.T) {
	for _, tc := range []struct {
		name, cloud, mode       string
		wantReattach, wantLocal string
	}{
		{"cloud and a mode", "cse_X", "auto",
			"claude remote-control --session-id cse_X", "claude --resume S --permission-mode auto --remote-control"},
		{"a recorded default", "cse_X", "default",
			"claude remote-control --session-id cse_X", "claude --resume S --permission-mode manual --remote-control"},
		{"no cloud id", "", "plan", "", "claude --resume S --permission-mode plan --remote-control"},
		{"no mode", "", "", "", "claude --resume S --remote-control"},
	} {
		r, l := RecoveryCommands("S", tc.cloud, tc.mode)
		if r != tc.wantReattach || l != tc.wantLocal {
			t.Errorf("%s:\n  reattach %q\n  want     %q\n  local    %q\n  want     %q", tc.name, r, tc.wantReattach, l, tc.wantLocal)
		}
	}
	if got := ModeLine(""); got != "MODE unknown — the resume starts in whatever mode the CLI chooses" {
		t.Errorf("the unknown MODE row = %q", got)
	}
	if got := ModeLine("plan"); got != "MODE plan" {
		t.Errorf("ModeLine(plan) = %q", got)
	}
	for in, want := range map[string]string{"session_X": "cse_X", "": "", "cse_X": "", "session_": ""} {
		if got := CloudID(in); got != want {
			t.Errorf("CloudID(%q) = %q, want %q", in, got, want)
		}
	}
}
