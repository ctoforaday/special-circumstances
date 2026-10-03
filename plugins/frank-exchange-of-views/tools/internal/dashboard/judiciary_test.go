package dashboard

import (
	"reflect"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

func judMint(t *testing.T, seat, id string, l, i recordpb.Grade, supersedes ...string) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, seat, &recordpb.Mint{
		GapId:           proto.String(id),
		Problem:         proto.String("p"),
		RequiredFix:     proto.String("f"),
		AcceptanceCheck: proto.String("the check runs"),
		Class:           proto.String("self-attestation"),
		CheckKind:       recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity:        recordtest.P(recordpb.Grade_GRADE_HIGH),
		Likelihood:      recordtest.P(l),
		Impact:          recordtest.P(i),
		Supersedes:      supersedes,
	})
}

func docketFiled(t *testing.T, motion, gap string) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.Motion{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Basis:    proto.String("impasse"),
		Filing:   &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(gap)}},
	})
}

func docketRuled(t *testing.T, motion string, d recordpb.Disposition, principle, opinion string) *recordpb.Event {
	t.Helper()
	r := &recordpb.DocketRuling{Disposition: recordtest.P(d), Principle: proto.String(principle), Tension: proto.String(""), ReviewFlag: proto.String(""), Settled: proto.String("whether the gap stands")}
	if d == recordpb.Disposition_DISPOSITION_REMANDED {
		r.ReopensOn = proto.String("the stated direction reporting back")
	} else {
		r.Final = proto.Bool(true)
	}
	return recordtest.Event(t, "judge", &recordpb.MotionRule{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_DOCKET),
		Opinion:  proto.String(opinion),
		Ruling:   &recordpb.MotionRule_Docket{Docket: r},
	})
}

func gradeFiled(t *testing.T, motion, gap string) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, "blue-respond", &recordpb.Motion{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Basis:    proto.String("over-graded"),
		Filing:   &recordpb.Motion_Grade{Grade: &recordpb.GradeMotion{GapId: proto.String(gap)}},
	})
}

func gradeRuled(t *testing.T, motion string, r recordpb.GradeRuling) *recordpb.Event {
	t.Helper()
	return recordtest.Event(t, "red-chair", &recordpb.MotionRule{
		MotionId: proto.String(motion),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_GRADE),
		Opinion:  proto.String("heard"),
		Ruling:   &recordpb.MotionRule_Grade{Grade: r},
	})
}

// judiciaryRun is a record holding every fact the Judiciary table shows: a docket sitting with
// two rulings, two grade motions answered both ways, a FAIL gate in epochs 1 and 2 and none yet
// in 3, and a chain — G2 supersedes G1, which was minted HIGH/HIGH and regraded LOW/LOW and is
// still open — beside a gap the bench closed in the epoch it was minted. A PASS would make the
// run terminal and put the final verdict in the tile, so the latest one here is a FAIL.
func judiciaryRun(t *testing.T) string {
	t.Helper()
	hi, med, lo := recordpb.Grade_GRADE_HIGH, recordpb.Grade_GRADE_MEDIUM, recordpb.Grade_GRADE_LOW
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		registered(t, "red-chair", "C1"), // opens epoch 1
		judMint(t, "red-lens-evidence", "G1", hi, hi),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}),
		registered(t, "red-chair", "C2"), // opens epoch 2
		judMint(t, "red-lens-logic", "G2", med, med, "G1"),
		judMint(t, "red-lens-logic", "G3", lo, lo),
		recordtest.Event(t, "red-lens-evidence", &recordpb.Regrade{GapId: proto.String("G1"), Likelihood: &lo, Impact: &lo, Basis: proto.String("the source was found")}),
		docketFiled(t, "M1", "G1"),
		docketFiled(t, "M2", "G3"),
		recordtest.Event(t, "judge", &recordpb.Register{AgentId: proto.String("J1"), Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum()}),
		docketRuled(t, "M1", recordpb.Disposition_DISPOSITION_REMANDED, "a deferral names its reporter", "remanded to the lane that owns it"),
		docketRuled(t, "M2", recordpb.Disposition_DISPOSITION_REPAIRED, "a repair is read at the leaf", "the repair holds"),
		gradeFiled(t, "M3", "G2"),
		gradeRuled(t, "M3", recordpb.GradeRuling_GRADE_RULING_ACCEPTED),
		gradeFiled(t, "M4", "G3"),
		gradeRuled(t, "M4", recordpb.GradeRuling_GRADE_RULING_REJECTED),
		recordtest.Event(t, "red-chair", &recordpb.Gate{Verdict: recordpb.Verdict_VERDICT_FAIL.Enum()}),
		registered(t, "red-chair", "C3"), // opens epoch 3: the chair sits, no verdict yet
	)
	return dir
}

// EVERY JUDICIARY FIGURE COMES OFF THE RECORD. The run's journal is empty here, as it is in effect
// on every real run: no seat's envelope schema declares `gaps` or `dispute_responses`, so a
// reader of those keys printed no verdict, no chains and no dispute answers over a record that
// holds all three. One subtest per figure, so a figure that falls back to zero names itself.
func TestTheJudiciaryReadsTheRecord(t *testing.T) {
	m := BuildModel(runtest.Open(t, judiciaryRun(t)), t.TempDir(), Config{}, 0)
	j := m.Judiciary

	t.Run("judge sittings", func(t *testing.T) {
		if j.JudgeSittings != 1 {
			t.Errorf("JudgeSittings = %d, want 1 (one docket register)", j.JudgeSittings)
		}
	})
	t.Run("rulings by disposition", func(t *testing.T) {
		if want := map[string]int{"remanded": 1, "repaired": 1}; !reflect.DeepEqual(j.Rulings, want) {
			t.Errorf("Rulings = %v, want %v", j.Rulings, want)
		}
	})
	t.Run("grade motions", func(t *testing.T) {
		if d := j.Disputes; d.Raised != 2 || d.Accepted != 1 || d.Rejected != 1 {
			t.Errorf("Disputes = %+v, want raised 2 · accepted 1 · rejected 1", d)
		}
	})
	t.Run("latest verdict", func(t *testing.T) {
		if j.LatestVerdict != "FAIL" || j.VerdictEpoch != 2 {
			t.Errorf("verdict = %q at epoch %d, want FAIL at epoch 2", j.LatestVerdict, j.VerdictEpoch)
		}
	})
	t.Run("argument chains", func(t *testing.T) {
		if want := map[int]int{1: 1, 3: 1}; j.Chains != 2 || !reflect.DeepEqual(j.ChainSpans, want) {
			t.Errorf("Chains = %d spans %v, want 2 chains spanning %v (G1→G2 open, epochs 1-3; G3 minted and closed in 2)", j.Chains, j.ChainSpans, want)
		}
	})
	t.Run("grade migration", func(t *testing.T) {
		if j.MigDown != 1 || j.MigUp != 0 || j.MigFlat != 0 {
			t.Errorf("migration down/up/flat = %d/%d/%d, want 1/0/0 (the G1 chain fell from HIGH/HIGH)", j.MigDown, j.MigUp, j.MigFlat)
		}
	})
	t.Run("the page shows them", func(t *testing.T) {
		h := RenderHTML(m)
		for _, want := range []string{"latest verdict (epoch 2)", "remanded: 1", "raised 2 · accepted 1 · rejected 1", "2 chains"} {
			if !strings.Contains(h, want) {
				t.Errorf("the rendered page lacks %q", want)
			}
		}
	})
}
