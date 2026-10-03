package record

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
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
		// act runs on a record holding grade motion M1 (gap G1), docket motion M2 (gap G2),
		// avenue Q1 and petition M3, all unruled; its LAST write is the one under test.
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
			// Worded for the act: an appellant is told it is appealing, and pointed at the appeal
			// the motion's own subject takes.
			wants:  []string{"motion M1 was filed as a grade motion and you are appealing it as a petition", "`motion grade appeal`"},
			motion: "M1", ruling: "rejected",
		},
		{
			name: "an appeal of a docket ruling is refused",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "judge"), docketRule("M2", "the bench rules"))
				_, err := Append(sit(t, run, "blue-respond"), &recordpb.MotionAppeal{MotionId: proto.String("M2"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET), Reason: proto.String("pressed past the bench")})
				return err
			},
			wants:  []string{"docket motion M2 has no appeal: the bench rules it", "file a NEW motion"},
			motion: "M2", ruling: "remanded",
		},
		{
			name: "an appeal of a petition ruling is refused",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "judge"), &recordpb.MotionRule{MotionId: proto.String("M3"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Opinion: proto.String("o"),
					Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_DENIED}})
				_, err := Append(sit(t, run, evLens), &recordpb.MotionAppeal{MotionId: proto.String("M3"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Reason: proto.String("pressed past the bench")})
				return err
			},
			wants:  []string{"petition motion M3 has no appeal: the bench rules it"},
			motion: "M3", ruling: "denied",
		},
		{
			name: "an appeal of a docket ruling under another subject is refused for the subject",
			act: func(t *testing.T, run Run) error {
				mustAppend(t, sit(t, run, "judge"), docketRule("M2", "the bench rules"))
				_, err := Append(sit(t, run, "blue-respond"), &recordpb.MotionAppeal{MotionId: proto.String("M2"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE), Reason: proto.String("pressed as a grade")})
				return err
			},
			// The appellant is told the motion takes no appeal at all, not sent to a verb that refuses it.
			wants:  []string{"motion M2 was filed as a docket motion and you are appealing it as a grade", "docket motion M2 has no appeal"},
			motion: "M2", ruling: "remanded",
		},
		{
			name: "a docket ruling on a grade motion is refused for the subject before its fields",
			act: func(t *testing.T, run Run) error {
				r := docketRule("M1", "the bench rules a grade")
				r.GetDocket().Settled = nil
				_, err := Append(sit(t, run, "judge"), r)
				return err
			},
			// ORDER IS THE MESSAGE: --settled is missing too, and naming it would send the bench
			// to supply a field for a ruling it cannot make.
			wants:  []string{"motion M1 was filed as a grade motion and you are ruling it as a docket", "the chair's `motion grade rule`"},
			motion: "M1",
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
				if strings.Contains(err.Error(), "settled") {
					t.Errorf("the refusal names a field before the subject:\n%v", err)
				}
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

// motionGuardRun is a record with four unruled motions: grade M1 on G1 (filed by the lens),
// docket M2 on G2 (filed by blue), avenue Q1 (proposed by blue) and petition M3 (filed by the lens).
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
	mustAppend(t, sit(t, run, evLens), &recordpb.Motion{MotionId: proto.String("M3"),
		Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION), Basis: proto.String("the run is proceeding past a safety objection"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY)}}})
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

// A MOTION IS ANSWERED ONCE WHEN EVERY SEAT ANSWERS AT ONCE. The first-wins guards read the record
// and then the write inserts; outside the writing transaction that is check-then-insert, and every
// seat that read "unanswered" before the first insert committed lands its own answer. Each
// goroutine is a different seat, so no idempotency key collides to hide a missing guard: exactly
// one answer may land, every other is refused in the guard's own words, and the two readers
// still name one answer.
func TestAMotionIsAnsweredOnceWhenSeatsWriteAtOnce(t *testing.T) {
	seats := []string{evLens, "red-lens-logic", "red-lens-voice", "red-lens-adversary",
		"red-lens-architecture", "red-lens-computation", "red-lens-dark-side", "red-chair", "blue-respond", "judge"}
	for _, tc := range []struct {
		name string
		// ruled says whether the motion is ruled before the race, which an appeal needs.
		ruled   bool
		act     func(i int) proto.Message
		refusal string
	}{
		{
			name: "rulings",
			act: func(i int) proto.Message {
				return &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
					Opinion: proto.String(fmt.Sprintf("ruling %d", i)), Ruling: &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED}}
			},
			refusal: "motion M1 is already ruled",
		},
		{
			name:  "appeals",
			ruled: true,
			act: func(i int) proto.Message {
				return &recordpb.MotionAppeal{MotionId: proto.String("M1"), Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
					Reason: proto.String(fmt.Sprintf("appeal %d", i))}
			},
			refusal: "motion M1 is already appealed",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := newStage(t).cast(seats...).ingest().register("red-chair")
			for _, s := range seats {
				if strings.HasPrefix(s, "red-lens-") {
					st = st.dispatch(2, s)
				}
			}
			run := st.register(evLens).mint(evLens, "G1", "high").seed()
			mustAppend(t, sit(t, run, evLens), &recordpb.Motion{MotionId: proto.String("M1"),
				Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE), Basis: proto.String("severity is overstated"),
				Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G1"),
					Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_SEVERITY), Proposed: recordtest.P(recordpb.Grade_GRADE_LOW)}}})
			if tc.ruled {
				mustAppend(t, sit(t, run, "red-chair"), &recordpb.MotionRule{MotionId: proto.String("M1"),
					Subject: recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE), Opinion: proto.String("the grade stands"),
					Ruling: &recordpb.MotionRule_Grade{Grade: recordpb.GradeRuling_GRADE_RULING_REJECTED}})
			}
			// Every seat sits BEFORE the race: a register is a write too, and the race under test
			// is the answers', not the sittings'.
			ids := make([]Identity, len(seats))
			for i, s := range seats {
				ids[i] = sit(t, run, s)
			}
			start := make(chan struct{})
			errs := make([]error, len(ids))
			var wg sync.WaitGroup
			for i := range ids {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					_, errs[i] = Append(ids[i], tc.act(i))
				}()
			}
			close(start)
			wg.Wait()
			landed := 0
			for i, err := range errs {
				switch {
				case err == nil:
					landed++
				case !strings.Contains(err.Error(), tc.refusal):
					t.Errorf("%s was refused for something other than the answer already standing: %v", seats[i], err)
				}
			}
			if landed != 1 {
				t.Errorf("%d of %d concurrent %s landed; want exactly one", landed, len(ids), tc.name)
			}
			got := motionReadersAgree(t, run)["M1"]
			if got.ruling != "rejected" || (tc.ruled && got.appealReason == "") {
				t.Errorf("motion M1 reads %+v after the race", got)
			}
		})
	}
}
