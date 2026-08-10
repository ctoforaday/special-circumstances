package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE RENDERER MUST SHOW A PRE-COLLAPSE RECORD'S EXCHANGES.
//
// This is the end of the chain the compat fixture exists for: record written by the old binary ->
// dual-read -> one section a human reads. If any link switches to the post-collapse types alone,
// the report renders an empty Motions section for an archived run — indistinguishable from a run
// that had no disputes and no petitions, which is exactly the plausible zero #344 set out to
// remove and would therefore be the worst possible way to fail.
func TestTheReportRendersAPreCollapseRecordsExchanges(t *testing.T) {
	b, err := record.BoardState("../record/testdata/pre-motion-run")
	if err != nil {
		t.Fatalf("the pre-motion fixture must still replay: %v", err)
	}
	out := motions(b)
	if out == "" {
		t.Fatal("the Motions section is EMPTY for a record carrying five retiring events — the exact failure this stage removes")
	}
	for _, want := range []string{
		"grade of",                 // the dispute, with its gap and dimension
		"petition (integrity",      // the filing, with its class
		"direction A1",             // red's ruling on a proposed line
		"ruled rejected",           // the grade answer
		"ruled granted",            // the petition answer
		"ruled out-of-scope",       // the direction answer
		"appealed",                 // contests_ruling, the legacy field with no counterpart
		"relief sought",            // the operative half of the petition
		"id assigned at read time", // legacy ids are marked, never passed off as recorded
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the pre-collapse render is missing %q:\n%s", want, out)
		}
	}
}

// AN UNRULED MOTION IS SAID, not omitted. A filing with no answer means the sitting did not
// happen, and silence there reads identically to a run that never asked.
func TestAnUnruledMotionIsReported(t *testing.T) {
	b := &record.Board{Events: []record.Event{{
		Round: 1, Type: "motion", SeatID: "blue-respond-r1",
		Payload: record.NewPayload().Set("motion_id", "M1").Set("subject", "petition").
			Set("class", "safety").Set("basis", "the demand would bury a hazard"),
	}}}
	out := motions(b)
	if !strings.Contains(out, "NOT RULED") || !strings.Contains(out, "1 motion(s) received no ruling") {
		t.Errorf("an unanswered motion must be named, both on its row and in the tally:\n%s", out)
	}
}
