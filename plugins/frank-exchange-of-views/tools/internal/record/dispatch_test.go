package record

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// stage is a seeded record built one act at a time, with the keys and timestamps the seed needs
// and the ids the folds compare against. Seeding bypasses the write path on purpose: these tests
// are about what the READERS make of a record, not about what the writers refuse.
type stage struct {
	t      *testing.T
	runDir string
	n      int
	evs    []*Event
}

func newStage(t *testing.T) *stage { return &stage{t: t, runDir: newRun(t)} }

func (b *stage) add(seat string, body proto.Message) *stage {
	b.n++
	b.evs = append(b.evs, recordtest.At(b.t, seat, fmt.Sprintf("%s:%d", seat, b.n), body))
	return b
}
func (b *stage) register(seat string) *stage { return b.add(seat, &recordpb.Register{}) }
func (b *stage) cast(seats ...string) *stage {
	return b.add("harness", &recordpb.Cast{SeatIds: seats})
}
func (b *stage) ingest() *stage {
	return b.add("harness", &recordpb.BaseIngest{Text: proto.String("# report")})
}
func (b *stage) mint(evLens, gap, severity string) *stage {
	sev, _ := GradeOf(severity)
	return b.add(evLens, &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
}

// mintSuperseding mints gap as the successor of ancestor — lineage the record keeps and the PASS
// gate reads: an open ancestor is a stranded gap.
func (b *stage) mintSuperseding(evLens, gap, severity, ancestor string) *stage {
	sev, _ := GradeOf(severity)
	return b.add(evLens, &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Supersedes: []string{ancestor}})
}

// docketMotion is a party escalating gap to the bench by hand — the route `motion docket file`
// keeps open beside the dispatch's own docketing at impasse.
func (b *stage) docketMotion(seat, id, gap string) *stage {
	return b.add(seat, &recordpb.Motion{MotionId: proto.String(id), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
		Basis: proto.String("escalated by hand"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}}})
}

func (b *stage) dispatch(pin int64, seat string, gaps ...string) *stage {
	return b.add("red-chair", &recordpb.Dispatch{Pin: proto.Int64(pin), SeatId: proto.String(seat), GapIds: gaps})
}
func (b *stage) edit(gap, old, new string) *stage {
	return b.add("blue-respond", &recordpb.BlueEdit{Answers: proto.String(gap), Old: proto.String(old), New: proto.String(new), EditKey: proto.String(fmt.Sprintf("E%d", b.n+1))})
}
func (b *stage) regrade(evLens, gap, severity string) *stage {
	sev, _ := GradeOf(severity)
	return b.add(evLens, &recordpb.Regrade{GapId: proto.String(gap), Severity: &sev, Basis: proto.String("b")})
}
func (b *stage) seed() Run {
	recordtest.Seed(b.t, b.runDir, b.evs...)
	return mustRun(b.t, b.runDir)
}

const evLens = "red-lens-evidence"

// One exchange is a red-party sitting engaged on G followed by a blue-party sitting engaged on G,
// both complete (the chair sat again). Movement in it resets the stall; a null turn — blue sat
// and recorded nothing on G — counts as an exchange and stalls it.
func TestAGapsExchangesAndStallsAreCountedFromTheRecord(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").                 // epoch 1: the lens mints
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1"). // epoch 2: the chair engages both
		register(evLens).register("blue-respond").edit("G1", "was", "is").                 // both sit; blue moves the text
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1"). // epoch 3
		register(evLens).register("blue-respond").                                         // blue sits and says nothing on G1
		register("red-chair")                                                              // epoch 4: the sittings above are complete
	run := b.seed()
	x, err := Exchanges(run, DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	g := x["G1"]
	if g == nil || g.Exchanges != 2 || g.Stalled != 1 || g.Impasse {
		t.Fatalf("G1 = %+v, want 2 exchanges, 1 stalled (the null turn), not at impasse under K=2", g)
	}
}

func TestASilentSittingIsANullTurnThatStillCountsAnExchange(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond"). // silence from both
		register("red-chair")
	x, err := Exchanges(b.seed(), DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	if g := x["G1"]; g.Exchanges != 1 || g.Stalled != 1 {
		t.Fatalf("a silent blue sitting = %+v, want one exchange, one stall — silence is a turn taken", g)
	}
	// The same shape with blue's sitting still OPEN (the chair has not sat again) counts nothing yet.
	b2 := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond")
	x2, _ := Exchanges(b2.seed(), DefaultParams)
	if g := x2["G1"]; g.Exchanges != 0 {
		t.Fatalf("an open sitting completed an exchange: %+v", g)
	}
}

// Impasse by stall: K consecutive exchanges without movement. Impasse by the monotone bound:
// KMax exchanges however much moved — a regrade per exchange keeps the stall at 0 forever and
// the bound still fires.
func TestImpasseFiresOnStallAndOnTheMonotoneBound(t *testing.T) {
	cycle := func(b *stage, move func(*stage)) {
		b.register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").register(evLens).register("blue-respond")
		if move != nil {
			move(b)
		}
	}
	stall := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().register("red-chair").register(evLens).mint(evLens, "G1", "high")
	cycle(stall, nil)
	cycle(stall, nil)
	stall.register("red-chair")
	x, _ := Exchanges(stall.seed(), Params{K: 2, KMax: 6, MintBudget: 5, ConvergenceFraction: 0.25})
	if g := x["G1"]; !g.Impasse || g.Stalled != 2 {
		t.Fatalf("two stalled exchanges under K=2: %+v, want impasse", g)
	}

	oscillate := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().register("red-chair").register(evLens).mint(evLens, "G1", "high")
	grades := []string{"medium", "high", "medium"}
	for i := 0; i < 3; i++ {
		g := grades[i]
		cycle(oscillate, func(b *stage) { b.regrade(evLens, "G1", g) })
	}
	oscillate.register("red-chair")
	x, _ = Exchanges(oscillate.seed(), Params{K: 2, KMax: 3, MintBudget: 5, ConvergenceFraction: 0.25})
	if g := x["G1"]; !g.Impasse || g.Stalled != 0 || g.Exchanges != 3 {
		t.Fatalf("a regrade every exchange under KMax=3: %+v, want impasse by the bound with no stall", g)
	}
	// A regrade to the SAME grade is an act, not movement.
	same := newStage(t).cast(evLens, "red-chair", "blue-respond").ingest().register("red-chair").register(evLens).mint(evLens, "G1", "high")
	cycle(same, func(b *stage) { b.regrade(evLens, "G1", "high") })
	same.register("red-chair")
	x, _ = Exchanges(same.seed(), DefaultParams)
	if g := x["G1"]; g.Stalled != 1 {
		t.Fatalf("a regrade to the same grade counted as movement: %+v", g)
	}
}

// A gap below material readies nobody. The lens is still ready under rule 1 when the head moved.
func TestASubMaterialGapDoesNotReadyBlue(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "low")
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parties) != 1 || plan.Parties[0].SeatID != evLens || len(plan.Parties[0].GapIDs) != 0 {
		t.Fatalf("parties = %+v, want only the lens, engaged on nothing (rule 1)", plan.Parties)
	}
	if !strings.Contains(strings.Join(plan.Why, "\n"), "G1: open but below material") {
		t.Errorf("the plan must say why the trifle readies nobody:\n%s", strings.Join(plan.Why, "\n"))
	}
	if plan.PassPermitted {
		t.Error("PASS is not permitted while a cast lens has not sat against the head")
	}
}

// A HAND-FILED DOCKET MOTION READIES THE BENCH BEFORE IMPASSE, and on a trifle. The escalation
// route is a party's, not only the dispatch's; the motion stands until ruled, and PASS is refused
// while it stands — so a docket nobody is dispatched to rule would strand the run.
func TestAHandFiledDocketReadiesTheBenchWithoutImpasse(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest(). // head = 2
											register("red-chair").dispatch(2, evLens).register(evLens).
											mint(evLens, "G1", "high").mint(evLens, "G2", "low").
											register("blue-respond").docketMotion("blue-respond", "M1", "G1").docketMotion("blue-respond", "M2", "G2").
											register("red-chair")
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	engaged := map[string][]string{}
	for _, p := range plan.Parties {
		engaged[p.SeatID] = p.GapIDs
	}
	if got := engaged["judge"]; len(got) != 2 || got[0] != "G1" || got[1] != "G2" {
		t.Errorf("the bench is engaged on %v, want G1 and G2 — both escalated by hand, the trifle included", got)
	}
	if _, ready := engaged[evLens]; ready {
		t.Error("the lens is engaged on a gap that is the bench's now")
	}
	if len(plan.Docket) != 0 {
		t.Errorf("the plan dockets %v again — a motion already stands on each", plan.Docket)
	}
	if plan.PassPermitted {
		t.Error("PASS is not permitted while a docket motion stands unruled")
	}
	why := strings.Join(plan.Why, "\n")
	if !strings.Contains(why, "G1: docketed and unruled — the bench is ready") || !strings.Contains(why, "G2: docketed and unruled — the bench is ready") {
		t.Errorf("the plan must say the bench is ready for each:\n%s", why)
	}
}

// A STRANDED ANCESTOR READIES ITS MINTER WHATEVER ITS GRADE. The PASS gate refuses a verdict while
// a superseded gap is open, so a sub-material ancestor that readied nobody left the plan saying
// "pass permitted" and the gate saying no — the run ended UNVERIFIED with nobody ready (found by
// the release sweep). Held as material: its minter and blue are engaged on it.
func TestAStrandedAncestorIsReadyWorkWhateverItsGrade(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest(). // head = 2
											register("red-chair").dispatch(2, evLens).register(evLens).
											mint(evLens, "G1", "low").mintSuperseding(evLens, "G2", "low", "G1").
											register("red-chair")
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	engaged := map[string][]string{}
	for _, p := range plan.Parties {
		engaged[p.SeatID] = p.GapIDs
	}
	if got := engaged[evLens]; len(got) != 1 || got[0] != "G1" {
		t.Errorf("the minter is engaged on %v, want G1 alone — the stranded ancestor, not the trifle that supersedes it", got)
	}
	if got := engaged["blue-respond"]; len(got) != 1 || got[0] != "G1" {
		t.Errorf("blue is engaged on %v, want G1 — the exchange must count toward impasse or the ancestor can never reach the bench", got)
	}
	if plan.PassPermitted {
		t.Error("PASS is not permitted over a stranded ancestor")
	}
	why := strings.Join(plan.Why, "\n")
	if !strings.Contains(why, "G1: open and superseded by G2") || !strings.Contains(why, "G2: open but below material") {
		t.Errorf("the plan must say the ancestor is held as material and the successor readies nobody:\n%s", why)
	}
}

// Empty is the termination signal: every cast lens sat against the head and no material gap is
// open — PASS permitted. With a material gap open and below its limits, its two parties are ready.
func TestAnEmptyDispatchIsTermination(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest(). // head = 2
											register("red-chair").dispatch(2, evLens).register(evLens). // the lens sat against the head
											register("red-chair")
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Parties) != 0 || !plan.PassPermitted || plan.Ceiling {
		t.Fatalf("plan = %+v, want no parties and pass_permitted", plan)
	}

	b2 := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high").
		register("red-chair")
	plan2, _ := PlanDispatch(b2.seed())
	want := map[string][]string{evLens: {"G1"}, "blue-respond": {"G1"}}
	if len(plan2.Parties) != 2 || plan2.PassPermitted {
		t.Fatalf("plan = %+v, want the lens and blue engaged on G1", plan2)
	}
	for _, p := range plan2.Parties {
		if w := want[p.SeatID]; len(w) != len(p.GapIDs) || (len(w) > 0 && w[0] != p.GapIDs[0]) {
			t.Errorf("%s engaged on %v, want %v", p.SeatID, p.GapIDs, w)
		}
	}
}

// At impasse the plan dockets the gap and readies the bench — nobody else.
func TestTheBenchIsReadyWhenAGapReachesImpasseAndTheVerbDocketsIt(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high")
	for i := 0; i < 2; i++ {
		b.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register(evLens).register("blue-respond")
	}
	b.register("red-chair")
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Docket) != 1 || plan.Docket[0] != "G1" {
		t.Fatalf("docket = %v, want [G1]", plan.Docket)
	}
	if len(plan.Parties) != 1 || plan.Parties[0].SeatID != "judge" || plan.Parties[0].GapIDs[0] != "G1" {
		t.Fatalf("parties = %+v, want the bench on G1 alone", plan.Parties)
	}
	if plan.PassPermitted || plan.Ceiling {
		t.Errorf("neither PASS nor CEILING while the bench has not ruled: %+v", plan)
	}
}

// No cast, no dispatch: the verb refuses rather than inventing an admissible list.
func TestTheDispatchPlanRefusesARecordWithNoCast(t *testing.T) {
	b := newStage(t).ingest().register("red-chair")
	if _, err := PlanDispatch(b.seed()); err == nil || !strings.Contains(err.Error(), "no cast") {
		t.Fatalf("a record with no cast planned a dispatch: %v", err)
	}
}

// The cast is refused at both doors: a dispatch naming a seat outside it, and a seat outside it
// registering. A record with no cast refuses the dispatch (nothing to admit against) but admits
// the register — the fixtures' world, and a migrated archive's before its cast is synthesized.
func TestTheDispatchVerbRefusesASeatOutsideTheCast(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	if _, err := Append(chair, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-voice")}); err == nil || !strings.Contains(err.Error(), "not in this run's cast") {
		t.Fatalf("a dispatch to a seat outside the cast was written: %v", err)
	}
	if _, err := Append(chair, &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String(evLens)}); err != nil {
		t.Fatalf("a dispatch to a cast seat was refused: %v", err)
	}
	if _, err := Append(chair, &recordpb.Dispatch{SeatId: proto.String(evLens)}); err == nil || !strings.Contains(err.Error(), "pin") {
		t.Fatalf("a dispatch with no pin was written: %v", err)
	}
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "red-lens-voice"}, ""); err == nil || !strings.Contains(err.Error(), "not in this run's cast") {
		t.Fatalf("a seat outside the cast registered: %v", err)
	}
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "judge-petition-red-chair"}, ""); err != nil {
		t.Fatalf("a petition sitting by a cast seat was refused: %v", err)
	}
	if _, err := Append(Identity{Run: run, SeatID: "harness"}, &recordpb.Cast{SeatIds: []string{"judge"}}); err == nil || !strings.Contains(err.Error(), "written ONCE") {
		t.Fatalf("a second cast was written: %v", err)
	}

	bare := newStage(t).ingest().register("red-chair").seed()
	if _, err := Append(Identity{Run: bare, SeatID: "red-chair"}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(evLens)}); err == nil || !strings.Contains(err.Error(), "no cast") {
		t.Fatalf("a dispatch on a record with no cast was written: %v", err)
	}
	if _, _, err := RegisterSeat(Identity{Run: bare, SeatID: "red-lens-voice"}, ""); err != nil {
		t.Fatalf("with no cast on the record, any dispatchable seat registers: %v", err)
	}
}

// PASS waits for every cast lens to have sat against the head; a lens engaged that never
// registered has not sat.
func TestPassIsRefusedUntilEveryCastLensSatAgainstTheHead(t *testing.T) {
	logic := "red-lens-logic"
	b := newStage(t).cast(evLens, logic, "red-chair", "blue-respond", "judge").ingest(). // head 2
												register("red-chair").dispatch(2, evLens).dispatch(2, logic).register(evLens). // only evidence sits
												register("red-chair")
	run := b.seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	pass := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}
	if _, err := Append(chair, pass); err == nil || !strings.Contains(err.Error(), logic) {
		t.Fatalf("PASS landed with %s never having sat against the head: %v", logic, err)
	}
	// The dispatch row alone does not count: logic was engaged. It sits now.
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: logic}, ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := RegisterSeat(chair, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(chair, pass); err != nil {
		t.Fatalf("PASS refused after every cast lens sat against the head: %v", err)
	}
}

// An open gap below material does not hold the gate; a material one does.
func TestABelowMaterialGapDoesNotHoldTheGate(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "low").
		register("red-chair")
	run := b.seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	pass := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_PASS)}
	if _, err := Append(chair, pass); err != nil {
		t.Fatalf("a low-severity open gap held the gate: %v", err)
	}
	b2 := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
		register("red-chair")
	if _, err := Append(Identity{Run: b2.seed(), SeatID: "red-chair"}, pass); err == nil || !strings.Contains(err.Error(), "material gap(s) still OPEN: G1") {
		t.Fatalf("a MEDIUM gap is material and must hold the gate: %v", err)
	}
}

// A lens's mint budget is the record's count of its mints against the run's term, superseding
// mints included; the chair is not a lens and is not bounded here.
func TestAMintPastTheBudgetIsRefused(t *testing.T) {
	runDir := newRun(t)
	writeRunConfig(t, runDir, `{"mintBudget":1}`)
	run := mustRun(t, runDir)
	lensID := Identity{Run: run, SeatID: evLens}
	if _, _, err := RegisterSeat(lensID, ""); err != nil {
		t.Fatal(err)
	}
	mint := func(id string, supersedes ...string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(id), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Supersedes: supersedes}
	}
	if _, err := Append(lensID, mint("G1")); err != nil {
		t.Fatalf("the first mint within a budget of 1 was refused: %v", err)
	}
	if _, err := Append(lensID, mint("G2")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("the second mint past a budget of 1 landed: %v", err)
	}
	if _, err := Append(lensID, mint("G3", "G1")); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("a superseding mint past the budget landed — lineage is a new gap: %v", err)
	}
	// Another lens has its own budget.
	other := Identity{Run: run, SeatID: "red-lens-logic"}
	if _, _, err := RegisterSeat(other, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(other, mint("G4")); err != nil {
		t.Fatalf("another lens's first mint was refused: %v", err)
	}
}

// A FAIL over a converged board is refused: nothing material open, mass a small fraction of the
// run's peak, no fresh material mint this epoch. A fresh material mint lifts the refusal.
func TestAFailOnAConvergentBoardIsRefused(t *testing.T) {
	stage := func() (*stage, Run) {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").register(evLens)
		// Epoch 1: a serious gap (mass 9) and a FAIL — the peak the run will be measured against.
		sev, _ := GradeOf("high")
		st.add(evLens, &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &sev,
			Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)})
		st.add("red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
		// Epoch 2: G1 closes; only a trifle (mass 1) is minted.
		st.register("red-chair").register(evLens).
			add(evLens, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
				AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")})
		low, _ := GradeOf("low")
		st.add(evLens, &recordpb.Mint{GapId: proto.String("G2"), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
			CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &low,
			Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)})
		return st, Run{}
	}
	st, _ := stage()
	run := st.seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	fail := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}
	if _, err := Append(chair, fail); err == nil || !strings.Contains(err.Error(), "converged") {
		t.Fatalf("a FAIL over a converged board landed: %v", err)
	}
	c, _ := convergenceOf(run)
	if c.Peak != 9 || c.Mass != 1 || c.MaxSeverityMass != 1 || c.FreshMaterialMints != 0 {
		t.Errorf("convergence = %+v, want peak 9, mass 1, top severity 1, no fresh material mint", c)
	}
}

func TestAMaterialFreshMintBlocksTheConvergenceRefusal(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").register(evLens)
	high, _ := GradeOf("high")
	st.add(evLens, &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &high,
		Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)})
	st.add("red-chair", &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)})
	st.register("red-chair").register(evLens).
		add(evLens, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")})
	// Fresh, superseding nothing, MEDIUM: material. Its mass (1) is still under the fraction; the
	// mint alone is what keeps the FAIL legal.
	med, _ := GradeOf("medium")
	st.add(evLens, &recordpb.Mint{GapId: proto.String("G2"), Class: proto.String("x"), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &med,
		Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)})
	run := st.seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}); err != nil {
		t.Fatalf("a FAIL with a fresh material mint this epoch was refused: %v", err)
	}
}

// A gap belongs to the lens that minted it: only that lens regrades or closes it. The chair may
// still carry a closure the archive holds; the bench disposes through its ruling.
func TestOnlyTheMintingLensMayRegradeOrCloseItsGap(t *testing.T) {
	run := newStage(t).cast(evLens, "red-lens-logic", "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).register("red-lens-logic").mint(evLens, "G1", "high").seed()
	closeG1 := func() *recordpb.Close {
		return &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
			AnchorSeat: proto.String("L1"), AnchorTool: proto.String("t"), AnchorTarget: proto.String("x"), Prose: proto.String("fixed")}
	}
	regrade := &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Basis: proto.String("b")}
	for _, other := range []string{"red-chair", "red-lens-logic"} {
		if _, err := Append(Identity{Run: run, SeatID: other}, regrade); err == nil || !strings.Contains(err.Error(), "minted by "+evLens) {
			t.Errorf("%s regraded another lens's gap: %v", other, err)
		}
		if _, err := Append(Identity{Run: run, SeatID: other}, closeG1()); err == nil || !strings.Contains(err.Error(), "originator") {
			t.Errorf("%s closed another lens's gap: %v", other, err)
		}
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, regrade); err != nil {
		t.Fatalf("the originator's regrade was refused: %v", err)
	}
	if _, err := Append(Identity{Run: run, SeatID: evLens}, closeG1()); err != nil {
		t.Fatalf("the originator's close was refused: %v", err)
	}
	// The chair may carry the closure the archive now holds.
	if _, _, err := RegisterSeat(Identity{Run: run, SeatID: "red-chair"}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, &recordpb.Close{GapId: proto.String("G1"),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), CarriedFrom: proto.String("1"), Prose: proto.String("as before")}); err != nil {
		t.Fatalf("the chair's carry of an archived closure was refused: %v", err)
	}
}
