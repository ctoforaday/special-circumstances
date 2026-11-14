// Package deliverytest completes a seat's delivery of report text for a fixture, as the delivery
// writer does: the cut from the record as it stands, then the report_read the writer appends when it
// hands the cut's last item (record.RecordDelivery). An act that stands only on delivered text — a
// close, a spot-check, a ruling on a motion about gaps — is then judged on a sitting that holds its
// delivery, which is the only state the live write accepts it in.
//
// It hands no item text: what reaches a seat's context is the writer's business, tested in
// sittingwrite. The fixture is held to the writer's refusal of an empty cut, so a seat owed no
// report text is a broken fixture, not a delivery.
package deliverytest

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// fitsAll packs a delivery into one piece per section or change: the fixture hands no item, so the
// item budget does not bind it.
func fitsAll(string, string) bool { return true }

// Deliver completes seatID's delivery in its open sitting and returns the head it was cut at.
func Deliver(t testing.TB, run record.Run, seatID string) int64 {
	t.Helper()
	w, err := record.WorkOfSeat(run, seatclass.RoleOfSeat(seatID), seatID)
	if err != nil {
		t.Fatalf("reading %s's work list to cut its delivery: %v", seatID, err)
	}
	cut, err := record.CutDelivery(run, seatID, record.ReportTextSubjects(w), fitsAll)
	if err != nil {
		t.Fatalf("cutting %s's delivery: %v", seatID, err)
	}
	if len(cut.Pieces) == 0 {
		t.Fatalf("%s is owed no report text: the cut holds none (no report yet, or nothing its role is handed)", seatID)
	}
	if err := record.RecordDelivery(run, seatID, cut.Head); err != nil {
		t.Fatalf("recording %s's delivery at head %d: %v", seatID, cut.Head, err)
	}
	return cut.Head
}

// DeliverIfOwed completes seatID's delivery when its work list holds report text still owed this
// sitting, and does nothing otherwise: a fixture appending a sequence of acts through the write path
// calls it before each, as a live seat's work list hands its report text before its other work.
func DeliverIfOwed(t testing.TB, run record.Run, seatID string) {
	t.Helper()
	w, err := record.WorkOfSeat(run, seatclass.RoleOfSeat(seatID), seatID)
	if err != nil {
		t.Fatalf("reading %s's work list: %v", seatID, err)
	}
	for _, it := range w.Sitting.Open {
		if it.Kind == record.ItemReportText {
			Deliver(t, run, seatID)
			return
		}
	}
}

// StandsOnDeliveredText is whether body is an act the delivered-text rule judges: a close, a
// spot-check or a motion ruling (record.DeliveredTextRefusal decides which of those it refuses).
func StandsOnDeliveredText(body proto.Message) bool {
	switch body.(type) {
	case *recordpb.Close, *recordpb.SpotCheck, *recordpb.MotionRule:
		return true
	}
	return false
}
