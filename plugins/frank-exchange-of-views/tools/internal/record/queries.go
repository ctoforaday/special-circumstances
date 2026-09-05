package record

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// This file is plans/record-sqlite.md §III step 4: the hand-written projections, replaced with
// queries, one at a time, each with its test (queries_parity_test.go holds each one against the
// fold it replaced). ConvergenceVsVerdict was the first and is the template; everything here
// follows its contract:
//
//   - `openRunForRead` returning (nil, nil) is a run that has recorded nothing yet — every
//     question below answers with its honest zero, exactly as the fold answered over an empty
//     event stream.
//   - A run that CANNOT be read (legacy shards, a vanished separated root) is an error, never an
//     empty answer — that distinction is the whole reason openRunForRead has three states.
//   - Each function keeps the error posture of the fold it replaced. Some folds deliberately
//     swallowed read errors (a board that cannot be read does not block a filing); the query
//     versions swallow exactly the same ones, and the comment at each site says so.

// queryRow runs one single-row query against the record. found=false is the honest zero: no
// record yet, or no row matched.
func queryRow(run Run, dest []any, q string, args ...any) (found bool, err error) {
	db, err := openRunForRead(run)
	if err != nil {
		return false, err
	}
	if db == nil {
		return false, nil
	}
	if err := db.QueryRow(q, args...).Scan(dest...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("record: asking the record: %w", err)
	}
	return true, nil
}

// recordHas answers an existence question. A missing record holds nothing.
func recordHas(run Run, q string, args ...any) (bool, error) {
	var one int
	return queryRow(run, []any{&one}, q, args...)
}

// BoardCounts is the board in two numbers, read from the board_counts view rather than folded —
// the site plans/record-sqlite.md names at view.go's headline count. A run with no record yet is
// an empty board: 0 and 0.
func BoardCounts(run Run) (open, closed int, err error) {
	if _, err := queryRow(run, []any{&open, &closed},
		`SELECT "open_gaps", "closed_gaps" FROM "board_counts"`); err != nil {
		return 0, 0, err
	}
	return open, closed, nil
}

// RecordedOutcome is the bench's own terminal act — the latest `outcome` event's verdict — and
// nothing else. "" means the bench has not recorded one, which is a fact, not a fallback: callers
// that want the record's derived verdict when the bench is silent layer DeriveVerdict on top
// (TerminalVerdict does), and callers that must not substitute a derivation for the bench's act
// (the dashboard's server lifetime, #270) read this directly.
//
// Read errors fold into "" deliberately — both former copies of this fold did the same, because
// every caller treats "cannot read" and "not recorded" identically: keep watching.
func RecordedOutcome(run Run) string {
	var v string
	if _, err := queryRow(run, []any{&v},
		`SELECT "verdict" FROM "outcome" ORDER BY "event_id" DESC LIMIT 1`); err != nil {
		return ""
	}
	return v
}

// MintCheckKind is the check_kind a gap was minted with. A gap not on the record answers
// UNSPECIFIED — the caller that must distinguish "no such gap" already ran requireGap.
func MintCheckKind(run Run, gapID string) (recordpb.CheckKind, error) {
	var word string
	found, err := queryRow(run, []any{&word},
		`SELECT COALESCE("check_kind", '') FROM "mint" WHERE "gap_id" = ?`, gapID)
	if err != nil || !found || word == "" {
		return recordpb.CheckKind_CHECK_KIND_UNSPECIFIED, err
	}
	vd, ok := recordpb.BySpelling(recordpb.CheckKind(0).Descriptor(), word)
	if !ok {
		return recordpb.CheckKind_CHECK_KIND_UNSPECIFIED,
			fmt.Errorf("record: gap %s was minted with check_kind %q, which this binary's schema does not declare", gapID, word)
	}
	return recordpb.CheckKind(vd.Number()), nil
}

// RoundsWithRevision counts the distinct rounds that filed a round record. Errors fold into 0,
// as the audit that reads it always treated an unreadable record: zero rounds it can vouch for.
func RoundsWithRevision(run Run) int {
	var n int
	if _, err := queryRow(run, []any{&n},
		`SELECT count(DISTINCT "round") FROM "events" WHERE "type" = ?`,
		recordpb.Word(recordpb.EventType_EVENT_TYPE_REVISION)); err != nil {
		return 0
	}
	return n
}

// RegisteredSeats lists the seat ids that registered in this run, in event order — every seat
// the BUILD registered, whether or not it then sat.
func RegisteredSeats(run Run) ([]string, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}
	rows, err := db.Query(`SELECT e."seat_id" FROM "register" r JOIN "events" e ON e."id" = r."event_id" ORDER BY r."event_id"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its registered seats: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// FindingMarkerRecorded reports whether any finding OR anchor event names this marker id — the
// membership question the torn-splice heal asks of an fx marker already sitting on the report.
// Read errors fold into false: the caller then takes the ordinary fresh path, exactly as it did
// when the merged read failed. Its proof twin below keeps the two namespaces as separate as the
// folds kept them.
func FindingMarkerRecorded(run Run, id string) bool {
	if id == "" {
		return false
	}
	found, err := recordHas(run, `SELECT 1 FROM "finding" WHERE "finding_id" = ?
	  UNION SELECT 1 FROM "anchor" WHERE "id" = ? LIMIT 1`, id, id)
	return err == nil && found
}

// ProofMarkerRecorded is FindingMarkerRecorded's twin for proof markers.
func ProofMarkerRecorded(run Run, id string) bool {
	if id == "" {
		return false
	}
	found, err := recordHas(run, `SELECT 1 FROM "proof" WHERE "proof_id" = ? LIMIT 1`, id)
	return err == nil && found
}

// gradeAtColumn names the regrade/mint column for a contested dimension; "" is a dimension this
// binary cannot read a grade at — the same refusal Gap.GradeAt returns false for.
func gradeAtColumn(d recordpb.GradeDimension) string {
	switch d {
	case recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY:
		return "severity"
	case recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD:
		return "likelihood"
	case recordpb.GradeDimension_GRADE_DIMENSION_IMPACT:
		return "impact"
	case recordpb.GradeDimension_GRADE_DIMENSION_COMPLEXITY:
		return "complexity_cost"
	}
	return ""
}

// gradeAt is a gap's CURRENT grade at one dimension — the gap view's regrade overlay, which is
// the one SQL statement of "latest regrade that touched the axis, else the mint's". found=false
// is a gap that is not on the record.
func gradeAt(run Run, gapID, col string) (word string, found bool, err error) {
	var v sql.NullString
	// The column name is drawn from gradeAtColumn's closed set, never from input.
	found, err = queryRow(run, []any{&v},
		fmt.Sprintf(`SELECT %q FROM "gap" WHERE "gap_id" = ?`, "current_"+col), gapID)
	return v.String, found, err
}
