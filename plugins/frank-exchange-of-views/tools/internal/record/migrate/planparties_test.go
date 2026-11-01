package migrate_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// planWithParties is a stored plan's printing, with the parties given as JSON.
func planWithParties(t *testing.T, parties string) string {
	t.Helper()
	p := record.Plan{Head: 2, Why: []string{"the bench is convened for a petition"}}
	raw := p.WithKey().Printed()
	var doc map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatal(err)
	}
	doc["parties"] = json.RawMessage(parties)
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// `occasions` LEAVES A STORED PLAN, AND THE PLAN STILL READS. A docket party keeps its gaps; a bench
// party with no gap was convened for a hearing alone, which is on no plan now, and leaves. The plan is
// re-keyed over what it now says, and the prose it printed keeps its words. A plan carrying none is
// left byte for byte.
func TestTranslatePlanPartiesTakesOccasionsOffAStoredPlan(t *testing.T) {
	pl := &recordpb.Plan{Printed: proto.String(planWithParties(t, `[
		{"seat_id":"blue-respond","gap_ids":["G-00000001"]},
		{"seat_id":"judge","gap_ids":["G-00000002"],"occasions":["docket"]},
		{"seat_id":"judge","gap_ids":[],"occasions":["petition"]}]`))}
	if err := migrate.TranslatePlanParties(pl); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(pl.GetPrinted(), `"occasions"`) {
		t.Errorf("the translated plan still prints occasions: %s", pl.GetPrinted())
	}
	p, _, err := record.ReadPlan([]byte(pl.GetPrinted()))
	if err != nil {
		t.Fatalf("the translated plan does not read: %v", err)
	}
	if len(p.Parties) != 2 || p.Parties[0].SeatID != "blue-respond" || p.Parties[1].SeatID != "judge" || p.Parties[1].GapIDs[0] != "G-00000002" {
		t.Errorf("parties = %+v, want blue and the docket party, and no hearing", p.Parties)
	}
	if want := p.WithKey().Key; p.Key != want {
		t.Errorf("the plan's key is %q and its own is %q", p.Key, want)
	}
	if len(p.Why) != 1 || p.Why[0] != "the bench is convened for a petition" {
		t.Errorf("the plan's prose moved: %v", p.Why)
	}

	bare := planWithParties(t, `[{"seat_id":"blue-respond","gap_ids":["G-00000001"]}]`)
	untouched := &recordpb.Plan{Printed: proto.String(bare)}
	if err := migrate.TranslatePlanParties(untouched); err != nil {
		t.Fatal(err)
	}
	if untouched.GetPrinted() != bare {
		t.Errorf("a plan carrying no occasions was rewritten:\nfrom %s\nto   %s", bare, untouched.GetPrinted())
	}

	for _, bad := range []string{`not json`, `{"parties":"none"}`, `{"parties":[{"seat_id":7,"occasions":[]}]}`, `{"parties":[{"seat_id":"judge","gap_ids":"x","occasions":[]}]}`} {
		if err := migrate.TranslatePlanParties(&recordpb.Plan{Printed: proto.String(bad)}); err == nil {
			t.Errorf("%s translated", bad)
		}
	}
}

// THE ARCHIVES' STORED PLANS LOSE THE KEY AND NOTHING ELSE THE TRANSLATION DOES NOT NAME. Their bench
// parties carry occasions, every one of them crosses, and the census counts both: an archive set in
// which no stored plan carried one tests this by nothing.
func TestArchivedPlansCrossWithNoOccasions(t *testing.T) {
	const name = "2026-10-18_is-91-prime-m28.tar.gz"
	res, _ := migrateArchive(t, name)
	if len(res.Refusals) != 0 {
		t.Fatalf("%d refusal(s); the first: %+v", len(res.Refusals), res.Refusals[0])
	}
	srcPlans, dstPlans, _, _, _ := planPairsOf(t, name)
	carried := 0
	for i := range srcPlans {
		carried += strings.Count(srcPlans[i].printed, `"occasions"`)
		if strings.Contains(dstPlans[i].printed, `"occasions"`) {
			t.Errorf("migrated plan %d still prints occasions: %s", i, dstPlans[i].printed)
		}
		if _, _, err := record.ReadPlan([]byte(dstPlans[i].printed)); err != nil {
			t.Errorf("migrated plan %d does not read: %v", i, err)
		}
	}
	if carried == 0 {
		t.Fatalf("fixture: %s's stored plans print no occasions, so the translation is tested by nothing", name)
	}
}

// A DISPATCH ROW LOSES ITS OCCASIONS, AND A BENCH ROW WITH NO GAP IS NO ROW. A docket row is kept with
// its gaps; a bench row that convened a hearing alone translates to no event, because the engine
// convenes a hearing off a close's read. Any other seat's row with no gap is the row it always was.
func TestADispatchRowCrossesWithoutItsOccasions(t *testing.T) {
	row := func(seat string, gaps, occasions []string) migrate.OldEvent {
		return migrate.OldEvent{ID: 9, SeatID: "red-chair", TS: ts1, Word: "dispatch",
			Fields: map[string]any{"seat_id": seat, "pin": int64(2)},
			Lists:  map[string][]string{"gap_ids": gaps, "occasions": occasions}}
	}
	for _, tc := range []struct {
		name string
		old  migrate.OldEvent
		gaps []string
	}{
		{"a docket", row("judge", []string{"G-00000001"}, []string{"docket"}), []string{"G-00000001"}},
		{"a lens with no gap", row("red-lens-evidence", nil, nil), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bodies, err := migrate.Entries().Translate(tc.old)
			if err != nil || len(bodies) != 1 {
				t.Fatalf("%d bodies (%v), want the row", len(bodies), err)
			}
			d := bodies[0].(*recordpb.Dispatch)
			if d.GetSeatId() != tc.old.Fields["seat_id"] || !slices.Equal(d.GetGapIds(), tc.gaps) {
				t.Errorf("the row = %+v", d)
			}
		})
	}
	bodies, err := migrate.Entries().Translate(row("judge", nil, []string{"petition"}))
	if err != nil || len(bodies) != 0 {
		t.Errorf("a hearing's row = %d bodies (%v), want no event", len(bodies), err)
	}
}
