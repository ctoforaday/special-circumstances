package view

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// inputT is the render input over evs as the write path stores them: every epoch and `seat #N` a
// view prints is the stored window, so a view's fixture is seeded and loaded back.
func inputT(t *testing.T, evs ...*record.Event) Input {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	in, err := InputOf(runtest.Open(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	return in
}
