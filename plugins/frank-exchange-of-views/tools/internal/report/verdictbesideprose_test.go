package report

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE PROSE CLAIMS ONE VERDICT AND THE RECORD HOLDS ANOTHER, so the projection must show the fact
// beside the claim. Measured in research/2026-09-02_quadratic-formula: red's round-5 position says
// "My verdict is PASS", `round_verdict` holds fail for all five rounds, and debate.md rendered only
// the sentence — so a reader of that document alone concluded the opposite of the record.
//
// The fix is deliberately NOT contradiction-detection. Matching "PASS" in prose would be a string
// test on a seat's argument, and a wrong one is worse than none. Render the recorded verdict; let
// the reader see both.
func TestTheRecordedVerdictRendersBesideRedsProse(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Gate{
			Verdict: recordtest.P(recordpb.Verdict_VERDICT_FAIL),
		}),
		recordtest.Event(t, "red-chair", &recordpb.Position{
			Text: proto.String("Nothing on the board is open. My verdict is PASS."),
		}),
	}
	got := debate((record.NewFamily(nil, evs)), evs)
	if !strings.Contains(got, "recorded verdict: FAIL") {
		t.Errorf("the epoch's recorded verdict is not rendered beside the prose that claims one:\n%s", got)
	}
	// Both must be present: suppressing the prose would be the opposite error, hiding a seat's
	// argument because the tool disagrees with it.
	if !strings.Contains(got, "My verdict is PASS") {
		t.Errorf("red's own words were dropped — the projection must show both:\n%s", got)
	}
}

// AN ABSENT VERDICT IS ITS OWN FACT, and it must not render as an epoch that simply had none worth
// mentioning. An epoch where red spoke and no verdict was recorded is a gate that never closed —
// which is exactly the ceiling case that produced this defect.
func TestAnEpochWithNoRecordedVerdictSaysSo(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "red-chair", &recordpb.Position{Text: proto.String("gap A stands")}),
	}
	got := debate((record.NewFamily(nil, evs)), evs)
	if !strings.Contains(got, "NO VERDICT RECORDED") {
		t.Errorf("an epoch with no recorded verdict renders indistinguishably from one that had a verdict:\n%s", got)
	}
}
