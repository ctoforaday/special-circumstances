package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// The Go clock and the SQL window are two spellings of one definition, and this is the only
// thing holding them level: a record with two lenses, a chair that sits twice, and work between
// the registers, read both ways and compared row by row.
func TestClockAgreesWithEventsW(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	mint := func(id string) *recordpb.Mint {
		return &recordpb.Mint{GapId: proto.String(id), AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Class: proto.String("x"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Problem: proto.String("p")}
	}
	for _, seat := range []string{"red-lens-evidence", "red-lens-logic", "red-chair"} {
		if _, _, err := RegisterSeat(Identity{Run: run, SeatID: seat}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, mint("G1")); err != nil {
		t.Fatal(err)
	}
	for _, seat := range []string{"red-chair", "red-lens-evidence"} {
		if _, _, err := RegisterSeat(Identity{Run: run, SeatID: seat}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: run, SeatID: "red-chair"}, mint("G2")); err != nil {
		t.Fatal(err)
	}

	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	evs, want, err := recordsql.EventsW(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 7 || len(want) != 7 {
		t.Fatalf("seeded 7 events, EventsW returned %d events and %d windows", len(evs), len(want))
	}
	var clk Clock
	for i, e := range evs {
		if got := clk.Advance(e); got != want[i] {
			t.Errorf("event %d (%s %s): Clock says %+v, events_w says %+v", i, e.GetSeatId(), recordpb.Word(e.GetType()), got, want[i])
		}
	}
	// The shape itself, so a parity of two wrong answers cannot pass: the second chair register
	// opens epoch 2 and is the chair's second sitting; the lens re-sitting is its second sitting
	// in that epoch.
	if w := want[4]; w != (recordsql.Window{Epoch: 2, Sitting: 2}) {
		t.Errorf("the chair's second register = %+v, want epoch 2 sitting 2", w)
	}
	if w := want[5]; w != (recordsql.Window{Epoch: 2, Sitting: 2}) {
		t.Errorf("the lens re-sitting = %+v, want epoch 2 sitting 2", w)
	}
	if got := CurrentEpochOf(evs[:5]); got != 1 {
		t.Errorf("CurrentEpochOf with the chair re-sat but no work yet = %d, want 1 (the last epoch with work in it)", got)
	}
	if got := CurrentEpochOf(evs); got != 2 {
		t.Errorf("CurrentEpochOf after work in epoch 2 = %d, want 2", got)
	}
}
