package sittingwrite

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests open records, so the process must not exit while a cached database handle
// outlives the directory it lived in — invisible on Linux, a Windows-leg failure. See
// recordtest.Main (#666).
func TestMain(m *testing.M) { recordtest.Main(m) }
