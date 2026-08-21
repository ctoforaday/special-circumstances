package record

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
var RequiredFields = map[string][]string{
	"mint": {"acceptance_check", "check_kind", "class", "likelihood", "impact", "problem"},
	// finding's label is TOOL-assigned now (not seat-provided), so it is not listed
	// here — same as mint's gap_id, which validate requires but no flag sets. validate
	// still enforces the finding-label INVARIANT; the table lists only seat-set fields.
	"close":           {"gap_id", "reason"},
	"closing":         {"reason"},
	"regrade":         {"reason"},
	"retire":          {"claim", "reason"},
	"line-of-inquiry": {"status", "line"},
	"inquiry-support": {"inquiry_id", "as", "reason"},
	"opinion":         {"gap_id", "disposition", "principle", "tension", "review_flag", "reason"},
	"halt":            {"reason"},
	"certify":         {"reason"},
	// The run's TERMINAL act, and it carried no reasoning at all until a bench seat reached for
	// --reason and filed its absence as friction (#375). The verdict is derived; how the sitting
	// ENDED is not, and on a judged deadlock nothing else records it.
	// `verdict` was enforced ONLY in the cobra verb, in RunE — which this file's neighbour warns
	// against in the `outcome` case itself: "a requirement the CLI holds and the record does not
	// is one every other caller skips". It also meant --help could not mark --as required and the
	// refusal fired after the seat had composed the whole invocation, which is the same defect
	// `spot-check` names in its own comment. Declared here, the mechanism marks the flag and
	// validate enforces it, and the hand-rolled check in RunE is gone.
	"outcome": {"reason", "verdict"},
	// A VERIFICATION OF NOTHING WAS RECORDABLE. `lens verify` required no flag at all: the bare
	// verb printed "source verified:" and appended an event, which then counted as red's audit
	// volume. The four fields here are what makes the row mean something — WHICH citation
	// (or an explicit --independent), WHAT the source did for the claim, HOW SURE red is of that
	// (a separate question), and the reading behind the verdict.
	"verify": {"claim", "outcome", "confidence", "reason"},
	// A DUTY DISCHARGED BY NOTHING. None of these required anything, so the bare verb recorded
	// an empty event and returned success — and two of them GATE THE SITTING. Measured:
	// `blue friction` then `blue revision`, no flags, took a seat from two outstanding duties to
	// complete:true, and the friction projection then read total:1 attested:0 with text:"".
	//
	// Same class as `verify` above, which was fixed at that one instance while its siblings sat
	// untouched. `spot-check` is deliberately absent: its bare form is a documented decision.
	"friction":      {"reason"},
	"friction-none": {"reason"},
	"position":      {"reason"},
	"revision":      {"reason"},
	"manifest-row":  {"gap_id", "row"},
}

// THIS TABLE DOES NOT ENFORCE. validate still owns enforcement, field by field, with its
// own message explaining WHY each is required — and those messages are the seat's teacher,
// so they are worth more than a generic "required flag not set".
//
// Making the table a second enforcer would have been the mistake this codebase keeps
// making: two readers of one rule, drifting. It would also have broken `opinion`
// immediately, because validate requires review_flag by PRESENCE while a generic
// "present and non-empty" check rejects a legitimate `--review-flag false` — the
// falsy-value confusion that has produced three separate defects here already.
//
// Instead the table is DOCUMENTATION, and a test proves it accurate behaviourally: for
// every (verb, field) listed, a payload missing that field must be rejected by validate.
// The table cannot drift from the code without that test failing.
