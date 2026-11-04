package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE MOTION READERS, HELD AGAINST THE FOLDS THEY REPLACED — step 4's third group: the
// ask-and-answer joins that used to be hand-written per reader (the eight-reader defect
// views.go's motion_state documents). The fixture walks one motion through its whole
// lifecycle — filed, ruled — asserting each guard's answer at each state.
func TestMotionQueriesAgreeWithTheFoldsTheyReplaced(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-chair"}
	blue := Identity{Run: run, SeatID: "blue-respond"}

	if _, err := Append(red, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId:           proto.String("G1"),
		Class:           recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(),
		Problem:         proto.String("p"),
		RequiredFix:     proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := Append(blue, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Basis:    proto.String("severity is understated"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
			GapId:     proto.String("G1"),
			Dimension: recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY.Enum(),
			Proposed:  recordtest.P(recordpb.Grade_GRADE_HIGH),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if got, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M1"); err != nil || got != "grade" {
		t.Errorf("RequireMotionSubjectRef on a filed motion = (%q, %v), want it resolved as filed: grade", got, err)
	}
	if _, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M9"); err == nil {
		t.Error("RequireMotionSubjectRef accepted a motion no filing created")
	}
	// The first-wins guards read through the handle they are given — the writing transaction, on a
	// write; the run's handle here.
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}

	// Unruled: filing exists, no answer yet.
	if err := RequireUnruledMotion(db, "M1", ""); err != nil {
		t.Errorf("RequireUnruledMotion before any ruling = %v", err)
	}

	if _, err := Append(red, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Opinion:  proto.String("the grade stands"),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED},
	}); err != nil {
		t.Fatal(err)
	}
	// Ruled: the second-ruling refusal quotes the FIRST ruling's word and ruler.
	err = RequireUnruledMotion(db, "M1", "")
	if err == nil || !strings.Contains(err.Error(), `ruled "rejected" by red-chair`) {
		t.Errorf("RequireUnruledMotion after a ruling = %v, want the first ruling quoted", err)
	}
}
