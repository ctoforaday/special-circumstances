package hookcmd

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records — the sitting hooks append to one — so the process must not
// exit while a cached database handle outlives the directory it lived in, the leak that is
// invisible on Linux and fails the Windows leg. See recordtest.Main (#666).
//
// The guard sweep caught this the moment sitting_test.go arrived, which is what it is for.
func TestMain(m *testing.M) { recordtest.Main(m) }
