package record

import (
	"slices"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// m38Archive is universe-m38: CEILING at epoch 12, with the adversary's G-ca807516 docketed in the final
// sitting with no lens re-audit of blue's two answers to it. It holds the bundle store, so it stands
// beside m37's, outside the migrate package's archive census.
const m38Archive = "bundle-store/2026-11-04_is-91-prime-m38.tar.gz"

// UNIVERSE-M38, REPLAYED: a fate blue changed changes what the gaps about it are about — the inquiry,
// and a section the tool composes that prints it (ruled by gblock 2026-11-05) — and, naming none of
// them, it is each one's raiser's turn and no answer (F-FATE, 2026-11-24).
//
//   - G-ca807516, about the section "Research areas", which the tool composes from fates. Blue recorded
//     eight fates again at e10, every status kept and every reason and method reworded. The plan that
//     followed held the adversary off it ("not until blue has acted on it"), and at e12 docketed it
//     with no re-audit. Folded again, the plan after blue's sitting readies the adversary on it — its
//     turn — the docket is not filed, and capture's docket-order audit names the archived one.
//   - G-845422d8, anchored on Q-b62beca0. Blue recorded `confirmed` again at e2 and at e3 with a
//     narrowed reason, naming no gap. Folded again, each is the lens's turn and moves nothing: the
//     rounds are stalls in which blue did not act on the gap.
func TestM38AChangedFateIsTheLensTurn(t *testing.T) {
	const section, anchored, adversary = "G-ca807516", "G-845422d8", "red-lens-adversary"
	whole := archivedEvents(t, m38Archive)
	evs, win := mustEvents(t, archiveAt(t, m38Archive, whole, len(whole)))
	if len(evs) != len(whole) {
		t.Fatalf("the replay holds %d events, the archive %d", len(evs), len(whole))
	}
	mintOf := func(g string) int {
		return slices.IndexFunc(evs, func(e *Event) bool { return e.GetMint().GetGapId() == g })
	}
	// blueFateAfter is the position of blue-respond's first record of a fate after from, on inquiry q
	// ("" for any).
	blueFateAfter := func(from int, q string) int {
		for i := from + 1; i < len(evs); i++ {
			if b, ok := recordpb.BodyAs[*recordpb.Inquiry](evs[i]); ok && evs[i].GetSeatId() == blueRespondSeat && (q == "" || b.GetInquiryId() == q) {
				return i
			}
		}
		t.Fatalf("no fate of blue's on %q after position %d", q, from)
		return 0
	}
	// nextPlan is the stream position of the first dispatch row of the first chair dispatch after from:
	// the record as the plan that wrote it read it.
	nextPlan := func(from int) int {
		for _, g := range DispatchGroups(evs, win) {
			if g.First > from {
				return g.First
			}
		}
		t.Fatalf("no dispatch after position %d", from)
		return 0
	}

	t.Run(section, func(t *testing.T) {
		answered := blueFateAfter(mintOf(section), "")
		at := nextPlan(answered)
		plan, err := PlanDispatch(archiveAt(t, m38Archive, whole, at))
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(gapsOf(plan, adversary), section) {
			t.Errorf("blue changed fates in a sitting engaged on %s, and the plan does not ready %s to re-audit them: parties %+v, why %q", section, adversary, plan.Parties, plan.Why)
		}
		docket := slices.IndexFunc(evs, func(e *Event) bool { return e.GetDocketing().GetGapId() == section })
		if docket < 0 {
			t.Fatalf("the archive holds no docketing on %s", section)
		}
		start := docket
		for start > 0 {
			if _, ok := recordpb.BodyAs[*recordpb.Docketing](evs[start-1]); !ok {
				break
			}
			start--
		}
		before, err := PlanDispatch(archiveAt(t, m38Archive, whole, start))
		if err != nil {
			t.Fatal(err)
		}
		if slices.Contains(docketGaps(before), section) {
			t.Errorf("the plan dockets %s with %s's audit of blue's answer still to come: to file %v, why %q", section, adversary, docketGaps(before), before.Why)
		}
		// CAPTURE NAMES IT.
		early, err := EarlyDockets(archiveAt(t, m38Archive, whole, len(whole)))
		if err != nil {
			t.Fatal(err)
		}
		if !slices.ContainsFunc(early, func(d EarlyDocket) bool {
			return d.GapID == section && d.Docketing == evs[docket].GetDocketing().GetDocketingId()
		}) {
			t.Errorf("capture's docket-order audit does not name the docket on %s: %+v", section, early)
		}
	})

	t.Run(anchored+" (the fold's own reading)", func(t *testing.T) {
		first := blueFateAfter(mintOf(anchored), "Q-b62beca0")
		second := blueFateAfter(first, "Q-b62beca0")
		x, err := Exchanges(archiveAt(t, m38Archive, whole, nextPlan(second)), DefaultParams)
		if err != nil {
			t.Fatal(err)
		}
		g := x[anchored]
		if g == nil || g.Exchanges == 0 || g.Stalled != g.Exchanges || !g.lensTurn {
			t.Fatalf("blue narrowed Q-b62beca0's reason in each round on %s naming no gap: every round a stall and the lens's turn come, got %+v", anchored, g)
		}
		for _, r := range g.rounds {
			if r.acted {
				t.Errorf("a round on %s reads blue as having acted on it, and no act of blue's named it: %+v", anchored, g.rounds)
			}
		}
	})
}

// mintedAbout is a stage at its second chair sitting: the lens minted G1, high, about the named report
// section, and nobody has acted on G1 since. The inquiry Q-1 stands `underway`, with reason "r".
func mintedAbout(t *testing.T, section string) *stage {
	sev, _ := GradeOf("high")
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	st.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	st.add("blue-respond", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r"))
	return st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).
		add(evLens, &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_SECTION.Enum(), AboutRef: proto.String(section), GapId: proto.String("G1"),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}).
		register("red-chair")
}

// fateOn is blue's record of q's fate.
func fateOn(q string, status recordpb.InquiryStatus, reason string) *recordpb.Inquiry {
	return &recordpb.Inquiry{InquiryId: proto.String(q), SupersedesStatus: proto.String("1"), Status: status.Enum(), Reason: proto.String(reason)}
}

// THE LENS'S TURN COMES WITH A FATE THAT CHANGED WHAT ITS GAP IS ABOUT, AND WITH NO OTHER (ruled by
// gblock 2026-11-05). Blue, engaged on G1, records Q-1's fate and nothing else. The plan after blue's
// sitting readies G1's lens only where the fate changed what G1 is about: a section that prints Q-1.
// A reworded `underway` on a gap about a section blue writes readies nobody, and a dispatch of the
// lens there is one capture's lens-turn audit names; so is one after a fate repeated word for word.
func TestALensTurnComesWithAFateAboutItsGap(t *testing.T) {
	for _, c := range []struct {
		name, section string
		fate          *recordpb.Inquiry
		turn          bool
	}{
		{"a changed fate the gap's section prints", "Research areas", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found"), true},
		{"a reworded underway on a gap about other text", "Findings", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r, narrowed"), false},
		{"a repeated fate the gap's section prints", "Research areas", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r"), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			st := mintedAbout(t, c.section)
			st.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-fates").add("blue-respond", c.fate).stop("blue-fates")
			plan, err := PlanDispatch(st.register("red-chair").seed())
			if err != nil {
				t.Fatal(err)
			}
			if got := slices.Contains(gapsOf(plan, evLens), "G1"); got != c.turn {
				t.Errorf("the plan readies %s on G1 = %v, want %v: parties %+v, why %q", evLens, got, c.turn, plan.Parties, plan.Why)
			}
			early, err := EarlyLensDispatches(st.more().dispatch(2, evLens, "G1").seed())
			if err != nil {
				t.Fatal(err)
			}
			if got := slices.ContainsFunc(early, func(d EarlyLensDispatch) bool { return d.GapID == "G1" }); got == c.turn {
				t.Errorf("capture names a dispatch of the lens on G1 early = %v, want %v: %+v", got, !c.turn, early)
			}
		})
	}
}

// remandedAbout is G1, about the named section, stalled to impasse under the default terms, docketed by
// the chair and remanded by the bench with a stated direction.
func remandedAbout(t *testing.T, section string) *stage {
	st := mintedAbout(t, section)
	st.exchangeOn("G1").exchangeOn("G1")
	return st.docketAndRemandWith("M1", "G1", remandDirection)
}

// A FATE ANSWERS A REMAND ONLY WHERE IT NAMES THE GAP (plans/findings-fold.md §III.B3, T-B45). Blue's
// sitting for the remand records Q-1's fate, changed, and nothing else. Naming G1, about the section
// that prints Q-1, it is blue's answer: its item stops holding the sitting, and the next plan readies
// the lens alone to audit it. Naming no gap, it is the lens's turn and discharges no remand: the item
// still blocks, and the lens is readied all the same. Where G1 is about a section blue writes, the fate
// changes nothing G1 is about, and a fate repeated word for word changes nothing.
func TestAChangedFateAnswersARemandOnlyOfAGapAboutIt(t *testing.T) {
	named := func(f *recordpb.Inquiry) *recordpb.Inquiry { f.Answers = proto.String("G1"); return f }
	for _, c := range []struct {
		name, section string
		fate          *recordpb.Inquiry
		answers, turn bool
	}{
		{"a changed fate the gap's section prints, naming it", "Research areas", named(fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found")), true, true},
		{"a changed fate the gap's section prints, naming no gap", "Research areas", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found"), false, true},
		{"an unrelated changed fate", "Findings", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found"), false, false},
		{"a repeated fate", "Research areas", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r"), false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			st := remandedAbout(t, c.section)
			st.register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-fates").add("blue-respond", c.fate)
			run := st.seed()
			if blocks := remandItemBlocks(t, run, "blue", "blue-respond"); blocks == c.answers {
				t.Errorf("blue's remand item on G1 blocks = %v, want %v", blocks, !c.answers)
			}
			after, err := PlanDispatch(st.more().stop("blue-fates").register("red-chair").seed())
			if err != nil {
				t.Fatal(err)
			}
			if got := slices.Contains(gapsOf(after, evLens), "G1"); got != c.turn {
				t.Errorf("the plan readies %s on G1 = %v, want %v: parties %+v, why %q", evLens, got, c.turn, after.Parties, after.Why)
			}
		})
	}
}
