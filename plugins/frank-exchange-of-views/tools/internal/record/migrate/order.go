package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
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
func replayInOrder(src Source, reg Registry, dst record.Run, opt Options) (*Result, error) {
	res, err := replay(src, reg, dst, opt, true)
	if err != nil {
		return nil, err
	}
	base, have, ops, err := record.ReportProjection(dst)
	if err != nil {
		return nil, err
	}
	if have && base != res.base {
		now, nerr := reportproj.Replay(base, true, ops)
		was, werr := reportproj.Replay(res.base, true, ops)
		if len(res.Refusals) == 0 && nerr == nil && werr == nil && now == report.Ordered(was) {
			res.ReportOrder = OrderReordered
			return res, nil
		}
		if err := recordsql.CloseUnder(dst.Records()); err != nil {
			return nil, err
		}
		for _, name := range dbSiblings {
			if err := os.Remove(filepath.Join(dst.Records(), name)); err != nil && !os.IsNotExist(err) {
				return nil, fmt.Errorf("replaying the record as written: %w", err)
			}
		}
		if res, err = replay(src, reg, dst, opt, false); err != nil {
			return nil, err
		}
	}
	if !have {
		res.ReportOrder = OrderNoReport
		return res, nil
	}
	text, err := reportproj.RenderFromRecord(dst)
	if err != nil {
		return nil, err
	}
	res.ReportOrder = OrderInOrder
	if _, out := report.OutOfOrderIn("", text); out {
		res.ReportOrder = OrderAsWritten
	}
	return res, nil
}
