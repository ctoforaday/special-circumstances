package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// inquiryRulingFold is the fold InquiryRuling replaced, kept HERE as the parity oracle: the
// query and the fold read the same record, and this test refuses to let them disagree.
func inquiryRulingFold(run Run, inquiryID string) string {
	b, err := BoardState(run)
	if err != nil {
		return ""
	}
	ruling := ""
	for _, e := range b.Events {
		mr, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || mr.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION || mr.GetMotionId() != inquiryID {
			continue
		}
		ruling = ""
		if d, isDirection := mr.GetRuling().(*recordpb.MotionRule_Direction); isDirection {
			ruling = strings.ReplaceAll(recordpb.Word(d.Direction), "_", "-")
		}
	}
	return ruling
}

// THE MOTION READERS, HELD AGAINST THE FOLDS THEY REPLACED — step 4's third group: the
// ask-and-answer joins that used to be hand-written per reader (the eight-reader defect
// views.go's motion_state documents). The fixture walks one motion through its whole
// lifecycle — filed, ruled, appealed — asserting each guard's answer at each state, and rules
// one line of inquiry to hold InquiryRuling to its fold.
func TestMotionQueriesAgreeWithTheFoldsTheyReplaced(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-merge-r1", Round: 1}
	blue := Identity{Run: run, SeatID: "blue-respond-r1", Round: 1}

	if _, err := Append(red, &recordpb.Mint{
		GapId:           proto.String("R1-1"),
		Class:           proto.String("self-attestation"),
		Problem:         proto.String("p"),
		RequiredFix:     proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"),
		CheckKind:       recordpb.CheckKind_CHECK_KIND_DOCUMENT.Enum(),
		Likelihood:      recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Impact:          recordtest.P(recordpb.Grade_GRADE_MEDIUM),
	}); err != nil {
		t.Fatal(err)
	}

	if id, err := MintMotionID(run); err != nil || id != "M1" {
		t.Errorf("MintMotionID on an unfiled record = (%q, %v)", id, err)
	}
	if _, err := Append(blue, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Basis:    proto.String("severity is understated"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{
			GapId:     proto.String("R1-1"),
			Dimension: recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY.Enum(),
			Proposed:  recordtest.P(recordpb.Grade_GRADE_HIGH),
		}},
	}); err != nil {
		t.Fatal(err)
	}
	if id, err := MintMotionID(run); err != nil || id != "M2" {
		t.Errorf("MintMotionID after one filing = (%q, %v)", id, err)
	}

	if got, err := motionSubjectOf(run, "M1"); err != nil || got != "grade" {
		t.Errorf("motionSubjectOf(M1) = (%q, %v)", got, err)
	}
	if err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M1"); err != nil {
		t.Errorf("RequireMotionSubjectRef on a filed motion = %v", err)
	}
	if err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M9"); err == nil {
		t.Error("RequireMotionSubjectRef accepted a motion no filing created")
	}

	// Unruled: filing exists, no answer yet.
	if err := RequireUnruledMotion(run, "M1"); err != nil {
		t.Errorf("RequireUnruledMotion before any ruling = %v", err)
	}
	if err := RequireRuledMotion(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M1"); err == nil {
		t.Error("RequireRuledMotion found a ruling nobody made")
	}

	if _, err := Append(red, &recordpb.MotionRule{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Opinion:  proto.String("the grade stands"),
		Ruling:   &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED},
	}); err != nil {
		t.Fatal(err)
	}
	// Ruled: the second-ruling refusal quotes the FIRST ruling's word and ruler.
	err := RequireUnruledMotion(run, "M1")
	if err == nil || !strings.Contains(err.Error(), `ruled "rejected" by red-merge-r1`) {
		t.Errorf("RequireUnruledMotion after a ruling = %v, want the first ruling quoted", err)
	}
	if err := RequireRuledMotion(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M1"); err != nil {
		t.Errorf("RequireRuledMotion after a ruling = %v", err)
	}
	if err := RequireUnappealedMotion(run, "M1"); err != nil {
		t.Errorf("RequireUnappealedMotion before any appeal = %v", err)
	}

	if _, err := Append(blue, &recordpb.MotionAppeal{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Reason:   proto.String("the ruling reads past the argument"),
	}); err != nil {
		t.Fatal(err)
	}
	err = RequireUnappealedMotion(run, "M1")
	if err == nil || !strings.Contains(err.Error(), "blue-respond-r1") ||
		!strings.Contains(err.Error(), "the ruling reads past the argument") {
		t.Errorf("RequireUnappealedMotion after an appeal = %v, want the appeal quoted", err)
	}

	// A line of inquiry: subject resolution falls through to the avenue, and the ruling read
	// agrees with the fold at every state.
	if _, err := Append(blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("a direction")}); err != nil {
		t.Fatal(err)
	}
	if got, err := motionSubjectOf(run, "Q1"); err != nil || got != "inquiry" {
		t.Errorf("motionSubjectOf(Q1) = (%q, %v) — a line of inquiry IS a direction motion by construction", got, err)
	}
	if _, err := motionSubjectOf(run, "Z9"); err == nil {
		t.Error("motionSubjectOf accepted an id that names neither a motion nor a line of inquiry")
	}
	if got, want := InquiryRuling(run, "Q1"), inquiryRulingFold(run, "Q1"); got != want || got != "" {
		t.Errorf("InquiryRuling before any ruling = %q, fold says %q", got, want)
	}
	if _, err := Append(red, &recordpb.MotionRule{
		MotionId: proto.String("Q1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DIRECTION),
		Opinion:  proto.String("worth the run's time"),
		Ruling:   &recordpb.MotionRule_Direction{Direction: recordpb.DirectionRuling_DIRECTION_RULING_OUT_OF_SCOPE},
	}); err != nil {
		t.Fatal(err)
	}
	if got, want := InquiryRuling(run, "Q1"), inquiryRulingFold(run, "Q1"); got != want || got != "out-of-scope" {
		t.Errorf("InquiryRuling = %q, fold says %q, want the hyphen join of the schema's word", got, want)
	}
}
