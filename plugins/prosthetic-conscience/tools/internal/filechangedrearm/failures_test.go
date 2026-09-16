package filechangedrearm

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/checkpoint"
	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/hookfailures"
)

// FILECHANGED DISPLAYS NOTHING. Every message this hook had — including the one its own comment
// calls "the severe case … all session, silently" — went to stderr at exit 0. They are recorded now,
// scoped by project, and the hook still writes nothing to stdout.

func rearmRecord(t *testing.T) []hookfailures.Failure {
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
		t.Fatal(err)
	}
	return rec.Failures
}

func recorded(fs []hookfailures.Failure, s hookfailures.Stage, scope string) bool {
	for _, f := range fs {
		if f.Stage == s && f.Scope == scope {
			return true
		}
	}
	return false
}

func fireOut(t *testing.T, dir, filePath string) (stdout string, code int) {
	t.Helper()
	in, _ := json.Marshal(hookInput{FilePath: filePath, Event: "change", SessionID: "s1"})
	var o, e bytes.Buffer
	code = run(nil, bytes.NewReader(in), &o, &e, dir, noon)
	return o.String(), code
}

// THE SEVERE CASE: a loop a reader counts and the parser opens nothing from. No file change re-arms
// anything for the whole session.
func TestALoopThatOpensNoChecksIsRecorded(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := project(t)
	// Every entry lettered: a reader counts two checks, the parser opens none.
	write(t, filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md"),
		"---\nschema: 3\n---\n## Validation loop\n1a. `a` · re-armed by: tools/\n1b. `b` · re-armed by: tools/\n")
	stdout, code := fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if code != 0 || stdout != "" {
		t.Fatalf("FileChanged must stay silent and never block: exit %d, stdout %q", code, stdout)
	}
	if !recorded(rearmRecord(t), StageLoopOpensNoChecks, dir) {
		t.Fatalf("a loop that opens no checks was not recorded: %+v", rearmRecord(t))
	}
}

// The re-arm state cannot be written: the change happened and nothing marked the check stale.
func TestARearmThatCannotBeWrittenIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := project(t)
	state := checkpoint.RearmPath(dir)
	if err := os.MkdirAll(state, 0o755); err != nil { // the state FILE as a directory
		t.Fatal(err)
	}
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if !recorded(rearmRecord(t), StageRearmWrite, dir) {
		t.Fatalf("an unwritable re-arm state was not recorded: %+v", rearmRecord(t))
	}

	if err := os.Remove(state); err != nil {
		t.Fatal(err)
	}
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if recorded(rearmRecord(t), StageRearmWrite, dir) {
		t.Errorf("the re-arm was written and its entry stayed: %+v", rearmRecord(t))
	}
}

// A CHANGE A CHECK CLAIMS is the loop working: it clears the "opens no checks" and "unclaimed" entries
// for that project. Without the clear, one bad note would be announced for the rest of the day
// after the note was fixed.
func TestAClaimedChangeClearsTheLoopEntries(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := project(t)
	notePath := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	write(t, notePath, "---\nschema: 3\n---\n## Validation loop\n1a. `a` · re-armed by: tools/\n1b. `b` · re-armed by: tools/\n")
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if !recorded(rearmRecord(t), StageLoopOpensNoChecks, dir) {
		t.Fatalf("setup: the bad loop was not recorded: %+v", rearmRecord(t))
	}
	// Seed an UNCLAIMED-change entry as well: the bad note above records "opens no checks" and never
	// reaches the unclaimed branch, so without a seed that clear would be unobservable.
	seed := hookfailures.New("prosthetic-conscience", "sc-filechanged-rearm", "FileChanged", noon, io.Discard)
	seed.FailIn(StageUnclaimedChange, dir, "seeded")
	seed.Settle()
	if !recorded(rearmRecord(t), StageUnclaimedChange, dir) {
		t.Fatal("setup: the seed did not land")
	}
	write(t, notePath, note) // the harness's healthy note, whose first check claims tools/
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	for _, s := range []hookfailures.Stage{StageLoopOpensNoChecks, StageUnclaimedChange} {
		if recorded(rearmRecord(t), s, dir) {
			t.Errorf("a claimed change left %s on the record: %+v", s, rearmRecord(t))
		}
	}
}

// TWO CHECKS NAMING ONE COMMAND make one re-arm record stand for several, and nothing is pruned while
// the identity is ambiguous — a prune that silently does nothing looks exactly like a file with
// nothing to prune. Recorded; distinct commands clear it.
func TestAmbiguousCheckKeysAreRecordedAndClear(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := project(t)
	notePath := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	write(t, notePath, "---\nschema: 3\n---\n## Validation loop\n"+
		"1. `go test ./...` → ok · re-armed by: tools/\n   last run: pass\n"+
		"2. `go test ./...` → ok · re-armed by: .qlty/qlty.toml\n   last run: pass\n")
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if !recorded(rearmRecord(t), StageAmbiguousChecks, dir) {
		t.Fatalf("two checks with one command were not recorded: %+v", rearmRecord(t))
	}
	write(t, notePath, note)
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if recorded(rearmRecord(t), StageAmbiguousChecks, dir) {
		t.Errorf("distinct commands left the entry: %+v", rearmRecord(t))
	}
}

// THE NOTE IS THERE AND CANNOT BE READ: no file change can re-arm anything, and returning 0 looked
// exactly like a note with nothing to re-arm.
func TestANoteTheRearmCannotReadIsRecordedAndClears(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	dir := project(t)
	notePath := filepath.Join(dir, ".claude", "checkpoints", "CHECKPOINT.md")
	if err := os.Chmod(notePath, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(notePath, 0o644) })
	if _, err := os.ReadFile(notePath); err == nil {
		t.Skip("chmod 000 did not stop this process reading the note — the premise never took")
	}
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if !recorded(rearmRecord(t), StageNoteRead, dir) {
		t.Fatalf("an unreadable note was not recorded: %+v", rearmRecord(t))
	}
	if err := os.Chmod(notePath, 0o644); err != nil {
		t.Fatal(err)
	}
	fireOut(t, dir, filepath.Join(dir, "tools", "internal", "x.go"))
	if recorded(rearmRecord(t), StageNoteRead, dir) {
		t.Errorf("a readable note left the entry: %+v", rearmRecord(t))
	}
}
