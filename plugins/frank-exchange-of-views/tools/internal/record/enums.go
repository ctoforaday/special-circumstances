package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// WHAT EACH CLOSED-SET FLAG ACCEPTS, DECLARED ONCE.
//
// Seven verbs spell an enum in their --help and then took ANY STRING. The help was the
// only statement of the set, and it was decoration: `--as pass` recorded a PASS, `--as
// banana` recorded a banana, and every gate downstream compares literally — so a
// near-miss never failed, it took the OTHER branch, silently. Measured on the shipped
// tree, one case each:
//
//	merge verdict --as PASS    -> refused, 1 gap still open   <- the gate working
//	merge verdict --as pass    -> RECORDED, gate never ran
//	merge verdict --as banana  -> RECORDED
//
// This table is the single declaration of the sets. `validate` enforces it at the one
// write path — so no route reaches the log around it — and the CLI reads it to BUILD the
// --as usage string, which is the contract a seat is told to read (`your contract is
// each verb's own --help`). Two hand-written statements of one rule is the mistake this
// codebase keeps making; here the help cannot drift from the check because it is
// generated from it, and a test asserts every entry's values are refused-when-wrong.
//
// NOT EVERY --as IS HERE, and that is deliberate:
//
//   - `bench opinion`'s disposition is an OPEN set by decision (its help ends in "...")
//     — closing it would mean a legitimate ruling failing hard mid-round. It is guarded
//     narrowly instead, in validate, for the one word that is another verb's act.
//   - `merge close`'s closure_class is likewise open, and its candidate values are not
//     yet consistent across the suite (the PASS refusal names `rebuttal_sustained`, the
//     red-auditor prompt names `evidence-rebutted`). Closing it before that is resolved
//     would refuse honest closures. It gets the same narrow guard, on the one class that
//     gates an invariant.
type EnumField struct {
	Key    string   // the payload key the value lands in
	Flag   string   // the flag a seat types — NOT derived: payload keys are not globally
	Values []string // unique, and flags.ForPayloadKey says so itself
	Why    string   // what a near-miss did before this was enforced; the seat reads it
}

var EnumFields = map[string]EnumField{
	"verdict": {
		Key: "verdict", Flag: flags.As, Values: []string{"PASS", "FAIL"},
		Why: "a PASS is checked against the open board by exact match, so any other spelling skips the check entirely and records an unadjudicated pass",
	},
	"outcome": {
		Key: "verdict", Flag: flags.As, Values: []string{"VERIFIED", "CEILING", "HALTED", "UNVERIFIED"},
		Why: "the report's verdict stamp switches on this word — an unrecognized one falls through to a bare stamp, so a lowercase CEILING loses the \"this is NOT a judged failure to verify\" caveat the stamp exists to carry",
	},
	"dispose": {
		Key: "disposition", Flag: flags.As, Values: []string{"minted-as", "folded-into", "declined", "banked"},
		Why: "a fate outside the four is a fifth meaning nobody defined, and one finding gets one fate — an unreadable one cannot be audited as either given or withheld",
	},
	"dispute-respond": {
		Key: "response", Flag: flags.As, Values: []string{"accepted", "rejected"},
		Why: "the orchestrator holds a dispute for a round only on an exact `rejected`; anything else falls through to the accepting branch, so a misspelt REJECTION silently applies blue's proposed grade",
	},
	"petition-rule": {
		Key: "ruling", Flag: flags.As, Values: []string{"granted", "denied"},
		Why: "relief binds the coming seats only on an exact `granted`, so a near-miss grants nothing while reading as a grant on the record — and a halt is the bench's own verb, which this set closing is what actually makes true",
	},
}

// Usage renders the flag's help from the set itself, so the contract a seat reads is the
// contract the write path enforces.
func (e EnumField) Usage(what string) string {
	return strings.Join(e.Values, " | ") + " — " + what
}

// Spelling is the set as a verb summary writes it: PASS|FAIL, no spaces.
func (e EnumField) Spelling() string { return strings.Join(e.Values, "|") }

// Allows reports whether v is in the set. Exact and case-sensitive by construction: the
// gates downstream compare literally, so anything looser here would re-open the hole one
// layer down.
func (e EnumField) Allows(v string) bool {
	for _, want := range e.Values {
		if v == want {
			return true
		}
	}
	return false
}

// sameWord reports whether two spellings differ only in case or in their separators —
// the typo class, and nothing wider. `closed-with-regression` and `Closed_With_Regression`
// are the same word; `closed` and `closed_with_regression` are not.
func sameWord(a, b string) bool {
	strip := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(s))
	}
	return strip(a) == strip(b)
}

// checkEnum refuses a value outside the declared set, naming what would have worked.
//
// The near-miss is called out BY NAME when the value differs only in case, because that
// is the failure that was actually measured (`--as pass`, `--as Pass`) and "PASS | FAIL"
// alone does not tell a seat that its lowercase spelling was the whole problem.
func checkEnum(typ string, p *Payload) error {
	e, ok := EnumFields[typ]
	if !ok {
		return nil
	}
	got := p.Str(e.Key)
	if e.Allows(got) {
		return nil
	}
	// The consequence (Why) is always stated: a seat that mistyped needs to know what
	// the mistype WOULD have done, not just that a set exists.
	detail := ""
	for _, want := range e.Values {
		if strings.EqualFold(got, want) {
			detail = fmt.Sprintf("%s differs from %s only in case, and ", jsonish(got), jsonish(want))
		}
	}
	if got == "" {
		detail = "nothing was passed, and "
	}
	return fmt.Errorf("record: %s requires --%s %s (got %s) — %s%s",
		typ, e.Flag, strings.Join(e.Values, "|"), jsonish(got), detail, e.Why)
}
