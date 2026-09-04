package fetchcache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// moduleCacheName is a FIXED path, not a unique one, and that is the cleanup design.
//
// The obvious shape — MkdirTemp plus a RemoveAll before os.Exit — cleans up on every path except
// the one that matters: `go test` panics the process on a timeout, which is exactly how this
// package failed before the cache was shared, and a panic runs no deferred removal. Measured: a
// hard timeout leaked a 5 MB compiled module into TMPDIR, once per crash, forever.
//
// A fixed name cannot accumulate. A crash leaves ONE directory; the next run finds it, uses it
// warm, and removes it on the way out. Residue is bounded at one either way, and self-healing
// rather than dependent on a cleanup step running.
//
// SAFE TO REUSE AND SAFE TO DELETE, because it holds nothing but wazero's compilation of a module
// that ships in the binary — keyed by content and wazero version, reproducible at any time for the
// 3,968 ms it costs. Two concurrent test binaries sharing it can at worst make each other
// recompile; there is no state here to corrupt.
const moduleCacheName = "feov-pdfium-module-cache"

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — the leak that is invisible on Linux and fails the Windows
// leg. See recordtest.Main (#666).
//
// IT NO LONGER *IS* recordtest.Main, because one more thing has to happen around m.Run: every test
// here builds its own run under its own t.TempDir(), and the PDFium module cache is keyed on the
// run's directory, so each of the 21 tests that touch a PDF was paying the full 3,968 ms compile —
// 68s for a package whose actual work is milliseconds, and past the 10-minute default timeout
// under `-race`. recordtest documents CheckOrphanedHandles as the seam for exactly this case: a
// package that needs its own TestMain keeps the handle check as a post-suite hook rather than
// giving it up.
//
// recordtest.Main calls os.Exit, so the cleanup below cannot be a defer.
func TestMain(m *testing.M) {
	dir := filepath.Join(os.TempDir(), moduleCacheName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "cannot create the shared PDFium module cache: %v\n", err)
		os.Exit(1)
	}
	moduleCacheDir = dir

	code := m.Run()

	os.RemoveAll(dir)
	if err := recordtest.CheckOrphanedHandles(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}
