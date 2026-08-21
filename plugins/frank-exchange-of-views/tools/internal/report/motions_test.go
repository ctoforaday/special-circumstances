package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// AN UNRULED MOTION IS SAID, not omitted. A filing with no answer means the sitting did not
// happen, and silence there reads identically to a run that never asked.
func TestAnUnruledMotionIsReported(t *testing.T) {
	b := &record.Board{Events: []*record.Event{{
		Round: 1, Type: "motion", SeatID: "blue-respond-r1",
		Payload: record.NewPayload().Set("motion_id", "M1").Set("subject", "petition").
			Set("class", "safety").Set("reason", "the demand would bury a hazard"),
	}}}
	out := motions(b)
	if !strings.Contains(out, "NOT RULED") || !strings.Contains(out, "1 motion(s) received no ruling") {
		t.Errorf("an unanswered motion must be named, both on its row and in the tally:\n%s", out)
	}
}
