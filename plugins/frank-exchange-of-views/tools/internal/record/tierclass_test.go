package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A seat's tier class is the join of the run's cast and the tier-class table (TierClassOfSeat). The
// served model is compared with the tier its class was configured for by capture's tier audit and
// `show tiers`; a seat whose join misses would never be compared at all.

func runWithTiers(t *testing.T) string {
	t.Helper()
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "inputs", "run-config.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// A lane is tier-bound because the CAST says it is a lane; its id carries no such fact.
	seats, laneSeats := CastFor(1)
	recordtest.Seed(t, run, recordtest.At(t, HarnessSeat, "harness:cast:1",
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats}))
	return run
}

// The operator is not a debating seat and rides no tier.
func TestTheOperatorRidesNoTier(t *testing.T) {
	if got := tierClassOfSeat(t, mustRun(t, runWithTiers(t)), OperatorRole); got != "" {
		t.Fatalf("the operator has no tier class, got %q", got)
	}
}

// The roster and the tier-class table are two lists, and TierClassOfSeat is the join. A seat
// whose base is not a key in seatclass would silently make that whole seat class ungated.
func TestEverySeatShapeJoinsToATierClass(t *testing.T) {
	r := mustRun(t, runWithTiers(t))
	for id, s := range seatclass.Seats {
		if s.Base == "" {
			continue // the operator, deliberately
		}
		if got := tierClassOfSeat(t, r, id); got != "bulk" && got != "judgment" {
			t.Errorf("seat %s (base %q) has no tier class — that class of seat would never be gated", id, s.Base)
		}
	}
	// A lane is not on the roster — its id cannot be enumerated (#1153b) — so it is the one seat
	// whose tier join would go unchecked by the loop above.
	if got := tierClassOfSeat(t, r, SampleSeatOf("blue")); got != "bulk" {
		t.Errorf("a lane seat has tier class %q, want bulk — lanes would never be gated", got)
	}
	// A PETITION SITTING IS NO LONGER ITS OWN SEAT, so there is nothing handled apart from the
	// table any more: it is the bench answering a different question, and the loop above already
	// covers `judge`. An id derived from a seat is not a seat.
	if got := tierClassOfSeat(t, r, "judge-petition-red-chair"); got != "" {
		t.Errorf("a derived petition id resolves to tier %q — it is not a seat the engine dispatches", got)
	}
}

// tierClassOfSeat is TierClassOfSeat with its read error failing the test rather than reading as "".
func tierClassOfSeat(t *testing.T, run Run, seatID string) string {
	t.Helper()
	c, err := TierClassOfSeat(run, seatID)
	if err != nil {
		t.Fatalf("TierClassOfSeat(%q): %v", seatID, err)
	}
	return c
}
