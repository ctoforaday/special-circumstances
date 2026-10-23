package lens

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The lens package's tests name record types, which the handle-guard sweep counts as opening a
// record; recordtest.Main runs the orphaned-handle guard over them.
func TestMain(m *testing.M) { recordtest.Main(m) }
