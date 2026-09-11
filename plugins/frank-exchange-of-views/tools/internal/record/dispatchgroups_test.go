package record

import (
	"strings"
	"testing"
)

// warmChair is the B6 record's shape in miniature: the chair registers once, asks for the plan
// twice in its first sitting, and comes back warm — no register — after the lens sat.
func warmChair(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest(). // head 2
											register("red-chair").dispatch(2, evLens).dispatch(2, evLens). // asked twice, nobody sat between
											register(evLens).mint(evLens, "G1", "high")
}

// One chair sitting's dispatch is the rows with no register between them: the repeated plan is
// one group, a party's register ends it, and each party's sitting is the register sittingFor finds.
func TestDispatchGroupsSplitWhereSomebodySat(t *testing.T) {
	b := warmChair(t).dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1").register("blue-respond")
	run := b.seed()
	evs := allEvents(t, run)
	gs := DispatchGroups(evs)
	if len(gs) != 2 {
		t.Fatalf("groups = %+v, want 2: the doubled first plan, then the warm second sitting's", gs)
	}
	if len(gs[0].rows) != 2 || len(gs[0].Parties) != 1 || gs[0].Parties[0] != evLens {
		t.Errorf("group 1 = %+v, want both rows of the doubled plan naming the lens once", gs[0])
	}
	if _, sat := gs[0].Sat[evLens]; !sat {
		t.Errorf("the lens registered after group 1 and is not recorded as having sat: %+v", gs[0].Sat)
	}
	if _, sat := gs[1].Sat[evLens]; sat {
		t.Errorf("the lens has not registered since group 2 and is recorded as having sat: %+v", gs[1].Sat)
	}
	if _, sat := gs[1].Sat["blue-respond"]; !sat {
		t.Errorf("blue registered after group 2 and is not recorded as having sat: %+v", gs[1].Sat)
	}
}

// chairOwed is the chair's blocking "register for this sitting" items.
func chairOwed(t *testing.T, run Run) []Item {
	t.Helper()
	var out []Item
	for _, it := range sittingOfRunT(t, run, "merge", "red-chair").Open {
		if it.Blocks && strings.Contains(it.What, "register for this sitting") {
			out = append(out, it)
		}
	}
	return out
}

// THE CHAIR OWES A REGISTER PER SITTING, and its work list says so once a party has sat for its
// last dispatch — the workflow has come back to it — while `dispatch next` refuses until it lands.
func TestAWarmChairOwesTheRegisterThatOpensItsSitting(t *testing.T) {
	// Its first sitting, still in progress: nobody has sat for the dispatch, nothing is owed.
	if owed := chairOwed(t, newStage(t).cast(evLens, "red-chair").ingest().register("red-chair").dispatch(2, evLens).seed()); len(owed) != 0 {
		t.Fatalf("the chair is still in the sitting that dispatched and owes %+v", owed)
	}
	if err := RequireChairSittingOpened(newStage(t).cast(evLens, "red-chair").ingest().register("red-chair").dispatch(2, evLens).seed()); err != nil {
		t.Fatalf("dispatch refused in the sitting that dispatched: %v", err)
	}

	// Back warm, after the lens sat: owed, and dispatch refuses.
	warm := warmChair(t).seed()
	owed := chairOwed(t, warm)
	if len(owed) != 1 || !strings.Contains(owed[0].What, evLens) || !strings.Contains(owed[0].What, "head 2") {
		t.Fatalf("owed = %+v, want one blocking item naming the lens that sat and head 2", owed)
	}
	if err := RequireChairSittingOpened(warm); err == nil || !strings.Contains(err.Error(), "Register for this sitting") {
		t.Fatalf("dispatch in an unopened chair sitting = %v, want the refusal naming the register", err)
	}

	// It registers: nothing is owed, and dispatch goes ahead.
	opened := warmChair(t).register("red-chair").seed()
	if owed := chairOwed(t, opened); len(owed) != 0 {
		t.Fatalf("the chair registered for its sitting and still owes %+v", owed)
	}
	if err := RequireChairSittingOpened(opened); err != nil {
		t.Fatalf("dispatch refused after the chair registered: %v", err)
	}

	// No other seat's list carries the chair's item.
	if owed := sittingOfRunT(t, warm, "lens", evLens); strings.Contains(strings.Join(whats(owed.Open), "|"), "the dispatch you recorded") {
		t.Fatalf("a lens was handed the chair's register: %+v", owed.Open)
	}
}

func whats(items []Item) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.What
	}
	return out
}

// The same plan, nobody sat since: it stands, and the verb records nothing new. A different plan,
// or anybody's register since, is a new decision.
func TestADispatchAskedAgainStands(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair").ingest().register("red-chair").dispatch(2, evLens).seed()
	same := Plan{Head: 2, Parties: []Party{{SeatID: evLens}}}
	if ok, err := DispatchStands(run, same); err != nil || !ok {
		t.Fatalf("the same plan with nobody sat since = %v (%v), want standing", ok, err)
	}
	for name, p := range map[string]Plan{
		"another head":  {Head: 3, Parties: []Party{{SeatID: evLens}}},
		"another gap":   {Head: 2, Parties: []Party{{SeatID: evLens, GapIDs: []string{"G1"}}}},
		"another party": {Head: 2, Parties: []Party{{SeatID: evLens}, {SeatID: "blue-respond"}}},
	} {
		if ok, _ := DispatchStands(run, p); ok {
			t.Errorf("%s: a different plan stood as the recorded one", name)
		}
	}
	sat := newStage(t).cast(evLens, "red-chair").ingest().register("red-chair").dispatch(2, evLens).register(evLens).seed()
	if ok, _ := DispatchStands(sat, same); ok {
		t.Error("the lens sat since the dispatch, and the same plan stood — a new sitting's dispatch would go unrecorded")
	}
}
