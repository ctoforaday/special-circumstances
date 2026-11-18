package record

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A RETIRE MOVES THE REPORT HEAD (gblock 2026-11-18): it takes text out of the report, so the head is
// the retire, a lens whose pin is behind it is ready again, and the retire is a change behind that
// lens's pin.
func TestARetireMovesTheReportHeadAndReadiesTheLensBehindIt(t *testing.T) {
	b := retBoard(t).readFully(2, evLens)
	if plan := planOf(t, b); lensReady(plan) {
		t.Fatalf("a lens that has read the report at the head is not ready before the retire: %+v", plan)
	}
	b = retBoard(t).readFully(2, evLens).retire("C1")
	retire := int64(b.n)
	run := b.seed()

	var head int64
	if err := readSnapshot(run, func(q recordsql.Querier) (err error) { head, err = ReportHeadAt(q); return }); err != nil {
		t.Fatal(err)
	}
	if head != retire {
		t.Fatalf("the report head is %d, want the retire at %d", head, retire)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Head != retire || !lensReady(plan) || !strings.Contains(whyOf(plan), "behind head "+strconv.FormatInt(retire, 10)) {
		t.Fatalf("the retire moved the head past the lens's pin, so the plan readies it at head %d: %+v", retire, plan)
	}

	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var changes []int64
	rs, err := db.Query(`SELECT "change" FROM "area_change" WHERE "seat_id" = ? ORDER BY "change"`, evLens)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Close()
	for rs.Next() {
		var c int64
		if err := rs.Scan(&c); err != nil {
			t.Fatal(err)
		}
		changes = append(changes, c)
	}
	if !slices.Equal(changes, []int64{retire}) {
		t.Fatalf("area_change for %s = %v, want the retire behind its pin, [%d]", evLens, changes, retire)
	}
}

// A RETIRE BEHIND A RETIRED LENS'S PIN MAKES ITS AREA STALE, AND THE SPOT-CHECK STAMPS IT: the change
// the chair bears for the area is the retire, so the area clears once a spot-check names it.
func TestARetireBehindARetiredLensIsAStaleAreaChangeTheSpotCheckStamps(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).sit(2, evLens).retire("C1")
	retire := int64(b.n)
	plan := planOf(t, b)
	if len(plan.StaleAreas) != 1 || plan.StaleAreas[0] != (StaleArea{SeatID: evLens, Pin: 2}) {
		t.Fatalf("stale_areas = %+v, want %s at pin 2 behind the retire at %d", plan.StaleAreas, evLens, retire)
	}
	spot := &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the change since the pin")}
	if err := readSnapshot(mustRun(t, b.runDir), func(q recordsql.Querier) error { return stampSpotCheckChanges(q, spot) }); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(spot.GetChanges(), []int64{retire}) {
		t.Fatalf("the spot-check stamped changes %v, want the retire [%d]", spot.GetChanges(), retire)
	}
}
