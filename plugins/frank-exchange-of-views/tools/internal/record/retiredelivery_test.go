package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// retireBoard is a lens past its full re-read that minted G1 on a report holding G1's and C1's
// tokens, at a pin of the report part (event 2).
func retireBoard(t *testing.T) *stage {
	report := "# report\n\n## Findings\n\nThe sky is green " + anchor.Token("G1") + ".\n\n## Method\n\nThe method is sound " + anchor.Token("C1") + ".\n"
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").
		add("harness", &recordpb.ReportPart{Text: proto.String(report)}).
		register("red-chair").readFully(2, evLens).mint(evLens, "G1", "high")
}

func closeRefusal(t *testing.T, run Run, seat, gap string) error {
	t.Helper()
	var refusal error
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		refusal = DeliveredTextRefusal(q, seat, &recordpb.Close{GapId: proto.String(gap)})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return refusal
}

func allFit(string, string) bool { return true }

// A RETIRE AFTER THE DELIVERY WAS CUT CHANGES NO DELIVERED TEXT (gblock 2026-11-20: a retire moves no
// head; the anchor leaves and its text stays). Release fuzz seed 32: a lens's delivery was cut at head
// 409, blue retired its gap's anchor at 410, and the lens's close was refused forever by the
// delivered-text rule, which counted the retire as a change to the gap's sentence. The close stands on
// the delivery it holds.
func TestACloseAfterARetireItsDeliveryPredatesStands(t *testing.T) {
	b := retireBoard(t)
	head := b.head()
	b = b.sitUnserved(head, evLens).served(evLens, head).retire("G1")
	run := b.seed()

	if err := closeRefusal(t, run, evLens, "G1"); err != nil {
		t.Fatalf("a close after a retire of its gap's anchor, on a delivery cut before it, is refused: %v", err)
	}
}

// A DELIVERY CUT AFTER A RETIRE STANDS AT THE HEAD BEFORE IT, and hands no piece for it: the retire
// moved no head and changed no text, so the lens is told the report has not changed since its pin, and
// its close stands on that delivery.
func TestADeliveryCutAfterARetireHandsNoPieceForIt(t *testing.T) {
	b := retireBoard(t)
	head := b.head()
	b = b.retire("G1")
	b = b.sitUnserved(b.head(), evLens)
	run := b.seed()

	cut, err := CutDelivery(run, evLens, nil, allFit)
	if err != nil {
		t.Fatal(err)
	}
	if cut.Head != head {
		t.Fatalf("the delivery was cut at head %d, want the head before the retire, %d", cut.Head, head)
	}
	for _, p := range cut.Pieces {
		if strings.Contains(p.Text, "RETIRED") || strings.Contains(p.Text, "it does not hold") {
			t.Fatalf("the delivery hands the retire as a change: %+v", cut.Pieces)
		}
	}
	if err := RecordDelivery(run, evLens, cut.Head); err != nil {
		t.Fatalf("the delivery cut after the retire is refused: %v", err)
	}
	if err := closeRefusal(t, run, evLens, "G1"); err != nil {
		t.Fatalf("a close on a delivery cut after the retire is refused: %v", err)
	}
}

// A RETIRE BEHIND A RETIRED LENS'S PIN GIVES THE CHAIR NOTHING TO BEAR: no stale area, and a cut of the
// lens's area holds no change.
func TestARetireBehindARetiredLensGivesTheChairNothingToBear(t *testing.T) {
	b := retireBoard(t)
	b = b.sit(b.head(), evLens).sit(b.head(), evLens).sit(b.head(), evLens).retire("C1").register("red-chair")
	run := b.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.StaleAreas) != 0 {
		t.Fatalf("stale_areas = %+v, want none: a retire moves no head", plan.StaleAreas)
	}
	var n int
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		return q.QueryRow(`SELECT COUNT(*) FROM "area_change" WHERE "seat_id" = ?`, evLens).Scan(&n)
	}); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("area_change holds %d changes for %s after a retire alone, want none", n, evLens)
	}
}
