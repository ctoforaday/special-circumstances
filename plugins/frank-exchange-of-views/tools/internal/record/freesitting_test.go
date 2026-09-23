package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A WOKEN SEAT WITH NOTHING TO DO RUNS NO COMMANDS AT ALL. This is the acceptance test for #1089.
//
// Measured across eight runs and 199 wakeups: 48% recorded nothing and cost as much as the
// productive ones — 46% of every command in the run. Two writes made an empty sitting cost
// anything: `register`, to put the sitting on the record, and `log --type nominal`, to close the
// log channel. Both restate what the hooks already capture at both ends of the sitting.
//
// The seat here NEVER REGISTERS AND NEVER LOGS. Its sitting exists because the SubagentStart hook
// wrote `sitting_open` naming the configuration it was dispatched as, and that configuration seats
// exactly one seat. So the dispatch is satisfied, the log is not owed, and the work list is
// complete — at a cost of zero commands.
// AND THE SEAT CAN READ ITS OWN WORK LIST WITHOUT REGISTERING, which is what makes the free
// sitting reachable rather than merely permitted.
//
// The CLI scopes its surface to whoever is asking, and it asks the record. While that lookup read
// only the register table, an unregistered agent had no identity and got the OPERATOR surface — on
// which `show work` does not exist. A seat could therefore skip `register` only by never looking,
// and could only learn it need not look BY looking. Measured: 8 of 9 sittings with nothing owed
// registered anyway. The sibling test above proved the RECORD owed nothing; nothing proved the
// seat could find that out.
func TestAHookOpenedSeatIsIdentifiedWithoutRegistering(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).
		add(HarnessSeat, &recordpb.SittingOpen{
			AgentId:   proto.String("agent-lens-1"),
			AgentType: proto.String("frank-exchange-of-views:red-lens-evidence"),
		}).seed()

	seat, found, err := SeatOfAgent(run, "agent-lens-1")
	if err != nil {
		t.Fatalf("SeatOfAgent: %v", err)
	}
	if !found || seat != evLens {
		t.Errorf("a hook-opened agent resolved to (%q, %v) — it must resolve to %q, or the seat is "+
			"handed the operator surface and has to register to see its own work list", seat, found, evLens)
	}

	// AN AGENT THE HARNESS NEVER BRACKETED IS STILL UNBOUND. The fallback reads a recorded
	// bracket, never a guess from the ambient environment.
	if _, found, err := SeatOfAgent(run, "agent-nobody"); err != nil || found {
		t.Errorf("an agent with no register and no bracket resolved to a seat (found=%v, err=%v)", found, err)
	}
}

func TestAWokenLensWithNothingToDoOwesNothingAndRunsNoCommands(t *testing.T) {
	hookOpened := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).
			add(HarnessSeat, &recordpb.SittingOpen{
				AgentId:   proto.String("agent-lens-1"),
				AgentType: proto.String("frank-exchange-of-views:red-lens-evidence"),
			})
	}

	s := sittingOfRunT(t, hookOpened().seed(), "lens", evLens)
	if owed := owedItems(s); len(owed) != 0 {
		t.Errorf("a hook-opened sitting still owes %+v — the seat would have to run a command to discharge it", owed)
	}
	if !s.Complete {
		t.Errorf("a hook-opened empty sitting is not complete, so the seat cannot end its turn: %+v", s.Open)
	}

	// AND THE SITTING COUNTED, which is the half that keeps the saving from becoming a loop.
	// `dispatch` readies a seat until it has SAT: if a hook-opened sitting did not count, the lens
	// would be re-dispatched for the same unsat dispatch forever.
	//
	// THE LENS IS STILL READIED, AND THAT IS CORRECT — an ACTIVE lens is ready every epoch by
	// design (retirement.go). What this asserts is the REASON: the plan must ready it as an active
	// lens, never as one that has not sat. The retirement fold counting a barren sitting is the
	// proof the hook-opened sitting reached every reader through the same ledger.
	plan, err := PlanDispatch(hookOpened().seed())
	if err != nil {
		t.Fatal(err)
	}
	why := strings.Join(plan.Why, " | ")
	if strings.Contains(why, "has not registered") || strings.Contains(why, "not sat") {
		t.Errorf("the lens sat (hook-opened) and dispatch still treats it as unsat: %s", why)
	}
	if !strings.Contains(why, "barren sitting") {
		t.Errorf("the retirement fold did not count the hook-opened sitting, so it reached the ledger "+
			"for the work list and not for the fold: %s", why)
	}
}

// BLUE STILL PAYS, AND THE TABLE SAYS WHY. `blue-researcher` is dispatched as blue-lane-N,
// blue-respond and frontier, so a sitting_open naming it cannot say which of the three sat. The
// honest answer is that the sitting is not knowable from the hook, and blue's register stays
// load-bearing — not a guess at the most likely seat.
func TestBlueGetsNoFreeSittingBecauseItsConfigurationSeatsThree(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, "blue-respond", "G1").
		add(HarnessSeat, &recordpb.SittingOpen{
			AgentId:   proto.String("agent-blue-1"),
			AgentType: proto.String("frank-exchange-of-views:blue-researcher"),
		})
	s := sittingOfRunT(t, st.seed(), "blue", "blue-respond")
	if len(owedItems(s)) == 0 {
		t.Error("blue's sitting was satisfied by a hook that cannot know which of three seats sat")
	}
}

// AN UNKNOWN CONFIGURATION OPENS NOTHING. While a run is live the SubagentStart hook attributes
// every subagent in that project to it — a dev session's included. A type this build has not been
// taught must resolve to no seat rather than forge a sitting for one.
func TestAnUnknownConfigurationOpensNoSitting(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).
		add(HarnessSeat, &recordpb.SittingOpen{
			AgentId:   proto.String("agent-stranger"),
			AgentType: proto.String("some-other-plugin:general-purpose"),
		})
	s := sittingOfRunT(t, st.seed(), "lens", evLens)
	if len(owedItems(s)) == 0 {
		t.Error("a subagent from another plugin opened a sitting for a lens that never ran")
	}
}
