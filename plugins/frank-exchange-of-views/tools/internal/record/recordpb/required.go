package recordpb

import (
	"sort"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// WHAT EACH VERB REQUIRES, DECLARED ONCE — as an ANNOTATION, never inferred from the schema.
//
// The obvious derivation is wrong and was tried: "a required field is one the message declares
// non-`optional`". Every field in this schema IS `optional`, deliberately, so that presence stays
// expressible (canonical.go and the presence test say why — three booleans carry meaning by
// presence alone and would collapse `false` into unset). Requiredness is therefore a SEPARATE
// FACT and needs a separate declaration; reading it off optionality would silently make every
// field optional at the same moment it made presence work.
//
// REQUIREDNESS IS A PRESENCE DUTY CHECKED AT THE WRITE, not a wire property. proto3 has no
// `required` and should not: a wire-level requirement is unremovable for the life of the schema,
// whereas this list changes when the debate's discipline changes.
//
// THE MESSAGES ARE THE POINT. required.go called them "the seat's teacher" and it was right —
// each states WHY the field is required and what its absence would have done, which a generic
// "required flag not set" cannot. They are carried verbatim from validate(), because every one
// of them was written in response to something that actually went wrong.
//
// ONLY UNCONDITIONAL REQUIREMENTS BELONG HERE. "closed_with_regression requires a successor" and
// "a declined avenue requires a reason" depend on ANOTHER FIELD'S VALUE, so they cannot be a
// static annotation and stay as logic in validate. They are documented on the field itself in
// record.proto, where the condition can be stated.
type Requirement struct {
	// Flag is the word a seat types. It is NOT derived from the field name: the payload key and
	// the flag are two vocabularies that move on different schedules — a close stores `prose`, an
	// opinion `rationale`, a certify `statement`, and the flag for all three is `--reason`.
	Flag string
	// Why is what the seat reads when it is refused.
	Why string
}

// required is keyed on the generated field's full name, so a rename breaks the build's test
// rather than silently disarming a check.
var required = map[protoreflect.FullName]Requirement{
	// mint — the gap. Grades first: an ungraded gap scores ZERO and reads as harmless.
	"feov.record.v1.Mint.likelihood":       {"likelihood", "it multiplies into the gap's mass, so an absent grade is scored as ZERO and the gap reads as harmless rather than ungraded"},
	"feov.record.v1.Mint.impact":           {"impact", "it multiplies into the gap's mass, so an absent grade is scored as ZERO and the gap reads as harmless rather than ungraded"},
	"feov.record.v1.Mint.acceptance_check": {"check", "the acceptance check red will run at re-audit — the pre-agreed contract"},
	"feov.record.v1.Mint.check_kind":       {"check-kind", "document | computation | source — what would SETTLE your acceptance check. A run where every check is a document probe can never ask for a computation, and measured across six runs no seat ever wrote one"},
	"feov.record.v1.Mint.class":            {"class", "the defect's kind (or --class-new with --definition/--neighbor/--distinguisher)"},
	"feov.record.v1.Mint.problem":          {"problem", "what is wrong; --reason fills it too — a gap with no stated problem cannot be repaired or re-audited"},

	// close — the gap ends. gap_id and the argument.
	"feov.record.v1.Close.gap_id": {"id", "which gap this closes"},
	"feov.record.v1.Close.prose":  {"reason", "the closure's argument — what was verified and why it holds; the report renders it and the re-audit reads it"},

	"feov.record.v1.Closing.text": {"reason", "the closing argument for this gap — the report renders it under the gap's docket"},

	"feov.record.v1.Regrade.basis": {"reason", "grade movement is recorded with its reason"},

	"feov.record.v1.Retire.claim":  {"claim", "quote the claim as it stood — a removal nobody can identify is not on the record"},
	"feov.record.v1.Retire.reason": {"reason", "refuted, superseded, merged, out of scope — substance leaves the report ONLY with its reason recorded"},

	"feov.record.v1.Avenue.status": {"status", "the line's fate; the lines-of-inquiry projection groups by it"},
	"feov.record.v1.Avenue.line":   {"line", "what you are going to try — an unnamed avenue teaches a future run nothing"},

	// opinion — the bench rules. Every field here was in validate's "opinions, not dispositions" loop.
	"feov.record.v1.Opinion.gap_id":      {"id", "which gap is being ruled on"},
	"feov.record.v1.Opinion.disposition": {"as", "how the gap ends, or `carried` to defer it with a stated research direction"},
	"feov.record.v1.Opinion.principle":   {"principle", "the rule you are applying, stated so a later sitting can apply it too"},
	"feov.record.v1.Opinion.tension":     {"tension", "what pulls the other way — a ruling with no acknowledged counterweight is an assertion"},
	"feov.record.v1.Opinion.review_flag": {"review-flag", "why a human should, or should not, look at this"},
	"feov.record.v1.Opinion.rationale":   {"reason", "the ruling's rationale — a disposition with no stated reasoning is indistinguishable from a default"},

	"feov.record.v1.Halt.opinion":      {"reason", "the written opinion capture relays verbatim — a halt nobody can read is a stop with no stated cause"},
	"feov.record.v1.Certify.statement": {"reason", "what you would want a human to re-examine — the bench keeps no memory between runs, so this statement is its continuity"},

	"feov.record.v1.Outcome.prose": {"reason", "how this run ended, in your words — the verdict is derived from the record, but your account of the sitting is not, and on a judged deadlock it is the only evidence that determination will ever have"},

	// verify — a verification of NOTHING was recordable before these four.
	"feov.record.v1.Verify.claim":      {"claim", "the claim you checked, quoted from the report — a verification that does not name what it verified cannot be re-checked, contested, or counted"},
	"feov.record.v1.Verify.outcome":    {"as", "what the source ACTUALLY DID for the claim — the negative half is the point, and until 0.60.0 there was no field for it"},
	"feov.record.v1.Verify.confidence": {"confidence", "how sure you are of that determination, which is a DIFFERENT question from what the determination was. `refutes` you would defend and `refutes` you are unsure of are different facts"},
	"feov.record.v1.Verify.text":       {"reason", "what the source says, in your words — a verdict with no reading behind it is the assertion this verb exists to replace"},
}

// Required returns the requirement for a field, if it has one.
func Required(fd protoreflect.FieldDescriptor) (Requirement, bool) {
	r, ok := required[fd.FullName()]
	return r, ok
}

// RequiredFieldsOf lists the fields a message requires, in declaration order — so a refusal or a
// help listing follows the schema's order rather than a map's.
func RequiredFieldsOf(md protoreflect.MessageDescriptor) []protoreflect.FieldDescriptor {
	var out []protoreflect.FieldDescriptor
	for i := 0; i < md.Fields().Len(); i++ {
		if fd := md.Fields().Get(i); required[fd.FullName()].Flag != "" {
			out = append(out, fd)
		}
	}
	return out
}

// protoFullName is the test helper's cast, kept here so the map's key type stays unexported
// knowledge of this file.
func protoFullName(s string) protoreflect.FullName { return protoreflect.FullName(s) }

// sortedRequiredKeys is used by the exhaustiveness test to report rot deterministically.
func sortedRequiredKeys() []string {
	out := make([]string, 0, len(required))
	for k := range required {
		out = append(out, string(k))
	}
	sort.Strings(out)
	return out
}
