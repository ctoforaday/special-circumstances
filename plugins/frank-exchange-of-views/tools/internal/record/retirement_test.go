package record

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// sit is one sitting of a lens: a dispatch naming it at pin, its register, and the whole report served
// to it at pin.
func (b *stage) sit(pin int64, lens string) *stage {
	return b.sitUnserved(pin, lens).served(lens, pin)
}

// sitUnserved is one sitting of a lens that was never served the whole report: a dispatch at pin and
// its register.
func (b *stage) sitUnserved(pin int64, lens string) *stage {
	return b.dispatch(pin, lens).register(lens)
}

// head is the stage's report head: the id of its latest ingest or edit — its count, as seeded.
func (b *stage) head() int64 {
	for i := len(b.evs) - 1; i >= 0; i-- {
		switch b.evs[i].GetType() {
		case recordpb.EventType_EVENT_TYPE_BLUE_EDIT, recordpb.EventType_EVENT_TYPE_BASE_INGEST:
			return int64(b.n - (len(b.evs) - 1 - i))
		}
	}
	return 0
}

// served is the whole report served to lens at head, in its current sitting.
func (b *stage) served(lens string, head int64) *stage {
	return b.add(lens, &recordpb.ReportRead{Head: proto.Int64(head)})
}

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

// THE RULE: a lens with no gap engaged is readied when it has never sat or the report head has
// moved since its last sitting, and not otherwise.
func TestALensIsReadiedOnlyWhenTheHeadHasMovedSinceItSat(t *testing.T) {
	never := planOf(t, retBoard(t))
	if !lensReady(never) || !strings.Contains(whyOf(never), "never sat") {
		t.Fatalf("a lens that has never sat is ready for its first look: %+v", never)
	}
	same := planOf(t, retBoard(t).sit(2, evLens))
	if lensReady(same) || !strings.Contains(whyOf(same), "served the report at the head (pin 2)") {
		t.Fatalf("a lens that sat at the head is not readied for a second look at the same report: %+v", same)
	}
	// A SITTING AT THE HEAD THAT WAS NEVER SERVED THE REPORT IS NOT A LOOK AT IT (#1458): the lens is
	// still behind the head, and readied to audit it.
	unserved := planOf(t, retBoard(t).sitUnserved(2, evLens))
	if !lensReady(unserved) || !strings.Contains(whyOf(unserved), "no sitting it sat through was served the whole report, behind head 2") {
		t.Fatalf("a lens that sat at the head and was never served the report is ready to audit it: %+v", unserved)
	}
	moved := planOf(t, retBoard(t).sit(2, evLens).ingest())
	if !lensReady(moved) || !strings.Contains(whyOf(moved), "the whole report was last served to it at its pin 2, behind head") {
		t.Fatalf("a lens is ready when the head has moved past its last sitting's pin: %+v", moved)
	}
}

// A lens that minted at the head is not readied by its mint: the mint does not ready a second audit
// of the report it has just read, and its gap waits for blue's answer (lensTurnOn) — blue is engaged
// on it, and the lens is not.
func TestALensThatMintedAtTheHeadWaitsForBlue(t *testing.T) {
	plan := planOf(t, retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")))
	if lensReady(plan) || !slices.Equal(gapsOf(plan, "blue-respond"), []string{"G1"}) || !strings.Contains(whyOf(plan), "served the report at the head (pin 2)") {
		t.Fatalf("a lens whose open material gap blue has not answered does not sit, and the head has not moved: %+v", plan)
	}
}

// A dispatch the lens never sat for is not a sitting: the head has still moved since it last sat.
func TestAnUnsatDispatchLeavesTheLensReady(t *testing.T) {
	b := retBoard(t).sit(2, evLens).ingest()
	head := int64(b.n)
	plan := planOf(t, b.dispatch(head, evLens))
	if !lensReady(plan) {
		t.Fatalf("a lens dispatched at the new head that has opened no sitting has not sat there: %+v", plan)
	}
}

func TestLensRetiresAtItsThirdBarrenSitting(t *testing.T) {
	// two barren sittings, each at its own head, and the head moved past the second
	twice := func() (*stage, int64) {
		b := retBoard(t).sit(2, evLens).ingest()
		b.sit(int64(b.n), evLens).ingest()
		return b, int64(b.n)
	}
	two, h3 := twice()
	if plan := planOf(t, two); !lensReady(plan) || !strings.Contains(whyOf(plan), "2 sitting(s) in a row with no fresh material mint; it retires at 3") {
		t.Fatalf("after two barren sittings a head move readies the lens: %+v", plan)
	}
	three, _ := twice()
	plan := planOf(t, three.sit(h3, evLens).ingest())
	if lensReady(plan) || !strings.Contains(whyOf(plan), evLens+": retired (pin ") {
		t.Fatalf("a third barren sitting retires the lens, and a head move does not ready it: %+v", plan)
	}
}

func TestRetiredLensIgnoresLaterHeadMoves(t *testing.T) {
	plan := planOf(t, retBoard(t).sit(2, evLens).sit(2, evLens).sit(2, evLens).ingest().ingest())
	if lensReady(plan) || !strings.Contains(whyOf(plan), evLens+": retired (pin 2)") {
		t.Fatalf("head moves do not ready a retired lens: %+v", plan)
	}
}

func TestAFreshMintReturnsARetiredLensToTheCount(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).sit(2, evLens)
	b.sit(2, evLens).add(evLens, freshHigh("G1")).ingest()
	if f := lensStateOf(t, b.seed()); f.retired() || f.barren != 0 || !f.ready {
		t.Fatalf("a productive sitting returns a retired lens to the count at zero, ready on the moved head: %+v", f)
	}
}

func TestNoLensReadyBeforeIngest(t *testing.T) {
	plan := planOf(t, newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").register("red-chair"))
	if lensReady(plan) || plan.PassPermitted || !strings.Contains(whyOf(plan), "no report has been ingested") {
		t.Fatalf("nothing is ready and nothing permits a PASS before a report is ingested: %+v", plan)
	}
}

func TestLineageMintIsNotFresh(t *testing.T) {
	// Sitting 1 mints fresh material; the three after it mint only successors of it — lineage, not
	// discovery — so they are barren and the lens retires.
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).
		sit(2, evLens).add(evLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high", "G1")).
		sit(2, evLens).add(evLens, cmMint("G3", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high", "G2")).
		sit(2, evLens).add(evLens, cmMint("G4", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high", "G3"))
	if f := lensStateOf(t, b.seed()); f.barren != retiresAfter {
		t.Fatalf("sittings that minted only lineage are barren: %+v", f)
	}
}

func TestNeverClassMintIsNotFresh(t *testing.T) {
	b := retBoard(t).sit(2, evLens).add(evLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")).
		sit(2, evLens).add(evLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high")).
		sit(2, evLens).add(evLens, cmMint("G3", recordpb.ClassMaterial_CLASS_MATERIAL_NEVER, "high"))
	if f := lensStateOf(t, b.seed()); f.barren != retiresAfter {
		t.Fatalf("sittings that minted only never-class gaps are barren: %+v", f)
	}
}

// retiredBehindHead is a lens retired by three barren sittings that the head has moved past
// since: stale. The pin is its last sitting's.
func retiredBehindHead(t *testing.T) (*stage, int64) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).ingest()
	pin := int64(b.n)
	b.sit(pin, evLens).ingest()
	return b, pin
}

func TestPassPermittedWithRetiredLensBehindHead(t *testing.T) {
	b, _ := retiredBehindHead(t)
	plan := planOf(t, b.register("red-chair"))
	if !plan.PassPermitted || lensReady(plan) {
		t.Fatalf("a retired lens behind the head is not ready, so the plan permits a PASS: %+v", plan)
	}
}

func TestPassRefusedWhileALensHasNotSeenTheHead(t *testing.T) {
	b := retBoard(t).sit(2, evLens).ingest().register("red-chair")
	run := b.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Fatalf("pass_permitted while the head has moved past a lens's last sitting: %+v", plan)
	}
	_, err = Append(Identity{Run: run, SeatID: "red-chair"}, passGate)
	if err == nil || !strings.Contains(err.Error(), evLens) || !strings.Contains(err.Error(), "the whole report was last served to it at its pin") {
		t.Fatalf("a chair's PASS over a lens the head has moved past must be refused at the write path, naming the lens: %v", err)
	}
}

// THE PLAN AND THE GATE READ ONE FOLD: in every state of it, pass_permitted and the write path's
// answer to a PASS agree.
func TestPassPermittedAndGateShareTheFold(t *testing.T) {
	const staleRead = "retired, stale area read"
	for _, c := range []struct {
		name  string
		build func(t *testing.T) *stage
	}{
		{"never sat", func(t *testing.T) *stage { return retBoard(t) }},
		{"sat at the head", func(t *testing.T) *stage { return retBoard(t).sit(2, evLens) }},
		{"head moved since it sat", func(t *testing.T) *stage { return retBoard(t).sit(2, evLens).ingest() }},
		{staleRead, func(t *testing.T) *stage {
			b, _ := retiredBehindHead(t)
			b.register("red-chair")
			return b.add("red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the changes since the pin")})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := c.build(t)
			if c.name != staleRead {
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

// A lens that sat at the head and found nothing holds no PASS: the plan permits one and the gate
// takes it, with no second sitting at the same head between.
func TestPassIsPermittedOnceEveryLensHasSatAtTheHead(t *testing.T) {
	run := retBoard(t).sit(2, evLens).register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.PassPermitted || len(plan.Parties) != 0 {
		t.Fatalf("every lens has sat at the head and nothing is open, so nobody sits and a PASS is permitted: %+v", plan)
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("the PASS the plan permits was refused: %v", err)
	}
}

// A LENS AT ITS BUDGET MINTS NOTHING, so its sittings are barren, and holding the gaps that spent the
// budget is not a debt: with nothing new on them it retires. Its material gap still engages it as a
// party to that gap's dispute once blue has acted on it, which is what a held gap needs of it.
func TestBudgetExhaustedLensRetiresAndIsStillAPartyToItsGap(t *testing.T) {
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).sit(2, evLens).sit(2, evLens).sit(2, evLens)
	writeRunConfig(t, b.runDir, `{"mintBudget":1}`)
	run := b.seed()
	if _, err := Append(Identity{Run: run, SeatID: evLens}, liveMint("G2", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_HIGH)); err == nil || !strings.Contains(err.Error(), "budget") {
		t.Fatalf("fixture: the lens is not at its budget: %v", err)
	}
	if f := lensStateOf(t, run); f.barren != retiresAfter || !f.retired() || len(f.owes) != 0 {
		t.Fatalf("a lens at its budget sits barren, owes nothing for holding G1, and retires: %+v", f)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(gapsOf(plan, evLens)) != 0 {
		t.Errorf("the retired lens is a party to G1 before blue has acted on it: %+v", plan.Parties)
	}
	answered, err := PlanDispatch(b.more().blueComputes("G1").seed())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(gapsOf(answered, evLens), []string{"G1"}) {
		t.Errorf("the retired lens is not a party to its material G1 once blue has acted on it: %+v", answered.Parties)
	}
}

func TestWhyCarriesState(t *testing.T) {
	b, pin := retiredBehindHead(t)
	for _, c := range []struct {
		name string
		plan Plan
		want string
	}{
		{"never sat", planOf(t, retBoard(t)), "never sat — audits the report at head 2"},
		{"served at the head", planOf(t, retBoard(t).sit(2, evLens)), "served the report at the head (pin 2) — ready when the head moves"},
		{"head moved", planOf(t, retBoard(t).sit(2, evLens).ingest()), "the whole report was last served to it at its pin 2, behind head 7 — audits the report (1 sitting(s) in a row with no fresh material mint; it retires at 3)"},
		{"retired", planOf(t, b), fmt.Sprintf("retired (pin %d) — its last 3 sittings minted no fresh material; not ready", pin)},
	} {
		if !strings.Contains(whyOf(c.plan), evLens+": ") || !strings.Contains(whyOf(c.plan), c.want) {
			t.Errorf("%s: the why must say %q:\n%s", c.name, c.want, whyOf(c.plan))
		}
	}
}

// THE LENS READS ITS LAST SITTING FROM ITS WORK VIEW, after registering for the dispatch it is
// sitting for — which is why the fact cannot be the latest dispatch's own sitting (sittingFor),
// which is the current one once the register lands.
func TestLensWorkCarriesLastSitting(t *testing.T) {
	last := func(t *testing.T, b *stage) LastReadJSON {
		t.Helper()
		run := b.seed()
		out, err := WorkJSONBytes(run, "lens", evLens)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(out), `"last_read"`) {
			t.Fatalf("the lens's work view carries no last_read key:\n%s", out)
		}
		var w struct {
			Sitting struct {
				LastRead LastReadJSON `json:"last_read"`
			} `json:"sitting"`
		}
		if err := json.Unmarshal(out, &w); err != nil {
			t.Fatal(err)
		}
		return w.Sitting.LastRead
	}
	if got := last(t, retBoard(t)); got.Kind != "undispatched" {
		t.Errorf("never dispatched: %+v", got)
	}
	if got := last(t, retBoard(t).sitUnserved(2, evLens)); got.Kind != "first" || got.Head != 2 {
		t.Errorf("first sitting, registered: %+v", got)
	}
	edited := func() (*stage, int64) {
		b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).
			add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b"), EditKey: proto.String("E1")})
		head := int64(b.n)
		return b.sitUnserved(head, evLens), head
	}
	behind, head := edited()
	if got := last(t, behind); got.Kind != "behind" || got.Pin != 2 || got.Head != head {
		t.Errorf("after a mint and a blue edit: %+v, want behind pin 2 head %d", got, head)
	}
	// THE FIELD IS AS OF NOW: the whole read this sitting asked for moves it to served at the head
	// served, on the same sitting (#1601 smoke: dark-side read it whole and was still told behind).
	behind, head = edited()
	if got := last(t, behind.served(evLens, head)); got.Kind != "served" || got.Pin != head || got.Head != head {
		t.Errorf("behind, then served whole in this sitting: %+v, want served pin %d head %d", got, head, head)
	}
	if got := last(t, retBoard(t).sit(2, evLens).sitUnserved(2, evLens)); got.Kind != "unchanged" || got.Pin != 2 || got.Head != 2 {
		t.Errorf("re-dispatched at the same head: %+v", got)
	}
	if got := last(t, retBoard(t).sit(2, evLens)); got.Kind != "served" || got.Pin != 2 || got.Head != 2 {
		t.Errorf("first sitting, served whole: %+v, want served pin 2 head 2", got)
	}
	b3 := retBoard(t).dispatch(2, evLens).ingest()
	h3 := int64(b3.n)
	if got := last(t, b3.sitUnserved(h3, evLens)); got.Kind != "first" {
		t.Errorf("an earlier dispatch the lens never sat for is not a prior sitting: %+v", got)
	}
}

// THE WORK LIST'S ITEM SAYS WHAT THE FIELD SAYS: behind before the read, served after it, and the
// behind item is gone once this sitting served the report whole.
func TestLensWorkItemFollowsTheServe(t *testing.T) {
	behind := func() (*stage, int64) {
		b := retBoard(t).sit(2, evLens).ingest()
		head := int64(b.n)
		return b.sitUnserved(head, evLens), head
	}
	items := func(st *stage) string {
		t.Helper()
		var whats []string
		for _, it := range sittingOfRunT(t, st.seed(), "lens", evLens).Open {
			whats = append(whats, it.What)
		}
		return strings.Join(whats, "\n")
	}
	b, _ := behind()
	if got := items(b); !strings.Contains(got, "before this sitting, the report was last served to you whole at 2") {
		t.Errorf("behind: the work list does not say so:\n%s", got)
	}
	b, head := behind()
	got := items(b.served(evLens, head))
	if strings.Contains(got, "before this sitting") || !strings.Contains(got, fmt.Sprintf("this sitting served you the report whole at head %d", head)) {
		t.Errorf("served in this sitting: the work list still speaks of before it, or not of the serve:\n%s", got)
	}
}

// ---- stale areas (§V #1b) ----

func TestPlanListsStaleAreas(t *testing.T) {
	b, pin := retiredBehindHead(t)
	plan := planOf(t, b)
	if len(plan.StaleAreas) != 1 || plan.StaleAreas[0] != (StaleArea{SeatID: evLens, Pin: pin}) {
		t.Fatalf("stale_areas = %+v, want %s at pin %d", plan.StaleAreas, evLens, pin)
	}
}

func TestPassRefusedUntilSpotCheckCoversStaleAreas(t *testing.T) {
	spot := &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the changes since the pin")}
	// A spot-check in an EARLIER chair sitting does not cover this one.
	b, _ := retiredBehindHead(t)
	b.add("red-chair", spot).register("red-chair")
	run := b.seed()
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, passGate); err == nil || !strings.Contains(err.Error(), "the report changed behind 1 retired lens(es)") {
		t.Fatalf("a PASS over an uncovered stale area must be refused: %v", err)
	}
	b2, _ := retiredBehindHead(t)
	b2.register("red-chair").add("red-chair", spot)
	if _, err := Append(Identity{Run: b2.seed(), SeatID: "red-chair"}, passGate); err != nil {
		t.Fatalf("a PASS after this sitting's spot-check named the stale area was refused: %v", err)
	}
}

func TestSpotCheckAreasValidatedAgainstCast(t *testing.T) {
	run := retBoard(t).seed()
	chair := Identity{Run: run, SeatID: "red-chair"}
	// The refused areas come first and are refused FOR THE AREA: a spot-check is once per sitting,
	// so one that lands makes every later one a refusal whatever it names.
	for _, area := range []string{"red-lens-architecture", "red-chair", "blue-respond", "red-lens-banana"} {
		_, err := Append(chair, &recordpb.SpotCheck{Areas: []string{area}, Reason: proto.String("r")})
		if err == nil || !strings.Contains(err.Error(), "not a lens seat in this run's cast") {
			t.Errorf("spot-check --areas %s is not refused as no lens seat of the cast: %v", area, err)
		}
	}
	if _, err := Append(chair, &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("r")}); err != nil {
		t.Errorf("spot-check --areas %s, a cast lens, is refused: %v", evLens, err)
	}
}

func TestMigratingSkipsStaleAreaGate(t *testing.T) {
	b, _ := retiredBehindHead(t)
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
	for _, s := range func() []string { s, _ := CastFor(1); return s }() {
		if strings.HasPrefix(s, "red-lens-") {
			lenses = append(lenses, strings.TrimPrefix(s, "red-lens-"))
		}
	}
	if len(lenses) != 7 || len(lenses) != len(LensAreas) {
		t.Fatalf("the cast seats %v, want all seven areas %v", lenses, LensAreas)
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

// A LENS'S AUDIT IS ON ITS WORK LIST, AND FOR A LONG TIME IT WAS THE ONE THING MISSING FROM IT.
//
// A lens is dispatched to find what is wrong with the report. That is not a gap, so it was in no
// item, so `open` came back EMPTY and `complete` came back TRUE for a lens whose report had just
// moved. Rendered on the 2026-09-23 run, the voice lens read `complete: true, open: []` with
// `last_sitting.kind: behind` — and went to the report anyway, where it found two real defects.
//
// A seat that cannot trust the work list reads the rest of the surface to find the work: 531 `show`
// calls across two runs, 331 of them a lens. This is what makes the list answer the question the
// seat actually has.
func TestALensAuditIsAnItemOnItsWorkList(t *testing.T) {
	// first sitting: nothing of the report is yet its own to have read
	first := sittingOfRunT(t, retBoard(t).sitUnserved(2, evLens).seed(), "lens", evLens)
	if !listsItem(first, "audit the report in full") {
		t.Errorf("a lens on its FIRST sitting is not told to audit: %+v", first.Open)
	}

	// behind: the FULL re-read is the surface and the item names the range it moved through.
	//
	// This asserted "audit what moved" and read, in its own comment, "the change is the surface" —
	// which is what the item used to say and what adversarial-audit forbids in terms: "a
	// change-summary is a navigation hint, never the audit surface… This is not permission to audit a
	// fragment." The item now states the pins as navigation and sends the lens to the whole document,
	// so the assertion moves with it.
	b := retBoard(t).sit(2, evLens).add(evLens, freshHigh("G1")).
		add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G1"), Old: proto.String("a"), New: proto.String("b"), EditKey: proto.String("E1")})
	head := int64(b.n)
	beh := sittingOfRunT(t, b.sitUnserved(head, evLens).seed(), "lens", evLens)
	if !listsItem(beh, "reads it in full") {
		t.Errorf("a lens whose report MOVED is not sent to the whole document: %+v", beh.Open)
	}
	if listsItem(beh, "not the whole document") {
		t.Errorf("the item tells the lens to audit a fragment, which adversarial-audit forbids: %+v", beh.Open)
	}
	if !listsItem(beh, "dispatched against head") {
		t.Errorf("the item does not name the range the lens must audit: %+v", beh.Open)
	}

	// NON-BLOCKING, because finding nothing is a legitimate audit. The item must appear without
	// holding the turn open the way an owed register does.
	for _, it := range beh.Open {
		if strings.Contains(it.What, "audit what moved") && it.Blocks {
			t.Error("the audit item BLOCKS — finding nothing is a legitimate audit and must not hold the turn")
		}
	}

	// unchanged: nothing moved, so there is nothing to audit and the list says so by silence.
	un := sittingOfRunT(t, retBoard(t).sit(2, evLens).sitUnserved(2, evLens).seed(), "lens", evLens)
	if listsItem(un, "audit") {
		t.Errorf("a lens dispatched against an UNMOVED report is told to audit it: %+v", un.Open)
	}
}

// AN EMPTY WORK LIST SAYS IT IS EMPTY, because `complete: true, open: []` is indistinguishable
// from a command that found nothing — and a seat that reads it as broken goes looking for the work
// in the rest of the surface. Measured across two runs: 531 `show` calls, 331 of them a lens,
// several against a list that had already said they were complete.
func TestAnEmptyWorkListSaysSoRatherThanReturningNothing(t *testing.T) {
	// a lens re-dispatched at the same head: nothing moved, nothing owed
	un := sittingOfRunT(t, retBoard(t).sit(2, evLens).sitUnserved(2, evLens).seed(), "lens", evLens)
	if len(un.Open) == 0 {
		t.Fatal("a seat with nothing to do got an EMPTY list, which reads as a broken command")
	}
	if !listsItem(un, "this list is complete, not empty") {
		t.Errorf("the list does not state that it is complete rather than empty: %+v", un.Open)
	}
	if !listsItem(un, "no other projection holds work for you") {
		t.Errorf("the list does not close off the hunt for work elsewhere: %+v", un.Open)
	}
	// AND IT STILL LETS THE SEAT STOP. Saying so must not become an obligation.
	if !un.Complete {
		t.Error("stating that the list is empty made the sitting incomplete — having nothing to do is not work")
	}
}

// THE SENTENCE IS NOT SAID TO A SEAT DISPATCHED TO WRITE. The frontier, a lane and the synthesizer
// hold no item for the inquiries, the draft or the report they were dispatched to write, so their
// list is empty before they have begun; "no other projection holds work for you … ending your turn
// now is the correct act" there denies the dispatch. universe-m21's synthesizer read it, looked at
// nothing else on the record, and proposed five approaches the record already held. Blue's
// responding seat is dispatched onto gaps, which are items, and keeps the sentence.
func TestTheEmptyListSentenceIsNotSaidToASeatDispatchedToWrite(t *testing.T) {
	seats := []string{"frontier", "blue-lane-1", "blue-synthesize", blueRespondSeat}
	b := newStage(t).cast(append([]string{evLens, "red-chair", "judge"}, seats...)...).ingest()
	for _, seat := range seats {
		b.register(seat) // a sitting that has opened and recorded nothing: the list as dispatch delivers it
	}
	run := b.seed()
	for _, seat := range seats {
		s := sittingOfRunT(t, run, "blue", seat)
		told := listsItem(s, "this list is complete, not empty")
		if want := seat == blueRespondSeat; told != want {
			t.Errorf("%s: told its list is complete and its turn should end = %v, want %v: %+v", seat, told, want, s.Open)
		}
		if seat != blueRespondSeat && len(s.Open) != 0 {
			t.Errorf("%s holds items on a record with nothing for it: %+v", seat, s.Open)
		}
		if !s.Complete {
			t.Errorf("%s: a list with nothing owed is not complete", seat)
		}
	}
}
