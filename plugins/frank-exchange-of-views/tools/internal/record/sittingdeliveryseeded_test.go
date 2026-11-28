package record_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/goldentest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EVERY SEEDED ARM OF THE SITTING AND DELIVERY QUESTIONS ANSWERS AS ITS GOLDEN SAYS. The archived
// runs hold no failed sitting, no stop that lands after its seat's next opening, no delivery in no
// sitting, no upheld position, no ruled docket motion and no stamped spot-check change, so the sweep
// over them
// (migrate's TestEveryArchivedRunAnswersItsSittingAndDeliveryQuestionsAsItsGoldenSays) pins nothing
// about those arms. Each record here holds one (record.SeededSittingDeliveryArms) and is rendered
// through the same calls (sittingDeliveryAnswers), under the acts it was seeded with, so a line of the
// golden is checked against the acts above it.
//
// EVERY QUESTION IS ANSWERED BY SOME RECORD HERE: a question whose total is zero fails the test, and
// the totals are logged. No call is refused on any of them.
func TestEverySeededSittingAndDeliveryArmAnswersAsItsGoldenSays(t *testing.T) {
	totals := map[string]int{}
	var out strings.Builder
	for _, arm := range record.SeededSittingDeliveryArms(t) {
		text, err := sittingDeliveryAnswers(arm.Run, totals, true)
		if err != nil {
			t.Fatalf("%s: %v", arm.Name, err)
		}
		fmt.Fprintf(&out, "== %s ==\n%s--\n%s\n", arm.Name, actsOf(t, arm.Run), text)
	}
	for _, q := range sittingDeliveryQuestions {
		t.Logf("%-50s %d", q, totals[q])
		switch {
		case q == qRefused && totals[q] != 0:
			t.Errorf("%d call(s) were refused on the seeded records: each is a line of the golden that says so", totals[q])
		case q != qRefused && totals[q] == 0:
			t.Errorf("no seeded record answers %q: nothing pins it", q)
		}
	}
	goldentest.Assert(t, "seeded_sitting_delivery_answers", out.String())
}

// actsOf is the record's acts, one line each: its id, its word, its seat, the sitting the write
// stored it in, and the fields the sitting and delivery questions read.
func actsOf(t *testing.T, run record.Run) string {
	t.Helper()
	m, err := record.MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, e := range m.Events {
		w := m.At.Of(e)
		sitting := "in no sitting"
		if w.SittingID != 0 {
			sitting = fmt.Sprintf("in sitting %d", w.SittingID)
		}
		fmt.Fprintf(&b, "act %d: %s by %s, %s", w.ID, recordpb.Word(e.GetType()), e.GetSeatId(), sitting)
		switch {
		case e.GetSittingOpen() != nil:
			fmt.Fprintf(&b, "; agent %s opens %s's", e.GetSittingOpen().GetAgentId(), e.GetSittingOpen().GetSeatId())
		case e.GetSittingClose() != nil:
			fmt.Fprintf(&b, "; agent %s", e.GetSittingClose().GetAgentId())
		case e.GetEnvelopeAdmitted() != nil:
			fmt.Fprintf(&b, "; agent %s", e.GetEnvelopeAdmitted().GetAgentId())
		case e.GetRegister() != nil && e.GetRegister().GetAgentId() != "":
			fmt.Fprintf(&b, "; agent %s", e.GetRegister().GetAgentId())
		case e.GetDispatch() != nil:
			fmt.Fprintf(&b, "; %s at pin %d on %v", e.GetDispatch().GetSeatId(), e.GetDispatch().GetPin(), e.GetDispatch().GetGapIds())
		case e.GetReportRead() != nil:
			fmt.Fprintf(&b, "; head %d", e.GetReportRead().GetHead())
		case e.GetMint() != nil:
			fmt.Fprintf(&b, "; %s", e.GetMint().GetGapId())
		case e.GetMotion() != nil:
			fmt.Fprintf(&b, "; %s", e.GetMotion().GetMotionId())
		case e.GetPosition() != nil:
			fmt.Fprintf(&b, "; %s", e.GetPosition().GetPositionId())
		case e.GetMotionRule() != nil:
			fmt.Fprintf(&b, "; %s", e.GetMotionRule().GetMotionId())
			if p := e.GetMotionRule().GetPosition(); p != nil {
				fmt.Fprintf(&b, " %s, naming %q", recordpb.Word(p.GetRuling()), p.GetLens())
			}
		case e.GetSpotCheck() != nil:
			fmt.Fprintf(&b, "; areas %v, pins %v, changes %v", e.GetSpotCheck().GetAreas(), e.GetSpotCheck().GetPins(), e.GetSpotCheck().GetChanges())
		}
		b.WriteString("\n")
	}
	return b.String()
}
