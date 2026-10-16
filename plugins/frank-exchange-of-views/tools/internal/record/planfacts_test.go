package record

import (
	"reflect"
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE PLAN RELAYS EACH ENGAGED GAP'S EXCHANGES, AND TELLS NOT-MEASURED FROM ZERO. The engine's
// no-progress valve reads these: a gap marching to impasse keeps its head and its parties while its
// count advances. A gap whose sittings the record cannot close carries them as `unresolved` beside a
// count of 0, never as a bare 0 — the bytes an undisputed gap produces (#1002).
func TestThePlanRelaysEachEngagedGapsExchanges(t *testing.T) {
	marched := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").edit("G1", "was", "is").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").edit("G1", "is", "will be").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").
		register("red-chair")
	unclosed := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond")
	for _, c := range []struct {
		name string
		st   *stage
		want []ExchangeCount
	}{
		{"two counted, the last epoch's sittings unclosed", marched, []ExchangeCount{{GapID: "G1", Count: 2, Unresolved: 2}}},
		{"nothing counted and two sittings unclosed", unclosed, []ExchangeCount{{GapID: "G1", Count: 0, Unresolved: 2}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			plan, err := PlanDispatch(c.st.seed())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(plan.Exchanges, c.want) {
				t.Fatalf("exchanges = %+v, want %+v (parties %+v)", plan.Exchanges, c.want, plan.Parties)
			}
		})
	}
	if got := emptyPlan().Exchanges; got == nil || len(got) != 0 {
		t.Errorf("the empty plan's exchanges = %#v, want an empty array, never null", got)
	}
}

// AT THE EPOCH LIMIT THE PLAN HAS NO PARTIES, AND SO NO EXCHANGES ENTRY: an entry is a gap a party
// of this plan is engaged on.
func TestAPlanAtTheEpochLimitRelaysNoExchanges(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair")
	run := b.seed()
	r, err := planReadsOf(run)
	if err != nil {
		t.Fatal(err)
	}
	p := DefaultParams
	p.MaxEpochs = 1
	plan := foldPlan(p, r)
	if !plan.EpochLimitReached || len(plan.Parties) != 0 {
		t.Fatalf("the fixture is not at its epoch limit: %+v", plan)
	}
	if len(plan.Exchanges) != 0 {
		t.Errorf("exchanges = %+v at the epoch limit, where no party is engaged", plan.Exchanges)
	}
}

func planReadsOf(run Run) (planReads, error) {
	var r planReads
	err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = planReadsAt(q)
		return err
	})
	return r, err
}

// THE FACTS ARE THE RECORD'S AS IT STOOD AT THE POINT ASKED, never as it stands now. Capture asks
// at the end of each chair sitting that relayed a plan; an answer that read the whole record would
// hold a faithful relay of an early plan to a later sitting's rulings, counts and dockets.
func TestPlanFactsAnswerForTheRecordAsItStoodAtThePoint(t *testing.T) {
	petition := &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
		Basis: proto.String("b"), Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}}}
	closeG3 := &recordpb.Close{GapId: proto.String("G3"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").mint(evLens, "G2", "high").mint(evLens, "G3", "high").
		add(evLens, petition).add(evLens, closeG3).
		// Chair sitting 2 engages the parties on G1 and dockets G2 by hand, and G3 — closed — too.
		register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").
		docketMotion("red-chair", "M2", "G2").docketMotion("red-chair", "M3", "G3")
	first := len(b.evs) // the end of chair sitting 2
	b.register(evLens).register("blue-respond").edit("G1", "was", "is").
		register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1")
	second := len(b.evs) // the end of chair sitting 3: its docket of sitting 2 is an earlier sitting's
	b.register(evLens).register("blue-respond").
		register("judge").add("judge", &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Opinion: proto.String("no"),
		Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED}}).
		// The chair carries G3's closure forward at the end: a second closing act, after every point
		// asked below. G3 reads closed from its FIRST.
		register("red-chair").add("red-chair", &recordpb.Close{GapId: proto.String("G3"),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), CarriedFrom: proto.String("1"), Prose: proto.String("as before")})
	run := b.seed()
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	r, err := PlanFactsOf(run, evs, win)
	if err != nil {
		t.Fatal(err)
	}
	subjects := func(f PlanFacts) []string {
		var out []string
		for _, b := range f.Blockers {
			out = append(out, string(b.Kind)+" "+b.Subject+" "+b.Owner)
		}
		slices.Sort(out)
		return out
	}
	at1, at2, now := r.At(first), r.At(second), r.At(len(evs))
	if got, want := subjects(at1), []string{"unruled_motion M1 judge", "unruled_motion M2 judge", "unruled_motion M3 judge"}; !slices.Equal(got, want) {
		t.Errorf("blockers at the end of chair sitting 2 = %v, want %v", got, want)
	}
	if got := subjects(now); slices.Contains(got, "unruled_motion M1 judge") {
		t.Errorf("blockers at the end of the record = %v — the fixture's petition is ruled by then, so the point above proved nothing", got)
	}
	if got, want := at1.Docketed, []string{"G2"}; !slices.Equal(got, want) {
		t.Errorf("docketed at the end of chair sitting 2 = %v, want %v: G2's docket is this sitting's, and G3 closed before its docket was filed", got, want)
	}
	if len(at2.Docketed) != 0 {
		t.Errorf("docketed at the end of chair sitting 3 = %v, want none: the docket was an earlier sitting's", at2.Docketed)
	}
	if got, want := at1.ExchangesOn("G1"), (ExchangeCount{GapID: "G1"}); got != want {
		t.Errorf("G1's exchanges at the end of chair sitting 2 = %+v, want %+v: nobody had sat on it", got, want)
	}
	if got, want := now.ExchangesOn("G1"), (ExchangeCount{GapID: "G1", Count: 1, Unresolved: 2}); got != want {
		t.Errorf("G1's exchanges at the end of the record = %+v, want %+v", got, want)
	}
	if got, want := at1.ExchangesOn("no-such-gap"), (ExchangeCount{GapID: "no-such-gap"}); got != want {
		t.Errorf("a gap no fold names = %+v, want %+v", got, want)
	}
}

// EVERY STREAM ARM OF THE GATE'S LIST ANSWERS FOR THE STREAM IT IS HANDED. The inquiry review is the
// arm that reads an epoch: cut before the epoch that records the review, the prefix still owes one.
func TestTheStreamArmsOfTheGateAnswerForAPrefix(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("A1"), Question: proto.String("a question"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}).
		register(evLens).mint(evLens, "G1", "high")
	cut := len(b.evs)
	b.register("red-chair").add("red-chair", &recordpb.InquiryReview{Reason: proto.String("read")})
	run := b.seed()
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	due := func(evs []*Event) bool {
		return slices.ContainsFunc(eventBlockersOf(evs, win.IDs(evs), win), func(b Blocker) bool { return b.Kind == BlockerInquiryReview })
	}
	if !due(evs[:cut]) {
		t.Fatalf("the prefix before the review owes no inquiry review — the arm read past the stream it was handed")
	}
	if due(evs) {
		t.Fatalf("the whole record still owes an inquiry review — the fixture's review does not answer it, so the prefix proved nothing")
	}
}

// THE RELAY-HELD KINDS ARE THE STREAM ARMS LESS THE CHAIR'S OWN REVIEW, and the gate's order is the
// table's whichever arm produced a blocker.
func TestTheRelayHeldKindsAreStatedOnTheKindTable(t *testing.T) {
	want := []BlockerKind{BlockerContradiction, BlockerReopened, BlockerNoReport, BlockerUnruledMotion}
	if got := RelayBlockerKinds(); !slices.Equal(got, want) {
		t.Fatalf("relay-held kinds = %v, want %v", got, want)
	}
	if got := BlockerKinds(); len(got) != len(blockerKinds) {
		t.Fatalf("BlockerKinds lists %d kinds, the table has %d", len(got), len(blockerKinds))
	}
	// No report, and a motion: the stream arms alone, and RelayBlockersOf in table order.
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").
		register(evLens).add(evLens, &recordpb.Motion{MotionId: proto.String("M1"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_PETITION.Enum(),
		Basis: proto.String("b"), Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}}})
	run := b.seed()
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	got := RelayBlockersOf(evs, win)
	want2 := []PlanBlocker{{Kind: BlockerNoReport}, {Kind: BlockerUnruledMotion, Subject: "M1", Owner: "judge"}}
	if !reflect.DeepEqual(got, want2) {
		t.Fatalf("relay blockers = %+v, want %+v", got, want2)
	}
}
