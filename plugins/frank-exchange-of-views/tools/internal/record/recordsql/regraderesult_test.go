package recordsql

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE INSERT STAMPS THE COMPUTED GRADES OVER WHATEVER A REGRADE CARRIES. Each regrade below arrives
// with a deliberately wrong result on every axis; what the record holds — read back through the
// generated decode — is the mint overlaid with the regrades so far, and an axis nothing graded is
// absent rather than the wrong value it arrived with.
func TestARegradeIsStampedWithTheComputedGradesOverWhatItCarries(t *testing.T) {
	db := store(t)
	mintRow(t, db, 0, "G1", recordpb.CheckKind_CHECK_KIND_DOCUMENT) // severity, likelihood, impact medium; no complexity
	wrong := recordpb.Grade_GRADE_TRIVIAL
	for i, r := range []*recordpb.Regrade{
		{GapId: proto.String("G1"), Severity: recordpb.Grade_GRADE_HIGH.Enum()},
		{GapId: proto.String("G1"), Impact: recordpb.Grade_GRADE_LOW.Enum()},
	} {
		r.Basis = proto.String("moved")
		r.ResultSeverity, r.ResultLikelihood, r.ResultImpact, r.ResultComplexityCost = wrong.Enum(), wrong.Enum(), wrong.Enum(), wrong.Enum()
		if _, err := Insert(db, event(t, int32(1+i), recordpb.EventType_EVENT_TYPE_REGRADE, r)); err != nil {
			t.Fatal(err)
		}
	}
	evs, _, err := Events(db)
	if err != nil {
		t.Fatal(err)
	}
	m := recordpb.Grade_GRADE_MEDIUM
	want := [][4]recordpb.Grade{
		{recordpb.Grade_GRADE_HIGH, m, m, recordpb.Grade_GRADE_UNSPECIFIED},
		{recordpb.Grade_GRADE_HIGH, m, recordpb.Grade_GRADE_LOW, recordpb.Grade_GRADE_UNSPECIFIED},
	}
	var got [][4]recordpb.Grade
	for _, e := range evs {
		r := e.GetRegrade()
		if r == nil {
			continue
		}
		if r.ResultComplexityCost != nil {
			t.Errorf("regrade %d stores complexity %s on a gap nothing graded on that axis; want it absent", len(got)+1, r.GetResultComplexityCost())
		}
		got = append(got, [4]recordpb.Grade{r.GetResultSeverity(), r.GetResultLikelihood(), r.GetResultImpact(), r.GetResultComplexityCost()})
	}
	if len(got) != len(want) {
		t.Fatalf("read back %d regrades, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("regrade %d stores %v, want the computed %v", i+1, got[i], want[i])
		}
	}
}
