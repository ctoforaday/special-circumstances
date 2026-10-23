package record_test

// A REGISTER IS ONCE PER DISPATCH, AND ITS DISPATCH NUMBER IS THE SITTING'S RANK. A second register
// in one sitting is refused, naming the registration the record holds, and writes nothing.
//
// Measured on universe-m31: four sittings held two registers from one agent — the seat's first act
// registered it silently, then the seat ran `register` itself — and each repeat advanced the
// dispatch count, so the chair's 12th sitting printed "dispatch 13".

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatenv"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/sittingwrite"
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

// refusedRepeat asserts err is the repeat-register refusal naming the stored dispatch.
func refusedRepeat(t *testing.T, err error, dispatch string) {
	t.Helper()
	var fe *feov.Error
	if err == nil || !errors.As(err, &fe) || fe.Code != feov.Conflict ||
		!strings.Contains(err.Error(), "already registered in this sitting ("+dispatch) {
		t.Errorf("the repeat = %v; want a conflict naming %s", err, dispatch)
	}
}

// THE M31 CHAIR, REPLAYED: twelve bracketed sittings, a register in each, and a second register in
// the 8th and the 12th. Every answer is the sitting's own number, the repeat is refused naming the
// dispatch the first answered and writes nothing, and the 12th sitting says 12.
func TestARepeatedRegisterIsRefusedAndWritesNothing(t *testing.T) {
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
		_, _, err = record.RegisterSeat(chair, "", "")
		refusedRepeat(t, err, fmt.Sprintf("dispatch %d", first))
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

// WITH NO BRACKET THE FIRST REGISTER OPENS THE SITTING, and the agent's repeat is refused and opens
// nothing new: one sitting, one register. This is the shape a seat of blue's researcher
// configuration has.
func TestARepeatedRegisterOpensNoSecondSitting(t *testing.T) {
	_, run := bracketedRun(t)
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	t.Setenv(seatenv.AgentVar, "agent_blue_1")
	if d, _, err := record.RegisterSeat(blue, "", ""); err != nil || d != 1 {
		t.Fatalf("register = dispatch %d, %v; want 1", d, err)
	}
	_, _, err := record.RegisterSeat(blue, "", "")
	refusedRepeat(t, err, "dispatch 1")
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

// THE BENCH'S REPEAT IS REFUSED UNDER EITHER OCCASION: under the same one it is a second register, and
// under another a second word for one dispatch that contradicts the record. Neither writes.
func TestABenchRepeatIsRefusedUnderEitherOccasion(t *testing.T) {
	dir, run := bracketedRun(t)
	bench := record.Identity{Run: run, SeatID: "judge"}
	bracket(t, dir, benchConfig, "agent_bench_1")
	t.Setenv(seatenv.AgentVar, "agent_bench_1")
	if _, _, err := record.RegisterSeat(bench, "", "docket"); err != nil {
		t.Fatal(err)
	}
	before := eventCount(t, run)
	_, _, err := record.RegisterSeat(bench, "", "docket")
	refusedRepeat(t, err, `dispatch 1, occasion "docket"`)
	_, _, err = record.RegisterSeat(bench, "", "petition")
	if err == nil || !strings.Contains(err.Error(), `"docket"`) || !strings.Contains(err.Error(), `"petition"`) {
		t.Errorf("a repeat naming another occasion = %v; want it refused naming the stored one", err)
	}
	if after := eventCount(t, run); after != before {
		t.Errorf("the refused repeat wrote %d event(s)", after-before)
	}
}

// A WARM SESSION KEEPS ITS AGENT ID ACROSS SITTINGS, and the stop recorded between them is what
// makes its next register the next sitting rather than a repeat of the last.
func TestARegisterAfterTheAgentsStopOpensTheNextSitting(t *testing.T) {
	dir, run := bracketedRun(t)
	blue := record.Identity{Run: run, SeatID: "blue-respond"}
	t.Setenv(seatenv.AgentVar, "agent_blue_warm")
	if d, _, err := record.RegisterSeat(blue, "", ""); err != nil || d != 1 {
		t.Fatalf("the first sitting's register = dispatch %d, %v; want 1", d, err)
	}
	if err := sittingwrite.Write(sittingwrite.Sitting{RunDir: dir, Phase: sittingwrite.Close, AgentID: "agent_blue_warm", AgentType: "frank-exchange-of-views:blue-researcher"}, io.Discard); err != nil {
		t.Fatal(err)
	}
	if d, _, err := record.RegisterSeat(blue, "", ""); err != nil || d != 2 {
		t.Errorf("the register after the agent's stop = dispatch %d, %v; want 2", d, err)
	}
}
