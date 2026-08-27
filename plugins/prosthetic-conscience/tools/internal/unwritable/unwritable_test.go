package unwritable

import (
	"os"
	"path/filepath"
	"testing"
)

// THE FALSE-NEGATIVE DIRECTION, which is the one that can be asserted at any privilege level.
//
// Took must never call a plainly writable directory restricted: if it did, Dir would run every
// caller's assertion against a premise that never held — the defect this package exists to stop,
// inverted. A 0o755 temp directory is writable to root and to nobody alike, so this arm holds
// wherever the suite runs.
//
// The other direction (0o500 IS restricted) is NOT asserted here, because whether it holds is
// exactly the environment fact Took is built to measure: it is true as an ordinary user and
// false as root, so an assertion either way would encode one machine. It is verified instead by
// the three call sites, which skip when it is false and run when it is true.
func TestTookDoesNotConvictAWritableDirectory(t *testing.T) {
	dir := t.TempDir()
	if Took(dir) {
		t.Fatalf("a fresh 0o755 temp directory was reported unwritable; every caller of Dir would "+
			"then skip unconditionally and report green for tests that never ran (uid %d)", os.Geteuid())
	}
}

// AND IT LEAVES NOTHING BEHIND. The probe writes into the directory under test, so a probe file
// that survived would be an artifact the code under test could then find — a helper that changes
// the state it is measuring.
func TestTookRemovesItsProbe(t *testing.T) {
	dir := t.TempDir()
	_ = Took(dir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		t.Errorf("the probe left %s behind in the directory under test", filepath.Join(dir, e.Name()))
	}
}
