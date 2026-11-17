package cli

import (
	"strings"
	"testing"
)

// THE BENCH'S ORDERS, ACROSS SEATS, THROUGH THE COMMANDS EACH SEAT TYPES (gblock, 2026-10-30): the
// bench regrades a gap it is hearing and the minting lens may not move that axis; the bench rules the
// gap closed.
func TestTheBenchSetsAGradeTheLensMayNotMove(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "bench-orders", "spec_underspecification")
	motion := docketFile(t, runDir, "red-chair", g, "contested, and the bench's to settle")

	if _, err := runAt(t, "regrade", "--run", runDir, "--seat-id", "judge", "--id", g, "--severity", "high",
		"--reason", "the defect misleads every reader of the section"); err != nil {
		t.Fatalf("the bench's regrade of a gap it is hearing: %v", err)
	}
	_, err := runAt(t, "regrade", "--run", runDir, "--seat-id", lensSeat, "--id", g, "--severity", "low",
		"--reason", "the lens disagrees")
	if err == nil || !strings.Contains(err.Error(), "which the bench set at event") {
		t.Fatalf("the minting lens moved the axis the bench set: %v", err)
	}

	args := benchRuleArgs(motion, "defect_accepted", "proportionality")
	args = append([]string{args[0], args[1], args[2], "--run", runDir, "--seat-id", "judge"}, args[3:]...)
	if _, err := runAt(t, args...); err != nil {
		t.Fatalf("motion docket rule: %v", err)
	}
}
