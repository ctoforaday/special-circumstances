package cli

import (
	"fmt"
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// This package drives a binary built by testbuild, and that build directory is shared by
// every caller in this test binary — so the process exit is the only point at which removing
// it is safe. Omitting this leaks one directory holding one linked binary per run into
// TMPDIR, which is a tmpfs on the paths that matter. See testbuild.Main.
// It also opens records, and TestMain is a package's ONE hook — so the orphaned-handle check
// rides here rather than being dropped for the build sandbox (#666).
//
// AND IT RENDERS PDFs. The `ocr` tests stub DefaultPageReader — the model call — but not the
// rasteriser, so six of them drive real PDFium through the CLI. The module cache is keyed on the
// run directory and each test makes its own, so each paid the full 3,968 ms compile: ~15s of this
// package's 33s. fetch_test.go's withExtractor says "EVERY cli test needs this"; the ocr tests are
// the ones it does not cover, because what they exercise IS the rendering path. Sharing one
// compiled module is the fix that does not weaken them.
func TestMain(m *testing.M) {
	release, err := fetchcache.UseSharedModuleCache()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	testbuild.Main(m, recordtest.CheckOrphanedHandles, release)
}
