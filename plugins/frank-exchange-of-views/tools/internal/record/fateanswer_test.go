package record

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// A CHANGED FATE IS THE LENS'S TURN ON EVERY GAP ABOUT WHAT IT CHANGED, WHATEVER BLUE'S SITTING WAS
// ENGAGED ON (ruled by gblock 2026-11-05; universe-m39 G-385fe4f4), and it answers only the gap it names
// (F-FATE, 2026-11-24). The lens minted G1 and G2, both about Q-1, and blue was engaged on G1 alone.
// Blue changed Q-1's fate naming G1: its answer to G1, and G2's lens's turn. The plan readies the lens
// on G2, and the lens's list says a fate answered G1 and changed what G2 is about — never that an edit
// changed a sentence, which `edited_since` would not hold.
func TestAChangedFateIsTheLensTurnOnAGapBlueWasNotEngagedOn(t *testing.T) {
	sev, _ := GradeOf("high")
	mint := func(gap string) *recordpb.Mint {
		return &recordpb.Mint{AboutKind: recordpb.AboutKind_ABOUT_KIND_INQUIRY.Enum(), AboutRef: proto.String("Q-1"), GapId: proto.String(gap),
			Class: recordpb.GapClass_GAP_CLASS_SPEC_UNDERSPECIFICATION.Enum(), Problem: proto.String("p"),
			AcceptanceCheck: proto.String("c"), CheckKind: recordtest.P(recordpb.CheckKind_CHECK_KIND_DOCUMENT),
			Severity: &sev, Likelihood: recordtest.P(recordpb.Grade_GRADE_MEDIUM), Impact: recordtest.P(recordpb.Grade_GRADE_MEDIUM)}
	}
	namedFate := fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED, "found")
	namedFate.Answers = proto.String("G1")
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	st.add("blue-respond", &recordpb.Inquiry{InquiryId: proto.String("Q-1"), Question: proto.String("q"), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	st.add("blue-respond", fateOn("Q-1", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, "r"))
	st.register("red-chair").dispatch(2, evLens).sitClosed(evLens).add(evLens, mint("G1")).add(evLens, mint("G2")).
		register("red-chair").dispatch(2, "blue-respond", "G1").registerAs("blue-respond", "blue-fates").
		add("blue-respond", namedFate).stop("blue-fates")
	run := st.register("red-chair").seed()

	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(gapsOf(plan, evLens), "G2") {
		t.Fatalf("blue changed Q-1's fate, G2's lens's turn, and the plan does not ready G2's lens: parties %+v, why %q", plan.Parties, plan.Why)
	}
	if !workMentions(t, run, "lens", evLens, "gap G1 has been ANSWERED by an inquiry record naming it — blue recorded the fate `confirmed` on Q-1, the inquiry the gap is anchored on") {
		t.Error("the lens's list does not say the fate naming G1 answered it")
	}
	if !workMentions(t, run, "lens", evLens, "gap G2: blue recorded the fate `confirmed` on Q-1, the inquiry the gap is anchored on, changing its fate, reason or method and naming no gap") {
		t.Error("the lens's list does not say the fate changed what G2 is about, naming no gap")
	}
	if workMentions(t, run, "lens", evLens, "gap G2 has been ANSWERED") {
		t.Error("the lens's list says a fate naming G1 answered G2")
	}
	if workMentions(t, run, "lens", evLens, "an edit changed its sentence") {
		t.Error("the lens's list speaks of an edit no `edited_since` holds")
	}
}
