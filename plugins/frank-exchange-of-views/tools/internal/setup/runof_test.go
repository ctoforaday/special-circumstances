package setup

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// runOf resolves a directory as the run these fixtures hand to setup.
//
// NewRun, not OpenRun, and the distinction is the one the two constructors exist to draw: most
// of these fixtures pass a directory to BuildSkeleton, which is what CREATES it. Requiring the
// directory to exist first would make the test set up the very thing the function under test is
// responsible for producing.
func runOf(t *testing.T, dir string) record.Run {
	t.Helper()
	r, err := record.NewRun(dir)
	if err != nil {
		t.Fatalf("resolving %q as a run: %v", dir, err)
	}
	return r
}
