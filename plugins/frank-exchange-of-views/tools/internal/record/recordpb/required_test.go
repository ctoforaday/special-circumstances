package recordpb

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoregistry"
)

// EVERY REQUIREMENT NAMES A LIVE FIELD.
//
// This is the direction that rots silently: a field renamed or removed leaves its requirement
// behind, matching nothing, and the check it encodes simply stops firing. Nothing else notices —
// the map still has an entry, the tests still pass, and the seat is no longer refused. That is
// the hand-kept-allowlist failure exactly, and the reason requiredness is keyed on the generated
// FullName rather than on a payload-key string.
func TestEveryRequirementNamesALiveField(t *testing.T) {
	fd, err := protoregistry.GlobalFiles.FindFileByPath("record.proto")
	if err != nil {
		t.Fatalf("cannot find the schema: %v — these checks would otherwise examine nothing", err)
	}
	live := map[string]bool{}
	for i := 0; i < fd.Messages().Len(); i++ {
		md := fd.Messages().Get(i)
		for j := 0; j < md.Fields().Len(); j++ {
			live[string(md.Fields().Get(j).FullName())] = true
		}
	}
	if len(live) == 0 {
		t.Fatal("zero fields found — a no-match must fail rather than read as a clean board")
	}
	for _, k := range sortedRequiredKeys() {
		if !live[k] {
			t.Errorf("required.go demands %s, which the schema does not declare — the requirement "+
				"matches nothing and has silently stopped being enforced", k)
		}
	}
}

// Every requirement carries BOTH halves: the flag a seat types and the reason it exists. A
// requirement with no flag cannot be reported; one with no `why` is the generic
// "required flag not set" that required.go was written to replace.
func TestEveryRequirementIsComplete(t *testing.T) {
	for _, k := range sortedRequiredKeys() {
		r := required[protoFullName(k)]
		if strings.TrimSpace(r.Flag) == "" {
			t.Errorf("%s has no flag — the refusal could not name what to pass", k)
		}
		if strings.TrimSpace(r.Why) == "" {
			t.Errorf("%s has no reason — a bare 'required' teaches a seat nothing, which is the "+
				"whole objection required.go raised against the CLI's own copy of this rule", k)
		}
	}
}

// The flag and the field name DIVERGE on purpose, and this asserts the divergence is real rather
// than an accident nobody noticed: prose lands in `prose`, `rationale`, `statement` and `opinion`
// depending on the verb, and the seat types `--reason` for all of them.
func TestProseFieldsAllMapToReason(t *testing.T) {
	for _, k := range []string{
		"feov.record.v1.Close.prose",
		"feov.record.v1.Opinion.rationale",
		"feov.record.v1.Certify.statement",
		"feov.record.v1.Halt.opinion",
		"feov.record.v1.Outcome.prose",
		"feov.record.v1.Verify.text",
	} {
		if got := required[protoFullName(k)].Flag; got != "reason" {
			t.Errorf("%s maps to --%s, want --reason: one word covers every verb's prose, and "+
				"the payload key differing per verb is the point", k, got)
		}
	}
}

// RequiredFieldsOf follows DECLARATION order, so a refusal lists fields the way the schema does
// rather than the way a map iterates.
func TestRequiredFieldsFollowDeclarationOrder(t *testing.T) {
	md := (&Verify{}).ProtoReflect().Descriptor()
	var got []string
	for _, fd := range RequiredFieldsOf(md) {
		got = append(got, string(fd.Name()))
	}
	want := []string{"claim", "outcome", "confidence", "text"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("RequiredFieldsOf(Verify) = %v, want %v", got, want)
	}
}
