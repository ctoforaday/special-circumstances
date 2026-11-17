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
// the SQL view must name the same answer. Both read the ruling that stands;
// TestEveryMotionReaderStatesTheRulingThatStands holds them to it beside a remand.
func TestTheMotionWriteRefusesWhatReadersCannotAgreeOn(t *testing.T) {
	finalDocketRule := func() *recordpb.MotionRule {
		r := docketRule("M2", "the bench rules again")
		r.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
		r.GetDocket().Direction = nil
		r.GetDocket().Final = proto.Bool(true)
		return r
	}

	for _, tc := range []struct {
		name string
		// act runs on a record holding docket motion M1 (gap G1, proposing a regrade), docket motion
		// M2 (gap G2) and petition M3, all unruled; its LAST write is the one under test.
		act func(t *testing.T, run Run) error
		// wants is what the refusal must say; nil means the last write is admitted.
		wants []string
		// ruling is the answer the motion must carry afterwards.
		motion, ruling string
	}{
		{
			name: "a second ruling is refused",
			act: func(t *testing.T, run Run) error {
				judge := sit(t, run, "judge")
				mustAppend(t, judge, actRule("M1", false, "judge"))
				_, err := Append(judge, actRule("M1", true, "judge"))
				return err
			},
			// The ruler is told the act that answers its ruling.
			wants:  []string{`motion M1 is already ruled "rejected" by judge`, "a declaration, which moves no gap"},
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
			name: "a ruling under another subject is refused",
			act: func(t *testing.T, run Run) error {
				_, err := Append(sit(t, run, "judge"), &recordpb.MotionRule{MotionId: proto.String("M1"), Opinion: proto.String("o"),
					Ruling: &recordpb.MotionRule_Petition{Petition: recordpb.PetitionRuling_PETITION_RULING_GRANTED}})
				return err
			},
			wants:  []string{"motion M1 was filed as a docket motion and you are ruling it as a petition", "the bench's `motion docket rule`"},
			motion: "M1",
		},
		{
			name: "a docket ruling on a petition is refused for the subject before its fields",
			act: func(t *testing.T, run Run) error {
				r := docketRule("M3", "the bench rules a petition")
				r.GetDocket().Settled = nil
				_, err := Append(sit(t, run, "judge"), r)
				return err
			},
			// ORDER IS THE MESSAGE: --settled is missing too, and naming it would send the bench
			// to supply a field for a ruling it cannot make.
			wants:  []string{"motion M3 was filed as a petition motion and you are ruling it as a docket", "the bench's `motion petition rule`"},
			motion: "M3",
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
			if got.ruling != tc.ruling {
				t.Errorf("motion %s reads ruled %q; want ruled %q", tc.motion, got.ruling, tc.ruling)
			}
		})
	}
}

// motionGuardRun is a record with three unruled motions: docket M1 on G1 proposing a regrade and
// docket M2 on G2 proposing no act (both filed as the dispatch files them), and petition M3 (filed by
// the lens); the chair's bench row on G1 and G2 then convenes the bench's next docket sitting on both.
func motionGuardRun(t *testing.T) Run {
	t.Helper()
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).
		mint(evLens, "G1", "high").mint(evLens, "G2", "high").seed()
	mustFileByDispatch(t, run, regradeMotion("M1", "G1", "severity is overstated"))
	mustFileByDispatch(t, run, docketMotion("M2", "G2"))
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(benchSeat), GapIds: []string{"G1", "G2"}})
	mustAppend(t, sit(t, run, evLens), &recordpb.Motion{MotionId: proto.String("M3"), Basis: proto.String("the run is proceeding past a safety objection"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY)}}})
	return run
}

// motionReadersAgree holds the Go fold and the SQL view to one answer per motion — the ruling and
// who ruled — and returns the answer each ruled motion carries.
func motionReadersAgree(t *testing.T, run Run) map[string]motionAnswer {
	t.Helper()
	fold := map[string]motionAnswer{}
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range MotionsOf(all.Events, all.At) {
		if !m.Ruled() {
			continue
		}
		fold[m.ID] = motionAnswer{ruling: m.Ruling, by: m.RulingBy}
	}
	db, err := openRunForRead(run)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := db.Query(`SELECT "motion_id", "ruling", "ruled_by" FROM "motion_answers" WHERE "ruled_by" IS NOT NULL`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	view := map[string]motionAnswer{}
	for rows.Next() {
		var id string
		var ruling, by sql.NullString
		if err := rows.Scan(&id, &ruling, &by); err != nil {
			t.Fatal(err)
		}
		view[id] = motionAnswer{ruling: ruling.String, by: by.String}
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
type motionAnswer struct{ ruling, by string }

// EVERY READER OF A MOTION'S ANSWER STATES THE RULING THAT STANDS, after the legal acts — an
// answer to a proposed act, a docket remand — went through Append. A motion holds one ruling: the
// write refuses a second, migrate's replay included, so no record holds two.
func TestEveryMotionReaderStatesTheRulingThatStands(t *testing.T) {
	run := motionGuardRun(t)

	// The legal acts, through Append.
	judge := sit(t, run, "judge")
	mustAppend(t, judge, actRule("M1", false, "judge"))
	mustAppend(t, judge, docketRule("M2", "the bench remands"))
	// A second motion on G2, proposing a regrade, once the first is answered (one at a time on a
	// gap), left unruled by the write.
	mustFileByDispatch(t, run, regradeMotion("M4", "G2", "likelihood is overstated"))

	motions := motionReadersAgree(t, run)
	if got, want := motions["M1"], (motionAnswer{ruling: "rejected", by: "judge"}); got != want {
		t.Errorf("motion M1 reads %+v; the ruling that stands is %+v", got, want)
	}
	if got, ok := motions["M4"]; ok {
		t.Errorf("motion M4 reads %+v; nobody ruled it", got)
	}
	if got := motions["M2"].ruling; got != "remanded" {
		t.Errorf("docket motion M2 reads ruled %q, want the bench's remand", got)
	}
}

// A MOTION IS ANSWERED ONCE WHEN BOTH ITS ANSWERERS ANSWER AT ONCE (T-B7). The first-wins guards
// read the record and then the write inserts; outside the writing transaction that is
// check-then-insert, and each answerer that read "unanswered" before the other's insert committed
// lands its own answer. The two seats a proposal's answer admits — the gap's assignee accepting, the
// bench rejecting — race, each on its own connection, so no idempotency key collides to hide a missing
// guard: exactly one answer lands, the other is refused in the guard's own words, and the two readers
// still name one answer.
func TestAMotionIsAnsweredOnceWhenSeatsWriteAtOnce(t *testing.T) {
	seats := []string{evLens, "judge"}
	for _, tc := range []struct {
		name    string
		act     func(i int) proto.Message
		refusal string
	}{
		{
			name: "rulings",
			act: func(i int) proto.Message {
				r := actRule("M1", seats[i] == evLens, seats[i])
				r.Opinion = proto.String(fmt.Sprintf("ruling %d", i))
				return r
			},
			refusal: "motion M1 is already ruled",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := newStage(t).cast(append([]string{"red-chair", "blue-respond"}, seats...)...).ingest().register("red-chair").dispatch(2, evLens)
			run := st.register(evLens).mint(evLens, "G1", "high").seed()
			mustFileByDispatch(t, run, regradeMotion("M1", "G1", "severity is overstated"))
			mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(benchSeat), GapIds: []string{"G1"}})
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
			if got.ruling != "rejected" && got.ruling != "accepted" {
				t.Errorf("motion M1 reads %+v after the race", got)
			}
		})
	}
}

// regradeMotion is a docket motion on gap proposing its severity low, argued by basis.
func regradeMotion(id, gap, basis string) *recordpb.Motion {
	return &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String(basis),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}},
		Act: &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{GapId: proto.String(gap),
			Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String(basis)}}}
}

// actRule answers a docket motion's proposed act: accepted or rejected, by seat. The bench's answer
// carries its opinion; another seat's carries none (requireBenchOpinion).
func actRule(id string, accepted bool, seat string) *recordpb.MotionRule {
	d := &recordpb.DocketRuling{Accepted: proto.Bool(accepted)}
	if seat == "judge" {
		d.Principle, d.Tension, d.ReviewFlag, d.Settled = proto.String("pr"), proto.String("tn"), proto.String("rf"), proto.String("st")
	}
	return &recordpb.MotionRule{MotionId: proto.String(id), Opinion: proto.String("ruled"),
		Ruling: &recordpb.MotionRule_Docket{Docket: d}}
}
