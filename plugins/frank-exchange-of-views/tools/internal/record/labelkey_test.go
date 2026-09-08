package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A label-keyed act dedups WITHIN a sitting and repeats ACROSS them. The seat id used to carry
// the round and so scoped the key; roundless the key carries the sitting itself, or a lens could
// verify each citation once per run (plans/roundless.md §III.A.1; found by migrating the
// quadratic-formula archive, 53 refusals).
func TestALabelKeyedActRepeatsAcrossSittingsButNotWithinOne(t *testing.T) {
	runDir := newRun(t)
	lens := Identity{Run: mustRun(t, runDir), SeatID: "red-lens-evidence"}
	if _, _, err := RegisterSeat(lens, ""); err != nil {
		t.Fatal(err)
	}
	verify := func() *recordpb.Verify {
		return &recordpb.Verify{Anchor: proto.String("s-1"), Claim: proto.String("c"), Text: proto.String("what the source says"),
			Confidence: recordtest.P(recordpb.Confidence_CONFIDENCE_HIGH),
			Outcome:    recordtest.P(recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS)}
	}
	first, err := Append(lens, verify())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(first.GetKey(), ":#1:s-1") {
		t.Errorf("key = %q, want it to carry sitting 1 and the anchor", first.GetKey())
	}
	if _, err := Append(lens, verify()); err == nil {
		t.Fatal("the same verification twice in one sitting was written twice — a crash-retry would double-count")
	} else if !strings.Contains(err.Error(), `"s-1"`) {
		t.Errorf("the refusal must name what collided: %v", err)
	}

	// The lens sits again and re-verifies the same anchor: new evidence, new act.
	if _, _, err := RegisterSeat(lens, ""); err != nil {
		t.Fatal(err)
	}
	second, err := Append(lens, verify())
	if err != nil {
		t.Fatalf("re-verifying a citation in a later sitting was refused: %v", err)
	}
	if !strings.HasSuffix(second.GetKey(), ":#2:s-1") {
		t.Errorf("key = %q, want sitting 2", second.GetKey())
	}
}
