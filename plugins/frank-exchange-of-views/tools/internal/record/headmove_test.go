package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A RETIRE DOES NOT MOVE THE REPORT HEAD (gblock 2026-11-20): the anchor leaves and its text stays,
// so the head stays at the last edit or part, a lens that read the report there is not readied again,
// and area_change holds nothing for it. The anchor fact reaches seats through their work lists.
func TestARetireDoesNotMoveTheReportHeadOrReadyTheLensBehindIt(t *testing.T) {
	plan0 := planOf(t, retBoard(t).sit(2, evLens))
	if lensReady(plan0) {
		t.Fatalf("a lens that has read the report at the head is not ready before the retire: %+v", plan0)
	}
	before := plan0.Head
	b := retBoard(t).sit(2, evLens).retire("C1")
	retire := int64(b.n)
	run := b.seed()

	var head int64
	if err := readSnapshot(run, func(q recordsql.Querier) (err error) { head, err = ReportHeadAt(q); return }); err != nil {
		t.Fatal(err)
	}
	if head == retire || head != before {
		t.Fatalf("the report head is %d, want it where it stood before the retire at %d: %d", head, retire, before)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Head != before || strings.Contains(whyOf(plan), "behind head") {
		t.Fatalf("a retire moves no head, so the plan readies no lens behind one: %+v", plan)
	}

	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM "area_change" WHERE "seat_id" = ?`, evLens).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("area_change holds %d changes for %s after a retire alone, want none", n, evLens)
	}
}

// A RETIRE BEHIND A RETIRED LENS'S PIN IS NO STALE AREA, AND NO SPOT-CHECK STAMPS IT: the chair bears
// changes to the report's text, and a retire changes none.
func TestARetireBehindARetiredLensIsNoStaleAreaChange(t *testing.T) {
	b := retBoard(t).sit(2, evLens).sit(2, evLens).sit(2, evLens).retire("C1")
	plan := planOf(t, b)
	if len(plan.StaleAreas) != 0 {
		t.Fatalf("stale_areas = %+v, want none: a retire moves no head", plan.StaleAreas)
	}
	spot := &recordpb.SpotCheck{Areas: []string{evLens}, Reason: proto.String("read the change since the pin")}
	if err := readSnapshot(mustRun(t, b.runDir), func(q recordsql.Querier) error { return stampSpotCheckChanges(q, spot) }); err != nil {
		t.Fatal(err)
	}
	if len(spot.GetChanges()) != 0 {
		t.Fatalf("the spot-check stamped changes %v, want none", spot.GetChanges())
	}
}
