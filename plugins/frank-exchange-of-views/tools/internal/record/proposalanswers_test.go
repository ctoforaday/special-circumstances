package record

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DOCKET MOTION IS A PROPOSED ACT, AND ITS ANSWER FITS IT. These tests drive the answers through
// record.Append — the write every writer shares — on motionGuardRun's record: docketing D1 holds M1,
// the chair's proposal of G1's regrade, docketing D2 stands on G2 at impasse, and the bench is
// convened on both. windowRun's record holds M1 in its assignee's window, before any docketing.

// windowRun is a record holding the chair's proposal M1 of G1's severity low, in its assignee's
// window: no docketing holds it, and the assignee may accept it.
func windowRun(t *testing.T) Run {
	t.Helper()
	return proposalRun(t, func(g string) *recordpb.Motion { return regradeMotion("M1", g, "severity is overstated") })
}

// benchFor is the bench's ruling of docketing, holding motion ("" where it holds none), siding with
// side, with its four opinion fields.
func benchFor(docketing, motion string, side recordpb.SidedWith) *recordpb.MotionRule {
	return ruling(docketing, motion, func(d *recordpb.DocketRuling) { d.SidedWith = side.Enum() })
}

// benchRemand is the bench's remand of docketing, holding motion ("" where it holds none).
func benchRemand(docketing, motion string) *recordpb.MotionRule {
	return ruling(docketing, motion, func(d *recordpb.DocketRuling) {
		d.Disposition, d.Direction = recordpb.Disposition_DISPOSITION_REMANDED.Enum(), proto.String("one more exchange")
	})
}

const (
	forFiler    = recordpb.SidedWith_SIDED_WITH_AUTHOR
	forAssignee = recordpb.SidedWith_SIDED_WITH_RAISER
	forBlue     = recordpb.SidedWith_SIDED_WITH_AUTHOR
)

// gapRow is the gap view's state columns for one gap, as one comparable value.
type gapRow struct {
	open, byBench, benchLast bool
	benchClose               sql.NullInt64
	mintedBy, severity       string
}

func gapRowOf(t *testing.T, run Run, g string) gapRow {
	t.Helper()
	var r gapRow
	found, err := queryRowAt(recordOf(t, run), []any{&r.open, &r.byBench, &r.benchLast, &r.benchClose, &r.mintedBy, &r.severity},
		`SELECT "open", COALESCE("closed_by_bench", 0), COALESCE("bench_closed_last", 0), "bench_close_event",
		        COALESCE("minted_by", ''), COALESCE("current_severity", '') FROM "gap" WHERE "gap_id" = ?`, g)
	if err != nil || !found {
		t.Fatalf("gap %s: found %v, %v", g, found, err)
	}
	return r
}

// withoutOpinion is r with its bench opinion field n cleared.
func withoutOpinion(r *recordpb.MotionRule, n string) *recordpb.MotionRule {
	r = proto.Clone(r).(*recordpb.MotionRule)
	d := r.GetDocket()
	switch n {
	case "principle":
		d.Principle = nil
	case "tension":
		d.Tension = nil
	case "review_flag":
		d.ReviewFlag = nil
	case "settled":
		d.Settled = nil
	}
	return r
}

// T-B6, and the ruler check: an answer fits the docketing or the proposal it answers — the bench's
// names its docketing, the proposal that docketing holds and no other, and a side of its pair; no other
// seat writes a docket ruling — the raiser takes a proposal by its own act (T-B48). Each refusal leaves the motion
// and the docketing standing, and the Go fold and the SQL view agree on it.
func TestADocketAnswerFitsTheMotionItAnswers(t *testing.T) {
	assigneeOnD2 := &recordpb.MotionRule{Opinion: proto.String("ruled"), Ruling: &recordpb.MotionRule_Docket{Docket: &recordpb.DocketRuling{
		DocketingId: proto.String("D2"), SidedWith: recordpb.SidedWith_SIDED_WITH_RAISER.Enum()}}}
	keepOnD2 := ruling("D2", "", func(d *recordpb.DocketRuling) { d.MovedGrades = recordpb.MovedGrades_MOVED_GRADES_KEPT.Enum() })
	closingWord := benchFor("D1", "M1", forFiler)
	closingWord.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
	for _, tc := range []struct {
		name   string
		run    func(*testing.T) Run
		seat   string
		body   *recordpb.MotionRule
		wants  []string
		motion string
		// docketing is the docketing that stands unruled after the refusal, "" for none.
		docketing string
	}{
		{"the raiser's answer to an impasse docketing", motionGuardRun, evLens, assigneeOnD2,
			[]string{"a docket ruling is the bench's"}, "", "D2"},
		{"a grade answer on an impasse docketing", motionGuardRun, "judge", keepOnD2,
			[]string{"docketing D2 is no grade movement"}, "", "D2"},
		{"a bench ruling naming a motion its docketing does not hold", motionGuardRun, "judge", benchFor("D2", "M1", forBlue),
			[]string{"docketing D2 holds no proposal, so its ruling names no motion; it names M1"}, "M1", "D2"},
		{"a bench ruling naming no docketing", motionGuardRun, "judge", actRule("M1", true, "judge"),
			[]string{"a bench docket ruling names the docketing it answers"}, "M1", "D1"},
		{"a closing word on a regrade proposal", motionGuardRun, "judge", closingWord,
			[]string{"--for-author on docketing D1 writes no closing word", "not_a_defect"}, "M1", "D1"},
		{"the raiser's ruling once its gap is docketed", motionGuardRun, evLens, actRule("M1", true, evLens),
			[]string{"a docket ruling is the bench's", "the regrade or close it proposes"}, "M1", "D1"},
		{"the raiser's ruling in its window", windowRun, evLens, actRule("M1", false, evLens),
			[]string{"a docket ruling is the bench's", "the regrade or close it proposes"}, "M1", ""},
		{"the author's ruling", windowRun, "blue-respond", actRule("M1", true, "blue-respond"),
			[]string{"a docket ruling is the bench's"}, "M1", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := tc.run(t)
			_, err := Append(sit(t, run, tc.seat), tc.body)
			mustRefuse(t, err, tc.wants...)
			if tc.motion != "" {
				if got := motionReadersAgree(t, run)[tc.motion]; got.ruling != "" {
					t.Errorf("motion %s reads ruled %q after a refused answer; want it standing", tc.motion, got.ruling)
				}
			}
			if tc.docketing != "" {
				if got := docketingWord(t, run, tc.docketing); got != "" {
					t.Errorf("docketing %s reads ruled %q after a refused answer; want it standing", tc.docketing, got)
				}
			}
		})
	}
}

// T-B15: ONLY THE BENCH OWES AN OPINION. A bench ruling missing any of its four opinion fields is
// refused naming it — each answer these records offer: for the filer and for the assignee on D1's
// proposal, for blue on D2's impasse, and a remand of each; the assignee's accept in its window lands
// with none and is refused carrying any.
func TestOnlyTheBenchOwesAnOpinion(t *testing.T) {
	fields := []struct{ name, flag string }{
		{"principle", "--principle"}, {"tension", "--tension"}, {"review_flag", "--review-flag"}, {"settled", "--settled"},
	}
	// regradeFirst writes the bench's regrade the filer's side applies, before its ruling.
	regradeFirst := func(t *testing.T, bench Identity) {
		mustAppend(t, bench, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("as proposed")})
	}
	// d1First rules D1, the docket's first, so D2 is first when its ruling is written.
	d1First := func(t *testing.T, bench Identity) {
		mustAppend(t, bench, benchFor("D1", "M1", forAssignee))
	}
	for _, answer := range []struct {
		name string
		pre  func(*testing.T, Identity)
		body *recordpb.MotionRule
	}{
		{"for the filer on D1", regradeFirst, benchFor("D1", "M1", forFiler)},
		{"for the assignee on D1", nil, benchFor("D1", "M1", forAssignee)},
		{"remand of D1", nil, benchRemand("D1", "M1")},
		{"for blue on D2", d1First, ruling("D2", "", func(d *recordpb.DocketRuling) {
			d.SidedWith, d.Disposition, d.Final = forBlue.Enum(), recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), proto.Bool(true)
		})},
		{"remand of D2", d1First, benchRemand("D2", "")},
	} {
		write := func(t *testing.T, body *recordpb.MotionRule) error {
			run := motionGuardRun(t)
			bench := sit(t, run, "judge")
			if answer.pre != nil {
				answer.pre(t, bench)
			}
			_, err := Append(bench, body)
			return err
		}
		for _, f := range fields {
			t.Run("the bench's ruling "+answer.name+" without "+f.name, func(t *testing.T) {
				mustRefuse(t, write(t, withoutOpinion(answer.body, f.name)), "requires "+f.flag)
			})
		}
		t.Run("the bench's ruling "+answer.name+" whole", func(t *testing.T) {
			if err := write(t, answer.body); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// proposalRun is a cast of two lenses, the chair, blue and the bench, with G1 minted by the evidence
// lens and the chair's proposal act("G1") filed on it, in its assignee's window.
func proposalRun(t *testing.T, act func(gap string) *recordpb.Motion) Run {
	t.Helper()
	run := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "high").seed()
	fileProposal(t, run, act("G1"))
	return run
}

func closeProposalOn(class recordpb.Disposition) func(string) *recordpb.Motion {
	return func(g string) *recordpb.Motion {
		return &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("b"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(g)}},
			Act: &recordpb.Motion_Close{Close: &recordpb.Close{GapId: proto.String(g), ClosureClass: class.Enum(),
				Prose: proto.String("p"), AnchorSeat: proto.String("blue-respond"), AnchorTool: proto.String("the report read"),
				AnchorTarget: proto.String("the report's sentence")}}}
	}
}

// T-B8: THE BENCH'S ANSWER PER ACT. Siding with the filer applies the proposal as filed: a close with
// its stored word, a regrade at its grades (the bench's regrade, before the ruling); siding with the
// assignee leaves G as it stood.
func TestTheBenchAnswersEachActItsOwnWay(t *testing.T) {
	t.Run("close: the closing word the proposal stores", func(t *testing.T) {
		run := proposalRun(t, closeProposalOn(recordpb.Disposition_DISPOSITION_MOOT))
		d := docketProposal(t, run, "G1", "M1")
		bench := convene(t, run, "G1")
		// The filer's side writes the stored word, and no ruling lands without it.
		_, err := Append(bench, benchFor(d, "M1", forFiler))
		mustRefuse(t, err, "--for-author on docketing "+d+" writes moot, the word that side stated")
		// Another closing word than the stored one is no statement of the filer's.
		other := benchFor(d, "M1", forFiler)
		other.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
		_, err = Append(bench, other)
		mustRefuse(t, err, "--for-author on docketing "+d+" writes moot", `carries "not_a_defect"`)
		stored := benchFor(d, "M1", forFiler)
		stored.GetDocket().Disposition, stored.GetDocket().Final = recordpb.Disposition_DISPOSITION_MOOT.Enum(), proto.Bool(true)
		mustAppend(t, bench, stored)
		g := gapRowOf(t, run, "G1")
		if id := rulingEventOf(t, run, "M1"); g.open || !g.byBench || !g.benchClose.Valid || g.benchClose.Int64 != id {
			t.Errorf("after the bench's moot: %+v, want closed by the bench at event %d", g, id)
		}
		if w := closureReasonOf(t, run, "G1"); w != "moot" {
			t.Errorf("G1's closure word = %q, want the stored moot", w)
		}
	})
	t.Run("regrade: the bench's regrade, then its ruling for the filer", func(t *testing.T) {
		run := motionGuardRun(t)
		bench := sit(t, run, "judge")
		mustAppend(t, bench, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("as proposed")})
		mustAppend(t, bench, benchFor("D1", "M1", forFiler))
		if g := gapRowOf(t, run, "G1"); !g.open || g.severity != "low" {
			t.Errorf("after the bench's ruling for the filer of the regrade: %+v, want G1 open at severity low", g)
		}
		if got := motionReadersAgree(t, run)["M1"]; got != (motionAnswer{ruling: "for author", by: "judge"}) {
			t.Errorf("M1 reads %+v, want for author by judge", got)
		}
	})
	t.Run("for the assignee on a regrade: the gap as it stood", func(t *testing.T) {
		run := motionGuardRun(t)
		before := gapRowOf(t, run, "G1")
		mustAppend(t, sit(t, run, "judge"), benchFor("D1", "M1", forAssignee))
		if after := gapRowOf(t, run, "G1"); after != before {
			t.Errorf("the bench's ruling for the assignee moved G1: %+v → %+v", before, after)
		}
		if got := motionReadersAgree(t, run)["M1"]; got != (motionAnswer{ruling: "for raiser", by: "judge"}) {
			t.Errorf("M1 reads %+v, want for raiser by judge", got)
		}
	})
}

// rulingEventOf is the event id of the ruling on motion.
func rulingEventOf(t *testing.T, run Run, motion string) int64 {
	t.Helper()
	var id int64
	if found, err := queryRowAt(recordOf(t, run), []any{&id}, `SELECT "event_id" FROM "motion_rule" WHERE "motion_id" = ?`, motion); err != nil || !found {
		t.Fatalf("the ruling on %s: found %v, %v", motion, found, err)
	}
	return id
}

// closureReasonOf is the replay's closure word for g.
func closureReasonOf(t *testing.T, run Run, g string) string {
	t.Helper()
	b, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	gap := b.Gap(g)
	if gap == nil {
		t.Fatalf("no gap %s on the board", g)
	}
	return gap.ClosureReason()
}

// THE STREAM'S MINTER IS THE STORED ONE (mintersOf against the gap view's minted_by): each gap reads
// the seat whose mint created it, on both readers.
func TestTheStreamsMinterIsTheStoredOne(t *testing.T) {
	run := proposalRun(t, closeProposalOn(recordpb.Disposition_DISPOSITION_NOT_A_DEFECT))
	other := sit(t, run, outsideLens)
	mustAppend(t, other, corrMint("G2"))

	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	minters := mintersOf(all.Events)
	for _, g := range []string{"G1", "G2"} {
		if stored := gapRowOf(t, run, g).mintedBy; minters[g] != stored {
			t.Errorf("%s: the stream's minter %q, the stored %q", g, minters[g], stored)
		}
	}
	want := map[string]string{"G1": evLens, "G2": outsideLens}
	for g, w := range want {
		if minters[g] != w {
			t.Errorf("%s's minter = %q, want %q", g, minters[g], w)
		}
	}
}

// gradeMovement.restoring as a unit: the first `from` against the last `to`, per axis.
func TestRestoringIsTheFirstFromAgainstTheLastTo(t *testing.T) {
	m := gradeMovement{moves: map[string][]gradeMove{
		"one":   {{axis: "severity", from: "high", to: "low"}},
		"round": {{axis: "severity", from: "high", to: "low"}, {axis: "severity", from: "low", to: "high"}},
		"mixed": {{axis: "severity", from: "high", to: "medium"}, {axis: "impact", from: "low", to: "high"}, {axis: "impact", from: "high", to: "low"}},
	}}
	if r := m.restoring("one", "b"); r == nil || r.GetSeverity() != recordpb.Grade_GRADE_HIGH {
		t.Errorf("one = %v, want severity high", r)
	}
	if r := m.restoring("round", "b"); r != nil {
		t.Errorf("round = %v, want nil", r)
	}
	if r := m.restoring("mixed", "b"); r == nil || r.GetSeverity() != recordpb.Grade_GRADE_HIGH || r.Impact != nil {
		t.Errorf("mixed = %v, want severity high alone", r)
	}
	if r := m.restoring("none", "b"); r != nil {
		t.Errorf("a gap with no movement = %v, want nil", r)
	}
}

// T-B9 (1), the convening row: the bench's docket sitting holds the docketings standing on the gaps
// the dispatch row that convened it names (D8). One on a gap the row does not name is not before the
// sitting — its ruling is refused there — while one on a gap it names is ruled in it.
func TestTheBenchRulesWhatItsSittingHolds(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).
		mint(evLens, "G1", "high").mint(evLens, "G2", "high").seed()
	for _, d := range [][2]string{{"D1", "G1"}, {"D2", "G2"}} {
		if _, err := docketByDispatchAs(t, run, d[0], d[1], recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, ""); err != nil {
			t.Fatal(err)
		}
	}
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(benchSeat), GapIds: []string{"G1"}})
	bench := sit(t, run, benchSeat)
	_, err := Append(bench, docketRule("D2", "on a gap the row does not name"))
	mustRefuse(t, err, "docketing D2 is not before this sitting")
	mustAppend(t, bench, docketRule("D1", "on the gap the row names"))
}
