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
// ONLY UNCONDITIONAL REQUIREMENTS BELONG HERE. A rule like "closed_with_regression
// requires --successor" or "a declined line of inquiry requires --reason" depends on another
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
// simply not marked — which is correct, and stated where that happens.
func RequiredFields(typ string) []RequiredField {
	md, ok := bodyDescriptor(typ)
	if !ok {
		return nil
	}
	var out []RequiredField
	for _, fd := range recordpb.RequiredOf(md) {
		out = append(out, RequiredField{Key: string(fd.Name()), Flag: recordpb.FlagFor(fd)})
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
}

// bodyDescriptor resolves an event type's WORD to the body message the `body` oneof pairs it with
// — the one place that pairing is declared, and the same one recordsql derives the schema from.
func bodyDescriptor(typ string) (protoreflect.MessageDescriptor, bool) {
	od := (&recordpb.Event{}).ProtoReflect().Descriptor().Oneofs().ByName("body")
	if od == nil {
		return nil, false
	}
	for i := 0; i < od.Fields().Len(); i++ {
		fd := od.Fields().Get(i)
		if string(fd.Name()) == typ && fd.Message() != nil {
			return fd.Message(), true
		}
	}
	return nil, false
}
