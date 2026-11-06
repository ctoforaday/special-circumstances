package runtest

import (
	"errors"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// DropReopened takes anchor out of the `reopened` of every blue edit naming it, and returns those
// edits' events.ids in record order: a record in which an edit changed the anchor's sentence and did
// not say so. It is a TEST SEAM, never a seat path — no verb writes such an edit, because `blue edit`
// derives `reopened` from the report itself (bluedoc.ReopenedAnchors) — and it exists for the
// consistency oracle's staleness rule (plans/compute-and-show.md III.1″.1), whose disagreement only a
// record built some other way can hold.
//
// It writes the body rows directly, through the process's one handle on the record: the events table
// is append-only, its body tables are not, and the event row stays as written. That handle is released
// after, so the next read in this process decodes the edits afresh rather than serving them from its
// stream. Call it where no seat is writing the record.
func DropReopened(run record.Run, anchor string) ([]int64, error) {
	path := filepath.Join(run.Records(), "record.db")
	db, err := recordsql.Open(path)
	if err != nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT "event_id" FROM "blue_edit_reopened" WHERE "value" = ? ORDER BY "event_id"`, anchor)
	if err != nil {
		return nil, err
	}
	var edits []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		edits = append(edits, id)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return nil, err
	}
	_, err = db.Exec(`DELETE FROM "blue_edit_reopened" WHERE "value" = ?`, anchor)
	return edits, errors.Join(err, recordsql.Close(path))
}
