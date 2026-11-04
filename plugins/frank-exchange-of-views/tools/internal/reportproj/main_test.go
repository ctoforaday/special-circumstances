package reportproj_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records (record.RenderRun), so they run under the record-handle guard:
// any test that leaks an open record.db handle fails here rather than corrupting a later run.
func TestMain(m *testing.M) { recordtest.Main(m) }
