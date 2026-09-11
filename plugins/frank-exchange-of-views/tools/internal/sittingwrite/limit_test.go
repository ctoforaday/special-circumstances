package sittingwrite

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

// registerAs registers seat under agent, as the hook-injected register does.
func registerAs(t *testing.T, run, agent, seat string) {
	t.Helper()
	t.Setenv(seatenv.AgentVar, agent)
	t.Setenv(seatenv.TypeVar, "")
	if _, _, err := record.RegisterSeat(record.Identity{Run: mustRun(t, run), SeatID: seat}, ""); err != nil {
		t.Fatal(err)
	}
	t.Setenv(seatenv.AgentVar, "")
}

// THE LIMIT LANDS ON THE RECORD, attributed to the harness and naming the seat, the sitting and
// the limit.
func TestASittingLimitLandsNamingTheSeatSittingAndLimit(t *testing.T) {
	run := newRun(t)
	registerAs(t, run, "agent_07", "red-chair")
	registerAs(t, run, "agent_07", "red-chair")
	if err := WriteLimit(run, "agent_07", "frank-exchange-of-views:red-chair", 2, 150); err != nil {
		t.Fatal(err)
	}
	m, err := record.MergedEvents(mustRun(t, run))
	if err != nil {
		t.Fatal(err)
	}
	var found *recordpb.SittingLimit
	for _, e := range m.Events {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_SITTING_LIMIT {
			if e.GetSeatId() != HookSeat {
				t.Errorf("seat_id = %q, want %q — the hook observed this, the seat did not claim it", e.GetSeatId(), HookSeat)
			}
			found = e.GetSittingLimit()
		}
	}
	if found == nil {
		t.Fatal("no sitting_limit event on the record")
	}
	if found.GetSeatId() != "red-chair" || found.GetSitting() != 2 || found.GetLimit() != 150 ||
		found.GetAgentId() != "agent_07" || found.GetAgentType() != "frank-exchange-of-views:red-chair" {
		t.Errorf("sitting_limit = %v, want red-chair sitting 2 limit 150 for agent_07", found)
	}
}

// A SITTING THE RECORD CANNOT ACCOUNT FOR IS REFUSED rather than written.
func TestASittingLimitTheRecordCannotAccountForIsRefused(t *testing.T) {
	run := newRun(t)
	if err := WriteLimit(run, "agent_ghost", "", 1, 150); err == nil || !strings.Contains(err.Error(), "no register") {
		t.Errorf("an agent that never registered: %v, want a refusal naming the missing register", err)
	}
	registerAs(t, run, "agent_07", "red-chair")
	if err := WriteLimit(run, "agent_07", "", 2, 150); err == nil {
		t.Error("sitting 2 of a seat that has opened one was written")
	}
	if err := WriteLimit(run, "agent_07", "", 1, 0); err == nil {
		t.Error("a limit of 0 was written")
	}
}
