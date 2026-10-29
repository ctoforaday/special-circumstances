package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

const m32Archive = "2026-10-24_is-91-prime-m32.tar.gz"

// M32 ENDED UNVERIFIED OVER A PROOF ITS RECORD CALLED UNSOUND. The computation lens judged
// P-5ba8edea's sentence unsound at event 365 — it names a sieve, and the script does trial division —
// minted no gap, and retired; the plan said pass_permitted with no blockers, and the final chair
// (event 556) withheld its verdict because the gate would have admitted a PASS over that sentence.
// Under the rule the judgement holds the PASS, names the proof and its judge, and readies the lens to
// mint its gap and blue to fix or retire the proof.
func TestM32UnsoundProofHoldsThePassAndReadies(t *testing.T) {
	whole := archivedEvents(t, m32Archive)
	r := whole[364].GetReproduce()
	if r.GetProofId() != "P-5ba8edea" || r.GetSoundness() != recordpb.Soundness_SOUNDNESS_UNSOUND || whole[364].GetSeatId() != compLens {
		t.Fatalf("event 365 is not the computation lens's unsound judgement of P-5ba8edea: %v", whole[364])
	}
	if whole[555].GetSeatId() != "red-chair" {
		t.Fatalf("event 556 is not the final chair's: %v", whole[555])
	}
	// THE CHAIR'S LAST SITTING: the record as the final chair read it, before its position.
	run := archiveAt(t, m32Archive, whole, 555)
	got := unsoundOf(t, run)
	if len(got) != 1 || got[0].Subject != "P-5ba8edea" || got[0].Owner != compLens || got[0].AnsweredBy() != "" {
		t.Fatalf("unsound blockers = %+v, want P-5ba8edea judged by %s and carried by no gap", got, compLens)
	}
	err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil || !strings.Contains(err.Error(), "P-5ba8edea") || !strings.Contains(err.Error(), compLens) {
		t.Errorf("the PASS over P-5ba8edea: %v, want it refused naming the proof and its judge", err)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over P-5ba8edea's standing unsound judgement")
	}
	readied := map[string]bool{}
	for _, p := range plan.Parties {
		readied[p.SeatID] = true
	}
	// A DEBT KEEPS ITS JUDGE FROM RETIRING: m32's computation lens retired with the judgement standing.
	if !readied[compLens] {
		t.Errorf("%s is retired over its own standing unsound judgement: parties %+v, why %q", compLens, plan.Parties, plan.Why)
	}

	// THE FIRST PLAN AFTER THE JUDGEMENT readies the lens to mint its gap. Blue sat between the
	// judgement and that plan — its edits at 396 to 405 — with the fix-or-retire item on its work list,
	// so its reason is that it sat since; it is a party all the same, for G-ae1849d4.
	first := 0
	for i := 365; i < len(whole); i++ {
		if whole[i].GetType() == recordpb.EventType_EVENT_TYPE_PLAN {
			first = i
			break
		}
	}
	if first == 0 {
		t.Fatal("no plan follows the judgement at event 365")
	}
	plan, err = PlanDispatch(archiveAt(t, m32Archive, whole, first))
	if err != nil {
		t.Fatal(err)
	}
	readied = map[string]bool{}
	for _, p := range plan.Parties {
		readied[p.SeatID] = true
	}
	wants := map[string]string{
		compLens:        compLens + "'s check of proof P-5ba8edea judged it unsound and no gap of its carries that to blue — ready",
		blueRespondSeat: blueRespondSeat + " has sat since proof P-5ba8edea's check failed",
	}
	for seat, why := range wants {
		said := false
		for _, w := range plan.Why {
			said = said || strings.Contains(w, why)
		}
		if !readied[seat] || !said {
			t.Errorf("the plan before event %d does not ready %s for P-5ba8edea: parties %+v, why %q", first+1, seat, plan.Parties, plan.Why)
		}
	}
	if plan.PassPermitted {
		t.Errorf("the plan before event %d permits a PASS over P-5ba8edea", first+1)
	}
}
