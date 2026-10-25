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

// THE PLAN STORED IS THE RECORD AS THE VERB LEFT IT, and its own rows move nothing that holds the
// PASS: the report changed after the chair's last inquiry review, so the review is owed before the
// rows are written and after them alike, and asked again in the sitting the plan stands.
func TestThePlanStoredCarriesTheReviewTheReportsChangeOwes(t *testing.T) {
	chair := func(run Run) Identity { return Identity{Run: run, SeatID: "red-chair"} }
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("I-00000001"), Question: proto.String("a question"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}).
		register("red-chair").add("red-chair", &recordpb.InquiryReview{Reason: proto.String("read")}).dispatch(2, evLens).
		register(evLens).mint(evLens, "G1", "high").ingest().register("red-chair").seed()
	before, err := PlanDispatch(run)
	if err != nil || len(before.Parties) == 0 || !slices.ContainsFunc(before.Blockers, func(b PlanBlocker) bool { return b.Kind == BlockerInquiryReview }) {
		t.Fatalf("fixture: %d parties, blockers %+v, %v; want parties and the review the report's change owes", len(before.Parties), before.Blockers, err)
	}
	plan, _, err := DispatchNext(chair(run))
	if err != nil || !reflect.DeepEqual(plan.Blockers, before.Blockers) {
		t.Fatalf("the stored plan's blockers %+v, %v; want the ones folded before its rows, %+v", plan.Blockers, err, before.Blockers)
	}
	if again, standing, err := DispatchNext(chair(run)); err != nil || !standing {
		t.Fatalf("asked again = standing %v, %v; want the plan standing: %s then %s", standing, err, plan.Printed(), again.Printed())
	}
	if stored, _ := storedOf(t, run); len(stored) != 1 {
		t.Fatalf("%d plans stored; want one", len(stored))
	}
}

// A STORED PLAN IS NOT WORK IN THE EPOCH. A chair sitting whose plan readies nobody does nothing in
// its epoch by asking: stored and asked again, the plan is the same, the current epoch is still the
// last one with work in it, and the sitting reads as having recorded nothing. The report has not
// changed since the last inquiry review, so the plan holds no review.
func TestAStoredPlanMovesNoEpoch(t *testing.T) {
	chair := func(run Run) Identity { return Identity{Run: run, SeatID: "red-chair"} }
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("I-00000001"), Question: proto.String("a question"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}).
		register("red-chair").add("red-chair", &recordpb.InquiryReview{Reason: proto.String("read")}).dispatch(2, evLens).
		register(evLens).served(evLens, 2).register("red-chair").seed()
	plan, _, err := DispatchNext(chair(run))
	if err != nil || len(plan.Parties) != 0 {
		t.Fatalf("fixture: %d parties, %v; want nobody ready", len(plan.Parties), err)
	}
	again, standing, err := DispatchNext(chair(run))
	if err != nil || !standing || again.Printed() != plan.Printed() || slices.ContainsFunc(again.Blockers, func(b PlanBlocker) bool { return b.Kind == BlockerInquiryReview }) {
		t.Fatalf("asked again: standing %v, blockers %+v, %v; want the plan standing with no review owed:\n%s\n%s", standing, again.Blockers, err, plan.Printed(), again.Printed())
	}
	evs, win, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	if got := win.CurrentEpoch(evs); got != 1 {
		t.Errorf("CurrentEpoch = %d after the chair stored a plan and did nothing else; want 1, the last epoch with work in it", got)
	}
	// Nor is it an act of the sitting: the chair recorded nothing, so no log is asked of it.
	if !sittingRecordedNothing(evs, win, "red-chair") {
		t.Errorf("the chair's sitting, whose only event is its stored plan, reads as having recorded something")
	}
}

// THE CHAIR'S OWN ACTS MOVE NOTHING THAT HOLDS THE PASS. Seen on a smoke run (2026-10-18): a plan
// readying nobody said pass_permitted with no blocker over a report unchanged since the chair's last
// inquiry review, the chair's work list agreed, and the chair's position then put the review on the
// work list and the gate. The other order passed over a review that predated the report's last
// change: a PASS issued as the sitting's first act was admitted. Here the plan, the work list after
// every act of the sitting, and the gate are one judgment: whatever the chair files that is not the
// review, the gate's list and the blocking items are the plan's; the gate admits the PASS exactly
// where the plan named nothing, and refuses it on the review the plan named, which the review then
// clears.
func TestTheChairsOwnActsMoveNothingThatHoldsThePass(t *testing.T) {
	position := &recordpb.Position{Text: proto.String("ask the bench to hold the PASS")}
	logged := &recordpb.Log{Text: proto.String("nothing cost a call"),
		Type: recordpb.LogType_LOG_TYPE_FRICTION.Enum(), Source: recordpb.LogSource_LOG_SOURCE_SEAT.Enum()}
	inquiry := &recordpb.Inquiry{InquiryId: proto.String("I-00000001"), Question: proto.String("a question"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()}
	boards := []struct {
		name string
		// build is the record the chair's new sitting opens on: inquiries, the chair's last review,
		// and a lens that has sat at the head, so nobody is ready.
		build      func(b *stage) *stage
		wantReview bool
	}{
		{"the report unchanged since the last review", func(b *stage) *stage {
			return b.ingest().add("blue-respond", inquiry).register("red-chair").
				add("red-chair", &recordpb.InquiryReview{Reason: proto.String("read")}).
				dispatch(2, evLens).register(evLens).served(evLens, 2)
		}, false},
		{"the report changed after the last review", func(b *stage) *stage {
			b.ingest().add("blue-respond", inquiry).register("red-chair").
				add("red-chair", &recordpb.InquiryReview{Reason: proto.String("read")}).ingest()
			return b.sit(b.at(), evLens)
		}, true},
	}
	acts := []struct {
		name string
		acts []proto.Message
	}{
		{"the verdict first", nil},
		{"a position, then the verdict", []proto.Message{position}},
		{"a log, then the verdict", []proto.Message{logged}},
		{"a position and a log, then the verdict", []proto.Message{position, logged}},
	}
	for _, bd := range boards {
		for _, c := range acts {
			t.Run(bd.name+"/"+c.name, func(t *testing.T) {
				run := bd.build(newStage(t).cast(evLens, "red-chair", "blue-respond", "judge")).register("red-chair").seed()
				chair := Identity{Run: run, SeatID: "red-chair"}
				plan, _, err := DispatchNext(chair)
				if err != nil || len(plan.Parties) != 0 || !plan.PassPermitted {
					t.Fatalf("fixture: %d parties, pass_permitted %v, %v; want nobody ready and PASS permitted", len(plan.Parties), plan.PassPermitted, err)
				}
				// The rule: a review recorded after the report and the inquiries last changed answers,
				// and one recorded before does not; the plan names the review owed as the chair's own.
				want := []PlanBlocker{}
				if bd.wantReview {
					want = []PlanBlocker{{Kind: BlockerInquiryReview, Owner: chairSeat}}
				}
				if !reflect.DeepEqual(plan.Blockers, want) {
					t.Errorf("the plan's blockers = %+v; want %+v", plan.Blockers, want)
				}
				atDispatch := blockingItems(sittingOfRunT(t, run, "chair", "red-chair"))
				blockersNow := func() []PlanBlocker {
					t.Helper()
					bs, err := PassBlockers(run)
					if err != nil {
						t.Fatal(err)
					}
					out := []PlanBlocker{}
					for _, b := range bs {
						out = append(out, PlanBlocker{Kind: b.Kind, Subject: b.Subject, Owner: b.Owner})
					}
					return out
				}
				for i, act := range c.acts {
					if _, err := Append(chair, act); err != nil {
						t.Fatalf("act %d: %v", i+1, err)
					}
					if got := blockersNow(); !reflect.DeepEqual(got, plan.Blockers) {
						t.Errorf("after act %d (%T) the gate's list is %+v; the plan said %+v", i+1, act, got, plan.Blockers)
					}
					if got := blockingItems(sittingOfRunT(t, run, "chair", "red-chair")); strings.Join(got, "\n") != strings.Join(atDispatch, "\n") {
						t.Errorf("after act %d (%T) the work list blocks on\n  %q\nand at the dispatch it blocked on\n  %q", i+1, act, got, atDispatch)
					}
				}
				_, gerr := Append(chair, passGate)
				if len(plan.Blockers) == 0 {
					if gerr != nil {
						t.Fatalf("the plan named no blocker and the gate refused the PASS: %v", gerr)
					}
					return
				}
				if gerr == nil || !strings.Contains(gerr.Error(), "no inquiry review is recorded since the report or its inquiries last changed") {
					t.Fatalf("PASS over the plan's blocker = %v; want it refused on the inquiry review the plan named", gerr)
				}
				if _, err := Append(chair, &recordpb.InquiryReview{Reason: proto.String("read the report as it now stands")}); err != nil {
					t.Fatal(err)
				}
				if got := blockersNow(); len(got) != 0 {
					t.Fatalf("after the review the gate's list is %+v; want it empty", got)
				}
				if _, err := Append(chair, passGate); err != nil {
					t.Fatalf("PASS after the review: %v", err)
				}
			})
		}
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
		{"a plan its rules refuse", emptied.WithKey().Printed(), "empty only when"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Append(sit(t, run, "red-chair"), &recordpb.Plan{Printed: proto.String(c.printed)}); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("storing %s = %v, want a refusal naming %q", c.name, err, c.want)
			}
		})
	}
	mustAppend(t, sit(t, run, "red-chair"), &recordpb.Plan{Printed: proto.String(good)})
}

// THE KEY IS WHAT THE PLAN DECIDES: the same whatever epoch asks it and whatever verdict the asking
// sitting recorded, and moved by any other field — a blocker of any kind and the reasons included.
func TestThePlanKeyLeavesOutOnlyWhatTheAskingSittingSays(t *testing.T) {
	var unkeyed []string
	for _, f := range PlanFields() {
		if f.Unkeyed {
			unkeyed = append(unkeyed, f.JSON)
		}
	}
	if strings.Join(unkeyed, ",") != "epoch,final_sitting,verdict" {
		t.Fatalf("the key leaves out %v, want epoch, final_sitting and verdict", unkeyed)
	}
	base := Plan{Head: 4, Parties: []Party{{SeatID: evLens, GapIDs: []string{}}}, Epoch: 2,
		Blockers: []PlanBlocker{{Kind: BlockerLensReady, Subject: evLens, Owner: evLens}}, Why: []string{"ready"}}.WithKey()
	later := base
	later.Epoch, later.FinalSitting, later.Verdict = 5, true, "FAIL"
	if later.WithKey().Key != base.Key {
		t.Errorf("the key moved with the epoch or the verdict")
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
		if p.WithKey().Key == base.Key {
			t.Errorf("the key did not move with %s", name)
		}
	}
}
