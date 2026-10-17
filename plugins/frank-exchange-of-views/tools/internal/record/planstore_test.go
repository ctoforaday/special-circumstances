package record

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// storedOf is every plan run's record stored, and how many dispatch rows it holds.
func storedOf(t *testing.T, run Run) ([]StoredPlan, int) {
	t.Helper()
	evs, _, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := StoredPlans(evs)
	if err != nil {
		t.Fatal(err)
	}
	rows := 0
	for _, e := range evs {
		if e.GetDispatch() != nil {
			rows++
		}
	}
	return stored, rows
}

// THE RECORD STORES THE PLAN THE VERB PRINTS, WHOLE — the one with no party as well, which writes no
// row and is the one the workflow reads its exit from.
func TestDispatchNextStoresThePlanItPrints(t *testing.T) {
	chair := func(run Run) Identity { return Identity{Run: run, SeatID: "red-chair"} }
	run := atImpasse(t)
	plan, standing, err := DispatchNext(chair(run))
	if err != nil || standing || len(plan.Parties) == 0 {
		t.Fatalf("fixture: dispatch next = %d parties, standing %v, %v", len(plan.Parties), standing, err)
	}
	stored, _ := storedOf(t, run)
	if len(stored) != 1 || stored[0].Printed != plan.Printed() {
		t.Fatalf("stored %d plan(s); want one, printed as the verb printed it:\n%+v\n%s", len(stored), stored, plan.Printed())
	}

	empty := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").register("red-chair").seed()
	plan, _, err = DispatchNext(chair(empty))
	if err != nil || len(plan.Parties) != 0 {
		t.Fatalf("fixture: a record with no report readies %d parties, %v", len(plan.Parties), err)
	}
	stored, rows := storedOf(t, empty)
	if len(stored) != 1 || stored[0].Printed != plan.Printed() || rows != 0 {
		t.Fatalf("a plan with no party: %d stored, %d rows; want it stored and no row", len(stored), rows)
	}
}

// ASKED AGAIN WITH NOTHING NEW, THE PLAN STANDS; changed only in what holds the PASS, it is stored
// again and its rows stand.
func TestAPlanAskedAgainStandsAndOneChangedInItsBlockersAloneWritesNoRows(t *testing.T) {
	chair := func(run Run) Identity { return Identity{Run: run, SeatID: "red-chair"} }
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("I-00000001"), Question: proto.String("a question"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}).
		register("red-chair").seed()
	first, _, err := DispatchNext(chair(run))
	if err != nil || len(first.Parties) == 0 || !slices.ContainsFunc(first.Blockers, func(b PlanBlocker) bool { return b.Kind == BlockerInquiryReview }) {
		t.Fatalf("fixture: %d parties, blockers %+v, %v; want a party and the chair's inquiry review owed", len(first.Parties), first.Blockers, err)
	}
	_, rows := storedOf(t, run)
	if _, standing, err := DispatchNext(chair(run)); err != nil || !standing {
		t.Fatalf("asked again = standing %v, %v; want the plan standing", standing, err)
	}
	if stored, again := storedOf(t, run); len(stored) != 1 || again != rows {
		t.Fatalf("asked again: %d stored, %d rows (was %d); want nothing new", len(stored), again, rows)
	}
	// The chair's own review, in the sitting it asked in: no sitting opens, and only what holds the
	// PASS moves.
	recordtest.Seed(t, run.Dir(), recordtest.At(t, "red-chair", "review", &recordpb.InquiryReview{Reason: proto.String("read")}))
	second, standing, err := DispatchNext(chair(run))
	if err != nil || standing || second.Key == first.Key || !reflect.DeepEqual(second.Parties, first.Parties) {
		t.Fatalf("after the review: standing %v, key %s (was %s), parties %+v (were %+v), %v; want a new plan readying the same parties",
			standing, second.Key, first.Key, second.Parties, first.Parties, err)
	}
	stored, again := storedOf(t, run)
	if len(stored) != 2 || stored[1].Printed != second.Printed() {
		t.Fatalf("after the review: %d stored; want the second plan stored as printed", len(stored))
	}
	if again != rows {
		t.Errorf("the parties' rows stand and %d more were written", again-rows)
	}
}

// THE WRITE REFUSES A STORED PLAN NO PRINTING OF A PLAN WRITES: a relay is compared with these bytes.
func TestTheWriteRefusesAPlanNoPrintingWrites(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	good := plan.Printed()
	emptied := plan
	emptied.Blockers = []PlanBlocker{}
	for _, c := range []struct{ name, printed, want string }{
		{"a key the plan does not declare", strings.Replace(good, `{"key"`, `{"extra":1,"key"`, 1), "not a plan"},
		{"bytes its printing would not write", strings.Replace(good, `,"parties"`, `, "parties"`, 1), "not as the verb prints it"},
		{"a key that is not the plan's", strings.Replace(good, plan.Key, "0000000000000000", 1), "key is"},
		{"a plan its rules refuse", keyed(emptied).Printed(), "empty only when"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Append(sit(t, run, "red-chair"), &recordpb.Plan{Printed: proto.String(c.printed)}); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("storing %s = %v, want a refusal naming %q", c.name, err, c.want)
			}
		})
	}
	mustAppend(t, sit(t, run, "red-chair"), &recordpb.Plan{Printed: proto.String(good)})
}

// THE KEY IS WHAT THE PLAN DECIDES: the same whatever epoch asks it, and moved by any other field —
// a blocker of any kind and the reasons included.
func TestThePlanKeyLeavesOutOnlyWhenThePlanWasAsked(t *testing.T) {
	var unkeyed []string
	for _, f := range PlanFields() {
		if f.Unkeyed {
			unkeyed = append(unkeyed, f.JSON)
		}
	}
	if strings.Join(unkeyed, ",") != "epoch,final_sitting" {
		t.Fatalf("the key leaves out %v, want epoch and final_sitting", unkeyed)
	}
	base := keyed(Plan{Head: 4, Parties: []Party{{SeatID: evLens, GapIDs: []string{}}}, Epoch: 2,
		Blockers: []PlanBlocker{{Kind: BlockerLensReady, Subject: evLens, Owner: evLens}}, Why: []string{"ready"}})
	later := base
	later.Epoch, later.FinalSitting = 5, true
	if keyed(later).Key != base.Key {
		t.Errorf("the key moved with the epoch")
	}
	for name, change := range map[string]func(*Plan){
		"the head": func(p *Plan) { p.Head = 5 },
		"a blocker": func(p *Plan) {
			p.Blockers = append(p.Blockers, PlanBlocker{Kind: BlockerMaterialGap, Subject: "G1", Owner: evLens})
		},
		"the reasons": func(p *Plan) { p.Why = []string{"ready, reworded"} },
	} {
		p := base
		change(&p)
		if keyed(p).Key == base.Key {
			t.Errorf("the key did not move with %s", name)
		}
	}
}
