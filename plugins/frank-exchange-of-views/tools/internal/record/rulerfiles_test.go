package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// filingBy is the body that FILES a motion on the subject, as any writer reaches Append with it: a
// motion — or, for the inquiry, the proposal the chair then rules. A subject with no arm fails the
// test that walks the enum, so a new subject is driven here on the day it is added.
func filingBy(subj recordpb.MotionSubject) proto.Message {
	m := &recordpb.Motion{MotionId: proto.String("M9"), Subject: recordtest.P(subj), Basis: proto.String("the ask")}
	switch subj {
	case recordpb.MotionSubject_MOTION_SUBJECT_GRADE:
		m.Filing = &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G2"),
			Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY), Proposed: recordtest.P(recordpb.Grade_GRADE_LOW)}}
	case recordpb.MotionSubject_MOTION_SUBJECT_DOCKET:
		m.Filing = &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}}
	case recordpb.MotionSubject_MOTION_SUBJECT_PETITION:
		m.Filing = &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY)}}
	case recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY:
		return &recordpb.Inquiry{InquiryId: proto.String("Q9"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum(), Question: proto.String("another question")}
	default:
		return nil
	}
	return m
}

// THE RECORD REFUSES A FILING BY THE SEAT THAT RULES ITS SUBJECT, whatever path the body took to
// Append — the verb's absence from the ruler's surface is the first refusal, and this is the one a
// fixture, a harness or a hand-built body meets. Walked over the MotionSubject enum: every subject
// is refused its ruler and admitted another seat, and migrate replays what an archive holds.
func TestTheWriteRefusesAFilingByTheSubjectsRuler(t *testing.T) {
	vals := recordpb.MotionSubject(0).Descriptor().Values()
	// A seat of another role that the write admits as the subject's filer.
	other := map[string]string{chairSeat: "blue-respond", benchSeat: "blue-respond"}
	checked := 0
	for i := 0; i < vals.Len(); i++ {
		subj := recordpb.MotionSubject(vals.Get(i).Number())
		if subj == recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED {
			continue
		}
		word := recordpb.Word(subj)
		body := filingBy(subj)
		if body == nil {
			t.Errorf("%s: this test has no filing body for it — add the arm, so its ruler's filing is driven", word)
			continue
		}
		g, err := gavelOf(word)
		if err != nil {
			t.Fatal(err)
		}
		ruler, _ := recordpb.SubjectRuler(subj)
		t.Run(word+", filed by its ruler", func(t *testing.T) {
			run := motionGuardRun(t)
			_, err := Append(sit(t, run, g.seat), filingBy(subj))
			mustRefuse(t, err, word+" motion is ruled by the "+ruler+" seat", "you are "+g.seat+" ("+ruler+")", "no other seat can rule")
		})
		t.Run(word+", filed by another seat", func(t *testing.T) {
			run := motionGuardRun(t)
			mustAppend(t, sit(t, run, other[g.seat]), filingBy(subj))
		})
		t.Run(word+", replayed by migrate", func(t *testing.T) {
			run := motionGuardRun(t)
			if _, err := replayed(sit(t, run, g.seat), filingBy(subj)); err != nil {
				t.Fatalf("migrate replays what an archive holds, and an archived record holds motions their ruler filed: %v", err)
			}
		})
		checked++
	}
	if checked < 4 {
		t.Errorf("only %d subjects were driven — the walk is meant to cover the whole enum", checked)
	}
}

// HoldsGavel is the one predicate, and it is the schema's `ruled_by` read once: the role that
// rules a subject, and no other.
func TestHoldsGavelIsTheSchemasRuler(t *testing.T) {
	for _, word := range MotionSubjects {
		subj, _ := MotionSubjectEnum(word)
		ruler, err := recordpb.SubjectRuler(subj)
		if err != nil {
			t.Fatal(err)
		}
		holders := 0
		for _, role := range seatclass.Roles() {
			if HoldsGavel(role, subj) {
				holders++
				if role != ruler {
					t.Errorf("%s: HoldsGavel says the %s role, the schema says %s", word, role, ruler)
				}
			}
		}
		if holders != 1 {
			t.Errorf("%s: %d roles hold its gavel, want one", word, holders)
		}
	}
	if HoldsGavel("chair", recordpb.MotionSubject_MOTION_SUBJECT_UNSPECIFIED) {
		t.Error("a role holds the gavel of no subject at all")
	}
}
