package migrate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A PLAN STORED BEFORE THE PLAN PRINTED ITS VERDICT carries, migrated, the verdict its chair sitting
// had recorded at that place — the record's own fact, read as `dispatch next` reads it — and a plan
// stored before the sitting recorded one carries none.
func TestAStoredPlanCarriesTheVerdictItsSittingHadRecorded(t *testing.T) {
	setup := func(dir string) []*recordpb.Event {
		at := func(seat, key string, body proto.Message) *recordpb.Event {
			return recordtest.Stamped(recordtest.At(t, seat, key, body), record.Stamp())
		}
		return []*recordpb.Event{
			at("harness", "c", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
			at("harness", "b", &recordpb.ReportPart{Text: proto.String("# r\n\nA sentence.")}),
			at("red-chair", "r1", &recordpb.Register{}),
		}
	}
	// The plan as the verb printed it before it printed a verdict: no `verdict` key.
	asked := recordtest.TmpRun(t)
	recordtest.Seed(t, asked, setup(asked)...)
	plan, _, err := record.DispatchNext(record.Identity{Run: runtest.Open(t, asked), SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	printed := plan.Printed()
	if strings.Contains(printed, `"verdict"`) {
		t.Fatalf("fixture: the plan asked before any verdict prints one: %s", printed)
	}

	dir := recordtest.TmpRun(t)
	evs := setup(dir)
	stamp := func(seat, key string, body proto.Message) *recordpb.Event {
		return recordtest.Stamped(recordtest.At(t, seat, key, body), record.Stamp())
	}
	fail := recordpb.Verdict_VERDICT_FAIL
	evs = append(evs,
		stamp("red-chair", "p1", &recordpb.Plan{Printed: proto.String(printed)}),
		stamp("red-chair", "v", &recordpb.Gate{Verdict: &fail}),
		stamp("red-chair", "p2", &recordpb.Plan{Printed: proto.String(printed)}))
	recordtest.Seed(t, dir, evs...)

	src, err := migrate.OpenSQLite(filepath.Join(dir, "records"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst := runtest.New(t, recordtest.TmpRun(t))
	res, err := migrate.Replay(src, migrate.Entries(), dst, migrate.Options{AcceptLoss: map[string]bool{}})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Refusals) != 0 {
		t.Fatalf("refusals: %+v", res.Refusals)
	}
	after := storedPlans(t, dst)
	if len(after) != 2 {
		t.Fatalf("migrated %d plans, want 2", len(after))
	}
	if after[0].Printed != printed {
		t.Errorf("the plan stored before the verdict changed: %s", after[0].Printed)
	}
	if after[1].Plan.Verdict != "FAIL" || after[1].Plan.Key != plan.Key {
		t.Errorf("the plan stored after the FAIL = verdict %q key %s, want FAIL under the same key %s", after[1].Plan.Verdict, after[1].Plan.Key, plan.Key)
	}
}
