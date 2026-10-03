package record

import (
	"database/sql"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE MOTION STATE GUARDS ARE THE WRITE'S, NOT THE COMMAND LINE'S (#1205). Every row goes through
// record.Append — the path a fixture, migrate, the fuzz or a future harness takes — so a guard
// that lives only in a verb fails here. After every row, admitted or refused, the Go fold and
// the SQL view must name the same answer: a record holding two rulings is the one that makes
// MotionsOf (last-wins) and motion_answers (first-wins) disagree.
func TestTheMotionWriteRefusesWhatReadersCannotAgreeOn(t *testing.T) {
	gradeRule := func(r recordpb.GradeRuling) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
			Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Grade{Grade: r}}
	}
	avenueRule := func(r recordpb.AvenueRuling) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String("Q1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE),
			Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Avenue{Avenue: r}}
	}
	appeal := func(subject recordpb.MotionSubject, reason string) *recordpb.MotionAppeal {
		return &recordpb.MotionAppeal{MotionId: proto.String("M1"), Subject: recordtest.P(subject), Reason: proto.String(reason)}
	}
	gradeAppeal := func(reason string) *recordpb.MotionAppeal {
		return appeal(recordpb.MotionSubject_MOTION_SUBJECT_GRADE, reason)
	}
	finalDocketRule := func() *recordpb.MotionRule {
		r := docketRule("M2", "the bench rules again")
		r.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
		r.GetDocket().ReopensOn = nil
		r.GetDocket().Final = proto.Bool(true)
		return r
	}

	for _, tc := range []struct {
		name string
		// act runs on a record holding grade motion M1 (gap G1), docket motion M2 (gap G2) and
		// avenue Q1, all unruled; its LAST write is the one under test.
		act func(t *testing.T, run Run) error
		// wants is what the refusal must say; nil means the last write is admitted.
		wants []string
		// ruling is the answer the motion must carry afterwards, and appealReason the argument
		// its appeal must read as — the replacement's, after a correction.
		motion, ruling, appealReason string
	}{
		{
			name: "a second ruling is refused",
			act: func(t *testing.T, run Run) error {
				chair := sit(t, run, "red-chair")
				mustAppend(t, chair, gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED))
				_, err := Append(chair, gradeRule(recordpb.GradeRuling_GRADE_RULING_ACCEPTED))
				return err
			},
			// The asker's own ruling from this sitting is offered the correction instead.
			wants:  []string{`motion M1 is already ruled "rejected" by red-chair`, "appeal", "It is your own ruling from this sitting", "--corrects red-chair:motion_rule:#"},
			motion: "M1", ruling: "rejected",
		},
		{
			name: "a second docket ruling is refused",
			act: func(t *testing.T, run Run) error {
				judge := sit(t, run, "judge")
				mustAppend(t, judge, docketRule("M2", "the bench rules"))
				_, err := Append(judge, finalDocketRule())
				return err
			},
			wants:  []string{`motion M2 is already ruled "remanded" by judge`},
			motion: "M2", ruling: "remanded",
		},
		{
			name: "a second avenue ruling is refused",
			act: func(t *testing.T, run Run) error {
				chair := sit(t, run, "red-chair")
				mustAppend(t, chair, avenueRule(recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE))
				_, err := Append(chair, avenueRule(recordpb.AvenueRuling_AVENUE_RULING_ENDORSED))
				return err
			},
			wants:  []string{`motion Q1 is already ruled "out_of_scope" by red-chair`},
			motion: "Q1", ruling: "out_of_scope",
		},
		{
			name: "a correction of the ruling in its sitting is admitted",
			act: func(t *testing.T, run Run) error {
				chair := sit(t, run, "red-chair")
				k := mustAppend(t, chair, gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED)).GetKey()
				// A correction restates the seat's own wording; the ruling it corrects stands.
				fixed := gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED)
				fixed.Opinion = proto.String("the grade stands, and here is why")
				_, err := Append(correcting(chair, recordpb.EventType_EVENT_TYPE_MOTION_RULE, k, "an unreasoned opinion"), fixed)
				return err
			},
			motion: "M1", ruling: "rejected",
		},
		{
			name: "a ruling under another subject is refused",
			act: func(t *testing.T, run Run) error {
				_, err := Append(sit(t, run, "judge"), &recordpb.MotionRule{MotionId: proto.String("M1"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Opinion: proto.String("o"),
					Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_GRANTED}})
				return err
			},
			wants:  []string{"motion M1 was filed as a grade motion and you are ruling it as a petition", "the chair's `motion grade rule`"},
			motion: "M1",
		},
		{
			name: "an appeal of an unruled motion is refused",
			act: func(t *testing.T, run Run) error {
				_, err := Append(sit(t, run, "blue-respond"), gradeAppeal("the grade is wrong"))
				return err
			},
			// The subject is the surface's word, never the enum's name.
			wants:  []string{"grade motion M1 has no ruling to appeal"},
			motion: "M1",
		},
		{
			name: "a second appeal is refused",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "red-chair"), gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED))
				mustAppend(t, sit(t, run, "blue-respond"), gradeAppeal("the ruling reads past the argument"))
				_, err := Append(sit(t, run, evLens), gradeAppeal("and again"))
				return err
			},
			wants:  []string{`motion M1 is already appealed by blue-respond ("the ruling reads past the argument")`},
			motion: "M1", ruling: "rejected", appealReason: "the ruling reads past the argument",
		},
		{
			name: "a correction of the appeal in its sitting is admitted",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "red-chair"), gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED))
				blue := sit(t, run, "blue-respond")
				k := mustAppend(t, blue, gradeAppeal("a typo")).GetKey()
				_, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_MOTION_APPEAL, k, "the typo"),
					gradeAppeal("the ruling reads past the argument"))
				return err
			},
			motion: "M1", ruling: "rejected", appealReason: "the ruling reads past the argument",
		},
		{
			name: "an appeal under another subject is refused",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "red-chair"), gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED))
				_, err := Append(sit(t, run, "blue-respond"), appeal(recordpb.MotionSubject_MOTION_SUBJECT_PETITION, "pressed as a petition"))
				return err
			},
			wants:  []string{"motion M1 was filed as a grade motion and you are ruling it as a petition"},
			motion: "M1", ruling: "rejected",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := motionGuardRun(t)
			err := tc.act(t, run)
			if tc.wants == nil {
				if err != nil {
					t.Fatalf("refused; want it admitted: %v", err)
				}
			} else {
				mustRefuse(t, err, tc.wants...)
				if strings.Contains(err.Error(), "MOTION_SUBJECT_") {
					t.Errorf("the refusal prints the enum's name rather than the subject's word:\n%v", err)
				}
			}
			got := motionReadersAgree(t, run)[tc.motion]
			if got.ruling != tc.ruling || got.appealReason != tc.appealReason {
				t.Errorf("motion %s reads ruled %q, appealed %q; want ruled %q, appealed %q",
					tc.motion, got.ruling, got.appealReason, tc.ruling, tc.appealReason)
			}
		})
	}
}

// motionGuardRun is a record with three unruled motions: grade M1 on G1 (filed by the lens),
// docket M2 on G2 (filed by blue) and avenue Q1 (proposed by blue).
func motionGuardRun(t *testing.T) Run {
	t.Helper()
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).register(evLens).
		mint(evLens, "G1", "high").mint(evLens, "G2", "high").seed()
	mustAppend(t, sit(t, run, evLens), &recordpb.Motion{MotionId: proto.String("M1"),
		Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE), Basis: proto.String("severity is overstated"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
			Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY), Proposed: recordtest.P(recordpb.Grade_GRADE_LOW)}}})
	blue := sit(t, run, "blue-respond")
	mustAppend(t, blue, docketMotion("M2", "G2"))
	mustAppend(t, blue, &recordpb.Avenue{AvenueId: proto.String("Q1"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("a direction")})
	return run
}

// motionReadersAgree holds the Go fold and the SQL view to one answer per motion — the ruling,
// who ruled, and the appeal — and returns the answer each motion carries.
func motionReadersAgree(t *testing.T, run Run) map[string]motionAnswer {
	t.Helper()
	fold := map[string]motionAnswer{}
	for _, m := range MotionsOf(allEvents(t, run)) {
		if !m.Ruled() && !m.Appealed {
			continue
		}
		a := motionAnswer{ruling: m.Ruling, by: m.RulingBy, appealReason: m.AppealReason}
		if m.Appealed {
			a.appealedBy = "appealed"
		}
		fold[m.ID] = a
	}
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT "motion_id", "ruling", "ruled_by", "appealed_by", "appeal_reason" FROM "motion_answers"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	view := map[string]motionAnswer{}
	for rows.Next() {
		var id string
		var ruling, by, appealedBy, reason sql.NullString
		if err := rows.Scan(&id, &ruling, &by, &appealedBy, &reason); err != nil {
			t.Fatal(err)
		}
		a := motionAnswer{ruling: ruling.String, by: by.String, appealReason: reason.String}
		if appealedBy.Valid {
			a.appealedBy = "appealed"
		}
		view[id] = a
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for id, f := range fold {
		if v := view[id]; v != f {
			t.Errorf("motion %s: MotionsOf reads %+v and motion_answers reads %+v", id, f, v)
		}
	}
	for id, v := range view {
		if _, ok := fold[id]; !ok {
			t.Errorf("motion %s: motion_answers reads %+v and MotionsOf has no answer", id, v)
		}
	}
	return fold
}

// motionAnswer is one motion's answer as a reader states it.
type motionAnswer struct{ ruling, by, appealedBy, appealReason string }
