package recordpb

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// WHAT EACH ENUM VALUE MEANS, DECLARED ONCE — and the SET is generated, only the prose is written.
//
// This replaces record.EnumValue, whose whole point was that a set rendered as six bare words
// leaves a seat guessing which situation warrants which. That prose is not decoration: it is
// generated into the CLI's --help AND into the refusal a seat reads when it gets one wrong, and
// enums.go was explicit that the help must not be allowed to drift from the check because it is
// BUILT from it. Proto enums carry no descriptions, so the prose lives here.
//
// WHY THIS IS NOT THE HAND-KEPT ALLOWLIST facts-are-fields WARNS ABOUT. That rule's objection is
// to a guard whose own list is maintained by hand — it reproduces the defect one level up. Here
// the authoritative set is the generated descriptor; this map supplies only the sentence. An
// exhaustiveness test walks the descriptor and fails when any value lacks prose, and fails again
// when this map names a value that no longer exists. Neither direction can rot silently, which
// is the property a bare list cannot offer.
//
// Rejected: custom proto options. They would put the prose in the .proto, which reads well, at
// the price of descriptor plumbing and a protoc-gen-go extension to read it back — a build
// dependency for a string table.
var enumValueDoc = map[protoreflect.FullName]string{
	// Verdict — red's binary gate.
	"feov.record.v1.VERDICT_PASS": "every gap on the board is resolved — this is CHECKED against the open board, not taken on your word",
	"feov.record.v1.VERDICT_FAIL": "at least one gap is still open, or you are not satisfied it was answered",

	// RunOutcome — how the sitting ended.
	"feov.record.v1.RUN_OUTCOME_VERIFIED":   "red passed the board and the bench agrees the question was answered",
	"feov.record.v1.RUN_OUTCOME_CEILING":    "the round ceiling was reached with work still open — NOT a judged failure to verify, and the stamp says so",
	"feov.record.v1.RUN_OUTCOME_HALTED":     "the bench ended the run on a safety, ethics, consent or integrity boundary",
	"feov.record.v1.RUN_OUTCOME_UNVERIFIED": "the run ended without the question being answered, and no ceiling or halt explains it",

	// CheckKind — what would SETTLE an acceptance check.
	"feov.record.v1.CHECK_KIND_DOCUMENT":    "reading a shipped artifact settles it — the check is answered by prose that quotes what is there",
	"feov.record.v1.CHECK_KIND_COMPUTATION": "RUNNING something settles it. This check CANNOT be closed by prose: it closes only when a proof answers the gap. Reach for it wherever the answer would be PRODUCED rather than asserted — arithmetic, a simulation, a forecast, a parse, a count, a re-derivation are common cases and not the whole of it; if you can imagine a script that would end the argument, this is the kind",
	"feov.record.v1.CHECK_KIND_SOURCE":      "verifying an external source settles it — the claim stands or falls on what the cited material actually says",

	// ClosureClass — how a gap ended.
	"feov.record.v1.CLOSURE_CLASS_CLOSED":                   "the repair was verified at the leaf and nothing regressed",
	"feov.record.v1.CLOSURE_CLASS_CLOSED_WITH_REGRESSION":   "repaired, but something else broke — REQUIRES a successor naming the gap that carries the regression forward",
	"feov.record.v1.CLOSURE_CLASS_AMENDS_PRIOR":             "a defect found BETWEEN two repairs that each closed clean earlier — REQUIRES supersedes so the lineage is explicit",
	"feov.record.v1.CLOSURE_CLASS_REBUTTAL_SUSTAINED":       "blue argued the finding was wrong and the argument held; nothing was repaired because nothing needed to be",
	"feov.record.v1.CLOSURE_CLASS_RISK_ACCEPTED":            "the fix costs more than the defect (complexity above likelihood x impact) and the risk is taken KNOWINGLY, with the argument on the record",
	"feov.record.v1.CLOSURE_CLASS_ROUTED_TO_INFRASTRUCTURE": "a real defect whose fix is owned outside this debate; it leaves here and is not silently dropped",

	// SourceOutcome — what the source DID for the claim.
	"feov.record.v1.SOURCE_OUTCOME_SUPPORTS":             "you read the source at the leaf and it says what the claim says",
	"feov.record.v1.SOURCE_OUTCOME_SUPPORTS_WITH_BRIDGE": "it supports the claim but you had to bridge something — a summary, a secondary citation, a near-restatement",
	"feov.record.v1.SOURCE_OUTCOME_WEAK":                 "it gestures at the claim, or is itself uncorroborated: thin support, not none",
	"feov.record.v1.SOURCE_OUTCOME_REFUTES":              "you read the source and it CONTRADICTS the claim — the strongest finding this verb can carry, and until 0.60.0 it had no field at all",
	"feov.record.v1.SOURCE_OUTCOME_ABSENT":               "you read the source and the claim is simply not in it. Distinct from `refutes`: silence is not contradiction, and a reader deciding what to do about it needs to know which it was",
	"feov.record.v1.SOURCE_OUTCOME_UNREACHABLE":          "you could not read it — paywall, dead link, a format you could not extract. Say what you tried in --reason; an untried \"unable to corroborate\" is an incomplete audit",

	// Confidence — how sure you are OF THE DETERMINATION, whatever it was.
	"feov.record.v1.CONFIDENCE_HIGH":   "you read the source at the leaf and would defend this determination as it stands",
	"feov.record.v1.CONFIDENCE_MEDIUM": "you are reasonably sure, but the reading bridges something — a summary, a secondary source, a near-restatement rather than the exact statement",
	"feov.record.v1.CONFIDENCE_LOW":    "your reading may be wrong: an ambiguous passage, thin evidence, or a source you could only partly read. This is a call for more evidence, NOT an automatic fail — blue digs further",

	// Soundness — reproducing is not proving.
	"feov.record.v1.SOUNDNESS_SOUND": "you READ the script and it computes what it claims to compute",

	// The three states of a line of inquiry against the CURRENT report. Two binary facts —
	// present, and backed — with the impossible fourth combination left out.
	"feov.record.v1.INQUIRY_STATE_CARRIED": "the report still carries this line and the text backs it as stated — nothing is owed",
	"feov.record.v1.INQUIRY_STATE_HOLLOW":  "the line is still in the report and the text NO LONGER backs it — blue owes a repair or a rebuttal, and saying so is not the same as saying it was cut",
	"feov.record.v1.INQUIRY_STATE_CUT":     "the line is not in the report at all — the document has dropped a direction it claims to have taken, so its account of its own research is false until the line is restored or withdrawn",
	"feov.record.v1.SOUNDNESS_UNSOUND":     "it re-runs cleanly and establishes nothing, or something other than the claim it is anchored to — the dangerous cell, because it looks maximally credible",

	// AvenueStatus — a line of inquiry's fate.
	"feov.record.v1.AVENUE_STATUS_PROPOSED":  "you intend to follow this line; the tool assigns it an id and red may rule on it",
	"feov.record.v1.AVENUE_STATUS_PURSUED":   "you took the line — what it produced belongs in the report",
	"feov.record.v1.AVENUE_STATUS_DEFERRED":  "not this run. REQUIRES a reason saying what a later run should pick it up FOR: a deferral with no stated reason is indistinguishable from forgetting, and this status exists precisely to be read by a run that has not happened yet",
	"feov.record.v1.AVENUE_STATUS_DECLINED":  "you considered it and chose not to. REQUIRES a reason — the road not taken is worthless without why",
	"feov.record.v1.AVENUE_STATUS_ABANDONED": "you started and stopped. REQUIRES a reason — what killed it is the part a future run actually needs",

	// Grade — the canonical set.
	"feov.record.v1.GRADE_TRIVIAL":     "cosmetic; nothing downstream changes if it is wrong",
	"feov.record.v1.GRADE_LOW":         "minor",
	"feov.record.v1.GRADE_LOW_MEDIUM":  "between minor and material",
	"feov.record.v1.GRADE_MEDIUM":      "material",
	"feov.record.v1.GRADE_MEDIUM_HIGH": "between material and serious",
	"feov.record.v1.GRADE_HIGH":        "serious",
	"feov.record.v1.GRADE_CERTAIN":     "the top of the scale — for LIKELIHOOD, reserve it for a consequence that is itself certain, never for a defect you merely verified exists",
	"feov.record.v1.GRADE_REALIZED":    "it has already happened. Contributes ZERO mass by design: mass forecasts what is still to come, and a realized defect is measured by its damage instead",

	// MotionSubject and the per-subject rulings.
	"feov.record.v1.MOTION_SUBJECT_GRADE":     "you contest a gap's grade on one dimension",
	"feov.record.v1.MOTION_SUBJECT_PETITION":  "you ask the bench to intervene — the constitutional short-circuit available to any party seat",
	"feov.record.v1.MOTION_SUBJECT_DIRECTION": "a ruling on a line of inquiry blue proposed; the id is the AVENUE's own, because the proposal IS the filing",

	"feov.record.v1.GRADE_RULING_ACCEPTED": "the proposed grade stands",
	"feov.record.v1.GRADE_RULING_REJECTED": "the grade on the board stands",

	"feov.record.v1.PETITION_RULING_GRANTED": "the relief asked for is ordered",
	"feov.record.v1.PETITION_RULING_DENIED":  "the petition fails; the run continues as it was",

	"feov.record.v1.DIRECTION_RULING_ENDORSED":     "worth this run's time — pursue it",
	"feov.record.v1.DIRECTION_RULING_OUT_OF_SCOPE": "a real question, but not THIS question",
	"feov.record.v1.DIRECTION_RULING_TOO_THIN":     "in scope, but the hypothesis does not carry its budget",

	"feov.record.v1.RULING_BINDS_ALL":   "every seat from here is bound by this ruling",
	"feov.record.v1.RULING_BINDS_FILER": "only the filing seat is bound",
	"feov.record.v1.RULING_BINDS_NONE":  "advisory — the ruling is on the record and obliges nobody",

	// FrictionKind — why a friction event exists. A FIELD, not inferred from the wording (#283).
	"feov.record.v1.FRICTION_KIND_ESTOPPEL": "the TOOL refused a mint because the defect lives in text blue applied verbatim from red's own --fix-new. Recorded by the tool, not filed by the seat: argue it on the original gap, or mint with --supersedes so the lineage is explicit",

	"feov.record.v1.FRICTION_KIND_TOOL_ERROR": "the TOOL failed internally — unparseable input, an undecodable row, a check that could not run. Recorded rather than printed or swallowed, because an error nobody learns about is one nothing improves on. Distinct from a seat's own friction so the counts an operator reads stay about capability gaps",

	// GradeDimension — which axis of a gap's grading is contested.
	"feov.record.v1.GRADE_DIMENSION_SEVERITY":   "how bad it is if it bites",
	"feov.record.v1.GRADE_DIMENSION_LIKELIHOOD": "how likely the CONSEQUENCE is — not how sure you are the defect exists, which is a separate axis",
	"feov.record.v1.GRADE_DIMENSION_IMPACT":     "how far the damage reaches",
	"feov.record.v1.GRADE_DIMENSION_COMPLEXITY": "what the fix costs; it is what makes risk_accepted arguable",

	// PetitionClass — what kind of intervention is asked for.
	"feov.record.v1.PETITION_CLASS_INTEGRITY": "the record or the process has been compromised",
	"feov.record.v1.PETITION_CLASS_SAFETY":    "a safety, ethics or consent boundary is in question",
	"feov.record.v1.PETITION_CLASS_PROCESS":   "the mechanics are obstructing the work",
	"feov.record.v1.PETITION_CLASS_SCOPE":     "the question being answered has drifted from the one asked",

	// EventType and SchemaVersion carry no seat-facing prose: no flag takes them. They are
	// exempted by name in the exhaustiveness test rather than given filler sentences, because a
	// filler sentence is indistinguishable from a considered one.
}

// undocumentedEnums are the enums whose values no seat ever types, so no --help renders them.
// Listed explicitly: an exemption somebody decided beats an exemption that emerges from a
// missing map entry.
var undocumentedEnums = map[protoreflect.FullName]bool{
	"feov.record.v1.EventType":     true, // stamped by the tool, never chosen
	"feov.record.v1.SchemaVersion": true, // the format discriminator
}

// EnumValueDoc returns the prose for one enum value.
//
// The miss is LOUD. A silent "" would render an empty --help line, which reads as a value with
// no meaning rather than a value whose meaning nobody wrote — and that is the failure this whole
// table exists to prevent.
func EnumValueDoc(v protoreflect.EnumValueDescriptor) (string, error) {
	if doc, ok := enumValueDoc[v.FullName()]; ok {
		return doc, nil
	}
	return "", fmt.Errorf("recordpb: no description for enum value %s — add one to descriptions.go; "+
		"a set rendered as bare words leaves a seat guessing which situation warrants which", v.FullName())
}

// Usage renders a flag's help from the enum itself, so the contract a seat reads is the contract
// the write path enforces. One declaration, two readers — which is the rule the old EnumFields
// table existed to keep and the reason it could not simply be deleted.
func Usage(e protoreflect.EnumDescriptor) (string, error) {
	var b strings.Builder
	for i := 0; i < e.Values().Len(); i++ {
		v := e.Values().Get(i)
		if isZeroValue(v) {
			continue
		}
		doc, err := EnumValueDoc(v)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "\n  %s — %s", Spelling(v), doc)
	}
	return b.String(), nil
}

// Spelling is the word a seat types: the enum value's name with its type prefix removed and
// lowercased, so CLOSURE_CLASS_RISK_ACCEPTED reads as `risk_accepted`.
func Spelling(v protoreflect.EnumValueDescriptor) string {
	prefix := enumPrefix(v.Parent().(protoreflect.EnumDescriptor))
	return strings.ToLower(strings.TrimPrefix(string(v.Name()), prefix))
}

// Names is the bare vocabulary, in declaration order, for the readers that only need the words.
func Names(e protoreflect.EnumDescriptor) []string {
	var out []string
	for i := 0; i < e.Values().Len(); i++ {
		if v := e.Values().Get(i); !isZeroValue(v) {
			out = append(out, Spelling(v))
		}
	}
	return out
}

// BySpelling resolves a seat's word back to its value, exactly and case-sensitively: the gates
// downstream compare literally, so anything looser here re-opens the hole one layer down.
func BySpelling(e protoreflect.EnumDescriptor, word string) (protoreflect.EnumValueDescriptor, bool) {
	for i := 0; i < e.Values().Len(); i++ {
		v := e.Values().Get(i)
		if !isZeroValue(v) && Spelling(v) == word {
			return v, true
		}
	}
	return nil, false
}

// SameWord reports whether two spellings differ only in case or separators — the typo class, and
// nothing wider. `closed-with-regression` and `Closed_With_Regression` are the same word;
// `closed` and `closed_with_regression` are not. Carried over from enums.go, where it exists
// because closure_class gates an invariant and a near-miss silently took the other branch.
func SameWord(a, b string) bool {
	strip := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(s))
	}
	return strip(a) == strip(b)
}

// NearMiss finds a declared value that differs from the seat's word only by the typo class, so a
// refusal can name what would have worked instead of only listing the set.
func NearMiss(e protoreflect.EnumDescriptor, word string) (string, bool) {
	for _, n := range Names(e) {
		if n != word && SameWord(n, word) {
			return n, true
		}
	}
	return "", false
}

// enumPrefix derives the SCREAMING_SNAKE prefix protoc-gen-go gives an enum's values.
func enumPrefix(e protoreflect.EnumDescriptor) string {
	var b strings.Builder
	name := string(e.Name())
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String()) + "_"
}

// isZeroValue reports whether this is the enum's UNSPECIFIED zero, which names no seat-facing
// choice. Grade's zero is the exception that proves the rule — it is the `undefined` sentinel and
// is still never something a seat TYPES.
func isZeroValue(v protoreflect.EnumValueDescriptor) bool { return v.Number() == 0 }

// sortedDocKeys is used by the exhaustiveness test to report rot deterministically.
func sortedDocKeys() []string {
	out := make([]string, 0, len(enumValueDoc))
	for k := range enumValueDoc {
		out = append(out, string(k))
	}
	sort.Strings(out)
	return out
}
