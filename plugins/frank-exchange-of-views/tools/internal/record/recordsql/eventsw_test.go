package recordsql_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE SITTING ORDINAL IS REGISTER-INCLUSIVE AND PER SEAT.
//
// A seat that registers, acts, registers again and acts again has two sittings, and the second
// register IS the second sitting — not the act after it. A different seat registering in between
// moves nothing: the window is partitioned by seat. This is the property that lets a seat id drop
// `-r<N>` — the record answers "which sitting" from the registers themselves, and nothing the seat
// types can put it on the wrong one.
func TestTheSittingOrdinalIsRegisterInclusiveAndPerSeat(t *testing.T) {
	dir := t.TempDir()
	reg := &recordpb.Register{}
	obs := &recordpb.Observe{Text: recordtest.P("x")}
	recordtest.Seed(t, dir,
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:register:#1", reg),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:observe:#1", obs),
		recordtest.At(t, "red-lens-logic", "red-lens-logic:register:#1", reg),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:register:#2", reg),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:observe:#2", obs),
	)
	db, err := recordsql.Open(runtest.Open(t, dir).Dir() + "/records/record.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "seat_id","type","sitting" FROM "events_w" ORDER BY "id"`)
	if err != nil {
		t.Fatalf("events_w did not answer: %v", err)
	}
	defer rows.Close()
	want := []struct {
		seat, typ string
		sitting   int
	}{
		{"red-lens-evidence", "register", 1},
		{"red-lens-evidence", "observe", 1},
		{"red-lens-logic", "register", 1}, // its own first, unmoved by evidence's
		{"red-lens-evidence", "register", 2},
		{"red-lens-evidence", "observe", 2},
	}
	i := 0
	for rows.Next() {
		var seat, typ string
		var sitting int
		if err := rows.Scan(&seat, &typ, &sitting); err != nil {
			t.Fatal(err)
		}
		if i >= len(want) {
			t.Fatalf("more rows than seeded: %s %s", seat, typ)
		}
		if seat != want[i].seat || typ != want[i].typ || sitting != want[i].sitting {
			t.Errorf("row %d: %s %s sitting %d, want %s %s sitting %d", i, seat, typ, sitting, want[i].seat, want[i].typ, want[i].sitting)
		}
		i++
	}
	if i != len(want) {
		t.Fatalf("events_w returned %d rows for %d events — a window that drops rows is a reader that undercounts", i, len(want))
	}
}

// THE EPOCH COUNTS CHAIR REGISTERS FOR EVERY ROW, WHOEVER WROTE IT.
//
// A lane's draft before the chair has ever sat is epoch 0. A lens act after the chair's first
// register is epoch 1 even though the lens never registered as the chair. The bench's act after the
// chair's second register is epoch 2. That is the whole reason it is a second window and not the
// sitting ordinal: the bucket readers ask "which dispatch cycle", and the answer must be the same
// for every seat's row in that cycle.
func TestTheEpochCountsChairRegistersForEveryRow(t *testing.T) {
	dir := t.TempDir()
	reg := &recordpb.Register{}
	obs := &recordpb.Observe{Text: recordtest.P("x")}
	recordtest.Seed(t, dir,
		recordtest.At(t, "blue-lane-1", "blue-lane-1:register:#1", reg),
		recordtest.At(t, "blue-lane-1", "blue-lane-1:observe:#1", obs),
		recordtest.At(t, "red-chair", "red-chair:register:#1", reg),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:register:#1", reg),
		recordtest.At(t, "red-lens-evidence", "red-lens-evidence:observe:#1", obs),
		recordtest.At(t, "red-chair", "red-chair:register:#2", reg),
		recordtest.At(t, "judge", "judge:register:#1", reg),
	)
	db, err := recordsql.Open(runtest.Open(t, dir).Dir() + "/records/record.db")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT "seat_id","type","epoch" FROM "events_w" ORDER BY "id"`)
	if err != nil {
		t.Fatalf("events_w did not answer: %v", err)
	}
	defer rows.Close()
	want := []int{0, 0, 1, 1, 1, 2, 2}
	i := 0
	for rows.Next() {
		var seat, typ string
		var epoch int
		if err := rows.Scan(&seat, &typ, &epoch); err != nil {
			t.Fatal(err)
		}
		if i >= len(want) {
			t.Fatalf("more rows than seeded")
		}
		if epoch != want[i] {
			t.Errorf("row %d (%s %s): epoch %d, want %d", i, seat, typ, epoch, want[i])
		}
		i++
	}
	if i != len(want) {
		t.Fatalf("events_w returned %d rows for %d events", i, len(want))
	}
}
