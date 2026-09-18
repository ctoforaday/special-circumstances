package capture

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// unclosedBlueSitting is the record of blue dispatched onto G1, registering while it was open, and
// editing to answer it — with no later register and no stop. Nothing about blue closes the sitting,
// so the two readings differ: running, what it owed is NOT MEASURED; after the run, the end of the
// record closes it and the owed row is scored.
func unclosedBlueSitting(t *testing.T) (record.Run, *record.Family) {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.Event(t, "blue-respond", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
	)
	run := runtest.Open(t, runDir)
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return run, &fam
}

// CAPTURE READS A FINISHED RUN (#1002, finding 3). It runs once the run has ended, so no sitting is
// still in flight and the end of the record closes the last one. Read as live instead, every card
// this writes carries NOT MEASURED for a sitting that simply ended — and these cards are appended to
// the feov-memory series, where that reading outlives the run that produced it.
//
// The mode is a parameter of scorecard.Compute, so the library is already tested at both readings;
// what nothing held was THE CALL. Flipping this one word left every test green.
func TestCaptureWritesTheScorecardsOfAFinishedRun(t *testing.T) {
	run, fam := unclosedBlueSitting(t)
	memory := t.TempDir()
	if r := WriteScorecards(run, nil, memory, fam); !r.Written {
		t.Fatalf("the cards were not written: %+v", r)
	}
	b, err := os.ReadFile(filepath.Join(memory, "blue-scorecard.md"))
	if err != nil {
		t.Fatal(err)
	}
	card := string(b)
	if strings.Contains(card, "NOT MEASURED") {
		t.Errorf("capture read the finished run as still running:\n%s", card)
	}
	if !strings.Contains(card, "1 of 1 owed gap(s) carry no row: G1") {
		t.Errorf("the owed gap is not scored:\n%s", card)
	}
}
