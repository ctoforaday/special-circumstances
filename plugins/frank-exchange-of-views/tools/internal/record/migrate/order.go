package migrate

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportproj"
)

// THE REPORT'S TEXT IS BROUGHT INTO THE ORDER IT SHIPS IN, WHERE THAT MOVES NOTHING BUT SECTIONS.
//
// A record written before the write held the order (report.TextOrder) can hold its sections in
// another: assembly placed three of them by name, so the order a seat read was not the one that
// shipped. Assembly ships the text as written now, and the migrated record holds it in the order
// the old assembly gave the reader: the base is reordered (report.Ordered), whole sections moved
// and nothing else, and every later act is replayed over it as it was recorded.
//
// ONLY WHERE THE REPLAY IS THE SAME REPORT. The acts locate by the text they quote, so one whose
// quote runs across two sections that are no longer neighbours does not replay, and an edit that
// itself wrote a section out of order leaves the text out of order whatever the base holds. The
// reordered record is kept when no act is refused and its acts, replayed over the base as its
// author wrote it, give the same text section for section, every anchor on the sentence it was
// on. Otherwise the record is replayed again as written: its text keeps its author's order,
// nothing is lost, and the manifest says which it was.
//
// The write refuses an out-of-order text only for what the write ADDS (report.OutOfOrderIn), so a
// record kept as written still takes every edit that does not put a further section out of place,
// the one that moves a section to its place included.

// The values of Manifest.ReportOrder.
const (
	OrderNoReport  = "no_report"  // the record holds no report text
	OrderInOrder   = "in_order"   // the text stands in the order it ships in, as its author wrote it
	OrderReordered = "reordered"  // the base was reordered and every act replayed over it
	OrderAsWritten = "as_written" // the text is out of order and reordering the base does not replay it
)

// replayInOrder is Replay with the report's text brought into order where that replays cleanly.
//
// THE TRIAL IS ONE BATCH, KEPT OR ROLLED BACK. The reordered replay is written in a batch; kept, it
// is committed, and not kept, it is rolled back and the record replayed as written in a fresh batch
// on the same database. The schema, applied when the record was opened and before the first batch
// began, stays; the rolled-back trial leaves no row, so the replay as written starts from the empty
// record the trial did. A failure of the trial's batch (record.ErrBatchFailed) is the migration's
// failure, never a reason to replay as written.
func replayInOrder(src Source, reg Registry, dst record.Run, opt Options) (*Result, error) {
	was := record.Now
	defer func() { record.Now = was }()
	record.Migrating = true
	defer func() { record.Migrating = false }()
	batch, err := record.BeginBatch(dst)
	if err != nil {
		return nil, err
	}
	// Every exit that does not commit — a refusal, an error, a panic — rolls the open batch back;
	// after a commit it does nothing. batch is the variable, so this reaches the as-written batch too.
	defer func() { _ = batch.Rollback() }()
	res, err := replay(src, reg, batch, dst, opt, true)
	if err != nil {
		return nil, err
	}
	base, have, ops, err := record.ReportProjectionAt(batch.Querier())
	if err != nil {
		return nil, err
	}
	if have && base != res.base {
		now, nerr := reportproj.Replay(base, true, ops)
		written, werr := reportproj.Replay(res.base, true, ops)
		if len(res.Refusals) == 0 && nerr == nil && werr == nil && now == report.Ordered(written) {
			if err := batch.Commit(); err != nil {
				return nil, err
			}
			res.ReportOrder = OrderReordered
			return res, nil
		}
		if err := batch.Rollback(); err != nil {
			return nil, fmt.Errorf("discarding the reordered replay: %w", err)
		}
		// The trial pinned the clock to its events; the replay as written starts from the clock the
		// trial did, since the synthesized cast pins it only when the first stamp parses.
		record.Now = was
		if batch, err = record.BeginBatch(dst); err != nil {
			return nil, err
		}
		if res, err = replay(src, reg, batch, dst, opt, false); err != nil {
			return nil, err
		}
	}
	if err := batch.Commit(); err != nil {
		return nil, err
	}
	if !have {
		res.ReportOrder = OrderNoReport
		return res, nil
	}
	text, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		return nil, fmt.Errorf("the migrated record's report does not render: %w", err)
	}
	res.ReportOrder = OrderInOrder
	if _, out := report.OutOfOrderIn("", text); out {
		res.ReportOrder = OrderAsWritten
	}
	return res, nil
}
