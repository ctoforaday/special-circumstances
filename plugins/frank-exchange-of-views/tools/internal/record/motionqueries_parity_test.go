package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// inquiryRulingFold is the fold InquiryRuling replaced, kept HERE as the parity oracle: the
// query and the fold read the same record, and this test refuses to let them disagree. It states
// the write's rule, FIRST-WINS: the first ruling on the inquiry is its answer. A record the fold
// cannot read fails the test: "" is the fold's answer for an inquiry red has not ruled.
func inquiryRulingFold(t *testing.T, run Run, inquiryID string) string {
	t.Helper()
	b, err := FamilyOf(run)
	if err != nil {
		t.Fatalf("the fold's own read of the record: %v", err)
	}
	ruling := ""
	for _, e := range b.Events {
		mr, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || mr.Subject() != recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY || mr.GetMotionId() != inquiryID {
			continue
		}
		if d, isInquiry := mr.GetRuling().(*recordpb.MotionRule_Inquiry); isInquiry {
			ruling = recordpb.Word(d.Inquiry)
		}
		break
	}
	return ruling
}

// THE MOTION READERS, HELD AGAINST THE FOLDS THEY REPLACED — step 4's third group: the
// ask-and-answer joins that used to be hand-written per reader (the eight-reader defect
// views.go's motion_state documents). The fixture walks one motion through its whole
// lifecycle — filed, ruled — asserting each guard's answer at each state, and rules
// one inquiry to hold InquiryRuling to its fold.
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

	// An inquiry: subject resolution falls through to the inquiry, and the ruling read
	// agrees with the fold at every state.
	if _, err := Append(blue, &recordpb.Inquiry{InquiryId: proto.String("Q1"),
		Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(), Question: proto.String("a question")}); err != nil {
		t.Fatal(err)
	}
	if got, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY, "Q1"); err != nil || got != "inquiry" {
		t.Errorf("RequireMotionSubjectRef(Q1) = (%q, %v) — an inquiry IS an inquiry motion by construction", got, err)
	}
	if _, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY, "Z9"); err == nil {
		t.Error("RequireMotionSubjectRef accepted an id that names no inquiry")
	}
	if got, err := InquiryRuling(run, "Q1"); err != nil || got != inquiryRulingFold(t, run, "Q1") || got != "" {
		t.Errorf("InquiryRuling before any ruling = (%q, %v), fold says %q", got, err, inquiryRulingFold(t, run, "Q1"))
	}
	if _, err := Append(red, &recordpb.MotionRule{
		MotionId: proto.String("Q1"),
		Opinion:  proto.String("worth the run's time"),
		Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE},
	}); err != nil {
		t.Fatal(err)
	}
	if got, err := InquiryRuling(run, "Q1"); err != nil || got != inquiryRulingFold(t, run, "Q1") || got != "out_of_scope" {
		t.Errorf("InquiryRuling = (%q, %v), fold says %q, want the schema's word out_of_scope", got, err, inquiryRulingFold(t, run, "Q1"))
	}
}
