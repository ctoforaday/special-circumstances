package recordsql

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// reportCheckpointEvery is how many report ops a reader replays at most, less one, after the
// checkpoint it starts from (ReportCheckpointDDL).
//
// THE SPACING TRADES A READER'S REPLAY AGAINST THE ROWS THE RECORD STORES. A reader replays up to
// reportCheckpointEvery-1 ops on every read of the report, and the record stores a copy of the text
// every reportCheckpointEvery ops. Measured on a 1,863-event run (257 ops, a 4.4 kB report): the
// whole replay is 16 ms, one op about 0.13 ms whatever the text's length, so the worst reader at 8 is
// under a millisecond, and the stored text is an eighth of a copy per op.
const reportCheckpointEvery = 8

// Newest is the event a read of the report as it stands now reads at: above every event's id.
const Newest int64 = math.MaxInt64

// ReportSource is where a read of the report as of an event starts: the newest checkpoint at or
// below that event, and the ops after it up to the event. Parted is whether a checkpoint was found —
// the text it holds always holds a part, since a fold that holds none fails. Where none was found,
// the record holds no op at or below the event, and Text is empty. Failure is the checkpoint's
// failure where the report did not replay through it (ReportCheckpointDDL): the report as of the
// event is that failure, whatever the ops after it.
type ReportSource struct {
	Text    string
	Parted  bool
	Failure string
	Ops     []reportproj.ReportOp
}

// Fold is the report src starts from with its ops replayed by replay (reportproj.Replay, or a stand-in
// a test passes), or the checkpoint's failure.
func (src ReportSource) Fold(replay func(text string, parted bool, ops []reportproj.ReportOp) (string, error)) (string, error) {
	if src.Failure != "" {
		return "", errors.New(src.Failure)
	}
	return replay(src.Text, src.Parted, src.Ops)
}

// ReportSourceAt reads the report's source as of event at (Newest for the report as it stands now).
//
// A MISSING CHECKPOINT IS AN ERROR, NEVER A LONGER REPLAY. Every insert that adds the record's first
// op writes a checkpoint, and so does every insert that brings the ops since the last one to
// reportCheckpointEvery (stampReportCheckpoint). So an op at or below at with no checkpoint below it,
// or a tail of reportCheckpointEvery ops or more, is a record this binary's write path did not write,
// and the read says so rather than replaying from wherever it can.
func ReportSourceAt(q Querier, at int64) (ReportSource, error) {
	var src ReportSource
	var cp int64
	var text, failure sql.NullString
	err := q.QueryRow(`SELECT "event_id", "text", "failure" FROM "report_checkpoint" WHERE "event_id" <= ? ORDER BY "event_id" DESC LIMIT 1`, at).
		Scan(&cp, &text, &failure)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return src, fmt.Errorf("reading the report checkpoint: %w", err)
	default:
		src.Text, src.Parted, src.Failure = text.String, true, failure.String
	}
	if src.Ops, err = ReportOpsBetween(q, cp, at); err != nil {
		return src, err
	}
	if !src.Parted && len(src.Ops) > 0 {
		return src, fmt.Errorf("the record holds report ops from event %d and no report checkpoint at or below them — "+
			"the insert of a record's first op writes one, so this record was not written by this binary's write path", src.Ops[0].EventID)
	}
	if len(src.Ops) >= reportCheckpointEvery {
		return src, fmt.Errorf("the record holds %d report ops after its checkpoint at event %d and none since — "+
			"the insert that brings them to %d writes one, so this record was not written by this binary's write path",
			len(src.Ops), cp, reportCheckpointEvery)
	}
	return src, nil
}

// ReportOpsBetween is the report_op rows of the events after event id `after` and at or before
// `upto`, in replay order. The second key orders the one case where an event yields several rows — a
// retire naming more than one anchor — so replay is deterministic rather than whatever SQLite returns.
func ReportOpsBetween(q Querier, after, upto int64) ([]reportproj.ReportOp, error) {
	rows, err := q.Query(`SELECT "event_id", "kind", "a", "b", "exact" FROM "report_op" WHERE "event_id" > ? AND "event_id" <= ? ORDER BY "event_id", "a"`, after, upto)
	if err != nil {
		return nil, fmt.Errorf("reading the report ops: %w", err)
	}
	defer rows.Close()
	var ops []reportproj.ReportOp
	for rows.Next() {
		var op reportproj.ReportOp
		var a, b sql.NullString
		if err := rows.Scan(&op.EventID, &op.Kind, &a, &b, &op.Exact); err != nil {
			return nil, err
		}
		op.A, op.B = a.String, b.String
		ops = append(ops, op)
	}
	return ops, rows.Err()
}

// stampReportCheckpoint writes the report's text as event id leaves it, where id is the first event
// to add a report op, or brings the ops since the last checkpoint to reportCheckpointEvery. It runs
// in InsertTx after the body is written, so the event's own ops are on the report_op view, and every
// write of an event — a seat's act, a migration's replay, a seeded fixture — keeps the checkpoints
// a reader requires (ReportSourceAt).
//
// AN EVENT THAT ADDS NO OP COSTS ONE QUERY, and an op between checkpoints one more. The fold runs
// only where a checkpoint lands, over the checkpoint before it and the ops since.
//
// AN OP THE REPORT CANNOT REPLAY THROUGH IS NOT REFUSED HERE: which ops land is the write path's
// guards' decision, and the checkpoint records what folding them gives — the replay's failure, where
// it fails, which every read as of the checkpoint or after it then states (ReportSourceAt).
func stampReportCheckpoint(tx *sql.Tx, id int64) error {
	var mine int64
	if err := tx.QueryRow(`SELECT count(*) FROM "report_op" WHERE "event_id" = ?`, id).Scan(&mine); err != nil {
		return fmt.Errorf("reading this event's report ops: %w", err)
	}
	if mine == 0 {
		return nil
	}
	var cp, since int64
	if err := tx.QueryRow(`SELECT c."at", (SELECT count(*) FROM "report_op" WHERE "event_id" > c."at")
	  FROM (SELECT coalesce(max("event_id"), 0) AS "at" FROM "report_checkpoint") c`).Scan(&cp, &since); err != nil {
		return fmt.Errorf("reading the report's last checkpoint: %w", err)
	}
	if cp > 0 && since < reportCheckpointEvery {
		return nil
	}
	// The source is read here rather than through ReportSourceAt: the tail this insert completes is
	// reportCheckpointEvery ops long, or the record's first, which a reader refuses as a missing
	// checkpoint.
	var src ReportSource
	if cp > 0 {
		var text, failure sql.NullString
		if err := tx.QueryRow(`SELECT "text", "failure" FROM "report_checkpoint" WHERE "event_id" = ?`, cp).Scan(&text, &failure); err != nil {
			return fmt.Errorf("reading the report checkpoint at event %d: %w", cp, err)
		}
		src = ReportSource{Text: text.String, Parted: true, Failure: failure.String}
	}
	var err error
	if src.Ops, err = ReportOpsBetween(tx, cp, id); err != nil {
		return err
	}
	var text, failure any
	if folded, err := src.Fold(reportproj.Replay); err != nil {
		failure = err.Error()
	} else {
		text = folded
	}
	if _, err := tx.Exec(`INSERT INTO "report_checkpoint" ("event_id", "text", "failure") VALUES (?, ?, ?)`, id, text, failure); err != nil {
		return fmt.Errorf("writing the report checkpoint at event %d: %w", id, err)
	}
	return nil
}
