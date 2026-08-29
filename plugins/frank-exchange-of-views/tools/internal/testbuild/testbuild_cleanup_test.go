package testbuild

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE CLAIM IS ABOUT DISK, SO IT IS MEASURED ON DISK. Asserting that Run "calls RemoveAll"
// would restate the implementation; what #643 is about is whether a directory still exists
// after the process that made it is done with it.
//
// Run's real teardown cannot be exercised in-process (it wraps m.Run, which owns this very
// binary), so this drives removeBuildDir — the half that does the work — against a directory
// standing in for the one Run would remove.
func TestRemoveBuildDirDeletesWhatWasBuilt(t *testing.T) {
	tmp := t.TempDir()
	stand := filepath.Join(tmp, "sc-testbuild-stand-in")
	if err := os.MkdirAll(stand, 0o755); err != nil {
		t.Fatal(err)
	}
	// A real one holds linked binaries, which are what make it 38 MB.
	if err := os.WriteFile(filepath.Join(stand, ExeName("feov-record")), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}

	saved := dir
	dir = stand
	t.Cleanup(func() { dir = saved })

	removeBuildDir()

	if _, err := os.Stat(stand); !os.IsNotExist(err) {
		t.Errorf("the build directory survived removeBuildDir: Stat err = %v — this is exactly "+
			"the leak #643 measured at 485 MB across one session", err)
	}
}

// A PROCESS THAT BUILT NOTHING MUST NOT REMOVE ANYTHING. `dir` is empty until the first
// Binary call, and an unguarded RemoveAll("") is a call on the current directory's parentage
// that no test should ever make.
func TestRemoveBuildDirDoesNothingWhenNothingWasBuilt(t *testing.T) {
	saved := dir
	dir = ""
	t.Cleanup(func() { dir = saved })
	removeBuildDir() // must not panic, must not touch anything
}

// THE GUARD IS THE REASON THIS CANNOT REGRESS. A package calling Binary without TestMain
// would leak in silence — passing tests, growing disk. underRun is what makes that a loud
// refusal instead.
//
// It already caught one: this very package calls Binary UNQUALIFIED from its own tests, so
// the `testbuild.Binary` grep used to enumerate callers missed it, and the guard refused the
// suite until testmain_test.go was added here too. That is the case the guard exists for,
// found on its first run.
func TestRunIsWhatSetsTheGuardFlag(t *testing.T) {
	// Nothing but Run writes underRun, so observing it true inside a package whose TestMain
	// calls Run is the evidence that the wiring is live. A false reading here means every
	// legitimate caller would be refused.
	if !underRun.Load() {
		t.Fatal("underRun is false inside a package whose TestMain calls Run — the guard would " +
			"refuse every legitimate caller, and no build directory would ever be removed")
	}
}

// AND THE REFUSAL ITSELF IS EXERCISED, IN BOTH DIRECTIONS. Testing only that the flag is true
// left the guard live but unexercised — deleting it failed nothing, which is how an untested
// guard rots into decoration.
func TestMissingTestMainRefusesOnlyWhenTestMainIsAbsent(t *testing.T) {
	if msg := missingTestMain(true); msg != "" {
		t.Errorf("missingTestMain(true) = %q, want no refusal — a package that DID wire Run "+
			"must be served", msg)
	}
	msg := missingTestMain(false)
	if msg == "" {
		t.Fatal("missingTestMain(false) returned no refusal — a package with no TestMain would " +
			"be served, and its build directory abandoned exactly as in #643")
	}
	// The refusal has to carry the remedy. "This package has no TestMain" alone leaves a
	// reader to guess the signature, and a guard that is annoying to satisfy gets deleted.
	for _, want := range []string{"TestMain", "testbuild.Run(m)", "#643"} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal = %q, want it to mention %q", msg, want)
		}
	}
}
