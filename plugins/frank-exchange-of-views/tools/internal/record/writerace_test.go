package record

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// overtaken runs write while another seat's acts commit at the instant the write is about to open
// its transaction (WriteBegins), in one commit on that seat's own connection (seatWriter) — as a
// seat in another process lands its act. The acts land once, at the first write to begin. A check
// made before the write's transaction has already been made by then and admits the write over them;
// a check made inside it reads them.
func overtaken(t *testing.T, run Run, acts []*Event, write func() error) error {
	t.Helper()
	db := seatWriter(t, filepath.Join(run.Records(), "record.db"))
	var once sync.Once
	was := WriteBegins
	t.Cleanup(func() { WriteBegins = was })
	WriteBegins = func() {
		once.Do(func() {
			tx, err := db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			for _, ev := range acts {
				// What the other seat's write path stamps: the time, and a mint's class material.
				ev.Ts = proto.String(stamp())
				if m := ev.GetMint(); m != nil && m.ClassMaterial == nil {
					m.ClassMaterial = recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE.Enum()
				}
				if _, err := recordsql.InsertTx(tx, ev); err != nil {
					t.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		})
	}
	return write()
}

// raceStage is a run every guard below can be asked on: the cast, the report, and a chair, a lens,
// blue and the bench each in a sitting, with G1 and G2 minted by the lens (both medium) and G1
// docketed as M1.
func raceStage(t *testing.T) Run {
	b := newStage(t).cast(evLens, "red-lens-logic", "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").register(evLens).
		mint(evLens, "G1", "medium").mint(evLens, "G2", "medium").
		docketMotion("red-chair", "M1", "G1").
		register("blue-respond").register("judge").register("red-lens-logic")
	return b.seed()
}

func gradeMotion(id, gap string, proposed recordpb.Grade) *recordpb.Motion {
	return &recordpb.Motion{MotionId: proto.String(id),
		Basis: proto.String("the likelihood is overstated"), Filing: &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String(gap),
			Dimension: recordpb.GradeDimension_GRADE_DIMENSION_LIKELIHOOD.Enum(), Proposed: proposed.Enum()}}}
}

func ordinalMint(gap, problem string) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_POLICY_WITHOUT_MECHANISM.Enum(), Problem: proto.String(problem),
		Location: proto.String("91 is the twenty-seventh semiprime."), RequiredFix: proto.String("fix"), AcceptanceCheck: proto.String("chk"),
		CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: recordtest.P(recordpb.Grade_GRADE_MEDIUM),
		Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
}

func highMint(gap string) *recordpb.Mint {
	return &recordpb.Mint{GapId: proto.String(gap), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("a material defect " + gap),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_HIGH), Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Impact: recordtest.P(recordpb.Grade_GRADE_HIGH)}
}

// A WRITE IS CHECKED AGAINST THE RECORD IT LANDS ON (#1252). Every guard on the write path that
// another seat's commit can turn from admitting to refusing is asked here three ways, each on a fresh
// run: before the other seat's acts (admitted — the fixture does not refuse on its own), after them
// (refused — the acts are what refuses), and with the acts committing on their own connection at the
// instant the write opens its transaction, which must answer as "after" does. Checked before the
// transaction, the third admits: the check has already passed over a record that no longer holds.
func TestAWriteIsCheckedAgainstTheRecordItLandsOn(t *testing.T) {
	for _, c := range []struct {
		name  string
		stage func(t *testing.T) Run
		// acts are the other seat's, committed beside the write.
		acts func(t *testing.T) []*Event
		seat string
		body func() proto.Message
		want string
	}{
		{
			name:  "a PASS while a lens mints a material gap",
			stage: func(t *testing.T) Run { return convergedBoard(t).seed() },
			acts:  func(t *testing.T) []*Event { return []*Event{recordtest.At(t, evLens, "race:mint:G9", highMint("G9"))} },
			seat:  "red-chair", body: func() proto.Message { return proto.Clone(passGate) },
			want: "PASS refused",
		},
		{
			name: "a FAIL while the lens closes the last material gap",
			stage: func(t *testing.T) Run {
				st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
					register("red-chair").register(evLens).add(evLens, highMint("G1")).
					add("red-chair", failVerdict()).register("red-chair").register(evLens)
				low, _ := GradeOf("low")
				st.add(evLens, &recordpb.Mint{GapId: proto.String("G2"), Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"), AcceptanceCheck: proto.String("c"),
					CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT), Severity: &low,
					Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Impact: recordtest.P(recordpb.Grade_GRADE_LOW)})
				return st.register("red-chair").sit(2, evLens).sit(2, evLens).register("red-chair").seed()
			},
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, evLens, "race:close:G1", repairClose("G1"))}
			},
			seat: "red-chair", body: func() proto.Message { return failVerdict() },
			want: "the board has converged",
		},
		{
			name: "the outcome while a party petitions the bench",
			stage: func(t *testing.T) Run {
				return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
					register("red-chair").register(evLens).register("judge").seed()
			},
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "blue-respond", "race:motion:M7", petitionMotion("M7"))}
			},
			seat: "judge", body: func() proto.Message {
				return &recordpb.Outcome{Verdict: recordpb.RunOutcome_RUN_OUTCOME_UNVERIFIED.Enum(), Prose: proto.String("ended")}
			},
			want: "M7 (petition, ruled by the bench seat)",
		},
		{
			name: "a mint while another lens mints the same defect",
			stage: func(t *testing.T) Run {
				return newStage(t).cast("red-lens-dark-side", "red-lens-computation").ingest().
					register("red-lens-dark-side").register("red-lens-computation").seed()
			},
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "red-lens-dark-side", "race:mint:G1",
					ordinalMint("G1", "91 is stated as the twenty-seventh semiprime and the second of the form 7·q; both ordinals are wrong."))}
			},
			seat: "red-lens-computation", body: func() proto.Message {
				return ordinalMint("G2", "The ordinal position of 91 among semiprimes is stated wrongly: twenty-seventh and second of the form 7·q.")
			},
			want: "the board already holds an open gap this may be",
		},
		{
			name: "a lens's two mints at once, one slot left",
			stage: func(t *testing.T) Run {
				dir := newRun(t)
				writeRunConfig(t, dir, `{"mintBudget":1}`)
				recordtest.Seed(t, dir, recordtest.At(t, evLens, "race:register", &recordpb.Register{}))
				return mustRun(t, dir)
			},
			acts: func(t *testing.T) []*Event { return []*Event{recordtest.At(t, evLens, "race:mint:G1", highMint("G1"))} },
			seat: evLens, body: func() proto.Message { return budgetMint("G2") },
			want: "holds 1 gap(s) open and its mint budget admits no more",
		},
		{
			name:  "a close while the bench's ruling closes the gap",
			stage: raceStage,
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "judge", "race:rule:M1", benchClosing("M1"))}
			},
			seat: evLens, body: func() proto.Message { return repairClose("G1") },
			want: "--id names gap G1, which is already CLOSED",
		},
		{
			name:  "a close superseded by a gap the bench closes",
			stage: raceStage,
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "judge", "race:rule:M1", benchClosing("M1"))}
			},
			seat: evLens, body: func() proto.Message {
				c := repairClose("G2")
				c.ClosureClass = recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION.Enum()
				c.Successor = proto.String("G1")
				return c
			},
			want: "--superseded-by names gap G1, which is already CLOSED",
		},
		{
			name:  "a regrade while the bench's ruling closes the gap",
			stage: raceStage,
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "judge", "race:rule:M1", benchClosing("M1"))}
			},
			seat: evLens, body: func() proto.Message {
				return &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("smaller than it read")}
			},
			want: "a grade only decides what happens NEXT",
		},
		{
			name:  "a grade motion while the bench's ruling closes the gap",
			stage: raceStage,
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, "judge", "race:rule:M1", benchClosing("M1"))}
			},
			seat: "blue-respond", body: func() proto.Message { return gradeMotion("M8", "G1", recordpb.Grade_GRADE_LOW) },
			want: "a grade motion asks for a DIFFERENT disposition",
		},
		{
			name:  "a grade motion while the lens regrades to what it proposes",
			stage: raceStage,
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, evLens, "race:regrade:G2", &recordpb.Regrade{GapId: proto.String("G2"),
					Likelihood: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("rarer than it read")})}
			},
			seat: "blue-respond", body: func() proto.Message { return gradeMotion("M8", "G2", recordpb.Grade_GRADE_LOW) },
			want: "so this motion asks for no change",
		},
		{
			name:  "the cast while setup's cast lands",
			stage: func(t *testing.T) Run { return mustRun(t, newRun(t)) },
			acts: func(t *testing.T) []*Event {
				return []*Event{recordtest.At(t, HarnessSeat, "race:cast", &recordpb.Cast{SeatIds: []string{"red-chair", evLens}})}
			},
			seat: HarnessSeat, body: func() proto.Message { return &recordpb.Cast{SeatIds: []string{"red-chair", evLens}} },
			want: "the cast is written ONCE",
		},
		{
			name:  "blue's edit naming a gap a lens mints",
			stage: raceStage,
			acts:  func(t *testing.T) []*Event { return []*Event{recordtest.At(t, evLens, "race:mint:G5", highMint("G5"))} },
			seat:  "blue-respond", body: func() proto.Message {
				return &recordpb.BlueEdit{Old: proto.String("was"), New: proto.String("is"), EditKey: proto.String("E-race"),
					Text: proto.String("G5: the sentence overclaims")}
			},
			want: "--reason names gap G5 but --answers is empty",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			before := c.stage(t)
			if _, err := Append(Identity{Run: before, SeatID: c.seat}, c.body()); err != nil {
				t.Fatalf("fixture: refused with nobody else writing: %v", err)
			}
			after := c.stage(t)
			recordtest.Seed(t, after.Dir(), c.acts(t)...)
			if _, err := Append(Identity{Run: after, SeatID: c.seat}, c.body()); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("fixture: written after the other seat's acts = %v, want a refusal containing %q", err, c.want)
			}
			raced := c.stage(t)
			err := overtaken(t, raced, c.acts(t), func() error {
				_, err := Append(Identity{Run: raced, SeatID: c.seat}, c.body())
				return err
			})
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("written while the other seat's acts committed = %v, want the refusal the record they land on gives (%q) — the check read a record the write did not land on", err, c.want)
			}
		})
	}
}
