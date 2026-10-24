package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A STANDING UNSOUND JUDGEMENT IS A PASS BLOCKER (gblock, 2026-10-24): the judging lens is readied to
// mint its gap, blue to fix or retire the proof, and it clears when the proof is judged sound again or
// retired. universe-m32 ended UNVERIFIED with pass_permitted true over exactly this board.

// unsoundBoard is a run whose computation lens re-ran P-1 and judged its sentence unsound in its
// sitting, with the chair seated after it.
func unsoundBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		proof("P-1", "s1").
		dispatch(2, compLens).register(compLens).
		reproduce(compLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
}

// aboutProof is a gap about proof anchor, graded low: not material, so it readies nobody itself.
func aboutProof(gap, anchor string) *recordpb.Mint {
	m := budgetMint(gap)
	m.Severity = recordpb.Grade_GRADE_LOW.Enum()
	m.Problem = proto.String("the script does trial division, and the sentence calls it a sieve")
	m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_PROOF.Enum(), proto.String(anchor)
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
		if b.Kind == BlockerUnsoundProof {
			out = append(out, b)
		}
	}
	return out
}

func TestAStandingUnsoundJudgementHoldsThePass(t *testing.T) {
	run := unsoundBoard(t).seed()
	got := unsoundOf(t, run)
	if len(got) != 1 || got[0].Subject != "P-1" || got[0].Owner != compLens {
		t.Fatalf("unsound blockers = %+v, want P-1 owned by its judge %s", got, compLens)
	}
	err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil {
		t.Fatal("the gate admits a PASS over a proof the record calls unsound")
	}
	for _, want := range []string{"P-1", compLens, "judged unsound"} {
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
			b.reproduce(compLens, "P-1", "s1", recordpb.Rerun_RERUN_NOT_RUN, recordpb.Soundness_SOUNDNESS_SOUND)
		}, true},
		{"blue retires the proof", func(b *stage) {
			b.add("blue-respond", &recordpb.Retire{Claim: proto.String("the sentence"), Reason: proto.String("refuted"), Anchors: []string{"P-1"}})
		}, true},
		{"a sound judgement after blue's edit", func(b *stage) {
			b.reopen("P-1").dispatch(b.at(), evLens).register(evLens).
				reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_NOT_RUN, recordpb.Soundness_SOUNDNESS_SOUND)
		}, true},
		{"another lens's sound beside it", func(b *stage) {
			b.dispatch(2, evLens).register(evLens).
				reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_NOT_RUN, recordpb.Soundness_SOUNDNESS_SOUND)
		}, false},
		{"the judge's gap alone", func(b *stage) { b.add(compLens, aboutProof("G-1", "P-1")) }, false},
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

// THE JUDGE IS READIED TO MINT ITS GAP, AND BLUE TO FIX OR RETIRE THE PROOF — each once per cause.
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
	if !parties(plan)[compLens] || !says(plan, compLens+" judged proof P-1 unsound and no gap of its carries that to blue — ready") {
		t.Errorf("the judge is not readied to mint its gap: parties %v, why %q", plan.Parties, plan.Why)
	}
	if !parties(plan)[blueRespondSeat] || !says(plan, "proof P-1 stands judged unsound by "+compLens+" — ready, to fix the proof or retire it") {
		t.Errorf("blue is not readied to fix or retire the proof: parties %v, why %q", plan.Parties, plan.Why)
	}
	lens := sittingOfRunT(t, run, "lens", compLens)
	if it, ok := itemNaming(lens, "proof P-1"); !ok || !it.Blocks {
		t.Errorf("the judge's work list does not owe its gap about P-1: %+v", lens.Open)
	}
	blue := sittingOfRunT(t, run, "blue", blueRespondSeat)
	if it, ok := itemNaming(blue, "proof P-1"); !ok || !it.Blocks || !strings.Contains(it.What, "retire") {
		t.Errorf("blue's work list does not owe the fix or retire of P-1: %+v", blue.Open)
	}

	// The judge mints its gap: it owes nothing more on it, and blue's cause moves to the mint.
	minted := func() *stage { return unsoundBoard(t).add(compLens, aboutProof("G-1", "P-1")) }
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
	if it, ok := itemNaming(lens, "proof P-1"); !ok || it.Blocks {
		t.Errorf("the judge's item on P-1 still blocks once its gap is minted: %+v", lens.Open)
	}

	// Blue sits after the mint and leaves it: not readied again for it.
	b := minted()
	b.dispatch(b.at(), blueRespondSeat).register(blueRespondSeat).register(compLens)
	plan, err = PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	if !says(plan, blueRespondSeat+" has sat since proof P-1 was judged unsound") {
		t.Errorf("blue is readied again for a cause it sat through: why %q", plan.Why)
	}
}

// A GAP THAT CARRIES THE LENS'S OWN UNSOUND JUDGEMENT TAKES NO MINT. Every other mint counts.
func TestTheGapAboutAnUnsoundJudgementTakesNoMint(t *testing.T) {
	b := unsoundBoard(t)
	writeRunConfig(t, b.runDir, `{"mintBudget":1}`)
	b.proof("P-2", "s2")
	run := b.seed()
	md := "# Report\n\nA paragraph that cites nothing."
	withRenderer(t, renders(&md))
	id := Identity{Run: run, SeatID: compLens}
	// The computation lens's budget is max(1, ceil(2 proofs × 3 / 2)) = 3.
	mustAppend(t, id, budgetMint("G-00000001"))
	mustAppend(t, id, budgetMint("G-00000002"))
	mustAppend(t, id, budgetMint("G-00000003"))
	_, err := Append(id, aboutProof("G-00000004", "P-2"))
	mustRefuse(t, err, "admits no more")
	mustAppend(t, id, aboutProof("G-00000005", "P-1"))
	w, err := WorkOfSeat(run, "lens", compLens)
	if err != nil {
		t.Fatal(err)
	}
	if mb := w.Sitting.MintBudget; mb == nil || mb.Open != 3 || mb.Answering != 1 {
		t.Errorf("mint_budget = %+v, want open 3, answering 1", mb)
	}
	_, err = Append(id, aboutProof("G-00000006", "P-1"))
	mustRefuse(t, err, "admits no more")
}

// A GAP OR FINDING ABOUT A PROOF NAMES ONE: the write refuses an anchor no proof carries.
func TestAnActAboutAProofNamesOne(t *testing.T) {
	run := unsoundBoard(t).seed()
	_, err := Append(Identity{Run: run, SeatID: compLens}, aboutProof("G-1", "P-9"))
	mustRefuse(t, err, "P-9", "no proof on this record carries")
}

// A GAP MINTED IN THE JUDGEMENT'S OWN SITTING ANSWERS IT, before or after the reproduce; one from an
// earlier sitting carried a judgement of a sentence since replaced, and answers nothing now.
func TestTheJudgementsOwnSittingsGapAnswersIt(t *testing.T) {
	b := newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		proof("P-1", "s1").dispatch(2, compLens).register(compLens).
		add(compLens, aboutProof("G-1", "P-1")).
		reproduce(compLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
	if got := unsoundOf(t, b.seed()); len(got) != 1 || got[0].AnsweredBy() != "G-1" {
		t.Errorf("unsound blockers = %+v, want P-1 answered by G-1, minted in the judgement's sitting", got)
	}
	b = newStage(t).cast(evLens, compLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		proof("P-1", "s1").dispatch(2, compLens).register(compLens).
		add(compLens, aboutProof("G-1", "P-1")).
		dispatch(2, compLens).register(compLens).
		reproduce(compLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
	if got := unsoundOf(t, b.seed()); len(got) != 1 || got[0].AnsweredBy() != "" {
		t.Errorf("unsound blockers = %+v, want P-1 answered by no gap: G-1 was minted in an earlier sitting", got)
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
		return slices.Contains(debts[compLens], "unsound_proof P-1")
	}
	before := owes(unsoundBoard(t).seed())
	after := owes(unsoundBoard(t).add(compLens, aboutProof("G-1", "P-1")).seed())
	if !before || after {
		t.Errorf("the judge owes unsound_proof P-1 before its gap = %v, after = %v; want true, then false", before, after)
	}
}

// BLUE'S CAUSE MOVES WITH THE GAP: blue that sat after the judgement and left it is readied again
// once the judge's gap states the defect and its fix.
func TestBlueIsReadiedAgainByTheJudgesGap(t *testing.T) {
	b := unsoundBoard(t)
	b.dispatch(b.at(), blueRespondSeat).register(blueRespondSeat).register(compLens).
		add(compLens, aboutProof("G-1", "P-1"))
	plan, err := PlanDispatch(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range plan.Why {
		if strings.HasPrefix(w, blueRespondSeat+": proof P-1 stands judged unsound") {
			return
		}
	}
	t.Errorf("blue is not readied by the judge's gap minted after its sitting: why %q", plan.Why)
}
