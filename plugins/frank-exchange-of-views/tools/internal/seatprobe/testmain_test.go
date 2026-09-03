package seatprobe

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// This package drives a binary built by testbuild, and that build directory is shared by
// every caller in this test binary — so the process exit is the only point at which removing
// it is safe. Omitting this leaks one directory holding one linked binary per run into
// TMPDIR, which is a tmpfs on the paths that matter. See testbuild.Main.
// It also opens records, and TestMain is a package's ONE hook — so the orphaned-handle check
// rides here rather than being dropped for the build sandbox (#666).
func TestMain(m *testing.M) { testbuild.Main(m, recordtest.CheckOrphanedHandles) }
