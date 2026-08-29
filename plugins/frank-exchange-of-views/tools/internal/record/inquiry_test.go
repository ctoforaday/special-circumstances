package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A LATER SEAT'S BARE `register` MUST NOT MAKE AN EARLIER SEAT'S DISCHARGE PERMANENTLY STALE.
//
// CurrentRound counted every event, and `register` is an event — so a seat that had merely been
// dispatched, having written nothing and decided nothing, advanced "now" past every earlier seat
// and left their round-scoped duties unsatisfiable. The act sat on the record at its own round and
// the gate compared it against a later one, forever.
//
// FOUND BY A SEAT, 2026-08-21. A round-1 merge discharged its round-1 duty, was told the act
// succeeded, and found the projection still demanding it. It retried ten to twelve times —
// different wording, different formatting, inline and from a file — then filed friction reporting
// that the tool returned success and nothing persisted. The events had persisted perfectly:
// CurrentRound was 2, and the only round-2 event on that board was `judge-r2` calling `register`.
//
// RETARGETED, NOT INHERITED. This arrived aimed at `UnvotedInquiriesAt`, the per-line support vote
// — a mechanism retired here in favour of ONE `inquiry-review` per round (see
// TestTheInquiryReviewOffersNoRetiredFlags). The carrier changed and the defect did not: the
// replacement asks the same "was this done in the board's HIGHEST round" question through the same
// CurrentRound, so the rewrite inherited the bug main had already measured and fixed. Retargeting
// is the point — deleting it with its old carrier would have dropped a guard over a live defect.
func TestARegisterFromALaterSeatDoesNotStaleAnEarlierReview(t *testing.T) {
	dir := newRun(t)
	blue := Identity{Run: mustRun(t, dir), SeatID: "blue-respond-r1", Round: 1}
	if _, err := Append(blue, &recordpb.Avenue{
		AvenueId: proto.String("Q1"),
		Status:   recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(),
		Line:     proto.String("a direction"),
	}); err != nil {
		t.Fatal(err)
	}
	merge := Identity{Run: mustRun(t, dir), SeatID: "red-merge-r1", Round: 1}
	if _, err := Append(merge, &recordpb.InquiryReview{
		Reason: proto.String("read the lines against the report as it now stands"),
	}); err != nil {
		t.Fatal(err)
	}

	b, err := BoardState(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if InquiryReviewDue(b) {
		t.Fatal("the round-1 review does not satisfy the round-1 duty")
	}

	// Now a LATER seat registers, and does nothing else.
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, dir), SeatID: "judge-r2", Round: 2}, ""); err != nil {
		t.Fatal(err)
	}
	b, err = BoardState(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := CurrentRound(b); got != 1 {
		t.Fatalf("a bare register advanced CurrentRound to %d — a seat that has written nothing must not move the board's idea of now", got)
	}
	if InquiryReviewDue(b) {
		t.Error("a bare register from judge-r2 made the round-1 merge's review stale.\n\n" +
			"The merge can never satisfy this — it acts at its own round and the gate has moved past it — " +
			"so the round's duty is refused forever while the verb keeps reporting success.")
	}

	// AND THE DUTY STILL BINDS. A round-2 seat doing real work advances the round, and the
	// round-1 review no longer answers for it — or this removed the check rather than repairing it.
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "blue-respond-r2", Round: 2}, &recordpb.Avenue{
		AvenueId: proto.String("Q2"),
		Status:   recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(),
		Line:     proto.String("another"),
	}); err != nil {
		t.Fatal(err)
	}
	b, err = BoardState(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := CurrentRound(b); got != 2 {
		t.Fatalf("real work in round 2 did not advance CurrentRound: got %d", got)
	}
	if !InquiryReviewDue(b) {
		t.Error("round 2 owes its own review and the round-1 one answered for it — the round check is gone, not fixed")
	}
}
