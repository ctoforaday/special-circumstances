package runtest

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// Family seeds evs through the shipped write path into a fresh run and returns them as the family a
// presenter takes, with gaps as given and the events as loaded — each with the sitting the write
// path stored it in (record.WindowIndex). A fold that reads which sitting an act belongs to asks
// that index, and a family built by hand has none: the fold panics rather than read a zero. The
// returned events are the loaded ones, not evs.
func Family(t *testing.T, gaps []*record.Gap, evs ...*record.Event) record.Family {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	m, err := record.MergedEvents(Open(t, dir))
	if err != nil {
		t.Fatalf("runtest: loading the seeded record: %v", err)
	}
	f := record.NewFamily(gaps, m.Events)
	f.At = m.At
	return f
}
