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
	"avenue-rule": {{
		Key: "ruling", Flag: flags.Ruling, Values: AvenueRulings,
		Why: "blue reads the ruling to decide whether to pursue, contest or drop a direction; an unrecognized fate reads as no ruling at all, so red's refusal of a line silently becomes permission",
	}},
	"dispose": {{
		Key: "disposition", Flag: flags.As, Values: []string{"minted-as", "folded-into", "declined", "banked"},
		Why: "a fate outside the four is a fifth meaning nobody defined, and one finding gets one fate — an unreadable one cannot be audited as either given or withheld",
	}},
	"dispute-respond": {{
		Key: "response", Flag: flags.As, Values: []string{"accepted", "rejected"},
		Why: "the orchestrator holds a dispute for a round only on an exact `rejected`; anything else falls through to the accepting branch, so a misspelt REJECTION silently applies blue's proposed grade",
	}},
	"petition-rule": {
		{
			Key: "ruling", Flag: flags.As, Values: []string{"granted", "denied"},
			Why: "relief binds the coming seats only on an exact `granted`, so a near-miss grants nothing while reading as a grant on the record — and a halt is the bench's own verb, which this set closing is what actually makes true",
		},
		{
			Key: "class", Flag: flags.PetitionClass, Values: []string{"ethical", "safety", "integrity", "constitutional"},
			Why:      "the four classes are what the bench is convened to hear; a fifth is a petition nobody defined a standard for, ruled on under whichever standard the seat happened to imagine",
			Optional: true,
		},
	},
	"petition": {{
		Key: "class", Flag: flags.PetitionClass, Values: []string{"ethical", "safety", "integrity", "constitutional"},
		Why:      "the class is what the seat is ASKING the bench to sit on, and the bench is convened per class; a fifth is a petition heard under whichever standard the ruling seat happened to imagine",
		Optional: true,
	}},
	"dispute": {{
		Key: "dimension", Flag: flags.Dimension, Values: []string{"severity", "likelihood", "impact", "complexity_cost"},
		Why:      "the orchestrator matches red's answer to blue's dispute on (gap_id, dimension) and then reads the gap's grade AT that dimension: an axis outside the four matches no answer and reads no grade, so the dispute auto-dockets and its accepted delta computes as zero",
		Optional: true,
	}},
	"mint": {{
		Key: "check_kind", Flag: flags.CheckKind, Values: []string{"document", "computation", "source"},
		Why: "the kind says WHAT WOULD SETTLE the acceptance check, and it is the lever the 2026-08-05 smoke measured missing: blue wrote zero programs across the run, not because it ignored the invitation but because NOTHING ASKED — all ten of red's checks were document probes, and R1-1 was literally \"execute the assembly step\". Red could only ever ask whether the report SAYS something. A `computation` check is a demand that cannot be answered in prose",
	}},
	"observe": {{
		Key: "kind", Flag: flags.Kind, Values: []string{"note", "checked-held"},
		Why:      "the two kinds are what an observation can BE — a note the merge may decline, or a check that was run and held. A third word exports into the findings projection as a flavour nothing downstream knows how to read",
		Optional: true,
	}},
	"cite": {{
		Key: "confidence", Flag: flags.Confidence, Values: []string{"high", "medium", "low"},
		Why:      "the grade is the whole content of a citation's claim about its source; an unreadable one makes the citation's confidence incomparable with every other row in the table it lands in",
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
