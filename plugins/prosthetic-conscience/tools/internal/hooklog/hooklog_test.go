package hooklog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendAccumulatesAndBatches(t *testing.T) {
	dir := t.TempDir()
	Append(dir, "first\n")
	// Several lines in ONE call is how the merged binary keeps a single writer single.
	Append(dir, "second\n", "third\n")

	b, err := os.ReadFile(Path(dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != "first\nsecond\nthird\n" {
		t.Errorf("log = %q; appends must accumulate in order", got)
	}
}

// An empty project dir must DISABLE the log, not write a relative .claude/ into whatever
// directory the process started in — the defect internal/hookenv exists to close.
func TestNoProjectDirWritesNothingRelative(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)
	Append("", "should not be written\n")
	if _, err := os.Stat(filepath.Join(wd, ".claude")); !os.IsNotExist(err) {
		t.Errorf("an empty projectDir created %s/.claude", wd)
	}
}

// Instrumentation must never cost the tool call.
func TestAppendNeverPanics(t *testing.T) {
	dir := t.TempDir()
	blocked := filepath.Join(dir, "not-a-dir")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	Append(blocked, "swallowed\n")

	Append(dir) // no lines at all
	if _, err := os.Stat(Path(dir)); !os.IsNotExist(err) {
		t.Error("a call with no lines must not create the file")
	}

	fresh := filepath.Join(t.TempDir(), "does", "not", "exist")
	Append(fresh, "created\n")
	if b, err := os.ReadFile(Path(fresh)); err != nil || !strings.Contains(string(b), "created") {
		t.Errorf("the log must create its own directory: %v", err)
	}
}
