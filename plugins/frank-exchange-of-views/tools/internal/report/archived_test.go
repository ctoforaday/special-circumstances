package report

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
