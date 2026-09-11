package record

import (
	"strings"
	"testing"
)

// epochTwo is a record at its second chair sitting with the lens and blue ready on G1.
func epochTwo(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "high").
		register("red-chair") // epoch 2
}

// THE EPOCH LIMIT IS A TERM, AND ITS CEILING IS DERIVED. The chair sitting that opens the last
// epoch dispatches nobody; the plan says why, and the record derives CEILING with that why —
// distinguishable from the carried-gaps ceiling — so `bench outcome --as CEILING` is accepted.
func TestTheSittingThatOpensTheLastEpochDispatchesNobodyAndTheRunEndsCeiling(t *testing.T) {
	b := epochTwo(t)
	run := b.seed()
	writeRunConfig(t, b.runDir, `{"maxEpochs":2}`)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.EpochLimitReached || !plan.Ceiling || plan.PassPermitted || len(plan.Parties) != 0 || plan.MaxEpochs != 2 {
		t.Fatalf("plan at epoch 2 of 2 = %+v, want nobody dispatched, ceiling, the limit named", plan)
	}
	if !strings.Contains(strings.Join(plan.Why, "\n"), "epoch limit 2 reached") {
		t.Errorf("why does not name the limit: %v", plan.Why)
	}
	got, why, ok := DeriveVerdict(run)
	if !ok || got != "CEILING" || !strings.Contains(why, "epoch limit 2 reached") {
		t.Errorf("derived (%q, %q, %v), want CEILING naming the epoch limit", got, why, ok)
	}

	// Under the limit the board's parties are dispatched as ever.
	writeRunConfig(t, b.runDir, `{"maxEpochs":3}`)
	if plan, _ := PlanDispatch(run); plan.EpochLimitReached || plan.Ceiling || len(plan.Parties) != 2 {
		t.Errorf("plan at epoch 2 of 3 = %+v, want the lens and blue on G1", plan)
	}
	// A run whose config states no limit is held to none.
	writeRunConfig(t, b.runDir, `{"k":2}`)
	if plan, _ := PlanDispatch(run); plan.EpochLimitReached || plan.MaxEpochs != 0 || len(plan.Parties) != 2 {
		t.Errorf("plan with no maxEpochs on the run = %+v, want no limit", plan)
	}
}

// A board that permits PASS at the last epoch is not at the epoch limit: nobody was ready anyway,
// and the chair may still record the PASS.
func TestAPassPermittedAtTheLastEpochIsNotTheEpochLimit(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).register("red-chair")
	run := b.seed()
	writeRunConfig(t, b.runDir, `{"maxEpochs":2}`)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.PassPermitted || plan.EpochLimitReached || plan.Ceiling {
		t.Fatalf("plan = %+v, want PASS permitted and no epoch-limit ceiling", plan)
	}
}
