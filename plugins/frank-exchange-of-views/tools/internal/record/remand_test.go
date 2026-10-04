package record

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// remandDirection is the research direction the bench's remand states, as the ruling's reopens_on.
const remandDirection = "re-derive the figure from the primary table rather than the press summary"

// remandRuling is the bench remanding a docket motion with direction as its stated research direction.
func remandRuling(motion, direction string) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String(motion), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
		Opinion: proto.String("the record cannot settle this yet"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED), Principle: proto.String("p"), Tension: proto.String("t"),
			ReviewFlag: proto.String("none"), Settled: proto.String("nothing"), ReopensOn: proto.String(direction)}}}
}

// sitClosed is one complete sitting of seat: its register, and the stop the SubagentStop hook
// records when its agent returns, so the exchange fold can close it without the seat sitting again.
func (b *stage) sitClosed(seat string) *stage {
	agent := fmt.Sprintf("%s-agent-%d", seat, b.n+1)
	return b.registerAs(seat, agent).stop(agent)
}

// exchangeOn is one exchange on gap: the chair engages the minting lens and blue, and each sits
// once and returns, recording nothing on the gap — a stalled exchange.
func (b *stage) exchangeOn(gap string) *stage {
	return b.register("red-chair").dispatch(2, evLens, gap).dispatch(2, "blue-respond", gap).
		sitClosed(evLens).sitClosed("blue-respond")
}

// docketAndRemand is the chair's verb docketing gap at impasse as motion, and the bench sitting for
// it and remanding it with remandDirection.
func (b *stage) docketAndRemand(motion, gap string) *stage {
	return b.register("red-chair").docketMotion("red-chair", motion, gap).
		dispatch(2, "judge", gap).register("judge").add("judge", remandRuling(motion, remandDirection))
}

// remandedAtImpasse is G1 stalled to impasse under the default terms (k 2), docketed by the chair,
// and remanded by the bench with a stated direction.
func remandedAtImpasse(t *testing.T) *stage {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.exchangeOn("G1").exchangeOn("G1")
	return st.docketAndRemand("M1", "G1")
}

// gapsOf is the gaps the plan engages seat on, and whether it readies the seat at all.
func gapsOf(plan Plan, seat string) []string {
	p, _ := partyOf(plan, seat)
	return p.GapIDs
}

func whyMentions(plan Plan, frag string) bool {
	return slices.ContainsFunc(plan.Why, func(w string) bool { return strings.Contains(w, frag) })
}

func workMentions(t *testing.T, run Run, role, seat, frag string) bool {
	t.Helper()
	w, err := WorkOfSeat(run, role, seat)
	if err != nil {
		t.Fatal(err)
	}
	return slices.ContainsFunc(w.Sitting.Open, func(it Item) bool { return strings.Contains(it.What, frag) })
}

// A BENCH REMAND SENDS THE GAP BACK TO THE DEBATE FOR ONE MORE EXCHANGE (#1210, gblock's ruling of
// 2026-09-29: the code follows the prose). The gap is at impasse and the bench has ruled it, and the
// next plan readies its minting lens and blue on it, with the ruling's direction stated in the plan
// and on both seats' work lists — so the direction the bench wrote reaches the two seats that owe it.
// While that exchange is owed the board is not at its ceiling, and every PASS blocker on the gap is
// owned by a seat the plan readies.
func TestARemandReadiesTheMinterAndBlueForOneMoreExchange(t *testing.T) {
	run := remandedAtImpasse(t).register("red-chair").seed()
	x, err := Exchanges(run, DefaultParams)
	if err != nil || !x["G1"].Impasse {
		t.Fatalf("setup: G1 must be at impasse before the remand (err %v, exchanges %+v)", err, x["G1"])
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("a remanded gap readies its minting lens and blue on it; parties %+v, why %q", plan.Parties, plan.Why)
	}
	if _, bench := partyOf(plan, benchSeat); bench || len(plan.ToFile) != 0 {
		t.Errorf("the remand's exchange is owed, so nothing goes back to the bench yet: parties %+v, to file %v", plan.Parties, plan.ToFile)
	}
	if !whyMentions(plan, remandDirection) {
		t.Errorf("the plan's reason for G1 must carry the ruling's direction %q; why %q", remandDirection, plan.Why)
	}
	if plan.Ceiling {
		t.Errorf("a remanded gap whose exchange is owed is not at its limit, so the board is not at its ceiling")
	}
	if v, _, ok, _ := DeriveVerdict(run); ok && v == "CEILING" {
		t.Errorf("CEILING derived while the remand's exchange is owed")
	}
	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		if !workMentions(t, run, seat.role, seat.id, remandDirection) {
			t.Errorf("%s's work list does not carry the remand's direction %q", seat.id, remandDirection)
		}
	}
	readied := map[string]bool{}
	for _, p := range plan.Parties {
		readied[p.SeatID] = true
	}
	for _, b := range plan.Blockers {
		if b.Subject == "G1" && !readied[b.Owner] {
			t.Errorf("blocker %s on G1 is owned by %q, whom the plan does not ready: parties %+v", b.Kind, b.Owner, plan.Parties)
		}
	}
}

// ONE EXCHANGE, NOT A LOOP. The minting lens and blue sit once on the remanded gap and nothing
// moves: the gap is at impasse again, and the plan dockets it for the bench afresh rather than
// readying the parties a second time. The work lists stop carrying the direction once its exchange
// is spent.
func TestAStalledRemandExchangeGoesBackToTheBench(t *testing.T) {
	run := remandedAtImpasse(t).exchangeOn("G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Errorf("the remand's one exchange is spent; G1 must not ready its parties again: parties %+v", plan.Parties)
	}
	if !slices.Equal(plan.ToFile, []string{"G1"}) || !slices.Contains(plan.Docket, "G1") || !slices.Contains(gapsOf(plan, benchSeat), "G1") {
		t.Errorf("a remanded gap still at impasse after its exchange is docketed for the bench again: to file %v, docket %v, parties %+v, why %q",
			plan.ToFile, plan.Docket, plan.Parties, plan.Why)
	}
	if plan.Ceiling {
		t.Errorf("a gap going back to the bench is not at its limit")
	}
	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		if workMentions(t, run, seat.role, seat.id, remandDirection) {
			t.Errorf("%s's work list still offers the remand's exchange after it was spent", seat.id)
		}
	}
}

// AN EXCHANGE THAT MOVES THE GAP takes it off impasse: it is below its limits again and debated as
// any other gap is, its parties readied by the ordinary route.
func TestARemandExchangeThatMovesTheGapLeavesItBelowItsLimits(t *testing.T) {
	run := remandedAtImpasse(t).register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").
		sitClosed(evLens).registerAs("blue-respond", "blue-moves").edit("G1", "old", "new").stop("blue-moves").
		register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !whyMentions(plan, "G1: open, material") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || len(plan.ToFile) != 0 {
		t.Errorf("a remand exchange that moved G1 leaves it below its limits: parties %+v, to file %v, why %q", plan.Parties, plan.ToFile, plan.Why)
	}
}

// THE CEILING IS A GAP THE BENCH REMANDED AFTER ITS REMAND EXCHANGE: remanded, one more exchange,
// still at impasse, docketed again, and remanded again. Then nobody is ready and the run ends CEILING.
func TestCeilingIsDerivedFromEveryMaterialGapAtItsLimit(t *testing.T) {
	run := remandedAtImpasse(t).exchangeOn("G1").docketAndRemand("M2", "G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parties) != 0 || !plan.Ceiling || plan.PassPermitted || len(plan.ToFile) != 0 {
		t.Fatalf("plan = %+v, want nobody ready, nothing to file, and ceiling", plan)
	}
	got, why, ok, err := DeriveVerdict(run)
	if err != nil || !ok || got != "CEILING" {
		t.Errorf("got %q (ok=%v, err=%v) — want CEILING: %s", got, ok, err, why)
	}
}
