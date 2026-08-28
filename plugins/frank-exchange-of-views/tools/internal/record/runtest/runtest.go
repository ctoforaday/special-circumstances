// Package runtest resolves a run directory into a record.Run for tests.
//
// It is separate from recordtest, and has to be. Files in `package record` import recordtest,
// so recordtest cannot import record without making record's own test binary a cycle. This
// package is imported only by tests of OTHER packages, which is why it may name the type.
package runtest

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Open resolves an EXISTING run, failing the test if it does not resolve.
//
// The refusal is the point of the type, so a fixture that cannot produce a real run should
// fail loudly here rather than hand a zero Run to the code under test — a zero Run reads every
// board as empty, which is the answer these tests most often assert against.
func Open(t *testing.T, dir string) record.Run {
	t.Helper()
	r, err := record.OpenRun(dir)
	if err != nil {
		t.Fatalf("resolving %q as a run: %v", dir, err)
	}
	return r
}

// New resolves a run that need not exist yet — for fixtures that hand a directory to the code
// responsible for creating it.
func New(t *testing.T, dir string) record.Run {
	t.Helper()
	r, err := record.NewRun(dir)
	if err != nil {
		t.Fatalf("resolving %q as a run to create: %v", dir, err)
	}
	return r
}
