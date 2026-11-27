package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE BENCH'S ORDERS, ACROSS SEATS, THROUGH THE COMMANDS EACH SEAT TYPES: the bench has no grade of
// its own; it sides with blue's regrade proposal, the tool applies the proposal's grade in the
// ruling's transaction, and the assignee may not move the axis the ruling decided.
func TestTheBenchSetsAGradeTheLensMayNotMove(t *testing.T) {
	runDir := seatRun(t)
	g := mintGap(t, runDir, "bench-orders", "spec_underspecification")
	exchangeOn(t, runDir, g)
	sit(t, runDir, "blue-respond")
	motion := motionID(t, runDir, "motion", "docket", "file", "regrade", "--seat-id", "blue-respond", "--id", g,
		"--severity", "low", "--reason", "the defect is cosmetic")
	d := docketsUnaccepted(t, runDir, lensSeat, g)
	if d.GetMotionId() != motion {
		t.Fatalf("the docketing holds %q, want the proposal %s", d.GetMotionId(), motion)
	}
	conveneBench(t, runDir, g)
	if out, err := answerDocketing(t, runDir, d.GetDocketingId(), "for-author"); err != nil {
		t.Fatalf("the bench's ruling for blue's regrade proposal: %v\n%s", err, out)
	}
	if got := lastBody(t, runDir, &recordpb.Regrade{}); got.GetSeverity() != recordpb.Grade_GRADE_LOW {
		t.Fatalf("the ruling applied severity %s, want the proposal's low", recordpb.Word(got.GetSeverity()))
	}
	_, err := lensRegrade(t, runDir, g, "severity", "high")
	if err == nil || !strings.Contains(err.Error(), "which the bench's ruling at event") {
		t.Fatalf("the assignee moved the axis the ruling decided: %v", err)
	}
}
