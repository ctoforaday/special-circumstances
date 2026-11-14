package surface

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// This package's tests build hook binaries and open records, so the process must remove the shared
// build directory at exit and must not exit while a cached database handle outlives the directory
// it lived in — the leak that is invisible on Linux and fails the Windows leg. See testbuild.Main
// and recordtest.CheckOrphanedHandles (#666).
func TestMain(m *testing.M) { testbuild.Main(m, recordtest.CheckOrphanedHandles) }
