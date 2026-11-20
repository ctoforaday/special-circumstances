package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A MIGRATED REPORT READ STANDS AT THE HEAD ITS TEXT WAS SERVED AT: the latest head move — an edit
// or a part — before it on the migrated record. Asserted per archive.
func TestAMigratedReportReadStandsAtTheLatestHeadMoveBeforeIt(t *testing.T) {
	reads := 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			_, dir := MigratedArchive(t, name)
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			ids := m.At.IDs(m.Events)
			var head int64
			for i, e := range m.Events {
				if recordsql.IsHeadMove(e.GetType()) {
					head = ids[i]
					continue
				}
				rr, ok := recordpb.BodyAs[*recordpb.ReportRead](e)
				if !ok {
					continue
				}
				reads++
				if rr.GetHead() != head {
					t.Errorf("the report read at %d names head %d; the latest head move before it is %d", ids[i], rr.GetHead(), head)
				}
			}
		})
	}
	if reads == 0 {
		t.Fatal("no archived run holds a report read, so no head was checked")
	}
}
