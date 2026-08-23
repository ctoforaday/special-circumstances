package checkpointseal

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const rowNote = "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. go test ./...\n"

// rows reads every seal row written under projectDir. A missing file is zero rows,
// which the caller must distinguish from an unparseable one — that is the whole
// point of the record.
func rows(t *testing.T, projectDir string) []map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(projectDir, ".claude", "checkpoints", "seals.jsonl"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("seal row is not JSON: %v\nline: %s", err, line)
		}
		out = append(out, m)
	}
	return out
}

func sealWith(t *testing.T, in hookInput, event string) (dir string) {
	t.Helper()
	dir = t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), rowNote)
	_, stderr, _ := call(t, input(t, in), dir, time.Date(2026, 8, 23, 1, 2, 3, 0, time.UTC), "-event", event)
	if strings.Contains(stderr, "cannot") {
		t.Fatalf("seal reported a failure: %s", stderr)
	}
	return dir
}

// The record has to exist before anything can be measured over it: the snapshot is
// pruned to keepSnapshots, so a baseline of 20 boundaries cannot be rebuilt from
// snapshots at all.
func TestEverySealAppendsARowNamingItsTrigger(t *testing.T) {
	for _, tc := range []struct{ event, trigger string }{
		{evPreCompact, "precompact"},
		{evSessionEnd, "sessionend"},
		{evSubagentStop, "seat_return"},
	} {
		t.Run(tc.event, func(t *testing.T) {
			dir := sealWith(t, hookInput{SessionID: "s1", HookEventName: tc.event}, tc.event)
			got := rows(t, dir)
			if len(got) != 1 {
				t.Fatalf("%s: got %d rows, want 1 — a trigger that writes no row can never appear in a group_by", tc.event, len(got))
			}
			if got[0]["seal_trigger"] != tc.trigger {
				t.Errorf("seal_trigger = %v, want %q", got[0]["seal_trigger"], tc.trigger)
			}
			if got[0]["session_id"] != "s1" {
				t.Errorf("session_id = %v, want s1", got[0]["session_id"])
			}
			if got[0]["nudge_answered"] != "n/a" {
				t.Errorf("nudge_answered = %v, want n/a — Phase 1 ships no nudge", got[0]["nudge_answered"])
			}
		})
	}
}

// A PreCompact payload carries no background_tasks key at all (spike §12). Writing 0
// there would be a manufactured measurement: the absent case and "no background work"
// would be the same bytes, on the column #506's verdict turns on.
func TestAbsentHandlesAreUnmeasuredRatherThanZero(t *testing.T) {
	dir := sealWith(t, hookInput{SessionID: "s1", HookEventName: evPreCompact}, evPreCompact)
	r := rows(t, dir)[0]
	if r["handles_measured"] != false {
		t.Errorf("handles_measured = %v, want false", r["handles_measured"])
	}
	if _, present := r["live_handles"]; present {
		t.Errorf("live_handles present (%v) on a payload with no background_tasks key; it must be omitted", r["live_handles"])
	}
}

// The discriminating case, and the normal seat return: the key IS there and IS empty.
// A []Task decode maps absent and [] to the same nil and gets this wrong, which would
// silently drop the only trigger able to measure handles.
func TestPresentButEmptyHandlesAreAMeasuredZero(t *testing.T) {
	empty := json.RawMessage(`[]`)
	dir := sealWith(t, hookInput{
		SessionID: "s1", HookEventName: evSubagentStop,
		BackgroundTasks: &empty,
	}, evSubagentStop)
	r := rows(t, dir)[0]
	if r["handles_measured"] != true {
		t.Errorf("handles_measured = %v, want true — the key was present", r["handles_measured"])
	}
	if r["live_handles"] != float64(0) {
		t.Errorf("live_handles = %v, want 0", r["live_handles"])
	}
}

// background_tasks includes SUBAGENTS (spike §12). At a seat_return seal the returning
// seat appears in the parent's own list, so a naive count reads high by exactly one and
// answers a different question from "did this note miss some background work".
func TestLiveHandlesExcludesSubagents(t *testing.T) {
	tasks := json.RawMessage(`[
		{"id":"a","type":"shell","status":"running"},
		{"id":"b","type":"subagent","status":"running","agent_type":"general-purpose"},
		{"id":"c","type":"shell","status":"running"}
	]`)
	crons := json.RawMessage(`[{"id":"cron1"}]`)
	dir := sealWith(t, hookInput{
		SessionID: "s1", HookEventName: evSubagentStop,
		BackgroundTasks: &tasks, SessionCrons: &crons,
	}, evSubagentStop)
	r := rows(t, dir)[0]
	if r["live_handles"] != float64(3) {
		t.Errorf("live_handles = %v, want 3 (2 shells + 1 cron, subagent excluded)", r["live_handles"])
	}
}

// The hash is of the NOTE, not of the stamped snapshot — the stamp carries a timestamp,
// so hashing the snapshot would make every seal differ and the drift check useless.
func TestBodySHAHashesTheNoteNotTheStampedSnapshot(t *testing.T) {
	dir := sealWith(t, hookInput{SessionID: "s1", HookEventName: evPreCompact}, evPreCompact)
	sum := sha256.Sum256([]byte(rowNote))
	want := hex.EncodeToString(sum[:])
	if got := rows(t, dir)[0]["body_sha"]; got != want {
		t.Errorf("body_sha = %v, want %s (sha256 of the note body)", got, want)
	}
}

// Append-only: the baseline is a distribution over boundaries, so a row that replaces
// its predecessor makes the record a gauge of one.
func TestRowsAccumulateAcrossSeals(t *testing.T) {
	dir := t.TempDir()
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), rowNote)
	for i := range 3 {
		call(t, input(t, hookInput{SessionID: "s1", HookEventName: evPreCompact}), dir,
			time.Date(2026, 8, 23, 1, 2, i, 0, time.UTC), "-event", evPreCompact)
	}
	if got := rows(t, dir); len(got) != 3 {
		t.Errorf("got %d rows after 3 seals, want 3", len(got))
	}
}

// Criterion 1: every seal record carries the note's age in three units. Without these
// the baseline is a count of boundaries, and the plan's whole argument — that nothing
// records how stale a seal was — is still true after shipping.
func TestSealRowCarriesTheNotesAge(t *testing.T) {
	dir := t.TempDir()
	// A note written at a known time, and a transcript with turns on both sides of it.
	note := "---\nschema: 3\nwritten_at: 2026-08-23T00:10:00Z\nhead: deadbee\n---\n## Validation loop\n1. x\n"
	writeNote(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), note)

	tpath := filepath.Join(dir, "t.jsonl")
	lines := []string{
		`{"type":"assistant","timestamp":"2026-08-23T00:05:00Z","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":10,"cache_creation_input_tokens":0}}}`,
		`{"type":"assistant","timestamp":"2026-08-23T00:20:00Z","message":{"usage":{"input_tokens":2,"cache_read_input_tokens":40,"cache_creation_input_tokens":0}}}`,
		`{"type":"assistant","timestamp":"2026-08-23T00:30:00Z","message":{"usage":{"input_tokens":3,"cache_read_input_tokens":90,"cache_creation_input_tokens":0}}}`,
	}
	if err := os.WriteFile(tpath, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	call(t, input(t, hookInput{SessionID: "s1", HookEventName: evPreCompact, TranscriptPath: tpath}),
		dir, time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC), "-event", evPreCompact)

	r := rows(t, dir)[0]
	if r["note_age_turns"] != float64(2) {
		t.Errorf("note_age_turns = %v, want 2 (assistant entries at or after written_at)", r["note_age_turns"])
	}
	if r["turns_measured"] != true {
		t.Errorf("turns_measured = %v on a transcript the window covers entirely", r["turns_measured"])
	}
	if r["ceiling_known"] != false {
		t.Errorf("ceiling_known = %v with no compact boundary in the transcript", r["ceiling_known"])
	}
}

// A seal with no transcript to read must still write its row — the trigger, the time
// and the handles are facts the seal itself holds. Dropping the row because one
// measurement failed would lose the boundary from the baseline entirely.
func TestASealWithNoTranscriptStillWritesARowWithTheAgeUnmeasured(t *testing.T) {
	dir := sealWith(t, hookInput{SessionID: "s1", HookEventName: evPreCompact}, evPreCompact)
	r := rows(t, dir)[0]
	if r["seal_trigger"] != "precompact" {
		t.Fatalf("row missing its trigger: %v", r)
	}
	if r["turns_measured"] != false {
		t.Errorf("turns_measured = %v with no transcript", r["turns_measured"])
	}
	if _, present := r["note_age_turns"]; present {
		t.Errorf("note_age_turns present (%v) with nothing to measure it from; it must be omitted", r["note_age_turns"])
	}
}
