package main

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// This package's tests import the record — the vocabulary region is rendered from its enums and
// record.Plan — so the process carries the orphaned-handle guard every such package does. See
// recordtest.Main (#666).
func TestMain(m *testing.M) { recordtest.Main(m) }
