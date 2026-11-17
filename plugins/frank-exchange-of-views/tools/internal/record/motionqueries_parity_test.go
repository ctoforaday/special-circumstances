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
	mustAppend(t, Identity{Run: run, SeatID: HarnessSeat}, castBody([]string{evLens, "red-chair", "blue-respond", "judge"}, nil))

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

	// Blue's proposal is admitted once G1 has had its one exchange.
	exchangeOn(t, run, "G1")
	if _, err := Append(blue, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Basis:    proto.String("severity is understated"),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
		Act: &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{
			GapId:    proto.String("G1"),
			Severity: recordtest.P(recordpb.Grade_GRADE_HIGH),
			Basis:    proto.String("severity is understated"),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	if got, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_DOCKET, "M1"); err != nil || got != "docket" {
		t.Errorf("RequireMotionSubjectRef on a filed motion = (%q, %v), want it resolved as filed: docket", got, err)
	}
	if _, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_DOCKET, "M9"); err == nil {
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

	if _, err := Append(convene(t, run, "G1"), actRule("M1", false, "judge")); err != nil {
		t.Fatal(err)
	}
	// Ruled: the second-ruling refusal quotes the FIRST ruling's word and ruler.
	err = RequireUnruledMotion(db, "M1", "")
	if err == nil || !strings.Contains(err.Error(), `ruled "rejected" by judge`) {
		t.Errorf("RequireUnruledMotion after a ruling = %v, want the first ruling quoted", err)
	}
}
