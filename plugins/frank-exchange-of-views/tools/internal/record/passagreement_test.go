package record

import (
	"encoding/json"
	"strings"
	"testing"
)

func planJSON(t *testing.T, p Plan) string {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// A PLAN THAT READIES NOBODY PRINTS ITS LISTS AS LISTS. B9's chair relayed a PASS-permitted plan
// verbatim — "parties": null — and the workflow, which refuses a plan without a parties list,
// aborted the run at the sitting that should have ended it.
func TestAPlanThatReadiesNobodyPrintsItsListsAsLists(t *testing.T) {
	pass := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).register("red-chair")
	plan, err := PlanDispatch(pass.seed())
	if err != nil {
		t.Fatal(err)
	}
	if !plan.PassPermitted {
		t.Fatalf("plan = %+v, want PASS permitted", plan)
	}
	js := planJSON(t, plan)
	for _, want := range []string{`"parties":[]`, `"docket":[]`} {
		if !strings.Contains(js, want) {
			t.Errorf("a PASS-permitted plan prints %s, want %s in it", js, want)
		}
	}

	limit := epochTwo(t)
	run := limit.seed()
	writeRunConfig(t, limit.runDir, `{"maxEpochs":2}`)
	plan, err = PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.EpochLimitReached {
		t.Fatalf("plan = %+v, want the epoch limit reached", plan)
	}
	if js := planJSON(t, plan); !strings.Contains(js, `"parties":[]`) {
		t.Errorf("the epoch-limit plan prints %s, want \"parties\":[] in it", js)
	}
}

// passItems is the chair's blocking items that name a gap refusing PASS.
func passItems(s SittingJSON) []string {
	var out []string
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, "PASS is refused") {
			out = append(out, it.What)
		}
	}
	return out
}

// THE CHAIR'S WORK LIST NAMES THE GAPS THE PASS GATE REFUSES OVER, AND ONLY THOSE. B9's chair was
// told two below-material gaps refused PASS while dispatch said pass_permitted and the verdict
// accepted it; it settled the contradiction by trying the verdict.
func TestTheChairsPassItemsAreTheGatesGaps(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		mint(evLens, "G1", "low").mint(evLens, "G2", "medium").
		register("red-chair")
	items := passItems(sittingOfRunT(t, b.seed(), "chair", "red-chair"))
	if len(items) != 1 || !strings.Contains(items[0], "gap G2 is open and material") {
		t.Fatalf("PASS items = %q, want one naming G2 as material and none for the low G1", items)
	}

	// A superseded ancestor holds the gate whatever its grade, and the list says so.
	s := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		mint(evLens, "G1", "low").mintSuperseding(evLens, "G3", "low", "G1").
		register("red-chair")
	items = passItems(sittingOfRunT(t, s.seed(), "chair", "red-chair"))
	if len(items) != 1 || !strings.Contains(items[0], "gap G1 is open and superseded") {
		t.Fatalf("PASS items = %q, want one naming the superseded G1 and none for the low G3", items)
	}
}
