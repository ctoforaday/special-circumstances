package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// EVERY FIELD OF THE RELAYED PLAN IS CLASSIFIED, AND EVERY ROW NAMES A FIELD. A field added to
// record.Plan with no row is a field the engine may come to branch on with nothing said about what
// stands behind it; a row naming no field is a claim about a plan that no longer carries it.
func TestEveryRelayedPlanFieldIsClassified(t *testing.T) {
	live := map[string]bool{}
	for _, f := range planFields() {
		live[f.goPath] = true
		if _, ok := relayedPlanFields[f.goPath]; !ok {
			t.Errorf("%s (relayed as %s) has no row in relayedPlanFields — say whether capture compares it, what gate stands behind it, or that the engine does not act on it", f.goPath, f.jsonPath)
		}
		if _, ok := planVariants[f.goPath]; !ok {
			t.Errorf("%s has no row in planVariants — nothing drives the engine or the audit with it altered, so its class is asserted and never tested", f.goPath)
		}
	}
	for path, f := range relayedPlanFields {
		if !live[path] {
			t.Errorf("relayedPlanFields names %s, which record.Plan does not carry", path)
		}
		if f.why == "" {
			t.Errorf("%s is classified %s with no reason", path, f.class)
		}
	}
	if len(live) < 15 {
		t.Fatalf("the walk found %d fields of record.Plan — it is not reaching the nested ones", len(live))
	}
}

// planVariant alters one field of a relayed plan. vary(plan, i) writes the i-th of three distinct
// values, none of them the fixture's own; names is the word a departure on the field carries.
type planVariant struct {
	names string
	vary  func(plan map[string]any, i int)
}

var variantLenses = []string{"red-lens-voice", "red-lens-computation", "red-lens-adversary"}

func entry(plan map[string]any, field string, k int) map[string]any {
	return plan[field].([]any)[k].(map[string]any)
}

// planVariants is keyed as relayedPlanFields is. One table drives both ends of each field's class:
// the engine's valve (is it sensitive to the field) and capture's audit (does it hold the field).
var planVariants = map[string]planVariant{
	"Plan.Head":             {"head", func(p map[string]any, i int) { p["head"] = 100 + i }},
	"Plan.Parties[].SeatID": {"parties", func(p map[string]any, i int) { entry(p, "parties", 0)["seat_id"] = variantLenses[i] }},
	"Plan.Parties[].GapIDs": {"gap_ids", func(p map[string]any, i int) { entry(p, "parties", 0)["gap_ids"] = []any{fmt.Sprintf("V%d", i)} }},
	"Plan.Parties[].Occasions": {"occasions", func(p map[string]any, i int) {
		entry(p, "parties", 2)["occasions"] = [][]any{{"docket", "petition"}, {"petition"}, {"docket"}}[(i+2)%3]
	}},
	"Plan.Docket":                 {"docket", func(p map[string]any, i int) { p["docket"] = []any{fmt.Sprintf("D%d", i)} }},
	"Plan.Exchanges[].GapID":      {"exchanges", func(p map[string]any, i int) { entry(p, "exchanges", 0)["gap_id"] = fmt.Sprintf("X%d", i) }},
	"Plan.Exchanges[].Count":      {"exchanges", func(p map[string]any, i int) { entry(p, "exchanges", 0)["count"] = 50 + i }},
	"Plan.Exchanges[].Unresolved": {"exchanges", func(p map[string]any, i int) { entry(p, "exchanges", 0)["unresolved"] = 50 + i }},
	"Plan.PassPermitted":          {"pass_permitted", func(p map[string]any, i int) { p["pass_permitted"] = i%2 == 1 }},
	"Plan.Ceiling":                {"ceiling", func(p map[string]any, i int) { p["ceiling"] = i%2 == 1 }},
	"Plan.MaxEpochs":              {"max_epochs", func(p map[string]any, i int) { p["max_epochs"] = 90 + i }},
	"Plan.EpochLimitReached":      {"", func(p map[string]any, i int) { p["epoch_limit_reached"] = true }},
	"Plan.Blockers[].Kind": {"blockers", func(p map[string]any, i int) {
		entry(p, "blockers", 0)["kind"] = []string{"contradiction", "reopened", "no_report"}[i]
	}},
	"Plan.Blockers[].Subject": {"blockers", func(p map[string]any, i int) { entry(p, "blockers", 0)["subject"] = fmt.Sprintf("S%d", i) }},
	"Plan.Blockers[].Owner": {"blockers", func(p map[string]any, i int) {
		entry(p, "blockers", 0)["owner"] = []string{"red-chair", "blue-respond", "red-lens-logic"}[i]
	}},
	"Plan.Why": {"", func(p map[string]any, i int) { p["why"] = []any{fmt.Sprintf("a reason reworded %d", i)} }},
	"Plan.StaleAreas[].SeatID": {"", func(p map[string]any, i int) {
		p["stale_areas"] = []any{map[string]any{"seat_id": variantLenses[i], "pin": 1}}
	}},
	"Plan.StaleAreas[].Pin": {"", func(p map[string]any, i int) {
		p["stale_areas"] = []any{map[string]any{"seat_id": "red-lens-voice", "pin": i}}
	}},
	"Plan.RemandOwed[].GapID": {"", func(p map[string]any, i int) {
		p["remand_owed"] = []any{map[string]any{"gap_id": fmt.Sprintf("R%d", i), "direction": "d"}}
	}},
	"Plan.RemandOwed[].Direction": {"", func(p map[string]any, i int) {
		p["remand_owed"] = []any{map[string]any{"gap_id": "R", "direction": fmt.Sprintf("d%d", i)}}
	}},
}

// relayFixture is a record whose first chair sitting dispatches a lens and blue on G1 and the bench
// on G2, which the chair docketed in that sitting, with a petition standing unruled; its second
// chair sitting readies nobody. The bench rules the petition after it, so the record as it stands
// at the end is NOT the record either plan was relayed against. first and last are the plans a
// faithful chair relays.
func relayFixture(t *testing.T) (run record.Run, first, last func() map[string]any) {
	t.Helper()
	n := 0
	ev := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	mint := func(gap string) *record.Event {
		return ev("red-lens-logic", &recordpb.Mint{GapId: proto.String(gap), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)})
	}
	petition := recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION)
	evs := []*record.Event{
		ev("harness", &recordpb.Cast{SeatIds: []string{"red-lens-evidence", "red-lens-logic", "red-chair", "blue-respond", "judge"}}),
		ev("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		ev("red-lens-logic", &recordpb.Register{}), mint("G1"), mint("G2"),
		ev("red-lens-logic", &recordpb.Motion{MotionId: proto.String("M1"), Subject: petition, Basis: proto.String("b"),
			Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{}}}),
		// Chair sitting 1.
		ev("red-chair", &recordpb.Register{}),
		ev("red-chair", &recordpb.Motion{MotionId: proto.String("M2"), Subject: recordpb.MotionSubject_MOTION_SUBJECT_DOCKET.Enum(),
			Basis: proto.String("at impasse"), Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G2")}}}),
		ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-logic"), GapIds: []string{"G1"}}),
		ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G2"}, Occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}}),
		ev("red-lens-logic", &recordpb.Register{}),
		ev("blue-respond", &recordpb.Register{}),
		ev("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}),
		// Chair sitting 2: nobody is readied, and the workflow reads its exit off this plan.
		ev("red-chair", &recordpb.Register{}),
		ev("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_TERMINAL.Enum()}),
		ev("judge", &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: petition, Opinion: proto.String("no"),
			Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED}}),
	}
	dir := t.TempDir()
	recordtest.Seed(t, dir, evs...)
	standing := func() []any {
		return []any{
			map[string]any{"kind": "unruled_motion", "subject": "M1", "owner": "judge"},
			map[string]any{"kind": "unruled_motion", "subject": "M2", "owner": "judge"},
		}
	}
	first = func() map[string]any {
		return map[string]any{"head": 2,
			"parties": []any{
				map[string]any{"seat_id": "red-lens-logic", "gap_ids": []any{"G1"}},
				map[string]any{"seat_id": "blue-respond", "gap_ids": []any{"G1"}},
				map[string]any{"seat_id": "judge", "gap_ids": []any{"G2"}, "occasions": []any{"docket"}},
			},
			"docket": []any{"G2"}, "remand_owed": []any{},
			"exchanges": []any{
				map[string]any{"gap_id": "G1", "count": 0, "unresolved": 0},
				map[string]any{"gap_id": "G2", "count": 0, "unresolved": 0},
			},
			"pass_permitted": false, "ceiling": false, "max_epochs": 0, "epoch_limit_reached": false,
			"why": []any{"a reason"}, "stale_areas": []any{}, "blockers": standing()}
	}
	last = func() map[string]any {
		p := first()
		p["parties"], p["docket"], p["exchanges"] = []any{}, []any{}, []any{}
		return p
	}
	return runtest.Open(t, dir), first, last
}

func auditOf(t *testing.T, run record.Run, plans ...map[string]any) Audit {
	t.Helper()
	var results []map[string]any
	for _, p := range plans {
		results = append(results, map[string]any{"plan": p})
	}
	return DispatchParityAudit(run, journalOf(t, results...), true)
}

// A FAITHFUL RELAY PASSES AT THE POINT IT WAS RELAYED. The fixture's petition is ruled after the
// chair's last sitting; an audit that read the record as it stands now would find both plans
// relaying a blocker the record no longer holds.
func TestAFaithfulRelayIsHeldToTheRecordAtItsOwnSittingsEnd(t *testing.T) {
	run, first, last := relayFixture(t)
	if a := auditOf(t, run, first(), last()); a.Verdict != "PASS" {
		t.Fatalf("a faithful relay = %s: %s", a.Verdict, a.Detail)
	}
	for _, c := range []struct {
		name, want string
		alter      func(first, last map[string]any)
	}{
		{"a docket entry dropped", "sitting 1: relayed docket []", func(f, _ map[string]any) { f["docket"] = []any{} }},
		{"an exchanges entry dropped", "sitting 1: relayed exchanges [G1: count 0, unresolved 0]", func(f, _ map[string]any) { f["exchanges"] = f["exchanges"].([]any)[:1] }},
		{"a blocker dropped on a plan that dispatches", `sitting 1: relayed blockers [unruled_motion "M2" (judge)]`, func(f, _ map[string]any) { f["blockers"] = f["blockers"].([]any)[1:] }},
		{"a blocker dropped on the last plan", `the last plan: relayed blockers [unruled_motion "M1" (judge)]`, func(_, l map[string]any) { l["blockers"] = l["blockers"].([]any)[:1] }},
		{"the epoch limit altered on the last plan", "the last plan: relayed max_epochs 7", func(_, l map[string]any) { l["max_epochs"] = 7 }},
		{"a field absent from the relay", "sitting 1: relayed max_epochs null", func(f, _ map[string]any) { delete(f, "max_epochs") }},
	} {
		t.Run(c.name, func(t *testing.T) {
			f, l := first(), last()
			c.alter(f, l)
			if a := auditOf(t, run, f, l); a.Verdict != "FAIL" || !strings.Contains(a.Detail, c.want) {
				t.Fatalf("%s = %s, want a FAIL naming %q: %s", c.name, a.Verdict, c.want, a.Detail)
			}
		})
	}
}

// drivenEpochs runs debate.js with the chair relaying plans[i] at its (i+1)th sitting and a plan
// readying nobody after, and reports how many chair sittings the debate ran to.
func drivenEpochs(t *testing.T, plans []map[string]any) int {
	t.Helper()
	chair := 0
	backend := func(seatID, label, prompt string) debatejs.Envelope {
		e := debatejs.Envelope{"synopsis": "t", "rulings": []any{}, "dispositions": []any{}, "holdings": []any{},
			"manifest": []any{}, "saturation_reached": false, "open_gaps": 0}
		if seatID == "red-chair" {
			chair++
			e["plan"] = map[string]any{"head": 2, "parties": []any{}, "docket": []any{}, "remand_owed": []any{}, "exchanges": []any{}, "pass_permitted": false,
				"ceiling": false, "max_epochs": 0, "epoch_limit_reached": false, "why": []any{}, "stale_areas": []any{}, "blockers": []any{}}
			if chair <= len(plans) {
				e["plan"] = plans[chair-1]
			}
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
	return out.Epochs
}

// THE VALVE READS ONLY WHAT CAPTURE COMPARES, AND CAPTURE COMPARES WHAT IT SAYS IT DOES. The
// engine's no-progress valve stops a run on the relayed plan alone. A field it is sensitive to that
// capture does not hold to the record is a field a chair can use to keep a stuck run spending, or
// to stop a moving one, unseen.
//
// EACH FIELD'S CLASS IS TESTED FROM BOTH ENDS, with one alteration (planVariants):
//   - the engine is driven through three chair sittings whose plans differ at that field alone. It
//     stops at the third when the valve does not read the field, runs on when it does, and stops at
//     the first when the field ends the debate before the valve is reached;
//   - a faithful relay with the field altered is put to the audit.
//
// A field the valve reads is classified compared; a field classified compared departs in the
// audit, naming the field and the sitting; a field that ends the debate before the valve is
// classified backstopped; and no other field departs. The same holds for each kind of blocker,
// where the two ends agree exactly: the valve keys on a kind when, and only when, a relay is held
// to the record on it — which is what holds debate.js's KEYED_BLOCKER_KINDS to
// record.RelayBlockerKinds.
func TestThePlanKeyReadsOnlyComparedFields(t *testing.T) {
	run, first, last := relayFixture(t)
	const stopsAtOnce, stopsAsStuck, runsOn = 1, 3, 4
	if got := drivenEpochs(t, []map[string]any{first(), first(), first()}); got != stopsAsStuck {
		t.Fatalf("three relays of the fixture's plan ran to %d chair sittings, want %d — the valve does not stop the unaltered plan, so no field below is measured", got, stopsAsStuck)
	}
	ends := func(vary func(map[string]any, int)) (epochs int, a Audit) {
		var plans []map[string]any
		for i := range 3 {
			p := first()
			vary(p, i)
			plans = append(plans, p)
		}
		altered := first()
		vary(altered, 1)
		return drivenEpochs(t, plans), auditOf(t, run, altered, last())
	}
	for _, f := range planFields() {
		v, ok := planVariants[f.goPath]
		if !ok {
			continue // TestEveryRelayedPlanFieldIsClassified names it
		}
		t.Run(f.jsonPath, func(t *testing.T) {
			class := relayedPlanFields[f.goPath].class
			epochs, a := ends(v.vary)
			switch epochs {
			case runsOn:
				if class != relayCompared {
					t.Errorf("the valve reads %s, and it is classified %q — the engine stops a run on a field capture does not hold to the record", f.jsonPath, class)
				}
			case stopsAtOnce:
				if class != relayBackstopped {
					t.Errorf("%s ends the debate before the valve, and it is classified %q", f.jsonPath, class)
				}
			case stopsAsStuck:
			default:
				t.Fatalf("%s altered: the debate ran to %d chair sittings, none of the three outcomes this test reads", f.jsonPath, epochs)
			}
			if class == relayCompared {
				if a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 1") || !strings.Contains(a.Detail, v.names) {
					t.Errorf("%s is classified compared, and altering it in a faithful relay = %s, want a FAIL naming sitting 1 and %q: %s", f.jsonPath, a.Verdict, v.names, a.Detail)
				}
			} else if a.Verdict != "PASS" {
				t.Errorf("%s is classified %q, and altering it in a faithful relay = %s: %s", f.jsonPath, class, a.Verdict, a.Detail)
			}
		})
	}
	held := record.RelayBlockerKinds()
	for _, kind := range record.BlockerKinds() {
		t.Run("blockers of kind "+string(kind), func(t *testing.T) {
			epochs, a := ends(func(p map[string]any, i int) {
				p["blockers"] = append(p["blockers"].([]any), map[string]any{"kind": string(kind), "subject": fmt.Sprintf("K%d", i), "owner": "red-chair"})
			})
			want := stopsAsStuck
			if slices.Contains(held, kind) {
				want = runsOn
			}
			if epochs != want {
				t.Errorf("a %s blocker changing each sitting ran the debate to %d chair sittings, want %d — debate.js's KEYED_BLOCKER_KINDS and record.RelayBlockerKinds disagree on the kind", kind, epochs, want)
			}
			if departs := a.Verdict == "FAIL" && strings.Contains(a.Detail, "sitting 1: relayed blockers"); departs != slices.Contains(held, kind) {
				t.Errorf("a %s blocker added to a faithful relay: departure = %v (%s: %s), and a relay held on the kind = %v", kind, departs, a.Verdict, a.Detail, !departs)
			}
		})
	}
	if len(held) == 0 || len(held) == len(record.BlockerKinds()) {
		t.Fatalf("a relay is held on %d of %d kinds — both arms of the kind check must run", len(held), len(record.BlockerKinds()))
	}
}
