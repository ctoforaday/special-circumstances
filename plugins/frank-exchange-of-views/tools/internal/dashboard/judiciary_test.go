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
// two rulings, two grade motions answered both ways, a FAIL gate in epochs 1 and 2 and an epoch
// 3 the chair has opened and nothing has happened in, and a chain — G2 supersedes G1, which was
// minted HIGH/HIGH and regraded LOW/LOW and is still open — beside a gap the bench closed in the
// epoch it was minted. A PASS would make the run terminal and put the final verdict in the tile,
// so the latest one here is a FAIL.
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
	// AN OPEN GAP LIVES TO THE CURRENT EPOCH, the last with work in it. Epoch 3 is a chair
	// register and nothing else, and counting it would lengthen every open chain by one.
	t.Run("argument chains", func(t *testing.T) {
		if want := map[int]int{1: 1, 2: 1}; j.Chains != 2 || !reflect.DeepEqual(j.ChainSpans, want) {
			t.Errorf("Chains = %d spans %v, want 2 chains spanning %v (G1→G2 open, epochs 1-2; G3 minted and closed in 2)", j.Chains, j.ChainSpans, want)
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

// EVERY BENCH SITTING IS A JUDGE SITTING, whatever it was convened to do: a petition hearing and
// the assembly sit as surely as a docket ruling. A register that repairs a sitting opens none.
func TestEveryBenchSittingIsCounted(t *testing.T) {
	bench := func(agent string, o recordpb.Occasion) *recordpb.Event {
		return recordtest.Event(t, "judge", &recordpb.Register{AgentId: proto.String(agent), Occasion: o.Enum()})
	}
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		registered(t, "red-chair", "C1"),
		bench("J1", recordpb.Occasion_OCCASION_DOCKET),
		recordtest.Event(t, "judge", &recordpb.Register{AgentId: proto.String("J1b"),
			Occasion: recordpb.Occasion_OCCASION_DOCKET.Enum(), RepairsSitting: proto.String("judge:register:#1")}),
		bench("J2", recordpb.Occasion_OCCASION_PETITION),
		bench("J3", recordpb.Occasion_OCCASION_TERMINAL),
		bench("J4", recordpb.Occasion_OCCASION_ASSEMBLE),
	)
	if n := BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0).Judiciary.JudgeSittings; n != 4 {
		t.Errorf("JudgeSittings = %d, want 4 (docket, petition, terminal, assemble; the repair opens none)", n)
	}
}

// A GAP SUPERSEDING SEVERAL ANCESTORS JOINS THEM ALL: G3 supersedes G1 and G2, so the three are one
// argument, not G1 alone beside G2 and G3.
func TestASupersedingGapJoinsEveryAncestorsChain(t *testing.T) {
	hi := recordpb.Grade_GRADE_HIGH
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir,
		registered(t, "red-chair", "C1"),
		judMint(t, "red-lens-evidence", "G1", hi, hi),
		judMint(t, "red-lens-logic", "G2", hi, hi),
		judMint(t, "red-lens-logic", "G3", hi, hi, "G1", "G2"),
	)
	if j := BuildModel(runtest.Open(t, dir), t.TempDir(), Config{}, 0).Judiciary; j.Chains != 1 {
		t.Errorf("Chains = %d (spans %v), want 1: G3 supersedes both G1 and G2", j.Chains, j.ChainSpans)
	}
}

// AN UNREAD RECORD IS NOT A BENCH THAT NEVER SAT. The scorecards on the same page say "not
// measured" for a record they cannot read; the Judiciary block and the verdict tile say the same,
// not "no judge sittings yet" and "—".
func TestTheJudiciaryOnAnUnreadRecordIsNotMeasured(t *testing.T) {
	m := BuildModel(runtest.Open(t, recordtest.TmpRun(t)), t.TempDir(), Config{}, 0)
	if m.Judiciary.Measured {
		t.Fatal("Judiciary.Measured with no record to read")
	}
	h := RenderHTML(m)
	for _, want := range []string{"<b>not measured</b><span>latest verdict</span>", "not measured — the record could not be read"} {
		if !strings.Contains(h, want) {
			t.Errorf("the rendered page lacks %q", want)
		}
	}
	if strings.Contains(h, "no judge sittings") {
		t.Error("the page reports a bench that never sat over a record it could not read")
	}
}
