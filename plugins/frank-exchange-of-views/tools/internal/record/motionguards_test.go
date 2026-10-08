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
// the SQL view must name the same answer. Both read the first ruling and the first appeal that
// stand; TestEveryMotionReaderStatesFirstWins holds them to it on a record seeded past the write.
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
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range MotionsOf(all.Events, all.At) {
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

// avenueReadersAgree holds AvenuesOf to motion_answers on every avenue: the ruling, the argument
// and epoch of the ruling event the view names, and the ruling an appeal contests — the view's
// pairing of the first appeal with the first ruling, whatever their order. It returns each
// avenue as AvenuesOf reads it.
func avenueReadersAgree(t *testing.T, run Run) map[string]*Avenue {
	t.Helper()
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]*Avenue{}
	for _, a := range AvenuesOf(all.Events, all.At) {
		out[a.ID] = a
		var ruling, contests, why sql.NullString
		var epoch sql.NullInt64
		err := db.QueryRow(`
		  SELECT a."avenue",
		         CASE WHEN a."appealed_seq" IS NOT NULL THEN a."avenue" END,
		         mr."opinion", w."epoch"
		  FROM "motion_answers" a
		  LEFT JOIN "motion_rule" mr ON mr."event_id" = a."ruled_seq"
		  LEFT JOIN "events_w" w ON w."id" = a."ruled_seq"
		  WHERE a."motion_id" = ?`, a.ID).Scan(&ruling, &contests, &why, &epoch)
		if err != nil && err != sql.ErrNoRows {
			t.Fatal(err)
		}
		got := [4]any{a.Ruling, a.Contests, a.RulingWhy, a.RuledEpoch}
		if want := [4]any{ruling.String, contests.String, why.String, int(epoch.Int64)}; got != want {
			t.Errorf("avenue %s: AvenuesOf reads (ruling, contests, why, epoch) %v and motion_answers reads %v", a.ID, got, want)
		}
	}
	return out
}

// EVERY READER OF A MOTION'S ANSWER STATES THE WRITE'S RULE: the first ruling that stands, the
// first appeal that stands. The write refuses a second of either, so the edges that would tell
// two rules apart are seeded past it, straight into the record, after the legal acts — a ruling
// and an appeal each corrected in their sitting, an avenue ruling, a docket remand — went through
// Append. A correction is not a second answer: the replacement stands in the original's place.
func TestEveryMotionReaderStatesFirstWins(t *testing.T) {
	run := motionGuardRun(t)
	grade := recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE)
	avenue := recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_AVENUE)
	gradeRule := func(r recordpb.GradeRuling, opinion string) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String("M1"), Subject: grade,
			Opinion: proto.String(opinion), Ruling: &recordpb.MotionRule_Grade{Grade: r}}
	}
	avenueRule := func(id string, r recordpb.AvenueRuling, opinion string) *recordpb.MotionRule {
		return &recordpb.MotionRule{MotionId: proto.String(id), Subject: avenue,
			Opinion: proto.String(opinion), Ruling: &recordpb.MotionRule_Avenue{Avenue: r}}
	}
	appeal := func(id string, subject *recordpb.MotionSubject, reason string) *recordpb.MotionAppeal {
		return &recordpb.MotionAppeal{MotionId: proto.String(id), Subject: subject, Reason: proto.String(reason)}
	}

	// A second grade motion and a second avenue, both left unruled by the write.
	mustAppend(t, sit(t, run, evLens), &recordpb.Motion{MotionId: proto.String("M4"),
		Subject: grade, Basis: proto.String("likelihood is overstated"),
		Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String("G2"),
			Dimension: recordtest.P(recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD), Proposed: recordtest.P(recordpb.Grade_GRADE_LOW)}}})
	mustAppend(t, sit(t, run, "blue-respond"), &recordpb.Avenue{AvenueId: proto.String("Q2"),
		Status: recordpb.AvenueStatus_AVENUE_STATUS_PROPOSED.Enum(), Line: proto.String("another direction")})

	// The legal acts, through Append.
	chair := sit(t, run, "red-chair")
	k := mustAppend(t, chair, gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED, "o")).GetKey()
	if _, err := Append(correcting(chair, recordpb.EventType_EVENT_TYPE_MOTION_RULE, k, "an unreasoned opinion"),
		gradeRule(recordpb.GradeRuling_GRADE_RULING_REJECTED, "the grade stands, and here is why")); err != nil {
		t.Fatal(err)
	}
	mustAppend(t, chair, avenueRule("Q1", recordpb.AvenueRuling_AVENUE_RULING_OUT_OF_SCOPE, "the first ruling's argument"))
	blue := sit(t, run, "blue-respond")
	k = mustAppend(t, blue, appeal("M1", grade, "a typo")).GetKey()
	if _, err := Append(correcting(blue, recordpb.EventType_EVENT_TYPE_MOTION_APPEAL, k, "the typo"),
		appeal("M1", grade, "the ruling reads past the argument")); err != nil {
		t.Fatal(err)
	}
	mustAppend(t, sit(t, run, "judge"), docketRule("M2", "the bench remands"))
	motionReadersAgree(t, run)
	avenueReadersAgree(t, run)

	// The forbidden edges, past the write. The chair sits again first, so the second avenue
	// ruling differs from the first in its word, its argument AND its epoch.
	sit(t, run, "red-chair")
	recordtest.Seed(t, run.dir,
		recordtest.Event(t, "judge", gradeRule(recordpb.GradeRuling_GRADE_RULING_ACCEPTED, "a second ruling")),
		recordtest.Event(t, evLens, appeal("M1", grade, "a second appeal")),
		recordtest.Event(t, "blue-respond", appeal("M4", grade, "an appeal of nothing")),
		recordtest.Event(t, "red-chair", avenueRule("Q1", recordpb.AvenueRuling_AVENUE_RULING_ENDORSED, "the second ruling's argument")),
		recordtest.Event(t, "blue-respond", appeal("Q2", avenue, "an appeal before any ruling")),
		recordtest.Event(t, "red-chair", avenueRule("Q2", recordpb.AvenueRuling_AVENUE_RULING_TOO_THIN, "ruled after its appeal")),
	)

	// Each edge is on the record: standing rulings and standing appeals per motion, and Q2's
	// appeal ahead of its ruling.
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	standing := func(table, id string) (n int, first int64) {
		t.Helper()
		var min sql.NullInt64
		if err := db.QueryRow(`SELECT count(*), MIN(l."pos") FROM "`+table+`" x JOIN "live_event" l ON l."event_id" = x."event_id"
		  WHERE x."motion_id" = ?`, id).Scan(&n, &min); err != nil {
			t.Fatal(err)
		}
		return n, min.Int64
	}
	for _, want := range []struct {
		id               string
		rulings, appeals int
	}{{"M1", 2, 2}, {"M4", 0, 1}, {"Q1", 2, 0}, {"Q2", 1, 1}, {"M2", 1, 0}} {
		r, _ := standing("motion_rule", want.id)
		a, _ := standing("motion_appeal", want.id)
		if r != want.rulings || a != want.appeals {
			t.Fatalf("motion %s holds %d standing rulings and %d standing appeals; the fixture seeds %d and %d",
				want.id, r, a, want.rulings, want.appeals)
		}
	}
	_, ruledAt := standing("motion_rule", "Q2")
	if _, appealedAt := standing("motion_appeal", "Q2"); appealedAt >= ruledAt {
		t.Fatalf("avenue Q2's appeal (pos %d) does not precede its ruling (pos %d)", appealedAt, ruledAt)
	}

	motions := motionReadersAgree(t, run)
	if got, want := motions["M1"], (motionAnswer{ruling: "rejected", by: "red-chair", appealedBy: "appealed", appealReason: "the ruling reads past the argument"}); got != want {
		t.Errorf("motion M1 reads %+v; the first ruling and the first appeal that stand are %+v", got, want)
	}
	if got, want := motions["M4"], (motionAnswer{appealedBy: "appealed", appealReason: "an appeal of nothing"}); got != want {
		t.Errorf("motion M4 reads %+v; it is unruled and appealed: %+v", got, want)
	}
	if got := motions["M2"].ruling; got != "remanded" {
		t.Errorf("docket motion M2 reads ruled %q, want the bench's remand", got)
	}
	avenues := avenueReadersAgree(t, run)
	if a := avenues["Q1"]; a == nil || a.Ruling != "out_of_scope" || a.RulingWhy != "the first ruling's argument" || a.Contests != "" {
		t.Errorf("avenue Q1 reads %+v; its first ruling is out_of_scope, with that ruling's argument, and nobody appealed it", a)
	}
	if a := avenues["Q2"]; a == nil || a.Ruling != "too_thin" || a.Contests != "too_thin" {
		t.Errorf("avenue Q2 reads %+v; its appeal contests the ruling the motion carries, too_thin", a)
	}
}

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
