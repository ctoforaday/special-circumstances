package fuzz

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE SPOT-CHECK STATES WHAT CHANGED BEHIND EACH PIN, AND THE TOOL STATES IT (gblock, 2026-11-10).
// m40's chair wrote "Changes since 260: 42 edits" over a record holding 4. The write now stamps each
// named area's pin and the event id of every change to the report after the earliest of them, and the
// chair's prose states neither.
//
// The oracle reads each spot-check naming areas at the run's end: each area's stamped pin is the pin
// the chair's last plan before it listed for that stale area — the drive names areas only from that
// plan — and its changes are every blue_edit and report_part on the record between the earliest pin and
// the spot-check's own event. Its tally gate: spot-checks naming areas were recorded, every one carries
// a stamp the record agrees with, and some stamp names changes.

// spotCheckStamps is the stamp arm's count over one run's record.
type spotCheckStamps struct {
	// named is the spot-checks naming a stale area, stamped those whose stamp the record agrees with,
	// and withChanges the stamped ones naming at least one change.
	named, stamped, withChanges int
}

// checkSpotCheckStamps is the stamp arm, read at the run's end. It returns the counts, and why the
// first spot-check whose stamp the record disagrees with fails the run.
func checkSpotCheckStamps(fam record.Family) (got spotCheckStamps, msg string) {
	planPins := map[string]int64{}
	var changes []int64
	for _, e := range fam.Events {
		id := fam.At.Of(e).ID
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_BLUE_EDIT, recordpb.EventType_EVENT_TYPE_REPORT_PART:
			changes = append(changes, id)
			continue
		}
		if p, ok := recordpb.BodyAs[*recordpb.Plan](e); ok {
			var plan struct {
				StaleAreas []struct {
					SeatID string `json:"seat_id"`
					Pin    int64  `json:"pin"`
				} `json:"stale_areas"`
			}
			if err := json.Unmarshal([]byte(p.GetPrinted()), &plan); err != nil {
				return got, fmt.Sprintf("spot-check stamp: the plan at event %d does not parse: %v", id, err)
			}
			planPins = map[string]int64{}
			for _, a := range plan.StaleAreas {
				planPins[a.SeatID] = a.Pin
			}
			continue
		}
		sc, ok := recordpb.BodyAs[*recordpb.SpotCheck](e)
		if !ok || len(sc.GetAreas()) == 0 {
			continue
		}
		got.named++
		if bad := stampDisagrees(sc, planPins, changes); bad != "" {
			if msg == "" {
				msg = fmt.Sprintf("spot-check stamp: the spot-check at event %d %s", id, bad)
			}
			continue
		}
		got.stamped++
		if len(sc.GetChanges()) > 0 {
			got.withChanges++
		}
	}
	return got, msg
}

// stampDisagrees is how a spot-check's stamp disagrees with the plan before it and the changes on the
// record before it, or "" where it agrees.
func stampDisagrees(sc *recordpb.SpotCheck, planPins map[string]int64, changes []int64) string {
	areas, pins := sc.GetAreas(), sc.GetPins()
	if len(pins) != len(areas) {
		return fmt.Sprintf("names %d area(s) and carries %d pin(s)", len(areas), len(pins))
	}
	earliest := pins[0]
	for i, a := range areas {
		want, listed := planPins[a]
		if !listed || want != pins[i] {
			return fmt.Sprintf("stamps %s at pin %d, and the plan before it lists it at %d (listed: %v)", a, pins[i], want, listed)
		}
		earliest = min(earliest, pins[i])
	}
	var want []int64
	for _, c := range changes {
		if c > earliest {
			want = append(want, c)
		}
	}
	if !slices.Equal(sc.GetChanges(), want) {
		return fmt.Sprintf("stamps changes %v after pin %d, and the record holds %v", sc.GetChanges(), earliest, want)
	}
	return ""
}
