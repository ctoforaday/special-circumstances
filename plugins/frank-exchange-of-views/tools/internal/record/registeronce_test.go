package record_test

// A REGISTER IS ONCE PER DISPATCH, AND ITS DISPATCH NUMBER IS THE SITTING'S RANK.
//
// Measured on universe-m31: four sittings held two registers from one agent — the seat's first act
// registered it silently, then the seat ran `register` itself — and each repeat advanced the
// dispatch count, so the chair's 12th sitting printed "dispatch 13".

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
)

func registersOf(t *testing.T, run record.Run, seat string) int {
	t.Helper()
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range fam.Events {
		if e.GetSeatId() == seat && e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
			n++
		}
	}
	return n
}

func eventCount(t *testing.T, run record.Run) int {
	t.Helper()
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return len(fam.Events)
}

// THE M31 CHAIR, REPLAYED: twelve bracketed sittings, a register in each, and a second register in
// the 8th and the 12th. Every answer is the sitting's own number, the repeat answers what the first
// did and writes nothing, and the 12th sitting says 12.
func TestARepeatedRegisterAnswersTheStoredOneAndWritesNothing(t *testing.T) {
	dir, run := bracketedRun(t)
	chair := record.Identity{Run: run, SeatID: "red-chair"}
	for i := 1; i <= 12; i++ {
		agent := "agent_chair_" + string(rune('a'+i))
		bracket(t, dir, chairConfig, agent)
		t.Setenv(seatenv.AgentVar, agent)
		first, _, err := record.RegisterSeat(chair, "", "")
		if err != nil {
			t.Fatalf("sitting %d: %v", i, err)
		}
		if first != i {
			t.Fatalf("sitting %d registered as dispatch %d", i, first)
		}
		if i != 8 && i != 12 {
			continue
		}
		before := eventCount(t, run)
		again, _, err := record.RegisterSeat(chair, "", "")
		if err != nil {
			t.Fatalf("sitting %d, the repeat: %v", i, err)
		}
		if again != first {
			t.Errorf("sitting %d: the repeat answered dispatch %d, the register it repeats said %d", i, again, first)
		}
		if after := eventCount(t, run); after != before {
			t.Errorf("sitting %d: the repeat wrote %d event(s)", i, after-before)
		}
	}
	if n := registersOf(t, run, "red-chair"); n != 12 {
		t.Errorf("twelve sittings hold %d registers", n)
	}
	if n, err := record.SittingsOf(run, "red-chair"); err != nil || n != 12 {
		t.Errorf("SittingsOf = %d, %v; want 12", n, err)
	}
}

// A SITTING THE BRACKET OPENED THAT THE SEAT NEVER REGISTERED IN IS STILL A DISPATCH OF IT, so the
// next register's number counts it: the number is the sitting's rank, not a count of registers.
func TestTheDispatchNumberCountsSittingsNotRegisters(t *testing.T) {
	dir, run := bracketedRun(t)
	chair := record.Identity{Run: run, SeatID: "red-chair"}
	bracket(t, dir, chairConfig, "agent_chair_1")
	bracket(t, dir, chairConfig, "agent_chair_2")
	bracket(t, dir, chairConfig, "agent_chair_3")
	t.Setenv(seatenv.AgentVar, "agent_chair_3")
	if d, _, err := record.RegisterSeat(chair, "", ""); err != nil || d != 3 {
		t.Errorf("the third sitting's register = dispatch %d, %v; want 3", d, err)
	}
}

// WITH NO BRACKET THE FIRST REGISTER OPENS THE SITTING, and the agent's repeat joins nothing new: one
// sitting, one register. This is the shape a seat of blue's researcher configuration has.
func TestARepeatedRegisterOpensNoSecondSitting(t *testing.T) {
	_, run := bracketedRun(t)
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	t.Setenv(seatenv.AgentVar, "agent_blue_1")
	for range 2 {
		if d, _, err := record.RegisterSeat(blue, "", ""); err != nil || d != 1 {
			t.Fatalf("register = dispatch %d, %v; want 1", d, err)
		}
	}
	if n, err := record.SittingsOf(run, "blue-respond"); err != nil || n != 1 {
		t.Errorf("SittingsOf = %d, %v; want 1", n, err)
	}
	if n := registersOf(t, run, "blue-respond"); n != 1 {
		t.Errorf("%d registers on the record", n)
	}
}

// A REGISTER NO AGENT ID TIES TO A DISPATCH OPENS A SITTING EACH TIME: nothing says it repeats one.
func TestARegisterWithNoAgentOpensASittingEachTime(t *testing.T) {
	_, run := bracketedRun(t)
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	for want := 1; want <= 2; want++ {
		if d, _, err := record.RegisterSeat(blue, "", ""); err != nil || d != want {
			t.Fatalf("register = dispatch %d, %v; want %d", d, err, want)
		}
	}
}

// THE BENCH'S REPEAT IS THE SAME REGISTRATION ONLY UNDER THE SAME OCCASION. A second word for one
// dispatch contradicts the record, and is refused without writing.
func TestABenchRepeatNamingAnotherOccasionIsRefused(t *testing.T) {
	dir, run := bracketedRun(t)
	bench := record.Identity{Run: run, SeatID: "judge"}
	bracket(t, dir, benchConfig, "agent_bench_1")
	t.Setenv(seatenv.AgentVar, "agent_bench_1")
	if _, _, err := record.RegisterSeat(bench, "", "docket"); err != nil {
		t.Fatal(err)
	}
	if d, _, err := record.RegisterSeat(bench, "", "docket"); err != nil || d != 1 {
		t.Errorf("the same occasion again = dispatch %d, %v; want the stored dispatch 1", d, err)
	}
	before := eventCount(t, run)
	_, _, err := record.RegisterSeat(bench, "", "petition")
	if err == nil || !strings.Contains(err.Error(), `"docket"`) {
		t.Errorf("a repeat naming another occasion = %v; want it refused naming the stored one", err)
	}
	if after := eventCount(t, run); after != before {
		t.Errorf("the refused repeat wrote %d event(s)", after-before)
	}
}
