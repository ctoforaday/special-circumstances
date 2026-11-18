package record

import (
	"errors"
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

// A DELIVERY CUT AFTER A RETIRE STANDS AT THE RETIRE (gblock 2026-11-18: a retire moves the head).
// Release fuzz seed 32: a lens's delivery was cut after blue retired its gap's anchor, stamped with the
// head before the retire, and its close was refused forever by the delivered-text rule, which counts
// the retire as a change after that head. Cut at the retire, the delivery holds it and the close stands.
func TestACloseAfterARetireAndADeliveryCutAfterItStands(t *testing.T) {
	b := retireBoard(t).retire("G1")
	retire := int64(b.n)
	b = b.sitUnserved(b.head(), evLens)
	run := b.seed()

	cut, err := CutDelivery(run, evLens, nil, allFit)
	if err != nil {
		t.Fatal(err)
	}
	if cut.Head != retire {
		t.Fatalf("the delivery was cut at head %d, want the retire at %d", cut.Head, retire)
	}
	var shown bool
	for _, p := range cut.Pieces {
		if strings.Contains(p.Text, "RETIRED, WITH G1:\nthe claim G1 backs") && strings.Contains(p.Text, "REASON:\nit does not hold") {
			shown = true
		}
	}
	if !shown {
		t.Fatalf("the retire since the pin is not delivered as the claim it retired and its reason: %+v", cut.Pieces)
	}
	if err := RecordDelivery(run, evLens, cut.Head); err != nil {
		t.Fatalf("the delivery cut at the retire is refused: %v", err)
	}
	if err := closeRefusal(t, run, evLens, "G1"); err != nil {
		t.Fatalf("a close on a delivery cut after the retire is refused: %v", err)
	}
}

// A RETIRE AFTER THE DELIVERY WAS CUT IS A CHANGE THE SEAT HAS NOT READ: the close is refused by the
// delivered-text rule until a later delivery holds it.
func TestACloseAfterARetireTheDeliveryWasCutBeforeIsRefused(t *testing.T) {
	b := retireBoard(t)
	b = b.sitUnserved(b.head(), evLens).served(evLens, b.head()).retire("G1")
	run := b.seed()

	var undelivered *UndeliveredError
	if err := closeRefusal(t, run, evLens, "G1"); !errors.As(err, &undelivered) {
		t.Fatalf("a close after a retire of its gap's anchor that its delivery predates is not refused as undelivered: %v", err)
	}
}

// A RETIRE BEHIND A RETIRED LENS'S PIN IS DELIVERED TO THE CHAIR: the stale area's change is the
// retire, the chair's cut holds it — a cut of nothing would end its turn — and its spot-check of the
// area stands on that delivery.
func TestARetireBehindAStaleAreaIsDeliveredToTheChair(t *testing.T) {
	b := retireBoard(t)
	b = b.sit(b.head(), evLens).sit(b.head(), evLens).sit(b.head(), evLens).retire("C1").register("red-chair")
	run := b.seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.StaleAreas) != 1 || plan.StaleAreas[0].SeatID != evLens {
		t.Fatalf("stale_areas = %+v, want %s behind the retire", plan.StaleAreas, evLens)
	}
	cut, err := CutDelivery(run, "red-chair", []string{evLens}, allFit)
	if err != nil {
		t.Fatal(err)
	}
	if len(cut.Pieces) != 1 || !strings.Contains(cut.Pieces[0].Text, "RETIRED") || !strings.Contains(cut.Pieces[0].Text, "C1") {
		t.Fatalf("the chair's cut of %s's area is not the retire behind its pin: %+v", evLens, cut.Pieces)
	}
	if err := RecordDelivery(run, "red-chair", cut.Head); err != nil {
		t.Fatal(err)
	}
	var refusal error
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		refusal = DeliveredTextRefusal(q, "red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the retire")})
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if refusal != nil {
		t.Fatalf("the chair's spot-check on its delivered retire is refused: %v", refusal)
	}
}
