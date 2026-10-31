package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// m37Archive is universe-m37: 1263 events, CEILING at epoch 12 under k-max 2, six dockets filed by the
// dispatch and six rulings on them. It holds the bundle store (computations/<bundle>/), which the
// migrate package's archive census does not read as the proof store of every run in run-archive/, so
// it stands in a directory of its own.
const m37Archive = "bundle-store/2026-10-31_is-91-prime-m37.tar.gz"

// UNIVERSE-M37, REPLAYED (ruled by gblock 2026-10-31). Each of the run's six dockets is folded again on
// the record as it stood before the plan that filed it. On all six the plan now files nothing and
// readies the gap's minting lens alone. What that lens did next is the run's own record: on five it
// closed the gap, each within minutes, so no docket exists and the bench has nothing to rule. On
// G-deced912's first docket it kept its grade, and that docket and its remand stand a sitting later:
// the one ruling of the six that decided anything.
//
//   - G-8405d5e1: its lens promised the close when its successor closed; stranded, its lens goes first.
//   - G-584d152f: blue's fix landed under G-07884fea's name; the edit's `reopened` gives the lens its turn.
//   - G-c3f84f98: blue's computation answered the kept grade; the lens reads it before any docket.
//   - G-deced912, first: the lens audits blue's answer, and keeps. The dispute is real.
//   - G-deced912, second: blue obeyed the remand with a computation, and no wording changed. The re-audit
//     the ruling ordered happens; the gap is not re-docketed for blue's obedience.
//   - G-37347c3c: blue's fix was a fate on the inquiry it is anchored on.
func TestM37DocketsComeAfterTheLensHasRead(t *testing.T) {
	whole := archivedEvents(t, m37Archive)
	want := []struct {
		gap     string
		remains bool
	}{{"G-8405d5e1", false}, {"G-584d152f", false}, {"G-c3f84f98", false}, {"G-deced912", true}, {"G-deced912", false}, {"G-37347c3c", false}}
	minted := mintedBy(whole)
	k, remain := 0, 0
	for i, e := range whole {
		m, ok := recordpb.BodyAs[*recordpb.Motion](e)
		if !ok || m.GetDocket().GetGapId() == "" {
			continue
		}
		if k == len(want) {
			t.Fatalf("a seventh docket motion on the record: %s", m.GetMotionId())
		}
		g := m.GetDocket().GetGapId()
		if g != want[k].gap {
			t.Fatalf("docket %d is on %s, want %s", k, g, want[k].gap)
		}
		start := i
		for start > 0 {
			if p, ok := recordpb.BodyAs[*recordpb.Motion](whole[start-1]); !ok || p.GetDocket().GetGapId() == "" {
				break
			}
			start--
		}
		plan, err := PlanDispatch(archiveAt(t, m37Archive, whole, start))
		if err != nil {
			t.Fatal(err)
		}
		lens := minted[g]
		if slices.Contains(plan.ToFile, g) || !slices.Contains(gapsOf(plan, lens), g) || slices.Contains(gapsOf(plan, blueRespondSeat), g) {
			t.Errorf("docket %s on %s: the plan must file nothing and ready %s alone on it: to file %v, parties %+v", m.GetMotionId(), g, lens, plan.ToFile, plan.Parties)
		}
		// WHAT THE LENS DID NEXT, off the run itself.
		closed := false
		for _, f := range whole[i:] {
			if f.GetSeatId() != lens {
				continue
			}
			if c := f.GetClose(); c != nil && c.GetGapId() == g {
				closed = true
				break
			}
			if r := f.GetRegrade(); r != nil && r.GetGapId() == g {
				break
			}
		}
		if closed == want[k].remains {
			t.Errorf("docket %s on %s: the lens closed it next = %v, want %v", m.GetMotionId(), g, closed, !want[k].remains)
		}
		if !closed {
			remain++
		}
		k++
	}
	if k != len(want) || remain != 1 {
		t.Errorf("%d docket(s), %d ruling(s) remaining, want %d and 1", k, remain, len(want))
	}
	// CAPTURE NAMES THEM.
	early, err := EarlyDockets(archiveAt(t, m37Archive, whole, len(whole)))
	if err != nil {
		t.Fatal(err)
	}
	if len(early) != len(want) {
		t.Errorf("m37's record holds %d early docket(s), want %d: %+v", len(early), len(want), early)
	}
}

// atLastAudit is the lens minting G1 under the terms kMax states, blue's first answer, the lens's keep
// and blue's second answer: under k-max 2 the next plan is the last audit, which reaches the limit.
func atLastAudit(t *testing.T, kMax string) (*stage, Run) {
	t.Helper()
	st := mintedOn(t).blueSitsOn("G1", true)
	st.register("red-chair").dispatch(2, evLens, "G1").dispatch(2, "blue-respond", "G1")
	st.registerAs(evLens, "lens-keeps").served(evLens, st.head()).keepGrade(evLens, "G1", keptReason).stop("lens-keeps")
	st.registerAs("blue-respond", "blue-again").edit("G1", "new", "newer").stop("blue-again")
	run := st.register("red-chair").seed()
	writeRunConfig(t, run.Dir(), kMax)
	return st, run
}

// THE ROUND ENDS ON THE LENS'S RE-AUDIT (ruled by gblock 2026-10-31). Under k-max 2 blue answers twice
// and the lens audits twice, and the docket comes after the second audit. The plan that opens the last
// audit readies the lens alone and files nothing: blue sits after the lens, and an answer of its there
// would be one nobody had read when the bench was asked.
func TestTheLastAuditReadiesTheLensAloneAndTheDocketWaits(t *testing.T) {
	st, run := atLastAudit(t, `{"kMax": 2}`)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.ToFile) != 0 || !slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("the second answer's audit reaches k-max: the lens alone is ready, and nothing is filed: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
	}
	if !whyMentions(plan, "G1: 1 exchange(s) (0 stalled) — this audit by "+evLens+" comes before any docket") {
		t.Errorf("the reason does not say the audit reaches the limit: %q", plan.Why)
	}
	// THE LENS IS SATISFIED: it closes, and no docket ever exists.
	closes := st.more()
	closes.dispatch(2, evLens, "G1").registerAs(evLens, "lens-closes").served(evLens, 2).closeGap(evLens, "G1").stop("lens-closes").register("red-chair").seed()
	after, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Docket) != 0 || slices.Contains(gapsOf(after, benchSeat), "G1") {
		t.Errorf("a satisfied lens closed before any docket: docket %v, parties %+v", after.Docket, after.Parties)
	}
}

// A REAL DISPUTE STILL REACHES THE BENCH WITHIN K-MAX. The lens keeps at the last audit, and the next
// plan files, with two answers and two audits had.
func TestARealDisputeReachesTheBenchWithinKMax(t *testing.T) {
	st, run := atLastAudit(t, `{"kMax": 2}`)
	kept := st.more()
	kept.dispatch(2, evLens, "G1").registerAs(evLens, "lens-keeps-again").served(evLens, 2).keepGrade(evLens, "G1", keptReason).stop("lens-keeps-again").register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(plan.ToFile, []string{"G1"}) {
		t.Errorf("two audits under k-max 2, and the lens kept: the gap is docketed: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
	}
	x, err := Exchanges(run, Params{K: 2, KMax: 2})
	if err != nil || x["G1"].Exchanges != 2 {
		t.Errorf("two exchanges had: %+v (%v)", x["G1"], err)
	}
}

// THE TRAP. Blue moves the gap's sentence under another gap's name every epoch, which gives the lens its
// turn every epoch, and blue sits last. A deferral resting on the turn alone would never file: the lens
// has one audit past the limit, and the docket then comes, turn or no turn.
func TestAnEditEveryEpochDoesNotHoldTheDocketOff(t *testing.T) {
	st, run := atLastAudit(t, `{"kMax": 2}`)
	for n := range 3 {
		more := st.more()
		lensAgent, blueAgent := "lens-"+string(rune('a'+n)), "blue-"+string(rune('a'+n))
		more.dispatch(2, evLens, "G1").registerAs(evLens, lensAgent).served(evLens, 2).keepGrade(evLens, "G1", keptReason).stop(lensAgent).
			registerAs("blue-respond", blueAgent).add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G2"), Old: proto.String("x"), New: proto.String("y"), EditKey: proto.String("sib" + lensAgent), Reopened: []string{"G1"}}).stop(blueAgent).
			register("red-chair").seed()
		plan, err := PlanDispatch(run)
		if err != nil {
			t.Fatal(err)
		}
		switch n {
		case 0:
			if len(plan.ToFile) != 0 || !slices.Contains(gapsOf(plan, evLens), "G1") {
				t.Fatalf("an edit moved G1's sentence after its last audit, so the lens has one more: to file %v, parties %+v", plan.ToFile, plan.Parties)
			}
		case 1:
			if !slices.Equal(plan.ToFile, []string{"G1"}) {
				t.Fatalf("the lens has had its audit past the limit, and the docket comes however the sentence moved: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
			}
			return
		}
		st = more
	}
}

// AN EDIT TO A GAP'S SENTENCE GIVES THE LENS ITS TURN, AS A TURN AND NOT AN ANSWER (ruled by gblock
// 2026-10-31; G-584d152f). The act fold still reads no answer on G1, the remand's exchange would not
// begin on it, and the list says which it is.
func TestAnEditUnderAnotherGapsNameGivesTheLensItsTurn(t *testing.T) {
	st := mintedOn(t)
	st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-sibling").
		add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G2"), Old: proto.String("x"), New: proto.String("y"), EditKey: proto.String("E-sib"), Reopened: []string{"G1"}}).
		stop("blue-sibling")
	run := st.register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G1") {
		t.Errorf("an edit changed G1's sentence, and its lens is not readied: parties %+v, why %q", plan.Parties, plan.Why)
	}
	if !workMentions(t, run, "lens", evLens, "gap G1: an edit changed its sentence since you last acted on it, and no act of blue's answers it") {
		t.Errorf("the lens's list does not say an edit moved the sentence")
	}
	if workMentions(t, run, "lens", evLens, "gap G1 has been ANSWERED") {
		t.Errorf("the edit is a turn and not an answer, and the list calls it an answer")
	}
	evs, win := mustEvents(t, run)
	if blueActed(gapActsOf(evs, win.IDs(evs))["G1"], 0, 1<<62) {
		t.Errorf("the edit answered another gap, and reads as blue's act on G1")
	}
}

// A COMPLIANT ANSWER THAT IS NOT AN EDIT IS MOVEMENT (ruled by gblock 2026-10-31). Blue's computation
// answering the gap, and a change of fate on the inquiry it is anchored on, each move it; a fate restated
// does not.
func TestAComputationOrAFateIsMovement(t *testing.T) {
	t.Run("a computation", func(t *testing.T) {
		st := mintedOn(t)
		st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-c").blueComputes("G1").stop("blue-c")
		run := st.reauditOn("G1").register("red-chair").seed()
		x, err := Exchanges(run, DefaultParams)
		if err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 0 {
			t.Errorf("blue's computation moved G1: %+v (%v)", x["G1"], err)
		}
	})
	fate := func(q string, s recordpb.InquiryStatus) *recordpb.Inquiry {
		return &recordpb.Inquiry{InquiryId: proto.String(q), Status: s.Enum(), Reason: proto.String("r")}
	}
	inquiryStage := func(t *testing.T, status recordpb.InquiryStatus) *stage {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
		st.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY.Enum()})
		st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).add(evLens, inquiryMint("G1", "Q-1")).register("red-chair")
		st.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-f").add("blue-respond", fate("Q-1", status)).stop("blue-f")
		return st.reauditOn("G1").register("red-chair")
	}
	t.Run("a fate", func(t *testing.T) {
		x, err := Exchanges(inquiryStage(t, recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED).seed(), DefaultParams)
		if err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 0 {
			t.Errorf("a fate on its inquiry moved G1: %+v (%v)", x["G1"], err)
		}
	})
	t.Run("a fate restated", func(t *testing.T) {
		x, err := Exchanges(inquiryStage(t, recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY).seed(), DefaultParams)
		if err != nil || x["G1"].Exchanges != 1 || x["G1"].Stalled != 1 {
			t.Errorf("a fate restated moves nothing: %+v (%v)", x["G1"], err)
		}
	})
}

// A REMAND THAT ORDERED NO WORDING CHANGE (G-deced912): blue complied with a computation. The lens is
// readied alone to re-audit it, as the ruling orders, and the gap is not re-docketed for blue's
// obedience, before that audit or after it.
func TestACompliantRemandAnswerIsNotRedocketed(t *testing.T) {
	st := remandedAtImpasse(t)
	st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-complies").blueComputes("G1").stop("blue-complies")
	run := st.register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.ToFile) != 0 || !slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("blue complied: the lens alone re-audits, and nothing is filed: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
	}
	audited := st.more()
	audited.dispatch(2, evLens, "G1").registerAs(evLens, "lens-re-audits").served(evLens, 2).stop("lens-re-audits").register("red-chair").seed()
	after, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.ToFile) != 0 || !slices.Contains(gapsOf(after, "blue-respond"), "G1") {
		t.Errorf("the remand's exchange moved G1, so it is below its limits: to file %v, parties %+v, why %q", after.ToFile, after.Parties, after.Why)
	}
}

// A STRANDED GAP KEEPS ITS CLOSE (G-8405d5e1). At its limits with nothing from blue to audit, its lens
// has one sitting before the docket, for the close its successor promised, and the close is admitted. If
// it does not close, the docket comes: a gap whose lens never acts still reaches the bench.
func TestAStrandedGapAtItsLimitsKeepsItsClose(t *testing.T) {
	build := func(t *testing.T) (*stage, Run) {
		st := mintedOn(t)
		st.add(evLens, successorMint("G2", "G1"))
		st.blueSitsOn("G1", false)
		st.blueSitsOn("G1", false)
		run := st.register("red-chair").seed()
		writeRunConfig(t, run.Dir(), `{"kMax": 2}`)
		return st, run
	}
	st, run := build(t)
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.ToFile) != 0 || !slices.Contains(gapsOf(plan, evLens), "G1") || slices.Contains(gapsOf(plan, "blue-respond"), "G1") {
		t.Fatalf("a stranded gap at its limits readies its lens alone, and files nothing: to file %v, parties %+v, why %q", plan.ToFile, plan.Parties, plan.Why)
	}
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_CLOSE, repairClose("G1")); err != nil {
		t.Errorf("the stranded gap's close is refused: %v", err)
	}
	sat := st.more()
	sat.dispatch(2, evLens, "G1").registerAs(evLens, "lens-silent").served(evLens, 2).stop("lens-silent").register("red-chair").seed()
	after, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after.ToFile, []string{"G1"}) {
		t.Errorf("its lens sat and did not close: the gap goes to the bench: to file %v, parties %+v, why %q", after.ToFile, after.Parties, after.Why)
	}
}

// ONCE DOCKETED, A GAP LEAVES THE PARTIES' LISTS BUT FOR ITS CLOSING, AND THE LENS'S CLOSE IS REFUSED
// (ruled by gblock 2026-10-31). Blue's answer and another lens's finding are both on the record, and
// neither reaches a list. The row rides `closings_owed` while the closing is owed and goes with it. The
// bench's list still holds the gap.
func TestADocketedGapLeavesTheListsButForItsClosing(t *testing.T) {
	st := mintedOn(t).blueSitsOn("G1", true)
	st.add("red-lens-logic", &recordpb.Finding{Id: proto.String("F-1"), Text: proto.String("the gap reads wrong"), AboutKind: recordpb.AboutKind_ABOUT_KIND_GAP.Enum(), AboutRef: proto.String("G1")})
	// A computation demand is an item the sitting's own arm writes, on every gap it reads.
	demand := highMint("G2")
	demand.CheckKind = recordpb.CheckKind_CHECK_KIND_COMPUTATION.Enum()
	st.add(evLens, demand)
	st.docketMotion("red-chair", "M1", "G1").docketMotion("red-chair", "M2", "G2")
	run := st.seed()

	err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_CLOSE, repairClose("G1"))
	if err == nil || !strings.Contains(err.Error(), "motion M1 stands unruled") || !strings.Contains(err.Error(), "closing argument") {
		t.Errorf("the close of a docketed gap: %v", err)
	}
	if feov.CodeOf(err) != string(feov.Conflict) {
		t.Errorf("the close of a docketed gap is refused as %v, want a conflict", feov.CodeOf(err))
	}

	for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
		w, err := WorkOfSeat(run, seat.role, seat.id)
		if err != nil {
			t.Fatal(err)
		}
		if len(w.Open) != 0 || w.Counts.Open != 0 {
			t.Errorf("%s: `open` holds the docketed gap: %+v", seat.id, w.Open)
		}
		if len(w.ClosingsOwed) != 2 || w.ClosingsOwed[0].ID != "G1" || w.ClosingsOwed[0].YoursToClose {
			t.Errorf("%s: closings_owed = %+v, want G1's and G2's rows, not its to close", seat.id, w.ClosingsOwed)
		}
		for _, gap := range []string{"G1", "G2"} {
			named := 0
			for _, it := range w.Sitting.Open {
				if strings.Contains(it.What, gap) {
					named++
					if !strings.Contains(it.What, "closing") {
						t.Errorf("%s: an item names docketed gap %s, and is not its closing: %q", seat.id, gap, it.What)
					}
				}
			}
			if named != 1 {
				t.Errorf("%s: %d item(s) name %s, want its closing alone: %+v", seat.id, named, gap, w.Sitting.Open)
			}
		}
	}
	filed := st.more()
	filed.closing(evLens, "G1").closing(evLens, "G2").seed()
	w, err := WorkOfSeat(run, "lens", evLens)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.ClosingsOwed) != 0 || len(w.Open) != 0 {
		t.Errorf("its closing filed, the docketed gap is off the lens's list: closings_owed %+v, open %+v", w.ClosingsOwed, w.Open)
	}
	w, err = WorkOfSeat(run, "bench", benchSeat)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Open) != 2 || w.Open[0].ID != "G1" {
		t.Errorf("the bench's list does not hold the gaps it rules: %+v", w.Open)
	}
}

// ANSWERED CLEARS ONCE THE LENS HAS AUDITED THAT ANSWER (defect 6 of the m37 audit), on a gap that is not
// material as on any other: the lens's keep, off-dispatch, reads it, and the item names only what is new.
func TestAnsweredClearsOnceTheLensHasAuditedIt(t *testing.T) {
	const answered = "gap G1 has been ANSWERED"
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "low").register("red-chair")
	st.dispatch(2, "blue-respond").registerAs("blue-respond", "blue-a").edit("G1", "was", "is").stop("blue-a")
	run := st.seed()
	if !workMentions(t, run, "lens", evLens, answered) {
		t.Fatalf("blue answered a gap that is not material, and its lens's list does not say so")
	}
	kept := st.more()
	kept.registerAs(evLens, "lens-keeps").served(evLens, 2).keepGrade(evLens, "G1", keptReason).seed()
	if workMentions(t, run, "lens", evLens, answered) {
		t.Errorf("the lens kept its grade after blue's answer, and the list still says it was answered")
	}
	again := kept.more()
	again.dispatch(2, "blue-respond").registerAs("blue-respond", "blue-b").edit("G1", "is", "was").stop("blue-b").seed()
	if !workMentions(t, run, "lens", evLens, answered) {
		t.Errorf("blue answered again, and the list does not say so")
	}
}

// A COMPUTATION IS AN ANSWER ON THE LIST (defect 7): the item says which, and not that nothing answered.
func TestTheListNamesAComputationAsTheAnswer(t *testing.T) {
	st := mintedOn(t)
	st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-c").blueComputes("G1").stop("blue-c")
	run := st.register("red-chair").seed()
	if !workMentions(t, run, "lens", evLens, "gap G1 has been ANSWERED by a computation") {
		t.Errorf("blue's computation answered G1, and the list does not say so")
	}
	if workMentions(t, run, "lens", evLens, "NO edit") {
		t.Errorf("the list says no edit answers a gap blue's computation answered")
	}
}

// THE ROW AND THE PLAN ARE ONE PREDICATE: `your_turn` on each party's row is exactly the plan's
// engagement of it, at every stage of one gap's life.
func TestYourTurnIsThePlansEngagement(t *testing.T) {
	check := func(t *testing.T, name string, run Run) {
		t.Helper()
		plan, err := PlanDispatch(run)
		if err != nil {
			t.Fatal(err)
		}
		for _, seat := range []struct{ role, id string }{{"lens", evLens}, {"blue", "blue-respond"}} {
			w, err := WorkOfSeat(run, seat.role, seat.id)
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range w.Open {
				if want := slices.Contains(gapsOf(plan, seat.id), row.ID); row.YourTurn != want {
					t.Errorf("%s: %s's row on %s says your_turn %v, and the plan's engagement is %v: parties %+v", name, seat.id, row.ID, row.YourTurn, want, plan.Parties)
				}
			}
		}
	}
	check(t, "fresh", mintedOn(t).seed())
	check(t, "answered", mintedOn(t).blueSitsOn("G1", true).register("red-chair").seed())
	_, last := atLastAudit(t, `{"kMax": 2}`)
	check(t, "the last audit", last)
	check(t, "remanded", remandedAtImpasse(t).register("red-chair").seed())
	st := remandedAtImpasse(t)
	st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-r").blueComputes("G1").stop("blue-r")
	check(t, "remand answered", st.register("red-chair").seed())
}

// successorMint is a lens's mint naming ancestor as the gap it supersedes.
func successorMint(gap, ancestor string) *recordpb.Mint {
	m := highMint(gap)
	m.Supersedes = []string{ancestor}
	return m
}

// AN ANSWER THE LENS HAS READ IS NO NEWS (defect 6): once it has audited blue's answer, a later edit of
// the sentence under another gap's name is that edit, and the list does not call the old answer new.
func TestAnAnswerTheLensHasReadIsNotNamedAgain(t *testing.T) {
	sibling := func(st *stage) Run {
		st.register("red-chair").dispatch(2, "blue-respond", "G2").registerAs("blue-respond", "blue-sibling").
			add("blue-respond", &recordpb.BlueEdit{Answers: proto.String("G2"), Old: proto.String("x"), New: proto.String("y"), EditKey: proto.String("E-sib"), Reopened: []string{"G1"}}).
			stop("blue-sibling")
		return st.register("red-chair").seed()
	}
	audited := func(st *stage, name string) *stage {
		return st.registerAs(evLens, name).served(evLens, st.head()).keepGrade(evLens, "G1", keptReason).stop(name)
	}
	const turn = "gap G1: an edit changed its sentence since you last acted on it"
	cases := []struct {
		name, answer string
		run          func(t *testing.T) Run
	}{
		{"an edit", "gap G1 has been ANSWERED by a change", func(t *testing.T) Run {
			st := mintedOn(t).blueSitsOn("G1", true)
			return sibling(audited(st.register("red-chair").dispatch(2, evLens, "G1"), "lens-e"))
		}},
		{"a computation", "gap G1 has been ANSWERED by a computation", func(t *testing.T) Run {
			st := mintedOn(t)
			st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-c").blueComputes("G1").stop("blue-c")
			return sibling(audited(st.register("red-chair").dispatch(2, evLens, "G1"), "lens-c"))
		}},
		{"a fate", "gap G1 has been ANSWERED by an inquiry record", func(t *testing.T) Run {
			st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
			st.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY.Enum()})
			st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).add(evLens, inquiryMint("G1", "Q-1")).register("red-chair")
			st.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-f").
				add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_ABANDONED.Enum(), Reason: proto.String("r")}).stop("blue-f")
			return sibling(audited(st.register("red-chair").dispatch(2, evLens, "G1"), "lens-f"))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			run := c.run(t)
			if workMentions(t, run, "lens", evLens, c.answer) {
				t.Errorf("the lens audited %s, and the list names it as new", c.name)
			}
			if !workMentions(t, run, "lens", evLens, turn) {
				t.Errorf("the sentence moved since the lens's audit of %s, and the list does not say so", c.name)
			}
		})
	}
}

// SILENCE IS A TURN TAKEN (ruled by gblock 2026-10-31): a lens engaged on the gap that sits and acts on
// nothing has had its turn, and the next plan does not ready it for the same answer. The sitting ended
// the exchange.
func TestASilentSittingTakesTheLensTurn(t *testing.T) {
	st := mintedOn(t).blueSitsOn("G1", true)
	st.register("red-chair").dispatch(2, evLens, "G1").registerAs(evLens, "lens-silent").served(evLens, st.head()).stop("lens-silent")
	run := st.register("red-chair").seed()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(gapsOf(plan, evLens), "G1") {
		t.Errorf("the lens sat through blue's answer in silence, and is readied on it again: parties %+v", plan.Parties)
	}
	x, err := Exchanges(run, DefaultParams)
	if err != nil || x["G1"].Exchanges != 1 {
		t.Errorf("the silent sitting ended the exchange: %+v (%v)", x["G1"], err)
	}
}

// THE LENS GOES BEFORE ANY DOCKET (ruled by gblock 2026-10-31), the one the open gaps' grade movement
// asks for too; an audit already had at impasse does not hold it off again.
func TestGradeMovementWaitsOnTheLensToo(t *testing.T) {
	for _, c := range []struct {
		name string
		x    GapExchanges
		want gapRoute
	}{
		{"its audit owed", GapExchanges{lensTurn: true, lensAudit: true}, routeLensFirst},
		{"its audit spent", GapExchanges{lensTurn: true}, routeGradeMovement},
		{"no turn", GapExchanges{}, routeGradeMovement},
	} {
		if got := routeOf(true, false, false, true, &c.x); got != c.want {
			t.Errorf("%s: route %v, want %v", c.name, got, c.want)
		}
	}
}
