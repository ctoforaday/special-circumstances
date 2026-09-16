package record

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// sit is one sitting of a lens: a dispatch naming it at pin, and its register.
func (b *stage) sit(pin int64, lens string) *stage { return b.dispatch(pin, lens).register(lens) }

// retBoard is the retirement fixtures' base: one cast lens, a report ingested at head 2, the chair
// seated. Each test adds the lens's sittings, mints and head moves it is about.
func retBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
}

// freshHigh is a mint of a by_grade class graded high: fresh, and material now.
func freshHigh(gap string) *recordpb.Mint {
	return cmMint(gap, recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high")
}

func planOf(t *testing.T, b *stage) Plan {
	t.Helper()
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func lensReady(p Plan) bool {
	for _, party := range p.Parties {
		if party.SeatID == evLens {
			return true
		}
	}
	return false
}

func whyOf(p Plan) string { return strings.Join(p.Why, "\n") }

// lensStateOf is the fold's state for the one cast lens, asked of the record.
func lensStateOf(t *testing.T, run Run) lensFold {
	t.Helper()
	g, ok, err := passLensGateOfRun(run)
	if err != nil || !ok || len(g.lenses) != 1 {
		t.Fatalf("gate: %+v %v %v", g, ok, err)
	}
	return g.lenses[0]
}

func TestLensRetiresAfterTwoBarrenSittings(t *testing.T) {
	one := planOf(t, retBoard(t).sit(2, evLens))
	if !lensReady(one) || !strings.Contains(whyOf(one), "active (1 barren sitting(s)") {
		t.Fatalf("after one barren sitting the lens is active and ready: %+v", one)
	}
	two := planOf(t, retBoard(t).sit(2, evLens).sit(2, evLens))
	if lensReady(two) || !strings.Contains(whyOf(two), "retired at the head") {
		t.Fatalf("after two barren sittings the lens is retired and not ready: %+v", two)
	}
}

func TestRetiredLensReArmedOnceByHeadMove(t *testing.T) {
	plan := planOf(t, retBoard(t).sit(2, evLens).sit(2, evLens).ingest())
	if !lensReady(plan) || !strings.Contains(whyOf(plan), "retired, re-armed once") {
		t.Fatalf("a retired lens is ready once when the head moves past its pin: %+v", plan)
	}
}

func TestBarrenReArmRetiresForGood(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest()
	head := int64(b.n)
	plan := planOf(t, b.sit(head, evLens))
	if lensReady(plan) || !strings.Contains(whyOf(plan), "retired for good") {
		t.Fatalf("a barren re-arm sitting retires the lens for good: %+v", plan)
	}
}

func TestRetiredForGoodIgnoresLaterHeadMoves(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest()
	head := int64(b.n)
	plan := planOf(t, b.sit(head, evLens).ingest().ingest())
	if lensReady(plan) || !strings.Contains(whyOf(plan), "retired for good") {
		t.Fatalf("head moves do not re-arm a lens retired for good: %+v", plan)
	}
}

func TestFreshMintOnReArmMakesLensActiveAgain(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest()
	head := int64(b.n)
	b.sit(head, evLens).add(evLens, freshHigh("G1"))
	run := b.seed()
	if f := lensStateOf(t, run); f.state != LensActive || f.barren != 0 || !f.ready {
		t.Fatalf("a productive re-arm sitting returns the lens to active with its re-arm unspent: %+v", f)
	}
}

func TestActiveLensReadyAtSameHead(t *testing.T) {
	plan := planOf(t, retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")))
	if !lensReady(plan) || !strings.Contains(whyOf(plan), "active (0 barren sitting(s)") {
		t.Fatalf("an active lens is ready at the same head: %+v", plan)
	}
}

func TestNoLensReadyBeforeIngest(t *testing.T) {
	plan := planOf(t, newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").register("red-chair"))
	if lensReady(plan) || plan.PassPermitted || !strings.Contains(whyOf(plan), "no report has been ingested") {
		t.Fatalf("nothing is ready and nothing permits a PASS before a report is ingested: %+v", plan)
	}
}

func TestLineageMintIsNotFresh(t *testing.T) {
	// Sitting 1 mints fresh material; sittings 2 and 3 mint only successors of it — lineage, not
	// discovery — so they are barren and the lens retires.
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).
		sit(2, evLens).add(evLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high", "G1")).
		sit(2, evLens).add(evLens, cmMint("G3", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high", "G2"))
	if f := lensStateOf(t, b.seed()); f.state != LensRetired {
		t.Fatalf("sittings that minted only lineage are barren: %+v", f)
	}
}

func TestNeverClassMintIsNotFresh(t *testing.T) {
	b := retBoard(t).sit(2, evLens).add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")).
		sit(2, evLens).add(evLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	if f := lensStateOf(t, b.seed()); f.state != LensRetired {
		t.Fatalf("sittings that minted only never-class gaps are barren: %+v", f)
	}
}

// retiredForGoodBehindHead is a lens that retired, spent its re-arm on a barren sitting, and has
// had the head move past it since: stale.
func retiredForGoodBehindHead(t *testing.T) (*stage, int64) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest()
	pin := int64(b.n)
	b.sit(pin, evLens).ingest()
	return b, pin
}

func TestPassPermittedWithLensRetiredForGoodBehindHead(t *testing.T) {
	b, _ := retiredForGoodBehindHead(t)
	plan := planOf(t, b.register("red-chair"))
	if !plan.PassPermitted || lensReady(plan) {
		t.Fatalf("a lens retired for good behind the head is not ready, so the plan permits a PASS: %+v", plan)
	}
}

func TestPassRefusedWhileReArmOwed(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest().register("red-chair")
	run := b.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Fatalf("pass_permitted over an owed re-arm: %+v", plan)
	}
	_, err = Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
	if err == nil || !strings.Contains(err.Error(), evLens) || !strings.Contains(err.Error(), "re-armed once") {
		t.Fatalf("a chair's PASS over an owed re-arm must be refused at the write path, naming the lens: %v", err)
	}
}

// THE PLAN AND THE GATE READ ONE FOLD: in every retirement state, pass_permitted and the write
// path's answer to a PASS agree.
func TestPassPermittedAndGateShareTheFold(t *testing.T) {
	for _, c := range []struct {
		name  string
		build func(t *testing.T) *stage
	}{
		{"active", func(t *testing.T) *stage { return retBoard(t).sit(2, evLens) }},
		{"retired at the head", func(t *testing.T) *stage { return retBoard(t).sit(2, evLens).sit(2, evLens) }},
		{"retired, re-arm owed", func(t *testing.T) *stage { return retBoard(t).sit(2, evLens).sit(2, evLens).ingest() }},
		{"retired for good, stale area read", func(t *testing.T) *stage {
			b, _ := retiredForGoodBehindHead(t)
			b.register("red-chair")
			return b.add("red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the changes since the pin")})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := c.build(t)
			if c.name != "retired for good, stale area read" {
				b.register("red-chair")
			}
			run := b.seed()
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			_, err = Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
			if plan.PassPermitted != (err == nil) {
				t.Fatalf("pass_permitted=%v but the PASS appended=%v (%v)", plan.PassPermitted, err == nil, err)
			}
		})
	}
}

func TestBudgetExhaustedLensRetires(t *testing.T) {
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).sit(2, evLens).sit(2, evLens)
	writeRunConfig(t, b.runDir, `{"mintBudget":1}`)
	run := b.seed()
	if _, err := Append(Identity{Run: run, SeatID: evLens}, liveMint("G2", "x", recordpb.Grade_GRADE_HIGH)); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("fixture: the lens is not at its budget: %v", err)
	}
	if f := lensStateOf(t, run); f.state != LensRetired {
		t.Fatalf("a lens at its budget mints nothing, so its sittings are barren and it retires: %+v", f)
	}
}

func TestWhyCarriesState(t *testing.T) {
	b, _ := retiredForGoodBehindHead(t)
	for _, c := range []struct {
		name string
		plan Plan
		want string
	}{
		{"never sat", planOf(t, retBoard(t)), "active, never sat"},
		{"active", planOf(t, retBoard(t).sit(2, evLens)), "active (1 barren sitting(s)"},
		{"retired at the head", planOf(t, retBoard(t).sit(2, evLens).sit(2, evLens)), "retired at the head (pin 2)"},
		{"re-armed", planOf(t, retBoard(t).sit(2, evLens).sit(2, evLens).ingest()), "retired, re-armed once"},
		{"retired for good", planOf(t, b), "retired for good"},
	} {
		if !strings.Contains(whyOf(c.plan), evLens+": "+c.want) {
			t.Errorf("%s: the why must name the state %q:\n%s", c.name, c.want, whyOf(c.plan))
		}
	}
}

// THE LENS READS ITS LAST SITTING FROM ITS WORK VIEW, after registering for the dispatch it is
// sitting for — which is why the fact cannot come from lensPins, whose answer is the current
// sitting once the register lands.
func TestLensWorkCarriesLastSitting(t *testing.T) {
	last := func(t *testing.T, b *stage) LastSittingJSON {
		t.Helper()
		run := b.seed()
		out, err := WorkJSONBytes(run, "lens", evLens)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), `"last_sitting"`) {
			t.Fatalf("the lens's work view carries no last_sitting key:\n%s", out)
		}
		var w struct {
			Sitting struct {
				LastSitting LastSittingJSON `json:"last_sitting"`
			} `json:"sitting"`
		}
		if err := json.Unmarshal(out, &w); err != nil {
			t.Fatal(err)
		}
		return w.Sitting.LastSitting
	}
	if got := last(t, retBoard(t)); got.Kind != "undispatched" {
		t.Errorf("never dispatched: %+v", got)
	}
	if got := last(t, retBoard(t).sit(2, evLens)); got.Kind != "first" || got.Head != 2 {
		t.Errorf("first sitting, registered: %+v", got)
	}
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).
		add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b"), EditKey: proto.String("E1")})
	head := int64(b.n)
	if got := last(t, b.sit(head, evLens)); got.Kind != "behind" || got.Pin != 2 || got.Head != head {
		t.Errorf("after a mint and a blue edit: %+v, want behind pin 2 head %d", got, head)
	}
	b2 := retBoard(t).sit(2, evLens).sit(2, evLens)
	if got := last(t, b2); got.Kind != "unchanged" || got.Pin != 2 || got.Head != 2 {
		t.Errorf("re-dispatched at the same head: %+v", got)
	}
	b3 := retBoard(t).dispatch(2, evLens).ingest()
	h3 := int64(b3.n)
	if got := last(t, b3.sit(h3, evLens)); got.Kind != "first" {
		t.Errorf("an earlier dispatch the lens never sat for is not a prior sitting: %+v", got)
	}
}

// ---- stale areas (§V #1b) ----

func TestPlanListsStaleAreas(t *testing.T) {
	b, pin := retiredForGoodBehindHead(t)
	plan := planOf(t, b)
	if len(plan.StaleAreas) != 1 || plan.StaleAreas[0] != (StaleArea{SeatID: evLens, Pin: pin}) {
		t.Fatalf("stale_areas = %+v, want %s at pin %d", plan.StaleAreas, evLens, pin)
	}
}

func TestPassRefusedUntilSpotCheckCoversStaleAreas(t *testing.T) {
	spot := &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the changes since the pin")}
	// A spot-check in an EARLIER chair sitting does not cover this one.
	b, _ := retiredForGoodBehindHead(t)
	b.add("red-chair", spot).register("red-chair")
	run := b.seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err == nil || !strings.Contains(err.Error(), "retired for good") {
		t.Fatalf("a PASS over an uncovered stale area must be refused: %v", err)
	}
	b2, _ := retiredForGoodBehindHead(t)
	b2.register("red-chair").add("red-chair", spot)
	if _, err := Append(Identity{Run: b2.seed(), SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS after this sitting's spot-check named the stale area was refused: %v", err)
	}
}

func TestSpotCheckAreasValidatedAgainstCast(t *testing.T) {
	run := retBoard(t).seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	for area, ok := range map[string]bool{"red-lens-architecture": false, "red-chair": false, evLens: true} {
		_, err := Append(chair, &recordpb.SpotCheck{Areas: []string{area}, Reason: proto.String("r")})
		if (err == nil) != ok {
			t.Errorf("spot-check --areas %s: err=%v, want accepted=%v", area, err, ok)
		}
	}
}

func TestMigratingSkipsStaleAreaGate(t *testing.T) {
	b, _ := retiredForGoodBehindHead(t)
	run := b.register("red-chair").seed()
	Migrating = true
	defer func() { Migrating = false }()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("an archived PASS over an uncovered stale area must replay: %v", err)
	}
}

// ---- the relay (§V #2) ----

// ARRAYS, NEVER null: the engine refuses a relayed plan whose array is anything else, so the verb
// emits `[]` for every list with nothing in it — an empty plan included.
func TestPlanJSONEmitsArraysNeverNull(t *testing.T) {
	plan := planOf(t, newStage(t).cast("red-chair", "blue-respond", "judge").ingest().register("red-chair"))
	b, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"parties", "docket", "why", "stale_areas"} {
		if !strings.Contains(string(b), `"`+field+`":[]`) {
			t.Errorf("an empty plan marshals %q as something other than []: %s", field, b)
		}
	}
}

// ---- the cast (§V #3) ----

func TestDefaultCastListsAllSevenAreas(t *testing.T) {
	var lenses []string
	for _, s := range CastFor(nil, 1) {
		if strings.HasPrefix(s, "red-lens-") {
			lenses = append(lenses, strings.TrimPrefix(s, "red-lens-"))
		}
	}
	if len(lenses) != 7 || len(lenses) != len(LensAreas) {
		t.Fatalf("the default cast seats %v, want all seven areas %v", lenses, LensAreas)
	}
	have := map[string]bool{}
	for _, a := range lenses {
		have[a] = true
	}
	for _, a := range LensAreas {
		if !have[a] {
			t.Errorf("the default cast lacks the %s lens", a)
		}
	}
	_ = recordtest.P[int] // the package's fixture helpers stay imported for the tests above
}
