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
// the model, which is the seat's tier. debate.js writes each at the call site, so a site can name
// one seat in its prompt and dispatch another seat's configuration — `blue-respond` under
// `blue-synthesizer` attests the same role, registers cleanly, and has its sitting attributed to
// `blue-synthesize`.
//
// THE SCRIPT IS RUN, NOT READ. The drive is the one TestEveryFragmentRendersOnceWhereItIsIncluded
// uses, which reaches every agent() call in the script or fails by line; each dispatch is then
// asked of the roster (seatclass.Seats, seatclass.Lane) and of the readers the record already has
// for the tier and the occasion. The two tier models differ here, so a site spreading the wrong
// tier dispatches the wrong model.
func TestEveryDispatchSeatsItsSeatUnderItsOwnConfigurationAndTier(t *testing.T) {
	const lanes = 3
	tierModel := map[string]string{"bulk": "m-bulk", "judgment": "m-judgment"}

	agentPaths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	var agentTypes []string
	for _, p := range agentPaths {
		agentTypes = append(agentTypes, strings.TrimSuffix(filepath.Base(p), ".md"))
	}
	ds, _ := driveEverySite(t, agentTypes, debatejs.Config{Topic: "dispatch identity", Lanes: lanes,
		Model: tierModel["bulk"], JudgmentModel: tierModel["judgment"]})

	laneIDs := record.LaneSeatIDs(lanes)
	occasionWords := record.OccasionWords()
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
		if want, ok := tierModel[class]; !ok || d.Model != want {
			t.Errorf("%s: seat %s is dispatched on model %q; its tier class is %q, whose model in this drive is %q",
				where, d.SeatID, d.Model, class, want)
		}
		if owes := record.SeatOwesOccasion(d.SeatID); owes != (d.Occasion != "") {
			t.Errorf("%s: seat %s is told OCCASION %q; a seat owes one at register: %v", where, d.SeatID, d.Occasion, owes)
		}
		if d.Occasion != "" {
			convened[d.Occasion] = true
			if !slices.Contains(occasionWords, d.Occasion) {
				t.Errorf("%s: the bench is convened on occasion %q, which the record does not carry (%s) — register refuses it",
					where, d.Occasion, strings.Join(occasionWords, ", "))
			}
		}
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
	for _, w := range occasionWords {
		if !convened[w] {
			t.Errorf("the record carries the occasion %q, and no sitting of the drive was convened on it", w)
		}
	}
}
