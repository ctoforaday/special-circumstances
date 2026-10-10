package record

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// ReportJSON is the report as a seat's `show report --json` reads it: the lines it asked for, each
// with what it takes to cite it, at the report head they were read at.
//
// LINES, NOT ONE STRING. Every reader of a seat's `show` is a machine, and what a machine does with
// the report is select from it — the lines under one section, the line an anchor sits on, the lines
// a pattern hits. One shape serves the whole document and each selection of it, so a seat that adds
// a selector parses the same keys it parsed without one.
type ReportJSON struct {
	// Head is the events.id of the latest blue_edit or base_ingest — the head a dispatch pins a
	// seat's reading to. Read on the snapshot the text was rendered from.
	Head int64 `json:"head"`
	// Lines is the report's lines in document order: all of them on a bare read, the window around
	// the anchor under --anchor, the matching lines under --match or --quote. Never omitted: a
	// selection that kept nothing is `[]`.
	Lines []ReportLineJSON `json:"lines"`
}

// ReportLineJSON is one line of the report. A markdown paragraph is a line.
type ReportLineJSON struct {
	// Line is the 1-based line number in the report as it stands — for saying what was read, not
	// for addressing it again after an edit.
	Line int `json:"line"`
	// Heading is the nearest markdown heading at or above the line, "" above the first.
	Heading string `json:"heading"`
	// Text is the line as the report holds it, anchors intact.
	Text string `json:"text"`
}

// ReportLines splits a rendered report into its lines, each under the heading it sits beneath.
// Joining every Text with a newline gives the report back byte for byte.
func ReportLines(report string) []ReportLineJSON {
	lines := strings.Split(report, "\n")
	out := make([]ReportLineJSON, 0, len(lines))
	heading := ""
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "#") {
			heading = strings.TrimSpace(ln)
		}
		out = append(out, ReportLineJSON{Line: i + 1, Heading: heading, Text: ln})
	}
	return out
}

// reportHeadAt is the report head on q: the latest blue_edit or base_ingest, by row id, 0 where
// the record holds neither.
func reportHeadAt(q recordsql.Querier) (int64, error) {
	var head int64
	_, err := queryRowAt(q, []any{&head},
		`SELECT COALESCE(MAX("id"), 0) FROM "events" WHERE "type" IN ('blue_edit', 'base_ingest')`)
	return head, err
}

// ReportAtHead renders the current report and names the head it stands at, both off one snapshot —
// a head read a moment after the text could name an edit the text does not hold. The text is the
// registered renderer's, the bytes every other reader of the report gets.
func ReportAtHead(run Run) (report string, head int64, err error) {
	err = readSnapshot(run, func(q recordsql.Querier) error {
		var rerr error
		if report, rerr = renderProjection(ReportProjectionAt(q)); rerr != nil {
			return rerr
		}
		head, rerr = reportHeadAt(q)
		return rerr
	})
	return report, head, err
}
