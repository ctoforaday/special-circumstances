package record

import (
	"fmt"
	"math"
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
// THE CHAIR'S PROSE IS ITS JUDGEMENT AND NOTHING ELSE. What there was to read is the record's: the
// archive's size at the epoch, and, for each stale area named, the pin and the changes to the report
// behind it, which the write stamps (stampSpotCheckChanges). m40's chair wrote "Changes since 260: 42
// edits" over a record holding 4 and claimed source re-reads it never made; the count and the ids
// are now the tool's, and no seat states them.

// SpotCheck is one spot-check the chair filed, joined to what the board says was available to
// sample.
type SpotCheck struct {
	Epoch   int
	Sitting int
	SeatID  string
	// Areas are the stale areas the chair named, each with the pin and the changes the write stamped.
	Areas []SpotCheckArea
	// Sampled are the archived closures the chair says it re-verified.
	Sampled []string
	// Prose is what the spot-check found — a defect in a stale area is stated here — or why it
	// named nothing.
	Prose string
	// Archived is the board's count of closures available at the START of this epoch — what there
	// was to sample.
	Archived int
}

// SpotCheckArea is one area a spot-check named: the lens seat, its pin when the chair named it, and
// the event id of each change to the report after that pin — the tool's count, never the chair's.
type SpotCheckArea struct {
	SeatID  string
	Pin     int64
	Changes []int64
}

// SpotCheckAreasOf pairs each named area with its stamped pin and the stamped changes after it.
func SpotCheckAreasOf(b *recordpb.SpotCheck) []SpotCheckArea {
	pins := b.GetPins()
	out := make([]SpotCheckArea, 0, len(b.GetAreas()))
	for i, a := range b.GetAreas() {
		ar := SpotCheckArea{SeatID: a}
		if i < len(pins) {
			ar.Pin = pins[i]
			ar.Changes = changesAfter(b.GetChanges(), ar.Pin)
		}
		out = append(out, ar)
	}
	return out
}

// changesAfter is each change id in changes after pin, in record order.
func changesAfter(changes []int64, pin int64) []int64 {
	out := []int64{}
	for _, c := range changes {
		if c > pin {
			out = append(out, c)
		}
	}
	return out
}

// stampSpotCheckChanges writes each named area's pin and the changes to the report behind the
// earliest of them onto a spot-check, which no writer may do itself. A spot-check that ARRIVES with
// either is refused: a seat that could state the changes could state 42 over a record holding 4.
//
// The pin is the lens fold's (lensStates): the latest head the whole report was served to the lens
// at in a sitting it sat through. A change is an event that moves the report head — a blue_edit or a
// report_part (reportHeadOf) — so the changes after a stale area's pin are exactly what moved the
// head past it.
func stampSpotCheckChanges(q recordsql.Querier, b *recordpb.SpotCheck) error {
	if len(b.GetPins()) > 0 || len(b.GetChanges()) > 0 {
		return fmt.Errorf("a spot-check's pins and changes are stamped by the tool from the record — the changes to the report behind each area's pin — and never supplied by a seat; record it without them")
	}
	if len(b.GetAreas()) == 0 {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	ids := win.IDs(evs)
	_, openings := dispatchLedger(evs, ids, win)
	ends := satThroughOf(evs, ids, win, openings)
	reads := servedOf(evs, ids, win)
	earliest := int64(math.MaxInt64)
	for _, a := range b.GetAreas() {
		pin := pinOf(reads[a], ends, a, math.MaxInt64)
		b.Pins = append(b.Pins, pin)
		earliest = min(earliest, pin)
	}
	for i, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_BLUE_EDIT, recordpb.EventType_EVENT_TYPE_REPORT_PART:
			if ids[i] > earliest {
				b.Changes = append(b.Changes, ids[i])
			}
		}
	}
	return nil
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
// The epoch is the closing event's, stamped by the record from the chair's sittings (events.epoch),
// so no seat-id shape stands between a closure and its epoch.
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
	rows, err := q.Query(`SELECT "gap_id", ` + gapClosureColumns + ` FROM "gap"`)
	if err != nil {
		return a, fmt.Errorf("asking the record for its archive: %w", err)
	}
	defer rows.Close()
	byID := eventsByID(evs, win)
	for rows.Next() {
		var id string
		var cl gapClosure
		if err := rows.Scan(append([]any{&id}, cl.dest()...)...); err != nil {
			return a, err
		}
		c, err := closeStateOf(id, cl, byID, win)
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
		Areas: SpotCheckAreasOf(body), Sampled: body.GetIds(), Prose: body.GetReason(),
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
	for _, e := range f.Events {
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
		named = append(named, "read stale area(s) "+DescribeAreas(s.Areas))
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

// DescribeAreas renders the areas a spot-check named with what the record holds behind each pin,
// grouped by pin in the order the areas were named: "a, b (pin 260: 4 change(s) to the report since —
// events 320, 321, 322, 323)". The count and the ids are the stamp's; the chair's echo and every
// reader of a spot-check print them from here.
func DescribeAreas(areas []SpotCheckArea) string {
	var order []int64
	byPin := map[int64][]SpotCheckArea{}
	for _, a := range areas {
		if _, seen := byPin[a.Pin]; !seen {
			order = append(order, a.Pin)
		}
		byPin[a.Pin] = append(byPin[a.Pin], a)
	}
	groups := make([]string, 0, len(order))
	for _, pin := range order {
		as := byPin[pin]
		seats := make([]string, 0, len(as))
		for _, a := range as {
			seats = append(seats, a.SeatID)
		}
		changes := as[0].Changes
		ids := make([]string, 0, len(changes))
		for _, c := range changes {
			ids = append(ids, fmt.Sprint(c))
		}
		behind := fmt.Sprintf("pin %d: %d change(s) to the report since", pin, len(changes))
		if len(ids) > 0 {
			behind += " — events " + strings.Join(ids, ", ")
		}
		groups = append(groups, strings.Join(seats, ", ")+" ("+behind+")")
	}
	return strings.Join(groups, "; ")
}
