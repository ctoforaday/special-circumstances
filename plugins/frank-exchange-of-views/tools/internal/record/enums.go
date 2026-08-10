package record

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
)

// WHAT EACH CLOSED-SET FLAG ACCEPTS, DECLARED ONCE.
//
// Ten flags spelled an enum in their --help and then took ANY STRING. The help was the
// only statement of the set, and it was decoration: `--as pass` recorded a PASS, `--as
// banana` recorded a banana, and every gate downstream compares literally — so a
// near-miss never failed, it took the OTHER branch, silently. Measured on the shipped
// tree, one case each:
//
//	merge verdict --as PASS    -> refused, 1 gap still open   <- the gate working
//	merge verdict --as pass    -> RECORDED, gate never ran
//	merge verdict --as banana  -> RECORDED
//
// The first pass at this fixed the five `--as` flags and stopped there, which is the
// same defect one level up: the class is "a flag whose help spells a set", not "the flag
// named --as". Sweeping by the SHAPE of the help string found five more — --kind,
// --confidence twice, --dimension, --petition-class — every one of them unenforced.
// internal/cli's sweep test now keys on that shape, so the next one cannot be missed the
// way these were.
//
// This table is the single declaration of the sets. `validate` enforces it at the one
// write path — so no route reaches the log around it — and the CLI reads it to BUILD the
// usage string, which is the contract a seat is told to read (`your contract is each
// verb's own --help`). Two hand-written statements of one rule is the mistake this
// codebase keeps making; here the help cannot drift from the check because it is
// generated from it.
//
// NOT EVERY SET IS HERE, and that is deliberate:
//
// BOTH CLOSURE SETS ARE NOW CLOSED (#342). This comment used to explain why they were not:
// "`merge close`'s closure_class is likewise open, and its candidate values are not yet
// consistent across the suite (the PASS refusal names `rebuttal_sustained`, the red-auditor
// prompt names `evidence-rebutted`). Closing it before that is resolved would refuse honest
// closures."
//
// That inconsistency is what #342 resolved, and it was worse than the note recorded — FOUR
// vocabularies for one concept: the record's close classes, the bench's dispositions, the
// envelope's `class` enum (which declared `rebuttal_accepted` and `risk_argued`, words no
// other surface used), and the prose in prompts (`evidence-rebutted`, `risk-accepted`). One
// concept, four spellings, and no mechanism could see them disagree because every set was
// open.
// DispositionCarried is the ONE bench disposition that does not end a gap: it defers the
// question to a later round with a stated research direction. Every other disposition is a
// ClosureClass.
const DispositionCarried = "carried"

// ClosureClasses is HOW A GAP ENDED — one vocabulary for both closing verbs (#342).
//
// `merge close` (red closes on verified repair) and `bench opinion` (the bench closes on
// judgement) are different acts with different evidence bars, and they stay different verbs.
// What they must not have is different WORDS for the same outcome: before this, a reader had
// to know which verb produced a closure before it could interpret the word, and four surfaces
// spelled the same three outcomes six ways.
var ClosureClasses = []string{
	// the repair was verified at the leaf
	"closed",
	// repaired, but something regressed — requires --successor, the one class that gates an
	// invariant and therefore the one the near-miss guard in validate protects by spelling
	"closed_with_regression",
	// a defect found BETWEEN two repairs that each closed clean earlier — requires --supersedes
	"amends_prior",
	// blue rebutted with evidence and the rebuttal was accepted. Spelled `rebuttal_accepted`
	// in the envelope and `evidence-rebutted` in prose before this
	"rebuttal_sustained",
	// the fix's complexity exceeds its likelihood x impact and the risk is taken knowingly.
	// Spelled `risk_argued` in the envelope and `risk-accepted` in prose before this
	"risk_accepted",
	// a valid finding whose fix is owned outside this debate
	"routed_to_infrastructure",
}

// benchDispositions is ClosureClasses plus the one word that does not close.
var benchDispositions = append(append([]string{}, ClosureClasses...), DispositionCarried)

type EnumField struct {
	Key    string   // the payload key the value lands in
	Flag   string   // the flag a seat types — NOT derived: payload keys are not globally
	Values []string // unique, and flags.ForPayloadKey says so itself
	Why    string   // what a near-miss did before this was enforced; the seat reads it

	// Optional means the field may be ABSENT. A present value is still policed; only
	// "not passed at all" is allowed through. Requiredness is a separate rule with a
	// separate declaration (required.go), and conflating the two here would silently
	// make several optional flags mandatory as a side effect of closing their sets.
	Optional bool
}

// EnumFields maps an event type to every closed set it carries. A LIST, not a single
// entry: `petition-rule` carries two (the ruling and the petition's class), and keying by
// verb alone is what made the first pass look complete when it covered one flag per verb.
var EnumFields = map[string][]EnumField{
	"verdict": {{
		Key: "verdict", Flag: flags.As, Values: []string{"PASS", "FAIL"},
		Why: "a PASS is checked against the open board by exact match, so any other spelling skips the check entirely and records an unadjudicated pass",
	}},
	"outcome": {{
		Key: "verdict", Flag: flags.As, Values: []string{"VERIFIED", "CEILING", "HALTED", "UNVERIFIED"},
		Why: "the report's verdict stamp switches on this word — an unrecognized one falls through to a bare stamp, so a lowercase CEILING loses the \"this is NOT a judged failure to verify\" caveat the stamp exists to carry",
	}},
	"avenue": {{
		Key: "status", Flag: flags.Status, Values: AvenueStatuses,
		Why: "the lines-of-inquiry projection groups BY status, so a status outside the set does not fail — it silently vanishes from the section that exists to show the roads not taken",
	}},
	// dispute, dispute-respond, petition, petition-rule and avenue-rule ARE ABSENT AND THAT IS
	// DELIBERATE (#344). EnumFields is checked at the WRITE, and nothing writes those types any
	// more — the verbs are gone. Their READ paths are permanent (record/compat.go), but a reader
	// does not re-validate: a record written in 2026 under a set that has since changed is still
	// the record. The motion sets that replaced them are keyed on (subject, key), which this map
	// cannot express, and live in record/motion.go.
	"close": {{
		Key: "closure_class", Flag: flags.As, Values: ClosureClasses,
		// Optional per this file's own rule: closing a SET is not the same decision as making
		// the flag REQUIRED, and conflating them here would make several flags mandatory as a
		// side effect. required.go owns requiredness.
		Optional: true,
		Why:      "the class is HOW the gap ended, and every downstream reader interprets it — the closure index, the repair_regression denominator, and the successor invariant that fires on closed_with_regression alone. An unrecognized class lands in no bucket and the gap reads as closed for no stated reason",
	}},
	"opinion": {{
		Key: "disposition", Flag: flags.As, Values: benchDispositions,
		Optional: true,
		Why:      "the bench's disposition both RULES and ends the gap; `carried` is the one value that defers instead of closing, and the replay keys the gap's whole fate on that distinction. A near-miss spelling silently carried a gap the bench meant to close, or closed one it meant to carry",
	}},
	"mint": {
		{
			Key: "check_kind", Flag: flags.CheckKind, Values: []string{"document", "computation", "source"},
			Why: "the kind says WHAT WOULD SETTLE the acceptance check, and it is the lever the 2026-08-05 smoke measured missing: blue wrote zero programs across the run, not because it ignored the invitation but because NOTHING ASKED — all ten of red's checks were document probes, and R1-1 was literally \"execute the assembly step\". Red could only ever ask whether the report SAYS something. A `computation` check is a demand that cannot be answered in prose",
		},
		{
			Key: "existence", Flag: flags.Existence, Values: []string{"verified", "suspected"},
			Why: "grading v2 SPLIT existence from consequence because v1's `certain` textual nits outweighed high-likelihood design flaws — likelihood now grades the consequence only, and this axis carries whether the defect was checked at the leaf at all. It was validated at the FLAG layer and nowhere else, so the record accepted whatever a non-flag write path put here and the record-level enum sweep never saw the field existed. A gap whose existence is unreadable is one whose grades cannot be interpreted: `medium likelihood` means something different for a defect confirmed at the leaf and one merely inferred",
			// Optional: records written before the write-path shipped carry no existence, and a
			// gap minted without it is a real (if incomplete) gap rather than a corrupt one.
			Optional: true,
		},
	},
	"reproduce": {{
		Key: "soundness", Flag: flags.As, Values: []string{"sound", "unsound"},
		Why: "REPRODUCING IS NOT PROVING. Re-running a script and getting the same bytes measures DETERMINISM; `print(\"7 is prime\")` reproduces perfectly forever. Whether the script actually establishes the claim it is anchored to cannot be computed — red must READ it — so it is judged, and it is required. The dangerous cell is reproduces+unsound: a proof that looks maximally credible and establishes nothing",
	}},
	"verify": {{
		Key: "trust", Flag: flags.Trust, Values: []string{"high", "medium", "low"},
		Why:      "the grade is the whole content of a verification's claim about its source; an unreadable one makes it incomparable with every other row in the ledger it lands in. Named `trust` because blue's `confidence` grades a CLAIM and this grades a SOURCE — one word for two questions is how the two acts came to share an event type (#341)",
		Optional: true,
	}},
	"confidence": {{
		Key: "grade", Flag: flags.Confidence, Values: []string{"high", "medium", "low"},
		Why:      "the report renders these as a confidence table meant to be read down the column, and a value outside the three renders verbatim into it — comparable-looking and not comparable",
		Optional: true,
	}},
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

// Enum returns one declared set by (event type, payload key), for the CLI's help.
func Enum(typ, key string) (EnumField, bool) {
	for _, e := range EnumFields[typ] {
		if e.Key == key {
			return e, true
		}
	}
	return EnumField{}, false
}

// MustEnum is Enum for package-level flag registration, where a missing entry is a
// programming error rather than a runtime condition — and a silently empty help string
// would be the very defect this table exists to remove.
func MustEnum(typ, key string) EnumField {
	e, ok := Enum(typ, key)
	if !ok {
		panic("record: no declared enum for " + typ + "." + key)
	}
	return e
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
	for _, e := range EnumFields[typ] {
		if e.Optional && !p.Has(e.Key) {
			continue
		}
		got := p.Str(e.Key)
		if e.Allows(got) {
			continue
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
	return nil
}
