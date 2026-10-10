package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A POSITION IS REFUSED FOR A SEAT THAT OWES NONE. The blue role is one surface, so a lane, the
// frontier and the synthesizer hold the verb; record.SeatOwesPosition is what says whose act it is,
// and the refusal names the seats it is true of — read from the predicate, so the seats a refused
// seat is pointed at are the seats the write path admits.
//
// Seventeen archived runs hold 48 positions from these three seat classes: argument filed where
// only the bench and the chair can read it, by seats nothing held to filing it.
func TestAPositionIsRefusedForASeatThatOwesNone(t *testing.T) {
	for _, seatID := range []string{"blue-lane-1", "blue-lane-3", "frontier", "blue-synthesize"} {
		t.Run(seatID, func(t *testing.T) {
			if record.SeatOwesPosition(seatID) {
				t.Fatalf("%s owes a position by the predicate: this test's premise is gone", seatID)
			}
			runDir := newRun(t)
			_, err := run(t, "position", "--run", runDir, "--seat-id", seatID, "--reason", "the draft covers the lane")
			if err == nil {
				t.Fatalf("%s filed a position: the verb admits a seat that owes none", seatID)
			}
			for _, owes := range []string{"blue-respond", "red-chair"} {
				if !strings.Contains(err.Error(), owes) {
					t.Errorf("the refusal does not name %s, a seat whose act it is:\n%v", owes, err)
				}
			}
			if !strings.Contains(err.Error(), seatID) {
				t.Errorf("the refusal does not name the seat it refuses (%s):\n%v", seatID, err)
			}
			// THE PAGE THE SEAT READS FIRST SAYS THE SAME: the blue role's one help page names the
			// seat whose act it is, so the refusal is not where a lane learns it.
			help, herr := run(t, "position", "--help", "--seat-id", seatID)
			if herr != nil || !strings.Contains(help, "blue-respond") || !strings.Contains(help, "refused") {
				t.Errorf("%s's help for the verb does not say whose act it is and that it is refused (err %v):\n%s", seatID, herr, help)
			}
			fam, ferr := record.FamilyOf(runtest.Open(t, runDir))
			if ferr != nil {
				t.Fatal(ferr)
			}
			for _, e := range fam.Events {
				if e.GetType() == recordpb.EventType_EVENT_TYPE_POSITION {
					t.Errorf("the refused position is on the record: %v", e)
				}
			}
		})
	}
	for _, seatID := range []string{"blue-respond", "red-chair"} {
		t.Run(seatID, func(t *testing.T) {
			if !record.SeatOwesPosition(seatID) {
				t.Fatalf("%s owes no position by the predicate", seatID)
			}
			if _, err := run(t, "position", "--run", newRun(t), "--seat-id", seatID, "--reason", "what the bench is asked to weigh"); err != nil {
				t.Fatalf("%s owes a position and was refused one: %v", seatID, err)
			}
		})
	}
}
