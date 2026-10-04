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
	// THE DIRECTION IS A FIELD ON THE PLAN, not only words in its reason: the workflow quotes
	// remand_owed's direction to blue and the lens, and reads it from nowhere else.
	if want := []RemandOwed{{GapID: "G1", Direction: remandDirection}}; !slices.Equal(plan.RemandOwed, want) {
		t.Errorf("remand_owed = %+v, want %+v", plan.RemandOwed, want)
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
	if len(plan.RemandOwed) != 0 {
		t.Errorf("the remand's exchange is spent, so the plan owes it nowhere: remand_owed %+v", plan.RemandOwed)
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

// movingExchangeOn is one exchange on gap in which blue's sitting moves it: an edit answering the gap
// with old != new.
func (b *stage) movingExchangeOn(gap string) *stage {
	b.register("red-chair").dispatch(2, evLens, gap).dispatch(2, "blue-respond", gap).sitClosed(evLens)
	agent := fmt.Sprintf("blue-moves-%d", b.n)
	return b.registerAs("blue-respond", agent).edit(gap, fmt.Sprintf("old %d", b.n), fmt.Sprintf("new %d", b.n)).stop(agent)
}

// remandStageOn is the `remand_stage` the seat's work list states for an open gap.
func remandStageOn(t *testing.T, run Run, role, seat, gap string) string {
	t.Helper()
	w, err := WorkOfSeat(run, role, seat)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range w.Open {
		if g.ID == gap {
			return g.RemandStage
		}
	}
	t.Fatalf("%s is not open on %s's work list", gap, seat)
	return ""
}

// THE STAGE IS A FIELD, so a seat acting on the JSON reads the stage the plan acts on rather than
// matching it out of a sentence: owed while the exchange is owed, spent once it has begun, at_limit
// once the bench has remanded the gap at impasse twice — and the chair's row at the limit says so.
func TestTheWorkListStatesTheRemandStageAsAField(t *testing.T) {
	for _, c := range []struct {
		name, want string
		build      func(*testing.T) *stage
	}{
		{"owed", "owed", func(t *testing.T) *stage { return remandedAtImpasse(t) }},
		{"spent", "spent", func(t *testing.T) *stage { return remandedAtImpasse(t).exchangeOn("G1") }},
		{"at its limit", "at_limit", func(t *testing.T) *stage {
			return remandedAtImpasse(t).exchangeOn("G1").docketAndRemand("M2", "G1")
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := c.build(t).register("red-chair").seed()
			for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}, {"chair", "red-chair"}} {
				if got := remandStageOn(t, run, seat.role, seat.id, "G1"); got != c.want {
					t.Errorf("%s: remand_stage = %q, want %q", seat.id, got, c.want)
				}
			}
			if c.want == "at_limit" && !workMentions(t, run, "chair", "red-chair", "at its limit") {
				t.Errorf("the chair's row for a gap at its limit does not say so")
			}
		})
	}
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high").register("red-chair").seed()
	if got := remandStageOn(t, run, "blue", "blue-respond", "G1"); got != "" {
		t.Errorf("a gap the bench never remanded states remand_stage %q, want none", got)
	}
}

// A REMAND COUNTS WHERE THE BENCH FOUND THE GAP AT IMPASSE. A party's docket, filed and remanded
// while the gap was below its limits, sends the gap nowhere the debate was not already going: it
// spends nothing, so when the gap reaches impasse and the bench remands it there, that remand is
// the first, and its exchange is owed — the gap is not at its limit.
func TestARemandBeforeImpasseSpendsNothing(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.register("blue-respond").docketMotion("blue-respond", "M1", "G1").
		register("red-chair").dispatch(2, "judge", "G1").register("judge").add("judge", remandRuling("M1", "an early direction"))
	st.exchangeOn("G1").exchangeOn("G1")
	run := st.docketAndRemand("M2", "G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || plan.Ceiling {
		t.Errorf("the remand at impasse is G1's first, so its exchange is owed: parties %+v, ceiling %v, why %q", plan.Parties, plan.Ceiling, plan.Why)
	}
	if got := remandStageOn(t, run, "blue", "blue-respond", "G1"); got != "owed" {
		t.Errorf("remand_stage = %q, want owed", got)
	}
}

// ONE BENCH SITTING IS ONE REMAND, however many of the gap's docket motions it rules: two motions on
// G1 remanded in the same sitting grant one exchange, and do not put the gap at its limit.
func TestTwoMotionsRemandedInOneSittingAreOneRemand(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.exchangeOn("G1").exchangeOn("G1")
	run := st.register("red-chair").docketMotion("red-chair", "M1", "G1").
		register("blue-respond").docketMotion("blue-respond", "M2", "G1").
		register("red-chair").dispatch(2, "judge", "G1").register("judge").
		add("judge", remandRuling("M1", remandDirection)).add("judge", remandRuling("M2", remandDirection)).
		register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || plan.Ceiling {
		t.Errorf("one bench sitting remanding two motions is one remand, its exchange owed: parties %+v, ceiling %v, why %q", plan.Parties, plan.Ceiling, plan.Why)
	}
}

// THE REMAND ITEMS FOLLOW THE PLAN, NOT THE REMAND. A gap that is not material readies nobody, at
// impasse or not, so a remand of one puts nothing on blue's or the minting lens's list: an item there
// would be an exchange the dispatch never readies.
func TestARemandedTrifleOffersNoExchange(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "low")
	st.exchangeOn("G1").exchangeOn("G1")
	run := st.docketAndRemand("M1", "G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("setup: a trifle readies nobody; parties %+v, why %q", plan.Parties, plan.Why)
	}
	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		if workMentions(t, run, seat.role, seat.id, remandDirection) {
			t.Errorf("%s's work list offers the remand's exchange on a gap the plan readies nobody for", seat.id)
		}
	}
}

// THE kMax ROUTE LEAVES IMPASSE TOO, under the release gate's terms (k 1, kMax 2). G1 moves on both
// of its exchanges and reaches impasse on the count alone; the count never decreases, so a remand
// exchange that moves the gap would leave it "at impasse" forever and send a gap making progress back
// to the bench and on to CEILING. After the remand the terms count from the ruling: the moving
// exchange takes G1 off impasse, the next one that reaches kMax sends it back, and one that leaves
// it unmoved sends it back at once.
func TestAMovingRemandExchangeLeavesAKMaxImpasse(t *testing.T) {
	const terms = `{"k": 1, "kMax": 2}`
	remanded := func(t *testing.T) *stage {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
		st.movingExchangeOn("G1").movingExchangeOn("G1")
		return st.docketAndRemand("M1", "G1")
	}
	seeded := func(t *testing.T, st *stage) (Run, Plan) {
		run := st.register("red-chair").seed()
		writeRunConfig(t, run.Dir(), terms)
		plan, err := PlanDispatch(run)
		if err != nil {
			t.Fatal(err)
		}
		return run, plan
	}
	t.Run("the kMax impasse is docketed and remanded", func(t *testing.T) {
		run, plan := seeded(t, remanded(t))
		x, err := Exchanges(run, Params{K: 1, KMax: 2})
		if err != nil || x["G1"].Stalled != 0 || x["G1"].Exchanges != 2 {
			t.Fatalf("setup: G1 must reach impasse on kMax with nothing stalled: %+v (err %v)", x["G1"], err)
		}
		if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || len(plan.ToFile) != 0 {
			t.Errorf("the remand's exchange is owed: parties %+v, to file %v, why %q", plan.Parties, plan.ToFile, plan.Why)
		}
	})
	t.Run("the remand exchange moves it off impasse", func(t *testing.T) {
		run, plan := seeded(t, remanded(t).movingExchangeOn("G1"))
		if !slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || len(plan.ToFile) != 0 || plan.Ceiling {
			t.Errorf("the remand exchange moved G1, so it is below its limits counted from the ruling: parties %+v, to file %v, why %q", plan.Parties, plan.ToFile, plan.Why)
		}
		if workMentions(t, run, "chair", "red-chair", "at its limit") || workMentions(t, run, "chair", "red-chair", "for the bench again") {
			t.Errorf("the chair's row sends a gap below its limits to the bench")
		}
	})
	t.Run("kMax counted from the ruling sends it back", func(t *testing.T) {
		_, plan := seeded(t, remanded(t).movingExchangeOn("G1").movingExchangeOn("G1"))
		if !slices.Equal(plan.ToFile, []string{"G1"}) || plan.Ceiling {
			t.Errorf("two exchanges since the ruling reach kMax 2: to file %v, why %q", plan.ToFile, plan.Why)
		}
	})
	t.Run("an unmoved remand exchange sends it back at once", func(t *testing.T) {
		_, plan := seeded(t, remanded(t).exchangeOn("G1"))
		if !slices.Equal(plan.ToFile, []string{"G1"}) || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
			t.Errorf("the remand's exchange left G1 unmoved: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
		}
	})
}

// AT ITS LIMIT IS A ROUTE, NOT A COUNT: a gap the bench remanded twice that then moves is below its
// limits, and the chair's row says what the dispatch does — readies its parties — not "at its limit".
func TestAGapRemandedTwiceThatMovesIsNotAtItsLimit(t *testing.T) {
	run := remandedAtImpasse(t).exchangeOn("G1").docketAndRemand("M2", "G1").movingExchangeOn("G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || plan.Ceiling {
		t.Errorf("G1 moved after its second remand, so it is below its limits: parties %+v, ceiling %v, why %q", plan.Parties, plan.Ceiling, plan.Why)
	}
	if workMentions(t, run, "chair", "red-chair", "at its limit") {
		t.Errorf("the chair's row says G1 is at its limit while the plan readies its parties")
	}
}
