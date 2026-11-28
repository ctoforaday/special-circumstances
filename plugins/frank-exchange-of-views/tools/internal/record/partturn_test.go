package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A QUESTION PART'S COVERAGE ANSWERS THE GAP ABOUT THE PART ONLY BY NAMING IT; A CHANGED COVERAGE NAMING
// NO GAP IS THE LENS'S TURN (plans/findings-fold.md §III.B3, F-FATE 2026-11-24, T-B45). Part S-1 stands
// `covered` with reason "r" before the lens mints G1 about it; blue's sitting engaged on G1 records the
// coverage given. Named and changed, the round it opens moves G1 and the lens's list says it was
// answered; changed and unnamed, the round stalls, blue did not act on G1, the lens's list says what
// changed and that no gap was named, and the lens owes that change as news (lensOwesOf) as it owes an
// answer.
func TestACoverageAnswersOnlyTheGapItNames(t *testing.T) {
	cover := func(cov recordpb.PartCoverage, reason, answers string) *recordpb.QuestionPart {
		p := &recordpb.QuestionPart{PartId: proto.String("S-1"), SupersedesCoverage: proto.String("1"), Coverage: cov.Enum(), Reason: proto.String(reason)}
		if answers != "" {
			p.Answers = proto.String(answers)
		}
		return p
	}
	staged := func(t *testing.T, recorded *recordpb.QuestionPart) *stage {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
		st.add("blue-respond", &recordpb.QuestionPart{PartId: proto.String("S-1"), Text: proto.String("whether 91 is prime")})
		st.add("blue-respond", cover(recordpb.PartCoverage_PART_COVERAGE_COVERED, "r", ""))
		m := corrMint("G1")
		m.AboutKind, m.AboutRef = recordpb.AboutKind_ABOUT_KIND_QUESTION.Enum(), proto.String("S-1")
		st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).add(evLens, m).register("red-chair")
		return st.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-p").add("blue-respond", recorded).stop("blue-p")
	}
	for _, c := range []struct {
		name      string
		part      *recordpb.QuestionPart
		stalled   int
		acted     bool
		item, not string
	}{
		{"a changed coverage naming the gap", cover(recordpb.PartCoverage_PART_COVERAGE_EXCLUDED, "out of scope", "G1"), 0, true,
			"gap G1 has been ANSWERED by a question part record naming it — blue recorded question part S-1 as `excluded`", "naming no gap"},
		{"a changed coverage naming no gap", cover(recordpb.PartCoverage_PART_COVERAGE_EXCLUDED, "out of scope", ""), 1, false,
			"gap G1: blue recorded question part S-1 as `excluded`, naming no gap — a change to what the gap is about and your turn", "has been ANSWERED"},
	} {
		t.Run(c.name, func(t *testing.T) {
			run := staged(t, c.part).register("red-chair").seed()
			if !workMentions(t, run, "lens", evLens, c.item) || workMentions(t, run, "lens", evLens, c.not) {
				t.Errorf("the lens's list does not say %q, or says %q", c.item, c.not)
			}
			fam, err := MergedEvents(run)
			if err != nil {
				t.Fatal(err)
			}
			owes := lensOwesOf(fam.Events, fam.At.IDs(fam.Events), fam.At, []blockerGap{{id: "G1", assignee: evLens}}, nil)
			if !slices.ContainsFunc(owes[evLens], func(s string) bool { return strings.Contains(s, "G1") }) {
				t.Errorf("the lens owes nothing on G1 after blue's coverage: %q", owes[evLens])
			}
			x, err := Exchanges(staged(t, c.part).reauditOn("G1").register("red-chair").seed(), DefaultParams)
			if err != nil {
				t.Fatal(err)
			}
			if g := x["G1"]; g == nil || g.Exchanges != 1 || g.Stalled != c.stalled || g.rounds[0].acted != c.acted {
				t.Errorf("want 1 exchange, %d stalled, acted %v: %+v", c.stalled, c.acted, g)
			}
		})
	}
}

// A CHANGED FATE NAMING NO GAP IS NEWS THE LENS OWES (gapNewsOf): it is no answer of blue's, and the
// lens owes its audit of the change all the same, as it owes a reaching retire.
func TestAnUnnamedFateIsNewsTheLensOwes(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	st.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	st.add("blue-respond", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r"))
	st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).add(evLens, inquiryMint("G1", "Q-1")).register("red-chair")
	run := st.dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-f").
		add("blue-respond", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found")).stop("blue-f").register("red-chair").seed()
	fam, err := MergedEvents(run)
	if err != nil {
		t.Fatal(err)
	}
	owes := lensOwesOf(fam.Events, fam.At.IDs(fam.Events), fam.At, []blockerGap{{id: "G1", assignee: evLens}}, nil)
	if !slices.Contains(owes[evLens], "blue's fate `confirmed` on Q-1, which changed what G1 is about") {
		t.Errorf("the lens does not owe the unnamed fate's change: %q", owes[evLens])
	}
}
