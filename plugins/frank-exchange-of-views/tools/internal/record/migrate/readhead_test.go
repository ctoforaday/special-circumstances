package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A MIGRATED REPORT READ STANDS AT THE HEAD ITS TEXT WAS SERVED AT (EventSchema 80: a retire moves
// the head). An archived read named the latest edit or part; a retire between that and the read was
// already in the text served, so on the migrated record every read names the latest head move before
// it. Asserted per archive; at least one archive holds a read the retire lifted, or nothing here
// exercised the translation.
func TestAMigratedReportReadStandsAtTheLatestHeadMoveBeforeIt(t *testing.T) {
	reads, lifted := 0, 0
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			_, dir := MigratedArchive(t, name)
			m, err := record.MergedEvents(runtest.Open(t, dir))
			if err != nil {
				t.Fatal(err)
			}
			ids := m.At.IDs(m.Events)
			var head int64
			var headType recordpb.EventType
			for i, e := range m.Events {
				if recordsql.IsHeadMove(e.GetType()) {
					head, headType = ids[i], e.GetType()
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
				if headType == recordpb.EventType_EVENT_TYPE_RETIRE {
					lifted++
				}
			}
		})
	}
	if reads == 0 {
		t.Fatal("no archived run holds a report read, so no head was checked")
	}
	if lifted == 0 {
		t.Error("no migrated report read stands at a retire, so nothing here exercised the translation")
	}
}
