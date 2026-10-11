package record

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHAT EACH VERB REQUIRES, DECLARED ONCE.
//
// A seat's contract is `--help`. The help was a flat alphabetical list in which `--check`,
// `--class` and `--problem` (mandatory) sat indistinguishable from `--comment`,
// `--found-by` and `--supersedes` (optional), so the only way to discover a requirement
// was to omit it and read the error — one round trip per missing flag, and wall clock is
// the constraint this project is actually fighting.
//
// Worse, requirement was expressed in two hand-written places: twelve checks here and
// fourteen in the CLI layer, with cobra's own MarkFlagRequired used exactly zero times.
// Nothing connected either to what a seat could read.
//
// So: this table is the single declaration. `validate` enforces it, and the CLI reads it
// to mark those flags REQUIRED in the help text. A drift test asserts the two agree, which
// is the check that names.go went without — and it went inert for a day as a result.
//
// KEYS ARE PAYLOAD KEYS, not flag names; flags.ForPayloadKey maps one to the other. The
// distinction is deliberate and load-bearing (see flags/names.go): a payload key is the
// event schema and a flag is a word a seat types, and they move on different schedules.
//
// ONLY UNCONDITIONAL REQUIREMENTS BELONG HERE. A rule like "repaired_with_regression
// requires --successor" or "a declined avenue requires --reason" depends on another
// field's value, so it cannot be a static annotation and stays as logic in validate. Those
// are documented in the flag's own description instead, where the condition can be stated.
// KEYS ARE PAYLOAD KEYS. The prose key differs per verb — a close stores `prose`, a
// dispute `evidence`, an opinion `rationale`, a certify `statement` — but the flag a seat
// types to fill any of them is the one word `--reason` (flags.ForPayloadKey maps each key
// back to it). The 2026-07-20 vocabulary collapse made every claim/judgment act require
// that prose: a ruling, a closure, a removal or a dispute with no stated reasoning is
// indistinguishable from a default, and the tool refuses it rather than let it through.
// RequiredFields lists the payload keys a verb must carry, DERIVED from the schema.
//
// It was a hand-written map of 20 verbs — the third copy of a fact the `(sql).required` annotation
// already carries on the field, beside recordpb's own map (deleted with this change) and the DDL's
// NOT NULL. Its own test said what a third copy costs: "the table grew and the check did not".
//
// The help is its one reader: seat.markRequired turns each key into the flag a seat types and
// marks it REQUIRED, so a verb whose schema requires a field it does not offer a flag for is
// simply not marked — which is correct, and stated where that happens. A field an ARM requires is
// marked on the verb that declares it fills that arm (seat.FillsArm), and on no other.
func RequiredFields(typ string) []RequiredField {
	body, ok := recordpb.BodyField(typ)
	if !ok {
		return nil
	}
	md := body.Message()
	var out []RequiredField
	for _, fd := range recordpb.RequiredOf(md) {
		out = append(out, RequiredField{Key: string(fd.Name()), Flag: recordpb.FlagFor(fd)})
	}
	// AND EACH ARM'S OWN, TAGGED WITH THE ARM. A docket ruling's reasoning fields are required of
	// the `docket` arm of a ruling and of no other subject's, so they are listed under that arm's
	// name and a reader marks them only on the verb that fills it.
	for _, arm := range recordpb.ArmsOf(md) {
		for _, fd := range recordpb.RequiredOf(arm.Message()) {
			out = append(out, RequiredField{Key: string(fd.Name()), Flag: recordpb.FlagFor(fd), Arm: string(arm.Name())})
		}
	}
	return out
}

// RequiredField is a required field and the word a SEAT TYPES for it — both off the one
// annotation, because they are one declaration.
//
// The flag is not derivable from the key: a ruling's `motion_id` is typed `--id`, the word every
// verb uses for the thing it is acting on, and a close's `prose` is typed `--reason`. Deriving it
// by rule (`s/_/-/`) produced `--motion-id`, which no command registers — so the contract gate
// read three verbs' requirements as invisible, and was right to.
type RequiredField struct {
	Key  string // the field the record holds
	Flag string // the word a seat types
	// Arm is the oneof arm whose message requires the field — `docket` for a docket ruling's
	// `principle` — and "" for a field the body itself requires. A body owes an arm's fields only
	// when it fills that arm.
	Arm string
}

// HasArm says whether an event type's body has a oneof arm of that name, whatever the arm holds.
// A verb declaring an arm its body does not have would be marked from nothing, silently; the
// declaration is checked against this where it is read.
func HasArm(typ, arm string) bool {
	body, ok := recordpb.BodyField(typ)
	if !ok {
		return false
	}
	md := body.Message()
	for i := 0; i < md.Oneofs().Len(); i++ {
		if od := md.Oneofs().Get(i); !od.IsSynthetic() && od.Fields().ByName(protoreflect.Name(arm)) != nil {
			return true
		}
	}
	return false
}
