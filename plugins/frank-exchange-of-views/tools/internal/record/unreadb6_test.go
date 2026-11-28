package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const b6Archive = "2026-09-11_is-91-prime-b6.tar.gz"

// B6 PASSED OVER TWO CITATIONS NOBODY COULD READ. Its evidence lens verified C-b718c6b6 and
// C-f5782056 `unreachable` at events 95 and 96 (the migrated record's ids: migrate writes the
// spot-check before them, which the delivered-text rule refuses, as no event), minted G-782e2cf3 on the sentence (quote-anchored,
// not about either anchor) and closed it repaired at 159 after blue cited a mirror; the two dead
// anchors stayed in the report, and the chair's PASS stood. Under the ruling (gblock, 2026-10-25:
// "Citation checking isn't optional. Ever.") both are on a failed check owned by the lens that read
// nothing, which owes a gap about each; the PASS is refused naming them; and the lens's close of the
// gap about either is refused while it stands unread.
func TestB6UnreadCitationsHoldThePassAndTheirGapDoesNotClose(t *testing.T) {
	whole, lost := archivedRun(t, b6Archive)
	unread := []string{"C-b718c6b6", "C-f5782056"}
	for i, a := range unread {
		v := whole[94+i].GetVerify()
		if v.GetAnchor() != a || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE || whole[94+i].GetSeatId() != evLens {
			t.Fatalf("event %d is not %s's unreachable verify of %s: %v", 95+i, evLens, a, whole[94+i])
		}
	}
	pass := slices.IndexFunc(whole, func(e *Event) bool {
		return e.GetType() == recordpb.EventType_EVENT_TYPE_VERDICT && e.GetSeatId() == "red-chair"
	})
	if pass < 0 {
		t.Fatal("b6 records no chair verdict")
	}

	// THE RECORD BEFORE THE CHAIR'S PASS: two failed checks, neither carried by a gap about it.
	run := archiveAt(t, b6Archive, whole, pass)
	var failed []string
	for _, b := range mustBlockers(t, run) {
		switch {
		case b.Kind == BlockerFailedCheck:
			if b.Owner != evLens || b.Detail != "" || b.AnsweredBy() != "" {
				t.Errorf("failed check %+v, want a citation's owned by %s and carried by no gap", b, evLens)
			}
			failed = append(failed, b.Subject)
		case b.Kind == BlockerUnchecked && slices.Contains(unread, b.Subject):
			t.Errorf("%s is unchecked, want it on a failed check: it was checked, and the check read nothing", b.Subject)
		}
	}
	if !slices.Equal(failed, unread) {
		t.Fatalf("failed checks before the PASS = %q, want %q", failed, unread)
	}
	err := validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil {
		t.Fatal("the gate admits b6's PASS over two citations nobody could read")
	}
	for _, a := range unread {
		if !strings.Contains(err.Error(), a+" (read nothing of its source by "+evLens) {
			t.Errorf("the PASS refusal does not name %s and the lens whose check read nothing: %v", a, err)
		}
	}

	// THE LENS'S CLOSE OF THE GAP ABOUT AN UNREAD CITATION IS REFUSED. Before G-782e2cf3's close, in the
	// evidence lens's sitting, the lens mints the gap it owes about C-b718c6b6,
	// and its close as repaired is refused while the citation stands unread.
	//
	// WHERE THE MIGRATION LOST THAT CLOSE (its manifest's lost_acts: the lens's sitting holds no
	// delivery of report text), the record holds no close of G-782e2cf3 and the gap reads open; the
	// arm runs at the place the close was written, with the lens's delivery cut there, so the close
	// it records is refused for the unread citation and not for want of a delivery.
	at := slices.IndexFunc(whole, func(e *Event) bool { return e.GetClose().GetGapId() == "G-782e2cf3" })
	closeLost := slices.IndexFunc(lost, func(a LostAct) bool { return a.Word == "close" && a.Gap == "G-782e2cf3" && a.Seat == evLens })
	switch {
	case closeLost >= 0:
		if at >= 0 {
			t.Fatalf("the migration lists %s's close of G-782e2cf3 as lost, and the record holds it at %d", evLens, at+1)
		}
		if open := gapOpenAt(t, archiveAt(t, b6Archive, whole, len(whole)), "G-782e2cf3"); !open {
			t.Error("G-782e2cf3, whose close the migration lost, reads closed")
		}
		at = slices.IndexFunc(whole, func(e *Event) bool { return e.GetTs() > lost[closeLost].TS })
	case at < 0 || whole[at].GetSeatId() != evLens:
		t.Fatalf("the record holds no %s close of G-782e2cf3 and the migration lost none", evLens)
	}
	run = archiveAt(t, b6Archive, whole, at)
	lens := Identity{Run: run, SeatID: evLens}
	if closeLost >= 0 {
		deliver(t, run, evLens)
	}
	mint := aboutComputation("G-1", unread[0])
	mint.Severity = recordpb.Grade_GRADE_HIGH.Enum()
	mint.Problem = proto.String("the cited source could not be read, and the report still cites it")
	mustAppend(t, lens, mint)
	_, err = Append(lens, &recordpb.Close{GapId: proto.String("G-1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String(evLens), AnchorTool: proto.String("show evidence"), AnchorTarget: proto.String(unread[0]), Prose: proto.String("blue cited a mirror")})
	mustRefuse(t, err, "gap G-1 carries citation "+unread[0], "the report still cites it", "Nothing is recorded")
}

// gapOpenAt is whether gap reads open on run's record.
func gapOpenAt(t *testing.T, run Run, gap string) bool {
	t.Helper()
	var open bool
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		return q.QueryRow(`SELECT "open" FROM "gap" WHERE "gap_id" = ?`, gap).Scan(&open)
	}); err != nil {
		t.Fatalf("gap %s: %v", gap, err)
	}
	return open
}
