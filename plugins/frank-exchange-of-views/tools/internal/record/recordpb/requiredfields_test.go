package recordpb

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
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
		Disposition: Disposition_DISPOSITION_CARRIED.Enum(),
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

// EVERY REQUIRED FIELD IS REFUSED WHEN ABSENT, on every body.
//
// This is what makes the per-verb `if` blocks in validate safe to delete. Each was an unconditional
// presence check on a field the schema now marks required, and CheckRequired runs BEFORE the switch
// — so the block became unreachable. Unreachable is only a safe thing to delete if the thing that
// shadows it fires for the same inputs, and this drives exactly that: a body carrying every
// required field but one, for each one.
func TestEveryRequiredFieldIsRefusedOnEveryBody(t *testing.T) {
	ev := (&Event{}).ProtoReflect().Descriptor()
	od := ev.Oneofs().ByName("body")
	checked := 0
	for i := 0; i < od.Fields().Len(); i++ {
		md := od.Fields().Get(i).Message()
		req := RequiredOf(md)
		if len(req) == 0 {
			continue
		}
		for _, omit := range req {
			m := dynamicpb.NewMessage(md)
			for _, fd := range req {
				if fd.FullName() == omit.FullName() {
					continue
				}
				m.Set(fd, sample(fd))
			}
			err := CheckRequired(string(md.Name()), m)
			if err == nil {
				t.Errorf("%s: a body missing only %s was accepted", md.Name(), omit.Name())
				continue
			}
			checked++
			// The refusal must name the omitted field, or it sends the seat to the wrong flag.
			if !strings.Contains(err.Error(), flagFor(omit, nil)) && !strings.Contains(err.Error(), string(omit.Name())) {
				t.Errorf("%s: omitting %s was refused with a message that does not name it: %v", md.Name(), omit.Name(), err)
			}
		}
	}
	if checked == 0 {
		t.Fatal("nothing was driven — a no-match here would pass on a schema with no requirements at all")
	}
	t.Logf("drove %d required-field omissions", checked)
}

func sample(fd protoreflect.FieldDescriptor) protoreflect.Value {
	switch fd.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString("x")
	case protoreflect.BoolKind:
		return protoreflect.ValueOfBool(true)
	case protoreflect.Int32Kind:
		return protoreflect.ValueOfInt32(1)
	case protoreflect.Int64Kind:
		return protoreflect.ValueOfInt64(1)
	case protoreflect.EnumKind:
		// The first NON-ZERO value: the zero is UNSPECIFIED, which the record reserves for absence,
		// so seeding it would make a present field read as missing and the test pass for the wrong
		// reason.
		return protoreflect.ValueOfEnum(fd.Enum().Values().Get(1).Number())
	}
	return protoreflect.ValueOfString("x")
}

// A CONDITIONAL REQUIREMENT MARKED UNCONDITIONAL DOES NOT BECOME STRICTER — IT BECOMES WRONG.
//
// `Avenue.line` carried `required: true` while the message's own doc comment said, in as many
// words, "a MOVE names the avenue and carries only the new status and its reason, which is why
// `line` cannot be required unconditionally". The annotation contradicted the paragraph above it,
// and it did two things at once: CheckRequired refused every move before the conditional check in
// record.go could run, and the derived DDL put NOT NULL on the column so the row could not be
// stored either. `blue line-of-inquiry move` was unusable.
//
// # Why this test is shaped as a census rather than a case
//
// One case would pin `Avenue.line` and nothing else. What recurs is the CLASS: a requirement that
// depends on another field, written as if it depended on nothing. The `check` extension exists for
// exactly those, and the split is what required.go always stated and could not enforce — "ONLY
// UNCONDITIONAL REQUIREMENTS BELONG HERE" — because it was a list somewhere else.
//
// So this asserts the property that makes the split checkable: every field marked `required` must
// be one that EVERY writer of that message supplies. It cannot verify that mechanically, so it
// does the next best thing and states the exceptions, which forces a decision at the moment a
// field is marked rather than at the moment a seat is refused.
func TestNoConditionallyRequiredFieldIsMarkedUnconditional(t *testing.T) {
	// The fields a verb may legitimately omit, each with the verb that omits it. A field marked
	// required that appears here is the contradiction this test exists to catch.
	conditional := map[string]string{
		"feov.record.v1.Avenue.line":        "a MOVE names an existing line and carries only its new status and reason",
		"feov.record.v1.Close.successor":    "only CLOSED_WITH_REGRESSION names one; the message-level check says so",
		"feov.record.v1.Close.carried_from": "a carry restates an earlier closure; an ordinary close has none",
		"feov.record.v1.SpotCheck.ids":      "the --none form samples nothing and says why",
	}
	od := (&Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		t.Fatal("Event has no `body` oneof — a broken walk would pass this test on every field")
	}
	checked := 0
	for i := 0; i < od.Fields().Len(); i++ {
		md := od.Fields().Get(i).Message()
		if md == nil {
			continue
		}
		for j := 0; j < md.Fields().Len(); j++ {
			fd := md.Fields().Get(j)
			o, _ := proto.GetExtension(fd.Options(), E_Sql).(*Sql)
			if !o.GetRequired() {
				continue
			}
			checked++
			if why, isConditional := conditional[string(fd.FullName())]; isConditional {
				t.Errorf("%s is marked `required` and is CONDITIONAL: %s.\n\nA requirement that "+
					"depends on another field belongs in validate (or in the message's `check`), not "+
					"on the field — marked unconditional it refuses correct work at the write path AND "+
					"puts NOT NULL on the column, so the row cannot be stored either.",
					fd.FullName(), why)
			}
		}
	}
	if checked == 0 {
		t.Fatal("no required fields found — an empty traversal passes this test on every schema")
	}
}
