package recordpb

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
)

// THE CHECK MUST ACTUALLY REFUSE, because the thing it replaces did not.
//
// `Outcome.verdict` was declared required in a map that no writer read. The declaration existed,
// the test that would have caught it could not run, and the requirement was inert — a run could
// record that it ENDED and not how. So the assertion here is not that the annotation is present
// but that a body missing it is REFUSED, and that the refusal names the flag a seat would type.
func TestARequiredFieldIsRefusedWhenAbsent(t *testing.T) {
	if err := CheckRequired("bench outcome", &Outcome{Prose: proto.String("ended on safety grounds")}); err == nil {
		t.Fatal("an outcome with no verdict was accepted — the run records that it ended and not how, " +
			"and every reader downstream sees a run that never reached a verdict")
	} else if !strings.Contains(err.Error(), "--as") {
		t.Errorf("the refusal does not name the flag that fixes it: %v", err)
	}

	complete := &Outcome{
		Verdict: RunOutcome_RUN_OUTCOME_HALTED.Enum(),
		Prose:   proto.String("ended on safety grounds"),
	}
	if err := CheckRequired("bench outcome", complete); err != nil {
		t.Errorf("a complete outcome was refused: %v — the requirement is a gate, not a wall", err)
	}
}

// PRESENT AND EMPTY IS NOT ABSENT, and the distinction is one a seat relies on.
//
// `--review-flag ""` is a legitimate ruling: the bench says there is nothing a human need look at.
// A check written against emptiness rather than presence would refuse it, and the seat would have
// no way to say that at all.
func TestAnEmptyValueTheSeatPassedSatisfiesTheRequirement(t *testing.T) {
	body := &Opinion{
		GapId:       proto.String("R1-1"),
		Disposition: proto.String("carried"),
		Principle:   proto.String("p"),
		Tension:     proto.String("t"),
		ReviewFlag:  proto.String(""), // said, and said to be nothing
		Rationale:   proto.String("r"),
	}
	if err := CheckRequired("bench opinion", body); err != nil {
		t.Errorf("an explicitly empty review flag was refused: %v", err)
	}
}

// EVERY REQUIRED FIELD CAN NAME THE FLAG THAT FIXES IT.
//
// A refusal that says only "required" sends a seat looking for a flag it has to guess, and a seat
// that cannot find the flag it was told to pass works around the verb — losing the capability for
// the whole run, which cli/motion/verbs.go records as measured behaviour.
func TestEveryRequiredFieldCarriesItsReason(t *testing.T) {
	ev := (&Event{}).ProtoReflect().Descriptor()
	od := ev.Oneofs().ByName("body")
	n := 0
	for i := 0; i < od.Fields().Len(); i++ {
		md := od.Fields().Get(i).Message()
		for _, fd := range RequiredOf(md) {
			n++
			o, _ := proto.GetExtension(fd.Options(), E_Sql).(*Sql)
			if o.GetWhy() == "" {
				t.Errorf("%s is required and says nothing about why — the seat is refused and told only that it was", fd.FullName())
			}
		}
	}
	if n == 0 {
		t.Fatal("no field is required anywhere — a no-match here would pass this test on an empty schema")
	}
}
