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

// A RECORD WRITTEN BEFORE THE GATE RECORDED ITS ADMISSIONS IS ADMITTED AT EVERY CLOSE, so no past
// sitting reads as failed. universe-m35 (epoch 45) is the run whose engine died on a seat: blue's
// fifth response had its envelope refused twice and ended with text. Each admission migrate writes
// answers exactly one sitting that would otherwise read failed, and after it none does.
//
// THE RESIDUE, ACCEPTED: the source cannot tell that sitting from one that finished, so blue's
// fifth sitting reads as finished — it is admitted like every other.
func TestAMigratedRecordAdmitsEveryClosedSitting(t *testing.T) {
	_, run := migrateArchive(t, "2026-10-27_is-91-prime-m35.tar.gz")
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	admitted := map[string]bool{}
	var without []*record.Event
	for _, e := range m.Events {
		if a, ok := recordpb.BodyAs[*recordpb.EnvelopeAdmitted](e); ok {
			if e.GetSeatId() != record.HarnessSeat {
				t.Errorf("a migrated admission is attributed to %q, want the harness", e.GetSeatId())
			}
			admitted[a.GetAgentId()] = true
			continue
		}
		without = append(without, e)
	}
	if len(admitted) == 0 {
		t.Fatal("the migrated record holds no admission, so nothing here exercised the translation")
	}
	if f := record.FailedSittingsOf(m.Events, m.At); len(f) != 0 {
		t.Errorf("%d migrated sitting(s) read failed, the first %+v — every closed sitting of the source is admitted", len(f), f[0])
	}
	if owed := record.FailedSittingsOf(without, m.At); len(owed) != len(admitted) {
		t.Errorf("migrate wrote %d admission(s) for %d sitting(s) that read failed without them — one per sitting, and none for any other", len(admitted), len(owed))
	}
	// Blue's fifth response — the sitting the engine died on — is the residue.
	var last string
	for _, e := range m.Events {
		if r, ok := recordpb.BodyAs[*recordpb.Register](e); ok && e.GetSeatId() == "blue-respond" {
			last = r.GetAgentId()
		}
	}
	if last == "" || !admitted[last] {
		t.Errorf("blue's last sitting (agent %q) is not admitted; the migration states it reads as finished", last)
	}
}

// A SOURCE THAT DECLARES THE ADMISSION IS GIVEN NONE: a sitting it closed with no admission failed,
// and stays failed.
func TestASourceThatRecordsAdmissionsKeepsItsFailedSittings(t *testing.T) {
	from := recordtest.TmpRun(t)
	recordtest.Seed(t, from,
		recordtest.Stamped(recordtest.At(t, record.HarnessSeat, "cast", &recordpb.Cast{SeatIds: []string{"red-chair"}}), ts1),
		recordtest.Stamped(recordtest.At(t, record.HarnessSeat, "open", &recordpb.SittingOpen{AgentId: proto.String("C1"),
			AgentType: proto.String("frank-exchange-of-views:red-chair"), SeatId: proto.String("red-chair")}), ts2),
		recordtest.Stamped(recordtest.At(t, record.HarnessSeat, "close", &recordpb.SittingClose{AgentId: proto.String("C1")}), ts3))
	src, err := migrate.OpenSQLite(filepath.Join(from, "records"), t.TempDir())
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
		t.Fatalf("refusals on a fixture built to land: %+v", res.Refusals)
	}
	f, err := record.FailedSittings(dst)
	if err != nil {
		t.Fatal(err)
	}
	if len(f) != 1 || f[0].SeatID != "red-chair" || f[0].AgentID != "C1" {
		t.Errorf("failed sittings = %+v, want the chair's one sitting, closed with no admission", f)
	}
}
