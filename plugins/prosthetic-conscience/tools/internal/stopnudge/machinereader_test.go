package stopnudge

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/machinereader"
)

// noteAndTranscript builds the with-note session the band gate emits on: 40 assistant turns
// against a note written an hour before. Returned as (project dir, transcript path).
func noteAndTranscript(t *testing.T) (string, string) {
	t.Helper()
	dir := withNote(t, "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. x\n")
	tp := filepath.Join(dir, "t.jsonl")
	var lines []string
	for range 40 {
		lines = append(lines, `{"type":"assistant","timestamp":"2026-08-23T00:30:00Z","message":{"usage":{"input_tokens":1,"cache_read_input_tokens":10,"cache_creation_input_tokens":0}}}`)
	}
	if err := os.WriteFile(tp, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, tp
}

// THE #1025 CASE. A session whose final message is contracted to a machine reader gets nothing:
// the nudge is a sentence addressed to a person, and a caller waiting for one JSON object cannot
// parse the answer a seat sends back to it.
//
// BOTH PATHS ARE DRIVEN — the with-note band gate and the no-note context gate — because they are
// separate decisions with separate state, and a check at only one of them leaves the other
// injecting. Each arm asserts the SAME session emits without the marker, so a green here cannot
// come from a fixture that was silent anyway.
func TestAContractedFinalMessageIsNeverInterrupted(t *testing.T) {
	th := Thresholds{
		TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120,
		ContextNotice: 150_000, ContextWarn: 300_000, ContextUrgent: 600_000,
	}
	for _, c := range []struct {
		name    string
		payload func(t *testing.T) (dir, in string)
	}{
		{"note past a band", func(t *testing.T) (string, string) {
			dir, tp := noteAndTranscript(t)
			return dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tp})
		}},
		{"no note, heavy context", func(t *testing.T) (string, string) {
			dir := t.TempDir()
			return dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir,
				"transcript_path": transcriptAt(t, 200_000)})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir, in := c.payload(t)
			if out, _ := driveMarked(t, dir, in, "", th); out == "" {
				t.Fatalf("the fixture emits nothing WITHOUT the marker — this arm proves nothing")
			}

			dir, in = c.payload(t) // a fresh session: the first drive spent the band
			out, errb := driveMarked(t, dir, in, "1", th)
			if out != "" {
				t.Errorf("stdout was written into a contracted final message:\n%s", out)
			}
			if errb != "" {
				t.Errorf("stderr was written: %q", errb)
			}
			if _, err := os.Stat(StatePath(dir)); err == nil {
				t.Error("state was written: the hook did not decline before touching the session's files")
			}
		})
	}
}

// PRESENCE, NOT TRUTHINESS, AND THE EXIT CODE. Every non-empty value declines, "0" included —
// a launcher computing the value sets the EMPTY string to mean no. The code is 0 throughout: a
// declining hook is not a failing one, and a non-zero exit would put a hook error in front of the
// reader this whole change exists to keep clear.
func TestEveryNonEmptyValueDeclinesAndTheHookStillExitsZero(t *testing.T) {
	dir, tp := noteAndTranscript(t)
	in := payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tp})
	for _, v := range []string{"1", "0", "true", "false", "seat"} {
		var o, e bytes.Buffer
		code := run(nil, strings.NewReader(in), &o, &e, dir, v,
			time.Date(2026, 8, 23, 1, 0, 0, 0, time.UTC),
			Thresholds{TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120})
		if code != 0 {
			t.Errorf("%s=%q exited %d, want 0", machinereader.Var, v, code)
		}
		if o.Len() != 0 || e.Len() != 0 {
			t.Errorf("%s=%q wrote stdout %q stderr %q", machinereader.Var, v, o.String(), e.String())
		}
	}
}

// THE UNMARKED SESSION IS UNCHANGED, asserted as the same fixture the arm above declines on. The
// nudge reaches the AGENT channel — additionalContext, not systemMessage — because the agent is
// who has to write the note; gblock's ruling keeps that in every session but a contracted one.
func TestWithoutTheMarkerTheNudgeStillReachesTheAgent(t *testing.T) {
	dir, tp := noteAndTranscript(t)
	in := payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tp})
	out, _ := driveMarked(t, dir, in, "", Thresholds{TurnsNotice: 30, TurnsWarn: 60, TurnsUrgent: 120})
	if out == "" {
		t.Fatal("an unmarked session past a band emitted nothing")
	}
	if !strings.Contains(out, `"additionalContext"`) {
		t.Errorf("the nudge did not travel on the agent's channel:\n%s", out)
	}
	if _, err := os.Stat(StatePath(dir)); err != nil {
		t.Errorf("an unmarked session wrote no band state: %v", err)
	}
}

// MAIN READS THE PUBLISHED SPELLING. run() takes the value, so nothing else in this package would
// notice if Main read a different variable — which is the whole of the launcher's side of the
// contract.
func TestMainReadsTheMarkerFromThePublishedVariable(t *testing.T) {
	t.Setenv(machinereader.Var, "1")
	dir, tp := noteAndTranscript(t)
	t.Setenv("CLAUDE_PROJECT_DIR", dir)
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	// os.Args UNDER `go test` CARRIES THE TEST BINARY'S OWN FLAGS, and Main hands them to the
	// preamble, which treats an unknown flag as a finished invocation. Without this the whole test
	// passes on a flag-parse failure — it did, and said PASS while Main never reached the payload.
	prevArgs := os.Args
	os.Args = []string{"sc-stop"}
	t.Cleanup(func() { os.Args = prevArgs })

	stdin := fileHolding(t, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tp}))
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	prevIn, prevOut := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdin, w
	code := Main()
	os.Stdin, os.Stdout = prevIn, prevOut
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	var got bytes.Buffer
	if _, err := got.ReadFrom(r); err != nil {
		t.Fatal(err)
	}

	if code != 0 {
		t.Errorf("Main exited %d", code)
	}
	if got.Len() != 0 {
		t.Errorf("Main emitted with %s set:\n%s", machinereader.Var, got.String())
	}
	if _, err := os.Stat(StatePath(dir)); err == nil {
		t.Error("Main wrote band state with the marker set")
	}
}

// fileHolding is stdin as a FILE rather than a pipe: Main reads os.Stdin to EOF, and a pipe
// whose writer is still open never reaches one.
func fileHolding(t *testing.T, body string) *os.File {
	t.Helper()
	p := filepath.Join(t.TempDir(), "stdin.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f
}
