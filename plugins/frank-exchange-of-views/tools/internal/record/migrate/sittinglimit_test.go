package migrate

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A SITTING LIMIT TRANSLATES TO NO EVENT (EventSchema 74): whatever its era's columns, it arrives as
// nothing and is not refused. A seat stopped at its agent definition's maxTurns is a failed sitting,
// which the record derives from the sitting's ends.
func TestASittingLimitTranslatesToNoEvent(t *testing.T) {
	for _, fields := range []map[string]any{
		{"agent_id": "a1", "agent_type": "frank-exchange-of-views:red-chair", "seat_id": "red-chair", "sitting": int64(2), "limit": int64(150)},
		{"seat_id": "red-chair", "sitting": int64(2), "limit": int64(150)},
	} {
		bodies, err := Entries().Translate(OldEvent{ID: 5, SeatID: "harness", Word: "sitting_limit", Fields: fields})
		if err != nil {
			t.Fatalf("a sitting limit was refused: %v", err)
		}
		if len(bodies) != 0 {
			t.Errorf("a sitting limit translated to %v, want no event", bodies)
		}
	}
}

// THE RESIDUE IS NONE: no archived run holds a sitting limit, so dropping the translation's output
// loses nothing an archive recorded. The census reads every archive's source, so an archive that
// holds one fails here and the residue has to be stated.
func TestNoArchivedRunHoldsASittingLimit(t *testing.T) {
	for _, name := range archivedRuns(t) {
		t.Run(name, func(t *testing.T) {
			evs, err := archivedSource(t, recordtest.ExtractArchive(t, name)).Events()
			if err != nil {
				t.Fatal(err)
			}
			for _, old := range evs {
				if old.Word == "sitting_limit" {
					t.Errorf("source event %d is a sitting limit, which migrates to nothing — state the residue", old.ID)
				}
			}
		})
	}
}
