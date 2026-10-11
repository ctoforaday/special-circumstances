package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/testbuild"
)

// TestMain is the orphaned-handle guard (#698): a package whose tests open runs must close
// them, and the guard is what notices when one does not. The migrated-archive store is removed
// FIRST, so a handle the store left open is one the guard then finds without its file.
func TestMain(m *testing.M) {
	testbuild.Main(m, migrate.RemoveArchiveStore, recordtest.CheckOrphanedHandles)
}
