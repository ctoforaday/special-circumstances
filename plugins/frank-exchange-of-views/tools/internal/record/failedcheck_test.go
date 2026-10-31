package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A STANDING UNSOUND JUDGEMENT IS A PASS BLOCKER (gblock, 2026-10-24): the judging lens is readied to
// mint its gap, blue to fix or retire the computation, and it clears when the computation is judged sound again or
// retired. universe-m32 ended UNVERIFIED with pass_permitted true over exactly this board.

// unsoundBoard is a run whose computation lens re-ran K-1 and judged its sentence unsound in its
// sitting, with the chair seated after it.
func unsoundBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		computation("K-1", "s1").
		dispatch(2, compLens).register(compLens).
		reproduce(compLens, "K-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
}

// aboutComputation is a gap about computation anchor, graded low: not material, so it readies nobody itself.
func aboutComputation(gap, anchor string) *recordpb.Mint {
	m := budgetMint(gap)
	m.Severity = recordpb.Grade_GRADE_LOW.Enum()
	m.Problem = proto.String("the script does trial division, and the sentence calls it a sieve")
	m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_ANCHOR.Enum(), proto.String(anchor)
	return m
}

func unsoundOf(t *testing.T, run Run) []Blocker {
	t.Helper()
	all, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []Blocker
	for _, b := range all {
		if b.Kind == BlockerFailedCheck {
			out = append(out, b)
		}
	}
	return out
}

func TestAStandingUnsoundJudgementHoldsThePass(t *testing.T) {
	run := unsoundBoard(t).seed()
	got := unsoundOf(t, run)
	if len(got) != 1 || got[0].Subject != "K-1" || got[0].Owner != compLens {
		t.Fatalf("unsound blockers = %+v, want K-1 owned by its judge %s", got, compLens)
	}
	err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil {
		t.Fatal("the gate admits a PASS over a computation the record calls unsound")
	}
	for _, want := range []string{"K-1", compLens, "judged unsound"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the PASS refusal does not name %q: %v", want, err)
		}
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over a standing unsound judgement (why %q)", plan.Why)
	}
}

func TestAnUnsoundJudgementClears(t *testing.T) {
	for _, c := range []struct {
		name  string
		then  func(b *stage)
		clear bool
	}{
		{"its judge re-judges it sound", func(b *stage) {
			b.reproduce(compLens, "K-1", "s1", notRun, recordpb.Soundness_SOUNDNESS_SOUND)
		}, true},
		{"blue retires the computation", func(b *stage) {
			b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the sentence"), Reason: proto.String("refuted"), Anchors: []string{"K-1"}})
		}, true},
		{"a sound judgement after blue's edit", func(b *stage) {
			b.reopen("K-1").dispatch(b.at(), evLens).register(evLens).
				reproduce(evLens, "K-1", "s1", notRun, recordpb.Soundness_SOUNDNESS_SOUND)
		}, true},
		{"another lens's sound beside it", func(b *stage) {
			b.dispatch(2, evLens).register(evLens).
				reproduce(evLens, "K-1", "s1", notRun, recordpb.Soundness_SOUNDNESS_SOUND)
		}, false},
		{"the judge's gap alone", func(b *stage) { b.add(compLens, aboutComputation("G-1", "K-1")) }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			b := unsoundBoard(t)
			c.then(b)
			if got := unsoundOf(t, b.seed()); (len(got) == 0) != c.clear {
				t.Errorf("unsound blockers = %+v, want cleared = %v", got, c.clear)
			}
		})
	}
}

// THE JUDGE IS READIED TO MINT ITS GAP, AND BLUE TO FIX OR RETIRE THE COMPUTATION — each once per cause.
func TestAnUnsoundJudgementReadiesItsJudgeAndBlue(t *testing.T) {
	parties := func(p Plan) map[string]bool {
		out := map[string]bool{}
		for _, x := range p.Parties {
			out[x.SeatID] = true
		}
		return out
	}
	says := func(p Plan, sub string) bool {
		for _, w := range p.Why {
			if strings.Contains(w, sub) {
				return true
			}
		}
		return false
	}

	run := unsoundBoard(t).seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !parties(plan)[compLens] || !says(plan, compLens+"'s check of computation K-1 judged it unsound and no gap of its carries that to blue — ready") {
		t.Errorf("the judge is not readied to mint its gap: parties %v, why %q", plan.Parties, plan.Why)
	}
	if !parties(plan)[blueRespondSeat] || !says(plan, "computation K-1 stands on a failed check by "+compLens+" — ready to remedy it: blue fixes the computation or retires it") {
		t.Errorf("blue is not readied to fix or retire the computation: parties %v, why %q", plan.Parties, plan.Why)
	}
	lens := sittingOfRunT(t, run, "lens", compLens)
	if it, ok := itemNaming(lens, "computation K-1"); !ok || !it.Blocks {
		t.Errorf("the judge's work list does not owe its gap about K-1: %+v", lens.Open)
	}
	blue := sittingOfRunT(t, run, "blue", blueRespondSeat)
	if it, ok := itemNaming(blue, "computation K-1"); !ok || !it.Blocks || !strings.Contains(it.What, "retire") {
		t.Errorf("blue's work list does not owe the fix or retire of K-1: %+v", blue.Open)
	}

	// The judge mints its gap: it owes nothing more on it, and blue's cause moves to the mint.
	minted := func() *stage { return unsoundBoard(t).add(compLens, aboutComputation("G-1", "K-1")) }
	run = minted().seed()
	plan, err = PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if says(plan, "no gap of its carries that to blue") {
		t.Errorf("the judge is still readied for a gap it minted: why %q", plan.Why)
	}
	if !parties(plan)[blueRespondSeat] {
		t.Errorf("blue is not readied once the gap is minted: parties %v", plan.Parties)
	}
	lens = sittingOfRunT(t, run, "lens", compLens)
	if it, ok := itemNaming(lens, "computation K-1"); !ok || it.Blocks {
		t.Errorf("the judge's item on K-1 still blocks once its gap is minted: %+v", lens.Open)
	}

	// Blue sits after the mint and leaves it: not readied again for it.
	b := minted()
	b.dispatch(b.at(), blueRespondSeat).register(blueRespondSeat).register(compLens)
	plan, err = PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if !says(plan, blueRespondSeat+" has sat since computation K-1's check failed") {
		t.Errorf("blue is readied again for a cause it sat through: why %q", plan.Why)
	}
}

// A GAP THAT CARRIES THE LENS'S OWN UNSOUND JUDGEMENT TAKES NO MINT. Every other mint counts.
func TestTheGapAboutAnUnsoundJudgementTakesNoMint(t *testing.T) {
	b := unsoundBoard(t)
	writeRunConfig(t, b.runDir, `{"mintBudget":1}`)
	b.computation("K-2", "s2")
	run := b.seed()
	md := "# Report\n\nA paragraph that cites nothing."
	withRenderer(t, renders(&md))
	id := Identity{Run: run, SeatID: compLens}
	// The computation lens's budget is max(1, ceil(2 computations × 3 / 2)) = 3.
	mustAppend(t, id, budgetMint("G-00000001"))
	mustAppend(t, id, budgetMint("G-00000002"))
	mustAppend(t, id, budgetMint("G-00000003"))
	_, err := Append(id, aboutComputation("G-00000004", "K-2"))
	mustRefuse(t, err, "admits no more")
	mustAppend(t, id, aboutComputation("G-00000005", "K-1"))
	w, err := WorkOfSeat(run, "lens", compLens)
	if err != nil {
		t.Fatal(err)
	}
	if mb := w.Sitting.MintBudget; mb == nil || mb.Open != 3 || mb.Answering != 1 {
		t.Errorf("mint_budget = %+v, want open 3, answering 1", mb)
	}
	_, err = Append(id, aboutComputation("G-00000006", "K-1"))
	mustRefuse(t, err, "admits no more")
}

// A GAP OR FINDING ABOUT A COMPUTATION NAMES ONE: the write refuses an anchor no computation carries.
func TestAnActAboutAComputationNamesOne(t *testing.T) {
	run := unsoundBoard(t).seed()
	_, err := Append(Identity{Run: run, SeatID: compLens}, aboutComputation("G-1", "K-9"))
	mustRefuse(t, err, "K-9", "no citation or computation on this record carries")
}

// A GAP MINTED IN THE JUDGEMENT'S OWN SITTING ANSWERS IT, before or after the reproduce; one from an
// earlier sitting carried a judgement of a sentence since replaced, and answers nothing now.
func TestTheJudgementsOwnSittingsGapAnswersIt(t *testing.T) {
	b := newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		computation("K-1", "s1").dispatch(2, compLens).register(compLens).
		add(compLens, aboutComputation("G-1", "K-1")).
		reproduce(compLens, "K-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
	if got := unsoundOf(t, b.seed()); len(got) != 1 || got[0].AnsweredBy() != "G-1" {
		t.Errorf("unsound blockers = %+v, want K-1 answered by G-1, minted in the judgement's sitting", got)
	}
	b = newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		computation("K-1", "s1").dispatch(2, compLens).register(compLens).
		add(compLens, aboutComputation("G-1", "K-1")).
		dispatch(2, compLens).register(compLens).
		reproduce(compLens, "K-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
	if got := unsoundOf(t, b.seed()); len(got) != 1 || got[0].AnsweredBy() != "" {
		t.Errorf("unsound blockers = %+v, want K-1 answered by no gap: G-1 was minted in an earlier sitting", got)
	}
}

// A REPEAT UNSOUND JUDGEMENT IS REFUSED while the judge's earlier `unsound` of the sentence stands —
// the sentence unchanged since — and the gap it minted about the computation is open. The refusal names the
// gap. A sound judgement, another lens's, a judgement after blue's edit, and one whose gap closed are
// each admitted.
func TestARepeatUnsoundJudgementIsRefused(t *testing.T) {
	again := func(s recordpb.Soundness) *recordpb.Reproduce {
		return &recordpb.Reproduce{ComputationId: proto.String("K-1"), BundleSha: proto.String("s1"), Rerun: recordpb.Rerun_RERUN_REPRODUCED.Enum(), Ran: proto.Bool(false),
			Soundness: s.Enum(), Note: proto.String("read again")}
	}
	answered := func() *stage { return unsoundBoard(t).add(compLens, aboutComputation("G-1", "K-1")) }
	for _, c := range []struct {
		name   string
		board  func() *stage
		seat   string
		as     recordpb.Soundness
		refuse bool
	}{
		{"in the judgement's own sitting", answered, compLens, recordpb.Soundness_SOUNDNESS_UNSOUND, true},
		{"in a later sitting", func() *stage {
			b := answered()
			return b.dispatch(b.at(), compLens).register(compLens)
		}, compLens, recordpb.Soundness_SOUNDNESS_UNSOUND, true},
		{"judged sound instead", answered, compLens, recordpb.Soundness_SOUNDNESS_SOUND, false},
		{"before its gap is minted", func() *stage { return unsoundBoard(t) }, compLens, recordpb.Soundness_SOUNDNESS_UNSOUND, false},
		{"its gap closed", func() *stage { return answered().closeGap(compLens, "G-1") }, compLens, recordpb.Soundness_SOUNDNESS_UNSOUND, false},
		{"after blue's edit changed the sentence", func() *stage {
			b := answered().reopen("K-1")
			return b.dispatch(b.at(), compLens).register(compLens)
		}, compLens, recordpb.Soundness_SOUNDNESS_UNSOUND, false},
		{"by another lens", func() *stage {
			return answered().dispatch(2, evLens).register(evLens)
		}, evLens, recordpb.Soundness_SOUNDNESS_UNSOUND, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, err := Append(Identity{Run: c.board().seed(), SeatID: c.seat}, again(c.as))
			if !c.refuse {
				if err != nil {
					t.Fatalf("refused: %v", err)
				}
				return
			}
			mustRefuse(t, err, "K-1", "gap G-1", "still open", "has not changed", "Nothing is recorded")
		})
	}
}

// THE JUDGE OWES ITS GAP, NOT THE FIX: its lens fold names the debt until the gap is minted, and not
// after, so a lens that has carried its judgement to blue may retire.
func TestTheJudgeOwesItsGapAndNoMore(t *testing.T) {
	owes := func(run Run) bool {
		fam, err := FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		ids := fam.At.IDs(fam.Events)
		debts := lensOwesOf(fam.Events, ids, fam.At, nil, owedBlockersOf(fam.Events, ids, fam.At, nil))
		return slices.Contains(debts[compLens], "failed_check K-1")
	}
	before := owes(unsoundBoard(t).seed())
	after := owes(unsoundBoard(t).add(compLens, aboutComputation("G-1", "K-1")).seed())
	if !before || after {
		t.Errorf("the judge owes failed_check K-1 before its gap = %v, after = %v; want true, then false", before, after)
	}
}

// BLUE'S CAUSE MOVES WITH THE GAP: blue that sat after the judgement and left it is readied again
// once the judge's gap states the defect and its fix.
func TestBlueIsReadiedAgainByTheJudgesGap(t *testing.T) {
	b := unsoundBoard(t)
	b.dispatch(b.at(), blueRespondSeat).register(blueRespondSeat).register(compLens).
		add(compLens, aboutComputation("G-1", "K-1"))
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Why {
		if strings.HasPrefix(w, blueRespondSeat+": computation K-1 stands on a failed check") {
			return
		}
	}
	t.Errorf("blue is not readied by the judge's gap minted after its sitting: why %q", plan.Why)
}
