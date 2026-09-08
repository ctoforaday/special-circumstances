package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// TestMain is the orphaned-handle guard (#698): a package whose tests open runs must close
// them, and the guard is what notices when one does not.
func TestMain(m *testing.M) { recordtest.Main(m) }
