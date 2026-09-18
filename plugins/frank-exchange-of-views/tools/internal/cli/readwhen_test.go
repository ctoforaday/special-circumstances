package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A SEAT AND THE OPERATOR BOTH READ A RUNNING RUN (#1002, finding 3). Blue is dispatched onto G1,
// registers while it is open and edits to answer it, with no later register and no stop: nothing
// about blue closes the sitting, and whether it is in flight or simply ended is a thing the record
// does not hold. Both of these verbs are asked DURING the run, so both say so; read as a finished
// run they would call the in-flight sitting closed and print what it owed as settled.
//
// scorecard.Compute takes the mode from its caller, so the library is tested at both readings and
// nothing held THE CALL: flipping the word at either verb left every test green.
func TestTheLiveScorecardVerbsReadTheRunAsStillRunning(t *testing.T) {
	runDir := newRun(t)
	recordtest.Seed(t, runDir,
		recordtest.Event(t, "red-chair", &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		recordtest.Event(t, "blue-respond", &recordpb.Register{AgentId: proto.String("blue-a")}),
		recordtest.Event(t, "blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1")}),
	)
	for _, tc := range []struct {
		name string
		argv []string
	}{
		{"the seat's own card", []string{"show", "scorecard", "--seat-id", "blue-respond"}},
		{"the operator's", []string{"scorecard", "--seat-id", "operator", "--card", "blue"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := run(t, append(tc.argv, "--run", runDir)...)
			if err != nil {
				t.Fatalf("%v: %v", tc.argv, err)
			}
			if !strings.Contains(out, "NOT MEASURED") {
				t.Errorf("a sitting that may still be in flight is reported as closed:\n%s", out)
			}
		})
	}
}
