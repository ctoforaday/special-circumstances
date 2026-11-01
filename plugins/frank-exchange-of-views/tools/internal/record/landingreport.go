package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// reportVersion names the state of everything the report's render reads, so a render made at one
// version is the render of any record at the same version.
//
// THE RENDER READS ONE SOURCE AND NOTHING ELSE. RenderAt is renderProjection over ReportProjectionAt
// (queries.go): the report_op view's rows, ordered, handed to the registered renderer, a pure
// function of them that asks the record nothing (reportRenderer). The view is report_part,
// blue_edit, anchor and retire_anchors, each joined to its event (recordsql/views.go). So the
// version is the count and the newest event id of the view.
//
// COUNT AND NEWEST ID ARE ENOUGH BECAUSE A ROW IS ONLY EVER ADDED. events is append-only by trigger,
// and its id is the rowid, so every event takes an id above every id before it. Each row of either
// view is written with the event it belongs to and carries that event's id, so a new row raises the
// newest id. No write path updates or deletes a row of the four tables the view reads; a deleted
// row would lower the count. A version that misses an input accepts a stale render, and the write is
// then decided on a report it does not land on (#1252).
type reportVersion struct {
	ops, lastOp int64
}

// reportVersionQuery reads the version in one statement.
const reportVersionQuery = `SELECT (SELECT count(*) FROM "report_op"), (SELECT coalesce(max("event_id"), 0) FROM "report_op")`

// reportVersionAt is the version of the report q holds. A nil q is a run with no record yet, which
// renders as a record with no rows does.
func reportVersionAt(q recordsql.Querier) (reportVersion, error) {
	var v reportVersion
	if noRecord(q) {
		return v, nil
	}
	_, err := queryRowAt(q, []any{&v.ops, &v.lastOp}, reportVersionQuery)
	return v, err
}

// landingReport is the report a write's checks read: rendered before the write takes the lock, and
// rendered again under it only where another write has moved the report in between.
//
// THE RENDER IS THE SLOW PART OF A CHECK, AND IT GROWS WITH THE RECORD. Made under the write lock, it
// holds every other seat's write for as long as it takes: about 35 ms at 700 events and 440 ms at
// 3,100, so seven lenses writing at once queue for seconds against the busy budget past which a
// write fails. Made before the lock, it is checked under the lock with one query: the version it was
// rendered at against the version the write lands on. Equal, the render is the landing report's, by
// what the version covers.
//
// MOVED IS THE COMMON CASE UNDER LOAD, NOT THE RARE ONE. Every placing act writes an anchor, and an
// anchor is a report op, so of seven lenses filing findings at once, the six that wait for the lock
// all find the version moved: measured, six of seven, every round. Rendering again from the empty text,
// each holds the lock as long as it did before (#1394), and the queue is as long. So the ops the other
// writes added are folded onto the render made before the lock: the replay is a left fold of the ops
// from the empty text (RegisterReportRenderer), and the ops added since are exactly the rows above the
// version's newest id, where the count says no other row moved. A fold that cannot be made —
// the count disagrees, a mutation does not apply — renders again from the empty text, so the
// report and any failure to replay it are the full render's own.
type landingReport struct {
	// declared is whether the write said it reads the report (CheckOnReport, or a validation that
	// reads it), and so made a render before the lock. A write that did not declare it and reads it
	// all the same is refused: its render would hold the lock.
	declared bool
	have     bool
	at       reportVersion
	text     string
	err      error
	// foldedInside and renderedInside count what the write did under the lock — folded the ops added
	// since onto its render, or rendered from the empty text — for the tests that hold each to its case.
	foldedInside, renderedInside int
}

// renderBefore renders run's report before the write takes the lock, from one read snapshot, so
// the version it is keyed by is the version of the rows it rendered. A snapshot that cannot be read
// leaves nothing to reuse, and the write renders under the lock.
func renderBefore(run Run) *landingReport {
	r := &landingReport{declared: true}
	var ops []ReportOp
	var perr error
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		v, err := reportVersionAt(q)
		if err != nil {
			return err
		}
		r.at = v
		ops, perr = ReportProjectionAt(q)
		return nil
	}); err != nil || perr != nil {
		return r // a read that failed is not the report at any version
	}
	// The renderer asks the record nothing, so it replays after the snapshot has closed. A render that
	// fails is kept: it fails the same way on the same rows.
	r.text, r.err = renderProjection(ops, nil)
	r.have = true
	return r
}

// report is the report the write lands on, asked of the writing transaction q.
func (r *landingReport) report(q recordsql.Querier) (string, error) {
	if !r.declared {
		return "", feov.Errorf(feov.Internal,
			"this write reads the report and did not declare it (record.CheckOnReport), so the report would be rendered under the write lock")
	}
	v, err := reportVersionAt(q)
	if err != nil {
		return "", err
	}
	if r.have && r.at == v {
		return r.text, r.err
	}
	if text, ok := r.fold(q, v); ok {
		r.text, r.at = text, v
		r.foldedInside++
		return r.text, nil
	}
	r.text, r.err = RenderAt(q)
	r.at, r.have = v, true
	r.renderedInside++
	return r.text, r.err
}

// fold is the report at version v: the ops added since r's render, replayed over it. ok is false where
// the fold is not the render — r holds no render that replayed, or a row other than
// the ones added since moved (the count of all rows is not the count before plus those) — or where a
// mutation does not apply, so that the full render states the failure in its own words.
func (r *landingReport) fold(q recordsql.Querier, v reportVersion) (string, bool) {
	if !r.have || r.err != nil || v.lastOp <= r.at.lastOp {
		return "", false
	}
	ops, err := reportOpsAfter(q, r.at.lastOp)
	if err != nil || int64(len(ops)) != v.ops-r.at.ops || reportRenderer == nil {
		return "", false
	}
	text, err := reportRenderer(r.text, true, ops)
	return text, err == nil
}

// report is the report this write lands on. A write with no landingReport is a batch's — a
// migration's replay, or `dispatch next` — which holds its own transaction across its acts and renders
// inside it.
func (w writeView) report() (string, error) {
	if w.rendering == nil {
		return RenderAt(w.q)
	}
	return w.rendering.report(w.q)
}
