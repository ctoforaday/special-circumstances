package record

import (
	"reflect"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE PLAN RELAYS EACH ENGAGED GAP'S EXCHANGES, AND TELLS NOT-MEASURED FROM ZERO. The engine's
// no-progress valve reads these: a gap marching to impasse keeps its head and its parties while its
// count advances. A gap whose sittings the record cannot close carries them as `unresolved` beside a
// count of 0, never as a bare 0 — the bytes an undisputed gap produces (#1002).
func TestThePlanRelaysEachEngagedGapsExchanges(t *testing.T) {
	marched := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").edit("G1", "was", "is").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").edit("G1", "is", "will be").
		register("red-chair").dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond").
		register("red-chair")
	unclosed := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		dispatch(1, evLens, "G1").dispatch(1, "blue-respond", "G1").
		register(evLens).register("blue-respond")
	for _, c := range []struct {
		name string
		st   *stage
		want []ExchangeCount
	}{
		{"two counted, the last epoch's sittings unclosed", marched, []ExchangeCount{{GapID: "G1", Count: 2, Unresolved: 2}}},
		{"nothing counted and two sittings unclosed", unclosed, []ExchangeCount{{GapID: "G1", Count: 0, Unresolved: 2}}},
	} {
		t.Run(c.name, func(t *testing.T) {
			plan, err := PlanDispatch(c.st.seed())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(plan.Exchanges, c.want) {
				t.Fatalf("exchanges = %+v, want %+v (parties %+v)", plan.Exchanges, c.want, plan.Parties)
			}
		})
	}
	if got := emptyPlan().Exchanges; got == nil || len(got) != 0 {
		t.Errorf("the empty plan's exchanges = %#v, want an empty array, never null", got)
	}
}

// AT THE EPOCH LIMIT THE PLAN HAS NO PARTIES, AND SO NO EXCHANGES ENTRY: an entry is a gap a party
// of this plan is engaged on.
func TestAPlanAtTheEpochLimitRelaysNoExchanges(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).mint(evLens, "G1", "high").
		register("red-chair")
	run := b.seed()
	r, err := planReadsOf(run)
	if err != nil {
		t.Fatal(err)
	}
	p := DefaultParams
	p.MaxEpochs = 1
	plan := foldPlan(p, r)
	if !plan.EpochLimitReached || len(plan.Parties) != 0 {
		t.Fatalf("the fixture is not at its epoch limit: %+v", plan)
	}
	if len(plan.Exchanges) != 0 {
		t.Errorf("exchanges = %+v at the epoch limit, where no party is engaged", plan.Exchanges)
	}
}

func planReadsOf(run Run) (planReads, error) {
	var r planReads
	err := readSnapshot(run, func(q recordsql.Querier) error {
		var err error
		r, err = planReadsAt(q)
		return err
	})
	return r, err
}
