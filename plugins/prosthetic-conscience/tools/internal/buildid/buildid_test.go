package buildid

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A NAME AND AN IDENTITY, always both. `-version` with the name alone tells a reader which tool
// they ran and nothing about which build, which is the state this package replaces.
func TestLineCarriesTheNameAndTheBuild(t *testing.T) {
	got := Line("sc-doctor")
	if !strings.HasPrefix(got, "sc-doctor ") {
		t.Errorf("Line = %q; must start with the tool name", got)
	}
	if strings.TrimSpace(strings.TrimPrefix(got, "sc-doctor")) == "" {
		t.Errorf("Line = %q; the identity half is empty, which reads as a bug rather than an answer", got)
	}
}

// UNSTAMPED MUST BE A WORD, NOT A GAP. The whole defect being fixed is a confident wrong number;
// replacing it with a confident empty string would be the same mistake in a new costume.
func TestRevisionIsNeverEmpty(t *testing.T) {
	if Revision() == "" {
		t.Fatal("Revision returned the empty string; an unknown build must SAY unknown")
	}
}

// THE LOAD-BEARING TEST. Everything above passes even if the toolchain stamps nothing — it would
// just report `unknown` forever, silently, which is the failure this package exists to prevent.
// So: build a real binary the ordinary way, run its -version, and require the printed revision to
// match the commit it was built from.
func TestAFreshlyBuiltBinaryReportsItsOwnCommit(t *testing.T) {
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("no go toolchain")
	}
	dir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	head, err := exec.Command("git", "-C", dir, "rev-parse", "--short=7", "HEAD").Output()
	if err != nil {
		t.Skip("not in a git checkout")
	}
	want := strings.TrimSpace(string(head))

	// `.exe` ON WINDOWS OR exec CANNOT FIND IT. `go build -o probe` writes exactly `probe`
	// there, and exec.Command then reports "executable file not found in %PATH%" — which reads
	// like a missing toolchain rather than a naming mistake. Caught by the Windows leg, not here.
	name := "probe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	out := filepath.Join(t.TempDir(), name)
	// No -ldflags: that absence is the assertion.
	build := exec.Command("go", "build", "-o", out, "./cmd/sc-doctor")
	build.Dir = dir
	if b, err := build.CombinedOutput(); err != nil {
		t.Skipf("build unavailable here: %v\n%s", err, b)
	}

	got, err := exec.Command(out, "-version").Output()
	if err != nil {
		t.Fatalf("-version failed: %v", err)
	}
	line := strings.TrimSpace(string(got))
	if !strings.Contains(line, "sc-doctor") {
		t.Errorf("-version = %q; must name the tool", line)
	}
	if !strings.Contains(line, want) {
		t.Errorf("-version = %q; must report the commit it was built from (%s). "+
			"If this fails with `unknown`, the toolchain stopped stamping vcs.revision and every "+
			"binary is now reporting nothing rather than something wrong", line, want)
	}
}
