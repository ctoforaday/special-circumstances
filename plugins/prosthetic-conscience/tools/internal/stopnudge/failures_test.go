package stopnudge

import (
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/unwritable"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func responseMessage(t *testing.T, stdout string) (context, message string) {
	t.Helper()
	if strings.TrimSpace(stdout) == "" {
		return "", ""
	}
	var out hookOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("stdout is not one response: %v (%q)", err, stdout)
	}
	return out.HookSpecificOutput.AdditionalContext, out.SystemMessage
}

// AN UNREADABLE STATE FILE turns the nudge off for the session, and that said nothing on any channel:
// Decide failed closed — correctly — and the hook exited 0 in silence. Stop displays a systemMessage,
// so it is said on the same turn.
func TestAnUnreadableNudgeStateIsSaidOnTheSameTurn(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	// The state path as a DIRECTORY: it exists and can never be read as a state file.
	if err := os.MkdirAll(statePath(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	tr := transcriptAt(t, 400_000) // over a context band edge, so the nudge has something to decide
	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tr}))
	ctx, msg := responseMessage(t, out)
	if ctx != "" {
		t.Errorf("the nudge spoke although it could not read its own state: %q", ctx)
	}
	if !strings.Contains(msg, "- "+string(StageStateRead)+" in ") {
		t.Fatalf("an unreadable nudge state was not said to the human: %q", out)
	}
}

// AN UNWRITABLE STATE FILE DISABLES THE NUDGE PERMANENTLY: every later turn re-reads a record that
// never records anything. Write-before-emit means nothing reaches the agent — right — and until now
// nothing reached anyone.
func TestAnUnwritableNudgeStateIsSaidOnTheSameTurn(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	parent := filepath.Dir(statePath(dir))
	if err := os.MkdirAll(parent, 0o755); err != nil {
		t.Fatal(err)
	}
	// Absent state, UNWRITABLE parent: the read succeeds (nothing there) and the save cannot. The
	// helper verifies the restriction took, and skips rather than passing when it did not.
	unwritable.Dir(t, parent)

	tr := transcriptAt(t, 400_000)
	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": tr}))
	ctx, msg := responseMessage(t, out)
	if ctx != "" {
		t.Errorf("the nudge spoke although its state could not be written — write-before-emit is broken: %q", ctx)
	}
	if !strings.Contains(msg, "- "+string(StageStateWrite)+" in ") {
		t.Fatalf("an unwritable nudge state was not said to the human: %q", out)
	}
}

// THE WITH-NOTE PATH has the same two failures at its own call sites, and they are separate code:
// Decide is not DecideContext. Driven at Decide with a reading past the floor, because building a
// transcript that crosses the floor through run() tests freshness, not this.
func TestDecideRecordsItsOwnStateFailuresAndClearsThem(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		dir := t.TempDir()
		if err := os.MkdirAll(statePath(dir), 0o755); err != nil { // the state FILE as a directory
			t.Fatal(err)
		}
		rec := hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", at(10), io.Discard)
		if d := Decide(rec, dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
			t.Errorf("spoke with an unreadable state: %q", d.Emit)
		}
		if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageStateRead)+" in ") {
			t.Fatalf("an unreadable state on the with-note path was not said: %q", msg)
		}
		if err := os.Remove(statePath(dir)); err != nil {
			t.Fatal(err)
		}
		rec = hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", at(30), io.Discard)
		Decide(rec, dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(30))
		rec.Settle()
		if onRecord(t, StageStateRead) {
			t.Error("a readable state left the entry on the record")
		}
	})
	t.Run("write", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", t.TempDir())
		dir := t.TempDir()
		parent := filepath.Dir(statePath(dir))
		if err := os.MkdirAll(parent, 0o755); err != nil {
			t.Fatal(err)
		}
		unwritable.Dir(t, parent)
		rec := hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", at(10), io.Discard)
		if d := Decide(rec, dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(10)); d.Emit != "" {
			t.Errorf("spoke although the band could not be recorded: %q", d.Emit)
		}
		if msg := rec.Settle(); !strings.Contains(msg, "- "+string(StageStateWrite)+" in ") {
			t.Fatalf("an unwritable state on the with-note path was not said: %q", msg)
		}
		if err := os.Chmod(parent, 0o755); err != nil {
			t.Fatal(err)
		}
		rec = hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", at(30), io.Discard)
		if d := Decide(rec, dir, "s1", false, stale(), "CHECKPOINT.md", at(1), bands(), at(30)); d.Emit == "" {
			t.Fatal("with the state writable again the nudge should speak")
		}
		rec.Settle()
		if onRecord(t, StageStateWrite) {
			t.Error("a written state left the entry on the record")
		}
	})
}

// THE RE-ARM SAVE. A context that drops back under an edge re-arms the band it had spent, and that
// is a SAVE on a path that speaks nothing to the agent — so its failure was invisible twice over.
func TestAContextRearmThatCannotBeSavedIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	heavy := payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": transcriptAt(t, 400_000)})
	if out, _ := drive(t, dir, heavy); out == "" {
		t.Fatal("the first heavy turn should spend a band")
	}
	unwritable.Dir(t, filepath.Dir(statePath(dir)))
	light := payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": transcriptAt(t, 20_000)})
	out, _ := drive(t, dir, light)
	if _, msg := responseMessage(t, out); !strings.Contains(msg, "- "+string(StageStateWrite)+" in ") {
		t.Fatalf("a re-arm that could not be saved was not said: %q", out)
	}
	if err := os.Chmod(filepath.Dir(statePath(dir)), 0o755); err != nil {
		t.Fatal(err)
	}
	drive(t, dir, light)
	if onRecord(t, StageStateWrite) {
		t.Error("a saved re-arm left the entry on the record")
	}
}

// THE NOTE IS THERE AND sc-stop CANNOT READ IT. NotePath stat'd it, the read failed, and the hook
// returned 0 in silence — exactly like a session with no note.
func TestStopRecordsANoteItCannotRead(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := withNote(t, "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n")
	note := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	if err := os.Chmod(note, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(note, 0o644) })
	if _, err := os.ReadFile(note); err == nil {
		t.Skip("chmod 000 did not stop this process reading the note — the premise never took")
	}
	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir}))
	if _, msg := responseMessage(t, out); !strings.Contains(msg, "- "+string(StageNoteRead)+" in ") {
		t.Fatalf("an unreadable note was not said: %q", out)
	}
	if err := os.Chmod(note, 0o644); err != nil {
		t.Fatal(err)
	}
	drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir}))
	if onRecord(t, StageNoteRead) {
		t.Error("a readable note left the entry on the record")
	}
}

// NOT REACHED BY ANY TEST, and why: the StageEncode call in emitResponse. json.Encoder cannot fail
// on hookOutput — two strings in a struct, written to an *os.File — short of a closed stdout, which
// the client never hands a hook. The call exists so that if the response ever grows a field that
// CAN fail to encode, the failure has somewhere to go. The same holds for sc-sessionstart's and
// sc-strike-counter's encode stage.

// STOP IS THE CARRIER. A seal on SessionEnd — which displays nothing — fails; the next Stop, with
// nothing of its own to nudge about, answers with the failure alone. That is the whole reason the
// record exists, driven through the real binary's response.
func TestStopCarriesAFailureRecordedWhereNothingCouldBeSaid(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	silent := hookfailures.New("prosthetic-conscience", "sc-sessionend", "SessionEnd", time.Now(), io.Discard)
	silent.Fail("snapshot-write", "sc-sessionend: cannot write snapshot: no space left on device")
	if msg := silent.Settle(); msg != "" {
		t.Fatalf("SessionEnd displays nothing and returned %q", msg)
	}

	dir := t.TempDir()
	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir}))
	ctx, msg := responseMessage(t, out)
	if ctx != "" {
		t.Errorf("nothing to nudge about, and the agent was told something: %q", ctx)
	}
	if !strings.Contains(msg, "snapshot-write") || !strings.Contains(msg, "no space left on device") {
		t.Fatalf("Stop did not carry the SessionEnd failure: %q", out)
	}
}

// onRecord reads the record FILE, not the message. A clear cannot be asserted from the message: an
// entry already shown is held back by the 10-minute throttle, so "not in the message" is true of a
// cleared entry and of one still there — the first version of these tests made exactly that
// mistake, and a mutation pass caught it.
func onRecord(t *testing.T, s hookfailures.Stage) bool {
	t.Helper()
	path, err := hookfailures.Path("prosthetic-conscience")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatal(err)
	}
	for _, f := range rec.Failures {
		if f.Stage == s {
			return true
		}
	}
	return false
}

// THE BAND-CROSSING SAVE CLEARS A WRITE FAILURE. It is the other save in DecideContext, on the path
// that speaks — so a machine whose state became writable again, and whose context then crossed an
// edge, must stop being told its nudge is off.
func TestASuccessfulBandSaveClearsAWriteFailure(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	seed := hookfailures.New("prosthetic-conscience", "sc-stop", "Stop", time.Now(), io.Discard)
	seed.FailIn(StageStateWrite, dir, "seeded: the state could not be written")
	seed.Settle()

	out, _ := drive(t, dir, payload(t, map[string]any{"session_id": "s1", "cwd": dir, "transcript_path": transcriptAt(t, 400_000)}))
	if ctx, _ := responseMessage(t, out); ctx == "" {
		t.Fatal("setup: a first heavy turn should cross a band and save")
	}
	if onRecord(t, StageStateWrite) {
		t.Error("the band was saved and the write failure stayed on the record")
	}
}
