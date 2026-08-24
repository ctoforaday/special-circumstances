package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// EVERY MOTION SUBJECT DECLARES WHO RULES IT.
//
// The gavel used to be a literal argument in internal/cli/motion and nowhere else, so the PASS
// gate — which lives here and cannot import the CLI — refused a merge seat over an unruled
// PETITION while telling it to go and rule the thing. requireRuler then refused that, because the
// bench holds that gavel. The seat had no legal verdict and the round wedged.
//
// Written against the DESCRIPTOR rather than a list of three subjects, so a fourth fails here on
// the day it is added rather than on the day a run blocks behind it.
func TestEveryMotionSubjectNamesItsRuler(t *testing.T) {
	vals := recordpb.MotionSubject(0).Descriptor().Values()
	checked := 0
	for i := 0; i < vals.Len(); i++ {
		v := vals.Get(i)
		if v.Number() == 0 {
			continue // UNSPECIFIED is the absence of a subject, not a subject
		}
		ruler, err := recordpb.RulerOf(v)
		if err != nil {
			t.Errorf("%s: %v", v.Name(), err)
			continue
		}
		// The role must be one a seat id can actually resolve to, or the refusal names a seat
		// that does not exist — which teaches worse than naming none.
		switch ruler {
		case "merge", "bench":
		default:
			t.Errorf("%s is ruled by %q, which is not a seat role — a refusal naming it sends the reader nowhere", v.Name(), ruler)
		}
		checked++
	}
	if checked < 3 {
		t.Errorf("only %d subjects carry a gavel — the sweep is meant to cover the whole enum", checked)
	}
}

// THE PASS REFUSAL NAMES THE GAVEL, AND NAMES THE MOVE FOR A SEAT THAT DOES NOT HOLD IT.
//
// This is the wedge, stated as a test: a clean gap board plus an unruled PETITION. The merge
// cannot rule it and cannot pass; the only legal act left is FAIL, and the seat learns that from
// this message or not at all.
func TestThePassRefusalNamesWhoHoldsTheGavel(t *testing.T) {
	runDir := tmpRun(t)
	for _, sid := range []string{"blue-respond-r1", "red-merge-r1"} {
		if _, _, err := RegisterSeat(Identity{RunDir: runDir, SeatID: sid, Round: RoundIn(runDir)(sid)}, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{RunDir: runDir, SeatID: "blue-respond-r1", Round: RoundIn(runDir)("blue-respond-r1")}, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
		Basis:    proto.String("the run is proceeding past a safety objection"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
			Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	err := requirePassClosesAllGaps(runDir)
	if err == nil {
		t.Fatal("PASS was allowed over an unruled petition")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bench") {
		t.Errorf("the refusal does not say the BENCH holds this gavel, so a merge seat reads it as work it owes: %v", err)
	}
	if !strings.Contains(msg, "--as FAIL") {
		t.Errorf("the refusal does not name the one act still open to a seat that cannot rule: %v", err)
	}
	// And it must not name the ruler as though the reader were it: the wedge came from an
	// instruction the seat could follow only into a second refusal.
	if !strings.Contains(msg, "IF THE GAVEL NAMED ABOVE IS YOURS") {
		t.Errorf("the rule-it instruction is unconditional, which is what walked a merge seat into requireRuler: %v", err)
	}
}
