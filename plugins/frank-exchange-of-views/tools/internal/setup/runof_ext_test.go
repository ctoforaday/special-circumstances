package setup_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// runOf is the external-test twin of the in-package helper: this file's tests drive setup
// through its exported surface, so they cannot reach the unexported one.
func runOf(t *testing.T, dir string) record.Run {
	t.Helper()
	r, err := record.NewRun(dir)
	if err != nil {
		t.Fatalf("resolving %q as a run: %v", dir, err)
	}
	return r
}
