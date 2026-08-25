package buildid

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A NAME AND AN IDENTITY, always both.
func TestLineCarriesTheNameAndTheBuild(t *testing.T) {
	got := Line("gray-area")
	if !strings.HasPrefix(got, "gray-area ") {
		t.Errorf("Line = %q; must start with the tool name", got)
	}
	if strings.TrimSpace(strings.TrimPrefix(got, "gray-area")) == "" {
		t.Errorf("Line = %q; the identity half is empty, which reads as a bug rather than an answer", got)
	}
}

// Unknown must be a WORD. Replacing a confident wrong number with a confident empty string is the
// same mistake in a new costume.
func TestRevisionIsNeverEmpty(t *testing.T) {
	if Revision() == "" {
		t.Fatal("Revision returned the empty string; an unknown build must SAY unknown")
	}
}

// THE LOAD-BEARING TEST, AND THE REASON THIS COPY HAS ONE.
//
// buildid is duplicated across two Go modules because the plugins ship independently and cannot
// import each other. Duplicated code with the guard in only ONE copy is worse than either: the
// unguarded copy silently reports `unknown` forever if the toolchain ever stops stamping, and the
// other module's green suite says nothing about it. So the check lives in both, on purpose.
//
// It builds a real binary the ordinary way — no -ldflags, which is the assertion — and reads its
// -version back.
func TestAFreshlyBuiltBinaryReportsItsOwnCommit(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	// THE REVISION GO STAMPS FROM, not this working tree's HEAD — they differ in every
	// worktree, which is where this repo is developed (#532).
	stampDir, ok := StampedFrom(dir)
	if !ok {
		t.Skip("not in a git checkout")
	}
	head, err := exec.Command("git", "-C", stampDir, "rev-parse", "--short=7", "HEAD").Output()
	if err != nil {
		t.Skip("not in a git checkout")
	}
	want := strings.TrimSpace(string(head))

	// `.exe` on Windows or exec cannot find it: `go build -o probe` writes exactly `probe`, and
	// the resulting error reads like a missing toolchain rather than a naming mistake.
	name := "probe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(t.TempDir(), name)
	build := exec.Command("go", "build", "-o", out, "./cmd/gray-area")
	build.Dir = dir
	if b, err := build.CombinedOutput(); err != nil {
		t.Skipf("build unavailable here: %v\n%s", err, b)
	}

	got, err := exec.Command(out, "-version").Output()
	if err != nil {
		t.Fatalf("-version failed: %v", err)
	}
	line := strings.TrimSpace(string(got))
	if !strings.Contains(line, "gray-area") {
		t.Errorf("-version = %q; must name the tool", line)
	}
	if !strings.Contains(line, want) {
		t.Errorf("-version = %q; must report the commit it was built from (%s). If this fails with "+
			"`unknown`, the toolchain stopped stamping vcs.revision and this binary now reports "+
			"nothing rather than something wrong", line, want)
	}
}
