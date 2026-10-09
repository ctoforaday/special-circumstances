package sittingwrite

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A CALL THE HOOK DENIED, OR ONE THAT FAILED, IS LOGGED AS THE SEAT'S — under the seat the agent registered as,
// as the tool's entry — and an agent that never registered is no seat, so nothing is written.
func TestAHookRefusalIsLoggedUnderTheRefusedSeat(t *testing.T) {
	runDir := newRun(t)
	registerAs(t, runDir, "agent-1", "red-lens-adversary")
	run := mustRun(t, runDir)
	if err := WriteToolEntry(runDir, Refusal, "agent-1", "refused a `Read` the run does not serve: the report is not a file a seat reads"); err != nil {
		t.Fatal(err)
	}
	if err := WriteToolEntry(runDir, Failure, "agent-1", "`Read` failed on `/x`: File does not exist <!--fx:F-1234abcd-->"); err != nil {
		t.Fatal(err)
	}
	if err := WriteToolEntry(runDir, Refusal, "agent-nobody", "refused a `Read`"); err != nil {
		t.Fatalf("an unregistered agent is no seat and is skipped, not an error: %v", err)
	}
	evs, _, err := record.EventsOf(run, recordpb.EventType_EVENT_TYPE_LOG)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want the registered seat's refusal and failure, got %d", len(evs))
	}
	for i, want := range []recordpb.LogType{recordpb.LogType_LOG_TYPE_REFUSAL, recordpb.LogType_LOG_TYPE_FAILURE} {
		l, _ := recordpb.BodyAs[*recordpb.Log](evs[i])
		if evs[i].GetSeatId() != "red-lens-adversary" || l.GetType() != want || l.GetSource() != recordpb.LogSource_LOG_SOURCE_TOOL {
			t.Errorf("entry %d is not the tool's %v under the seat: %s %v", i, want, evs[i].GetSeatId(), l)
		}
		if strings.Contains(l.GetText(), "<!--") {
			t.Errorf("a live anchor reached the log: %q", l.GetText())
		}
	}
}
