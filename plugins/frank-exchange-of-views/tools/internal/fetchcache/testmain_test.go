package fetchcache

import (
	"fmt"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — the leak that is invisible on Linux and fails the Windows
// leg. See recordtest.Main (#666).
//
// IT NO LONGER *IS* recordtest.Main, because one more thing has to happen around m.Run: every test
// here builds its own run under its own t.TempDir(), and the PDFium module cache is keyed on the
// run's directory, so each of the 21 tests that touch a PDF paid the full 3,968 ms compile — 68s
// for a package whose real work is milliseconds, and past the default timeout under -race.
// UseSharedModuleCache holds the whole argument and the cleanup design; this is one of its two
// callers. recordtest documents CheckOrphanedHandles as the seam for exactly this case: a package
// that needs its own TestMain keeps the handle check as a post-suite hook rather than giving it up.
//
// recordtest.Main calls os.Exit, so the release below cannot be a defer.
func TestMain(m *testing.M) {
	release, err := UseSharedModuleCache()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	code := m.Run()

	if err := release(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := recordtest.CheckOrphanedHandles(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
