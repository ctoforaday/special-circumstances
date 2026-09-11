package catalogue

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// corpus builds a synthesized project tree carrying all THREE tiers of the vendor's layout.
func corpus(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	proj := filepath.Join(root, "-home-u-work")
	mk := func(p, body string) {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	body := lines(userRec, bashOK, bashRes)
	mk(filepath.Join(proj, "S.jsonl"), body)
	mk(filepath.Join(proj, "S", "subagents", "agent-a1.jsonl"), body)
	mk(filepath.Join(proj, "S", "subagents", "workflows", "wf_x", "agent-a2.jsonl"), body)
	return root
}

// THE THIRD TIER IS THE ONE THAT BITES. A rule naming only <sid>.jsonl and subagents/agent-*
// leaves the workflow seats unattributed — on the real corpus that was 56 files, all under one
// live session, whose turns would go uncatalogued while the queue read empty.
func TestAttributionCoversAllThreeTiers(t *testing.T) {
	got, err := TranscriptFiles(corpus(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("attributed %d files, want 3 — a missing tier reads as an empty queue, not an error", len(got))
	}
	var agents []string
	for _, f := range got {
		if f.SessionID != "S" {
			t.Errorf("%s attributed to session %q, want S", f.Path, f.SessionID)
		}
		agents = append(agents, f.AgentID)
	}
	sort.Strings(agents)
	want := []string{"", "a1", "a2"}
	for i := range want {
		if agents[i] != want[i] {
			t.Fatalf("agent ids = %v, want %v — the workflow seat is the one a two-tier rule drops", agents, want)
		}
	}
}

// sessionFile picks the session's OWN transcript by name. An earlier draft took files[0] and
// silently got the agent file — WalkDir visits the "S" directory before "S.jsonl" — so a test
// asserting on agent_id=” looked at rows that were never going to be there. The failure read
// like a bug in reprojection and was a bug in the test.
func sessionFile(t *testing.T, root string) TranscriptFile {
	t.Helper()
	files, err := TranscriptFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.AgentID == "" {
			return f
		}
	}
	t.Fatal("no session-level transcript in the fixture")
	return TranscriptFile{}
}

// A second pass over unchanged bytes must add nothing. This is what makes `backfill` safe to
// re-run and what a crash mid-sweep relies on.
func TestIngestIsIdempotent(t *testing.T) {
	root := corpus(t)
	db, err := Open(filepath.Join(t.TempDir(), "catalogue.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files, _ := TranscriptFiles(root)

	count := func() (int, int) {
		var a, w int
		db.QueryRow(`SELECT count(*) FROM act`).Scan(&a)
		db.QueryRow(`SELECT count(*) FROM word`).Scan(&w)
		return a, w
	}
	for _, f := range files {
		if _, err := IngestFile(db, f); err != nil {
			t.Fatal(err)
		}
	}
	a1, w1 := count()
	if a1 == 0 || w1 == 0 {
		t.Fatalf("first pass stored nothing: %d acts, %d words", a1, w1)
	}
	for _, f := range files {
		r, err := IngestFile(db, f)
		if err != nil {
			t.Fatal(err)
		}
		if r.BytesRead != 0 {
			t.Errorf("%s re-read %d bytes on an unchanged file", f.Path, r.BytesRead)
		}
	}
	if a2, w2 := count(); a2 != a1 || w2 != w1 {
		t.Errorf("second pass changed the store: %d/%d acts/words, want %d/%d", a2, w2, a1, w1)
	}
}

// APPENDING is resumed from the offset, not re-read.
func TestAppendIsResumedNotReread(t *testing.T) {
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
	extra := `{"uuid":"a7","parentUuid":"u1","timestamp":"2026-09-08T10:01:00Z","message":{"role":"assistant","content":[{"type":"text","text":"appended"}]}}` + "\n"
	f, _ := os.OpenFile(target.Path, os.O_APPEND|os.O_WRONLY, 0o600)
	f.WriteString(extra)
	f.Close()

	r, err := IngestFile(db, target)
	if err != nil {
		t.Fatal(err)
	}
	if r.Reprojected {
		t.Error("an APPEND triggered a full reprojection; the head fingerprint should be unchanged")
	}
	if r.BytesRead != int64(len(extra)) {
		t.Errorf("read %d bytes, want exactly the %d appended", r.BytesRead, len(extra))
	}
	var n int
	db.QueryRow(`SELECT count(*) FROM word WHERE text='appended'`).Scan(&n)
	if n != 1 {
		t.Errorf("the appended text was stored %d times, want once", n)
	}
}

// A REWRITTEN file is reprojected from zero rather than resumed into the middle of a record. The
// head sha is what tells the two apart; a size check alone cannot, because a rewrite can be the
// same length.
func TestRewrittenFileIsReprojectedFromZero(t *testing.T) {
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
	var before int
	db.QueryRow(`SELECT count(*) FROM act`).Scan(&before)

	// Same session, entirely different content — the shape a resumed or forked transcript takes.
	os.WriteFile(target.Path, []byte(lines(
		`{"uuid":"z1","type":"user","promptId":"P9","sessionId":"S","timestamp":"2026-09-08T11:00:00Z","message":{"role":"user","content":"different"}}`,
		`{"uuid":"z2","parentUuid":"z1","timestamp":"2026-09-08T11:00:01Z","message":{"role":"assistant","content":[{"type":"tool_use","id":"q1","name":"Write","input":{"file_path":"/n"}}]}}`,
		`{"uuid":"z3","parentUuid":"z2","timestamp":"2026-09-08T11:00:02Z","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"q1"}]}}`)), 0o600)

	r, err := IngestFile(db, target)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Reprojected {
		t.Fatal("a rewritten file was resumed from its stale offset instead of reprojected")
	}
	var tools string
	db.QueryRow(`SELECT group_concat(tool) FROM act WHERE agent_id=''`).Scan(&tools)
	if tools != "Write" {
		t.Errorf("acts after reprojection = %q, want only the new content — the old rows survived", tools)
	}
	_ = before
}

// A transcript that vanishes mid-sweep is DATA, not a failure: it must not cost the hook its
// exit code, because a manifest that cannot be written must never cost a session its turn.
func TestAVanishedTranscriptIsNotAnError(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := IngestFile(db, TranscriptFile{Path: filepath.Join(t.TempDir(), "gone.jsonl"), SessionID: "S"}); err != nil {
		t.Errorf("a missing transcript returned an error: %v", err)
	}
}

// A SAME-LENGTH rewrite is the case only the fingerprint can catch, and the case a size check
// silently misses. TestRewrittenFileIsReprojectedFromZero does not reach it: its replacement
// happens to be shorter, so `st.Size() < offset` fires and the sha is never consulted — a
// mutation disabling the sha comparison survived that test and is killed by this one.
func TestSameLengthRewriteIsDetectedByTheFingerprint(t *testing.T) {
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
	original, err := os.ReadFile(target.Path)
	if err != nil {
		t.Fatal(err)
	}
	// Same byte count, different content: flip the tool name inside the existing bytes.
	rewritten := bytes.Replace(original, []byte(`"name":"Bash"`), []byte(`"name":"Xash"`), 1)
	if len(rewritten) != len(original) {
		t.Fatalf("fixture is not the same length: %d vs %d", len(rewritten), len(original))
	}
	if bytes.Equal(rewritten, original) {
		t.Fatal("fixture was not actually rewritten")
	}
	if err := os.WriteFile(target.Path, rewritten, 0o600); err != nil {
		t.Fatal(err)
	}

	r, err := IngestFile(db, target)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Reprojected {
		t.Fatal("a same-length rewrite was resumed from its stale offset — only the head sha can see this, " +
			"and a size check cannot")
	}
	var tools string
	db.QueryRow(`SELECT group_concat(tool) FROM act WHERE agent_id=''`).Scan(&tools)
	if tools != "Xash" {
		t.Errorf("acts after reprojection = %q, want only the rewritten content", tools)
	}
}

// THE STORED ROLE IS WHO SPOKE. The same no-origin prompt is the human's at top level and the
// lead's in a seat's transcript, so this is asserted through IngestFile — the one place that knows
// the tier — and not through Project.
func TestStoredRoleIsTheSpeaker(t *testing.T) {
	root := t.TempDir()
	proj := filepath.Join(root, "-home-u-work")
	top := lines(
		`{"uuid":"h1","type":"user","origin":{"kind":"human"},"timestamp":"2026-09-08T10:00:00Z","message":{"role":"user","content":"human prompt"}}`,
		`{"uuid":"p1","type":"user","isMeta":true,"origin":{"kind":"peer"},"timestamp":"2026-09-08T10:00:01Z","message":{"role":"user","content":"Another Claude session sent a message: peer turn"}}`,
		`{"uuid":"m1","type":"user","isMeta":true,"timestamp":"2026-09-08T10:00:02Z","message":{"role":"user","content":[{"type":"text","text":"meta record"}]}}`,
		`{"uuid":"n1","type":"user","origin":{"kind":"task-notification"},"timestamp":"2026-09-08T10:00:03Z","message":{"role":"user","content":"notification turn"}}`,
		`{"uuid":"q1","type":"attachment","timestamp":"2026-09-08T10:00:04Z","attachment":{"type":"queued_command","prompt":"queued prompt","commandMode":"prompt","origin":{"kind":"human"}}}`,
		`{"uuid":"a1","parentUuid":"h1","timestamp":"2026-09-08T10:00:05Z","message":{"role":"assistant","content":[{"type":"text","text":"assistant reply"}]}}`)
	seat := lines(
		`{"uuid":"s1","type":"user","timestamp":"2026-09-08T10:00:00Z","message":{"role":"user","content":"seat prompt"}}`)
	for p, body := range map[string]string{
		filepath.Join(proj, "S.jsonl"):                          top,
		filepath.Join(proj, "S", "subagents", "agent-a1.jsonl"): seat,
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files, _ := TranscriptFiles(root)
	for _, f := range files {
		if _, err := IngestFile(db, f); err != nil {
			t.Fatal(err)
		}
	}
	for text, want := range map[string]string{
		"human prompt": "user",
		"Another Claude session sent a message: peer turn": "peer",
		"meta record":       "harness",
		"notification turn": "notification",
		"seat prompt":       "lead",
		"assistant reply":   "assistant",
	} {
		var got string
		if err := db.QueryRow(`SELECT role FROM v_word WHERE text = ?`, text).Scan(&got); err != nil {
			t.Errorf("%q: %v", text, err)
			continue
		}
		if got != want {
			t.Errorf("%q stored role %q, want %q", text, got, want)
		}
	}
	// The word tier stores turns; a mid-turn delivery is `find`'s to report, never a word row.
	var n int
	db.QueryRow(`SELECT count(*) FROM v_word WHERE text = 'queued prompt'`).Scan(&n)
	if n != 0 {
		t.Errorf("a queued_command attachment was stored as %d word row(s)", n)
	}
}

// The session row's project_dir must name the SESSION's directory, not whichever file was
// ingested last. A session with subagents has most of its files under <sid>/subagents/, so
// taking the last one made the row report the subagents directory — a false field that reads
// as plausible, which is the shape this plugin exists to refuse.
func TestProjectDirComesFromTheSessionTranscript(t *testing.T) {
	root := corpus(t)
	db, err := Open(filepath.Join(t.TempDir(), "c.db"), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	files, _ := TranscriptFiles(root)
	// Ingest the session's own transcript FIRST, then the agent files, so a naive
	// last-write-wins would overwrite the correct value with the subagents directory.
	sess := sessionFile(t, root)
	if _, err := IngestFile(db, sess); err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.AgentID != "" {
			if _, err := IngestFile(db, f); err != nil {
				t.Fatal(err)
			}
		}
	}
	var dir string
	db.QueryRow(`SELECT project_dir FROM session WHERE session_id='S'`).Scan(&dir)
	if want := filepath.Dir(sess.Path); dir != want {
		t.Errorf("project_dir = %q, want %q — an agent file overwrote it", dir, want)
	}
}
