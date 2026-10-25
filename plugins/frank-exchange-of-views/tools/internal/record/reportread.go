package record

import (
	"database/sql"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// A LENS'S PIN IS THE HEAD THE WHOLE REPORT WAS LAST SERVED TO IT AT (#1458).
//
// The pin was the head of the dispatch whose sitting the lens opened, and a sitting opens when the
// hook brackets it — so a lens that crashed, skipped its audit or read only at anchors left the
// same record as one that read the report and found nothing: an empty sitting, pin = head. At the
// final head nothing re-readied it, and its area went uncovered at the PASS. The read verb now
// records what it served (ReportRead), and the fold reads the pin from there (lensStates): a lens
// that was never served the report at the head is still ready at it, through the rule that readies
// a lens behind the head, until it is served it or retires — and a retired lens behind the head is a
// stale area the chair's spot-check covers, as before.
//
// SERVED, NOT READ. The tool cannot see a seat read; it can see that the whole text at a head
// reached the seat in a sitting, which a read at an anchor or a selection cannot claim.

// ServedPinNote is what moves a lens's pin, as its work list's audit item states it.
const ServedPinNote = "Your pin moves when the tool serves you the whole report — `show report`, bare or with --json and no selector — in a sitting you sit through; a read at an anchor or a selection serves part of it and moves nothing, and a sitting that ends with the report not served to you at the head leaves you ready at it"

// ServeReport is the report as the whole-report read serves it to a lens, and its head, with the
// ReportRead that records the serving — one transaction, so the head recorded is the head of the
// text returned. It writes nothing where the run has its outcome or was captured, or where the seat's current
// sitting already holds a read at this head; the text is served either way. A write the record
// cannot take — a lock it could not get — is the caller's error, and nothing is served.
func ServeReport(id Identity) (text string, head int64, err error) {
	// A CAPTURED RUN IS SERVED AND RECORDS NOTHING, asked before the append, which a captured run
	// refuses as the write it is (requireRunNotCaptured): the seat asked for a read.
	if captured, err := RunCaptured(id.Run); err != nil {
		return "", 0, err
	} else if captured {
		return ReportAtHead(id.Run)
	}
	body := &recordpb.ReportRead{}
	_, err = Append(id, body, CheckOnReport(func(l *Landing) error {
		if text, err = l.Report(); err != nil {
			return err
		}
		if head, err = reportHeadAt(l.w.q); err != nil {
			return err
		}
		body.Head = proto.Int64(head)
		if v, err := l.Outcome(); err != nil {
			return err
		} else if v != "" {
			return ErrWritesNothing
		}
		served, err := servedThisSittingAt(l.w.q, id.SeatID, head)
		if err != nil {
			return err
		}
		if served {
			return ErrWritesNothing
		}
		return nil
	}))
	if errors.Is(err, ErrWritesNothing) {
		err = nil
	}
	if err != nil {
		return "", 0, err
	}
	return text, head, nil
}

// servedThisSittingAt is whether seatID's current sitting already holds a read at head.
func servedThisSittingAt(q recordsql.Querier, seatID string, head int64) (bool, error) {
	var one int
	err := q.QueryRow(`SELECT 1 FROM "report_read" r JOIN "events" e ON e."id" = r."event_id"
	  WHERE e."seat_id" = ?1 AND r."head" = ?2
	    AND e."sitting_id" IS (SELECT max("id") FROM "sittings" WHERE "seat_id" = ?1) LIMIT 1`, seatID, head).Scan(&one)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("asking whether %s was already served the report at head %d this sitting: %w", seatID, head, err)
	}
	return true, nil
}

// requireReportReadAt refuses a read by any seat but a lens — no other seat's audit is pinned — and
// one whose head is not the report's head on the record it lands on: the head is the text served.
func requireReportReadAt(q recordsql.Querier, seatID string, b *recordpb.ReportRead) error {
	if seatclass.RoleOfSeat(seatID) != "lens" {
		return fmt.Errorf("a report read records the whole report served to a lens, and %s is not a lens", seatID)
	}
	head, err := reportHeadAt(q)
	if err != nil {
		return err
	}
	if b.GetHead() != head {
		return fmt.Errorf("a report read names head %d and the report's head is %d — it records the head the text was served at", b.GetHead(), head)
	}
	return nil
}

// served is one ReportRead a lens's sitting holds: the place its sitting opened at, and the head.
type served struct {
	opening, head int64
}

// servedOf is each seat's reads, in stream order, by the place its sitting opened at — dispatchLedger's
// places, so they compare with its openings and with satThrough's. A read in no sitting is in none.
func servedOf(evs []*Event, ids []int64, win WindowIndex) map[string][]served {
	place := placesOf(evs, ids, win)
	out := map[string][]served{}
	for _, e := range evs {
		r, ok := recordpb.BodyAs[*recordpb.ReportRead](e)
		if !ok {
			continue
		}
		o, in := place[win.Of(e).SittingID]
		if !in || win.Of(e).SittingID == 0 {
			continue
		}
		out[e.GetSeatId()] = append(out[e.GetSeatId()], served{opening: o, head: r.GetHead()})
	}
	return out
}

// pinOf is the lens's pin: the latest head of its reads in a sitting it sat through whose opening
// precedes before, 0 where none does.
func pinOf(reads []served, ends satThrough, seat string, before int64) int64 {
	var pin int64
	for _, r := range reads {
		if r.opening < before && ends.sat(seat, r.opening) && r.head > pin {
			pin = r.head
		}
	}
	return pin
}
