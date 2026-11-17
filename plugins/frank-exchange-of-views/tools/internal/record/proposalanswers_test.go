package record

import (
	"database/sql"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A DOCKET MOTION IS A PROPOSED ACT, AND ITS ANSWER FITS IT. These tests drive the answers through
// record.Append — the write every writer shares — on motionGuardRun's record: M1 proposes G1's
// regrade, M2 (the dispatch's impasse filing) proposes no act, and the bench is convened on both.

// gapRow is the gap view's state columns for one gap, as one comparable value.
type gapRow struct {
	open, byBench, benchLast bool
	benchClose, supersede    sql.NullInt64
	assignee, severity       string
}

func gapRowOf(t *testing.T, run Run, g string) gapRow {
	t.Helper()
	var r gapRow
	found, err := queryRowAt(recordOf(t, run), []any{&r.open, &r.byBench, &r.benchLast, &r.benchClose, &r.supersede, &r.assignee, &r.severity},
		`SELECT "open", COALESCE("closed_by_bench", 0), COALESCE("bench_closed_last", 0), "bench_close_event", "supersede_event",
		        COALESCE("assignee", ''), COALESCE("current_severity", '') FROM "gap" WHERE "gap_id" = ?`, g)
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

// T-B6, T-B8 (no-act and closing-word rows), and the ruler check: each refusal leaves its motion
// standing, and the Go fold and the SQL view agree on it.
func TestADocketAnswerFitsTheMotionItAnswers(t *testing.T) {
	judgeAccept := func(id string, accepted bool) *recordpb.MotionRule { return actRule(id, accepted, "judge") }
	closingWord := func(id string) *recordpb.MotionRule {
		r := docketRule(id, "the bench closes it")
		r.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum()
		r.GetDocket().Direction = nil
		r.GetDocket().Final = proto.Bool(true)
		return r
	}
	for _, tc := range []struct {
		name   string
		seat   string
		body   *recordpb.MotionRule
		wants  []string
		motion string
	}{
		{"the assignee's accept of a no-act motion", evLens, actRule("M2", true, evLens),
			[]string{"motion M2 proposes no act", "remanded"}, "M2"},
		{"the bench's accept of a no-act motion", "judge", judgeAccept("M2", true),
			[]string{"motion M2 proposes no act", "not_a_defect", "remanded"}, "M2"},
		{"the bench's reject of a no-act motion", "judge", judgeAccept("M2", false),
			[]string{"motion M2 proposes no act", "not_a_defect", "remanded"}, "M2"},
		{"a closing word on a regrade proposal", "judge", closingWord("M1"),
			[]string{"motion M1 proposes a regrade of G1", "--accept or --reject"}, "M1"},
		{"the assignee's reject", evLens, actRule("M1", false, evLens),
			[]string{"answered by the bench, and by one other seat", "accepting the act it proposes"}, "M1"},
		{"another seat's accept", "blue-respond", actRule("M1", true, "blue-respond"),
			[]string{"answered by the bench, and by one other seat", "Gap G1 is " + evLens + "'s; you are blue-respond"}, "M1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run := motionGuardRun(t)
			_, err := Append(sit(t, run, tc.seat), tc.body)
			mustRefuse(t, err, tc.wants...)
			if got := motionReadersAgree(t, run)[tc.motion]; got.ruling != "" {
				t.Errorf("motion %s reads ruled %q after a refused answer; want it standing", tc.motion, got.ruling)
			}
		})
	}
}

// T-B7: A MOTION IS ANSWERED ONCE, whichever writer answers first.
func TestAProposalIsAnsweredOnce(t *testing.T) {
	t.Run("the bench's reject after the assignee's accept", func(t *testing.T) {
		run := motionGuardRun(t)
		mustAppend(t, sit(t, run, evLens), actRule("M1", true, evLens))
		_, err := Append(sit(t, run, "judge"), actRule("M1", false, "judge"))
		mustRefuse(t, err, `motion M1 is already ruled "accepted" by `+evLens)
		if got := motionReadersAgree(t, run)["M1"]; got.ruling != "accepted" {
			t.Errorf("M1 reads %q, want the assignee's accept", got.ruling)
		}
	})
	t.Run("the assignee's accept after the bench's reject", func(t *testing.T) {
		run := motionGuardRun(t)
		mustAppend(t, sit(t, run, "judge"), actRule("M1", false, "judge"))
		_, err := Append(sit(t, run, evLens), actRule("M1", true, evLens))
		mustRefuse(t, err, `motion M1 is already ruled "rejected" by judge`)
		if got := motionReadersAgree(t, run)["M1"]; got.ruling != "rejected" {
			t.Errorf("M1 reads %q, want the bench's reject", got.ruling)
		}
	})
}

// T-B15: ONLY THE BENCH OWES AN OPINION. A bench ruling missing any of its four opinion fields is
// refused — an accept, a reject and a remand alike; the assignee's accept lands with none and is
// refused carrying any.
func TestOnlyTheBenchOwesAnOpinion(t *testing.T) {
	fields := []string{"principle", "tension", "review_flag", "settled"}
	for _, ruling := range []struct {
		name string
		body *recordpb.MotionRule
	}{
		{"accept", actRule("M1", true, "judge")},
		{"reject", actRule("M1", false, "judge")},
		{"remand", docketRule("M1", "one more exchange")},
	} {
		for _, f := range fields {
			t.Run("the bench's "+ruling.name+" without "+f, func(t *testing.T) {
				run := motionGuardRun(t)
				_, err := Append(sit(t, run, "judge"), withoutOpinion(ruling.body, f))
				if err == nil {
					t.Fatalf("the bench's %s without its %s landed", ruling.name, f)
				}
			})
		}
		t.Run("the bench's "+ruling.name+" whole", func(t *testing.T) {
			run := motionGuardRun(t)
			mustAppend(t, sit(t, run, "judge"), ruling.body)
		})
	}
	t.Run("the assignee's accept with none", func(t *testing.T) {
		run := motionGuardRun(t)
		mustAppend(t, sit(t, run, evLens), actRule("M1", true, evLens))
	})
	for _, f := range fields {
		t.Run("the assignee's accept carrying "+f, func(t *testing.T) {
			run := motionGuardRun(t)
			r := actRule("M1", true, evLens)
			d := r.GetDocket()
			switch f {
			case "principle":
				d.Principle = proto.String("p")
			case "tension":
				d.Tension = proto.String("t")
			case "review_flag":
				d.ReviewFlag = proto.String("r")
			case "settled":
				d.Settled = proto.String("s")
			}
			_, err := Append(sit(t, run, evLens), r)
			mustRefuse(t, err, "is the bench's opinion on a ruling it writes")
		})
	}
}

// proposalRun is a cast of two lenses, the chair, blue and the bench, with G1 minted by the evidence
// lens and the bench convened on it after the motion filed by fileMotion.
func proposalRun(t *testing.T, act func(gap string) *recordpb.Motion) Run {
	t.Helper()
	run := newStage(t).cast(evLens, outsideLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).mint(evLens, "G1", "high").seed()
	mustFileByDispatch(t, run, act("G1"))
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

func supersedeProposalOn(successor string) func(string) *recordpb.Motion {
	return func(g string) *recordpb.Motion {
		return &recordpb.Motion{MotionId: proto.String("M1"), Basis: proto.String("b"),
			Filing: &recordpb.Motion_Docket{Docket: &recordpb.DocketMotion{GapId: proto.String(g)}},
			Act:    &recordpb.Motion_Supersede{Supersede: successorMint(successor, g)}}
	}
}

// T-B8: THE BENCH'S ANSWER PER ACT.
func TestTheBenchAnswersEachActItsOwnWay(t *testing.T) {
	t.Run("close: the closing ruling the proposal stores", func(t *testing.T) {
		run := proposalRun(t, closeProposalOn(recordpb.Disposition_DISPOSITION_MOOT))
		bench := convene(t, run, "G1")
		// The bench's accept is the stored disposition; `accepted: true` is not how it lands.
		_, err := Append(bench, actRule("M1", true, "judge"))
		mustRefuse(t, err, "the closing ruling it proposes — --as moot")
		// Another closing word than the stored one answers no proposal.
		other := docketRule("M1", "the bench closes it")
		other.GetDocket().Disposition, other.GetDocket().Direction, other.GetDocket().Final =
			recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(), nil, proto.Bool(true)
		_, err = Append(bench, other)
		mustRefuse(t, err, "motion M1 proposes a close of G1", "--accept or --reject")
		stored := proto.Clone(other).(*recordpb.MotionRule)
		stored.GetDocket().Disposition = recordpb.Disposition_DISPOSITION_MOOT.Enum()
		mustAppend(t, bench, stored)
		g := gapRowOf(t, run, "G1")
		if id := rulingEventOf(t, run, "M1"); g.open || !g.byBench || !g.benchClose.Valid || g.benchClose.Int64 != id {
			t.Errorf("after the bench's moot: %+v, want closed by the bench at event %d", g, id)
		}
		if w := closureReasonOf(t, run, "G1"); w != "moot" {
			t.Errorf("G1's closure word = %q, want the stored moot", w)
		}
	})
	t.Run("regrade: the bench's regrade, then its accept", func(t *testing.T) {
		run := motionGuardRun(t)
		bench := sit(t, run, "judge")
		mustAppend(t, bench, &recordpb.Regrade{GapId: proto.String("G1"), Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Basis: proto.String("as proposed")})
		mustAppend(t, bench, actRule("M1", true, "judge"))
		if g := gapRowOf(t, run, "G1"); !g.open || g.severity != "low" {
			t.Errorf("after the bench's accept of the regrade: %+v, want G1 open at severity low", g)
		}
		if got := motionReadersAgree(t, run)["M1"]; got.ruling != "accepted" || got.by != "judge" {
			t.Errorf("M1 reads %+v, want accepted by judge", got)
		}
	})
	t.Run("supersede: the bench's close of G, and the successor its assignee's", func(t *testing.T) {
		run := proposalRun(t, supersedeProposalOn("G9"))
		bench := convene(t, run, "G1")
		mustAppend(t, bench, actRule("M1", true, "judge"))
		g := gapRowOf(t, run, "G1")
		if id := rulingEventOf(t, run, "M1"); g.open || !g.byBench || !g.benchLast || g.benchClose.Int64 != id || g.supersede.Valid {
			t.Errorf("after the bench's accept of the supersede: %+v, want G1 closed by the bench at the ruling %d, no supersede event", g, id)
		}
		if w := closureReasonOf(t, run, "G1"); w != ClosedSuperseded {
			t.Errorf("G1's closure word = %q, want %q", w, ClosedSuperseded)
		}
		// The bench mints nothing but the successor it accepted.
		_, err := Append(bench, successorMint("G8", "G1"))
		mustRefuse(t, err, "the bench mints no gap of its own")
		mustAppend(t, bench, successorMint("G9", "G1"))
		if s := gapRowOf(t, run, "G9"); !s.open || s.assignee != evLens {
			t.Errorf("the bench's successor: %+v, want open and %s's", s, evLens)
		}
	})
	t.Run("supersede: the assignee's accept is no bench closure", func(t *testing.T) {
		run := proposalRun(t, supersedeProposalOn("G9"))
		lens := sit(t, run, evLens)
		mustAppend(t, lens, actRule("M1", true, evLens))
		mustAppend(t, lens, successorMint("G9", "G1"))
		if g := gapRowOf(t, run, "G1"); g.open || g.byBench {
			t.Errorf("after the assignee's accept and successor: %+v, want G1 closed, not by the bench", g)
		}
	})
	t.Run("reject: the gap as it stood", func(t *testing.T) {
		run := motionGuardRun(t)
		before := gapRowOf(t, run, "G1")
		mustAppend(t, sit(t, run, "judge"), actRule("M1", false, "judge"))
		if after := gapRowOf(t, run, "G1"); after != before {
			t.Errorf("the bench's reject moved G1: %+v → %+v", before, after)
		}
		if got := motionReadersAgree(t, run)["M1"]; got.ruling != "rejected" {
			t.Errorf("M1 reads %q, want rejected", got.ruling)
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

// T-B10, subset: an accept and a reject are answers every reader reads as answers, and the bench-only
// readers count the bench's answer and never the assignee's.
func TestAcceptedAndRejectedRulingsReadAsAnswers(t *testing.T) {
	run := motionGuardRun(t)
	mustAppend(t, sit(t, run, evLens), actRule("M1", true, evLens))
	answers := motionReadersAgree(t, run)
	if answers["M1"].ruling != "accepted" {
		t.Errorf("M1 reads %+v, want accepted", answers["M1"])
	}
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range MotionsOf(all.Events, all.At) {
		if m.ID == "M1" && !m.Ruled() {
			t.Errorf("M1, accepted, reads unruled")
		}
	}
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bs {
		if b.Subject == "M1" {
			t.Errorf("the accepted M1 still holds the PASS: %+v", b)
		}
	}
	if w := RulingWord(rulingOf(t, run, "M1")); w != "accepted" {
		t.Errorf("RulingWord of the assignee's accept = %q, want accepted", w)
	}

	if lead := leadOf(t, run); len(lead) != 0 {
		t.Errorf("the debate's LEAD carries the assignee's accept: %+v", lead)
	}

	rejected := motionGuardRun(t)
	mustAppend(t, sit(t, rejected, "judge"), actRule("M1", false, "judge"))
	if w := RulingWord(rulingOf(t, rejected, "M1")); w != "rejected" {
		t.Errorf("RulingWord of the bench's reject = %q, want rejected", w)
	}
	if lead := leadOf(t, rejected); len(lead) != 1 || lead[0].Disposition != "rejected" || lead[0].GapID != "G1" {
		t.Errorf("the debate's LEAD = %+v, want the bench's reject of G1's proposal", lead)
	}
}

// leadOf is every bench opinion the debate projection carries, over all epochs.
func leadOf(t *testing.T, run Run) []DebateOpinionJSON {
	t.Helper()
	f, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	var out []DebateOpinionJSON
	for _, e := range DebateJSONOfFamily(f).Epochs {
		out = append(out, e.Lead...)
	}
	return out
}

// rulingOf is the MotionRule answering motion on run's record.
func rulingOf(t *testing.T, run Run, motion string) *recordpb.MotionRule {
	t.Helper()
	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range all.Events {
		if r, ok := recordpb.BodyAs[*recordpb.MotionRule](e); ok && r.GetMotionId() == motion {
			return r
		}
	}
	t.Fatalf("no ruling on %s", motion)
	return nil
}

// THE STREAM'S ASSIGNEE IS THE STORED ONE (mintChain against gap_state.assignee): a lineage successor
// another lens mints on a closed gap is its origin's, and the bench's successor is its ancestor's
// assignee's; the acting seat of the bench's Mint is that assignee too.
func TestAssigneeAndActingSeatReadersOfASuccessor(t *testing.T) {
	run := proposalRun(t, supersedeProposalOn("G9"))
	bench := convene(t, run, "G1")
	mustAppend(t, bench, actRule("M1", true, "judge"))
	mustAppend(t, bench, successorMint("G9", "G1"))
	// A lineage successor of the closed G9's sibling: G2, the outside lens's gap, closed by it, then
	// superseded by the evidence lens.
	other := sit(t, run, outsideLens)
	mustAppend(t, other, corrMint("G2"))
	mustAppend(t, other, &recordpb.Close{GapId: proto.String("G2"), ClosureClass: recordpb.Disposition_DISPOSITION_NOT_A_DEFECT.Enum(),
		Prose: proto.String("p"), AnchorSeat: proto.String(outsideLens), AnchorTool: proto.String("read"), AnchorTarget: proto.String("the report")})
	mustAppend(t, sit(t, run, evLens), successorMint("G3", "G2"))

	all, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	assignee, acting := mintChain(all.Events)
	for _, g := range []string{"G1", "G9", "G2", "G3"} {
		if stored := gapRowOf(t, run, g).assignee; assignee[g] != stored {
			t.Errorf("%s: the stream's assignee %q, the stored %q", g, assignee[g], stored)
		}
	}
	want := map[string]string{"G9": evLens, "G3": outsideLens}
	for g, w := range want {
		if assignee[g] != w {
			t.Errorf("%s's assignee = %q, want %q", g, assignee[g], w)
		}
	}
	if acting["G9"] != evLens || acting["G3"] != evLens {
		t.Errorf("acting seats G9 %q, G3 %q; want %s (the bench's successor acts as its assignee) and %s (the writer)", acting["G9"], acting["G3"], evLens, evLens)
	}
}

// T-B4: THE DISPATCH'S GRADE-MOVEMENT FILING PROPOSES THE RESTORING REGRADE. Driven through
// DispatchNext: a gap moved one way is filed with the regrade restoring its axis; a gap whose axis went
// down and back up carries movement and no net difference, and is filed with no act (L25); a gap with
// net movement on one axis and a round trip on another is restored on the one axis alone.
func TestTheGradeMovementFilingRestoresTheNetMovement(t *testing.T) {
	b := movedBoard(t, [2]string{"G1", "low"}, [2]string{"G2", "low"}, [2]string{"G2", "medium"})
	// G1's likelihood goes up and back down: a round trip beside its net severity move.
	b.add(outsideLens, &recordpb.Regrade{GapId: proto.String("G1"), Likelihood: recordtest.P(recordpb.Grade_GRADE_HIGH), Basis: proto.String("b")})
	b.add(outsideLens, &recordpb.Regrade{GapId: proto.String("G1"), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Basis: proto.String("b")})
	b.register("red-chair")
	run := b.seed()
	if _, _, err := DispatchNext(Identity{Run: run, SeatID: "red-chair"}); err != nil {
		t.Fatal(err)
	}
	evs, _, err := eventsAt(recordOf(t, run))
	if err != nil {
		t.Fatal(err)
	}
	acts := map[string]*recordpb.Motion{}
	for _, e := range evs {
		if m := e.GetMotion(); m != nil && m.GetDocket() != nil {
			acts[m.GetDocket().GetGapId()] = m
			if e.GetSeatId() != "red-chair" {
				t.Errorf("the filing on %s is %s's, want the chair's", m.GetDocket().GetGapId(), e.GetSeatId())
			}
		}
	}
	if len(acts) != 2 {
		t.Fatalf("filings on %d gaps (%v), want G1 and G2", len(acts), acts)
	}
	r := acts["G1"].GetRegrade()
	if r == nil || r.GetSeverity() != recordpb.Grade_GRADE_HIGH || r.Likelihood != nil || r.Impact != nil {
		t.Errorf("G1's filing proposes %v, want severity restored to high and no other axis", r)
	}
	if a := acts["G2"].GetAct(); a != nil {
		t.Errorf("G2's filing proposes %v, want no act: its axis went down and back", a)
	}
	// The next plan plans.
	if _, err := PlanDispatch(run); err != nil {
		t.Errorf("the plan after the filings: %v", err)
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

// T-B9 (1), the convening row: the bench's docket sitting holds the motions filed before the dispatch
// that convened it. One filed after that row is not before the sitting — its ruling is refused there —
// while one filed before it is ruled in it.
func TestTheBenchRulesWhatItsSittingHolds(t *testing.T) {
	run := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").sit(2, evLens).
		mint(evLens, "G1", "high").mint(evLens, "G2", "high").seed()
	mustFileByDispatch(t, run, regradeMotion("M1", "G1", "filed before the bench was convened"))
	mustAppend(t, Identity{Run: run, SeatID: chairSeat}, &recordpb.Dispatch{Pin: proto.Int64(1), SeatId: proto.String(benchSeat), GapIds: []string{"G1", "G2"}})
	mustFileByDispatch(t, run, regradeMotion("M2", "G2", "filed after the bench was convened"))
	bench := sit(t, run, benchSeat)
	_, err := Append(bench, actRule("M2", false, "judge"))
	mustRefuse(t, err, "motion M2 is not before this sitting")
	mustAppend(t, bench, actRule("M1", false, "judge"))
}
