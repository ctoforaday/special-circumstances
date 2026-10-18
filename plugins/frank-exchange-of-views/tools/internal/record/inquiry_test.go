package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
// CurrentRound was 2, and the only round-2 event on that board was `judge` calling `register`.
//
// The inquiry review asks "is a review recorded since the report or an inquiry last changed", so a
// later seat's register, which changes neither, must not stale it.
func TestARegisterFromALaterSeatDoesNotStaleAnEarlierReview(t *testing.T) {
	dir := newRun(t)
	// The chair sits first so the acts below land in epoch 1 — the round this fixture's path names.
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, dir), SeatID: "red-chair"}, "", ""); err != nil {
		t.Fatal(err)
	}
	// A report exists: before one does there is nothing to read, and no review is owed.
	recordtest.Seed(t, dir, recordtest.At(t, "harness", "ingest", &recordpb.BaseIngest{Text: proto.String("# report")}))
	blue := Identity{Run: mustRun(t, dir), SeatID: "blue-respond"}
	if _, err := Append(blue, &recordpb.Inquiry{
		InquiryId: proto.String("Q1"),
		Status:    recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(),
		Question:  proto.String("a question"),
	}); err != nil {
		t.Fatal(err)
	}
	merge := Identity{Run: mustRun(t, dir), SeatID: "red-chair"}
	if _, err := Append(merge, &recordpb.InquiryReview{
		Reason: proto.String("read the inquiries against the report as it now stands"),
	}); err != nil {
		t.Fatal(err)
	}

	b, err := FamilyOf(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if InquiryReviewDueOf(b.Events, b.At) {
		t.Fatal("the round-1 review does not satisfy the round-1 duty")
	}

	// Now a LATER seat registers, and does nothing else.
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, dir), SeatID: "judge"}, "", "docket"); err != nil {
		t.Fatal(err)
	}
	b, err = FamilyOf(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := b.At.CurrentEpoch(b.Events); got != 1 {
		t.Fatalf("a bare register advanced CurrentEpoch to %d — a seat that has written nothing must not move the board's idea of now", got)
	}
	if InquiryReviewDueOf(b.Events, b.At) {
		t.Error("a bare register from judge made the round-1 merge's review stale.\n\n" +
			"The chair can never satisfy this — it acts at its own round and the gate has moved past it — " +
			"so the round's duty is refused forever while the verb keeps reporting success.")
	}

	// AND THE DUTY STILL BINDS. An inquiry proposed after the review is one it never read, so the
	// review no longer answers — or this removed the check rather than repairing it. The chair sits
	// again: what follows is epoch 2, the "round 2" this fixture means.
	if _, _, err := RegisterSeat(Identity{Run: mustRun(t, dir), SeatID: "red-chair"}, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: "blue-respond"}, &recordpb.Inquiry{
		InquiryId: proto.String("Q2"),
		Status:    recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(),
		Question:  proto.String("another"),
	}); err != nil {
		t.Fatal(err)
	}
	b, err = FamilyOf(mustRun(t, dir))
	if err != nil {
		t.Fatal(err)
	}
	if got := b.At.CurrentEpoch(b.Events); got != 2 {
		t.Fatalf("real work in epoch 2 did not advance CurrentEpoch: got %d", got)
	}
	if !InquiryReviewDueOf(b.Events, b.At) {
		t.Error("an inquiry proposed after the review owes a new one, and the earlier review answered for it — the check is gone, not fixed")
	}
}

// THE REVIEW IS OWED WHEN WHAT IT READ CHANGES, AND ONLY THEN. Each step is one act on the record and
// the answer once the record ends there: no report owes nothing; the report's arrival owes a read;
// the read settles it; an edit, a proposal and a fate each owe a new one; the chair's own acts, a
// lens's register and the chair's next sitting owe nothing.
func TestTheInquiryReviewIsOwedWhenWhatItReadChanges(t *testing.T) {
	inquiry := func(id string, status recordpb.InquiryStatus) *recordpb.Inquiry {
		return &recordpb.Inquiry{InquiryId: proto.String(id), Question: proto.String("a question"), Status: status.Enum()}
	}
	review := &recordpb.InquiryReview{Reason: proto.String("read")}
	steps := []struct {
		what string
		act  func(b *stage)
		want bool
	}{
		{"an inquiry and no report", func(b *stage) { b.add("blue-respond", inquiry("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED)) }, false},
		{"the report ingested", func(b *stage) { b.ingest() }, true},
		{"the chair's review", func(b *stage) { b.add("red-chair", review) }, false},
		{"the chair's position", func(b *stage) { b.add("red-chair", &recordpb.Position{Text: proto.String("p")}) }, false},
		{"a lens's register", func(b *stage) { b.register(evLens) }, false},
		{"the chair's next sitting", func(b *stage) { b.register("red-chair") }, false},
		{"the report edited", func(b *stage) { b.edit("", "# report", "# report, revised") }, true},
		{"the chair's review again", func(b *stage) { b.add("red-chair", review) }, false},
		{"an inquiry proposed", func(b *stage) { b.add("blue-respond", inquiry("Q2", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED)) }, true},
		{"the chair's review a third time", func(b *stage) { b.add("red-chair", review) }, false},
		{"an inquiry given a fate", func(b *stage) { b.add("blue-respond", inquiry("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY)) }, true},
	}
	for n := range steps {
		b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").register("red-chair")
		for _, s := range steps[:n+1] {
			s.act(b)
		}
		fam, err := FamilyOf(b.seed())
		if err != nil {
			t.Fatal(err)
		}
		if got := InquiryReviewDueOf(fam.Events, fam.At); got != steps[n].want {
			t.Errorf("after %s: review owed = %v, want %v", steps[n].what, got, steps[n].want)
		}
	}
}
