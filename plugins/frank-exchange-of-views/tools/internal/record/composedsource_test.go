package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// composedProblem and composedFix are a gap's problem and fix whose first sentences the risk matrix
// prints, and which the voice lens's item quotes.
const (
	composedProblem = "The composed row states as confirmed what its evidence does not reach. It names two sequences."
	composedFix     = "Re-record the fate so that it matches its evidence. Name both sequences."
)

// voiceBoard is a run whose voice lens has sat through one sitting at the head and sits in a second,
// so neither its pin nor its first sitting readies it: what follows is all that can.
func voiceBoard(t *testing.T) *stage {
	b := newStage(t).cast(composedSourceAuditor, evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair")
	return b.sit(b.head(), composedSourceAuditor).register("red-chair").sit(b.head(), composedSourceAuditor).register("red-chair")
}

// composedMint is a gap minted by lens about a section, with the problem and fix the matrix prints.
func composedMint(gap string) *recordpb.Mint {
	return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String("Findings"), GapId: proto.String(gap),
		Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String(composedProblem), RequiredFix: proto.String(composedFix),
		AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
		Severity: recordtest.P(recordpb.Grade_GRADE_LOW), Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
}

// voiceParty is the plan's voice party and the why line that readied it for composed text, or false.
func voiceParty(t *testing.T, run Run) (string, bool) {
	t.Helper()
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(plan.Parties, func(p Party) bool { return p.SeatID == composedSourceAuditor }) {
		return "", false
	}
	return lineOf(plan, composedSourceAuditor+": "), true
}

// THE TEXT THE TOOL COMPOSES READIES THE VOICE LENS AT AN UNMOVED HEAD (gblock 2026-11-10): a fate
// with a reason and another lens's mint each ready it, name what it owes, hold the PASS while it is
// ready, and hand its work list the source — until it sits through a sitting opened after them.
func TestTheVoiceLensAuditsTheTextTheToolComposes(t *testing.T) {
	if _, ready := voiceParty(t, voiceBoard(t).seed()); ready {
		t.Fatal("a voice lens served the report at the head, with no composed text recorded since, is ready")
	}
	proposed := voiceBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true))
	if _, ready := voiceParty(t, proposed.seed()); ready {
		t.Error("a proposal readies the voice lens: it prints no reason or method, and no act could answer a gap about its voice")
	}

	fated := voiceBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, false)).
		add(evLens, composedMint("G1")).add(composedSourceAuditor, composedMint("G-own"))
	run := fated.seed()
	why, ready := voiceParty(t, run)
	if !ready {
		t.Fatal("a fate and another lens's mint since the voice lens last sat do not ready it")
	}
	for _, owed := range []string{"the printed fate of inquiry Q1", "the printed problem and fix of gap G1"} {
		if !strings.Contains(why, owed) {
			t.Errorf("the readiness does not name %q: %q", owed, why)
		}
	}
	if strings.Contains(why, "G-own") {
		t.Errorf("the voice lens owes its own mint: %q", why)
	}
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(bs, func(b Blocker) bool { return b.Kind == BlockerLensReady && b.Subject == composedSourceAuditor }) {
		t.Errorf("the PASS gate does not hold the voice lens ready for the composed text: %+v", bs)
	}

	w, err := WorkOfSeat(run, "lens", composedSourceAuditor)
	if err != nil {
		t.Fatal(err)
	}
	var subjects []string
	for _, it := range w.Sitting.Open {
		if it.Kind != ItemComposedSource {
			continue
		}
		if it.Blocks {
			t.Errorf("a composed source blocks the sitting: %+v", it)
		}
		subjects = append(subjects, string(it.Subject.Kind)+" "+it.Subject.ID)
		if it.Subject.Kind == SubjectGap && !(strings.Contains(it.What, anchor.Concise(composedProblem)) && strings.Contains(it.What, anchor.Concise(composedFix))) {
			t.Errorf("the gap's item does not quote the cells the risk matrix prints: %q", it.What)
		}
		if it.Subject.Kind == SubjectInquiry && !strings.Contains(it.What, "Future research directions") {
			t.Errorf("the inquiry's item does not name the section that prints it: %q", it.What)
		}
	}
	if want := []string{"inquiry Q1", "gap G1"}; !slices.Equal(subjects, want) {
		t.Errorf("the voice lens's composed sources are %v, want %v", subjects, want)
	}
	if other, err := WorkOfSeat(run, "lens", evLens); err != nil {
		t.Fatal(err)
	} else if slices.ContainsFunc(other.Sitting.Open, func(i Item) bool { return i.Kind == ItemComposedSource }) {
		t.Error("a lens other than the voice lens is handed composed sources")
	}
	q := QueueOf(w)
	if len(q) == 0 || q[len(q)-1].Kind != ItemComposedSource || len(q[len(q)-1].Items) != 2 {
		t.Errorf("the queue does not end with one entry holding both composed sources: %+v", q)
	}

	// A DOCKETED GAP IS THE BENCH'S UNTIL IT RULES, and leaves the voice lens's list and debt with it.
	docketed := voiceBoard(t).add(evLens, composedMint("G1")).docketMotion("red-chair", "M-1", "G1")
	if why, ready := voiceParty(t, docketed.seed()); ready && strings.Contains(why, "G1") {
		t.Errorf("a gap docketed and unruled readies the voice lens: %q", why)
	}

	// THE SITTING IT IS IN STILL LISTS THEM: a register no agent opened is sat through the moment it
	// opens, and the list reads the sittings before it.
	sitting := voiceBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, false))
	sitting = sitting.register("red-chair").sit(sitting.head(), composedSourceAuditor)
	if w, err := WorkOfSeat(sitting.seed(), "lens", composedSourceAuditor); err != nil {
		t.Fatal(err)
	} else if !slices.ContainsFunc(w.Sitting.Open, func(i Item) bool { return i.Kind == ItemComposedSource }) {
		t.Error("the sitting the voice lens was readied for does not list the fate it was readied by")
	}

	// A RETIRED VOICE LENS IS READIED ALL THE SAME: the composed text is its debt (lensOwesOf), and a
	// lens that owes does not retire.
	retired := voiceBoard(t)
	retired = retired.sit(retired.head(), composedSourceAuditor).register("red-chair").sit(retired.head(), composedSourceAuditor).register("red-chair")
	if why, ready := voiceParty(t, retired.seed()); ready {
		t.Fatalf("the quiet voice lens is ready before any composed text: %q", why)
	}
	retired = voiceBoard(t)
	retired = retired.sit(retired.head(), composedSourceAuditor).register("red-chair").sit(retired.head(), composedSourceAuditor).register("red-chair").
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED, false))
	if why, ready := voiceParty(t, retired.seed()); !ready || !strings.Contains(why, "the printed fate of inquiry Q1") {
		t.Errorf("a voice lens three quiet sittings in is not readied for a fate since: ready %v, %q", ready, why)
	}

	// A SITTING OPENED AFTER THEM, SAT THROUGH, ENDS THE DEBT: its next opening closes it.
	had := voiceBoard(t).add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_DEFERRED, false)).
		add(evLens, composedMint("G1"))
	had = had.register("red-chair").sit(had.head(), composedSourceAuditor).register("red-chair").sit(had.head(), composedSourceAuditor)
	if why, ready := voiceParty(t, had.seed()); ready {
		t.Errorf("a voice lens that sat through a sitting opened after the fate and the mint is still ready: %q", why)
	}
}

// THE COMPOSED SOURCES ARE QUEUED ONLY WHEN NOTHING ELSE IS: no act discharges them, so queued behind
// other work they would hold an item arriving later out of `current`.
func TestTheComposedEntryYieldsToOtherWork(t *testing.T) {
	composed := Item{Kind: ItemComposedSource, Subject: about(SubjectInquiry, "Q1"), What: "inquiry Q1 prints"}
	owed := Item{Kind: ItemKind(BlockerUncheckedInquiry), Subject: about(SubjectInquiry, "Q2"), What: "inquiry Q2 is settled", Blocks: true}
	alone := QueueOf(WorkJSON{Sitting: SittingJSON{Open: []Item{composed}}})
	if len(alone) != 1 || alone[0].Kind != ItemComposedSource {
		t.Errorf("composed sources alone are not the queue: %+v", alone)
	}
	beside := QueueOf(WorkJSON{Sitting: SittingJSON{Open: []Item{composed, owed}}})
	if len(beside) != 1 || beside[0].Kind != ItemKind(BlockerUncheckedInquiry) {
		t.Errorf("composed sources beside owed work are queued: %+v", beside)
	}
}

// A CAST WITHOUT THE VOICE LENS OWES NOBODY THE COMPOSED TEXT.
func TestNoVoiceLensNoComposedDebt(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().register("red-chair").
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED, true)).
		add("blue-respond", fate("Q1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, false))
	db, err := openRunForRead(b.seed())
	if err != nil {
		t.Fatal(err)
	}
	evs, win, err := eventsAt(db)
	if err != nil {
		t.Fatal(err)
	}
	if owed := composedOwedOf(evs, win.IDs(evs), win, nil); owed != nil {
		t.Errorf("a cast with no voice lens owes composed text: %+v", owed)
	}
}
