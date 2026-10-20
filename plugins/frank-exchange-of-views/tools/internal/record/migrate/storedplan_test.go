package migrate_test

import (
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A STORED PLAN MIGRATES AS PRINTED WHERE NOTHING RENUMBERS IT. A seeded run drops no event, so
// each plan's head and pins name the same events after migrate and the bytes migrated are the bytes
// stored: the write admits each again, key and all. A record whose replay drops an event has its
// plans translated through newID (TestAStoredPlanCrossesARenumbering).
func TestAStoredPlanMigratesAsPrinted(t *testing.T) {
	dir := recordtest.TmpRun(t)
	at := func(seat, key string, body proto.Message) *recordpb.Event {
		return recordtest.Stamped(recordtest.At(t, seat, key, body), record.Stamp()) // as the write stamps it, which migrate reads
	}
	recordtest.Seed(t, dir,
		at("harness", "c", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		at("harness", "b", &recordpb.BaseIngest{Text: proto.String("# r\n\nA sentence.")}),
		at("red-chair", "r1", &recordpb.Register{}))
	run := runtest.Open(t, dir)
	if _, _, err := record.DispatchNext(record.Identity{Run: run, SeatID: "red-chair"}); err != nil {
		t.Fatal(err)
	}
	before := storedPlans(t, run)
	if len(before) != 1 {
		t.Fatalf("fixture: %d stored plans, want 1", len(before))
	}

	src, err := migrate.OpenSQLite(filepath.Join(dir, "records"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals migrating a record that stores its plan: %+v", res.Refusals)
	}
	after := storedPlans(t, dst)
	if len(after) != 1 || after[0].Printed != before[0].Printed {
		t.Fatalf("migrated plans %+v, want the one stored, as printed: %s", after, before[0].Printed)
	}
}

func storedPlans(t *testing.T, run record.Run) []record.StoredPlan {
	t.Helper()
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := record.StoredPlans(fam.Events)
	if err != nil {
		t.Fatal(err)
	}
	return stored
}
