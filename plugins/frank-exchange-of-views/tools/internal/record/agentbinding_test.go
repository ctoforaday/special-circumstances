package record_test

import (
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// registerAs is a seat's register under agent ("" writes none), with occasion where one is given.
func registerAs(t *testing.T, seat, agent string, occasion ...recordpb.Occasion) *recordpb.Event {
	t.Helper()
	reg := &recordpb.Register{}
	if agent != "" {
		reg.AgentId = proto.String(agent)
	}
	if len(occasion) > 0 {
		reg.Occasion = occasion[0].Enum()
	}
	return recordtest.Event(t, seat, reg)
}

// bracketAs is the SubagentStart hook's sitting_open under agent, naming seat when the configuration
// resolves to one ("" for a configuration that seats several, as blue's lanes do).
func bracketAs(t *testing.T, seat, agent string) *recordpb.Event {
	t.Helper()
	open := &recordpb.SittingOpen{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:x")}
	if seat != "" {
		open.SeatId = proto.String(seat)
	}
	return recordtest.Event(t, record.HarnessSeat, open)
}

// boundRun seeds evs into a fresh run and returns it with its bindings as record.SeatBindings reads them.
func boundRun(t *testing.T, evs ...*recordpb.Event) (record.Run, string, map[string]record.SeatBinding) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	run := runtest.Open(t, dir)
	got, err := record.SeatBindings(run)
	if err != nil {
		t.Fatal(err)
	}
	return run, dir, got
}

func wantBindings(t *testing.T, got, want map[string]record.SeatBinding) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("bound %d agents, want %d: %+v", len(got), len(want), got)
	}
	for id, w := range want {
		if got[id] != w {
			t.Errorf("%s = %+v, want %+v", id, got[id], w)
		}
	}
}

// THE EPOCH AND THE SITTING ARE THE ONES THE RECORD STORES EACH REGISTER IN, NOT STAMPED. A lens
// that sits before the chair has ever sat is in epoch 0; the chair's own register opens the epoch
// it is in; a seat's second sitting — here its second register — is its sitting 2 whatever epoch it
// lands in.
func TestSeatBindingsReadTheStoredWindowOfEachRegister(t *testing.T) {
	_, _, got := boundRun(t,
		registerAs(t, "red-lens-evidence", "L1"), // epoch 0, sitting 1
		registerAs(t, "red-chair", "C1"),         // opens epoch 1, sitting 1
		registerAs(t, "blue-respond", "B1"),      // epoch 1, sitting 1
		registerAs(t, "red-chair", "C2"),         // opens epoch 2, sitting 2
		registerAs(t, "red-lens-evidence", "L2"), // epoch 2, sitting 2
		registerAs(t, "judge", ""),               // no agent id: binds nothing
	)
	wantBindings(t, got, map[string]record.SeatBinding{
		"L1": {SeatID: "red-lens-evidence", Epoch: 0, Sitting: 1},
		"C1": {SeatID: "red-chair", Epoch: 1, Sitting: 1},
		"B1": {SeatID: "blue-respond", Epoch: 1, Sitting: 1},
		"C2": {SeatID: "red-chair", Epoch: 2, Sitting: 2},
		"L2": {SeatID: "red-lens-evidence", Epoch: 2, Sitting: 2},
	})
}

// THE LAST REGISTER WINS, as the agent_binding view has it: a resumed seat arrives under a new agent
// id claiming a seat already bound, and an agent that registers twice is bound to its latest claim.
func TestSeatBindingsTakeTheLatestRegister(t *testing.T) {
	_, _, got := boundRun(t, registerAs(t, "red-chair", "A"), registerAs(t, "red-chair", "A"))
	if got["A"] != (record.SeatBinding{SeatID: "red-chair", Epoch: 2, Sitting: 2}) {
		t.Errorf("A = %+v, want the second register's window", got["A"])
	}
}

// A SITTING THE HOOK BRACKETED AND NO REGISTER JOINED IS STILL BOUND, by its bracket: a lens woken
// with nothing to do need not register (#1089). Where a register names the agent too, the register
// wins, whichever came first; a bracket naming no seat binds nothing.
func TestSeatBindingsBindABracketOnlySitting(t *testing.T) {
	_, _, got := boundRun(t,
		bracketAs(t, "red-lens-voice", "V1"),   // bracket only: voice sitting 1, epoch 0
		bracketAs(t, "red-chair", "C1"),        // epoch 1
		registerAs(t, "red-chair", "C1"),       // joins C1's bracket: the register binds
		bracketAs(t, "red-lens-voice", "V2"),   // bracket only: voice sitting 2, epoch 1
		bracketAs(t, "", "LANE"),               // names no seat
		registerAs(t, "red-lens-logic", "X"),   // a register, then a bracket naming another seat:
		bracketAs(t, "red-lens-evidence", "X"), // the register wins
	)
	wantBindings(t, got, map[string]record.SeatBinding{
		"V1": {SeatID: "red-lens-voice", Epoch: 0, Sitting: 1},
		"C1": {SeatID: "red-chair", Epoch: 1, Sitting: 1},
		"V2": {SeatID: "red-lens-voice", Epoch: 1, Sitting: 2},
		"X":  {SeatID: "red-lens-logic", Epoch: 1, Sitting: 1},
	})
}

// THE OCCASION IS THE REGISTER'S. The bench's four sittings share one id, so its register says
// which this one was convened for, and the binding carries it; a seat whose id already says, and an
// agent bound by its bracket, read "" — read as a string, so a NULL that reached the scan would fail
// here rather than silently become a docket ruling.
func TestSeatBindingsCarryTheRegistersOccasion(t *testing.T) {
	_, _, got := boundRun(t,
		registerAs(t, "red-chair", "C1"),
		registerAs(t, "judge", "J1", recordpb.Occasion_OCCASION_TERMINAL),
		registerAs(t, "judge", "J2", recordpb.Occasion_OCCASION_ASSEMBLE),
		bracketAs(t, "red-lens-voice", "V1"),
	)
	for agent, want := range map[string]string{"C1": "", "J1": "terminal", "J2": "assemble", "V1": ""} {
		if b, ok := got[agent]; !ok || b.Occasion != want {
			t.Errorf("%s = %+v (bound %v), want occasion %q", agent, b, ok, want)
		}
	}
}

// A RUN WITH NO RECORD YET BINDS NOTHING, and that is not an error: the cost report and the
// dashboard read a run before its first write.
func TestSeatBindingsOfARunWithNoRecordIsEmpty(t *testing.T) {
	got, err := record.SeatBindings(runtest.Open(t, t.TempDir()))
	if err != nil || got != nil {
		t.Errorf("a run with no record = (%v, %v), want (nil, nil)", got, err)
	}
}

// AN AGENT'S LATEST BRACKET DECIDES, SEAT-NAMING OR NOT. Bracketed once naming a seat and again naming
// none, with no register, the agent is unbound: the later bracket says its configuration seats
// several, and naming the earlier seat would be a guess the record no longer makes. Every reader of
// the binding says so — the rule's view, the windowed view, the hook's lookup and the bulk read.
func TestALaterSeatlessBracketLeavesTheAgentUnbound(t *testing.T) {
	run, dir, got := boundRun(t,
		bracketAs(t, "red-lens-voice", "E"),
		bracketAs(t, "", "E"),
		bracketAs(t, "red-lens-voice", "OK"), // the control: a seat-naming latest bracket binds
	)
	if _, bound := got["E"]; bound {
		t.Errorf("SeatBindings binds E to %+v", got["E"])
	}
	if got["OK"].SeatID != "red-lens-voice" {
		t.Errorf("SeatBindings does not bind the control: %+v", got)
	}
	if seat, found, err := record.SeatOfAgent(run, "E"); err != nil || found {
		t.Errorf("SeatOfAgent(E) = (%q, %v, %v), want unbound", seat, found, err)
	}
	db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}
	for _, view := range []string{"agent_binding", "seat_of_agent"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM "` + view + `" WHERE "agent_id" = 'E'`).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 0 {
			t.Errorf("%s holds %d row(s) for E, want none", view, n)
		}
	}
}
