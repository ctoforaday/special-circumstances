package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// repairs is a register by seat's agent naming the sitting key opened as the one it repairs.
func (b *stage) repairs(seat, agent, key string) *stage {
	return b.add(seat, &recordpb.Register{AgentId: proto.String(agent), RepairsSitting: proto.String(key)})
}

func (b *stage) position(seat string) *stage {
	return b.add(seat, &recordpb.Position{Text: proto.String("the position")})
}

// lastKey is the key of the event the stage wrote last.
func (b *stage) lastKey() string { return b.evs[len(b.evs)-1].GetKey() }

// THE SHIPPED ENGINE'S SHAPE (#1002, ruling 2). Blue's agent sits, edits, and returns without its
// position or revision; its stop closes the sitting. The engine re-prompts: a NEW agent registers as
// the repair of that sitting, files both, and returns. Its acts are the repaired sitting's, so the
// sitting carries what it owed — while running and after the run — and the exchange fold ends the
// sitting where the repair ended.
func TestARepairRegistersActsCountAsTheSittingItRepairs(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().
		register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").stop("lens-1").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		registerAs(evLens, "lens-2").stop("lens-2").
		registerAs("blue-respond", "blue-a")
	opened := b.lastKey()
	b.edit("G1", "old", "new").stop("blue-a").
		repairs("blue-respond", "blue-b", opened).position("blue-respond").revision("blue-respond").stop("blue-b")
	run := b.seed()
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, when := range []ReadWhen{WhileRunning, AfterTheRun} {
		ss := BlueSittings(m.Events, when)
		if len(ss) != 1 || ss[0].Unresolved {
			t.Fatalf("when=%d: sittings = %+v, want the one sitting, closed by the repair's stop", when, ss)
		}
		if owes := ss[0].Owes(); len(owes) != 0 {
			t.Errorf("when=%d: the repaired sitting still owes %v — the repair's position and revision are its acts", when, owes)
		}
	}
	ids, err := eventIDsOfRun(run)
	if err != nil {
		t.Fatal(err)
	}
	if g := exchangesOf(m.Events, ids, DefaultParams, WhileRunning)["G1"]; g == nil || g.Exchanges != 1 || g.Unresolved != 0 {
		t.Errorf("G1 = %+v, want one exchange: a repair register opens no sitting of its own", g)
	}
	// The repair's register is not a sitting for a dispatch: blue's work list owes no register.
	if _, owed := owedSitting(m.Events, "blue-respond"); owed {
		t.Error("the repair register left blue owing a register — it was counted as a sitting, or as none")
	}
}

// A REPAIR STILL IN FLIGHT LEAVES ITS SITTING UNRESOLVED while the run runs: the repair's agent has
// not returned, so what the sitting filed may still grow. After the run the end of the record closes it.
func TestAnInFlightRepairLeavesTheSittingUnresolvedUntilTheRunEnds(t *testing.T) {
	b := newStage(t).cast("red-chair", "blue-respond").ingest().
		register("red-chair").dispatch(1, "blue-respond", "G1").
		registerAs("blue-respond", "blue-a")
	opened := b.lastKey()
	b.stop("blue-a").repairs("blue-respond", "blue-b", opened).position("blue-respond")
	m, err := MergedEvents(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if ss := BlueSittings(m.Events, WhileRunning); len(ss) != 1 || !ss[0].Unresolved {
		t.Errorf("while running: %+v, want the sitting unresolved while its repair is in flight", ss)
	}
	ss := BlueSittings(m.Events, AfterTheRun)
	if len(ss) != 1 || ss[0].Unresolved || len(ss[0].Owes()) != 1 || ss[0].Owes()[0] != recordpb.EventType_EVENT_TYPE_REVISION {
		t.Errorf("after the run: %+v, want the sitting closed, owing only its revision", ss)
	}
}

// THE WRITER NAMES THE SITTING, AND REFUSES A REPAIR THE RECORD DOES NOT BEAR OUT. The register
// writes the key of the seat's latest opening register, and checkRepair refuses a sitting that is
// not this seat's, not its latest, dispatched past, or owing nothing — so no seat can claim a
// repair it is not doing.
func TestARegisterClaimingARepairIsRefusedUnlessTheRecordBearsItOut(t *testing.T) {
	owing := func(t *testing.T) (*stage, string) {
		b := newStage(t).cast("red-chair", "blue-respond", "blue-synthesize", evLens).ingest().
			register("red-chair").dispatch(1, "blue-respond", "G1").
			registerAs("blue-respond", "blue-a")
		opened := b.lastKey()
		b.edit("G1", "old", "new").stop("blue-a")
		return b, opened
	}
	blue := func(run Run) Identity { return Identity{Run: run, SeatID: "blue-respond"} }

	t.Run("the latest sitting owes its position and revision: the repair is written, naming it", func(t *testing.T) {
		b, opened := owing(t)
		run := b.seed()
		_, repairs, err := RegisterRepair(blue(run), "")
		if err != nil || repairs != opened {
			t.Fatalf("RegisterRepair = %q, %v; want the sitting %q", repairs, err, opened)
		}
		m, err := MergedEvents(run)
		if err != nil {
			t.Fatal(err)
		}
		last := m.Events[len(m.Events)-1]
		if r, ok := recordpb.BodyAs[*recordpb.Register](last); !ok || r.GetRepairsSitting() != opened {
			t.Errorf("the written register = %v, want repairs_sitting %q", last, opened)
		}
	})

	refused := func(t *testing.T, err error, want string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want a refusal saying %q", err, want)
		}
	}
	t.Run("another seat's sitting", func(t *testing.T) {
		b, _ := owing(t)
		b.registerAs(evLens, "lens-1")
		m, _ := MergedEvents(b.seed())
		refused(t, checkRepair(m.Events, "blue-respond", b.lastKey()), "is not a register that opened a sitting of blue-respond")
	})
	t.Run("an earlier sitting than the latest", func(t *testing.T) {
		b, first := owing(t)
		b.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-c").edit("G1", "new", "newer").stop("blue-c")
		m, _ := MergedEvents(b.seed())
		refused(t, checkRepair(m.Events, "blue-respond", first), "opened an earlier sitting of blue-respond")
	})
	t.Run("dispatched again since the sitting began", func(t *testing.T) {
		b, _ := owing(t)
		b.dispatch(2, "blue-respond", "G1")
		_, _, err := RegisterRepair(blue(b.seed()), "")
		refused(t, err, "was dispatched again after the sitting")
	})
	t.Run("the sitting found every gap closed", func(t *testing.T) {
		b := newStage(t).cast("red-chair", "blue-respond", evLens).ingest().
			register("red-chair").registerAs(evLens, "lens-1").mint(evLens, "G1", "high").
			dispatch(1, "blue-respond", "G1").
			closeGap(evLens, "G1").
			registerAs("blue-respond", "blue-a").stop("blue-a")
		_, _, err := RegisterRepair(blue(b.seed()), "")
		refused(t, err, "owes nothing: every gap it was engaged on was closed before it sat")
	})
	t.Run("the sitting already carries both", func(t *testing.T) {
		b := newStage(t).cast("red-chair", "blue-respond").ingest().
			register("red-chair").dispatch(1, "blue-respond", "G1").
			registerAs("blue-respond", "blue-a").edit("G1", "old", "new").
			position("blue-respond").revision("blue-respond").stop("blue-a")
		_, _, err := RegisterRepair(blue(b.seed()), "")
		refused(t, err, "already carries its position and its revision")
	})
	t.Run("a seat that never opened a sitting", func(t *testing.T) {
		b, _ := owing(t)
		_, _, err := RegisterRepair(Identity{Run: b.seed(), SeatID: "blue-synthesize"}, "")
		refused(t, err, "has no sitting to repair")
	})
	t.Run("a red seat", func(t *testing.T) {
		b, _ := owing(t)
		b.registerAs(evLens, "lens-1")
		_, _, err := RegisterRepair(Identity{Run: b.seed(), SeatID: evLens}, "")
		refused(t, err, "only a blue sitting owes them")
	})
	t.Run("the synthesizer's sitting that filed its revision", func(t *testing.T) {
		b, _ := owing(t)
		b.registerAs("blue-synthesize", "synth-a").revision("blue-synthesize").stop("synth-a")
		_, _, err := RegisterRepair(Identity{Run: b.seed(), SeatID: "blue-synthesize"}, "")
		refused(t, err, "already carries its revision")
	})
	t.Run("the synthesizer's sitting that did not", func(t *testing.T) {
		b, _ := owing(t)
		b.registerAs("blue-synthesize", "synth-a").stop("synth-a")
		if _, _, err := RegisterRepair(Identity{Run: b.seed(), SeatID: "blue-synthesize"}, ""); err != nil {
			t.Errorf("a synthesizer sitting owing its revision refused its repair: %v", err)
		}
	})
	t.Run("a register naming another seat's sitting through the general write", func(t *testing.T) {
		b, opened := owing(t)
		run := b.seed()
		_, err := Append(Identity{Run: run, SeatID: "blue-synthesize"}, &recordpb.Register{RepairsSitting: proto.String(opened)})
		refused(t, err, "is not a register that opened a sitting of blue-synthesize")
	})
}

// A REPAIR REGISTER SITS FOR NO DISPATCH, whatever the writer admitted. The writer refuses a repair
// after blue is dispatched again, but the read does not lean on that refusal: here a repair lands
// after the chair's next row naming blue, and it is still no sitting for that row — blue owes the
// register, and the fold has one blue sitting, not two.
func TestARepairRegisterSitsForNoDispatch(t *testing.T) {
	b := newStage(t).cast("red-chair", "blue-respond").ingest().
		register("red-chair").dispatch(1, "blue-respond", "G1").
		registerAs("blue-respond", "blue-a")
	opened := b.lastKey()
	b.stop("blue-a").register("red-chair").dispatch(2, "blue-respond", "G1").
		repairs("blue-respond", "blue-b", opened).position("blue-respond").stop("blue-b")
	m, err := MergedEvents(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if ss := BlueSittings(m.Events, AfterTheRun); len(ss) != 1 {
		t.Errorf("blue sittings = %+v, want one: the repair opened no sitting for the second dispatch", ss)
	}
	if _, owed := owedSitting(m.Events, "blue-respond"); !owed {
		t.Error("blue is not owed its register for the second dispatch — the repair register was taken as its sitting")
	}
}
