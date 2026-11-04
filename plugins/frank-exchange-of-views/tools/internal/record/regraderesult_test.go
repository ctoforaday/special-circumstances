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

// THE BENCH LOCK'S FOLD READS THE GRADES THE GAP VIEW READS. benchSetAxesOf folds the lock over the
// regrades and takes each axis's grade from the result the regrade stored; the gap view takes its
// current grades from gap_grade's standing row. The two must agree on every axis of every gap, or the
// lock refuses a move against a grade the board does not show.
func TestTheBenchLockFoldAgreesWithTheGapView(t *testing.T) {
	run, lens := liveRun(t)
	appendOK(t, lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_LOW))
	appendOK(t, lens, liveMint("G2", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_HIGH))
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("b")})
	if _, _, err := RegisterSeat(lens, "", ""); err != nil { // a sitting holds one regrade per gap
		t.Fatal(err)
	}
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G1"), Impact: recordtest.P(recordpb.Grade_GRADE_LOW),
		ComplexityCost: recordtest.P(recordpb.Grade_GRADE_CERTAIN), Basis: proto.String("b")})
	appendOK(t, lens, &recordpb.Regrade{GapId: proto.String("G2"), Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("b")})

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open: %v", err)
	}
	axes, err := benchSetAxesAt(writeView{q: db})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT "gap_id", "current_severity", "current_likelihood", "current_impact", "current_complexity_cost" FROM "gap"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var id string
		var v [4]sql.NullString
		if err := rows.Scan(&id, &v[0], &v[1], &v[2], &v[3]); err != nil {
			t.Fatal(err)
		}
		n++
		for k := range v {
			if got := recordpb.Word(axes[id][k].grade); got != v[k].String {
				t.Errorf("gap %s's %s: the bench lock's fold reads %q, the gap view %q", id, recordpb.Word(gradeAxes[k]), got, v[k].String)
			}
		}
	}
	if n != 2 {
		t.Fatalf("the gap view holds %d gaps, want 2", n)
	}
}
