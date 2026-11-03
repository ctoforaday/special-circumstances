package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// AN ARCHIVED REPORT IS ITS RECORD'S ONE PART. The archive's report was written whole by its synthesis
// seat and recorded once; migrated, it is the record's one part, by the seat that recorded it, holding
// the text as written (or with its sections put in order, where the archive's were not), and the first
// thing the report's fold reads.
func TestAnArchivedReportIsTheMigratedRecordsOnePart(t *testing.T) {
	const archive = "2026-10-27_is-91-prime-m35.tar.gz"
	var seat, written string
	src := openRO(t, recordtest.RequireArchive(t, archive))
	if err := src.QueryRow(`SELECT e."seat_id", b."text" FROM "base_ingest" b JOIN "events" e ON e."id" = b."event_id"`).Scan(&seat, &written); err != nil {
		t.Fatal(err)
	}
	m, dst := migrateArchive(t, archive)
	if len(m.Refusals) != 0 {
		t.Fatalf("%d event(s) refused migrating m35; the first: %+v", len(m.Refusals), m.Refusals[0])
	}
	f, err := record.FamilyOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	var parts []*recordpb.Event
	for _, e := range f.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REPORT_PART {
			parts = append(parts, e)
		}
	}
	if len(parts) != 1 {
		t.Fatalf("the migrated m35 holds %d part(s); its archive recorded its report once", len(parts))
	}
	if parts[0].GetSeatId() != seat {
		t.Errorf("the part is %s's; the archive's report was recorded by %s", parts[0].GetSeatId(), seat)
	}
	p, _ := recordpb.BodyAs[*recordpb.ReportPart](parts[0])
	if got := p.GetText(); got != written && got != report.Ordered(written) {
		t.Errorf("the part holds %d byte(s) that are neither the archive's %d nor them in order", len(got), len(written))
	}
	ops, err := record.ReportProjection(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(ops) == 0 || ops[0].Kind != "append" {
		t.Fatalf("the migrated record's first report op is not its part: %+v", ops[:min(1, len(ops))])
	}
	for i, op := range ops[1:] {
		if op.Kind == "append" {
			t.Errorf("report op %d is a second part", i+1)
		}
	}
}
