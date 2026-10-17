package capture

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

const (
	firstDirection  = "re-derive the figure from the primary table"
	secondDirection = "count the table's rows by hand"
)

// remandRelayFixture is a record on which the bench remands G1 twice — each in a chair sitting that
// convened it for the docket — and the third chair sitting dispatches G1's minting lens and blue on
// the exchange the second remand grants, each row recording it owed. faithful(k) is the plan a
// chair relaying faithfully relays at chair sitting k (1-based), every field as the record holds it
// at that sitting's end.
func remandRelayFixture(t *testing.T) (run record.Run, faithful func(k int) map[string]any) {
	t.Helper()
	n := 0
	ev := func(seat string, body proto.Message) *record.Event {
		n++
		return recordtest.At(t, seat, fmt.Sprintf("%s:%d", seat, n), body)
	}
	remand := func(motion, direction string) *record.Event {
		return ev("judge", &recordpb.MotionRule{MotionId: proto.String(motion), Opinion: proto.String("not yet settled"),
			Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
				Disposition: recordtest.P(recordpb.Disposition_DISPOSITION_REMANDED), Principle: proto.String("p"), Tension: proto.String("t"),
				ReviewFlag: proto.String("none"), Settled: proto.String("nothing"), Direction: proto.String(direction)}}})
	}
	docket := func(motion string) *record.Event {
		return ev("red-chair", &recordpb.Motion{MotionId: proto.String(motion), Basis: proto.String("at impasse"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}})
	}
	bench := func() *record.Event {
		return ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("judge"), GapIds: []string{"G1"},
			Occasions: []recordpb.Occasion{recordpb.Occasion_OCCASION_DOCKET}})
	}
	evs := []*record.Event{
		ev("harness", &recordpb.Cast{SeatIds: []string{"red-lens-logic", "red-chair", "blue-respond", "judge"}}),
		ev("harness", &recordpb.BaseIngest{Text: proto.String("# r")}),
		ev("red-lens-logic", &recordpb.Register{}),
		ev("red-lens-logic", &recordpb.Mint{GapId: proto.String("G1"), Class: proto.String("x"), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}),
		// Chair sitting 1: the bench is convened on G1, and remands it.
		ev("red-chair", &recordpb.Register{}), docket("M1"), bench(),
		ev("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}), remand("M1", firstDirection),
		// Chair sitting 2: again, on a fresh docket.
		ev("red-chair", &recordpb.Register{}), docket("M2"), bench(),
		ev("judge", &recordpb.Register{Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}), remand("M2", secondDirection),
		// Chair sitting 3: the minting lens and blue, on the exchange the second remand grants.
		ev("red-chair", &recordpb.Register{}),
		ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("red-lens-logic"), GapIds: []string{"G1"}, RemandOwed: []string{"G1"}}),
		ev("red-chair", &recordpb.Dispatch{Pin: proto.Int64(2), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}, RemandOwed: []string{"G1"}}),
		ev("red-lens-logic", &recordpb.Register{}),
		ev("blue-respond", &recordpb.Register{}),
	}
	dir := t.TempDir()
	recordtest.Seed(t, dir, evs...)
	run = runtest.Open(t, dir)
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	facts, err := record.PlanFactsOf(run, fam.Events, fam.At)
	if err != nil {
		t.Fatal(err)
	}
	groups := record.DispatchGroups(fam.Events, fam.At)
	if len(groups) != 3 {
		t.Fatalf("setup: %d dispatch groups, want 3", len(groups))
	}
	faithful = func(k int) map[string]any {
		g := groups[k-1]
		at := facts.At(chairSittingEnd(fam.Events, fam.At, g.Last))
		var parties, exchanges, owed, blockers, docketed []any
		for _, seat := range g.Parties {
			row := g.PartyRows[seat]
			p := map[string]any{"seat_id": seat, "gap_ids": toAny(row.GapIDs)}
			if len(row.Occasions) > 0 {
				p["occasions"] = toAny(row.Occasions)
			}
			parties = append(parties, p)
			for _, gap := range row.GapIDs {
				x := at.ExchangesOn(gap)
				exchanges = append(exchanges, map[string]any{"gap_id": gap, "count": x.Count, "unresolved": x.Unresolved})
			}
			for _, gap := range row.RemandOwed {
				if seat == "blue-respond" {
					d, _ := at.RemandDirection(gap)
					owed = append(owed, map[string]any{"gap_id": gap, "direction": d})
				}
			}
		}
		for _, b := range at.Blockers {
			blockers = append(blockers, map[string]any{"kind": string(b.Kind), "subject": b.Subject, "owner": b.Owner})
		}
		for _, d := range at.Docketed {
			docketed = append(docketed, d)
		}
		return map[string]any{"head": 2, "parties": parties, "docket": nonNil(docketed), "remand_owed": nonNil(owed),
			"exchanges": nonNil(exchanges), "pass_permitted": false, "ceiling": false, "max_epochs": at.MaxEpochs, "epoch": k,
			"final_sitting": false, "epoch_limit_reached": false, "why": []any{}, "stale_areas": []any{}, "blockers": nonNil(blockers)}
	}
	return run, faithful
}

func toAny(xs []string) []any {
	out := []any{}
	for _, x := range xs {
		out = append(out, x)
	}
	return out
}

func nonNil(xs []any) []any {
	if xs == nil {
		return []any{}
	}
	return xs
}

// A RELAYED REMAND IS HELD TO THE ROWS AND TO THE REMAND. Its gap is held to the remand_owed the
// sitting's dispatch rows recorded, and its direction to the latest remand on the gap written by the
// sitting's end — so a chair that drops the exchange, adds one, rewords the direction or quotes the
// remand the bench has since replaced departs, naming the field and the sitting. The last plan,
// which names no party, has no rows: its directions are held all the same.
func TestARelayedRemandIsHeldToTheRowsAndTheRemandsDirection(t *testing.T) {
	run, faithful := remandRelayFixture(t)
	relays := func(alter func(third map[string]any)) []map[string]any {
		third := faithful(3)
		if alter != nil {
			alter(third)
		}
		return []map[string]any{faithful(1), faithful(2), third}
	}
	if owed := faithful(3)["remand_owed"].([]any); len(owed) != 1 || owed[0].(map[string]any)["direction"] != secondDirection {
		t.Fatalf("setup: the third plan owes %v, want G1 on the second remand's direction", owed)
	}
	if a := auditOf(t, run, relays(nil)...); a.Verdict != "PASS" || !strings.Contains(a.Detail, "remand_owed[].gap_id, remand_owed[].direction") {
		t.Fatalf("a faithful relay = %s, want a PASS naming remand_owed among the compared fields: %s", a.Verdict, a.Detail)
	}
	entry := func(gap, direction string) []any {
		return []any{map[string]any{"gap_id": gap, "direction": direction}}
	}
	for _, c := range []struct {
		name  string
		alter func(map[string]any)
		want  string
	}{
		{"the remand dropped", func(p map[string]any) { p["remand_owed"] = []any{} },
			"sitting 3: relayed remand_owed [], its dispatch rows recorded [G1]"},
		{"a remand added on a gap the bench never remanded", func(p map[string]any) {
			p["remand_owed"] = append(p["remand_owed"].([]any), entry("G2", "d")...)
		}, `sitting 3: relayed remand_owed G2 with direction "d", the record at the end of the chair's sitting holds no remand on it`},
		{"the direction reworded", func(p map[string]any) { p["remand_owed"] = entry("G1", "look harder") },
			`sitting 3: relayed remand_owed G1 with direction "look harder", the record at the end of the chair's sitting holds its remand's direction "` + secondDirection + `"`},
		{"the earlier remand's direction", func(p map[string]any) { p["remand_owed"] = entry("G1", firstDirection) },
			`relayed remand_owed G1 with direction "` + firstDirection + `"`},
	} {
		t.Run(c.name, func(t *testing.T) {
			if a := auditOf(t, run, relays(c.alter)...); a.Verdict != "FAIL" || !strings.Contains(a.Detail, c.want) {
				t.Errorf("%s = %s, want a FAIL naming %q: %s", c.name, a.Verdict, c.want, a.Detail)
			}
		})
	}
	t.Run("a remand owed on the first plan, before any remand", func(t *testing.T) {
		first := faithful(1)
		first["remand_owed"] = entry("G1", firstDirection)
		if a := auditOf(t, run, first, faithful(2), faithful(3)); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "sitting 1: relayed remand_owed G1") {
			t.Errorf("= %s, want a FAIL naming sitting 1's remand_owed: %s", a.Verdict, a.Detail)
		}
	})
	t.Run("the last plan's direction", func(t *testing.T) {
		last := faithful(3)
		last["parties"], last["exchanges"] = []any{}, []any{}
		if a := auditOf(t, run, append(relays(nil), last)...); a.Verdict != "PASS" {
			t.Fatalf("a faithful last plan carrying the remand = %s: %s", a.Verdict, a.Detail)
		}
		last["remand_owed"] = entry("G1", firstDirection)
		if a := auditOf(t, run, append(relays(nil), last)...); a.Verdict != "FAIL" || !strings.Contains(a.Detail, "the last plan: relayed remand_owed G1") {
			t.Errorf("the last plan quoting the replaced remand = %s, want a FAIL naming it: %s", a.Verdict, a.Detail)
		}
	})
}
