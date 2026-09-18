package dashboard

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE DASHBOARD IS A LIVE VIEW (#1002, finding 3). It renders a run that is still going, so a blue
// sitting with no later register and no stop MAY BE IN FLIGHT — and the record holds no fact that
// tells that apart from one that ended with nothing after it. Read as a finished run, this page
// calls the sitting closed and prints what blue owed as settled while blue is still sitting.
//
// scorecard.Compute takes the mode from its caller, so the library is tested at both readings and
// nothing held THE CALL: flipping this one word left every test green.
func TestTheDashboardReadsTheRunAsStillRunning(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	recordtest.Seed(t, runDir,
		recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.Event(t, "blue-respond", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
	)
	got := scorecardSection(runtest.Open(t, runDir))
	if !strings.Contains(got, "NOT MEASURED") {
		t.Errorf("a sitting that may still be in flight is rendered as closed:\n%s", got)
	}
}
