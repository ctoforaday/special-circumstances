package cli

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A MINT SUPERSEDES ONE GAP (T-B18). Two ancestors are refused in both forms — a comma in the value
// and the flag given twice — naming the road: supersede one and close the rest. Neither form is
// narrowed to one of its ids. One ancestor lands, and the mint stores it.
func TestAMintSupersedesOneGap(t *testing.T) {
	runDir := seatRun(t)
	a := mintGap(t, runDir, "one-ancestor-a", "spec_underspecification")
	b := mintGap(t, runDir, "one-ancestor-b", "policy_without_mechanism")
	mint := func(supersedes ...string) (string, error) {
		args := []string{"--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
			"--class", "spec_underspecification", "--problem", "the defect restated", "--fix", "the fix",
			"--check-kind", "document", "--check", "the acceptance check red runs at re-audit",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low"}
		for _, s := range supersedes {
			args = append(args, "--supersedes", s)
		}
		return runMint(t, args...)
	}
	for name, form := range map[string][]string{
		"a comma":        {a + "," + b},
		"the flag twice": {a, b},
	} {
		out, err := mint(form...)
		if err == nil || !strings.Contains(err.Error()+out, flags.OneAncestorRefusal) {
			t.Errorf("--supersedes with %s = %v %q, want refused with %q", name, err, out, flags.OneAncestorRefusal)
		}
	}
	if _, err := mint(a); err != nil {
		t.Fatalf("a mint superseding one gap was refused: %v", err)
	}
	if got := lastBody(t, runDir, &recordpb.Mint{}).GetSupersedes(); got != a {
		t.Errorf("the mint stores supersedes %q, want %q", got, a)
	}
}

// THE GRADE SUBJECT IS GONE (T-B1): no surface mounts `motion grade`, and the schema reserves the
// subject's value and both arms' numbers, so no stored byte is reread as something else.
func TestNoSurfaceMountsAGradeMotion(t *testing.T) {
	runDir := seatRun(t)
	registerLensOnce(t, runDir)
	if out, err := run(t, "motion", "grade", "file", "--run", runDir, "--seat-id", lensSeat); err == nil {
		t.Errorf("`motion grade file` ran: %q", out)
	}
	subj := recordpb.MotionSubject(0).Descriptor()
	if subj.Values().ByName("MOTION_SUBJECT_GRADE") != nil || !subj.ReservedNames().Has("MOTION_SUBJECT_GRADE") || !subj.ReservedRanges().Has(1) {
		t.Error("MOTION_SUBJECT_GRADE is a value, or its name and number are not reserved")
	}
	for _, md := range []protoreflect.MessageDescriptor{
		(&recordpb.Motion{}).ProtoReflect().Descriptor(), (&recordpb.MotionRule{}).ProtoReflect().Descriptor(),
	} {
		if md.Fields().ByName("grade") != nil || !md.ReservedNames().Has("grade") || !md.ReservedRanges().Has(10) {
			t.Errorf("%s.grade is a field, or its name and number 10 are not reserved", md.Name())
		}
	}
	docket := (&recordpb.DocketMotion{}).ProtoReflect().Descriptor()
	for _, n := range []protoreflect.Name{"grounds", "grounds_quote"} {
		if docket.Fields().ByName(n) != nil || !docket.ReservedNames().Has(n) {
			t.Errorf("DocketMotion.%s is a field, or its name is not reserved", n)
		}
	}
}
