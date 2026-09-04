package surface

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — the leak that is invisible on Linux and fails the Windows
// leg. See recordtest.Main (#666). The guard came for free while these gates lived in
// integration/fuzz, whose TestMain carries it; the split (#694) moved the tests and left the
// guard behind, which is exactly the tenth mistake the guard sweep exists to catch — and did.
func TestMain(m *testing.M) { recordtest.Main(m) }
