package record

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// fileByDispatch writes m as `dispatch next` files its dockets: under the chair, in a write of its
// own, marked byDispatch — so the seat-only check (one counted exchange first) is not asked, and
// every other check is. A fixture that needs a gap docketed, and is not about who filed it, files it
// this way rather than as a seat no rule would admit.
func fileByDispatch(t *testing.T, run Run, m *recordpb.Motion) error {
	t.Helper()
	db, err := openRun(run)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := recordsql.BeginWrite(db)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	ev := &Event{}
	typ, err := recordpb.SetBody(ev, m)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeIn(tx.Tx, ev, chairSeat, typ, m, func(*sql.Tx) error {
		return validateAct(writeView{q: tx, files: filesOf(run), byDispatch: true}, chairSeat, typ, m)
	}, stamp); err != nil {
		return err
	}
	return tx.Commit()
}

// mustFileByDispatch is fileByDispatch, failing the test on a refusal.
func mustFileByDispatch(t *testing.T, run Run, m *recordpb.Motion) {
	t.Helper()
	if err := fileByDispatch(t, run, m); err != nil {
		t.Fatal(err)
	}
}

// convene readies the bench on gaps as `dispatch next` does — a bench row naming them, written by the
// chair — and sits the bench in the docket sitting that row convenes, which holds every docket motion
// on those gaps filed before the row (benchHeld).
func convene(t *testing.T, run Run, gaps ...string) Identity {
	t.Helper()
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(benchSeat), GapIds: gaps})
	return sit(t, run, benchSeat)
}

// exchangeOn gives gap its one counted exchange, the cheapest the record counts: the chair engages
// blue on it, and blue sits and sits again having recorded nothing on it — a null turn, which ends the
// exchange and stalls it.
func exchangeOn(t *testing.T, run Run, gap string) {
	t.Helper()
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(blueRespondSeat), GapIds: []string{gap}})
	sit(t, run, blueRespondSeat)
	sit(t, run, blueRespondSeat)
}
