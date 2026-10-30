package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

const b6Archive = "2026-09-11_is-91-prime-b6.tar.gz"

// B6 PASSED OVER TWO CITATIONS NOBODY COULD READ. Its evidence lens verified C-b718c6b6 and
// C-f5782056 `unreachable` at events 100 and 101, minted G-782e2cf3 on the sentence (quote-anchored,
// not about either anchor) and closed it repaired at 178 after blue cited a mirror; the two dead
// anchors stayed in the report, and the chair's PASS stood. Under the ruling (gblock, 2026-10-25:
// "Citation checking isn't optional. Ever.") both are on a failed check owned by the lens that read
// nothing, which owes a gap about each; the PASS is refused naming them; and the gap about either
// does not close while it stands unread.
func TestB6UnreadCitationsHoldThePassAndTheirGapDoesNotClose(t *testing.T) {
	whole := archivedEvents(t, b6Archive)
	unread := []string{"C-b718c6b6", "C-f5782056"}
	for i, a := range unread {
		v := whole[99+i].GetVerify()
		if v.GetAnchor() != a || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE || whole[99+i].GetSeatId() != evLens {
			t.Fatalf("event %d is not %s's unreachable verify of %s: %v", 100+i, evLens, a, whole[99+i])
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

	// THE GAP ABOUT AN UNREAD CITATION DOES NOT CLOSE. Before G-782e2cf3's close (event 178), in the
	// evidence lens's sitting, the lens mints the gap it owes about C-b718c6b6 — outside its budget —
	// and its close as repaired is refused while the citation stands unread.
	if c := whole[177].GetClose(); c.GetGapId() != "G-782e2cf3" || whole[177].GetSeatId() != evLens {
		t.Fatalf("event 178 is not %s's close of G-782e2cf3: %v", evLens, whole[177])
	}
	run = archiveAt(t, b6Archive, whole, 177)
	lens := Identity{Run: run, SeatID: evLens}
	mint := aboutComputation("G-1", unread[0])
	mint.Severity = recordpb.Grade_GRADE_HIGH.Enum()
	mint.Problem = proto.String("the cited source could not be read, and the report still cites it")
	mustAppend(t, lens, mint)
	_, err = Append(lens, &recordpb.Close{GapId: proto.String("G-1"), ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED),
		AnchorSeat: proto.String(evLens), AnchorTool: proto.String("show evidence"), AnchorTarget: proto.String(unread[0]), Prose: proto.String("blue cited a mirror")})
	mustRefuse(t, err, "gap G-1 carries citation "+unread[0], "the report still cites it", "Nothing is recorded")
}
