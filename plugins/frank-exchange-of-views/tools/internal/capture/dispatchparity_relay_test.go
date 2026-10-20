package capture

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// journalOf writes results as journal.jsonl lines and reads them back through ReadJournal, so the
// audit sees exactly what capture hands it — numbers as json.Number included.
func journalOf(t *testing.T, results ...map[string]any) []map[string]any {
	t.Helper()
	dir := t.TempDir()
	var b strings.Builder
	for i, r := range results {
		line, err := json.Marshal(map[string]any{"agentId": fmt.Sprintf("a%d", i), "result": r})
		if err != nil {
			t.Fatal(err)
		}
		b.Write(line)
		b.WriteString("\n")
	}
	if err := os.WriteFile(filepath.Join(dir, "journal.jsonl"), []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return ReadJournal(dir)
}

// relayed is a chair result whose plan is p as the verb printed it, decoded as the journal holds it.
func relayed(t *testing.T, p record.Plan) map[string]any {
	t.Helper()
	var plan map[string]any
	if err := json.Unmarshal([]byte(p.Printed()), &plan); err != nil {
		t.Fatal(err)
	}
	return map[string]any{"plan": plan}
}

// storedRun is a record the verb dispatched on: the chair asks `dispatch next` at each of its
// sittings, and every party it names sits before the next. It returns the run and the plan each
// sitting printed — the chair's faithful relay, sitting by sitting. With petition set, a petition is
// filed in the first sitting after the chair asked, and the chair asks again.
func storedRun(t *testing.T, sittings int, petition bool) (record.Run, []record.Plan) {
	return dispatchedRun(t, sittings, petition, true)
}

// dispatchedRun is storedRun, with the parties sitting or never sitting.
func dispatchedRun(t *testing.T, sittings int, petition, sit bool) (record.Run, []record.Plan) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	n := 0
	seed := func(seat string, body proto.Message) {
		n++
		recordtest.Seed(t, dir, recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body))
	}
	seed("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}})
	seed("harness", &recordpb.BaseIngest{Text: proto.String("# r\n\nA sentence.")})
	run := runtest.Open(t, dir)
	var plans []record.Plan
	for k := range sittings {
		seed("red-chair", &recordpb.Register{})
		p, _, err := record.DispatchNext(record.Identity{Run: run, SeatID: "red-chair"})
		if err != nil {
			t.Fatal(err)
		}
		if petition && k == 0 {
			seed("red-chair", &recordpb.Motion{MotionId: proto.String("M-00000001"), Basis: proto.String("b"),
				Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}}})
			if p, _, err = record.DispatchNext(record.Identity{Run: run, SeatID: "red-chair"}); err != nil {
				t.Fatal(err)
			}
		}
		plans = append(plans, p)
		for _, party := range p.Parties {
			if !sit {
				break
			}
			if len(party.Occasions) == 0 {
				seed(party.SeatID, &recordpb.Register{})
			}
			for _, w := range party.Occasions {
				o, _ := record.OccasionOf(w)
				seed(party.SeatID, &recordpb.Register{Occasion: o.Enum()})
			}
		}
	}
	return run, plans
}

// A FAITHFUL RELAY IS THE STORED PLAN, sitting by sitting, the plan with no party included.
func TestAFaithfulRelayIsThePlanTheRecordStored(t *testing.T) {
	run, plans := storedRun(t, 2, false)
	var js []map[string]any
	for _, p := range plans {
		js = append(js, relayed(t, p))
	}
	a := DispatchParityAudit(run, journalOf(t, js...), true)
	if a.Verdict != "PASS" || !strings.Contains(a.Detail, "2 relayed plan(s) each equal to the plan the record stored last in its sitting") {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}
}

// ANY FIELD ALTERED IN THE RELAY DEPARTS, by equality, and the departure names it: no field is
// outside the comparison.
func TestARelayAlteredInAnyFieldDeparts(t *testing.T) {
	run, plans := storedRun(t, 1, false)
	for _, f := range record.PlanFields() {
		t.Run(f.JSON, func(t *testing.T) {
			r := relayed(t, plans[0])
			plan := r["plan"].(map[string]any)
			switch f.Type {
			case "integer":
				plan[f.JSON] = json.Number("7")
			case "boolean":
				plan[f.JSON] = !plan[f.JSON].(bool)
			case "string":
				plan[f.JSON] = "0000000000000000"
			case "array":
				if xs := plan[f.JSON].([]any); len(xs) > 0 {
					plan[f.JSON] = xs[1:]
				} else if f.Of == nil {
					plan[f.JSON] = []any{"added"}
				} else {
					t.Skipf("the fixture's %s is empty, and an entry added would have to be a whole one", f.JSON)
				}
			}
			a := DispatchParityAudit(run, journalOf(t, r), true)
			if a.Verdict != "FAIL" || !strings.Contains(a.Detail, f.JSON) {
				t.Fatalf("a relay altered in %s = %s: %s", f.JSON, a.Verdict, a.Detail)
			}
		})
	}
}

// A KEY THE PLAN DOES NOT DECLARE IS A RELAY ALTERED ON THE WAY; a lens's empty `occasions` is not,
// because absent and empty are one state there.
func TestARelayCarryingAnUndeclaredKeyDepartsAndALenssEmptyOccasionsDoesNot(t *testing.T) {
	run, plans := storedRun(t, 1, false)
	r := relayed(t, plans[0])
	r["plan"].(map[string]any)["nested"] = map[string]any{}
	if a := DispatchParityAudit(run, journalOf(t, r), true); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "does not read as a plan") {
		t.Fatalf("an undeclared key = %s: %s", a.Verdict, a.Detail)
	}
	r = relayed(t, plans[0])
	for _, p := range r["plan"].(map[string]any)["parties"].([]any) {
		p.(map[string]any)["occasions"] = []any{}
	}
	if a := DispatchParityAudit(run, journalOf(t, r), true); a.Verdict != "PASS" {
		t.Fatalf("a lens's empty occasions = %s: %s", a.Verdict, a.Detail)
	}
}

// THE PLAN RELAYED IS THE ONE THE CHAIR ASKED FOR LAST: a motion filed after the first asking makes
// the first plan stale, and relaying it departs on what the filing moved.
func TestARelayOfAPlanTheSittingMovedPastDeparts(t *testing.T) {
	run, plans := storedRun(t, 1, true)
	stored, err := record.StoredPlans(mustEvents(t, run))
	if err != nil || len(stored) != 2 {
		t.Fatalf("fixture: %d stored plans, %v; want the plan before the petition and the one after", len(stored), err)
	}
	if a := DispatchParityAudit(run, journalOf(t, relayed(t, plans[0])), true); a.Verdict != "PASS" {
		t.Fatalf("the last plan of the sitting = %s: %s", a.Verdict, a.Detail)
	}
	if a := DispatchParityAudit(run, journalOf(t, relayed(t, stored[0].Plan)), true); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "epoch 1: the relayed plan departs from the plan the record stored last in that sitting on key, parties, blockers, why") {
		t.Fatalf("the plan before the petition = %s: %s", a.Verdict, a.Detail)
	}
}

// A SITTING THE RECORD STORED A PLAN AT IS RELAYED, and a relay names a sitting the record stored a
// plan at; a duplicated chair result — a resume re-journaling it — is compared like any other.
func TestEverySittingsPlanIsRelayedAndEveryRelayIsOfAStoredPlan(t *testing.T) {
	run, plans := storedRun(t, 2, false)
	if a := DispatchParityAudit(run, journalOf(t, relayed(t, plans[0])), true); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "epoch 2: the record stored a plan and the journal relays none") {
		t.Fatalf("a sitting's plan not relayed = %s: %s", a.Verdict, a.Detail)
	}
	if a := DispatchParityAudit(run, journalOf(t, relayed(t, plans[0]), relayed(t, plans[0]), relayed(t, plans[1])), true); a.Verdict != "PASS" {
		t.Fatalf("a chair result journaled twice = %s: %s", a.Verdict, a.Detail)
	}
	never := plans[1]
	never.Epoch = 3
	if a := DispatchParityAudit(run, journalOf(t, relayed(t, plans[0]), relayed(t, plans[1]), relayed(t, never)), true); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "relays a plan at epoch 3, and the record stored none there") {
		t.Fatalf("a relay of a plan never stored = %s: %s", a.Verdict, a.Detail)
	}
}

// NO JOURNAL, OR NO STORED PLAN: nothing was compared, and the detail says so of whichever is absent.
func TestARelayWithNothingToCompareSaysSo(t *testing.T) {
	run, plans := storedRun(t, 1, false)
	if a := DispatchParityAudit(run, nil, false); !strings.Contains(a.Detail, "no workflow journal, so the relayed plans were NOT compared with the plans the record stored") {
		t.Fatalf("no journal = %s: %s", a.Verdict, a.Detail)
	}
	// A record from before the record stored its plans: the rows alone.
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		recordtest.At(t, "harness", "c", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-chair", "blue-respond", "judge"}}),
		recordtest.At(t, "harness", "b", &recordpb.BaseIngest{Text: proto.String("# r")}),
		recordtest.At(t, "red-chair", "r1", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "d1", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-evidence")}),
		recordtest.At(t, "red-lens-evidence", "r2", &recordpb.Register{}))
	old := runtest.Open(t, dir)
	if a := DispatchParityAudit(old, journalOf(t, relayed(t, plans[0])), true); !strings.Contains(a.Detail, "the record stores no plan, so the 1 relayed plan(s) were NOT compared") {
		t.Fatalf("no stored plan = %s: %s", a.Verdict, a.Detail)
	}
}

func mustEvents(t *testing.T, run record.Run) []*record.Event {
	t.Helper()
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return fam.Events
}

// THE ENGINE'S VALVE AND THIS AUDIT READ ONE PLAN. Parties the record dispatches and that never sit
// leave the board as it was, so the record stores the same plan sitting after sitting — the same key,
// the epoch apart — and the engine, handed those relays, stops at the third. The audit holds each
// relay to the stored plan by equality, so the key the valve reads is the record's.
func TestTheValveStopsOnTheKeyTheRecordRepeats(t *testing.T) {
	run, plans := dispatchedRun(t, 3, false, false)
	if plans[0].Key != plans[1].Key || plans[1].Key != plans[2].Key || len(plans[0].Parties) == 0 {
		t.Fatalf("fixture: keys %s %s %s over %d parties; want one key, readying parties", plans[0].Key, plans[1].Key, plans[2].Key, len(plans[0].Parties))
	}
	var js []map[string]any
	for _, p := range plans {
		js = append(js, relayed(t, p))
	}
	if a := DispatchParityAudit(run, journalOf(t, js...), true); strings.Contains(a.Detail, "departs") {
		t.Fatalf("a faithful relay of the stuck plans departs: %s", a.Detail)
	}
	chair := 0
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{"rulings": []any{}, "dispositions": []any{},
			"manifest": []any{}, "open_gaps": 0}
		if seatID == "red-chair" {
			chair++
			e["plan"] = js[min(chair, len(js))-1]["plan"]
		}
		return e
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	_, out, err := debatejs.CaptureRun(debatejs.ScriptPath(filepath.Join(wd, "..", "..")), debatejs.Config{Topic: "t", RunDir: t.TempDir(), BinDir: t.TempDir(),
		Lanes: 1, Model: "haiku", JudgmentModel: "haiku", Backend: backend})
	if err != nil {
		t.Fatalf("driving debate.js: %v", err)
	}
	if out.Epochs != 3 || out.Verdict != "UNVERIFIED" {
		t.Fatalf("the engine ran %d chair sitting(s) to %s on one key repeating; want UNVERIFIED at the third", out.Epochs, out.Verdict)
	}
}
