package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// forgedRegister is a register event with NO BODY.
//
// It has to be forged, and that is the point of it. recordpb.SetBody stamps the type from the body,
// so no fixture helper can express the disagreement; recordsql.Insert refuses an event with no body;
// and recordsql's loader refuses a `register` row whose body row is missing, so the shape cannot
// come back out of a database either. The degradation rule opensASitting states is therefore about
// a record this tree did not write — which is exactly the record no reader was ever compared on,
// and why six copies of one predicate could answer it two different ways with every gate green.
func forgedRegister(seat, key, ts string) *Event {
	return &recordpb.Event{
		SeatId: proto.String(seat),
		Ts:     proto.String(ts),
		Type:   recordpb.EventType_EVENT_TYPE_REGISTER.Enum(),
		Key:    proto.String(key),
	}
}

// forgedRegister stages one on the builder, for the tables that need the unreadable case.
func (b *stage) forgedRegister(seat string) *stage {
	b.n++
	b.evs = append(b.evs, forgedRegister(seat, fmt.Sprintf("%s:%d", seat, b.n),
		fmt.Sprintf("2026-01-01T00:00:%02dZ", b.n)))
	return b
}

// EVERY READER THAT ATTRIBUTES AN ACT AGREES A REGISTER IT CANNOT READ OPENS A SITTING (#1040).
//
// They did not. Five copies of the predicate failed open and three failed closed, and the two
// answers decided the same thing about the same event: which sitting an act belongs to. They share
// one definition now, so this fixture asks each of them and expects one answer — and the assertions
// are written so that flipping opensASitting reddens more than one of them.
func TestEveryAttributionReaderAgreesAnUnreadableRegisterOpensASitting(t *testing.T) {
	blue := func(key string, body proto.Message) *Event { return recordtest.At(t, "blue-respond", key, body) }
	evs := []*Event{
		recordtest.At(t, HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{"red-chair", "blue-respond"}}),
		recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}),
		recordtest.At(t, "red-chair", "red-chair:dispatch:d1", &recordpb.Dispatch{
			Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		blue("blue-respond:register:#1", &recordpb.Register{AgentId: proto.String("blue-a")}),
		blue("blue-respond:blue_edit:e1", &recordpb.BlueEdit{Answers: proto.String("G1")}),
		blue("blue-respond:revision:#1", &recordpb.Revision{Text: proto.String("what sitting 1 filed")}),
		recordtest.At(t, HarnessSeat, "harness:sitting_close:s1", &recordpb.SittingClose{AgentId: proto.String("blue-a")}),
		recordtest.At(t, "red-chair", "red-chair:dispatch:d2", &recordpb.Dispatch{
			Pin: proto.Int64(1), SeatId: proto.String("blue-respond"), GapIds: []string{"G1"}}),
		// The forgery: blue's second register, and nothing can read it.
		forgedRegister("blue-respond", "blue-respond:register:#2", "2026-01-01T00:01:00Z"),
		blue("blue-respond:blue_edit:e2", &recordpb.BlueEdit{Answers: proto.String("G1")}),
	}
	const forged, afterForged = 8, 9
	if _, ok := recordpb.BodyAs[*recordpb.Register](evs[forged]); ok {
		t.Fatal("the forgery carries a readable body, so this test asserts nothing")
	}

	t.Run("ActClock gives the acts after it a new sitting", func(t *testing.T) {
		var ac ActClock
		var got []int
		for _, e := range evs {
			got = append(got, ac.Advance(e).Sitting)
		}
		if got[forged] != 2 || got[afterForged] != 2 {
			t.Errorf("the forged register and the act after it are sitting #%d and #%d, want #2 and #2",
				got[forged], got[afterForged])
		}
	})

	t.Run("seatDidThisSitting starts the window at it", func(t *testing.T) {
		// The revision is sitting 1's. The forged register opened sitting 2, so sitting 2 has
		// filed none — which is the answer that inverts if the register stops opening one.
		if seatDidThisSitting(evs, "blue-respond", recordpb.EventType_EVENT_TYPE_REVISION) {
			t.Error("blue's sitting reads as having filed its revision, but the revision is the previous sitting's")
		}
	})

	t.Run("dispatchLedger counts it among the seat's sittings", func(t *testing.T) {
		seq := make([]int64, len(evs))
		for i := range seq {
			seq[i] = int64(i)
		}
		_, registers := dispatchLedger(evs, seq)
		want := []int64{3, forged}
		if got := registers["blue-respond"]; len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
			t.Errorf("blue's opening registers are at %v, want %v", got, want)
		}
	})

	t.Run("sittingCloser bounds the first sitting at it", func(t *testing.T) {
		seq := make([]int64, len(evs))
		for i := range seq {
			seq[i] = int64(i)
		}
		_, registers := dispatchLedger(evs, seq)
		c := sittingCloserOf(evs, seq, registers, WhileRunning)
		if _, end, closed := c.bounds("blue-respond", 3); !closed || end > forged {
			t.Errorf("blue's first sitting ends at %d (closed=%v), want it closed no later than the forged register at %d",
				end, closed, forged)
		}
	})

	t.Run("BlueSittings reads two sittings", func(t *testing.T) {
		ss := BlueSittings(evs, WhileRunning)
		if len(ss) != 2 {
			t.Fatalf("blue has %d sitting(s), want 2 — the forged register opens the second", len(ss))
		}
	})

	// THE CHAIR'S SPOT-CHECK WINDOW IS THE SAME QUESTION, on its own fixture because the window is
	// the chair's — and on a REPAIR rather than the forgery, because that is where this reader's
	// two possible answers differ. No verb can reach a chair's repair (checkRepair admits a blue
	// role only), so the record is staged by hand; the reading is the shared one so that it stays
	// right if that gate ever moves, which is what spotcheck.go already did for the same reason.
	t.Run("the chair's spot-check window opens at the register that opened its sitting", func(t *testing.T) {
		chair := []*Event{
			recordtest.At(t, HarnessSeat, "harness:cast", &recordpb.Cast{SeatIds: []string{"red-chair", "red-lens-logic"}}),
			recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}),
			recordtest.At(t, "red-chair", "red-chair:spot_check:#1", &recordpb.SpotCheck{
				Areas: []string{"red-lens-logic"}, Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")}),
			recordtest.At(t, "red-chair", "red-chair:register:#2", &recordpb.Register{
				RepairsSitting: proto.String("red-chair:register:#1")}),
		}
		ids := make([]int64, len(chair))
		for i := range ids {
			ids[i] = int64(i + 1)
		}
		if g := passLensGateOf(chair, ids, nil); !g.covered["red-lens-logic"] {
			t.Error("the chair's sitting reads as not covering red-lens-logic, but the spot-check is an act of the sitting its repair completes")
		}
	})
}

// THE CLAIM CHECK REFUSES A TARGET IT CANNOT READ, WHICH IS THE OTHER ANSWER AND IS MEANT TO BE
// (#1040). Attribution invents nothing and lets an unreadable register open a sitting; a repair
// CLAIM against that register would go on the record as a repair of a sitting nothing can bound, so
// the write path refuses it — and refuses it on the claimUnfounded branch, because telling the seat
// there is nothing to file would assert what the unreadable sitting owed.
func TestTheRepairClaimCheckRefusesARegisterItCannotRead(t *testing.T) {
	b := newStage(t)
	b.cast(evLens, "red-chair", "blue-respond").ingest().register("red-chair").
		dispatch(1, "blue-respond", "G1").forgedRegister("blue-respond")
	key := b.lastKey()

	for _, c := range []struct {
		name string
		call func() error
	}{
		{"named by the seat", func() error { return checkRepair(b.evs, "blue-respond", key) }},
		{"chosen by the tool", func() error { _, err := repairTarget(b.evs, "blue-respond"); return err }},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := c.call()
			if err == nil {
				t.Fatal("the repair was admitted against a register this binary cannot read")
			}
			if !strings.Contains(err.Error(), "whose body this binary cannot read") {
				t.Fatalf("the refusal does not name the unreadable target:\n%v", err)
			}
			if strings.Contains(err.Error(), RepairNothingToFile) {
				t.Errorf("the refusal puts the seat on the nothing-to-file branch, which asserts what the unreadable sitting owed:\n%v", err)
			}
		})
	}
}

// THE SQL AND THE GO PREDICATE ANSWER THE SAME RECORD THE SAME WAY, the unreadable register
// included (#1040).
//
// They cannot be joined by a round-trip: recordsql's loader REFUSES a register row with no body row
// ("an envelope with no body replays as a seat that did nothing"), so the Go readers never see one
// through a database. The SQL queries read the database directly and DO see it — the LEFT JOIN
// leaves repairs_sitting NULL — which makes SQL the only side of this that can meet the case at all.
// So the fixture is one record expressed twice, and what is compared is the ANSWER.
func TestTheSQLAndGoReadingsOfWhatOpensASittingAgree(t *testing.T) {
	dir := recordtest.TmpRun(t)
	if err := os.MkdirAll(filepath.Join(dir, "records"), 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
	if err != nil {
		t.Fatal(err)
	}

	stamp := func(i int) string { return fmt.Sprintf("2026-01-01T00:00:%02dZ", i) }
	opening := recordtest.Stamped(recordtest.At(t, "blue-respond", "blue-respond:register:#1",
		&recordpb.Register{AgentId: proto.String("blue-a")}), stamp(2))
	evs := []*Event{
		recordtest.Stamped(recordtest.At(t, "red-chair", "red-chair:register:#1", &recordpb.Register{}), stamp(1)),
		opening,
		recordtest.Stamped(recordtest.At(t, "blue-respond", "blue-respond:position:#1",
			&recordpb.Position{Text: proto.String("sitting 1's position")}), stamp(3)),
		// A repair: a register that does NOT open a sitting, so the two answers cannot coincide
		// by the fixture holding only one kind of register.
		recordtest.Stamped(recordtest.At(t, "blue-respond", "blue-respond:register:#2",
			&recordpb.Register{AgentId: proto.String("blue-b"), RepairsSitting: proto.String(opening.GetKey())}), stamp(4)),
	}
	for i, ev := range evs {
		if _, err := recordsql.Insert(db, ev); err != nil {
			t.Fatalf("seeding event %d: %v", i, err)
		}
	}
	// The forgery goes in as a bare row: no `register` row joins it, which is what an undecodable
	// body IS on this side of the fence.
	forgedKey := "blue-respond:register:#3"
	if _, err := db.Exec(`INSERT INTO "events" ("seat_id", "ts", "type", "key") VALUES (?, ?, ?, ?)`,
		"blue-respond", stamp(5), "register", forgedKey); err != nil {
		t.Fatal(err)
	}
	after := recordtest.Stamped(recordtest.At(t, "blue-respond", "blue-respond:position:#2",
		&recordpb.Position{Text: proto.String("sitting 2's position")}), stamp(6))
	if _, err := recordsql.Insert(db, after); err != nil {
		t.Fatal(err)
	}
	evs = append(evs, forgedRegister("blue-respond", forgedKey, stamp(5)), after)

	// THE GO SIDE: the sitting each act is attributed to, and the id of the seat's latest opening
	// register. The ids the database assigned are 1..len(evs) in insertion order.
	var ac ActClock
	sittings := make([]int, len(evs))
	latestOpening := int64(0)
	for i, e := range evs {
		sittings[i] = ac.Advance(e).Sitting
		if opensASitting(e) && e.GetSeatId() == "blue-respond" {
			latestOpening = int64(i + 1)
		}
	}

	t.Run("sittingBeforeAndNowSQL", func(t *testing.T) {
		// Bound as (seat, the act's event id, seat): the sitting that act was filed in, and the
		// sitting the seat is in now.
		for _, c := range []struct{ id, want int }{{3, sittings[2]}, {6, sittings[5]}} {
			var before, now int
			if err := db.QueryRow(sittingBeforeAndNowSQL, "blue-respond", c.id, "blue-respond").
				Scan(&before, &now); err != nil {
				t.Fatal(err)
			}
			if before != c.want {
				t.Errorf("SQL files event %d in sitting #%d; ActClock puts it in #%d", c.id, before, c.want)
			}
			if want := sittings[len(sittings)-1]; now != want {
				t.Errorf("SQL reads the seat as in sitting #%d now; ActClock reads #%d", now, want)
			}
		}
	})

	t.Run("oncePerSittingSQL", func(t *testing.T) {
		// The standing position is the one filed in the seat's CURRENT sitting — the sitting the
		// forged register opened, not the one before it.
		var stands string
		if err := db.QueryRow(oncePerSittingSQL, "blue-respond", "position", "blue-respond").Scan(&stands); err != nil {
			t.Fatal(err)
		}
		if stands != after.GetKey() {
			t.Errorf("the standing position this sitting is %q, want %q — the window opens at the register the Go predicate opens a sitting at", stands, after.GetKey())
		}
	})

	// ONE FRAGMENT, NOT THREE SPELLINGS. The queries agree with each other BY CONSTRUCTION only
	// while they are composed from the fragment; a query that re-inlines the predicate would still
	// pass the assertions above on this fixture and would be free to drift on the next change.
	t.Run("both queries are composed from the one fragment", func(t *testing.T) {
		for name, q := range map[string]string{
			"sittingBeforeAndNowSQL": sittingBeforeAndNowSQL,
			"oncePerSittingSQL":      oncePerSittingSQL,
		} {
			if !strings.Contains(q, openingRegistersOfSeatSQL) {
				t.Errorf("%s states the predicate itself instead of wrapping openingRegistersOfSeatSQL, so a change to what opens a sitting can reach one query and not the other:\n%s", name, q)
			}
		}
	})

	t.Run("openingRegistersOfSeatSQL", func(t *testing.T) {
		var got int64
		if err := db.QueryRow(`SELECT max("id") FROM (`+openingRegistersOfSeatSQL+`)`, "blue-respond").Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != latestOpening {
			t.Errorf("SQL's latest opening register of blue-respond is event %d; opensASitting's is %d", got, latestOpening)
		}
	})
}
