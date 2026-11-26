package cli

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

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
