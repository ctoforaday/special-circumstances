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
// The PASS gate lives here and cannot import the CLI, whose command tree is the gavel. A gate
// that does not know who rules a subject refuses a chair seat over an unruled PETITION while
// telling it to go and rule the thing — a verb its tree does not carry, because the bench holds
// that gavel. The seat has no legal verdict and the round wedges.
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
		// The role must resolve, through the roster, to one seat, or the refusal names a seat
		// that does not exist — which teaches worse than naming none.
		if _, err := gavelSeatOf(ruler); err != nil {
			t.Errorf("%s: %v", v.Name(), err)
		}
		checked++
	}
	if checked < 3 {
		t.Errorf("only %d subjects carry a gavel — the sweep is meant to cover the whole enum", checked)
	}
}

// A GAVEL RESOLVES THROUGH THE ROLE ROSTER TO ONE SEAT, and a word no single seat holds is an
// error rather than a role word printed as a seat id.
func TestAGavelResolvesToOneSeat(t *testing.T) {
	for ruler, want := range map[string]string{"chair": chairSeat, "bench": benchSeat} {
		if got, err := gavelSeatOf(ruler); err != nil || got != want {
			t.Errorf("gavelSeatOf(%q) = %q, %v; want %q", ruler, got, err, want)
		}
	}
	for _, ruler := range []string{"lens", "blue", "nobody", ""} {
		if got, err := gavelSeatOf(ruler); err == nil {
			t.Errorf("gavelSeatOf(%q) = %q, want an error — no single seat holds that gavel", ruler, got)
		}
	}
}

// THE PASS REFUSAL NAMES THE GAVEL, AND NAMES THE MOVE FOR A SEAT THAT DOES NOT HOLD IT.
//
// This is the wedge, stated as a test: a clean gap board plus an unruled PETITION. The chair
// cannot rule it and cannot pass; the only legal act left is FAIL, and the seat learns that from
// this message or not at all.
func TestThePassRefusalNamesWhoHoldsTheGavel(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, sid := range []string{"blue-respond", "red-chair"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: sid}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
		Basis:    proto.String("the run is proceeding past a safety objection"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
			Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	err := passRefusalOver(t, mustRun(t, runDir), BlockerUnruledMotion)
	if err == nil {
		t.Fatal("PASS was allowed over an unruled petition")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bench") {
		t.Errorf("the refusal does not say the BENCH holds this gavel, so a chair seat reads it as work it owes: %v", err)
	}
	if !strings.Contains(msg, "--as FAIL") {
		t.Errorf("the refusal does not name the one act still open to a seat that cannot rule: %v", err)
	}
	// And it must not name the ruler as though the reader were it: the wedge came from an
	// instruction the seat could follow only into a second refusal.
	if !strings.Contains(msg, "IF THE GAVEL NAMED ABOVE IS YOURS") {
		t.Errorf("the rule-it instruction is unconditional, which walks a chair seat into a verb only the bench's tree carries: %v", err)
	}
}

// THE SITTING VIEW AND THE PASS REFUSAL DESCRIBE ONE BLOCKAGE, SO THEY NAME THE SAME GAVEL.
//
// The refusal was rewritten to say who rules each unruled motion, after a blocked chair seat
// followed an unconditional "rule it" instruction into a second refusal. The sitting view — the
// list a seat reads to find out what it owes — still said only that the motion stood, so the item
// the seat could not rule looked exactly like the ones it could.
//
// WHAT THIS DOES NOT CHANGE IS WHO IS BLOCKED. An unruled petition still refuses a chair PASS:
// the run is not finished until the bench answers it. Both halves are asserted, because a "fix"
// that resolved the divergence by dropping the item would pass the first alone.
func TestTheSittingViewNamesTheGavelAndStillBlocks(t *testing.T) {
	runDir := recordtest.TmpRun(t)
	for _, sid := range []string{"blue-respond", "red-chair"} {
		if _, _, err := RegisterSeat(Identity{Run: mustRun(t, runDir), SeatID: sid}, "", ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Append(Identity{Run: mustRun(t, runDir), SeatID: "blue-respond"}, &recordpb.Motion{
		MotionId: proto.String("M1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_PETITION),
		Basis:    proto.String("the run is proceeding past a safety objection"),
		Filing: &recordpb.Motion_Petition{Petition: &recordpb.PetitionMotion{
			Class: recordtest.P(recordpb.PetitionClass_PETITION_CLASS_SAFETY),
		}},
	}); err != nil {
		t.Fatal(err)
	}

	s := sittingOfRunT(t, mustRun(t, runDir), "chair", "red-chair")

	var line string
	for _, o := range s.Open {
		if strings.Contains(o.What, "M1") {
			line = o.What
		}
	}
	if line == "" {
		t.Fatal("the unruled petition is not on the chair's outstanding list at all")
	}
	if !strings.Contains(line, "bench") {
		t.Errorf("the sitting view does not say the BENCH rules this, so the seat reads it as work it owes: %q", line)
	}
	if s.Complete {
		t.Error("the chair sitting is complete over an unruled petition — naming the gavel must not change WHO is blocked")
	}
}
