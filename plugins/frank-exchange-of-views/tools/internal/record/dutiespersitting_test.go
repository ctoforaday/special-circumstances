package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func hasItem(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, sub) {
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

func (b *stage) revision(seat string) *stage {
	return b.add(seat, &recordpb.Revision{Text: proto.String("the sitting's edits")})
}

func (b *stage) closeGap(lens, gap string) *stage {
	return b.add(lens, &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")})
}

// THE LOG IS OWED EVERY SITTING. It read the whole record, so a seat's first log entry discharged
// every later sitting's.
func TestTheLogIsOwedEverySitting(t *testing.T) {
	first := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).logNominal(evLens)
	}
	if hasItem(sittingOfRunT(t, first().seed(), "lens", evLens), "the log is open") {
		t.Fatal("a lens that logged this sitting is told its log is open")
	}
	again := first().register("red-chair").dispatch(2, evLens).register(evLens)
	if !hasItem(sittingOfRunT(t, again.seed(), "lens", evLens), "the log is open") {
		t.Fatal("a lens in its second sitting, with only its first sitting's log on the record, is not told the log is open")
	}
}

// BLUE-RESPOND OWES A REVISION FOR EACH SITTING THAT HAD SOMETHING TO ANSWER, and none for a
// sitting that found every engaged gap closed before it sat (gblock's ruling, 2026-09-11). In B9
// blue sat in epoch 5 with G4 open, filed no revision, and read complete because its epoch-4
// revision was on the record.
func TestBlueRespondsRevisionIsOwedPerSittingThatHadAGapOpen(t *testing.T) {
	sat := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
			register("red-chair").dispatch(2, "blue-respond", "G1").register("blue-respond").revision("blue-respond")
	}
	const missing = "this sitting's revision is missing"
	if hasItem(sittingOfRunT(t, sat().seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue revised this sitting and is told its revision is missing")
	}
	again := sat().register("red-chair").dispatch(2, "blue-respond", "G1").register("blue-respond")
	if !hasItem(sittingOfRunT(t, again.seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue's second sitting on an open G1 has no revision, and the work list does not say so")
	}
	closedFirst := sat().register("red-chair").dispatch(2, "blue-respond", "G1").closeGap(evLens, "G1").register("blue-respond")
	if hasItem(sittingOfRunT(t, closedFirst.seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue's sitting found G1 closed before it sat and is still told a revision is owed")
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

// Every other blue seat sits once for its act, and owes that sitting's revision.
func TestTheSynthesizerOwesItsSittingsRevision(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "blue-synthesize", "judge").register("blue-synthesize")
	if !hasItem(sittingOfRunT(t, b.seed(), "blue", "blue-synthesize"), "this sitting's revision is missing") {
		t.Fatal("the synthesizer filed no revision and is not told it is missing")
	}
}

// WHETHER A BLUE SITTING OWES A REVISION IS FIXED AT ITS REGISTER. The latest sitting is unresolved
// exactly while blue sits it — the moment its work list is read — so revisionOwed reads the Open set
// and never where the sitting ended: an unresolved sitting owes what it found open, as a closed one does.
func TestRevisionOwedReadsWhatTheSittingFoundOpenWhetherOrNotItHasEnded(t *testing.T) {
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
			ss := BlueSittings(c.evs, WhileRunning)
			if len(ss) != 1 || ss[0].Unresolved != c.unresolved {
				t.Fatalf("sittings = %+v, want one with Unresolved=%v", ss, c.unresolved)
			}
			if got := revisionOwed(c.evs, "blue-respond"); got != c.owes {
				t.Errorf("revisionOwed = %v, want %v", got, c.owes)
			}
		})
	}
}

// A REPAIR DOES NOT RE-OPEN A DUTY THE SITTING IT REPAIRS DISCHARGED (#1026). seatDidThisSitting is
// a reader that ATTRIBUTES — it answers "did this sitting file its log", not "has this turn" — so
// its window opens at the register that OPENED the sitting, never at a repair's own. Starting it at
// the seat's latest register of any kind told a re-prompted seat that the log channel was open
// while scorecard.channel_closure, reading the same attribution, scored that duty discharged: the
// prompt says "put it on the record NOW — nothing else" and the work list then named something the
// sitting did not owe.
//
// THE REVISION IS STILL OWED, which is the half that must not move with it: the repair exists
// because that duty is outstanding, and a window that swallowed it would report the sitting complete.
func TestARepairDoesNotReopenTheDutiesTheSittingDischarged(t *testing.T) {
	sat := func(t *testing.T) (*stage, string) {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
			register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-a")
		opened := b.lastKey()
		return b.logNominal("blue-respond").position("blue-respond").stop("blue-a"), opened
	}

	b, opened := sat(t)
	s := sittingOfRunT(t, b.repairs("blue-respond", "blue-b", opened).seed(), "blue", "blue-respond")
	if hasItem(s, "the log is open") {
		t.Error("inside the repair the work list asks for a log the sitting it repairs already filed")
	}
	if !hasItem(s, "this sitting's revision is missing") {
		t.Error("inside the repair the work list does not name the revision the repair exists to file")
	}

	// THE CONTROL: a plain register is a real second sitting, and every per-sitting duty is owed again.
	again, _ := sat(t)
	if !hasItem(sittingOfRunT(t, again.register("blue-respond").seed(), "blue", "blue-respond"), "the log is open") {
		t.Error("a genuine second sitting has filed no log, and the work list does not say so")
	}
}
