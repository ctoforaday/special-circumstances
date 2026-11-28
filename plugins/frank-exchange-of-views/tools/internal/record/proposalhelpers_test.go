package record

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// docketByDispatch writes the docketing `dispatch next` writes on gap — under the chair, in a write of
// its own, through the write path's validation — with cause and, for a PROPOSAL docketing, the
// proposal it holds. A fixture that needs a gap before the bench, and is not about how the plan came to
// docket it, writes it this way. It returns the docketing's id.
func docketByDispatch(t *testing.T, run Run, gap string, cause recordpb.DocketingCause, motion string) (string, error) {
	t.Helper()
	return docketByDispatchAs(t, run, NewID("docketing"), gap, cause, motion)
}

// docketByDispatchAs is docketByDispatch with the docketing's id given, for a fixture that names it.
func docketByDispatchAs(t *testing.T, run Run, id, gap string, cause recordpb.DocketingCause, motion string) (string, error) {
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
	d := &recordpb.Docketing{DocketingId: proto.String(id), GapId: proto.String(gap), Cause: cause.Enum(), Basis: proto.String(impasseBasis(gap))}
	if motion != "" {
		d.MotionId = proto.String(motion)
	}
	ev := &Event{}
	typ, err := recordpb.SetBody(ev, d)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writeIn(tx.Tx, ev, chairSeat, typ, d, func(*sql.Tx) error {
		return validateAct(writeView{q: tx, files: filesOf(run)}, chairSeat, typ, d)
	}, stamp); err != nil {
		return "", err
	}
	return d.GetDocketingId(), tx.Commit()
}

// mustDocketByDispatch is docketByDispatch at impasse, failing the test on a refusal.
func mustDocketByDispatch(t *testing.T, run Run, gap string) string {
	t.Helper()
	id, err := docketByDispatch(t, run, gap, recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, "")
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// fileProposal files m, a proposal with its act, as blue's responding seat — the author of a report
// subject, the one party that files on a gap — after the one counted exchange on its gap every
// proposal needs (requireExchangedOn).
func fileProposal(t *testing.T, run Run, m *recordpb.Motion) {
	t.Helper()
	exchangeOn(t, run, m.GetDocket().GetGapId())
	mustAppend(t, Identity{Run: run, SeatID: blueRespondSeat}, m)
}

// docketProposal writes the PROPOSAL docketing holding proposal motion on gap, as the end of its
// assignee's window does, and returns its id.
func docketProposal(t *testing.T, run Run, gap, motion string) string {
	t.Helper()
	id, err := docketByDispatch(t, run, gap, recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, motion)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// docketGaps is the gaps a plan dockets (Plan.ToDocket), in its order.
func docketGaps(p Plan) []string {
	out := []string{}
	for _, d := range p.ToDocket {
		out = append(out, d.gapID)
	}
	return out
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
