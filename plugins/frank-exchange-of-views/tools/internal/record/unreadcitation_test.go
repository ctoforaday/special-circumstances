package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// ONE STATE PER ANCHOR (anchorStanding.state), each reached by the board that should reach it, so
// an arm deleted from the switch reads one of these boards as another state and fails here.
func TestEveryAnchorStateIsReached(t *testing.T) {
	retire := &recordpb.Retire{Claim: proto.String("the claim C-1 backs"), Reason: proto.String("gone"), Anchors: []string{"C-1"}}
	for _, c := range []struct {
		name   string
		build  func() *stage
		anchor string
		want   anchorState
	}{
		{"a citation read and backed", func() *stage { return reopenedBoard(t) }, "C-1", stateHolds},
		{"a citation retired", func() *stage { return reopenedBoard(t).add("blue-respond", retire) }, "C-1", stateRetired},
		{"a citation nobody checked", func() *stage { return reopenedBoard(t).cite("C-2") }, "C-2", stateUnchecked},
		{"a citation an edit changed since its check", func() *stage { return reopenedBoard(t).reopen("C-1") }, "C-1", stateUnchecked},
		{"a citation whose check read nothing", func() *stage {
			return reopenedBoard(t).cite("C-2").verify(evLens, "C-2", recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE)
		}, "C-2", stateUnread},
		{"a citation read against", func() *stage {
			return reopenedBoard(t).cite("C-2").verify(evLens, "C-2", recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES)
		}, "C-2", stateFoundAgainst},
		{"a proof judged sound", func() *stage {
			return reopenedBoard(t).proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND)
		}, "P-1", stateHolds},
		{"a proof nobody judged", func() *stage { return reopenedBoard(t).proof("P-1", "s1") }, "P-1", stateUnchecked},
		{"a proof judged unsound", func() *stage {
			return reopenedBoard(t).proof("P-1", "s1").reproduce(evLens, "P-1", "s1", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_UNSOUND)
		}, "P-1", stateUnsound},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := loadedT(t, c.build().evs...)
			s := anchorChecksOf(m.Events, m.At).at(c.anchor)
			if s == nil {
				t.Fatalf("the ledger holds no %s", c.anchor)
			}
			if got := s.state(); got != c.want {
				t.Errorf("%s's state = %d, want %d", c.anchor, got, c.want)
			}
		})
	}
}

// aboutAnchor is a material gap about anchor, by the lens whose check of it failed.
func aboutAnchor(gap, anchor string) *recordpb.Mint {
	m := aboutProof(gap, anchor)
	m.Severity = recordpb.Grade_GRADE_HIGH.Enum()
	m.Problem = proto.String("the cited source could not be read, and the report still cites it")
	return m
}

// unreadBoard is C-1 verified unreachable by the evidence lens in its sitting: a failed check, owned
// by that lens.
func unreadBoard(t *testing.T) *stage {
	return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingestCited().register("red-chair").
		dispatch(2, evLens).register(evLens).cite("C-1").sit(2, evLens).
		verify(evLens, "C-1", recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE)
}

func failedOf(t *testing.T, run Run) []Blocker {
	t.Helper()
	var out []Blocker
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerFailedCheck {
			out = append(out, b)
		}
	}
	return out
}

// THE BENCH DOES NOT CLEAR AN UNREAD CITATION (gblock, 2026-10-25: "Bench: STILL BLOCKS"). The
// bench may rule on the gap that carries it — here it accepts the defect, which closes the gap — and
// the citation still stands unread: the failed check holds the PASS, and the run that ends there
// names it.
func TestABenchRulingDoesNotClearAnUnreadCitation(t *testing.T) {
	run := unreadBoard(t).add(evLens, aboutAnchor("G-1", "C-1")).
		register("red-chair").docketMotion("red-chair", "M1", "G-1").
		dispatch(2, "judge", "G-1").register("judge").add("judge", docketRuling("M1", recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED)).
		register("red-chair").seed()
	fam, err := FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	for _, g := range workStatesOfFamilyT(fam) {
		if g.ID == "G-1" && g.Open {
			t.Fatalf("fixture: G-1 is open after the bench accepted the defect")
		}
	}
	got := failedOf(t, run)
	if len(got) != 1 || got[0].Subject != "C-1" || got[0].AnsweredBy() != "G-1" {
		t.Fatalf("failed checks after the bench's ruling = %+v, want C-1's, carried by the closed G-1", got)
	}
	err = validateRun(run, "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate)
	if err == nil || !strings.Contains(err.Error(), "C-1 (read nothing of its source by "+evLens) || !strings.Contains(err.Error(), "no bench ruling clears it") {
		t.Errorf("the PASS over C-1 after the bench's ruling: %v, want it refused naming C-1", err)
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if plan.PassPermitted {
		t.Errorf("pass_permitted over an unread citation the bench ruled on: %q", plan.Why)
	}
	if got := anchorsStanding(plan.Blockers); got != "; on a failed check: C-1" {
		t.Errorf("the outcome's clause = %q, want it to name C-1", got)
	}
}

// THE GAP THAT CARRIES A FAILED CHECK DOES NOT CLOSE WHILE THE CHECK STANDS (requireFailedCheckAnswered),
// whatever its class; one that names a live successor hands the defect on and is recorded; and once
// blue retires the citation, it closes.
func TestTheGapAboutAFailedCheckClosesOnlyOnceTheCheckIsAnswered(t *testing.T) {
	closeAs := func(gap string, word recordpb.Disposition) *recordpb.Close {
		return &recordpb.Close{GapId: proto.String(gap), ClosureClass: recordtest.P(word),
			AnchorSeat: proto.String(evLens), AnchorTool: proto.String("show evidence"), AnchorTarget: proto.String("C-1"), Prose: proto.String("checked")}
	}
	run := unreadBoard(t).seed()
	lens := Identity{Run: run, SeatID: evLens}
	mustAppend(t, lens, aboutAnchor("G-1", "C-1"))
	for _, word := range []recordpb.Disposition{recordpb.Disposition_DISPOSITION_REPAIRED, recordpb.Disposition_DISPOSITION_NOT_A_DEFECT, recordpb.Disposition_DISPOSITION_DEFECT_ACCEPTED} {
		_, err := Append(lens, closeAs("G-1", word))
		mustRefuse(t, err, "gap G-1 carries citation C-1", "Nothing is recorded", "retires the citation")
	}
	mustAppend(t, lens, budgetMint("G-2"))
	handed := closeAs("G-1", recordpb.Disposition_DISPOSITION_REPAIRED_WITH_REGRESSION)
	handed.Successor = proto.String("G-2")
	mustAppend(t, lens, handed)

	// Blue retires the citation and keeps its sentence (the verb's guard is cli/blue's; the record
	// layer is under test here).
	run = unreadBoard(t).add(evLens, aboutAnchor("G-1", "C-1")).
		add("blue-respond", &recordpb.Retire{Claim: proto.String("the claim C-1 backs"), Reason: proto.String("unreadable"), Anchors: []string{"C-1"}, ClaimKept: proto.Bool(true)}).seed()
	lens = Identity{Run: run, SeatID: evLens}
	if got := failedOf(t, run); len(got) != 0 {
		t.Fatalf("failed checks after blue retired C-1 = %+v, want none", got)
	}
	mustAppend(t, lens, closeAs("G-1", recordpb.Disposition_DISPOSITION_REPAIRED))
}

// A SECOND READING THAT READS NOTHING, BY THE SEAT WHOSE GAP CARRIES THE FIRST, IS REFUSED while the
// gap is open and the sentence unchanged; a reading of the source is recorded and clears it.
func TestARepeatUnreadReadingIsRefused(t *testing.T) {
	run := unreadBoard(t).add(evLens, aboutAnchor("G-1", "C-1")).seed()
	unreachable := &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("the claim C-1 backs"), Text: proto.String("still a dead link"),
		Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_UNREACHABLE.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(),
		CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()}
	err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_VERIFY, unreachable)
	mustRefuse(t, err, "your check of citation C-1", "gap G-1", "Nothing is recorded")
	reads := proto.Clone(unreachable).(*recordpb.Verify)
	reads.Outcome = recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum()
	if err := validateRun(run, evLens, recordpb.EventType_EVENT_TYPE_VERIFY, reads); err != nil && strings.Contains(err.Error(), "your check of citation") {
		t.Errorf("a reading of the source is refused as a repeat: %v", err)
	}
	// Another lens's reading that reads nothing is its own.
	if err := validateRun(run, "red-lens-logic", recordpb.EventType_EVENT_TYPE_VERIFY, unreachable); err != nil && strings.Contains(err.Error(), "your check of citation") {
		t.Errorf("another lens's unreachable is refused as this one's repeat: %v", err)
	}
}

// THE UNREAD CITATION'S GAP IS OUTSIDE THE BUDGET of the lens whose check read nothing, and of no
// other lens.
func TestTheGapAboutAnUnreadCitationTakesNoMint(t *testing.T) {
	run := unreadBoard(t).seed()
	q := recordOf(t, run)
	if got, err := answersOwnFailedCheckAt(q, evLens, aboutAnchor("G-1", "C-1")); err != nil || got != "C-1" {
		t.Errorf("the evidence lens's gap about C-1 answers %q (%v), want C-1", got, err)
	}
	if got, err := answersOwnFailedCheckAt(q, "red-lens-logic", aboutAnchor("G-1", "C-1")); err != nil || got != "" {
		t.Errorf("another lens's gap about C-1 answers %q (%v), want nothing", got, err)
	}
}
