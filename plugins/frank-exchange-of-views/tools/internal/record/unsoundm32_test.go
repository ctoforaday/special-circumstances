package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

const m32Archive = "2026-10-24_is-91-prime-m32.tar.gz"

// M32 ENDED UNVERIFIED OVER A COMPUTATION ITS RECORD CALLED UNSOUND. The computation lens judged
// K-5ba8edea's sentence unsound at event 353 — it names a sieve, and the script does trial division —
// minted no gap, and retired; the plan said pass_permitted with no blockers, and the final chair
// (event 535) withheld its verdict because the gate would have admitted a PASS over that sentence.
// Under the rule the judgement holds the PASS, names the computation and its judge, and readies the lens to
// mint its gap and blue to fix or retire the computation. (Events the migration writes as no
// event — m32's findings and their anchors among them — leave every event after them earlier than
// the archive placed it, so each act is found by what it is, not by where the archive put it.)
func TestM32UnsoundComputationHoldsThePassAndReadies(t *testing.T) {
	whole := archivedEvents(t, m32Archive)
	judged, final := -1, -1
	for i, e := range whole {
		if r := e.GetReproduce(); r.GetComputationId() == "K-5ba8edea" && r.GetSoundness() == recordpb.Soundness_SOUNDNESS_UNSOUND && e.GetSeatId() == compLens {
			judged = i
		}
		if e.GetPosition() != nil && e.GetSeatId() == "red-chair" {
			final = i
		}
	}
	if judged < 0 || final < judged {
		t.Fatalf("the migrated m32 holds no unsound judgement of K-5ba8edea by %s (at %d) before the final chair's position (at %d)", compLens, judged, final)
	}
	// THE CHAIR'S LAST SITTING: the record as the final chair read it, before its position.
	run := archiveAt(t, m32Archive, whole, final)
	got := unsoundOf(t, run)
	if len(got) != 1 || got[0].Subject != "K-5ba8edea" || got[0].Owner != compLens || got[0].AnsweredBy() != "" {
		t.Fatalf("unsound blockers = %+v, want K-5ba8edea judged by %s and carried by no gap", got, compLens)
	}
	err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil || !strings.Contains(err.Error(), "K-5ba8edea") || !strings.Contains(err.Error(), compLens) {
		t.Errorf("the PASS over K-5ba8edea: %v, want it refused naming the computation and its judge", err)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over K-5ba8edea's standing unsound judgement")
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
	// judgement and that plan — its edits at 384 to 393 — with the fix-or-retire item on its work list,
	// so its reason is that it sat since; it is a party all the same, for G-ae1849d4.
	first := 0
	for i := judged + 1; i < len(whole); i++ {
		if whole[i].GetType() == recordpb.EventType_EVENT_TYPE_PLAN {
			first = i
			break
		}
	}
	if first == 0 {
		t.Fatal("no plan follows the judgement")
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
		compLens:        compLens + "'s check of computation K-5ba8edea judged it unsound and no gap of its carries that to blue — ready",
		blueRespondSeat: blueRespondSeat + " has sat since computation K-5ba8edea's check failed",
	}
	for seat, why := range wants {
		said := false
		for _, w := range plan.Why {
			said = said || strings.Contains(w, why)
		}
		if !readied[seat] || !said {
			t.Errorf("the plan before event %d does not ready %s for K-5ba8edea: parties %+v, why %q", first+1, seat, plan.Parties, plan.Why)
		}
	}
	if plan.PassPermitted {
		t.Errorf("the plan before event %d permits a PASS over K-5ba8edea", first+1)
	}
}
