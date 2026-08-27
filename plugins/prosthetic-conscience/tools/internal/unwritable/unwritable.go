// Package unwritable makes a directory unwritable for a test, and REFUSES TO PRETEND when
// it cannot.
//
// Three tests in this module took the same shape: chmod a checkpoints directory to 0o500,
// then assert the code under test copes with a write it cannot perform. On a machine where
// the chmod does not restrict the caller, all three exercise nothing — and MEASURED
// 2026-08-27 in a container running as uid 0, two of them went on PASSING while testing
// nothing (checkpointseal's seal-survives and postcompactobserve's always-exits-zero both
// assert exit 0, which a successful write also produces), and the third failed with a
// message about the nudge that named no cause.
//
// The silent pair is the expensive one, and stopnudge_test.go already had the reasoning
// written down — for Windows, in the skip directly above the arm this package replaces: "a
// test that passes without testing is worse than one that says it did not run". The Windows
// hole was seen because os.Chmod's no-op there is documented. The root hole is the same hole
// and was not.
//
// SO THE RESTRICTION IS VERIFIED, NOT ASSUMED. Dir attempts a write and skips if it lands.
// That is a CAPABILITY check rather than an identity check, which matters in both
// directions: a container can drop CAP_DAC_OVERRIDE from root (where a Geteuid()==0 skip
// would then decline to run a test that would have worked), and a process can hold the
// capability without being uid 0 (where the same check would report a green it did not earn).
package unwritable

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Dir makes dir unwritable for the remainder of the test, restoring it afterwards.
//
// It SKIPS — never fails, and never silently continues — when the platform or the caller's
// privileges mean the directory is still writable, because a test whose premise did not hold
// has not disproved anything.
func Dir(t *testing.T, dir string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod cannot make a directory unwritable on Windows; the portable arms of " +
			"these suites (a FILE where the directory must be) cover the same property")
	}
	if err := os.Chmod(dir, 0o500); err != nil { // readable, not writable
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	// THE PROBE IS THE POINT. Everything above this line is what the three call sites already
	// did, and it is the part that can quietly do nothing.
	if !Took(dir) {
		t.Skipf("chmod 0500 did not stop this process writing to %s — the caller holds the "+
			"privilege to override it (uid %d), so this test's premise never took and a result "+
			"either way would be about the machine, not the code", dir, os.Geteuid())
	}
}

// Took reports whether dir is genuinely unwritable to THIS process, by attempting a write and
// removing what it created. It is separate from Dir so the decision can be tested without a
// testing.T to fake.
//
// A create that FAILS for any reason is read as restricted, and that is deliberate: this is a
// premise check for a test that is about to assert something cannot be written, so every
// failure mode it could confuse (a missing directory, a full disk) is one where that test's
// premise does hold. The cost of being wrong runs one way only — an unnecessary skip, which
// says so — never a green nobody earned.
func Took(dir string) bool {
	probe := filepath.Join(dir, ".unwritable-probe")
	f, err := os.OpenFile(probe, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return true
	}
	_ = f.Close()
	_ = os.Remove(probe)
	return false
}
