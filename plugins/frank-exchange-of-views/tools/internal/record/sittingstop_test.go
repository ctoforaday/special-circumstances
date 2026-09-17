package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// registerAs is a register carrying the agent handle the PreToolUse hook injects, which is what
// every register in a hook-reached run carries.
func (b *stage) registerAs(seat, agent string) *stage {
	return b.add(seat, &recordpb.Register{AgentId: proto.String(agent)})
}

// stop is the sitting_close the SubagentStop hook writes when that agent returns: attributed to the
// harness, naming the agent and its configuration and no seat.
func (b *stage) stop(agent string) *stage {
	return b.add(HarnessSeat, &recordpb.SittingClose{AgentId: proto.String(agent), AgentType: proto.String("frank-exchange-of-views:x")})
}

// boardRun is the board the epoch limit lost a ruling on: a lens mints one high gap at its first
// sitting and never mints again, blue answers it every epoch without moving it, and the chair asks
// the record who sits and dispatches exactly that. With stops, every party sitting is a fresh
// subagent — one agent() call — whose stop the hook records when it returns, which is how the
// shipped workflow runs a seat.
type boardRun struct {
	stops     bool // the SubagentStop hook records each party's return
	warm      bool // the chair registers only once
	maxEpochs int  // 0: no epoch limit
}

// boardEnd is where a boardRun stopped: the epoch whose chair sitting ended it and what that
// sitting's plan said.
type boardEnd struct {
	epoch   int
	docket  []string
	ceiling bool
}

func (c boardRun) play(t *testing.T) boardEnd {
	t.Helper()
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	minted := false
	for epoch := 1; epoch <= 12; epoch++ {
		if !c.warm || epoch == 1 {
			b.registerAs("red-chair", "chair-agent")
		}
		run := newStage(t)
		if c.maxEpochs > 0 {
			writeRunConfig(t, run.runDir, fmt.Sprintf(`{"maxEpochs":%d}`, c.maxEpochs))
		}
		run.evs, run.n = append(run.evs, b.evs...), b.n
		plan, err := PlanDispatch(run.seed())
		if err != nil {
			t.Fatal(err)
		}
		if len(plan.Docket) > 0 || plan.Ceiling || plan.PassPermitted || len(plan.Parties) == 0 {
			return boardEnd{epoch: epoch, docket: plan.Docket, ceiling: plan.Ceiling}
		}
		for _, p := range plan.Parties {
			b.dispatch(plan.Head, p.SeatID, p.GapIDs...)
		}
		for _, p := range plan.Parties {
			agent := fmt.Sprintf("agent-%s-e%d", p.SeatID, epoch)
			b.registerAs(p.SeatID, agent)
			if p.SeatID == evLens && !minted {
				b.add(evLens, freshHigh("G1"))
				minted = true
			}
			if c.stops {
				b.stop(agent)
			}
		}
	}
	t.Fatal("the board did not end within 12 epochs")
	return boardEnd{}
}

// THE EPOCH LIMIT LANDING ON THE IMPASSE EPOCH (#1002). Two stalled exchanges put G1 at impasse
// under K=2, and the record holds both by the end of epoch 3: every party that sat has returned.
// The chair sitting that opens epoch 4 is the run's last under an epoch limit of 4, and it must
// docket G1 for the bench — the docket stands at the limit, and the terminal bench rules it. A fold
// that closed a sitting only when the seat sat again saw one exchange at that moment, ended the run
// at CEILING, and left the material gap unruled.
func TestAnEpochLimitOnTheImpasseEpochStillDocketsTheGap(t *testing.T) {
	end := boardRun{stops: true, maxEpochs: 4}.play(t)
	if end.epoch != 4 || len(end.docket) != 1 || end.docket[0] != "G1" {
		t.Fatalf("the board ended at epoch %d with docket %v (ceiling %v); want G1 docketed at epoch 4, the epoch its second stalled exchange ended",
			end.epoch, end.docket, end.ceiling)
	}
}

// The same board with no limit, and under a warm chair: the stop puts impasse on the epoch the
// exchanges ended in either way, and the lens — retired by then for two barren sittings — does not
// keep the gap from reaching the bench.
func TestTheRetiredLensBoardReachesImpasseOnTheEpochItsExchangesEnded(t *testing.T) {
	for _, c := range []struct {
		name string
		run  boardRun
		want int
	}{
		{"cold chair, stops recorded", boardRun{stops: true}, 4},
		{"warm chair, stops recorded", boardRun{stops: true, warm: true}, 4},
		// No stop on the record — a headless seat, or a run the hook never reached: each sitting
		// closes when the seat sits again, one epoch later, and the gap still reaches the bench.
		{"cold chair, no stops", boardRun{}, 5},
	} {
		t.Run(c.name, func(t *testing.T) {
			end := c.run.play(t)
			if end.epoch != c.want || len(end.docket) != 1 || end.docket[0] != "G1" {
				t.Fatalf("ended at epoch %d with docket %v (ceiling %v), want G1 docketed at epoch %d", end.epoch, end.docket, end.ceiling, c.want)
			}
		})
	}
}

// THE STOP CLOSES THE SITTING WHEN IT ENDS. Both parties sat once and returned; neither has sat
// again. Their stops are on the record, so the exchange is complete now — it is counted, and its
// silence stalls it — rather than waiting for the parties' next register.
func TestAStopClosesTheSittingBeforeTheSeatSitsAgain(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-2").stop("lens-2").
		registerAs("blue-respond", "blue-1").stop("blue-1")
	run := b.seed()
	x, err := Exchanges(run, DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 1 || g.Stalled != 1 || g.Unresolved != 0 {
		t.Fatalf("G1 = %+v, want the one exchange both stops closed, stalled by its silence, nothing unresolved", g)
	}
	if line := planLine(t, run, "G1"); !strings.Contains(line, "1 exchange(s) (1 stalled)") {
		t.Errorf("the plan's line for G1 is %q, want the counted exchange", line)
	}

	// The window is the sitting's own: movement after blue returned is not blue's answer. The lens
	// regrades G1 after blue's stop and before anyone sits again, so the exchange is still a stall.
	late := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-2").stop("lens-2").
		registerAs("blue-respond", "blue-1").stop("blue-1").
		regrade(evLens, "G1", "medium")
	x, err = Exchanges(late.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 1 || g.Stalled != 1 {
		t.Fatalf("G1 = %+v, want the exchange closed at blue's stop, before the later regrade", g)
	}
}

// A STOP THAT JOINS NO SITTING'S REGISTER CLOSES NOTHING. The hook records every typed subagent in
// a project whose run marker is live — a developer's own general-purpose and plan-auditor agents
// included — and a register with no agent_id has nothing to join. Neither may close a party's
// sitting, and nor may another seat's agent returning.
func TestAStopThatJoinsNoRegisterClosesNothing(t *testing.T) {
	stranger := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).registerAs("blue-respond", "blue-1").
		stop("af4ca9baf79c3eb22").stop("aa81ad8fc6308b432").stop("")
	x, err := Exchanges(stranger.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 2 {
		t.Fatalf("G1 = %+v, want nothing closed — no stop joins either register", g)
	}

	// Blue's agent returned; the lens's has not. Blue's stop closes blue's sitting and not the
	// lens's, so there is still no exchange, and the lens's sitting is the one not measured.
	other := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-1").registerAs("blue-respond", "blue-1").stop("blue-1")
	x, err = Exchanges(other.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 1 {
		t.Fatalf("G1 = %+v, want the lens's sitting still unresolved — another seat's agent returning closes nothing of it", g)
	}

	// The same agent's EARLIER stop does not close a later sitting it registered for.
	earlier := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").registerAs(evLens, "lens").mint(evLens, "G1", "high").stop("lens").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens").registerAs("blue-respond", "blue").stop("blue")
	x, err = Exchanges(earlier.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 1 {
		t.Fatalf("G1 = %+v, want the lens's second sitting unresolved — its agent's stop came before it began", g)
	}
}

// NEITHER A REGISTER NOR A STOP: NOT MEASURED. The parties' registers carry agents and neither
// agent has returned, so the record cannot say either sitting ended, and the line says so — naming
// both facts that would have closed it.
func TestASittingWithNeitherARegisterNorAStopIsNotMeasured(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-2").registerAs("blue-respond", "blue-1")
	run := b.seed()
	x, err := Exchanges(run, DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 2 {
		t.Fatalf("G1 = %+v, want nothing counted and both sittings unresolved", g)
	}
	line := planLine(t, run, "G1")
	if !strings.Contains(line, "NOT MEASURED") || strings.Contains(line, "0 exchange(s)") {
		t.Errorf("the plan's line for G1 is %q, want it not measured rather than zero", line)
	}
	if !strings.Contains(line, "stop") {
		t.Errorf("the plan's line for G1 is %q — a stop closes a sitting too, so the line must not name the register alone", line)
	}
}

func planLine(t *testing.T, run Run, gap string) string {
	t.Helper()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Why {
		if strings.HasPrefix(w, gap+": ") {
			return w
		}
	}
	return ""
}

// WHAT A SITTING NOTHING CLOSED IS DEPENDS ON WHEN THE RECORD IS READ, AND ONLY THAT (#1002, ruling
// 1). Both parties sat once and neither registered again nor returned. While the run runs, the
// sittings may be in flight: the fold counts nothing and blue's sitting is unresolved. After the
// run — capture's read — the end of the record closes both: one exchange, and a closed blue sitting.
// One closer answers both readers, so the two readings cannot drift apart.
func TestTheEndOfTheRecordClosesASittingOnlyAfterTheRun(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().
		register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-2").registerAs("blue-respond", "blue-1")
	run := b.seed()
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := eventIDsOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	if g := exchangesOf(m.Events, ids, DefaultParams, WhileRunning)["G1"]; g == nil || g.Exchanges != 0 || g.Unresolved != 2 {
		t.Errorf("while running: G1 = %+v, want nothing counted and both sittings unresolved", g)
	}
	if g := exchangesOf(m.Events, ids, DefaultParams, AfterTheRun)["G1"]; g == nil || g.Exchanges != 1 || g.Unresolved != 0 {
		t.Errorf("after the run: G1 = %+v, want the one exchange the end of the record closes, nothing unresolved", g)
	}
	if ss := BlueSittings(m.Events, WhileRunning); len(ss) != 1 || !ss[0].Unresolved {
		t.Errorf("while running: blue's sittings = %+v, want one unresolved", ss)
	}
	if ss := BlueSittings(m.Events, AfterTheRun); len(ss) != 1 || ss[0].Unresolved {
		t.Errorf("after the run: blue's sittings = %+v, want one, closed", ss)
	}
}
