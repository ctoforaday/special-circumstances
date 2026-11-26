package record

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A SEAT'S ACT THAT DEPENDS ON REPORT TEXT STANDS ONLY ON TEXT DELIVERED TO IT (gblock 2026-11-13).
//
// The tool writes one report_read per completed delivery of report text into a seat's context: the
// delivery writer appends it when it hands the delivery's last item, its head the report head the
// delivery was cut at. Everything else about deliveries is read off that row and the report_delivery
// view as they stand: whether a sitting holds its delivery, which delivery a lens is owed next, its
// pin, and whether the text an act depends on has changed since it was delivered. Each is an id
// comparison over rows after an id; nothing is rebuilt as it stood at a past event.

// deliveryThisSittingSQL is the seat's report_read in its sitting ?2, and its head.
const deliveryThisSittingSQL = `SELECT r."head" FROM "report_read" r JOIN "events" e ON e."id" = r."event_id"
  WHERE e."seat_id" = ?1 AND e."sitting_id" = ?2 ORDER BY r."event_id" DESC LIMIT 1`

// DeliveryThisSitting is the head of the delivery the seat's current sitting holds, and whether it
// holds one.
func DeliveryThisSitting(q recordsql.Querier, seatID string) (head int64, held bool, err error) {
	sitting, _, err := latestSittingAt(q, seatID)
	if err != nil || !sitting.Valid {
		return 0, false, err
	}
	err = q.QueryRow(deliveryThisSittingSQL, seatID, sitting.Int64).Scan(&head)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return 0, false, nil
	case err != nil:
		return 0, false, fmt.Errorf("asking whether %s's sitting holds its delivery: %w", seatID, err)
	}
	return head, true, nil
}

// nextDeliveryWholeSQL is whether the seat's next delivery is the whole report, by report_delivery's
// own rule asked of the row the writer is about to append.
var nextDeliveryWholeSQL = `SELECT ` + recordsql.WholeDeliverySQL(`(SELECT COALESCE(MAX("id"), 0) + 1 FROM "events")`, `?1`)

// NextDeliveryWhole is whether a lens's next delivery is the whole report: it has no complete whole
// delivery and its full re-read in sittings it sat through, or an upheld position named it after its
// latest delivery. Any other seat's delivery is never the whole report.
func NextDeliveryWhole(q recordsql.Querier, seatID string) (bool, error) {
	if seatclass.RoleOfSeat(seatID) != "lens" {
		return false, nil
	}
	var whole bool
	if err := q.QueryRow(nextDeliveryWholeSQL, seatID).Scan(&whole); err != nil {
		return false, fmt.Errorf("asking which delivery %s is owed: %w", seatID, err)
	}
	return whole, nil
}

// pinSQL is a seat's pin: the head of its latest delivery in a sitting it sat through.
const pinSQL = `SELECT COALESCE(MAX("head"), 0) FROM "report_delivery" WHERE "seat_id" = ?1 AND "sat_through"`

// PinAt is the seat's pin, 0 where it has had no complete delivery in a sitting it sat through.
func PinAt(q recordsql.Querier, seatID string) (int64, error) {
	var pin int64
	if err := q.QueryRow(pinSQL, seatID).Scan(&pin); err != nil {
		return 0, fmt.Errorf("asking %s's pin: %w", seatID, err)
	}
	return pin, nil
}

// changedSinceSQL is the first head move after ?1 that changed the sentence of anchor ?2: an edit
// whose reopened lists it, which includes one that left it bare. These are the head moves (head_move)
// that touch one anchor's sentence; a part appends and touches none. A retire is no head move: it takes
// an anchor out and leaves the text the delivery carried, so it never stales a delivered text (gblock
// 2026-11-20).
const changedSinceSQL = `SELECT MIN(r."event_id") FROM "blue_edit_reopened" r WHERE r."value" = ?2 AND r."event_id" > ?1`

// DeliveredTextRefusal is the delivered-text rule asked of one act, as the write asks it: the refusal
// of body under seatID, or nil where the act depends on no report text or stands on its delivery. The
// rule governs an act that depends on report text its seat was handed: a close (not a carry) depends
// on its gap's text, a spot-check on its sitting's delivery. A ruling depends on no report text: the
// bench rules on the evidence the parties put on the record, their statements and closings, as a
// court rules on what the parties present, so no ruling stands on a delivery.
// The validate arms and migrate's replay ask this one function, so an archived act migrate drops is
// exactly one the live write refuses.
func DeliveredTextRefusal(q recordsql.Querier, seatID string, body proto.Message) error {
	switch b := body.(type) {
	case *recordpb.Close:
		if b.CarriedFrom != nil {
			return nil
		}
		return requireDeliveredText(q, seatID, "a close of "+b.GetGapId(), []string{b.GetGapId()})
	case *recordpb.SpotCheck:
		return requireDeliveredText(q, seatID, "a spot-check", nil)
	}
	return nil
}

// UndeliveredError is the delivered-text rule's refusal: the act, and why its text is not delivered.
// It reaches the seat coded conflict (undelivered): the act is well formed, and the record's state —
// the sitting's delivery — refuses it until the seat has read what it is owed.
type UndeliveredError struct{ msg string }

func (e *UndeliveredError) Error() string { return e.msg }

// undelivered is the delivered-text refusal as the write returns it: an *UndeliveredError, which
// migrate's replay finds by type, under a conflict code, which a seat's --json envelope reads.
func undelivered(msg string) error {
	return feov.Wrap(feov.Conflict, &UndeliveredError{msg}, "%s", msg)
}

// requireDeliveredText refuses an act of seatID that depends on the text around anchors unless the
// seat's current sitting holds its delivery and no head move after that delivery's head
// changed any of those anchors' sentences. what names the act and where its text arrives.
func requireDeliveredText(q recordsql.Querier, seatID, what string, anchors []string) error {
	head, held, err := DeliveryThisSitting(q, seatID)
	if err != nil {
		return err
	}
	if !held {
		return undelivered(fmt.Sprintf("%s stands only on report text delivered to you, and this sitting holds no complete delivery: the text arrives on your work list, one item after each of your tool calls — finish it, then record this", what))
	}
	for _, id := range anchors {
		var at sql.NullInt64
		if err := q.QueryRow(changedSinceSQL, head, id).Scan(&at); err != nil {
			return fmt.Errorf("asking whether %s changed since your delivery: %w", anchor.Token(id), err)
		}
		if at.Valid {
			return undelivered(fmt.Sprintf("%s stands only on report text delivered to you, and the text around %s changed after your delivery was cut: the change arrives on your work list at your next sitting", what, id))
		}
	}
	return nil
}

// requireReportRead refuses a report_read the tool may not write: one under a seat that is not a
// lens or the chair; a second in the seat's sitting; one whose head is not a head move at
// or before the report's head.
func requireReportRead(q recordsql.Querier, seatID string, b *recordpb.ReportRead) error {
	switch seatclass.RoleOfSeat(seatID) {
	case "lens", "chair":
	default:
		return fmt.Errorf("a report read records report text delivered to a lens or the chair, and %s is neither", seatID)
	}
	if _, held, err := DeliveryThisSitting(q, seatID); err != nil {
		return err
	} else if held {
		return fmt.Errorf("%s's sitting already holds its delivery: one report read per seat and sitting", seatID)
	}
	head, err := ReportHeadAt(q)
	if err != nil {
		return err
	}
	var moved bool
	if err := q.QueryRow(`SELECT EXISTS (SELECT 1 FROM "head_move" WHERE "event_id" = ?)`, b.GetHead()).Scan(&moved); err != nil {
		return err
	}
	if !moved || b.GetHead() > head {
		return fmt.Errorf("a report read names head %d, which is no head move — a part or an edit — at or before the report's head %d: it records the head its delivery was cut at", b.GetHead(), head)
	}
	return nil
}

// gapsNamed is ids joined for a message.
func gapsNamed(ids []string) string { return strings.Join(ids, ", ") }

// reportTextItems is the report text the seat is owed into its context this sitting, as blocking
// items: none once its sitting holds its delivery, and none for a seat that is handed no report text.
// open is the seat's list so far, from which the chair's stale areas are read.
//
//   - a lens: the whole report (its first whole delivery, or its full re-read, or after an upheld
//     position named it), or the changes since its pin and its own gaps' passages;
//   - the chair: each stale area with changes it has not borne (area_change).
//
// The bench is handed none: its rulings stand on the parties' statements on the record.
func reportTextItems(run Run, role, seatID string, open []Item) ([]Item, error) {
	if role != "lens" && role != "chair" {
		return nil, nil
	}
	var out []Item
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		head, err := ReportHeadAt(q)
		if err != nil || head == 0 {
			return err
		}
		sitting, _, err := latestSittingAt(q, seatID)
		if err != nil || !sitting.Valid {
			return err
		}
		if _, held, err := DeliveryThisSitting(q, seatID); err != nil || held {
			return err
		}
		item := func(sub *ItemSubject, what string) {
			out = append(out, Item{What: what, Blocks: true, Kind: ItemReportText, Subject: sub})
		}
		switch role {
		case "lens":
			whole, err := NextDeliveryWhole(q, seatID)
			if err != nil {
				return err
			}
			if whole {
				item(about(SubjectReport, strconv.FormatInt(head, 10)), "the whole report arrives on your work list this sitting, one item after each of your tool calls, and your envelope waits until all of it has been delivered to you. "+wholeDeliveryNote)
			} else {
				item(about(SubjectReport, strconv.FormatInt(head, 10)), "the changes to the report since your pin — each edit as the whole paragraphs it touched, before and after — and the passages of your own gaps arrive on your work list this sitting, one item after each of your tool calls, and your envelope waits until all of it has been delivered to you. Read anything else through the selector reads")
			}
		case "chair":
			for _, it := range open {
				if it.Kind != ItemKind(BlockerStaleArea) || it.Subject == nil {
					continue
				}
				var n int
				if err := q.QueryRow(`SELECT COUNT(*) FROM "area_change" WHERE "seat_id" = ? AND "change" <= ?`, it.Subject.ID, head).Scan(&n); err != nil {
					return err
				}
				if n > 0 {
					item(about(SubjectArea, it.Subject.ID), fmt.Sprintf("the %d change(s) behind %s's pin that you have not borne arrive on your work list this sitting, each as the paragraphs it touched before and after, one item after each of your tool calls; a spot-check of the area stands only once all of them have been delivered, and your envelope waits until they have", n, it.Subject.ID))
				}
			}
		}
		return nil
	})
	return out, err
}

// wholeDeliveryNote is what the whole-report delivery is for, as a lens's item says it.
const wholeDeliveryNote = "A lens is delivered the whole report at its first sitting and once more at its next — its full re-read — and at every other sitting the changes since its pin and its own gaps' passages; the PASS is refused while a lens that has sat still owes either whole delivery"

// RecordDelivery writes seatID's report_read: the delivery of report text cut at head has reached the
// seat's context. The delivery writer calls it as it hands the delivery's last item.
func RecordDelivery(run Run, seatID string, head int64) error {
	_, err := Append(Identity{Run: run, SeatID: seatID}, &recordpb.ReportRead{Head: proto.Int64(head)})
	return err
}

// ReportTextSubjects is what a seat's report-text items are about, in list order: the chair's stale
// areas. The delivery is cut from them (CutDelivery); a lens's items name the report.
func ReportTextSubjects(w WorkJSON) []string {
	var out []string
	for _, it := range w.Sitting.Open {
		if it.Kind == ItemReportText && it.Subject != nil && it.Subject.Kind == SubjectArea {
			out = append(out, it.Subject.ID)
		}
	}
	return out
}
