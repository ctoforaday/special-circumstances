package buildid

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// The assertion that MATTERS is not the string — it is that a binary built under this name
// can actually be started. A suffix test would pass on Linux while asserting nothing, since
// Linux is the platform where the answer is "add nothing".
func TestABinaryBuiltUnderThisNameCanBeExecuted(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles a binary")
	}
	out := filepath.Join(t.TempDir(), ExeName("probe"))
	if msg, err := exec.Command("go", "build", "-o", out, "../../cmd/sc-doctor").CombinedOutput(); err != nil {
		t.Fatalf("building the probe: %v\n%s", err, msg)
	}
	if err := exec.Command(out, "-version").Run(); err != nil {
		t.Errorf("a binary built as %s could not be executed: %v — on Windows this reports "+
			"\"executable file not found\", which reads like a missing toolchain rather than "+
			"a naming mistake", filepath.Base(out), err)
	}
}

func TestTheSuffixFollowsThePlatform(t *testing.T) {
	got := ExeName("sc-doctor")
	want := "sc-doctor"
	if runtime.GOOS == "windows" {
		want = "sc-doctor.exe"
	}
	if got != want {
		t.Errorf("ExeName = %q, want %q on %s", got, want, runtime.GOOS)
	}
}
