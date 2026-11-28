package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// THE LOUD SURFACE TAKES ITS WORD EXACTLY: the schema's word in capitals, and nothing the tool would
// have to fold into it.
func TestTheLoudOutcomeWordIsTakenExactly(t *testing.T) {
	if v, ok := RunOutcomeOf("VERIFIED"); !ok || v != recordpb.RunOutcome_RUN_OUTCOME_VERIFIED {
		t.Errorf("RunOutcomeOf(VERIFIED) = %v, %v", v, ok)
	}
	for _, w := range []string{"verified", "Unverified", " VERIFIED", ""} {
		if v, ok := RunOutcomeOf(w); ok {
			t.Errorf("RunOutcomeOf(%q) = %v; want it refused", w, v)
		}
	}
}

// AN ABOUT ANCHOR IS TAKEN AS TYPED: space around either half is refused, and a pair of space alone
// is not the absent pair.
func TestAnAboutAnchorIsTakenAsTyped(t *testing.T) {
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, recordtest.At(t, HarnessSeat, "harness:report", &recordpb.ReportPart{Text: proto.String("# H\n\nSeven is prime.\n")}))
	run := mustRun(t, dir)
	admitted := MustEnum("mint", "about_kind")
	for _, c := range [][2]string{{" section", "H"}, {"section", "H "}, {"  ", "  "}, {"section", "   "}} {
		if k, ref, err := ResolveAbout(run, admitted, c[0], c[1]); err == nil {
			t.Errorf("ResolveAbout(%q, %q) = %v %v; want it refused", c[0], c[1], k, ref)
		}
	}
	if k, ref, err := ResolveAbout(run, admitted, "section", "H"); err != nil || k == nil || *ref != "H" {
		t.Errorf("ResolveAbout(section, H) = %v %v %v", k, ref, err)
	}
	if k, ref, err := ResolveAbout(run, admitted, "", ""); err != nil || k != nil || ref != nil {
		t.Errorf("ResolveAbout of no anchor = %v %v %v; want none", k, ref, err)
	}
}
