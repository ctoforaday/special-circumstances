package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// THE SPOT-CHECKS ON THE RECORD, JOINED TO THE BOARD.
//
// A spot-check is owed before a PASS and at no other time, and the write enforces it: a PASS is
// refused while a stale area stands that no spot-check in the chair's sitting names
// (BlockerStaleArea, passblockers.go). Nothing here is a second statement of that rule. This file
// is the reader's account of the spot-checks the chair filed — which stale areas each read, which
// closures it re-read, what it found, and what the board says the archive held at that point, a
// number no seat can author.
//
// One claim in a spot-check is checkable against that number: "there was nothing to sample". A
// spot-check that says so while the board holds archived closures is reported to the reader as a
// claim the record contradicts.

// SpotCheck is one spot-check the chair filed, joined to what the board says was available to
// sample.
type SpotCheck struct {
	Epoch   int
	Sitting int
	SeatID  string
	// Areas are the stale areas the chair says it read the changes against.
	Areas []string
	// Sampled are the archived closures the chair says it re-verified.
	Sampled []string
	// Prose is what the spot-check found — a defect in a stale area is stated here — or why it
	// named nothing.
	Prose string
	// Archived is the board's count of closures available at the START of this epoch — what there
	// was to sample.
	Archived int
}

// Archive is the epoch each closed gap on the board closed in: what the archive held at the start
// of any epoch, replayed rather than reported.
type Archive struct{ closed []int }

// ArchiveOf is the archive of a family's gaps.
func ArchiveOf(gaps []*Gap) Archive {
	var a Archive
	for _, g := range gaps {
		if g != nil && g.HasClosed {
			a.closed = append(a.closed, g.ClosedEpoch)
		}
	}
	return a
}

// Before is the archive at the START of epoch R: every gap closed in an epoch strictly before R.
//
// The epoch is the closing event's, derived by the record from the chair's stored sittings
// (events_w), so no seat-id shape stands between a closure and its epoch.
func (a Archive) Before(epoch int) int {
	n := 0
	for _, c := range a.closed {
		if c < epoch {
			n++
		}
	}
	return n
}

// archiveOfGapView is the archive as the `gap` view closes each gap, with the closing acts' bodies looked
// up in evs. A narrowed read that renders spot-checks loads the close and docket-ruling families
// for it, and closeStateOf refuses loudly when it did not.
func archiveOfGapView(q recordsql.Querier, evs []*Event, win WindowIndex) (Archive, error) {
	var a Archive
	if noRecord(q) {
		return a, nil
	}
	rows, err := q.Query(`SELECT "gap_id", "close_event", "bench_close_event", "closed_by_bench" FROM "gap"`)
	if err != nil {
		return a, fmt.Errorf("asking the record for its archive: %w", err)
	}
	defer rows.Close()
	byID := eventsByID(evs, win)
	for rows.Next() {
		var id string
		var closeEvent, benchEvent sql.NullInt64
		var closedByBench bool
		if err := rows.Scan(&id, &closeEvent, &benchEvent, &closedByBench); err != nil {
			return a, err
		}
		c, err := closeStateOf(id, closeEvent, benchEvent, closedByBench, byID, win)
		if err != nil {
			return a, err
		}
		if c != nil {
			a.closed = append(a.closed, c.closedEpoch)
		}
	}
	return a, rows.Err()
}

// SpotCheckOf is the spot-check e records, joined to the archive at the start of its epoch; false
// when e is not a spot-check. Every reader that shows a spot-check takes it from here and renders
// it with Describe.
func SpotCheckOf(e *Event, at WindowIndex, a Archive) (SpotCheck, bool) {
	// The BODY is the type test. Named `body` because the row this fills is ALSO called SpotCheck —
	// record.SpotCheck is the reader's row, recordpb.SpotCheck the event body.
	body, ok := recordpb.BodyAs[*recordpb.SpotCheck](e)
	if !ok {
		return SpotCheck{}, false
	}
	w := at.Of(e)
	return SpotCheck{
		Epoch: w.Epoch, Sitting: w.Sitting, SeatID: e.GetSeatId(),
		Areas: body.GetAreas(), Sampled: body.GetIds(), Prose: body.GetReason(),
		Archived: a.Before(w.Epoch),
	}, true
}

// SpotChecks is each standing spot-check, in order, joined to the board's archive.
func SpotChecks(f Family) (checks []SpotCheck) {
	if len(f.Gaps) == 0 && len(f.Events) == 0 {
		return nil
	}
	// THE STORED EPOCH, because the number goes on a recorded ACT (the spot-check the chair filed),
	// and the record holds which of the chair's sittings each act is in.
	a := ArchiveOf(f.Gaps)
	for _, e := range f.Live() {
		sc, ok := SpotCheckOf(e, f.At, a)
		if !ok {
			continue
		}
		checks = append(checks, sc)
	}
	return checks
}

// Describe renders one spot-check as a reader-facing line: what it names, then its prose. It is the
// ONE rendering of a spot-check — the board's list and every debate transcript print it — so what
// the chair found reaches each of them whatever the spot-check names.
//
// THE PROSE IS PRINTED ON EVERY FORM. A spot-check naming no closure printed "naming no closures"
// and dropped its prose, so a defect the chair stated in a stale area's spot-check — where its duty
// puts it, with no verdict recorded — reached no reader.
func (s SpotCheck) Describe() string {
	var named []string
	if len(s.Areas) > 0 {
		named = append(named, "read stale area(s) "+strings.Join(s.Areas, ", "))
	}
	if len(s.Sampled) > 0 {
		named = append(named, fmt.Sprintf("re-verified %s of %d archived closure(s)", strings.Join(s.Sampled, ", "), s.Archived))
	}
	if len(named) == 0 {
		named = append(named, "naming no stale area and no closure")
	}
	line := fmt.Sprintf("#%d (%s): %s", s.Sitting, s.SeatID, strings.Join(named, "; "))
	if s.Prose != "" {
		line += " — " + s.Prose
	}
	return line
}
