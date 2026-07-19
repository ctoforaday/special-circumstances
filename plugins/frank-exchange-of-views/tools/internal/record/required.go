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
// requires --successor" or "a declined avenue requires --reason" depends on another
// field's value, so it cannot be a static annotation and stays as logic in validate. Those
// are documented in the flag's own description instead, where the condition can be stated.
var RequiredFields = map[string][]string{
	"mint":    {"acceptance_check", "class", "likelihood", "impact"},
	"close":   {"gap_id"},
	"dispose": {"disposition"},
	"regrade": {"basis"},
	"retire":  {"claim", "reason"},
	"avenue":  {"status", "line"},
	"opinion": {"gap_id", "disposition", "principle", "tension", "review_flag"},
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
