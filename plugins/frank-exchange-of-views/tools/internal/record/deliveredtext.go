package record

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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

// requireDeliveredText refuses an act of seatID that depends on the text around anchors unless the
// seat's current sitting holds its delivery and no head move or retire after that delivery's head
// changed any of those anchors' sentences. what names the act and where its text arrives.
func requireDeliveredText(q recordsql.Querier, seatID, what string, anchors []string) error {
	head, held, err := DeliveryThisSitting(q, seatID)
	if err != nil {
		return err
	}
	if !held {
		return fmt.Errorf("%s stands only on report text delivered to you, and this sitting holds no complete delivery: the text arrives on your work list, one item after each of your calls to the record — finish it, then record this", what)
	}
	for _, id := range anchors {
		var at sql.NullInt64
		if err := q.QueryRow(changedSinceSQL, head, id).Scan(&at); err != nil {
			return fmt.Errorf("asking whether %s changed since your delivery: %w", anchor.Token(id), err)
		}
		if at.Valid {
			return fmt.Errorf("%s stands only on report text delivered to you, and the text around %s changed after your delivery was cut: the change arrives on your work list at your next sitting", what, id)
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
