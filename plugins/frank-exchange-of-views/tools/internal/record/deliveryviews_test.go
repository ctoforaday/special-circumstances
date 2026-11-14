package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

func deliveriesT(t *testing.T, run Run) []recordsql.DeliveryRow {
	t.Helper()
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := recordsql.ReportDeliveries(db)
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

// A LENS'S FIRST TWO SAT-THROUGH DELIVERIES ARE WHOLE — its first, then its full re-read — and every
// later one is the changes since its pin; a delivery in a sitting it did not sit through does not count
// toward the two.
func TestReportDeliveryWholeIsTheFirstTwoSatThrough(t *testing.T) {
	b := retBoard(t).sitHooked(2, evLens, "a1", false).ingest() // not sat through: a later opening, no stop
	b.sitHooked(b.at(), evLens, "a2", true).ingest()
	b.sitHooked(b.at(), evLens, "a3", true).ingest()
	b.sitHooked(b.at(), evLens, "a4", true)
	rows := deliveriesT(t, b.seed())
	var whole, sat []bool
	for _, r := range rows {
		if r.SeatID == evLens {
			whole, sat = append(whole, r.Whole), append(sat, r.SatThrough)
		}
	}
	if want := []bool{true, true, true, false}; len(whole) != 4 || whole[0] != want[0] || whole[1] != want[1] || whole[2] != want[2] || whole[3] != want[3] {
		t.Fatalf("whole = %v, want %v (sat through %v)", whole, want, sat)
	}
	if sat[0] {
		t.Fatalf("the first sitting has a later opening and no stop: sat through %v", sat)
	}
}

// A CHANGE THE CHAIR HAS BORNE IS NOT AN AREA CHANGE: area_change lists the head moves after both the
// lens's pin and the highest change stamped on the latest spot-check naming it.
func TestAreaChangeIsTheUnborneHeadMovesAfterThePin(t *testing.T) {
	areaChanges := func(run Run) []int64 {
		db, err := openRunForRead(run)
		if err != nil {
			t.Fatal(err)
		}
		rs, err := db.Query(`SELECT "change" FROM "area_change" WHERE "seat_id" = ? ORDER BY "change"`, evLens)
		if err != nil {
			t.Fatal(err)
		}
		defer rs.Close()
		var out []int64
		for rs.Next() {
			var c int64
			if err := rs.Scan(&c); err != nil {
				t.Fatal(err)
			}
			out = append(out, c)
		}
		return out
	}
	board := func(spot bool) (*stage, int64, int64, int64) {
		b := retBoard(t).sitHooked(2, evLens, "a1", true).ingest()
		first := b.head()
		b.ingest()
		second := b.head()
		if spot {
			// Seeded as the write stamps it: the area's pin and the changes behind it.
			b.add("red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Pins: []int64{2}, Changes: []int64{first, second}, Reason: proto.String("read")})
		}
		b.ingest()
		return b, first, second, b.head()
	}
	b, first, second, third := board(false)
	if got := areaChanges(b.seed()); len(got) != 3 || got[0] != first || got[1] != second || got[2] != third {
		t.Fatalf("with no spot-check: %v, want [%d %d %d]", got, first, second, third)
	}
	b, _, _, third = board(true)
	if got := areaChanges(b.seed()); len(got) != 1 || got[0] != third {
		t.Fatalf("after a spot-check that stamped the first two: %v, want [%d]", got, third)
	}
}
