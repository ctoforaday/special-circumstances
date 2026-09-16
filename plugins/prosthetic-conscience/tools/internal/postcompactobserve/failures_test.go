package postcompactobserve

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

// POSTCOMPACT DISPLAYS NOTHING, and an observation that cannot be written used to be a stderr line at
// exit 0. A lost observation is cheap; a lost observation nobody knows about is what makes a thin
// corpus read like a quiet one — which is the exact failure this hook exists to measure.
func TestAnObservationThatCannotBeWrittenIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := withNote(t, "---\nschema: 3\n---\n## Objective\nkeep the thread\n")
	obs := filepath.Join(dir, ".claude", "checkpoints", "compaction-observations.jsonl")
	if err := os.MkdirAll(obs, 0o755); err != nil { // the corpus FILE as a directory
		t.Fatal(err)
	}
	stdin := `{"compact_summary":"keep the thread"}`
	stdout, _, code := call(t, dir, stdin)
	if code != 0 || stdout != "" {
		t.Fatalf("PostCompact must stay silent and never block: exit %d, stdout %q", code, stdout)
	}
	if !has(t, StageObservationAppend, dir) {
		t.Fatalf("an unwritable observation was not recorded")
	}
	if err := os.Remove(obs); err != nil {
		t.Fatal(err)
	}
	call(t, dir, stdin)
	if has(t, StageObservationAppend, dir) {
		t.Errorf("a written observation left the entry")
	}
	if len(rows(t, dir)) == 0 {
		t.Error("setup: the repaired run wrote no observation, so the clear proves nothing")
	}
}

func has(t *testing.T, s hookfailures.Stage, scope string) bool {
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
		if f.Stage == s && f.Scope == scope {
			return true
		}
	}
	return false
}

// THE NOTE IS THERE AND CANNOT BE READ: this compaction gets no observation, and nothing said so.
func TestANoteTheObserverCannotReadIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := withNote(t, "---\nschema: 3\n---\n## Objective\nkeep the thread\n")
	note := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	if err := os.Chmod(note, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(note, 0o644) })
	if _, err := os.ReadFile(note); err == nil {
		t.Skip("chmod 000 did not stop this process reading the note — the premise never took")
	}
	call(t, dir, `{"compact_summary":"keep the thread"}`)
	if !has(t, StageNoteRead, dir) {
		t.Fatal("an unreadable note was not recorded")
	}
	if err := os.Chmod(note, 0o644); err != nil {
		t.Fatal(err)
	}
	call(t, dir, `{"compact_summary":"keep the thread"}`)
	if has(t, StageNoteRead, dir) {
		t.Error("a readable note left the entry")
	}
}
