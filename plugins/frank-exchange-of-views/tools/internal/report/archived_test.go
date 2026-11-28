package report

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// archived writes body as an archived record holds it, through the replay writer: the shape is
// asked, and none of the live write's checks on who files when. A fixture whose subject is how the
// report renders a docketed gap seeds its docket motion this way — a no-act docket on a gap the
// filer minted is a record migrated from before docket motions proposed acts, and the live write
// refuses it (F1).
func archived(t *testing.T, runDir, seatID string, body proto.Message) {
	t.Helper()
	run := runtest.Open(t, runDir)
	b, err := record.BeginReplay(run)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := b.Write(record.Identity{Run: run, SeatID: seatID}, body, time.Now()); err != nil {
		_ = b.Rollback()
		t.Fatalf("replaying %s/%T: %v", seatID, body, err)
	}
	if err := b.Commit(); err != nil {
		t.Fatal(err)
	}
}

// benchSits opens the bench's sitting of occasion, as its register does.
func benchSits(t *testing.T, runDir, occasion string) {
	t.Helper()
	if _, _, err := record.RegisterSeat(record.Identity{Run: runtest.Open(t, runDir), SeatID: "judge"}, "", occasion); err != nil {
		t.Fatalf("the bench's %s sitting: %v", occasion, err)
	}
}

// benchConvenedOn seats the bench in the docket sitting the chair's dispatch convenes on gaps: a
// remand is ruled where a plan follows it (partiesSitAfter), never at the assembly. The cast the
// dispatch row is admitted against is seeded first; the chair has sat.
func benchConvenedOn(t *testing.T, runDir string, gaps ...string) {
	t.Helper()
	recordtest.Seed(t, runDir, recordtest.At(t, record.HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: []string{"red-chair", "blue-synthesize", "blue-respond", "judge"}}))
	if _, err := record.Append(record.Identity{Run: runtest.Open(t, runDir), SeatID: "red-chair"},
		&recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("judge"), GapIds: gaps}); err != nil {
		t.Fatalf("convening the bench on %v: %v", gaps, err)
	}
	benchSits(t, runDir, "docket")
}
