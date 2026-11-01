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

// THE SQL AND THE GO PREDICATE ANSWER THE SAME RECORD THE SAME WAY (#1040, #1151).
//
// They are one answer now rather than two that agree: the Go predicate decides at the WRITE and the
// database stores what it decided, so the SQL below reads `sitting_id` and has nothing of its own to
// get wrong. What this fixture still holds level is the two ENDS — that the SQL scoped to a sitting
// answers what the loader's window index carries for each act.
//
// THE UNREADABLE REGISTER IS NO LONGER A CASE, and its absence is the finding rather than an
// omission. A register with no decodable body could only reach the database as a forged row, which
// is why six copies of one predicate could answer it two ways with every gate green. A register
// carries its sitting now and the envelope CHECK refuses one that does not, so the shape is
// unrepresentable; the subtest below is what proves the refusal rather than assuming it.
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
			&recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("sitting 1's position")}), stamp(3)),
	}
	for i, ev := range evs {
		if _, err := recordsql.Insert(db, ev); err != nil {
			t.Fatalf("seeding event %d: %v", i, err)
		}
	}
	// THE FORGERY IS REFUSED BY THE RECORD ITSELF. A bare register row — no `register` row joining
	// it, which is what an undecodable body IS on this side of the fence — names no sitting, and a
	// register that names no sitting cannot be written.
	forgedKey := "blue-respond:register:#2"
	_, forgeErr := db.Exec(`INSERT INTO "events" ("seat_id", "ts", "type", "key") VALUES (?, ?, ?, ?)`,
		"blue-respond", stamp(5), "register", forgedKey)
	if forgeErr == nil {
		t.Fatal("a register with no sitting went onto the record — the envelope CHECK is what makes the two readings one")
	}
	// A real second sitting, opened by a register.
	opening2 := recordtest.Stamped(recordtest.At(t, "blue-respond", forgedKey,
		&recordpb.Register{AgentId: proto.String("blue-c")}), stamp(5))
	if _, err := recordsql.Insert(db, opening2); err != nil {
		t.Fatal(err)
	}
	after := recordtest.Stamped(recordtest.At(t, "blue-respond", "blue-respond:position:#2",
		&recordpb.Position{PositionId: proto.String(NewID("position")), Text: proto.String("sitting 2's position")}), stamp(6))
	if _, err := recordsql.Insert(db, after); err != nil {
		t.Fatal(err)
	}
	evs = append(evs, opening2, after)

	// THE LOADED SIDE: the sitting each act is attributed to, and the id of the seat's latest
	// opening, as the window index carries them. The ids the database assigned are 1..len(evs) in
	// insertion order.
	loaded, ws, err := recordsql.Events(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != len(evs) {
		t.Fatalf("loaded %d events, seeded %d", len(loaded), len(evs))
	}
	win := windowIndexOf(loaded, ws)
	sittings := make([]int, len(loaded))
	for i, e := range loaded {
		sittings[i] = win.Of(e).Sitting
	}
	latestOpening := win.LatestSittingOf("blue-respond")

	t.Run("oncePerSittingSQL", func(t *testing.T) {
		// The standing position is the one filed in the seat's CURRENT sitting — the sitting the
		// forged register opened, not the one before it.
		var stands string
		if err := db.QueryRow(oncePerSittingSQL, "blue-respond", "position").Scan(&stands); err != nil {
			t.Fatal(err)
		}
		if stands != after.GetKey() {
			t.Errorf("the standing position this sitting is %q, want %q — the window opens at the register the Go predicate opens a sitting at", stands, after.GetKey())
		}
	})

	// NO QUERY STATES THE PREDICATE ANY MORE. A sitting is a stored field (#1151), so a query that
	// scopes to one names `sittings` or `sitting_id` and reads what the write path decided. A query
	// that went back to deciding for itself — counting registers — would still pass the assertions
	// above on this fixture and would be free to drift on the next change.
	t.Run("no query re-derives what opens a sitting", func(t *testing.T) {
		for name, q := range map[string]string{
			"oncePerSittingSQL": oncePerSittingSQL,
		} {
			if !strings.Contains(q, `"sittings"`) && !strings.Contains(q, `"sitting_id"`) {
				t.Errorf("%s scopes to a sitting without reading the stored one:\n%s", name, q)
			}
			for _, tell := range []string{"'register'"} {
				if strings.Contains(q, tell) {
					t.Errorf("%s decides for itself what opens a sitting (%q), instead of reading sitting_id:\n%s", name, tell, q)
				}
			}
		}
	})

	t.Run("the sittings view and the window index agree", func(t *testing.T) {
		var got int64
		if err := db.QueryRow(`SELECT max("id") FROM "sittings" WHERE "seat_id" = ?`, "blue-respond").Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != latestOpening {
			t.Errorf("the sittings view's latest sitting of blue-respond is event %d; the window index's is %d", got, latestOpening)
		}
	})
}

// A DISPATCH IS BRACKETED *AND* REGISTERS, AND THAT IS ONE SITTING (#1151).
//
// SubagentStart writes the bracket; the seat then runs `register` in the turn that bracket opened.
// Both are openings on their own terms — either can be the only one a dispatch gets — so admitting
// the bracket without joining the pair doubled every seat's ordinal. It shipped that way: a live run
// carried 37 registers against 33 resolved brackets, every one of them a pair under one agent id.
//
// The second dispatch is the other half of the gate. A fix that made a register never open a sitting
// would pass the first assertion and lose every resumed dispatch, which carries no bracket at all.
func TestABracketAndItsOwnRegisterAreOneSitting(t *testing.T) {
	const seat = "red-lens-evidence"
	bracket := func(agent string) *Event {
		return recordtest.At(t, HarnessSeat, "harness:sitting_open:"+agent, &recordpb.SittingOpen{
			AgentId:   proto.String(agent),
			AgentType: proto.String("frank-exchange-of-views:red-lens-evidence"),
			SeatId:    proto.String(seat),
		})
	}
	register := func(agent, key string) *Event {
		return recordtest.At(t, seat, key, &recordpb.Register{AgentId: proto.String(agent)})
	}
	for _, c := range []struct {
		name string
		evs  []*Event
		want int
	}{
		{"one bracketed dispatch that also registers", []*Event{
			bracket("a1"), register("a1", seat+":register:#1")}, 1},
		{"two of them", []*Event{
			bracket("a1"), register("a1", seat+":register:#1"),
			bracket("a2"), register("a2", seat+":register:#2")}, 2},
		{"a resume, which carries no bracket", []*Event{
			bracket("a1"), register("a1", seat+":register:#1"),
			register("a2", seat+":register:#2")}, 2},
		{"a bracket the seat never answered", []*Event{bracket("a1")}, 1},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			recordtest.Seed(t, dir, c.evs...)
			db, err := recordsql.Open(filepath.Join(dir, "records", "record.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			var stored, view int
			if err := db.QueryRow(`SELECT count(*) FROM "sittings" WHERE "seat_id" = ?`, seat).Scan(&stored); err != nil {
				t.Fatal(err)
			}
			// Asked of the rows BELONGING to the seat's sittings, not of the rows the seat wrote:
			// a bracket is harness-seated, so a sitting the seat never answered has none of its own.
			if err := db.QueryRow(`SELECT COALESCE(max(w."sitting"), 0) FROM "events_w" w
				  JOIN "events" e ON e."id" = w."id"
				  JOIN "sittings" s ON s."id" = e."sitting_id"
				 WHERE s."seat_id" = ?`, seat).Scan(&view); err != nil {
				t.Fatal(err)
			}
			if stored != c.want || view != c.want {
				t.Errorf("the record holds %d sitting(s) and events_w reads %d; want %d — a bracket and the register of its own agent are one sitting, and a register without one is still a sitting", stored, view, c.want)
			}
		})
	}
}
