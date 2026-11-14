package record

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
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

// changedSinceSQL is the first head move after ?1 that changed the sentence of anchor ?2 — an edit
// whose reopened lists it, which includes one that left it bare — or a retire that removed it.
const changedSinceSQL = `SELECT MIN(x) FROM (
  SELECT r."event_id" AS x FROM "blue_edit_reopened" r WHERE r."value" = ?2 AND r."event_id" > ?1
  UNION ALL
  SELECT a."event_id" FROM "retire_anchors" a WHERE a."value" = ?2 AND a."event_id" > ?1)`

// DeliveredTextRefusal is the delivered-text rule asked of one act, as the write asks it: the refusal
// of body under seatID, or nil where the act depends on no report text or stands on its delivery. A
// close (not a carry) depends on its gap's text, a spot-check on its sitting's delivery, and a ruling
// on a motion about gaps on their passages; a ruling on a position names no passage and needs none.
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
	case *recordpb.MotionRule:
		if b.GetMotionId() == "" {
			return nil
		}
		gaps, err := motionGaps(q, b.GetMotionId())
		if err != nil || len(gaps) == 0 {
			return err
		}
		return requireDeliveredText(q, seatID, "a ruling on a motion about "+gapsNamed(gaps), gaps)
	}
	return nil
}

// UndeliveredError is the delivered-text rule's refusal: the act, and why its text is not delivered.
type UndeliveredError struct{ msg string }

func (e *UndeliveredError) Error() string { return e.msg }

// requireDeliveredText refuses an act of seatID that depends on the text around anchors unless the
// seat's current sitting holds its delivery and no head move or retire after that delivery's head
// changed any of those anchors' sentences. what names the act and where its text arrives.
func requireDeliveredText(q recordsql.Querier, seatID, what string, anchors []string) error {
	head, held, err := DeliveryThisSitting(q, seatID)
	if err != nil {
		return err
	}
	if !held {
		return &UndeliveredError{fmt.Sprintf("%s stands only on report text delivered to you, and this sitting holds no complete delivery: the text arrives on your work list, one item after each of your calls to the record — finish it, then record this", what)}
	}
	for _, id := range anchors {
		var at sql.NullInt64
		if err := q.QueryRow(changedSinceSQL, head, id).Scan(&at); err != nil {
			return fmt.Errorf("asking whether %s changed since your delivery: %w", anchor.Token(id), err)
		}
		if at.Valid {
			return &UndeliveredError{fmt.Sprintf("%s stands only on report text delivered to you, and the text around %s changed after your delivery was cut: the change arrives on your work list at your next sitting", what, id)}
		}
	}
	return nil
}

// requireReportRead refuses a report_read the tool may not write: one under a seat that is not a
// lens, the chair or the bench; a second in the seat's sitting; one whose head is not a head move at
// or before the report's head.
func requireReportRead(q recordsql.Querier, seatID string, b *recordpb.ReportRead) error {
	switch seatclass.RoleOfSeat(seatID) {
	case "lens", "chair", "bench":
	default:
		return fmt.Errorf("a report read records report text delivered to a lens, the chair or the bench, and %s is none of them", seatID)
	}
	if _, held, err := DeliveryThisSitting(q, seatID); err != nil {
		return err
	} else if held {
		return fmt.Errorf("%s's sitting already holds its delivery: one report read per seat and sitting", seatID)
	}
	head, err := reportHeadAt(q)
	if err != nil {
		return err
	}
	var kind string
	err = q.QueryRow(`SELECT "type" FROM "events" WHERE "id" = ?`, b.GetHead()).Scan(&kind)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if (kind != "blue_edit" && kind != "report_part") || b.GetHead() > head {
		return fmt.Errorf("a report read names head %d, which is no edit or part at or before the report's head %d: it records the head its delivery was cut at", b.GetHead(), head)
	}
	return nil
}

// motionGapsSQL is the gaps a motion is about: a docket's, a remand's (a docket ruling) and a grade
// motion's.
const motionGapsSQL = `SELECT d."gap_id" FROM "motion" m JOIN "motion_docket" d ON d."event_id" = m."event_id" WHERE m."motion_id" = ?1
  UNION SELECT g."gap_id" FROM "motion" m JOIN "motion_grade" g ON g."event_id" = m."event_id" WHERE m."motion_id" = ?1 AND g."gap_id" IS NOT NULL`

// motionGaps is the gaps motion id is about, none for a motion about no gap.
func motionGaps(q recordsql.Querier, id string) ([]string, error) {
	rows, err := q.Query(motionGapsSQL, id)
	if err != nil {
		return nil, fmt.Errorf("asking which gaps motion %s is about: %w", id, err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var g string
		if err := rows.Scan(&g); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// gapsNamed is ids joined for a message.
func gapsNamed(ids []string) string { return strings.Join(ids, ", ") }

// reportTextItems is the report text the seat is owed into its context this sitting, as blocking
// items: none once its sitting holds its delivery, and none for a seat that is handed no report text.
// open is the seat's list so far, from which the chair's stale areas are read.
//
//   - a lens: the whole report (its first whole delivery, or its full re-read, or after an upheld
//     position named it), or the changes since its pin and its own gaps' passages;
//   - the chair: each stale area with changes it has not borne (area_change);
//   - the bench: each gap an unruled motion is about.
func reportTextItems(run Run, role, seatID string, open []Item) ([]Item, error) {
	if role != "lens" && role != "chair" && role != "bench" {
		return nil, nil
	}
	var out []Item
	err := readSnapshot(run, func(q recordsql.Querier) error {
		if noRecord(q) {
			return nil
		}
		head, err := reportHeadAt(q)
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
				item(about(SubjectReport, strconv.FormatInt(head, 10)), "the whole report arrives on your work list this sitting, one item after each of your calls to the record, and your envelope waits until all of it has been delivered to you. "+wholeDeliveryNote)
			} else {
				item(about(SubjectReport, strconv.FormatInt(head, 10)), "the changes to the report since your pin — each edit as the whole paragraphs it touched, before and after — and the passages of your own gaps arrive on your work list this sitting, one item after each of your calls to the record, and your envelope waits until all of it has been delivered to you. Read anything else through the selector reads")
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
					item(about(SubjectArea, it.Subject.ID), fmt.Sprintf("the %d change(s) behind %s's pin that you have not borne arrive on your work list this sitting, each as the paragraphs it touched before and after, one item after each of your calls to the record; a spot-check of the area stands only once all of them have been delivered, and your envelope waits until they have", n, it.Subject.ID))
				}
			}
		case "bench":
			rows, err := q.Query(motionPassagesSQL, "")
			if err != nil {
				return err
			}
			defer rows.Close()
			for rows.Next() {
				var gap, loc string
				if err := rows.Scan(&gap, &loc); err != nil {
					return err
				}
				item(about(SubjectGap, gap), "the passage of gap "+gap+", which a motion before you is about, arrives on your work list this sitting, one item after each of your calls to the record; a ruling on a motion about it stands only once it has been delivered, and your envelope waits until it has")
			}
			return rows.Err()
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

// ReportTextAreas is the stale areas whose changes a seat's list holds as report text: the chair's
// delivery is cut from them.
func ReportTextAreas(w WorkJSON) []string {
	var out []string
	for _, it := range w.Sitting.Open {
		if it.Kind == ItemReportText && it.Subject != nil && it.Subject.Kind == SubjectArea {
			out = append(out, it.Subject.ID)
		}
	}
	return out
}
