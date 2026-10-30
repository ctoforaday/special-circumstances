package record

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// remandDirection is the research direction the bench's remand states.
const remandDirection = "re-derive the figure from the primary table rather than the press summary"

// remandRuling is the bench remanding a docket motion with direction as its stated research direction.
func remandRuling(motion, direction string) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String(motion),
		Opinion: proto.String("the record cannot settle this yet"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED), Principle: proto.String("p"), Tension: proto.String("t"),
			ReviewFlag: proto.String("none"), Settled: proto.String("nothing"), Direction: proto.String(direction)}}}
}

// sitClosed is one complete sitting of seat: its register, the whole report served to it where it is a
// lens, and the stop the SubagentStop hook records when its agent returns, so the exchange fold can
// close it without the seat sitting again.
func (b *stage) sitClosed(seat string) *stage {
	agent := fmt.Sprintf("%s-agent-%d", seat, b.n+1)
	b.registerAs(seat, agent)
	if seatclass.RoleOfSeat(seat) == "lens" {
		b.served(seat, b.head())
	}
	return b.stop(agent)
}

// exchangeOn is one exchange on gap: the chair engages the minting lens and blue, and each sits
// once and returns, recording nothing on the gap — a stalled exchange.
func (b *stage) exchangeOn(gap string) *stage {
	return b.register("red-chair").dispatch(2, evLens, gap).dispatch(2, "blue-respond", gap).
		sitClosed(evLens).sitClosed("blue-respond")
}

// actingExchangeOn is one exchange on gap in which both parties act on it without moving it: the
// minting lens keeps the grade, and blue records a proof answering it. It is the remand's exchange
// when one is owed, because blue acted in it.
func (b *stage) actingExchangeOn(gap string) *stage {
	b.lensActs(gap)
	agent := fmt.Sprintf("blue-proves-%d", b.n)
	return b.registerAs("blue-respond", agent).blueProves(gap).stop(agent)
}

// lensOnlyExchangeOn is one exchange on gap in which only the minting lens acts: it sits first and
// keeps the grade, and blue sits after it and records nothing on the gap — universe-m35's order. It
// begins no remand's exchange.
func (b *stage) lensOnlyExchangeOn(gap string) *stage {
	return b.lensActs(gap).sitClosed("blue-respond")
}

// lensActs is the chair dispatching the minting lens and blue on gap, and the lens sitting first and
// keeping the grade.
func (b *stage) lensActs(gap string) *stage {
	b.register("red-chair").dispatch(2, evLens, gap).dispatch(2, "blue-respond", gap)
	agent := fmt.Sprintf("%s-keeps-%d", evLens, b.n)
	return b.registerAs(evLens, agent).served(evLens, b.head()).keepGrade(evLens, gap, keptReason).stop(agent)
}

// blueProves is blue recording a proof answering gap: an act on it that moves nothing.
func (b *stage) blueProves(gap string) *stage {
	return b.add("blue-respond", &recordpb.Proof{ProofId: proto.String(fmt.Sprintf("P-%d", b.n+1)),
		ProofSha: proto.String(fmt.Sprintf("sha-%d", b.n+1)), Answers: proto.String(gap)})
}

// docketAndRemand is the chair's verb docketing gap at impasse as motion, and the bench sitting for
// it and remanding it with remandDirection.
func (b *stage) docketAndRemand(motion, gap string) *stage {
	return b.docketAndRemandWith(motion, gap, remandDirection)
}

// docketAndRemandWith is docketAndRemand with the bench stating direction.
func (b *stage) docketAndRemandWith(motion, gap, direction string) *stage {
	return b.register("red-chair").docketMotion("red-chair", motion, gap).
		dispatch(2, "judge", gap).register("judge").add("judge", remandRuling(motion, direction))
}

// remandedAtImpasse is G1 stalled to impasse under the default terms (k 2), docketed by the chair,
// and remanded by the bench with a stated direction.
func remandedAtImpasse(t *testing.T) *stage {
	return remandedAtImpasseWith(t, remandDirection)
}

// remandedAtImpasseWith is remandedAtImpasse with the bench stating direction.
func remandedAtImpasseWith(t *testing.T, direction string) *stage {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.exchangeOn("G1").exchangeOn("G1")
	return st.docketAndRemandWith("M1", "G1", direction)
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
// 2026-09-29: the code follows the prose), AND THE EXCHANGE BEGINS WITH BLUE (ruled by gblock
// 2026-10-27 and 2026-10-30). The gap is at impasse and the bench has ruled it, and the next plan
// readies blue on it — not its minting lens, whose turn comes with blue's act — with the ruling's
// direction stated in the plan and on both seats' work lists, so the direction the bench wrote
// reaches both parties. While that exchange is owed the board is not at its ceiling. Once blue has
// answered, the plan readies the lens on the gap to audit the answer.
func TestARemandReadiesBlueForOneMoreExchange(t *testing.T) {
	run := remandedAtImpasse(t).register("red-chair").seed()
	x, err := Exchanges(run, DefaultParams)
	if err != nil || !x["G1"].Impasse {
		t.Fatalf("setup: G1 must be at impasse before the remand (err %v, exchanges %+v)", err, x["G1"])
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("a remanded gap readies blue on it, and not its minting lens before blue has acted; parties %+v, why %q", plan.Parties, plan.Why)
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
		t.Errorf("a remanded gap whose exchange is owed readies its parties, so the board is not at its ceiling")
	}
	if v, _, ok, _ := DeriveVerdict(run); ok && v == "CEILING" {
		t.Errorf("CEILING derived while the remand's exchange is owed")
	}
	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		if !workMentions(t, run, seat.role, seat.id, remandDirection) {
			t.Errorf("%s's work list does not carry the remand's direction %q", seat.id, remandDirection)
		}
	}
	if remandItemBlocks(t, run, "lens", evLens) {
		t.Errorf("the lens's remand item blocks its list before blue has acted on G1")
	}
	answered, err := PlanDispatch(remandedAtImpasse(t).register("red-chair").dispatch(2, "blue-respond", "G1").
		registerAs("blue-respond", "blue-answers-remand").edit("G1", "old", "new").stop("blue-answers-remand").
		register("red-chair").seed())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(answered, evLens), "G1") || len(answered.RemandOwed) != 0 {
		t.Errorf("blue's sitting answering the remand is its exchange, and the lens's turn on G1 has come: parties %+v, remand_owed %+v, why %q", answered.Parties, answered.RemandOwed, answered.Why)
	}
}

// A REMAND LINE PRINTS THE BENCH'S DIRECTION LAST, AFTER RemandDirectionPrefix, AS WRITTEN. migrate
// carries everything after the prefix in an archived plan's line as seat prose — neither translating
// a number in it nor reading a quote in it — so the verb printing the direction anywhere but last,
// or quoted, would hand migrate a tail that is not the bench's words. The direction holds an id
// phrase and an unclosed quote, the two things the tail rule exists to leave alone.
func TestARemandLineCarriesItsDirectionAsWritten(t *testing.T) {
	const direction = `re-read pin 412 and the claim "head 412`
	run := remandedAtImpasseWith(t, direction).register("red-chair").seed()
	plan, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := storedOf(t, run)
	if len(stored) != 1 || stored[0].Printed != plan.Printed() {
		t.Fatalf("stored %d plan(s); want the one the verb printed", len(stored))
	}
	got := stored[0].Plan
	var line string
	for _, w := range got.Why {
		if strings.HasPrefix(w, "G1: ") {
			line = w
		}
	}
	if !strings.HasSuffix(line, RemandDirectionPrefix+direction) || strings.Count(line, RemandDirectionPrefix) != 1 {
		t.Errorf("the stored plan's line for G1 must end with the prefix and the direction as written, once; got %q", line)
	}
	if want := []RemandOwed{{GapID: "G1", Direction: direction}}; !slices.Equal(got.RemandOwed, want) {
		t.Errorf("remand_owed = %+v, want %+v", got.RemandOwed, want)
	}
}

// ONE EXCHANGE, NOT A LOOP. The minting lens and blue sit once on the remanded gap, both act on it,
// and nothing moves: the gap is at impasse again, and the plan dockets it for the bench afresh
// rather than readying the parties a second time. The work lists stop carrying the direction once its
// exchange is spent.
func TestAStalledRemandExchangeGoesBackToTheBench(t *testing.T) {
	run := remandedAtImpasse(t).actingExchangeOn("G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Errorf("the remand's one exchange is spent; G1 must not ready its parties again: parties %+v", plan.Parties)
	}
	// DOCKETED AGAIN, AND ARGUED AGAIN: its parties are readied for their closings on the new
	// docketing, engaged on no gap, and the bench sits once neither is awaited.
	if !slices.Equal(plan.ToFile, []string{"G1"}) || !slices.Contains(plan.Docket, "G1") || slices.Contains(gapsOf(plan, benchSeat), "G1") ||
		!slices.ContainsFunc(plan.Why, func(w string) bool {
			return strings.HasPrefix(w, "G1: docketed — the closing of "+evLens+" and blue-respond")
		}) {
		t.Errorf("a remanded gap still at impasse after its exchange is docketed for the bench again, its parties readied for their closings: to file %v, docket %v, parties %+v, why %q",
			plan.ToFile, plan.Docket, plan.Parties, plan.Why)
	}
	if plan.Ceiling {
		t.Errorf("a gap going back to the bench is no ceiling")
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

// A SITTING THAT DOES NOT TOUCH THE GAP SPENDS NOTHING (gblock, 2026-10-25). The minting lens and blue
// sit once after the ruling and neither records an act on G1: the remand's exchange is still owed, so
// the plan readies blue on it again with the direction — and not the lens, whose turn comes with
// blue's act — nothing goes back to the bench, the stage stays owed and the PASS is held. The next
// exchange, in which blue acts, is the remand's.
func TestASittingThatDoesNotTouchTheGapSpendsNothing(t *testing.T) {
	run := remandedAtImpasse(t).exchangeOn("G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || len(plan.ToFile) != 0 {
		t.Errorf("a null exchange spends nothing, so the remand's exchange is still owed, blue's to begin: parties %+v, to file %v, why %q", plan.Parties, plan.ToFile, plan.Why)
	}
	if want := []RemandOwed{{GapID: "G1", Direction: remandDirection}}; !slices.Equal(plan.RemandOwed, want) {
		t.Errorf("remand_owed = %+v, want %+v", plan.RemandOwed, want)
	}
	if got := remandStageOn(t, run, "blue", "blue-respond", "G1"); got != "owed" {
		t.Errorf("remand_stage = %q, want owed", got)
	}

	spent := remandedAtImpasse(t).exchangeOn("G1").actingExchangeOn("G1").register("red-chair").seed()
	after, err := PlanDispatch(spent)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after.ToFile, []string{"G1"}) || len(after.RemandOwed) != 0 {
		t.Errorf("the exchange in which blue acted is the remand's, and it left G1 at impasse: to file %v, remand_owed %+v, why %q", after.ToFile, after.RemandOwed, after.Why)
	}

	// The null exchange is not the remand's, so it cannot hold G1 at impasse: the exchange after it
	// moved G1, and G1 is below its limits.
	moved := remandedAtImpasse(t).exchangeOn("G1").movingExchangeOn("G1").register("red-chair").seed()
	below, err := PlanDispatch(moved)
	if err != nil {
		t.Fatal(err)
	}
	if !whyMentions(below, "G1: open, material") || len(below.ToFile) != 0 {
		t.Errorf("the remand's exchange moved G1 after a null one, so it is below its limits: to file %v, why %q", below.ToFile, below.Why)
	}
}

// BLUE'S ENVELOPE IS REFUSED UNTIL IT ACTS ON THE REMANDED GAP (gblock, 2026-10-25), AND THE LENS'S
// NEVER IS (ruled by gblock 2026-10-30). Blue and the minting lens each sit after the ruling. Blue's
// envelope is refused naming the remanded gap and its direction until blue records an act on it — an
// edit answering it; the lens's act clears nothing for blue. The exchange begins with blue's act, so
// the lens owes the gap nothing until blue has acted, and its envelope is not held by the remand.
func TestBlueIsRefusedItsEnvelopeUntilItActsOnTheRemandedGap(t *testing.T) {
	sat := func(t *testing.T, blueActs, lensActs bool) Run {
		st := remandedAtImpasse(t).register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1")
		st.registerAs(evLens, "lens-after-ruling").served(evLens, st.head())
		if lensActs {
			st.keepGrade(evLens, "G1", keptReason)
		}
		st.registerAs("blue-respond", "blue-after-ruling")
		if blueActs {
			st.edit("G1", "old", "new")
		}
		return st.seed()
	}
	holds := func(t *testing.T, run Run, seat string) bool {
		t.Helper()
		r, err := EnvelopeRefusal(run, seat)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Contains(r, "gap G1") && strings.Contains(r, remandDirection)
	}
	for _, c := range []struct {
		name               string
		blueActs, lensActs bool
	}{{"neither acts", false, false}, {"blue acts", true, false}, {"the lens acts", false, true}, {"both act", true, true}} {
		t.Run(c.name, func(t *testing.T) {
			run := sat(t, c.blueActs, c.lensActs)
			if got := holds(t, run, "blue-respond"); got == c.blueActs {
				t.Errorf("blue acted %v; its envelope refused on the remand: %v", c.blueActs, got)
			}
			if holds(t, run, evLens) {
				t.Errorf("the lens acted %v; its envelope is refused on the remand, whose exchange begins with blue", c.lensActs)
			}
		})
	}
}

// A REMAND PAST THE LIMIT IS REFUSED, NAMING THE DISPOSITIONS THAT CLOSE THE GAP (gblock,
// 2026-10-25). The bench remanded G1 at impasse; once that remand's exchange left it at impasse — or
// while the exchange is still owed and a party dockets it again — a remand from another sitting would
// be the gap's second at impasse, and nothing would hand its direction to any seat. The bench's work
// list says so before it rules; a ruling that closes the gap is admitted; a second motion remanded in
// the sitting that already remanded the gap is the same hearing, and is admitted.
func TestARemandPastTheLimitIsRefused(t *testing.T) {
	again := func(st *stage) Run {
		return st.register("red-chair").docketMotion("red-chair", "M2", "G1").
			dispatch(2, "judge", "G1").register("judge").seed()
	}
	for _, c := range []struct {
		name string
		run  func(*testing.T) Run
	}{
		{"after the remand's exchange", func(t *testing.T) Run { return again(remandedAtImpasse(t).actingExchangeOn("G1")) }},
		{"while the remand's exchange is owed", func(t *testing.T) Run { return again(remandedAtImpasse(t)) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := c.run(t)
			err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("M2", "a second direction"))
			if err == nil {
				t.Fatal("a second remand of G1 at impasse was admitted")
			}
			for _, want := range append([]string{"M2", "G1", RemandLimitRule}, ClosingDispositions()...) {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("the refusal does not name %q: %v", want, err)
				}
			}
			if strings.Contains(strings.Join(ClosingDispositions(), " "), "remanded") {
				t.Errorf("remanded is offered as a disposition that closes the gap: %v", ClosingDispositions())
			}
			if !workMentions(t, run, "bench", "judge", "a remand of G1 is refused") {
				t.Errorf("the bench's work list does not say, before it rules, that a remand of G1 is refused")
			}
			if err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, closingRuling("M2")); err != nil {
				t.Errorf("a ruling that closes G1 is refused: %v", err)
			}
		})
	}
	t.Run("in the sitting that remanded it", func(t *testing.T) {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
		st.exchangeOn("G1").exchangeOn("G1")
		run := st.register("red-chair").docketMotion("red-chair", "M1", "G1").
			register("blue-respond").docketMotion("blue-respond", "M2", "G1").
			register("red-chair").dispatch(2, "judge", "G1").register("judge").
			add("judge", remandRuling("M1", remandDirection)).seed()
		if err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("M2", remandDirection)); err != nil {
			t.Errorf("a second motion remanded in the sitting that remanded G1 is the same hearing, and was refused: %v", err)
		}
	})
}

// closingRuling is the bench ruling a docket motion with a disposition that closes its gap.
func closingRuling(motion string) *recordpb.MotionRule {
	return &recordpb.MotionRule{MotionId: proto.String(motion),
		Opinion: proto.String("the exchange settled nothing the record can"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED), Principle: proto.String("p"), Tension: proto.String("t"),
			ReviewFlag: proto.String("none"), Settled: proto.String("nothing"), Final: proto.Bool(true)}}}
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
// matching it out of a sentence: owed while the exchange is owed, spent once it has been had.
func TestTheWorkListStatesTheRemandStageAsAField(t *testing.T) {
	for _, c := range []struct {
		name, want string
		build      func(*testing.T) *stage
	}{
		{"owed", "owed", func(t *testing.T) *stage { return remandedAtImpasse(t) }},
		{"spent", "spent", func(t *testing.T) *stage { return remandedAtImpasse(t).actingExchangeOn("G1") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := c.build(t).register("red-chair").seed()
			for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}, {"chair", "red-chair"}} {
				if got := remandStageOn(t, run, seat.role, seat.id, "G1"); got != c.want {
					t.Errorf("%s: remand_stage = %q, want %q", seat.id, got, c.want)
				}
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
// the first, and its exchange is owed.
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
	if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || slices.Contains(gapsOf(plan, evLens), "G1") || plan.Ceiling {
		t.Errorf("the remand at impasse is G1's first, so its exchange is owed, blue's to begin: parties %+v, ceiling %v, why %q", plan.Parties, plan.Ceiling, plan.Why)
	}
	if got := remandStageOn(t, run, "blue", "blue-respond", "G1"); got != "owed" {
		t.Errorf("remand_stage = %q, want owed", got)
	}
}

// ONE BENCH SITTING IS ONE REMAND, however many of the gap's docket motions it rules: two motions on
// G1 remanded in the same sitting grant one exchange.
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
	if !slices.Contains(gapsOf(plan, "blue-respond"), "G1") || slices.Contains(gapsOf(plan, evLens), "G1") || plan.Ceiling {
		t.Errorf("one bench sitting remanding two motions is one remand, its exchange owed, blue's to begin: parties %+v, ceiling %v, why %q", plan.Parties, plan.Ceiling, plan.Why)
	}
}

// THE BENCH'S REMAND OUTRANKS MATERIALITY (gblock, 2026-10-24). A gap that is not material readies
// nobody until the bench remands it; then blue is readied for the one more exchange the remand
// grants, which begins with its act, each party's work list carries the ruling's direction, the stage is owed and
// the PASS is held — and once that exchange has had, the gap readies nobody again and holds nothing.
// On universe-m33 the remand of a gap regraded below material was stored and never acted on, and the
// run passed against the bench's holding.
func TestARemandedTrifleIsOwedItsExchange(t *testing.T) {
	for _, atImpasse := range []bool{true, false} {
		t.Run(fmt.Sprintf("at impasse %v", atImpasse), func(t *testing.T) {
			st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
				register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "low")
			if atImpasse {
				st.exchangeOn("G1").exchangeOn("G1")
			}
			run := st.docketAndRemand("M1", "G1").register("red-chair").seed()
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") ||
				len(plan.RemandOwed) != 1 || plan.RemandOwed[0].GapID != "G1" || plan.PassPermitted {
				t.Fatalf("the remand of a trifle readies blue for its exchange, and holds the PASS: parties %+v, remand_owed %+v, pass_permitted %v, why %q",
					plan.Parties, plan.RemandOwed, plan.PassPermitted, plan.Why)
			}
			for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
				if !workMentions(t, run, seat.role, seat.id, remandDirection) {
					t.Errorf("%s's work list does not carry the remand's direction", seat.id)
				}
			}
			if got := remandStageOn(t, run, "blue", "blue-respond", "G1"); got != "owed" {
				t.Errorf("remand_stage = %q, want owed", got)
			}
			if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err == nil || !strings.Contains(err.Error(), "G1") {
				t.Errorf("the gate admits a PASS over the remand's owed exchange: %v", err)
			}

			// The remand's exchange, on a stage that holds none of the report's events: the lens is
			// served the report at its head, 2, stated rather than read off this stage.
			more := st.more()
			lensAgent := fmt.Sprintf("%s-agent-%d", evLens, more.n+1)
			more.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").
				registerAs(evLens, lensAgent).served(evLens, 2).keepGrade(evLens, "G1", keptReason).stop(lensAgent).
				registerAs("blue-respond", "blue-proves-trifle").blueProves("G1").stop("blue-proves-trifle").
				register("red-chair").seed()
			after, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			if slices.Contains(gapsOf(after, evLens), "G1") || slices.Contains(gapsOf(after, "blue-respond"), "G1") || len(after.RemandOwed) != 0 {
				t.Errorf("after the remand's exchange the trifle readies nobody: parties %+v, remand_owed %+v, why %q", after.Parties, after.RemandOwed, after.Why)
			}
			bs, err := PassBlockers(run)
			if err != nil {
				t.Fatal(err)
			}
			if slices.ContainsFunc(bs, func(b Blocker) bool { return b.Subject == "G1" }) {
				t.Errorf("after the remand's exchange G1 still holds the PASS: %v", kindsOf(bs))
			}
			if workMentions(t, run, "blue", "blue-respond", remandDirection) {
				t.Errorf("blue's work list still offers the remand's exchange once it has been had")
			}
		})
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
		if workMentions(t, run, "chair", "red-chair", "for the bench again") {
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
		_, plan := seeded(t, remanded(t).actingExchangeOn("G1"))
		if !slices.Equal(plan.ToFile, []string{"G1"}) || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
			t.Errorf("the remand's exchange left G1 unmoved: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
		}
	})
}

// THE STORED PLAN CARRIES WHAT THE REMAND OWES. The workflow tells blue the remand's duty and quotes
// its direction off the chair's copy of the plan, and capture compares that copy with the plan the
// record stored, which names the gap and the direction.
func TestTheStoredPlanCarriesTheRemandOwed(t *testing.T) {
	run := remandedAtImpasse(t).register("red-chair").seed()
	plan, standing, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil || standing {
		t.Fatalf("dispatch next = (standing %v, %v)", standing, err)
	}
	if want := []RemandOwed{{GapID: "G1", Direction: remandDirection}}; !slices.Equal(plan.RemandOwed, want) {
		t.Fatalf("setup: remand_owed = %+v, want %+v", plan.RemandOwed, want)
	}
	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := StoredPlans(fam.Events)
	if err != nil || len(stored) == 0 {
		t.Fatalf("stored plans = %d, %v", len(stored), err)
	}
	if got := stored[len(stored)-1].Plan.RemandOwed; !slices.Equal(got, plan.RemandOwed) {
		t.Errorf("the stored plan owes %+v, the printed one %+v", got, plan.RemandOwed)
	}
}

// A LENS'S ACT BEFORE BLUE'S BEGINS NO REMAND'S EXCHANGE (ruled by gblock 2026-10-27). The lenses sit
// before blue in an epoch, so the minting lens acts on the remanded gap first: on universe-m35 its
// regrade ("Blue has not answered the remand") was counted as the exchange, blue's item vanished when
// blue's sitting closed, and blue never acted. Here the lens keeps the grade and blue sits without
// touching G1: the exchange has not begun — the plan readies blue again with the direction, and not
// the lens, nothing goes back to the bench, the stage is owed — the lens's item does not hold it, and
// blue's still blocks. The exchange in which blue then acts is the remand's.
func TestALensActBeforeBlueLeavesTheRemandExchangeUnbegun(t *testing.T) {
	run := remandedAtImpasse(t).lensOnlyExchangeOn("G1").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if want := []RemandOwed{{GapID: "G1", Direction: remandDirection}}; !slices.Equal(plan.RemandOwed, want) || len(plan.ToFile) != 0 ||
		slices.Contains(gapsOf(plan, evLens), "G1") || !slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Errorf("only the lens acted, so the remand's exchange is still owed: remand_owed %+v, to file %v, parties %+v, why %q",
			plan.RemandOwed, plan.ToFile, plan.Parties, plan.Why)
	}
	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}, {"chair", "red-chair"}} {
		if got := remandStageOn(t, run, seat.role, seat.id, "G1"); got != "owed" {
			t.Errorf("%s: remand_stage = %q, want owed", seat.id, got)
		}
	}
	if r, err := EnvelopeRefusal(run, evLens); err != nil || strings.Contains(r, remandDirection) {
		t.Errorf("the remand does not hold the lens's envelope: %q (%v)", r, err)
	}
	if !remandItemBlocks(t, run, "blue", "blue-respond") {
		t.Errorf("blue has not acted on G1, and its remand item does not block its list")
	}

	spent := remandedAtImpasse(t).lensOnlyExchangeOn("G1").actingExchangeOn("G1").register("red-chair").seed()
	after, err := PlanDispatch(spent)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after.ToFile, []string{"G1"}) || len(after.RemandOwed) != 0 {
		t.Errorf("blue acted in the next exchange, which is the remand's, and it left G1 at impasse: to file %v, remand_owed %+v, why %q", after.ToFile, after.RemandOwed, after.Why)
	}
}

// remandItemBlocks reports whether the seat's work list holds the remand's item on G1 as blocking —
// and fails the test if it holds it at all with a reason that says it does not.
func remandItemBlocks(t *testing.T, run Run, role, seat string) bool {
	t.Helper()
	w, err := WorkOfSeat(run, role, seat)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range w.Sitting.Open {
		if strings.Contains(it.What, "gap G1") && strings.Contains(it.What, remandDirection) {
			return it.Blocks
		}
	}
	return false
}

// BLUE'S REMAND ITEM BLOCKS, AND SAYS SO TRULY, BEFORE AND AFTER ITS REGISTER, AND STAYS WHILE OWED.
// Blue's harness bracket names no seat — its configuration sits several — so until blue registers,
// the latest sitting of blue-respond is the one before the ruling. On universe-m35 the item read "the
// ruling landed after this sitting opened" (not blocking) before register, blocked after it, and
// vanished when the sitting closed. Read off the dispatch blue is sitting for, it blocks at each step
// while the exchange is owed, and stops blocking only once blue has acted.
func TestBlueRemandItemBlocksBeforeAndAfterRegister(t *testing.T) {
	st := remandedAtImpasse(t).lensActs("G1")
	st.add(HarnessSeat, &recordpb.SittingOpen{AgentId: proto.String("blue-m35"), AgentType: proto.String("frank-exchange-of-views:blue-researcher")})
	run := st.seed()
	steps := []struct {
		name  string
		act   func(*stage)
		block bool
	}{
		{"bracketed, before register", func(*stage) {}, true},
		{"registered", func(b *stage) { b.registerAs("blue-respond", "blue-m35") }, true},
		{"its sitting closed without an act", func(b *stage) { b.stop("blue-m35") }, true},
		{"dispatched again and acted", func(b *stage) {
			b.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-again").blueProves("G1")
		}, false},
	}
	for _, s := range steps {
		// Each step's acts are appended to the seeded record (more), so the record grows as the
		// sitting does.
		st = st.more()
		s.act(st)
		st.seed()
		w, err := WorkOfSeat(run, "blue", "blue-respond")
		if err != nil {
			t.Fatal(err)
		}
		var item *Item
		for i, it := range w.Sitting.Open {
			if strings.Contains(it.What, "gap G1") && strings.Contains(it.What, remandDirection) {
				item = &w.Sitting.Open[i]
			}
		}
		switch {
		case item == nil:
			t.Errorf("%s: blue's remand item on G1 vanished while the exchange is owed", s.name)
		case item.Blocks != s.block:
			t.Errorf("%s: blue's remand item blocks %v, want %v: %q", s.name, item.Blocks, s.block, item.What)
		case strings.Contains(item.What, "landed after"):
			t.Errorf("%s: blue's remand item says the ruling landed after its dispatch, and the ruling preceded it: %q", s.name, item.What)
		}
	}
}

// A REMAND IS ANSWERED ONLY AFTER IT (ruled by gblock 2026-10-27). Blue answered G1 with an edit, the
// exchanges after it stalled it to impasse, and the bench remanded it: the edit is what the bench
// heard, and it answers nothing the remand asks. On universe-m35 the lens's list said "ANSWERED by an
// edit" beside the remand item while every answering edit predated the ruling. Once blue edits after
// the ruling, the lens's list says it was answered.
func TestAnEditBeforeTheRemandDoesNotAnswerIt(t *testing.T) {
	answered := "gap G1 has been ANSWERED by a change to the report"
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.movingExchangeOn("G1").exchangeOn("G1").exchangeOn("G1").docketAndRemand("M1", "G1")
	run := st.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").seed()
	if got := remandStageOn(t, run, "lens", evLens, "G1"); got != "owed" {
		t.Fatalf("setup: remand_stage = %q, want owed", got)
	}
	if workMentions(t, run, "lens", evLens, answered) {
		t.Errorf("the lens's list counts blue's edit from before the remand as its answer")
	}
	st.more().registerAs("blue-respond", "blue-edits-after").edit("G1", "old after", "new after").stop("blue-edits-after").seed()
	if !workMentions(t, run, "lens", evLens, answered) {
		t.Errorf("blue edited G1 after the remand, and the lens's list does not say it was answered")
	}
}
