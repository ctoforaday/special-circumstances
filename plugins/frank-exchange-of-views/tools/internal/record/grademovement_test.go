package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// movedBoard is a run whose lens sat twice at the head, with G1 and G2 minted by a lens outside
// the cast (so no lens is ready) and graded by moves: each regrades one gap's severity.
func movedBoard(t *testing.T, moves ...[2]string) *stage {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		dispatch(2, evLens).register(evLens).dispatch(2, evLens).register(evLens)
	b.add(outsideLens, cmMint("G1", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "high"))
	b.add(outsideLens, cmMint("G2", recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE, "medium"))
	for _, m := range moves {
		b.regrade(outsideLens, m[0], m[1])
	}
	return b
}

// kindsOf is the blockers' kinds and subjects, as "kind:subject".
func kindsOf(bs []Blocker) []string {
	var out []string
	for _, b := range bs {
		out = append(out, string(b.Kind)+":"+b.Subject)
	}
	return out
}

func planKindsOf(bs []PlanBlocker) []string {
	var out []string
	for _, b := range bs {
		out = append(out, string(b.Kind)+":"+b.Subject)
	}
	return out
}

// THE THRESHOLD IS STRICT AND IN TOTAL. high -> low is 2.0 mass, which the bench does not rule on;
// two gaps moved by 2.0 and 0.5 total 2.5, which it does — each gap below the threshold alone, so a
// change sliced across gaps is held as one.
func TestGradeMovementPastTheThresholdInTotalConvenesTheBench(t *testing.T) {
	t.Run("at the threshold nothing changes", func(t *testing.T) {
		b := movedBoard(t, [2]string{"G1", "low"}, [2]string{"G2", "low_medium"}) // 2.0 + 0.5, G2 below material
		b.add(outsideLens, &recordpb.Close{GapId: proto.String("G2"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("fixed")})
		openChairSitting(b)
		run := b.seed()
		bs, err := PassBlockers(run)
		if err != nil {
			t.Fatal(err)
		}
		if len(bs) != 0 {
			t.Fatalf("2.0 mass on an open gap holds %v, want nothing — the threshold is strict", kindsOf(bs))
		}
		plan, err := PlanDispatch(run)
		if err != nil {
			t.Fatal(err)
		}
		if !plan.PassPermitted || len(plan.ToFile) != 0 {
			t.Fatalf("pass_permitted %v, to file %v, want a PASS permitted and nothing filed", plan.PassPermitted, plan.ToFile)
		}
		if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err != nil {
			t.Fatalf("the gate refuses a PASS at the threshold: %v", err)
		}
	})

	fail := &recordpb.Gate{Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL)}
	b := movedBoard(t, [2]string{"G1", "low"}, [2]string{"G2", "low_medium"}) // 2.0 + 0.5 = 2.5
	openChairSitting(b)
	b.add("red-chair", fail) // the terminal act, so its duty does not hold the work list open
	run := b.seed()

	// The gate list: both gaps, the bench's.
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"grade_movement:G1", "grade_movement:G2"}
	if got := kindsOf(bs); !slices.Equal(got, want) {
		t.Fatalf("blockers = %v, want %v", got, want)
	}
	for _, bl := range bs {
		if bl.Owner != benchSeat {
			t.Errorf("%s is owned by %q, want the bench", bl.Subject, bl.Owner)
		}
	}
	// The gate refuses PASS naming each, and the total; a FAIL is not held.
	passErr := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if passErr == nil {
		t.Fatal("the gate admits a PASS over grade movement past the threshold")
	}
	for _, w := range []string{"G1", "G2", "2.5", "severity high → low", "severity medium → low_medium"} {
		if !strings.Contains(passErr.Error(), w) {
			t.Errorf("the PASS refusal does not say %q: %v", w, passErr)
		}
	}
	if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, fail); err != nil {
		t.Errorf("a FAIL is refused over a bench-owned blocker: %v", err)
	}
	// The plan readies the bench on both, dockets both, and permits no PASS.
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Error("the plan permits a PASS the gate refuses")
	}
	if !slices.Equal(plan.Docket, []string{"G1", "G2"}) || !slices.Equal(plan.ToFile, []string{"G1", "G2"}) {
		t.Errorf("docket %v, to file %v, want G1 and G2 in both", plan.Docket, plan.ToFile)
	}
	if !slices.ContainsFunc(plan.Parties, func(p Party) bool {
		return p.SeatID == benchSeat && slices.Equal(p.GapIDs, []string{"G1", "G2"}) && slices.Equal(p.Occasions, []string{occasionDocket})
	}) {
		t.Errorf("parties = %+v, want the bench convened for the docket on G1 and G2", plan.Parties)
	}
	if got := planKindsOf(plan.Blockers); !slices.Equal(got, want) {
		t.Errorf("plan blockers = %v, want the gate's %v", got, want)
	}
	// The chair's work list blocks on the gate's items.
	s := sittingOfRunT(t, run, "chair", "red-chair")
	var items []string
	for _, bl := range bs {
		items = append(items, bl.WorkItem())
	}
	if got := blockingItems(s); !slices.Equal(got, items) {
		t.Errorf("blocking items = %q, want the gate's %q", got, items)
	}
}

// DISPATCH FILES THE DOCKET, AND THE RULING CLEARS IT. The verb files one docket motion per gap with
// the movement as its basis; the gate list then holds each as an unruled motion, the bench's; the
// bench's ruling rules on the movement, and nothing holds. Each state, the plan, the gate and the
// work list agree.
func TestTheDispatchDocketsGradeMovementAndTheBenchsRulingClearsIt(t *testing.T) {
	b := movedBoard(t, [2]string{"G1", "low"}, [2]string{"G2", "low_medium"})
	b.register("red-chair")
	run := b.seed()
	plan, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"})
	if err != nil {
		t.Fatal(err)
	}
	evs, _, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	motions := map[string]string{}
	for _, e := range evs {
		if m := e.GetMotion(); m != nil && m.GetDocket() != nil {
			motions[m.GetDocket().GetGapId()] = m.GetMotionId()
			if !strings.Contains(m.GetBasis(), "grades moved") || !strings.Contains(m.GetBasis(), "past the 2.0") {
				t.Errorf("the docket motion on %s asks %q, want the grade movement", m.GetDocket().GetGapId(), m.GetBasis())
			}
		}
	}
	if len(motions) != 2 {
		t.Fatalf("docket motions filed on %v, want G1 and G2", motions)
	}
	if got := planKindsOf(plan.Blockers); !slices.Equal(got, []string{"unruled_motion:" + motions["G1"], "unruled_motion:" + motions["G2"]}) {
		t.Errorf("the relayed plan's blockers = %v, want the two docket motions", got)
	}
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	if got := kindsOf(bs); !slices.Equal(got, planKindsOf(plan.Blockers)) {
		t.Errorf("the gate's blockers %v disagree with the plan's %v", got, planKindsOf(plan.Blockers))
	}

	// The bench sits and remands both, so both stay open with their movement: the rulings alone are
	// what rule on it.
	ruled := newStage(t)
	ruled.runDir = run.Dir()
	ruled.n = 1000
	ruled.register("judge").add("judge", remandRuling(motions["G1"], "the leaf supports low")).
		add("judge", remandRuling(motions["G2"], "the leaf supports low_medium"))
	openChairSitting(ruled)
	ruled.seed()
	bs, err = PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(bs) != 0 {
		t.Fatalf("after the bench ruled, the gate holds %v, want nothing", kindsOf(bs))
	}
	after, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !after.PassPermitted || len(after.ToFile) != 0 || len(after.Blockers) != 0 {
		t.Errorf("after the rulings: pass_permitted %v, to file %v, blockers %v — want a PASS permitted, nothing filed", after.PassPermitted, after.ToFile, after.Blockers)
	}
	if err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err != nil {
		t.Errorf("the gate refuses the PASS after the bench ruled: %v", err)
	}
}

// WHAT COUNTS: both directions, the axes the gate reads, open gaps, and only what no docket ruling
// on the gap has followed.
func TestGradeMovementCountsWhatTheGateReads(t *testing.T) {
	regrade := func(gap string, axis recordpb.GradeDimension, word recordpb.Grade) *recordpb.Regrade {
		r := &recordpb.Regrade{GapId: proto.String(gap), Basis: proto.String("b")}
		switch axis {
		case recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD:
			r.Likelihood = word.Enum()
		case recordpb.GradeDimension_GRADE_DIMENSION_IMPACT:
			r.Impact = word.Enum()
		case recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY:
			r.ComplexityCost = word.Enum()
		}
		return r
	}
	cases := []struct {
		name  string
		build func(b *stage)
		want  []string
	}{
		{"inflation counts: likelihood and impact up 1.5 each", func(b *stage) {
			b.add(outsideLens, regrade("G2", recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD, recordpb.Grade_GRADE_CERTAIN))
			b.add(outsideLens, regrade("G2", recordpb.GradeDimension_GRADE_DIMENSION_IMPACT, recordpb.Grade_GRADE_CERTAIN))
		}, []string{"grade_movement:G2", "material_gap:G1", "material_gap:G2"}},
		{"a move down and back up is two movements", func(b *stage) {
			b.regrade(outsideLens, "G1", "low_medium").regrade(outsideLens, "G1", "high") // 1.5 + 1.5
		}, []string{"grade_movement:G1", "material_gap:G1", "material_gap:G2"}},
		{"complexity is read by no gate", func(b *stage) {
			b.add(outsideLens, regrade("G1", recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY, recordpb.Grade_GRADE_CERTAIN))
			b.add(outsideLens, regrade("G2", recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY, recordpb.Grade_GRADE_TRIVIAL))
		}, []string{"material_gap:G1", "material_gap:G2"}},
		{"a closed gap's movement drops out", func(b *stage) {
			b.regrade(outsideLens, "G1", "trivial") // 2.5, then closed
			b.add(outsideLens, &recordpb.Close{GapId: proto.String("G1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("fixed")})
		}, []string{"material_gap:G2"}},
		{"a docket ruling before the movement does not rule on it", func(b *stage) {
			b.docketMotion("red-chair", "M1", "G1").register("judge").add("judge", remandRuling("M1", "d"))
			b.regrade(outsideLens, "G1", "trivial")
		}, []string{"grade_movement:G1", "material_gap:G2"}},
		{"a docket motion standing on the gap rules on it: no second filing", func(b *stage) {
			b.docketMotion("red-chair", "M1", "G1")
			b.regrade(outsideLens, "G1", "trivial")
		}, []string{"material_gap:G2", "unruled_motion:M1"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b := movedBoard(t)
			c.build(b)
			openChairSitting(b)
			run := b.seed()
			bs, err := PassBlockers(run)
			if err != nil {
				t.Fatal(err)
			}
			got := kindsOf(bs)
			slices.Sort(got)
			if !slices.Equal(got, c.want) {
				t.Errorf("blockers = %v, want %v", got, c.want)
			}
			plan, err := PlanDispatch(run)
			if err != nil {
				t.Fatal(err)
			}
			pk := planKindsOf(plan.Blockers)
			slices.Sort(pk)
			if !slices.Equal(pk, got) {
				t.Errorf("plan blockers %v disagree with the gate's %v", pk, got)
			}
		})
	}
}

// THE CHAIR'S ROW SAYS WHAT THE DISPATCH DOES WITH A REMANDED GAP WHOSE GRADES MOVED SINCE: it is
// docketed again for its movement, by the plan's own route.
func TestTheChairsRowForARemandedGapWhoseGradesMoved(t *testing.T) {
	b := movedBoard(t)
	b.docketMotion("red-chair", "M1", "G1").register("judge").add("judge", remandRuling("M1", "d"))
	b.add(outsideLens, &recordpb.Regrade{GapId: proto.String("G1"), Likelihood: recordpb.Grade_GRADE_CERTAIN.Enum(),
		Impact: recordpb.Grade_GRADE_HIGH.Enum(), Basis: proto.String("b")}) // 1.5 + 1.0, still material
	openChairSitting(b)
	run := b.seed()
	s := sittingOfRunT(t, run, "chair", "red-chair")
	found := false
	for _, it := range s.Open {
		found = found || strings.Contains(it.What, "gap G1 is open because the BENCH REMANDED it, and its grades have moved since")
	}
	if !found {
		t.Errorf("the chair's list has no row docketing remanded G1 for its movement: %q", whats(s.Open))
	}
}
