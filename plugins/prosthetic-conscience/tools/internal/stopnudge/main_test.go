package stopnudge

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
)

// payload builds hook stdin by MARSHALLING, never by concatenation.
//
// A path embedded in a JSON string literal is not a path on Windows: C:\Users\...\t.jsonl
// carries \U and \t, and \t is a TAB. The transcript path then arrives corrupted, the file
// is not found, and the read comes back unmeasured — which is what CI caught, on a test
// asserting turns_measured. json.Marshal escapes it correctly on every platform.
func payload(t *testing.T, kv map[string]any) string {
	t.Helper()
	b, err := json.Marshal(kv)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func withNote(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func drive(t *testing.T, dir, payload string) (stdout, stderr string) {
	t.Helper()
	return driveWith(t, dir, payload, configured())
}

func driveWith(t *testing.T, dir, payload string, th Thresholds) (stdout, stderr string) {
	t.Helper()
	var o, e bytes.Buffer
	run(strings.NewReader(payload), &o, &e, dir, time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC), th)
	return o.String(), e.String()
}

// newest is what bands re-arm against, so it decides whether a nudge repeats. It takes
// the LATER of the two timestamps, because a re-affirmation is an answer exactly as a
// rewrite is — the skill clause calls a reasoned "still accurate" valid, and a band that
// ignored it would nag a session that had complied.
func TestNewestTakesTheLaterOfWrittenAndReaffirmed(t *testing.T) {
	for _, tc := range []struct {
		name, body, want string
	}{
		{"reaffirmed is later", "---\nwritten_at: 2026-08-23T01:00:00Z\nreaffirmed_at: 2026-08-23T05:00:00Z\n---\n", "2026-08-23T05:00:00Z"},
		{"written is later", "---\nwritten_at: 2026-08-23T09:00:00Z\nreaffirmed_at: 2026-08-23T05:00:00Z\n---\n", "2026-08-23T09:00:00Z"},
		{"only written", "---\nwritten_at: 2026-08-23T09:00:00Z\n---\n", "2026-08-23T09:00:00Z"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := newest(checkpoint.Parse(tc.body))
			if got.Format(time.RFC3339) != tc.want {
				t.Errorf("newest = %s, want %s", got.Format(time.RFC3339), tc.want)
			}
		})
	}
}

// A note with no parseable timestamp yields the zero time, which Decide treats as
// "never answered" rather than "answered at the epoch". The distinction matters: an
// epoch answer would make every band look permanently re-armed.
func TestNewestOnAnUnparseableNoteIsZero(t *testing.T) {
	for _, body := range []string{
		"---\nschema: 2\nupdated: 2026-08-23T01:00:00Z\n---\n", // schema 2: no written_at at all
		"---\nwritten_at: not-a-time\n---\n",
		"no frontmatter here",
	} {
		if got := newest(checkpoint.Parse(body)); !got.IsZero() {
			t.Errorf("newest(%q) = %v, want the zero time", body, got)
		}
	}
}

// No note means nothing whose age could be reported. Silence, and no state file: the
// seal record reads the nudge's liveness from whether that file exists.
func TestRunWithNoNoteSaysNothingAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	out, errb := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir}))
	if out != "" {
		t.Errorf("emitted with no note present:\n%s", out)
	}
	if strings.Contains(errb, "panic") {
		t.Errorf("stderr: %s", errb)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "nudge.json")); err == nil {
		t.Error("created nudge.json with no note")
	}
}

// THE CURRENT SHIPPED STATE, asserted rather than assumed: thresholds do not exist yet,
// so the nudge emits nothing AND leaves no trace. If this ever starts failing, either
// bands were configured (which Phase 2 does deliberately) or the inert path started
// writing — and the second would silently mark every Phase 1 baseline row as
// nudge_enabled, disarming criterion 6.
func TestRunIsInertAndLeavesNoTraceWhileThresholdsAreUnset(t *testing.T) {
	dir := withNote(t, "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. x\n")
	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "stop_hook_active": false}))
	if out != "" {
		t.Errorf("an unconfigured nudge emitted:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "nudge.json")); err == nil {
		t.Error("an unconfigured nudge wrote state; nudge_enabled would read true for every baseline row")
	}
	if configured().Configured() {
		t.Error("configured() now reports thresholds — this test's premise no longer holds, and the " +
			"inert-path assertions above stopped meaning anything")
	}
}

// A hook is fed whatever the client sends, including nothing. It must not fail the event.
func TestRunSurvivesAnUnusablePayload(t *testing.T) {
	dir := withNote(t, "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n")
	for _, payload := range []string{"", "{", "null", `{"session_id":123}`} {
		out, errb := drive(t, dir, payload)
		if out != "" {
			t.Errorf("payload %q produced output: %s", payload, out)
		}
		if strings.Contains(errb, "panic") {
			t.Errorf("payload %q panicked: %s", payload, errb)
		}
	}
}

// THE EMIT PATH, end to end: payload in, hookSpecificOutput out. Until thresholds are
// configured this branch is unreachable in production, which is exactly why it needs a
// test — otherwise its first execution ever is its first real emission.
//
// The response SHAPE is the assertion. hookEventName and additionalContext are the two
// fields measured to make an injection arrive as a hook_additional_context attachment
// (spike §8/§13); a wrong shape is silently ignored by the client, so nothing downstream
// would report it.
func TestTheEmitPathProducesTheMeasuredResponseShape(t *testing.T) {
	dir := withNote(t, "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. x\n")
	tp := filepath.Join(dir, "t.jsonl")
	var lines []string
	for i := range 40 { // well past the floor of 20 turns
		lines = append(lines, `{"type":"assistant","timestamp":"2026-08-23T00:30:00Z","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":10,"cache_creation_input_tokens":0}}}`)
		_ = i
	}
	if err := os.WriteFile(tp, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	in := payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tp, "stop_hook_active": false})
	out, _ := driveWith(t, dir, in, Thresholds{TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120})
	if out == "" {
		t.Fatal("no emission from a note 40 turns old with a NOTICE edge at 30")
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("response is not JSON: %v\n%s", err, out)
	}
	hso, ok := got["hookSpecificOutput"].(map[string]any)
	if !ok {
		t.Fatalf("no hookSpecificOutput in the response:\n%s", out)
	}
	if hso["hookEventName"] != "Stop" {
		t.Errorf("hookEventName = %v, want Stop — the client keys on this", hso["hookEventName"])
	}
	ctx, _ := hso["additionalContext"].(string)
	if ctx == "" {
		t.Error("additionalContext empty; the injection would carry nothing")
	}
	if len(ctx) > 200 {
		t.Errorf("additionalContext is %d bytes, over the 200-byte budget:\n%s", len(ctx), ctx)
	}
	if strings.Contains(ctx, "%") {
		t.Errorf("emitted a percentage:\n%s", ctx)
	}
	// And it recorded the band BEFORE returning: a second boundary must be silent.
	if again, _ := driveWith(t, dir, in, Thresholds{TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120}); again != "" {
		t.Errorf("emitted twice for one band — the write did not happen before the emit:\n%s", again)
	}
}
