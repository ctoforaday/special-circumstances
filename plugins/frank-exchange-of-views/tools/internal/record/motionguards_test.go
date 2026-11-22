package record

import (
	"database/sql"
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
		r := docketRule("D2", "the bench rules again")
		r.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
		r.GetDocket().SidedWith = recordpb.SidedWith_SIDED_WITH_BLUE.Enum()
		r.GetDocket().Direction = nil
		r.GetDocket().Final = proto.Bool(true)
		return r
	}

	for _, tc := range []struct {
		name string
		// act runs on a record holding proposal M1 (gap G1, proposing a regrade) docketed as D1,
		// docketing D2 (gap G2, at impasse) and petition M3, all unruled; its LAST write is the one
		// under test.
		act func(t *testing.T, run Run) error
		// wants is what the refusal must say; nil means the last write is admitted.
		wants []string
		// ruling is the answer the motion, or the docketing, must carry afterwards.
		motion, docketing, ruling string
	}{
		{
			name: "a second ruling is refused",
			act: func(t *testing.T, run Run) error {
				judge := sit(t, run, "judge")
				mustAppend(t, judge, heldBy(actRule("M1", false, "judge"), "D1"))
				_, err := Append(judge, benchRemand("D1", "M1"))
				return err
			},
			// The ruler is told the act that answers its ruling.
			wants:  []string{`motion M1 is already ruled "for assignee" by judge`, "a declaration, which moves no gap"},
			motion: "M1", ruling: "for assignee",
		},
		{
			name: "a second docket ruling is refused",
			act: func(t *testing.T, run Run) error {
				judge := sit(t, run, "judge")
				mustAppend(t, judge, docketRule("D2", "the bench rules"))
				_, err := Append(judge, finalDocketRule())
				return err
			},
			// One bench ruling per docketing, refused in words (requireUnanswered's docketing arm) and
			// not only at the column.
			wants:     []string{"docketing D2", "already"},
			docketing: "D2", ruling: "remanded",
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
				r := docketRule("D2", "the bench rules a petition")
				r.MotionId = proto.String("M3")
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
			if tc.docketing != "" {
				got = motionAnswer{ruling: docketingWord(t, run, tc.docketing)}
			}
			if got.ruling != tc.ruling {
				t.Errorf("motion %s / docketing %s reads ruled %q; want ruled %q", tc.motion, tc.docketing, got.ruling, tc.ruling)
			}
		})
	}
}

// motionGuardRun is a record with three unruled asks: proposal M1 on G1 proposing a regrade, which
// its assignee's window passed and docketing D1 holds; docketing D2 on G2 at impasse (both written as
// the dispatch writes them); and petition M3 (filed by the lens). The chair's bench row on G1 and G2
// then convenes the bench's next docket sitting on both.
func motionGuardRun(t *testing.T) Run {
	t.Helper()
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).
		mint(evLens, "G1", "high").mint(evLens, "G2", "high").seed()
	fileProposal(t, run, regradeMotion("M1", "G1", "severity is overstated"))
	if _, err := docketByDispatchAs(t, run, "D1", "G1", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "M1"); err != nil {
		t.Fatal(err)
	}
	if _, err := docketByDispatchAs(t, run, "D2", "G2", recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, ""); err != nil {
		t.Fatal(err)
	}
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

// docketingWord is the word of docketing id's ruling as the docketings fold states it, "" unruled.
func docketingWord(t *testing.T, run Run, id string) string {
	t.Helper()
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range DocketingsOf(all.Events) {
		if d.ID == id {
			return d.Word()
		}
	}
	t.Fatalf("no docketing %s on the record", id)
	return ""
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
	mustAppend(t, judge, heldBy(actRule("M1", false, "judge"), "D1"))
	mustAppend(t, judge, docketRule("D2", "the bench remands"))
	// A proposal on G2, a regrade, once the bench's remand lifted its docketing, left unruled by the
	// write.
	fileProposal(t, run, regradeMotion("M4", "G2", "likelihood is overstated"))

	motions := motionReadersAgree(t, run)
	if got, want := motions["M1"], (motionAnswer{ruling: "for assignee", by: "judge"}); got != want {
		t.Errorf("motion M1 reads %+v; the ruling that stands is %+v", got, want)
	}
	if got, ok := motions["M4"]; ok {
		t.Errorf("motion M4 reads %+v; nobody ruled it", got)
	}
	if got := docketingWord(t, run, "D2"); got != "remanded" {
		t.Errorf("docketing D2 reads ruled %q, want the bench's remand", got)
	}
}

// A PROPOSAL IS ANSWERED ONCE WHEN ITS ASSIGNEE ACCEPTS AS THE DISPATCH DOCKETS IT (T-B7). The guards
// read the record and then the write inserts; outside the writing transaction that is
// check-then-insert. The two writes that can end a proposal's window race, each on its own connection:
// the gap's assignee accepting it (refused while a docketing stands, L27), and the dispatch docketing it
// for the bench (refused once the proposal is answered, L28). Exactly one lands, the other is refused in
// its guard's own words, and the record holds either the accept or the docketing, never both.
func TestAMotionIsAnsweredOnceWhenSeatsWriteAtOnce(t *testing.T) {
	for round := 0; round < 5; round++ {
		st := newStage(t).cast("red-chair", "blue-respond", evLens, "judge").ingest().register("red-chair").dispatch(2, evLens)
		run := st.register(evLens).mint(evLens, "G1", "high").seed()
		fileProposal(t, run, regradeMotion("M1", "G1", "severity is overstated"))
		// The assignee sits BEFORE the race: a register is a write too, and the race under test is the
		// answer's and the docketing's, not the sitting's.
		lens := sit(t, run, evLens)
		start := make(chan struct{})
		var acceptErr, docketErr error
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			_, acceptErr = Append(lens, actRule("M1", true, evLens))
		}()
		go func() {
			defer wg.Done()
			<-start
			_, docketErr = docketByDispatchAs(t, run, "D1", "G1", recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL, "M1")
		}()
		close(start)
		wg.Wait()
		switch {
		case acceptErr == nil && docketErr == nil:
			t.Fatalf("round %d: the accept of M1 and its docketing both landed; want exactly one", round)
		case acceptErr != nil && docketErr != nil:
			t.Fatalf("round %d: neither landed: accept %v; docketing %v", round, acceptErr, docketErr)
		case acceptErr != nil && !strings.Contains(acceptErr.Error(), "gap G1 is docketed (D1)"):
			t.Errorf("round %d: the accept was refused for something other than the docketing standing: %v", round, acceptErr)
		case docketErr != nil && !strings.Contains(docketErr.Error(), "motion M1 is no docket motion standing unanswered"):
			t.Errorf("round %d: the docketing was refused for something other than the accept standing: %v", round, docketErr)
		}
		_, ruled := motionReadersAgree(t, run)["M1"]
		if ruled != (acceptErr == nil) {
			t.Errorf("round %d: M1 reads answered %v, and the accept landed %v", round, ruled, acceptErr == nil)
		}
	}
}

// heldBy is r naming docketing, the one holding the proposal it answers: the bench's ruling of a
// proposal its assignee did not accept in its window.
func heldBy(r *recordpb.MotionRule, docketing string) *recordpb.MotionRule {
	r.GetDocket().DocketingId = proto.String(docketing)
	return r
}

// forRed is r siding with red: the bench's answer against blue's proposal.
func forRed(r *recordpb.MotionRule) *recordpb.MotionRule {
	r.GetDocket().SidedWith = recordpb.SidedWith_SIDED_WITH_RED.Enum()
	return r
}

// regradeMotion is a docket motion on gap proposing its severity low, argued by basis.
func regradeMotion(id, gap, basis string) *recordpb.Motion {
	return &recordpb.Motion{MotionId: proto.String(id), Basis: proto.String(basis),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}},
		Act: &recordpb.Motion_Regrade{Regrade: &recordpb.Regrade{GapId: proto.String(gap),
			Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String(basis)}}}
}

// actRule answers a docket motion's proposed act, by seat: for the filer (a red seat's proposal), or
// for the assignee. The bench's answer carries its opinion; another seat's carries none
// (requireBenchOpinion). B2's shape: the bench's answer on a docketing names it (DocketRuling.docketing_id),
// which these callers are re-keyed to with the dispatch's docketings.
func actRule(id string, accepted bool, seat string) *recordpb.MotionRule {
	side := recordpb.SidedWith_SIDED_WITH_ASSIGNEE
	if accepted {
		side = recordpb.SidedWith_SIDED_WITH_FILER
	}
	d := &recordpb.DocketRuling{SidedWith: side.Enum()}
	if seat == "judge" {
		d.Principle, d.Tension, d.ReviewFlag, d.Settled = proto.String("pr"), proto.String("tn"), proto.String("rf"), proto.String("st")
	}
	return &recordpb.MotionRule{MotionId: proto.String(id), Opinion: proto.String("ruled"),
		Ruling: &recordpb.MotionRule_Docket{Docket: d}}
}
