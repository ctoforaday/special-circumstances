package sittingwrite

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A CALL THE HOOK DENIED IS LOGGED AS THE SEAT'S REFUSAL, under the seat the agent registered as,
// as the tool's entry — and an agent that never registered is no seat, so nothing is written.
func TestAHookRefusalIsLoggedUnderTheRefusedSeat(t *testing.T) {
	runDir := newRun(t)
	registerAs(t, runDir, "agent-1", "red-lens-adversary")
	run := mustRun(t, runDir)
	if err := WriteRefusal(runDir, "agent-1", "refused a `Read` the run does not serve: the report is not a file a seat reads"); err != nil {
		t.Fatal(err)
	}
	if err := WriteRefusal(runDir, "agent-nobody", "refused a `Read`"); err != nil {
		t.Fatalf("an unregistered agent is no seat and is skipped, not an error: %v", err)
	}
	evs, err := record.EventsOf(run, recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("want the one registered seat's refusal, got %d", len(evs))
	}
	l, _ := recordpb.BodyAs[*recordpb.Log](evs[0])
	if evs[0].GetSeatId() != "red-lens-adversary" || l.GetType() != recordpb.LogType_LOG_TYPE_REFUSAL || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
		t.Errorf("the refusal is not the tool's refusal under the refused seat: %s %v", evs[0].GetSeatId(), l)
	}
}
