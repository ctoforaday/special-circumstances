package postcompactobserve

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/unwritable"
)

const note = `---
schema: 3
session_id: sess-abc
---
## Validation loop
1. go test ./...  → all packages ok  · re-armed by: any .go edit
## Next intended steps
1. wire the hooks into hooks.json
## In-flight handles
- background task bg-77
`

func withNote(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func call(t *testing.T, dir, stdin string, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(stdin), &o, &e, dir,
		time.Date(2026, 7, 29, 5, 0, 0, 0, time.UTC))
	return o.String(), e.String(), code
}

func rows(t *testing.T, dir string) []observation {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", "compaction-observations.jsonl"))
	if err != nil {
		return nil
	}
	var out []observation
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		if line == "" {
			continue
		}
		var o observation
		if err := json.Unmarshal([]byte(line), &o); err != nil {
			t.Fatalf("row is not JSON: %v\n%s", err, line)
		}
		out = append(out, o)
	}
	return out
}

// The companion to the restore regression: this hook must never inject. It
// cannot — PostCompact has no additionalContext — but a future edit reaching for
// stdout would send text to the HUMAN while reading as a restore in review.
func TestNeverWritesToStdout(t *testing.T) {
	dir := withNote(t, note)
	stdout, _, code := call(t, dir, `{"trigger":"auto","compact_summary":"the session ran go test","session_id":"s1"}`)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if stdout != "" {
		t.Errorf("PostCompact stdout reaches the user, not the model — emitted %q", stdout)
	}
}

// A summary that reused a section's vocabulary scores higher than one that did
// not. This is the whole measurement, so it is asserted as a comparison rather
// than against a magic threshold.
func TestOverlapDistinguishesKeptFromDropped(t *testing.T) {
	kept := overlap(note, "The session established a validation loop: go test across all packages, re-armed by any edit.")
	dropped := overlap(note, "The user asked about the weather and we discussed nothing else whatsoever.")

	find := func(rs []sectionOverlap, h string) sectionOverlap {
		for _, r := range rs {
			if r.Heading == h {
				return r
			}
		}
		t.Fatalf("section %q not scored", h)
		return sectionOverlap{}
	}
	k := find(kept, "Validation loop")
	d := find(dropped, "Validation loop")
	if !(k.Ratio > d.Ratio) {
		t.Errorf("kept ratio %.2f not above dropped %.2f — the probe measures nothing", k.Ratio, d.Ratio)
	}
	if d.Ratio != 0 {
		t.Errorf("unrelated summary scored %.2f, want 0", d.Ratio)
	}
	if k.Tokens == 0 {
		t.Error("counts must travel with the ratio; 0 tokens makes the ratio meaningless")
	}
}

// Counts travel with the ratio, and every scored section carries both.
func TestRowCarriesCountsNotJustRatios(t *testing.T) {
	dir := withNote(t, note)
	call(t, dir, `{"trigger":"manual","compact_summary":"go test packages","session_id":"s1","agent_id":"a9"}`)
	got := rows(t, dir)
	if len(got) != 1 {
		t.Fatalf("rows = %d, want 1", len(got))
	}
	r := got[0]
	if r.Probe != probe {
		t.Errorf("probe = %q — an unlabelled row can be mistaken for a reading", r.Probe)
	}
	if r.SessionID != "s1" || r.AgentID != "a9" || r.Trigger != "manual" {
		t.Errorf("row lost its identity: %+v", r)
	}
	if r.SummaryB != len("go test packages") {
		t.Errorf("summary_bytes = %d", r.SummaryB)
	}
	if len(r.Sections) == 0 {
		t.Fatal("no sections scored")
	}
	for _, s := range r.Sections {
		if s.Tokens == 0 {
			t.Errorf("section %q scored with 0 tokens", s.Heading)
		}
	}
}

// Append-only: thrash means this fires repeatedly, and each boundary is a row.
func TestRowsAccumulate(t *testing.T) {
	dir := withNote(t, note)
	for range 3 {
		call(t, dir, `{"trigger":"auto","compact_summary":"x","session_id":"s1"}`)
	}
	if got := rows(t, dir); len(got) != 3 {
		t.Errorf("rows = %d, want 3 — observations must accumulate across boundaries", len(got))
	}
}

// Empty sections are skipped, not scored as absent: "nothing to measure" and
// "measured as absent" are different facts.
func TestEmptySectionsAreNotScored(t *testing.T) {
	got := overlap("---\nschema: 3\n---\n## Validation loop\n\n## Next intended steps\n1. do the thing properly\n", "unrelated")
	for _, s := range got {
		if s.Heading == "Validation loop" {
			t.Error("scored a section with no content")
		}
	}
	if len(got) != 1 {
		t.Errorf("scored %d sections, want 1", len(got))
	}
}

func TestNeverBlocks(t *testing.T) {
	cases := []struct{ name, dir, stdin string }{
		{"no project dir", "", `{"trigger":"auto"}`},
		{"no note", t.TempDir(), `{"trigger":"auto"}`},
		{"malformed stdin", withNote(t, note), `{not json`},
		{"empty stdin", withNote(t, note), ``},
		{"no summary", withNote(t, note), `{"trigger":"auto"}`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, _, code := call(t, c.dir, c.stdin); code != 0 {
				t.Errorf("exit %d, want 0", code)
			}
		})
	}
}

func TestVersion(t *testing.T) {
	stdout, _, code := call(t, t.TempDir(), ``, "-version")
	if code != 0 || !strings.Contains(stdout, "sc-postcompact-observe") {
		t.Errorf("version: code=%d stdout=%q", code, stdout)
	}
}

// The compaction row carries the same three age figures as the seal row, and for a
// different question: the seal asks how stale the note was at a boundary, this asks
// what the note looked like when the summary was built. Criterion 6's falsification
// reads nudge_enabled from here.
func TestObservationCarriesTheNotesAge(t *testing.T) {
	dir := t.TempDir()
	note := "---\nschema: 3\nwritten_at: 2026-08-23T00:10:00Z\n---\n## Validation loop\n1. x\n"
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	tpath := filepath.Join(dir, "t.jsonl")
	lines := `{"type":"assistant","timestamp":"2026-08-23T00:05:00Z","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":10,"cache_creation_input_tokens":0}}}
{"type":"assistant","timestamp":"2026-08-23T00:20:00Z","message":{"usage":{"input_tokens":2,"cache_read_input_tokens":40,"cache_creation_input_tokens":0}}}
`
	if err := os.WriteFile(tpath, []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}

	// MARSHALLED, not concatenated: a Windows path in a JSON string literal turns \t into a
	// TAB, the transcript is not found, and turns_measured comes back false. CI caught
	// exactly that on this test.
	inb, err := json.Marshal(map[string]any{
		"session_id": "s1", "transcript_path": tpath, "trigger": "auto", "compact_summary": "a summary",
	})
	if err != nil {
		t.Fatal(err)
	}
	in := string(inb)
	var out, errb bytes.Buffer
	run(nil, strings.NewReader(in), &out, &errb, dir, time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC))

	b, err := os.ReadFile(filepath.Join(dir, ".claude", "checkpoints", "compaction-observations.jsonl"))
	if err != nil {
		t.Fatalf("no observation row written: %v (stderr: %s)", err, errb.String())
	}
	var row map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(b), &row); err != nil {
		t.Fatal(err)
	}
	if row["note_age_turns"] != float64(1) {
		t.Errorf("note_age_turns = %v, want 1", row["note_age_turns"])
	}
	if row["turns_measured"] != true {
		t.Errorf("turns_measured = %v", row["turns_measured"])
	}
	// Phase 1 ships no nudge, and the falsification compares rows written with it live
	// against rows written without. A row that does not say which it was cannot be put
	// in either population.
	if row["nudge_enabled"] != false {
		t.Errorf("nudge_enabled = %v, want false in Phase 1", row["nudge_enabled"])
	}
}

// A PostCompact hook must ALWAYS exit 0, whatever it finds. It runs at the moment a
// session has just lost most of its context; a non-zero exit there risks wedging the
// compaction itself, and the observation this hook exists to record is worth strictly
// less than the compaction completing.
//
// Every failure shape it can actually meet, asserted on the exit code rather than on the
// absence of a panic — "did not crash" and "reported success" are different claims.
func TestTheObserverAlwaysExitsZero(t *testing.T) {
	noteDir := func(t *testing.T) string {
		t.Helper()
		d := t.TempDir()
		if err := os.MkdirAll(filepath.Join(d, ".claude", "checkpoints"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, ".claude", "checkpoints", "CHECKPOINT.md"),
			[]byte("---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		return d
	}

	for _, tc := range []struct {
		name    string
		dir     func(*testing.T) string
		payload string
	}{
		{"no note at all", func(t *testing.T) string { return t.TempDir() }, `{"session_id":"s1","trigger":"auto"}`},
		{"empty payload", noteDir, ``},
		{"malformed payload", noteDir, `{`},
		{"note present, transcript missing", noteDir, `{"session_id":"s1","transcript_path":"/nope/none.jsonl","trigger":"auto"}`},
		// PROBED, NOT ASSUMED — this case asserts exit 0, which a successful write also
		// produces, so where the chmod does not restrict the caller it passed while exercising
		// nothing. Measured green in a root container against a writable directory.
		{"unwritable checkpoints dir", func(t *testing.T) string {
			d := noteDir(t)
			unwritable.Dir(t, filepath.Join(d, ".claude", "checkpoints"))
			return d
		}, `{"session_id":"s1","trigger":"auto","compact_summary":"x"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.dir(t)
			var out, errb bytes.Buffer
			code := run(nil, strings.NewReader(tc.payload), &out, &errb,
				dir, time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC))
			if code != 0 {
				t.Errorf("exit = %d, want 0 — a failing PostCompact hook risks the compaction itself\nstderr: %s", code, errb.String())
			}
		})
	}
}

// A compaction with no note writes NO row. The baseline is a distribution over notes'
// ages, so a row with no note behind it is not a zero-age observation — it is not an
// observation, and counting it would pull every median toward fresh.
func TestNoNoteMeansNoRow(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	run(nil, strings.NewReader(`{"session_id":"s1","trigger":"auto"}`), &out, &errb,
		dir, time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC))
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "compaction-observations.jsonl")); err == nil {
		t.Error("wrote an observation row with no note to observe")
	}
}
