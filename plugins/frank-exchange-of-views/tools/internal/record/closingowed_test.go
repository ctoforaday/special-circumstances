package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

func (b *stage) closing(seat, gap string) *stage {
	return b.add(seat, &recordpb.Closing{GapId: proto.String(gap), Text: proto.String("the case on " + gap)})
}

// docketed is G1, minted by the evidence lens and docketed by a party's motion, with nobody having
// sat since: both parties' closings are awaited.
func docketed(t *testing.T) *stage {
	return newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "blue-lane-1", "judge").ingest().
		register("red-chair").dispatch(2, evLens).dispatch(2, outsideLens).register(evLens).register(outsideLens).
		served(evLens, 2).served(outsideLens, 2).mint(evLens, "G1", "medium").
		register("red-chair").docketMotion("blue-respond", "M1", "G1")
}

func planOfT(t *testing.T, run Run) Plan {
	t.Helper()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func closingOwners(t *testing.T, run Run) []string {
	t.Helper()
	bl, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, b := range bl {
		if b.Kind == BlockerClosingOwed {
			out = append(out, b.Owner)
		}
	}
	return out
}

// A DOCKETED GAP OWES A CLOSING FROM EACH PARTY, AND THE BENCH WAITS. The minting lens and
// blue-respond are readied with no gap engaged — the closing is not an exchange — and the bench is
// not convened while either closing is awaited. Once both are filed, the bench sits.
func TestADocketedGapOwesAClosingFromBothPartiesAndTheBenchWaits(t *testing.T) {
	plan := planOfT(t, docketed(t).seed())
	for _, seat := range []string{evLens, "blue-respond"} {
		p, ok := partyOf(plan, seat)
		if !ok || len(p.GapIDs) != 0 {
			t.Errorf("%s: want readied with no gap engaged for its closing, got %+v (named %v)", seat, p, ok)
		}
	}
	if _, ok := partyOf(plan, "judge"); ok {
		t.Errorf("the bench is convened while both closings are awaited: %+v", plan.Parties)
	}
	if _, ok := partyOf(plan, outsideLens); ok {
		t.Errorf("a lens that is no party to G1 is readied for it: %+v", plan.Parties)
	}
	if !slices.ContainsFunc(plan.Why, func(w string) bool { return strings.Contains(w, "the closing of "+evLens+" and blue-respond is owed") }) {
		t.Errorf("the plan does not say whose closings it awaits: %q", plan.Why)
	}
	if got := closingOwners(t, docketed(t).seed()); !slices.Equal(got, []string{evLens, "blue-respond"}) {
		t.Errorf("closing blockers owned by %v, want both parties", got)
	}

	t.Run("one filed, the other awaited: the bench still waits", func(t *testing.T) {
		plan := planOfT(t, docketed(t).closing("blue-respond", "G1").seed())
		if _, ok := partyOf(plan, "judge"); ok {
			t.Errorf("the bench is convened with the lens's closing awaited: %+v", plan.Parties)
		}
		if _, ok := partyOf(plan, "blue-respond"); ok {
			t.Errorf("blue-respond filed and is readied again: %+v", plan.Parties)
		}
		if _, ok := partyOf(plan, evLens); !ok {
			t.Errorf("the lens's closing is awaited and it is not readied: %+v", plan.Parties)
		}
	})
	t.Run("both filed: the bench sits", func(t *testing.T) {
		run := docketed(t).closing(evLens, "G1").closing("blue-respond", "G1").seed()
		p, ok := partyOf(planOfT(t, run), "judge")
		if !ok || !slices.Contains(p.GapIDs, "G1") {
			t.Errorf("both closings are filed and the bench is not convened on G1: %+v", p)
		}
		if got := closingOwners(t, run); len(got) != 0 {
			t.Errorf("both closings are filed and %v still owe one", got)
		}
	})
	// A CLOSING ANSWERS THE DOCKETING IT FOLLOWS: one filed before the motion answers none.
	t.Run("a closing filed before the docketing answers nothing", func(t *testing.T) {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").sit(2, evLens).mint(evLens, "G1", "medium").
			closing(evLens, "G1").closing("blue-respond", "G1").
			register("red-chair").docketMotion("blue-respond", "M1", "G1")
		if got := closingOwners(t, b.seed()); len(got) != 2 {
			t.Errorf("closings before the docketing satisfied it: owed by %v, want both", got)
		}
	})
}

// A PARTY THAT SAT AND FILED NONE IS MISSING: it is not readied again for that docketing, the bench
// sits on the record as it stands, and the item stays on the party's own list until the bench rules.
func TestAPartyThatSatWithoutItsClosingDoesNotHoldTheBench(t *testing.T) {
	run := docketed(t).closing(evLens, "G1").dispatch(2, "blue-respond").registerAs("blue-respond", "blue-a").stop("blue-a").seed()
	plan := planOfT(t, run)
	if _, ok := partyOf(plan, "blue-respond"); ok {
		t.Errorf("blue sat since the docketing and is readied again for the same closing: %+v", plan.Parties)
	}
	if p, ok := partyOf(plan, "judge"); !ok || !slices.Contains(p.GapIDs, "G1") {
		t.Errorf("no party is awaited and the bench is not convened: %+v", plan.Parties)
	}
	if !slices.ContainsFunc(plan.Why, func(w string) bool {
		return strings.Contains(w, "blue-respond sat since the docketing without filing a closing")
	}) {
		t.Errorf("the plan does not say the bench sits without blue's closing: %q", plan.Why)
	}
	s := sittingOfRunT(t, run, "blue", "blue-respond")
	if !hasItem(s, "has sat since the docketing without filing it") {
		t.Errorf("blue's list no longer names the closing it still owes: %+v", s.Open)
	}
	duties := ClosingDuties(mustEvents(t, run))
	if i := slices.IndexFunc(duties, func(d ClosingDuty) bool { return d.Party == "blue-respond" }); i < 0 || duties[i].State != ClosingMissing {
		t.Errorf("blue's closing = %+v, want missing", duties)
	}
}

// THE STATED NOT-OWED STATES: the gap closed after its docketing, and a minter the cast seats as no
// lens. Each holds no blocker and readies nobody for a closing.
func TestAClosingIsNotOwedOnAClosedGapOrByAMinterTheCastDoesNotSeat(t *testing.T) {
	t.Run("the gap closed after the docketing", func(t *testing.T) {
		run := docketed(t).closeGap(evLens, "G1").seed()
		if got := closingOwners(t, run); len(got) != 0 {
			t.Errorf("a closed gap owes closings from %v", got)
		}
		for _, d := range ClosingDuties(mustEvents(t, run)) {
			if d.State != ClosingNotOwed || d.NotOwedWhy != closingGapClosed {
				t.Errorf("%s: %+v, want not owed because the gap closed", d.Party, d)
			}
		}
	})
	t.Run("the minter is no cast lens", func(t *testing.T) {
		run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").mint("red-lens-voice", "G1", "medium").docketMotion("blue-respond", "M1", "G1").seed()
		if got := closingOwners(t, run); !slices.Equal(got, []string{"blue-respond"}) {
			t.Errorf("closings owed by %v, want blue-respond's alone", got)
		}
		duties := ClosingDuties(mustEvents(t, run))
		if len(duties) != 2 || duties[0].State != ClosingNotOwed || duties[0].NotOwedWhy != closingMinterNoLens {
			t.Errorf("duties = %+v, want the minter's not owed for want of a cast lens", duties)
		}
	})
}

// EACH PARTY'S WORK LIST BLOCKS ON ITS OWN CLOSING AND ON NO OTHER: the minting lens and blue-respond
// are blocked; another lens, a lane and the chair are not owed it (the chair's list carries it as
// another seat's). Filing clears it.
func TestTheWorkListBlocksEachPartyOnItsOwnClosing(t *testing.T) {
	const item = "is docketed for the bench (motion M1)"
	run := docketed(t).seed()
	for _, c := range []struct{ role, seat string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		if s := sittingOfRunT(t, run, c.role, c.seat); !hasItem(s, item) || s.Complete {
			t.Errorf("%s owes its closing on G1 and its list does not block on it: %+v", c.seat, s.Open)
		}
	}
	for _, c := range []struct{ role, seat string }{{"lens", outsideLens}, {"blue", "blue-lane-1"}} {
		if s := sittingOfRunT(t, run, c.role, c.seat); listsItem(s, item) {
			t.Errorf("%s is no party to G1 and its list names the closing: %+v", c.seat, s.Open)
		}
	}
	chair := sittingOfRunT(t, run, "chair", "red-chair")
	_, board := chair.Owed()
	if n := len(slices.DeleteFunc(whats(board), func(w string) bool { return !strings.Contains(w, item) })); n != 2 {
		t.Errorf("the chair's list carries %d closing item(s) as another seat's, want 2: %+v", n, chair.Open)
	}
	if s := sittingOfRunT(t, docketed(t).closing(evLens, "G1").seed(), "lens", evLens); listsItem(s, item) {
		t.Errorf("the lens filed its closing and its list still names it: %+v", s.Open)
	}
}

// A CLOSING THE DOCKETING NO LONGER OWES IS STATED, NOT DROPPED. The minting lens files its closing
// and closes the gap before blue sits (#1601 smoke, G-0354ac40): blue's list names the docketing and
// why its closing is not owed and whose close ended it, and blocks on nothing for it; the lens that
// closed the gap is told nothing it did not do itself.
func TestTheWorkListStatesAClosingTheDocketingNoLongerOwes(t *testing.T) {
	const item = "gap G1 is docketed for the bench (motion M1) and no closing argument is owed from you on it: " + closingGapClosed + " (" + evLens + " closed it)"
	s := sittingOfRunT(t, docketed(t).closing(evLens, "G1").closeGap(evLens, "G1").seed(), "blue", "blue-respond")
	if !listsItem(s, item) || hasItem(s, "is docketed for the bench (motion M1)") {
		t.Errorf("blue's list does not state, without blocking, that the closing is not owed: %+v", s.Open)
	}
	if s := sittingOfRunT(t, docketed(t).closeGap(evLens, "G1").seed(), "lens", evLens); listsItem(s, "no closing argument is owed") {
		t.Errorf("the lens closed the gap itself and is told so as news: %+v", s.Open)
	}
	if s := sittingOfRunT(t, docketed(t).seed(), "blue", "blue-respond"); listsItem(s, item) {
		t.Errorf("an owed closing is stated as not owed: %+v", s.Open)
	}
}

// THE WRITE PATH ADMITS A CLOSING FROM A PARTY ON A DOCKETED GAP, AND NOTHING ELSE: not the chair,
// which is party to no dispute; not a lens that did not mint the gap; not a lane; not on a gap no
// unruled docket motion stands on.
func TestTheWritePathAdmitsAClosingOnlyFromAPartyOnADocketedGap(t *testing.T) {
	run := docketed(t).mint(evLens, "G2", "medium").seed()
	closingOn := func(gap string) *recordpb.Closing {
		return &recordpb.Closing{GapId: proto.String(gap), Text: proto.String("the case")}
	}
	for _, seat := range []string{evLens, "blue-respond"} {
		if err := validateRun(run, seat, recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("G1")); err != nil {
			t.Errorf("%s is a party to docketed G1 and its closing is refused: %v", seat, err)
		}
	}
	for _, seat := range []string{"red-chair", outsideLens, "blue-lane-1"} {
		err := validateRun(run, seat, recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("G1"))
		if err == nil || !strings.Contains(err.Error(), "is party to none") {
			t.Errorf("%s is no party to G1: want the closing refused naming the parties, got %v", seat, err)
		}
	}
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("G2")); err == nil || !strings.Contains(err.Error(), "is not before the bench") {
		t.Errorf("G2 is not docketed: want the closing refused, got %v", err)
	}
	// AND UNDER MIGRATING ALIKE: migrate writes no event for an archived closing this refuses, so
	// the refusal holds a replayed act to the live rule rather than standing aside for it.
	migrating(t)
	if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("G1")); err == nil {
		t.Error("under Migrating the chair's closing on G1 was admitted; the write refuses it whoever replays it")
	}
	if err := validateRun(run, "blue-respond", recordpb.EventType_EVENT_TYPE_CLOSING, closingOn("G2")); err == nil {
		t.Error("under Migrating blue's closing on undocketed G2 was admitted; the write refuses it whoever replays it")
	}
}

// THE WRITE PATH ADMITS A POSITION FROM THE CHAIR AND FROM NO OTHER SEAT, whatever surface mounts
// the verb and under Migrating alike: a party argues a docketed gap in its closing.
func TestTheWritePathAdmitsAPositionOnlyFromTheChair(t *testing.T) {
	run := docketed(t).seed()
	position := &recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("why the PASS was withheld")}
	for _, replaying := range []bool{false, true} {
		if replaying {
			migrating(t)
		}
		if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_POSITION, position); err != nil {
			t.Errorf("Migrating=%v: the chair's position is refused: %v", replaying, err)
		}
		for _, seat := range []string{"blue-respond", "blue-lane-1", "blue-synthesize", "frontier", evLens, "judge"} {
			err := validateRun(run, seat, recordpb.EventType_EVENT_TYPE_POSITION, position)
			if err == nil || !strings.Contains(err.Error(), "is not the chair") {
				t.Errorf("Migrating=%v: %s is not the chair: want its position refused naming the chair, got %v", replaying, seat, err)
			}
		}
	}
}

// A LENS THAT OWES A CLOSING IS NOT RETIRED (lensOwesOf): three barren sittings retire a lens owing
// nothing, and a closing it has not filed on a docketed gap of its own is something it owes — so
// its readiness says it does not retire, and a head move readies it.
func TestALensOwingAClosingDoesNotRetire(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "medium")
	for range 3 {
		b.register("red-chair").sit(2, evLens)
	}
	b.docketMotion("blue-respond", "M1", "G1").closing("blue-respond", "G1").register("red-chair").sit(2, evLens).register("red-chair")
	run := b.seed()
	if f := foldOf(t, run, evLens); f.retired() || !slices.Contains(f.owes, "closing_owed G1") {
		t.Errorf("the gate's fold for the lens = %+v, want it unretired and owing closing_owed G1", f)
	}
	plan := planOfT(t, run)
	if !slices.ContainsFunc(plan.Why, func(w string) bool {
		return strings.HasPrefix(w, evLens+":") && strings.Contains(w, "it does not retire while it owes closing_owed G1")
	}) {
		t.Errorf("a lens owing its closing on G1 reads as retired or owing nothing: %q", plan.Why)
	}
}

// ONE DOCKETING STANDS PER GAP: two unruled docket motions on one gap owe each party one closing,
// the newest's, and the gate lists each party once.
func TestTwoUnruledDocketMotionsOweOneClosingEach(t *testing.T) {
	run := docketed(t).docketMotion(evLens, "M2", "G1").seed()
	if got := closingOwners(t, run); !slices.Equal(got, []string{evLens, "blue-respond"}) {
		t.Errorf("closing blockers owned by %v, want each party once", got)
	}
}

// mustEvents is the run's events and index.
func mustEvents(t *testing.T, run Run) ([]*Event, WindowIndex) {
	t.Helper()
	m, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	return m.Events, m.At
}
