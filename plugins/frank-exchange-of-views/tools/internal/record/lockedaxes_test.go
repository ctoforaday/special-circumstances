package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE AXES A GRADE-MOVEMENT RULING LOCKS ARE THE AXES ITS WINDOW MOVED (L35′, L41). benchSetAxesOf
// decides which axes a --keep or --restore ruling decided from its own fold of the party regrades since
// each axis's last ruling; gradeMovementOf is the window the docketing asked about. The two folds are
// held level: the axes locked after the ruling are exactly the axes the window held moves on before
// it, an axis moved and moved back included, an axis never moved excluded; and a second movement
// ruling re-locks at its own place, never lifting the first.
func TestTheLockedAxesAreTheWindowsMovedAxes(t *testing.T) {
	const lens = "red-lens-logic"
	regrade := func(sev, lik *recordpb.Grade) *recordpb.Regrade {
		return &recordpb.Regrade{GapId: proto.String("G1"), Severity: sev, Likelihood: lik, Basis: proto.String("b")}
	}
	keep := func(d string) *recordpb.MotionRule {
		return &recordpb.MotionRule{Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			DocketingId: proto.String(d), MovedGrades: recordpb.MovedGrades_MOVED_GRADES_KEPT.Enum(),
			Principle: proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s")}}}
	}
	docketing := func(d string) *recordpb.Docketing {
		return &recordpb.Docketing{DocketingId: proto.String(d), GapId: proto.String("G1"),
			Cause: recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT.Enum(), Basis: proto.String("grades moved")}
	}
	b := newStage(t)
	b.register(lens).add(lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_LOW))
	b.register(lens).add(lens, regrade(recordtest.P(recordpb.Grade_GRADE_HIGH), nil))
	b.register(lens).add(lens, regrade(nil, recordtest.P(recordpb.Grade_GRADE_HIGH)))
	b.register(lens).add(lens, regrade(nil, recordtest.P(recordpb.Grade_GRADE_MEDIUM))) // likelihood moved and back
	b.register("red-chair").add("red-chair", docketing("D1"))
	b.register("judge").add("judge", keep("D1"))
	// A SECOND RULING RE-LOCKS: blue's regrade proposal of the locked axis, ruled for blue — the bench's
	// regrade and the ruling in one transaction — leaves it locked at the later ruling's place and grade.
	b.register("blue-respond").add("blue-respond", &recordpb.Motion{MotionId: proto.String("M2"), Basis: proto.String("b"),
		Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String("G1")}},
		Act:    &recordpb.Motion_Regrade{Regrade: regrade(recordtest.P(recordpb.Grade_GRADE_MEDIUM), nil)}})
	b.register("red-chair").add("red-chair", &recordpb.Docketing{DocketingId: proto.String("D2"), GapId: proto.String("G1"),
		Cause: recordpb.DocketingCause_DOCKETING_CAUSE_PROPOSAL.Enum(), MotionId: proto.String("M2"), Basis: proto.String("unaccepted")})
	b.register("judge").add("judge", regrade(recordtest.P(recordpb.Grade_GRADE_MEDIUM), nil)).
		add("judge", &recordpb.MotionRule{MotionId: proto.String("M2"), Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			DocketingId: proto.String("D2"), SidedWith: recordpb.SidedWith_SIDED_WITH_BLUE.Enum(),
			Principle: proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s")}}})
	b.seed()

	db, err := openRunForRead(mustRun(t, b.runDir))
	if err != nil || db == nil {
		t.Fatalf("open: %v", err)
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	ids := win.IDs(evs)
	ruling := 0
	for evs[ruling].GetMotionRule() == nil {
		ruling++
	}
	gm := gradeMovementOf(evs[:ruling], ids[:ruling], win, map[string]bool{"G1": true})
	moved := map[string]bool{}
	for _, mv := range gm.moves["G1"] {
		moved[mv.axis] = true
	}
	axes := benchSetAxesOf(evs[:ruling+1], ids[:ruling+1])["G1"]
	for k := range 3 {
		axis := recordpb.Word(gradeAxes[k])
		if locked := axes[k].benchAt != 0; locked != moved[axis] {
			t.Errorf("%s: locked %v after the --keep ruling, and the window it ruled on moved it: %v", axis, locked, moved[axis])
		}
	}
	if !moved["severity"] || !moved["likelihood"] || moved["impact"] {
		t.Fatalf("the fixture's window = %v, want severity and likelihood moved, impact not", moved)
	}
	first := axes[0].benchAt

	again := benchSetAxesOf(evs, win.IDs(evs))["G1"]
	if again[0].benchAt == 0 || again[0].benchAt == first || again[0].grade != recordpb.Grade_GRADE_MEDIUM {
		t.Errorf("severity after a second ruling = %+v (first locked at %d), want re-locked at the second ruling's place at medium", again[0], first)
	}
}

// The fold under T-B36's reword rows (the rows themselves are driven through the verbs by internal/cli's
// TestNoLaterDocketingOverturnsARuling): A RULING'S LOCK ENDS WITH ITS TOMBSTONE (gblock 2026-11-21, "Ends
// with the tombstone"). R1 (--keep) locks severity and likelihood, R2 (for blue on a regrade proposal)
// re-locks severity; the row that ends G1's ruling tombstone — blue's reword's — ends every lock a
// ruling on G1 set, and a successor's inherited lock with them; a later ruling's placement revives no
// ended lock, and a later grade ruling locks afresh.
func TestARulingsLockEndsWithItsTombstone(t *testing.T) {
	const lens = "red-lens-logic"
	sev := recordtest.P(recordpb.Grade_GRADE_HIGH)
	keep := func(d string) *recordpb.MotionRule {
		return &recordpb.MotionRule{Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
			DocketingId: proto.String(d), MovedGrades: recordpb.MovedGrades_MOVED_GRADES_KEPT.Enum(),
			Principle: proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s")}}}
	}
	movement := func(d string) *recordpb.Docketing {
		return &recordpb.Docketing{DocketingId: proto.String(d), GapId: proto.String("G1"),
			Cause: recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT.Enum(), Basis: proto.String("grades moved")}
	}
	stone := func(d string, ends bool) *recordpb.Tombstone {
		s := &recordpb.Tombstone{GapId: proto.String("G1"), Kind: recordpb.TombstoneKind_TOMBSTONE_KIND_RULING.Enum()}
		if ends {
			s.Ends = proto.Bool(true)
		} else {
			s.DocketingId = proto.String(d)
		}
		return s
	}
	b := newStage(t)
	b.register(lens).add(lens, liveMint("G1", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_LOW))
	b.register(lens).add(lens, &recordpb.Regrade{GapId: proto.String("G1"), Severity: sev, Likelihood: sev, Basis: proto.String("b")})
	b.register("red-chair").add("red-chair", movement("D1"))
	b.register("judge").add("judge", keep("D1")).add("judge", stone("D1", false))
	b.register(lens).add(lens, &recordpb.Regrade{GapId: proto.String("G1"), Impact: sev, Basis: proto.String("b")})
	b.register("red-chair").add("red-chair", movement("D2"))
	b.register("judge").add("judge", keep("D2")).add("judge", stone("D2", false))
	b.register(lens).add(lens, func() *recordpb.Mint {
		m := liveMint("G2", recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION, recordpb.Grade_GRADE_HIGH)
		m.Supersedes = proto.String("G1")
		return m
	}())
	b.register("blue-respond").add("blue-respond", stone("", true))
	b.register("red-chair").add("red-chair", &recordpb.Docketing{DocketingId: proto.String("D3"), GapId: proto.String("G1"),
		Cause: recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE.Enum(), Basis: proto.String("impasse")})
	b.register("judge").add("judge", &recordpb.MotionRule{Opinion: proto.String("o"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
		DocketingId: proto.String("D3"), SidedWith: recordpb.SidedWith_SIDED_WITH_BLUE.Enum(), Disposition: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Final: proto.Bool(true), Principle: proto.String("p"), Tension: proto.String("t"), ReviewFlag: proto.String("none"), Settled: proto.String("s")}}}).
		add("judge", stone("D3", false))
	b.seed()

	db, err := openRunForRead(mustRun(t, b.runDir))
	if err != nil || db == nil {
		t.Fatalf("open: %v", err)
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	ids := win.IDs(evs)
	end := 0
	for evs[end].GetTombstone() == nil || !evs[end].GetTombstone().GetEnds() {
		end++
	}
	locked := func(axes [4]axisGrade) []string {
		var out []string
		for k, ax := range axes {
			if ax.benchAt != 0 {
				out = append(out, recordpb.Word(gradeAxes[k]))
			}
		}
		return out
	}
	before := benchSetAxesOf(evs[:end], ids[:end])
	if got := locked(before["G1"]); len(got) != 3 {
		t.Fatalf("before the end, G1's locked axes = %v, want severity, likelihood (R1) and impact (R2)", got)
	}
	if got := locked(before["G2"]); len(got) != 3 {
		t.Fatalf("before the end, the successor G2 inherits %v, want G1's three", got)
	}
	after := benchSetAxesOf(evs, ids)
	if got := locked(after["G1"]); len(got) != 0 {
		t.Errorf("after the end row and a later closing ruling, G1's locked axes = %v, want none: the lock ended with its tombstone and a ruling on no grade revives none", got)
	}
	if got := locked(after["G2"]); len(got) != 0 {
		t.Errorf("after the end row, the successor's inherited lock = %v, want none", got)
	}
}
