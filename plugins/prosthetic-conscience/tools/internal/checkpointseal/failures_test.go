package checkpointseal

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/unwritable"
)

// NONE OF THIS HOOK'S THREE EVENTS DISPLAYS ANYTHING. Every failure below was a stderr line at exit
// 0, which reaches the debug log and nobody; the seal is where a failure costs the most, because a
// snapshot that was never written is invisible until the session that needed it. These tests break
// the real thing — a file where a directory belongs, a directory where a file belongs, a transcript
// that is not there — and assert each failure reaches the record under its own stage, that the
// hook still says NOTHING on stdout, and that the stage's own next success clears it.

func sealRecord(t *testing.T) []hookfailures.Failure {
	t.Helper()
	path, err := hookfailures.Path("prosthetic-conscience")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var rec hookfailures.Record
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("the record does not parse: %v\n%s", err, b)
	}
	return rec.Failures
}

func hasStage(fs []hookfailures.Failure, s hookfailures.Stage) (hookfailures.Failure, bool) {
	for _, f := range fs {
		if f.Stage == s {
			return f, true
		}
	}
	return hookfailures.Failure{}, false
}

// sealProject is a project with a note in it, and a state directory of its own.
func sealProject(t *testing.T, note string) string {
	t.Helper()
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	cp := filepath.Join(dir, ".claude", "checkpoints")
	if err := os.MkdirAll(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cp, "CHECKPOINT.md"), []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

const badLoopNote = "---\nschema: 3\n---\n" +
	"## Validation loop\n1. `a` · re-armed by: tools/\n7b. `b` · re-armed by: tools/\n" +
	"## Next intended steps\n1. only one step\n"

func TestEverySealFailureIsRecordedAndClearedByItsOwnSuccess(t *testing.T) {
	goodLoop := noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass")
	cases := []struct {
		name  string
		stage hookfailures.Stage
		note  string
		brk   func(t *testing.T, dir string)
		fix   func(t *testing.T, dir string)
		stdin func(dir string) string
	}{
		{
			// seals.jsonl as a DIRECTORY: the snapshot is written and its row cannot be.
			name: "seal-row", stage: StageSealRow, note: goodLoop,
			brk: func(t *testing.T, dir string) {
				if err := os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints", "seals.jsonl"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			fix: func(t *testing.T, dir string) {
				if err := os.Remove(filepath.Join(dir, ".claude", "checkpoints", "seals.jsonl")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			// A transcript that is not there: the drift check cannot run, and that must not read as
			// "no drift".
			name: "note-check", stage: StageNoteCheck, note: goodLoop,
			stdin: func(dir string) string {
				return `{"transcript_path":"` + filepath.ToSlash(filepath.Join(dir, "gone.jsonl")) + `"}`
			},
		},
		{
			// A loop entry the parser cannot open: ACTION-NEEDED, and the human is the one who can
			// fix the live note while the context that explains it still exists.
			name: "note-loop-problems", stage: StageLoopProblems,
			// A LETTERED entry ("7b."): a reader counts it, the parser opens nothing from it. The
			// fixture is the one checkpoint's own NoteLoopProblems test pins.
			note: badLoopNote,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := sealProject(t, tc.note)
			if tc.brk != nil {
				tc.brk(t, dir)
			}
			stdin := `{}`
			if tc.stdin != nil {
				stdin = tc.stdin(dir)
			}
			for _, ev := range []string{evPreCompact, evSessionEnd, evSubagentStop} {
				stdout, _, code := call(t, stdin, dir, noon, "-event", ev)
				if code != 0 {
					t.Fatalf("%s: the seal must never block: exit %d", ev, code)
				}
				if ev != evPreCompact && stdout != "" {
					// PreCompact may steer the summarizer on stdout; the other two must never
					// speak, and SubagentStop least of all — an emission there re-invokes the seat.
					t.Errorf("%s wrote to stdout: %q", ev, stdout)
				}
				if strings.Contains(stdout, "systemMessage") {
					t.Errorf("%s displays nothing, and carried a systemMessage: %q", ev, stdout)
				}
			}
			f, ok := hasStage(sealRecord(t), tc.stage)
			if !ok {
				t.Fatalf("%s failed and nothing recorded it: %+v", tc.stage, sealRecord(t))
			}
			if f.Error == "" {
				t.Errorf("recorded with no detail: %+v", f)
			}

			if tc.fix == nil && tc.stdin == nil {
				return // a note problem is fixed by editing the note, which the next case covers
			}
			if tc.fix != nil {
				tc.fix(t, dir)
			}
			call(t, `{}`, dir, noon, "-event", evPreCompact)
			if tc.stage == StageNoteCheck {
				return // with no transcript the check is not RUN, so it cannot clear; asserted below
			}
			if _, still := hasStage(sealRecord(t), tc.stage); still {
				t.Errorf("the stage worked and its entry stayed: %+v", sealRecord(t))
			}
		})
	}
}

// A note whose loop is repaired clears the loop entry: the finding is about the note, and the note
// is what changed.
func TestARepairedLoopClearsItsEntry(t *testing.T) {
	dir := sealProject(t, badLoopNote)
	call(t, `{}`, dir, noon, "-event", evPreCompact)
	if _, ok := hasStage(sealRecord(t), StageLoopProblems); !ok {
		t.Fatalf("the bad loop was not recorded: %+v", sealRecord(t))
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		[]byte(noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass")), 0o644); err != nil {
		t.Fatal(err)
	}
	call(t, `{}`, dir, noon, "-event", evPreCompact)
	if _, ok := hasStage(sealRecord(t), StageLoopProblems); ok {
		t.Errorf("the repaired loop left its entry: %+v", sealRecord(t))
	}
}

// A snapshot directory that cannot be created is the seal failing outright: no snapshot at all. The
// note lives in a workspace (projects/<name>/CHECKPOINT.md, which NotePath searches first) so it is
// still found while .claude/checkpoints is a FILE and the snapshot directory cannot be made.
func TestASnapshotThatCannotBeWrittenIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := t.TempDir()
	ws := filepath.Join(dir, "projects", "work")
	if err := os.MkdirAll(ws, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "CHECKPOINT.md"), []byte(noteWith("1. `go test` → ok\n   last run: pass")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, ".claude"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".claude", "checkpoints"), []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, code := call(t, `{}`, dir, noon, "-event", evSessionEnd)
	if code != 0 || stdout != "" {
		t.Fatalf("SessionEnd must stay silent and never block: exit %d, stdout %q", code, stdout)
	}
	if _, ok := hasStage(sealRecord(t), StageSnapshotDir); !ok {
		t.Fatalf("a snapshot directory that cannot be created was not recorded: %+v", sealRecord(t))
	}

	// Repaired, the next seal writes its snapshot and clears the entry.
	if err := os.Remove(filepath.Join(dir, ".claude", "checkpoints")); err != nil {
		t.Fatal(err)
	}
	call(t, `{}`, dir, noon, "-event", evSessionEnd)
	if _, ok := hasStage(sealRecord(t), StageSnapshotDir); ok {
		t.Errorf("the snapshot directory was created and its entry stayed: %+v", sealRecord(t))
	}
}

// A NOTE THE SEAL CANNOT READ fails twice, at two separate reads: the steering read on PreCompact and
// the seal's own. Both are recorded, and a readable note clears both.
func TestANoteTheSealCannotReadIsRecordedAtBothReads(t *testing.T) {
	dir := sealProject(t, noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass"))
	note := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	if err := os.Chmod(note, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(note, 0o644) })
	if _, err := os.ReadFile(note); err == nil {
		t.Skip("chmod 000 did not stop this process reading the note — the premise never took")
	}
	call(t, `{}`, dir, noon, "-event", evPreCompact)
	for _, s := range []hookfailures.Stage{StageSteerNoteRead, StageSnapshotRead} {
		if _, ok := hasStage(sealRecord(t), s); !ok {
			t.Errorf("%s was not recorded: %+v", s, sealRecord(t))
		}
	}
	if err := os.Chmod(note, 0o644); err != nil {
		t.Fatal(err)
	}
	call(t, `{}`, dir, noon, "-event", evPreCompact)
	for _, s := range []hookfailures.Stage{StageSteerNoteRead, StageSnapshotRead} {
		if _, ok := hasStage(sealRecord(t), s); ok {
			t.Errorf("a readable note left %s: %+v", s, sealRecord(t))
		}
	}
}

// A SNAPSHOT THAT CANNOT BE WRITTEN into a directory that exists: the seal's whole product is lost.
func TestASnapshotWriteThatFailsIsRecordedAndClears(t *testing.T) {
	dir := sealProject(t, noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass"))
	cp := filepath.Join(dir, ".claude", "checkpoints")
	unwritable.Dir(t, cp)
	call(t, `{}`, dir, noon, "-event", evSessionEnd)
	if _, ok := hasStage(sealRecord(t), StageSnapshotWrite); !ok {
		t.Fatalf("a snapshot that could not be written was not recorded: %+v", sealRecord(t))
	}
	if err := os.Chmod(cp, 0o755); err != nil {
		t.Fatal(err)
	}
	call(t, `{}`, dir, noon, "-event", evSessionEnd)
	if _, ok := hasStage(sealRecord(t), StageSnapshotWrite); ok {
		t.Errorf("a written snapshot left the entry: %+v", sealRecord(t))
	}
}

// THE CLEARING CALLS FOR THE FINDINGS. A drift finding, a check that could not run, and an impossible
// written_at each persist until the next seal where the thing is fine — and "fine" has to clear
// them, or a note fixed at 10:00 is still reported as broken at 17:00. Seeded through the recorder
// (the failure paths have their own tests above) so this asserts the CLEAR call sites alone.
func TestHealthySealsClearTheFindings(t *testing.T) {
	dir := sealProject(t, "---\nschema: 3\nwritten_at: 2026-07-28T11:00:00Z\n---\n## Validation loop\n"+
		"1. `go test` → ok · re-armed by: tools/\n   last run: pass\n")
	seed := hookfailures.New("prosthetic-conscience", "sc-precompact", evPreCompact, noon, io.Discard)
	// Seeded IN THIS PROJECT: every seal stage is about one project's files, so a seed at machine
	// scope would be a different entry that no seal here could ever clear.
	seed.FailIn(StageDrift, dir, "seeded drift")
	seed.FailIn(StageNoteCheck, dir, "seeded unreadable transcript")
	seed.FailIn(StageWrittenAt, dir, "seeded impossible written_at")
	seed.Settle()

	// A transcript that EXISTS, holds a record, and records no writes: the check runs, and there is
	// nothing to drift. (An EMPTY file is "held no records", which is the unreadable arm.)
	tr := filepath.Join(t.TempDir(), "t.jsonl")
	line := `{"type":"user","timestamp":"2026-07-28T11:30:00Z","message":{"role":"user","content":"carry on"}}` + "\n"
	if err := os.WriteFile(tr, []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	call(t, `{"transcript_path":"`+filepath.ToSlash(tr)+`"}`, dir, noon, "-event", evPreCompact)
	for _, s := range []hookfailures.Stage{StageDrift, StageNoteCheck, StageWrittenAt} {
		if f, ok := hasStage(sealRecord(t), s); ok {
			t.Errorf("a healthy seal left %s on the record: %+v", s, f)
		}
	}
}

// A PRUNE THAT WORKS CLEARS A PRUNE FAILURE. The failure itself is not reached by a fixture: listing
// fails only on a directory the seal has just written a snapshot into, and a snapshot that cannot be
// removed means a directory that cannot be written — where the new snapshot fails first. So the
// failure paths are named here and the CLEAR is tested, seeded in this project.
func TestAHealthyPruneClearsAPruneFailure(t *testing.T) {
	dir := sealProject(t, noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass"))
	seed := hookfailures.New("prosthetic-conscience", "sc-precompact", evPreCompact, noon, io.Discard)
	seed.FailIn(StagePrune, dir, "seeded: snapshots could not be pruned")
	seed.Settle()
	call(t, `{}`, dir, noon, "-event", evSessionEnd)
	if _, ok := hasStage(sealRecord(t), StagePrune); ok {
		t.Errorf("a healthy prune left the entry: %+v", sealRecord(t))
	}
}

// THE FALSE ALARM, DRIVEN THROUGH THE REAL HOOK. Found live on 2026-09-17, the day this record shipped:
// a main-agent turn end fires SubagentStop with no agent_type and a predicted transcript nothing writes
// (#189), and the drift check recorded "cannot check the note" on every one — shown every ten
// minutes, about nothing. A turn end records no note-check at all; a typed seat whose transcript is
// missing is still recorded, because that one is real.
func TestATurnEndIsNotANoteCheckFailureButAMissingSeatTranscriptIs(t *testing.T) {
	goodLoop := noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass")
	for name, tc := range map[string]struct {
		agentType string
		recorded  bool
	}{
		"turn end":                      {"", false},
		"seat whose transcript is gone": {"frank-exchange-of-views:red-lens-logic", true},
	} {
		t.Run(name, func(t *testing.T) {
			dir := sealProject(t, goodLoop)
			predicted := filepath.ToSlash(filepath.Join(dir, "subagents", "agent-never-written.jsonl"))
			payload := `{"session_id":"s1","agent_id":"a1","agent_type":"` + tc.agentType + `","agent_transcript_path":"` + predicted + `"}`
			stdout, _, code := call(t, payload, dir, noon, "-event", evSubagentStop)
			if code != 0 || stdout != "" {
				t.Fatalf("SubagentStop must stay silent and never block: exit %d, stdout %q", code, stdout)
			}
			if _, got := hasStage(sealRecord(t), StageNoteCheck); got != tc.recorded {
				t.Errorf("note-check recorded = %v, want %v: %+v", got, tc.recorded, sealRecord(t))
			}
		})
	}
}

// A TURN END CLEARS NOTHING. Its path is empty, and an empty path reads as "no transcript" — which
// the check would otherwise take as passing, and clear a real note-check failure recorded by an
// earlier seal with no evidence the check works now.
func TestATurnEndDoesNotClearARealNoteCheckFailure(t *testing.T) {
	dir := sealProject(t, noteWith("1. `go test` → ok · re-armed by: tools/\n   last run: pass"))
	seed := hookfailures.New("prosthetic-conscience", "sc-precompact", evPreCompact, noon, io.Discard)
	seed.FailIn(StageNoteCheck, dir, "a real failure from an earlier seal")
	seed.Persist()

	predicted := filepath.ToSlash(filepath.Join(dir, "subagents", "agent-never-written.jsonl"))
	call(t, `{"session_id":"s1","agent_id":"a1","agent_transcript_path":"`+predicted+`"}`, dir, noon, "-event", evSubagentStop)
	f, ok := hasStage(sealRecord(t), StageNoteCheck)
	if !ok || f.Error != "a real failure from an earlier seal" {
		t.Errorf("a turn end cleared or rewrote a real note-check failure: %+v", sealRecord(t))
	}
}
