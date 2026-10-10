package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// hasItem is a BLOCKING item carrying sub. It is named for what it asks rather than for `Open`,
// because the list holds both kinds: an item may be present without blocking (Item.Blocks), which
// is how a lens's audit is stated — work it owes that must not hold its turn open.
func hasItem(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

// listsItem is any item carrying sub, blocking or not.
func listsItem(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

// listed is hasItem for an AFFORDANCE — availableOf's lines carry Blocks:false, open work rather
// than owed work, so a blocking-only match can never see them.
func listed(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

// repairs writes a register that names key as the sitting it repairs. NO COMMAND WRITES ONE: the
// stored field stays readable, so the readers that bound a sitting by it keep a fixture.
func (b *stage) repairs(seat, agent, key string) *stage {
	return b.add(seat, &recordpb.Register{AgentId: proto.String(agent), RepairsSitting: proto.String(key)})
}

func (b *stage) position(seat string) *stage {
	return b.add(seat, &recordpb.Position{Text: proto.String("the position")})
}

// lastKey is the key of the event the stage wrote last.
func (b *stage) lastKey() string { return b.evs[len(b.evs)-1].GetKey() }

func (b *stage) closeGap(lens, gap string) *stage {
	return b.add(lens, &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")})
}

// THE LOG IS OFFERED EVERY SITTING THAT DID ANYTHING, AND BLOCKS NONE. The item read the whole
// record, so a seat's first log entry closed the channel for every later sitting — the first half.
// The second half: a sitting that acted and hit nothing is COMPLETE without an entry. While the item
// blocked, seats padded the log to reach `complete` — audit summaries and arguments about other
// lenses' gaps on universe-m12 — in the channel meant only for what got in their way.
//
// THE SECOND SITTING HERE MINTS, and that is load-bearing rather than scene-setting. A sitting with
// no acts at all is not offered the channel (#1089), so a second sitting that recorded NOTHING would
// pass the first half while asserting the other rule. The mint makes it a sitting that acted.
func TestTheLogIsOfferedEverySittingAndBlocksNone(t *testing.T) {
	first := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).logEntry(evLens)
	}
	if listsItem(sittingOfRunT(t, first().seed(), "lens", evLens), "the log is open") {
		t.Fatal("a lens that logged this sitting is told its log is open")
	}
	again := first().register("red-chair").dispatch(2, evLens).register(evLens).add(evLens, &recordpb.Finding{Id: proto.String("F-f0000001"), Text: proto.String("something this sitting actually did")})
	s := sittingOfRunT(t, again.seed(), "lens", evLens)
	if !listsItem(s, "the log is open") {
		t.Fatal("a lens in its second sitting, which acted and filed no log, is not told the log is open")
	}
	if !s.Complete {
		t.Errorf("a sitting that acted and hit nothing cannot finish without a log entry: %+v", s.Open)
	}
}

// AND A SITTING THAT DID NOTHING IS OFFERED NOTHING. This is the other half of the rule above, and the
// whole point of #1089: a woken seat with no work should be able to end its turn having run no
// commands at all. Its sitting is on the record either way — the hooks write sitting_open and
// sitting_close around it, each carrying the agent's id and type.
func TestASittingThatRecordedNothingOwesNoLog(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens)
	s := sittingOfRunT(t, st.seed(), "lens", evLens)
	if listsItem(s, "the log is open") {
		t.Error("a sitting that recorded nothing is told to file a log saying it recorded nothing")
	}
	if !s.Complete {
		t.Errorf("a sitting that recorded nothing is not complete, so the seat cannot end its turn: %+v", s.Open)
	}
}

// NO BLUE SEAT IS TOLD A REVISION IS MISSING, because none owes one. The work list is where a seat
// reads what its sitting still owes, so an item there for an act no surface offers is an
// instruction the tool then refuses. Every blue seat sits here having filed nothing at all; the
// one sitting-record item any of them reads is blue-respond's position, which it owes because the
// gap it was engaged on is open.
func TestNoBlueSeatIsToldARevisionIsMissing(t *testing.T) {
	for _, seat := range []string{"blue-lane-1", "frontier", "blue-synthesize", "blue-respond"} {
		t.Run(seat, func(t *testing.T) {
			b := newStage(t).cast(evLens, "red-chair", "blue-lane-1", "frontier", "blue-synthesize", "blue-respond", "judge").ingest().
				register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
				register("red-chair").dispatch(2, "blue-respond", "G1").register(seat)
			s := sittingOfRunT(t, b.seed(), "blue", seat)
			for _, it := range s.Open {
				if strings.Contains(strings.ToLower(it.What), "revision") {
					t.Errorf("%s's work list names a revision, an act no surface offers: %q", seat, it.What)
				}
			}
			if owes := seat == "blue-respond"; hasItem(s, "this sitting's position is missing") != owes {
				t.Errorf("%s's work list holds a position item = %v, want %v: %+v", seat, !owes, owes, s.Open)
			}
			if seat != "blue-respond" && !s.Complete {
				t.Errorf("%s sat, owes nothing, and its work list is not complete: %+v", seat, s.Open)
			}
		})
	}
}

// THE CHAIR RE-SAMPLES THE ARCHIVE EVERY SITTING it is not empty — its item already said "this
// sitting has sampled none of it", and read the whole record.
func TestTheSpotCheckIsOwedEverySittingTheArchiveHoldsAClosure(t *testing.T) {
	sampled := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").closeGap(evLens, "G1").
			register("red-chair").add("red-chair", &recordpb.SpotCheck{Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")})
	}
	const unsampled = "this sitting has sampled none of it"
	if listed(sittingOfRunT(t, sampled().seed(), "chair", "red-chair"), unsampled) {
		t.Fatal("the chair sampled the archive this sitting and is told it has not")
	}
	if !listed(sittingOfRunT(t, sampled().register("red-chair").seed(), "chair", "red-chair"), unsampled) {
		t.Fatal("the chair's next sitting has sampled nothing, and the work list does not say so")
	}
}

// WHETHER A BLUE SITTING OWES AN ANSWER IS FIXED AT ITS REGISTER. The latest sitting is unresolved
// exactly while blue sits it — the moment its work list is read — so what it owes is its Open set
// and never where the sitting ended: an unresolved sitting owes what it found open, as a closed one does.
func TestWhatABlueSittingOwesIsWhatItFoundOpenWhetherOrNotItHasEnded(t *testing.T) {
	cases := map[string]struct {
		evs              []*Event
		unresolved, owes bool
	}{
		"in flight, G1 open": {
			evs:        []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a")},
			unresolved: true, owes: true,
		},
		"in flight, a later chair row naming blue changes nothing": {
			evs:        []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), chairDispatches(t, "blue-respond", "G1")},
			unresolved: true, owes: true,
		},
		"returned, G1 open": {
			evs:  []*Event{dispatchBlue(t, "G1"), registersAs(t, "blue-respond", "blue-a"), agentStops(t, "blue-a")},
			owes: true,
		},
		"in flight, G1 closed before blue sat": {
			evs:        []*Event{dispatchBlue(t, "G1"), closeGap(t, "red-lens-logic", "G1"), registersAs(t, "blue-respond", "blue-a")},
			unresolved: true,
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			m := loadedT(t, append(gapsExist(t, "G1"), c.evs...)...)
			ss := BlueSittings(m.Events, m.At, WhileRunning)
			if len(ss) != 1 || ss[0].Unresolved != c.unresolved {
				t.Fatalf("sittings = %+v, want one with Unresolved=%v", ss, c.unresolved)
			}
			if got := len(ss[0].Open) > 0; got != c.owes {
				t.Errorf("the sitting owes an answer = %v, want %v", got, c.owes)
			}
			p, sitting := positionSittingNow(m.Events, m.At, "blue-respond")
			if !sitting || (p.State != PositionNotOwed) != c.owes {
				t.Errorf("the work list's reading of the sitting = %q (found %v), want it owing a position = %v", p.State, sitting, c.owes)
			}
		})
	}
}

// A REPAIR DOES NOT RE-OPEN WHAT THE SITTING IT REPAIRS ALREADY DID (#1026). seatDidThisSitting is
// a reader that ATTRIBUTES — it answers "did this sitting file its log", not "has this turn" — so
// its window is the sitting the record stored the acts in, from its OPENING (a hook's bracket or a
// register), never a repair's own register. Starting it at the seat's latest register of any kind
// would tell the seat that the log channel is open for a sitting that has already filed there, and
// the work list would name something the sitting has done. No command writes such a register; the
// record can hold one, and this reader answers for it.
//
// WHAT THE SITTING STILL OWES IS STILL OWED, which is the half that must not move with it: a window
// that swallowed the missing position would report the sitting complete.
func TestARepairDoesNotReopenTheDutiesTheSittingDischarged(t *testing.T) {
	sat := func(t *testing.T) (*stage, string) {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
			register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-a")
		opened := b.lastKey()
		return b.logEntry("blue-respond").stop("blue-a"), opened
	}

	b, opened := sat(t)
	s := sittingOfRunT(t, b.repairs("blue-respond", "blue-b", opened).seed(), "blue", "blue-respond")
	if listsItem(s, "the log is open") {
		t.Error("inside the repair the work list asks for a log the sitting it repairs already filed")
	}
	if !hasItem(s, "this sitting's position is missing") {
		t.Error("inside the repair the work list does not name the position the repaired sitting still owes")
	}

	// THE CONTROL: a plain register is a real second sitting, and every per-sitting duty is owed
	// again — for a sitting that DID something. An empty second sitting is offered no log (#1089), so the
	// act here is what makes this a control on the repair window rather than on that rule.
	again, _ := sat(t)
	second := again.register("blue-respond").
		add("blue-respond", &recordpb.Position{Text: proto.String("this sitting took a position")}).seed()
	if !listsItem(sittingOfRunT(t, second, "blue", "blue-respond"), "the log is open") {
		t.Error("a genuine second sitting acted, filed no log, and the work list does not say so")
	}
}

// A REFUSAL THE TOOL LOGGED IS NOT THE SEAT SPEAKING, AND IT IS FRICTION. The tool records every
// refusal it gives a seat; if that entry counted as the seat's log, the first refused call would
// close the channel it exists to open. And a sitting whose only events are refused calls recorded
// no act, yet plainly met friction, so it is asked — with the count, so the seat knows the tool
// already holds the refusals and only its expectation is missing.
func TestARefusalTheToolLoggedAsksTheSeatForWhatItExpected(t *testing.T) {
	refusal := func(b *stage) *stage {
		return b.add(evLens, &recordpb.Log{Text: proto.String("refused `mint` with --acceptance-check: unknown flag"),
			Type:   recordpb.LogType_LOG_TYPE_REFUSAL.Enum(),
			Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum()})
	}
	sat := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens)
	}
	s := sittingOfRunT(t, refusal(sat()).seed(), "lens", evLens)
	if !listsItem(s, "has already recorded 1 refused or failed call") {
		t.Fatalf("a sitting whose only events are refused calls is not asked what it expected: %+v", s.Open)
	}
	if !s.Complete {
		t.Errorf("the ask blocks the sitting: %+v", s.Open)
	}
	if listsItem(sittingOfRunT(t, refusal(sat()).logEntry(evLens).seed(), "lens", evLens), "the log is open") {
		t.Error("the seat filed its own entry and is still told the log is open")
	}
}
