package sittingwrite

import (
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hookgate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// REPORT TEXT IS HANDED ONE ITEM PER CALL TO THE RECORD, AND ONLY HERE.
//
// A seat's delivery is cut once, at its first item, from the report as it then stands
// (record.CutDelivery), and kept in the sitting's .work-delivery file with the head it was cut at; every
// later item is handed from that file, so every item of one delivery comes from one text. When the
// last item is handed, the tool writes the seat's report_read — the record of what reached its
// context — and the item is emitted after the row is written. An item is rendered alone, with no
// standing, index or notes, and never clipped: the cutter packs to the item budget, so a render over
// it is a cutter defect, refused here as a delivery that cannot advance.
//
// THE SEAT IS STOPPED ON A DELIVERY THAT CANNOT ADVANCE: the writer returns an ErrDeliveryStuck, the
// hook logs it as the tool's defect and ends the seat's turn.

// reportCut is a sitting's delivery as cut, and how far it has been handed.
type reportCut struct {
	record.DeliveryCut
	Handed int `json:"handed"`
}

// ErrDeliveryStuck is a delivery of report text that cannot advance: Reason says why, for the
// operator's log and the seat's stop.
type ErrDeliveryStuck struct{ Reason string }

func (e *ErrDeliveryStuck) Error() string { return hookgate.ErrDeliveryStuck.Error() + ": " + e.Reason }

// Unwrap is hookgate.ErrDeliveryStuck, the name the writer's exit and the hook read it by.
func (e *ErrDeliveryStuck) Unwrap() error { return hookgate.ErrDeliveryStuck }

// stuck wraps err as a delivery that cannot advance.
func stuck(format string, args ...any) error {
	return &ErrDeliveryStuck{Reason: fmt.Sprintf(format, args...)}
}

// IsDeliveryStuck reports whether err is a delivery that cannot advance, and its reason.
func IsDeliveryStuck(err error) (string, bool) {
	var s *ErrDeliveryStuck
	if errors.As(err, &s) {
		return s.Reason, true
	}
	return "", false
}

// reportItemPreamble is the fixed line before every item of report text.
const reportItemPreamble = "REPORT TEXT, item %d of %d of this sitting's delivery — %s. It is handed to you whole, " +
	"one item after each of your calls to the record, and the tool records the delivery when the last item " +
	"reaches you; your envelope waits until then. Read any other text through the selector reads."

// reportItemDelivery renders one item of report text: its preamble and its text, nothing else.
func reportItemDelivery(p record.DeliveryPiece, i, n int) string {
	return fmt.Sprintf(reportItemPreamble, i, n, p.Label) + "\n\n" + p.Text + "\n"
}

// renderedSize is the bytes a delivery's text takes in the hook's JSON document.
func renderedSize(text string) int {
	b, _ := json.Marshal(text)
	return len(b)
}

// itemFits is whether a piece rendered as an item is within the item budget, at the widest numbering.
func itemFits(label, text string) bool {
	return renderedSize(reportItemDelivery(record.DeliveryPiece{Label: label, Text: text}, 999, 999)) <= DeliveryBudget
}

// handReportText hands the seat's next item of report text, cutting the delivery at the first, and
// writes the report_read when it hands the last. now is what the sitting's file holds after.
func handReportText(run record.Run, seat string, staleAreas []string, prev delivered) (string, delivered, error) {
	now := prev
	if now.Report == nil {
		cut, err := record.CutDelivery(run, seat, staleAreas, itemFits)
		if err != nil {
			return "", prev, stuck("the record could not be read to cut %s's delivery: %v", seat, err)
		}
		if len(cut.Pieces) == 0 {
			return "", prev, stuck("%s is owed report text and the cut holds none", seat)
		}
		now.Report = &reportCut{DeliveryCut: cut}
	}
	r := now.Report
	if r.Handed >= len(r.Pieces) {
		return "", now, nil
	}
	i := r.Handed
	text := reportItemDelivery(r.Pieces[i], i+1, len(r.Pieces))
	if renderedSize(text) > DeliveryBudget {
		return "", prev, stuck("item %d of %s's delivery renders to %d bytes, over the item budget of %d", i+1, seat, renderedSize(text), DeliveryBudget)
	}
	cp := *r
	cp.Handed = i + 1
	now.Report = &cp
	if cp.Handed == len(cp.Pieces) {
		if _, err := record.Append(record.Identity{Run: run, SeatID: seat}, &recordpb.ReportRead{Head: proto.Int64(cp.Head)}); err != nil {
			return "", prev, stuck("the record refused %s's delivery: %v", seat, err)
		}
	}
	return text, now, nil
}

// logDefect appends the stuck delivery's reason to the tool's (TOOL, DEFECT) log under the seat, in
// its open sitting: the operator reads it in run.md's Log and capture's failed-sittings audit.
func logDefect(run record.Run, seat, reason string) error {
	_, err := record.Append(record.Identity{Run: run, SeatID: seat}, &recordpb.Log{
		Source: recordpb.LogSource_LOG_SOURCE_TOOL.Enum(),
		Type:   recordpb.LogType_LOG_TYPE_DEFECT.Enum(),
		Text:   proto.String("the delivery of report text to this seat could not advance, and its turn was ended: " + reason),
	})
	return err
}
