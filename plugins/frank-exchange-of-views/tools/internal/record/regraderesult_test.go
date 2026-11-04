package record

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE WRITE STAMPS EACH REGRADE WITH THE GAP'S GRADES ONCE IT STANDS: the axes it names, and on the
// rest the grades the gap held before it — the mint's for the first regrade, the previous regrade's
// result after that. An axis nothing graded stays absent, and a seat may not supply the result.
func TestARegradeStoresTheGradesItLeaves(t *testing.T) {
	run, lens := liveRun(t)
	appendOK(t, lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_LOW))
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("b")})
	resit := func() {
		t.Helper()
		if _, _, err := RegisterSeat(lens, "", ""); err != nil { // a sitting holds one regrade per gap
			t.Fatal(err)
		}
	}
	resit()
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Impact: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("b")})
	resit()

	_, err := Append(lens, &recordpb.Regrade{GapId: proto.String("G1"), Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH),
		ResultLikelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("b")})
	mustRefuse(t, err, "stamped by the tool")

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open: %v", err)
	}
	rows, err := db.Query(`SELECT "result_severity", "result_likelihood", "result_impact", "result_complexity_cost"
	  FROM "regrade" ORDER BY "event_id"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	want := [][4]string{
		{"high", "medium", "medium", ""}, // severity named; the mint's on the rest
		{"high", "medium", "low", ""},    // impact named; the first regrade's result on the rest
	}
	var got [][4]string
	for rows.Next() {
		var r [4]sql.NullString
		if err := rows.Scan(&r[0], &r[1], &r[2], &r[3]); err != nil {
			t.Fatal(err)
		}
		got = append(got, [4]string{r[0].String, r[1].String, r[2].String, r[3].String})
		if r[3].Valid {
			t.Errorf("result_complexity_cost is %q on a gap nothing graded on that axis; want it absent", r[3].String)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("the record holds %d regrades, want %d (the seat-supplied one refused): %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("regrade %d stores %v, want %v", i+1, got[i], want[i])
		}
	}
}
