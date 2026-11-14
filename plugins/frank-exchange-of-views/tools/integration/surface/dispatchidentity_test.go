package surface

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/debatejs"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EVERY DISPATCH SEATS THE SEAT ITS PROMPT NAMES, UNDER THAT SEAT'S OWN CONFIGURATION AND TIER.
//
// A dispatch states its identity three times: the SEAT_ID in the prompt, which the seat types at
// register; the agent type, which the hook attests and the record reads the seat back from; and
// the model, which is the seat's tier. The configuration and tier of every seat are debate.js's
// generated seat block, written from the roster (devcmd/enginegen), so neither can be misspelled.
// What a call site authors is WHICH seat, and it says so twice — once to the prompt and once
// to `dispatchAs` — so a site can name one seat in its prompt and dispatch another's row:
// `blue-respond` under `blue-synthesizer` attests the same role, registers cleanly, and has its
// sitting attributed to `blue-synthesize`. A site also authors which occasion it convenes the
// bench on, and whether a seat has a site at all.
//
// THE SCRIPT IS RUN, NOT READ. The drive is the one TestEveryFragmentRendersOnceWhereItIsIncluded
// uses, which reaches every agent() call in the script or fails by line; each dispatch is then
// asked of the roster (seatclass.Seats, seatclass.Lane) and of the readers the record already has
// for the tier and the occasion. The two tier models differ here, so a site spreading the wrong
// tier dispatches the wrong model.
func TestEveryDispatchSeatsItsSeatUnderItsOwnConfigurationAndTier(t *testing.T) {
	const lanes = 3
	tierModel := map[string]string{"bulk": "m-bulk", "judgment": "m-judgment"}

	seats := debatejs.TieredSeats(tierModel["bulk"], tierModel["judgment"])
	override := seats["red-lens-dark-side"]
	override.Model = "m-lens-override"
	seats["red-lens-dark-side"] = override

	agentPaths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	var agentTypes []string
	for _, p := range agentPaths {
		agentTypes = append(agentTypes, strings.TrimSuffix(filepath.Base(p), ".md"))
	}
	ds, _ := driveEverySite(t, agentTypes, debatejs.Config{Topic: "dispatch identity", Lanes: lanes,
		Seats: seats})

	laneIDs := record.LaneSeatIDs(lanes)
	sat := map[string]bool{}
	convened := map[string]bool{}
	for _, d := range ds {
		where := fmt.Sprintf("debate.js:%d (%s)", d.Site, d.Label)
		row, onRoster := seatclass.Seats[d.SeatID]
		if slices.Contains(laneIDs, d.SeatID) {
			row, onRoster = seatclass.Lane, true
		}
		if !onRoster || row.AgentType == "" {
			t.Errorf("%s: the prompt states SEAT_ID %q, which is no seat the roster dispatches — the seat is refused at register, its first act", where, d.SeatID)
			continue
		}
		sat[d.SeatID] = true
		if d.AgentType != row.AgentType {
			t.Errorf("%s: seat %s is dispatched under %q; the roster seats it under %q — the hook attests another seat's configuration",
				where, d.SeatID, d.AgentType, row.AgentType)
		}
		class := seatclass.ClassOf(row.Base)
		want := tierModel[class]
		seatKey := d.SeatID
		if slices.Contains(laneIDs, d.SeatID) {
			seatKey = seatclass.Lane.Base
		}
		if spec, ok := seats[seatKey]; ok && spec.Model != "" {
			want = spec.Model
		}
		if want == "" || d.Model != want {
			t.Errorf("%s: seat %s is dispatched on model %q; want %q (tier class %q)",
				where, d.SeatID, d.Model, want, class)
		}
		if owes := record.SeatOwesOccasion(d.SeatID); owes != (d.Occasion != "") {
			t.Errorf("%s: seat %s is told OCCASION %q; a seat owes one at register: %v", where, d.SeatID, d.Occasion, owes)
		}
		convened[d.Occasion] = true
	}

	// AND EVERY SEAT AND OCCASION THE RECORD ADMITS IS ONE THE ENGINE REACHES. A roster seat no
	// dispatch seats is an id register admits and no run produces; an occasion no sitting is
	// convened on is an arm every reader of it carries for nothing.
	for id, s := range seatclass.Seats {
		if s.AgentType != "" && !sat[id] {
			t.Errorf("the roster seats %s under %s, and the drive dispatched no such seat", id, s.AgentType)
		}
	}
	for _, id := range laneIDs {
		if !sat[id] {
			t.Errorf("a run of %d lanes seats %s, and the drive dispatched no such seat", lanes, id)
		}
	}
	for _, w := range record.OccasionWords() {
		if !convened[w] {
			t.Errorf("the record carries the occasion %q, and no sitting of the drive was convened on it", w)
		}
	}
}
