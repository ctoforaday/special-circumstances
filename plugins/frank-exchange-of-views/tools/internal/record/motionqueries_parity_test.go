package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// avenueRulingFold is the fold AvenueRuling replaced, kept HERE as the parity oracle: the
// query and the fold read the same record, and this test refuses to let them disagree. A record
// the fold cannot read fails the test: "" is the fold's answer for an avenue red has not ruled.
func avenueRulingFold(t *testing.T, run Run, avenueID string) string {
	t.Helper()
	b, err := FamilyOf(run)
	if err != nil {
		t.Fatalf("the fold's own read of the record: %v", err)
	}
	ruling := ""
	for _, e := range b.Events {
		mr, ok := recordpb.BodyAs[*recordpb.MotionRule](e)
		if !ok || mr.GetSubject() != recordpb.MotionSubject_MOTION_SUBJECT_AVENUE || mr.GetMotionId() != avenueID {
			continue
		}
		ruling = ""
		if d, isAvenue := mr.GetRuling().(*recordpb.MotionRule_Avenue); isAvenue {
			ruling = recordpb.Word(d.Avenue)
		}
	}
	return ruling
}

// THE MOTION READERS, HELD AGAINST THE FOLDS THEY REPLACED — step 4's third group: the
// ask-and-answer joins that used to be hand-written per reader (the eight-reader defect
// views.go's motion_state documents). The fixture walks one motion through its whole
// lifecycle — filed, ruled, appealed — asserting each guard's answer at each state, and rules
// one avenue to hold AvenueRuling to its fold.
func TestMotionQueriesAgreeWithTheFoldsTheyReplaced(t *testing.T) {
	runDir := newRun(t)
	run := mustRun(t, runDir)
	red := Identity{Run: run, SeatID: "red-chair"}
	blue := Identity{Run: run, SeatID: "blue-respond"}

	if _, err := Append(red, &recordpb.Mint{Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		GapId:           proto.String("G1"),
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

	if _, err := Append(blue, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
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
	if err := RequireUnruledMotion(db, "M1", "", ""); err != nil {
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
	err = RequireUnruledMotion(db, "M1", "", "")
	if err == nil || !strings.Contains(err.Error(), `ruled "rejected" by red-chair`) {
		t.Errorf("RequireUnruledMotion after a ruling = %v, want the first ruling quoted", err)
	}
	if err := RequireRuledMotion(run, recordpb.MotionSubject_MOTION_SUBJECT_GRADE, "M1"); err != nil {
		t.Errorf("RequireRuledMotion after a ruling = %v", err)
	}
	if err := RequireUnappealedMotion(db, "M1", "", ""); err != nil {
		t.Errorf("RequireUnappealedMotion before any appeal = %v", err)
	}

	if _, err := Append(blue, &recordpb.MotionAppeal{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Reason:   proto.String("the ruling reads past the argument"),
	}); err != nil {
		t.Fatal(err)
	}
	err = RequireUnappealedMotion(db, "M1", "", "")
	if err == nil || !strings.Contains(err.Error(), "blue-respond") ||
		!strings.Contains(err.Error(), "the ruling reads past the argument") {
		t.Errorf("RequireUnappealedMotion after an appeal = %v, want the appeal quoted", err)
	}

	// An avenue: subject resolution falls through to the avenue, and the ruling read
	// agrees with the fold at every state.
	if _, err := Append(blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("a direction")}); err != nil {
		t.Fatal(err)
	}
	if got, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_AVENUE, "Q1"); err != nil || got != "avenue" {
		t.Errorf("RequireMotionSubjectRef(Q1) = (%q, %v) — an avenue IS an avenue motion by construction", got, err)
	}
	if _, err := RequireMotionSubjectRef(run, recordpb.MotionSubject_MOTION_SUBJECT_AVENUE, "Z9"); err == nil {
		t.Error("RequireMotionSubjectRef accepted an id that names no avenue")
	}
	if got, err := AvenueRuling(run, "Q1"); err != nil || got != avenueRulingFold(t, run, "Q1") || got != "" {
		t.Errorf("AvenueRuling before any ruling = (%q, %v), fold says %q", got, err, avenueRulingFold(t, run, "Q1"))
	}
	if _, err := Append(red, &recordpb.MotionRule{
		MotionId: proto.String("Q1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE),
		Opinion:  proto.String("worth the run's time"),
		Ruling:   &recordpb.MotionRule_Avenue{Avenue: recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE},
	}); err != nil {
		t.Fatal(err)
	}
	if got, err := AvenueRuling(run, "Q1"); err != nil || got != avenueRulingFold(t, run, "Q1") || got != "out_of_scope" {
		t.Errorf("AvenueRuling = (%q, %v), fold says %q, want the schema's word out_of_scope", got, err, avenueRulingFold(t, run, "Q1"))
	}
}
