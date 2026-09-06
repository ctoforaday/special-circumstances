package blue

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The blue package's tests now open real record handles (render_record_test.go), so they must run
// the orphaned-handle guard — a cached DB handle leaked from t.TempDir() instead of
// recordtest.TmpRun passes on Linux and fails the Windows leg. recordtest.Main enforces it.
func TestMain(m *testing.M) { recordtest.Main(m) }
