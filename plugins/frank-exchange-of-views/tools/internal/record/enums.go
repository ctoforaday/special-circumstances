package record

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/flags"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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
// consistent across the suite (the PASS refusal names `not_a_defect`, the red-auditor
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
// question to a later round with a stated research direction.
//
// It stays a named constant because it is the word the CLI defaults to and the seat-facing help
// reaches for, but it is no longer the DEFINITION of anything. "Does this end the gap" is
// recordpb.Closes, read off the value's own annotation. The two were the same statement while
// `carried` was the only deferring word, and that coincidence is exactly what made the negative
// rule "everything except carried" look correct right up until a second deferring word existed.
const DispositionCarried = "carried"

// Dispositions is HOW A GAP ENDED — one vocabulary for both closing verbs (#342).
//
// `merge close` (red closes on verified repair) and `motion docket rule` (the bench closes on
// judgement) are different acts with different evidence bars, and they stay different verbs.
// What they must not have is different WORDS for the same outcome: before this, a reader had
// to know which verb produced a closure before it could interpret the word, and four surfaces
// spelled the same three outcomes six ways.
//
// READ OFF THE SCHEMA, not re-typed beside it. This table used to carry all six meanings as Go
// string literals while record.proto carried the same six as `(means)` annotations — one
// statement, two files, and nothing in the build that could see them disagree. That is the exact
// defect the annotations were added to remove, reproduced one layer up from it.
var Dispositions = dispositionsWhere(func(bool) bool { return true })

// ClosureClasses is the subset a MERGE may write: the values that actually close a gap.
//
// Derived from the same annotation the schema builds its CHECK from, so the refusal a seat reads
// and the constraint the database enforces cannot admit different words.
var ClosureClasses = dispositionsWhere(func(closes bool) bool { return closes })

// DeferringDispositions is the complement of ClosureClasses: the words that do NOT end the gap.
//
// It exists so that no surface has to SAY which those are. The help text used to read "every value
// ends the gap except `carried`" — a sentence that was true when it was written, is a copy of an
// annotation that now answers the question, and would have gone quietly wrong the moment a second
// deferring word was added. That is not a hypothetical: it is what happened to the predicate this
// vocabulary replaced.
var DeferringDispositions = dispositionsWhere(func(closes bool) bool { return !closes })

// closureClassNames is the bare vocabulary, for the readers that only need the words.
func closureClassNames() []string { return Names(ClosureClasses) }

// dispositionsWhere reads the vocabulary out of the descriptor.
//
// A value that never declared `closes` PANICS at init rather than defaulting to false. That is
// deliberate and it is the whole point of the annotation: adding a word must force the question,
// and answering it here on the author's behalf is precisely how a gap the bench deferred came to
// be retired with nobody having decided to retire it. An init panic is loud, immediate, and
// impossible to ship past; a default is none of those.
func dispositionsWhere(keep func(closes bool) bool) []EnumValue {
	ed := recordpb.Disposition(0).Descriptor()
	var out []EnumValue
	for i := 0; i < ed.Values().Len(); i++ {
		vd := ed.Values().Get(i)
		word := recordpb.Word(recordpb.Disposition(vd.Number()))
		if word == "" {
			continue // the UNSPECIFIED zero is absence, not a choice a seat makes
		}
		closes, declared, err := recordpb.Facet(vd, "closes")
		if err != nil || !declared {
			panic(fmt.Sprintf("record: disposition %s does not declare whether it closes the gap — "+
				"a value added without answering that reads as closing BY DEFAULT, which is how "+
				"`grade_adjusted` retired a gap the bench had explicitly deferred", vd.Name()))
		}
		if !keep(closes) {
			continue
		}
		means, err := recordpb.EnumValueDoc(vd)
		if err != nil {
			panic(fmt.Sprintf("record: disposition %s carries no meaning, so the set would reach a "+
				"seat as a bare noun: %v", vd.Name(), err))
		}
		out = append(out, ev(word, means))
	}
	return out
}

// ArtifactState is the SECOND AXIS a closure carries, and for a long time nothing could read it.
//
// A disposition answers "is this still contested". It does NOT answer "does the report still
// carry a known defect", and those diverge: three of the six classes settle the dispute while
// leaving a real defect in the shipped artifact. `open`/`closed` is the DOCKET axis, and every
// consumer that read it as the artifact axis was wrong.
//
// MEASURED 2026-08-22, and it is the reason this exists. On the sqlite-schema run, at the same
// moment, for the same gap R1-1:
//
//	the board          open: 0 — nothing outstanding
//	assembly-screen    FAIL — "1 source(s) red found AGAINST are still cited in the report"
//
// Two gates, one run, flatly contradicting each other in English. Both were right under their
// own definition, and only one of them knew there were two definitions.
//
// IT IS DERIVED, NOT STORED, and that is deliberate. The mapping is total, so a written field
// would be a second hand-kept copy of a fact the disposition already determines — two writers of
// one fact, which is the drift this codebase keeps finding. [[facts-are-fields]] says it
// outright: prefer GENERATING the derived carrier over guarding two copies of it. A derivation
// also cannot be FORGOTTEN, which a new required field demonstrably can — the friction channel
// went unclosed in eighteen consecutive sittings until its empty case became assertable.
//
// The claim stays falsifiable: assembly-screen detects a live defect in the assembled report
// from the artifact itself, so a wrong derivation here is catchable rather than merely asserted.
type ArtifactState string

const (
	// ArtifactRepaired: the defect is gone and the repair was verified.
	ArtifactRepaired ArtifactState = "repaired"
	// ArtifactNoDefect: there was never a defect to fix — red was wrong.
	ArtifactNoDefect ArtifactState = "no_defect"
	// ArtifactDefectLive: the defect is REAL and the report ships carrying it. The dispute is
	// over; the problem is not.
	ArtifactDefectLive ArtifactState = "defect_live"
	// ArtifactUnexamined: nobody reached the merits, so nothing is known either way. It must not
	// collapse into "no defect" — an unasked question and an answered one are not the same.
	ArtifactUnexamined ArtifactState = "unexamined"
	// ArtifactUnknown: the class is not a word this vocabulary carries. NO RECORD CAN PRODUCE
	// IT — the schema's CHECK is generated from the same enum this map is checked against, and
	// TestArtifactStateCoversEveryDisposition fails the moment the map falls behind that enum.
	// It survives for input that reached here without passing the schema, where reporting the
	// miss as itself beats folding it into a healthy value.
	ArtifactUnknown ArtifactState = "unknown"
)

// artifactByClass is the total part of the mapping. `amends_prior` is absent on purpose: it
// inherits from the ruling it amends and needs a lookup, not a table row.
var artifactByClass = map[string]ArtifactState{
	"repaired":                 ArtifactRepaired,
	"repaired_with_regression": ArtifactDefectLive, // repaired here, and a live successor carries the remainder
	"not_a_defect":             ArtifactNoDefect,
	"defect_accepted":          ArtifactDefectLive,
	"defect_owed_elsewhere":    ArtifactDefectLive,
	DispositionCarried:         ArtifactUnexamined, // still live; the question is open, not answered
}

// ArtifactStateOf answers what a closure class says about the ARTIFACT.
//
// `amends_prior` returns ok=false: it is defined relative to an earlier ruling, so this function
// cannot answer alone and says so rather than guessing. The caller walks the lineage — and a
// caller that cannot must report the miss, not substitute a plausible value.
func ArtifactStateOf(class string) (ArtifactState, bool) {
	if class == "amends_prior" {
		return "", false
	}
	if s, ok := artifactByClass[class]; ok {
		return s, true
	}
	return ArtifactUnknown, true
}

type EnumField struct {
	Key  string // the payload key the value lands in
	Flag string // the flag a seat types — NOT derived: payload keys are not globally
	// unique, and flags.ForPayloadKey says so itself.
	//
	// Values carry their MEANINGS, not only their spellings: a set rendered as six words and
	// one shared sentence leaves a seat to guess which situation warrants which, and the
	// guessing is measurable (see enumvalue.go).
	Values []EnumValue
	Why    string // what a near-miss did before this was enforced; the seat reads it

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
		Key: "verdict", Flag: flags.As, Values: []EnumValue{
			ev("PASS", "every gap on the board is resolved — this is CHECKED against the open board, not taken on your word"),
			ev("FAIL", "at least one gap is still open, or you are not satisfied it was answered"),
		},
		Why: "a PASS is checked against the open board by exact match, so any other spelling skips the check entirely and records an unadjudicated pass",
	}},
	"outcome": {{
		Key: "verdict", Flag: flags.As, Values: []EnumValue{
			ev("VERIFIED", "red passed the board and the bench agrees the question was answered"),
			ev("CEILING", "the round ceiling was reached with work still open — NOT a judged failure to verify, and the stamp says so"),
			ev("HALTED", "the bench ended the run on a safety, ethics, consent or integrity boundary"),
			ev("UNVERIFIED", "the run ended without the question being answered, and no ceiling or halt explains it"),
		},
		Why: "the report's verdict stamp switches on this word — an unrecognized one falls through to a bare stamp, so a lowercase CEILING loses the \"this is NOT a judged failure to verify\" caveat the stamp exists to carry",
	}, {
		Key: "ended", Flag: flags.Ended, Optional: true, Values: []EnumValue{
			ev("deadlock", "the bench JUDGED the exchange deadlocked — the one terminal state the record cannot derive, so --reason is the only account of it there will ever be"),
			ev("ceiling", "the run stopped against its safety or round ceiling rather than against a judgement"),
		},
		// A SWITCH OVER TWO BOOLEANS IS AN ENUM WITH A SILENT FOURTH STATE. `--deadlocked` and
		// `--exhausted` were separate flags stored as separate fields and read back in a
		// first-match switch by every consumer, so setting both recorded a contradiction the
		// tool accepted and the reader resolved by argument order.
		Why: "the verdict stamp reads this to say HOW a non-pass ended; an unrecognized word decorates the stamp with nothing, which reads exactly like a run that ended for no stated reason",
	}},
	// `log`, and the TYPE is the whole reason the channel is worth reading. Measured on
	// 2026-09-02_quadratic-formula the channel ran 142,891 characters with no types on it, so an
	// operator had to READ all of it to learn which entries were actionable — 46% were, 48% was
	// mandated ceremony, and nothing on the entry said which was which.
	"log": {{
		Key: "type", Flag: flags.Type, Values: []EnumValue{
			ev("nominal", "the surface met the work — the clean sitting, said in the POSITIVE, because an entry saying nothing is still an entry and silence cannot say it"),
			ev("defect", "something is broken: it did the wrong thing, or failed where it should have worked"),
			ev("request", "a capability that does not exist — the act you wanted was on no surface, so there was nothing to get wrong"),
			ev("friction", "the work was impeded and you are NOTING it; NOT necessarily actionable and not necessarily advisable to change, which is why it has its own word rather than posing as a defect"),
			ev("estoppel", "the TOOL refused a mint against text the other side prescribed — recorded by the tool, not filed by a seat"),
		},
		Why: "the operator triages this channel by FILTERING on the type; an untyped entry hands the reader back the reading this field exists to replace, so the write refuses one",
	}},
	// `finding`, and the ABOUT is the anchor a quote could not provide. Measured: a missing line
	// of inquiry pinned to a sentence the finding called fine, and a missing risk matrix pinned to
	// a section opening, because a live quote was the only target the verb accepted.
	"finding": {{
		Key: "about_kind", Flag: flags.AboutKind, Optional: true, Values: []EnumValue{
			ev("section", "a named report section, for something MISSING from it — the anchor a quote cannot give, because the text you object to is not there"),
			ev("inquiry", "a line of inquiry, by its avenue id: an argument against the REASON it was declined, deferred or abandoned"),
			ev("gap", "a gap already on the docket, by its id — a defect in the record rather than in the report"),
		},
		Why: "an absence has no sentence to quote, so it used to borrow an innocent one as a handle and the gap list pointed a reader at good prose. These targets are references the record can CHECK: an avenue id either names a line this run proposed or it does not",
	}},
	// `cite`, and the READING is the half a citation could not previously state. A citation says
	// a source backs a sentence; it never said whether anyone had read the source. That gap was
	// the load-bearing caveat of a whole run, carried as the prose substring "unreachable from
	// this container" and miscounted twice by grep.
	"cite": {{
		Key: "source_text_read", Flag: flags.SourceText, Optional: true, Values: []EnumValue{
			ev("leaf", "the source's own text, in the bytes this run cached — the only reading that licenses a claim about what it SAYS"),
			ev("summary_only", "someone else's account of it: an abstract, a secondary description, or the summary of an INTERESTED party"),
			ev("unread", "never read — the citation rests on a record that the source EXISTS, not on anything it says"),
		},
		Why: "a claim about what a source SAYS rests on having read it; without this the report cannot distinguish a source read at the leaf from one known only through the summary of the party whose case depends on it",
	}},
	// `avenue`, the schema's word. It was "line-of-inquiry" — an event type the schema does not
	// declare — so this whole set was advertised against a body that does not exist, and nothing
	// noticed because the key was only ever looked up by the same stale name.
	"avenue": {{
		Key: "status", Flag: flags.As, Values: InquiryStatuses,
		Why: "the lines-of-inquiry projection groups BY status, so a status outside the set does not fail — it silently vanishes from the section that exists to show the roads not taken",
	}},
	// `inquiry-review` HAS NO ENTRY BECAUSE IT HAS NO CLOSED SET, and the absence is the ruling.
	// Its predecessor `inquiry-support` carried a four-value `--as` (supported / weakened /
	// unsupported / absent) answering "does the report still carry this line". Presence is not a
	// question — the lines are generated onto the page from the record, so blue cannot cut them —
	// and the surviving question, whether the body delivered the research, is an ORDINARY GAP with
	// the grade vocabulary it already has. The review carries prose and nothing else.
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
		Why:      "the class is HOW the gap ended, and every downstream reader interprets it — the closure index, the repair_regression denominator, and the successor invariant that fires on repaired_with_regression alone. An unrecognized class lands in no bucket and the gap reads as closed for no stated reason",
	}},
	"mint": {
		{
			Key: "check_kind", Flag: flags.CheckKind, Values: []EnumValue{
				ev("document", "reading a shipped artifact settles it — the check is answered by prose that quotes what is there"),
				ev("computation", "RUNNING something settles it. This check CANNOT be closed by prose: it closes only when a proof answers the gap. Reach for it wherever the answer would be PRODUCED rather than asserted — arithmetic, a simulation, a forecast, a parse, a count, a re-derivation are common cases and not the whole of it; if you can imagine a script that would end the argument, this is the kind"),
				ev("source", "verifying an external source settles it — the claim stands or falls on what the cited material actually says"),
			},
			Why: "the kind says WHAT WOULD SETTLE the acceptance check, and it is the lever the 2026-08-05 smoke measured missing: blue wrote zero programs across the run, not because it ignored the invitation but because NOTHING ASKED — all ten of red's checks were document probes, and R1-1 was literally \"execute the assembly step\". Red could only ever ask whether the report SAYS something. A `computation` check is a demand that cannot be answered in prose",
		},
	},
	"reproduce": {{
		Key: "soundness", Flag: flags.As, Values: []EnumValue{
			ev("sound", "you READ the script and it computes what it claims to compute"),
			ev("unsound", "it re-runs cleanly and establishes nothing, or something other than the claim it is anchored to — the dangerous cell, because it looks maximally credible"),
		},
		Why: "REPRODUCING IS NOT PROVING. Re-running a script and getting the same bytes measures DETERMINISM; `print(\"7 is prime\")` reproduces perfectly forever. Whether the script actually establishes the claim it is anchored to cannot be computed — red must READ it — so it is judged, and it is required. The dangerous cell is reproduces+unsound: a proof that looks maximally credible and establishes nothing",
	}},
	"verify": {{
		Key: "outcome", Flag: flags.As, Values: []EnumValue{
			ev("supports", "you read the source at the leaf and it says what the claim says"),
			// UNDERSCORE, matching the schema. It was `supports-with-bridge` — the only hyphenated value in
			// any set — so `--help` offered a word `SourceOutcomeOf` then refused: "not a source outcome
			// this record can carry", for the value the tool had just told the seat to use. That is #342
			// in miniature, and it survived because nothing compared the advertised set to the schema's.
			ev("supports_with_bridge", "it supports the claim but you had to bridge something — a summary, a secondary citation, a near-restatement"),
			ev("weak", "it gestures at the claim, or is itself uncorroborated: thin support, not none"),
			ev("refutes", "you read the source and it CONTRADICTS the claim — the strongest finding this verb can carry, and until 0.60.0 it had no field at all"),
			ev("absent", "you read the source and the claim is simply not in it. Distinct from `refutes`: silence is not contradiction, and a reader deciding what to do about it needs to know which it was"),
			ev("unreachable", "you could not read it — paywall, dead link, a format you could not extract. Say what you tried in --reason; an untried \"unable to corroborate\" is an incomplete audit"),
		},
		Why: "THE NEGATIVE HALF, WHICH DID NOT EXIST. Red could say how a citation held and had no way whatever to record that it did NOT — so the strongest adversarial finding available on this axis had to leave as prose, and the capture audit built to catch a report shipping a refuted citation went looking for a verdict no field could carry: it reported PASS over an empty file on every record-mode run (#296). This is WHAT THE SOURCE DID, and it is a different question from how sure you are of it, which is --confidence",
	}, {
		Key: "confidence", Flag: flags.Confidence, Values: []EnumValue{
			ev("high", "you read the source at the leaf and would defend this determination as it stands"),
			ev("medium", "you are reasonably sure, but the reading bridges something — a summary, a secondary source, a near-restatement rather than the exact statement"),
			ev("low", "your reading may be wrong: an ambiguous passage, thin evidence, or a source you could only partly read. This is a call for more evidence, NOT an automatic fail — blue digs further"),
		},
		Why: "CONFIDENCE IS IN THE DETERMINATION, WHATEVER THE DETERMINATION WAS. It is orthogonal to --outcome and always has been: `refutes` at low confidence (this source may contradict the claim, I am not certain) and `refutes` at high confidence (I read it, it says the opposite) are different facts, and a reader who cannot tell them apart cannot decide what to do about either.\n\nThe original plan specified exactly this — \"for each statement ↔ reference pair it assigns a confidence that the source actually corroborates the statement (facts are rarely black and white); low confidence → needs more evidence, blue digs further, not an automatic fail\" — and this field is what shipped from it.\n\nIt spent time called `--trust`, a rename made in #341 to dodge a collision with `blue confidence` (one word carrying two questions). That verb was DELETED in 0.54.0, so the collision has not existed for six releases while the dodge did — and the substitute word invited its own misreading: `trust` sounds like a property of the SOURCE, so its own value descriptions drifted into a support scale (\"the source supports the claim but you had to bridge something\"), and the axis read as a positive-only outcome. It is not one; it is how sure you are",
	}},
}

// Usage renders the flag's help from the set itself, so the contract a seat reads is the
// contract the write path enforces.
func (e EnumField) Usage(what string) string {
	return strings.Join(Names(e.Values), " | ") + " — " + what
}

// Spelling is the set as a verb summary writes it: PASS|FAIL, no spaces.
func (e EnumField) Spelling() string { return strings.Join(Names(e.Values), "|") }

// allows reports whether v is in the set. Exact and case-sensitive by construction: the
// gates downstream compare literally, so anything looser here would re-open the hole one
// layer down.
func (e EnumField) allows(v string) bool {
	for _, want := range e.Values {
		if v == want.Name {
			return true
		}
	}
	return false
}

// enum returns one declared set by (event type, payload key), for the CLI's help.
func enum(typ, key string) (EnumField, bool) {
	for _, e := range EnumFields[typ] {
		if e.Key == key {
			return e, true
		}
	}
	return EnumField{}, false
}

// MustEnum is enum for package-level flag registration, where a missing entry is a
// programming error rather than a runtime condition — and a silently empty help string
// would be the very defect this table exists to remove.
func MustEnum(typ, key string) EnumField {
	e, ok := enum(typ, key)
	if !ok {
		panic("record: no declared enum for " + typ + "." + key)
	}
	return e
}

// sameWord reports whether two spellings differ only in case or in their separators —
// the typo class, and nothing wider. `closed-with-regression` and `Closed_With_Regression`
// are the same word; `closed` and `repaired_with_regression` are not.
func sameWord(a, b string) bool {
	strip := func(s string) string {
		return strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(s))
	}
	return strip(a) == strip(b)
}

// checkOpenSets refuses a value outside the ONE set the schema cannot close.
//
// MOST OF WHAT checkEnum POLICED IS NOW UNREPRESENTABLE. verdict, outcome's verdict, avenue
// status, closure_class, check_kind, soundness, verify's outcome and confidence are closed proto
// enums: a value outside the set cannot be built, let alone written, so a runtime check for it
// would be dead code asserting the type system works.
//
// `Opinion.disposition` WAS THE SECOND, AND THE REASON GIVEN FOR IT DID NOT SURVIVE READING.
//
// It was listed here as "kept open on the operator's decision (plan §II.3): closing it means a
// legitimate bench ruling fails HARD mid-round, and a bench that cannot rule is worse than a
// vocabulary that drifts." Two things were wrong with that. The cited section is in no plan in
// `plans/`. And the behaviour it described as the cost of closing the set was what this function
// ALREADY DID: the arm below refused any word outside `benchDispositions`, from record.go:1131, on
// the write path a bench actually uses. The set was closed. Only its DECLARATION was loose, one
// file from the field, where the schema could not read it — so the DDL could not build a foreign
// key, the vocabulary table had no row for `carried`, and "does this word close the gap" had to be
// answered by a hand-written predicate that guessed.
//
// The drift it was meant to tolerate happened anyway, and could not be seen: the engine and the
// bench's own constitution instructed three dispositions this arm refused.
//
// So one set remains, and it is genuinely open:
//
//   - `Outcome.ended` — how the sitting ended.
//
// The near-miss is called out BY NAME when the value differs only in case, because that is the
// failure that was actually measured (`--as pass`, `--as Pass`) and "PASS | FAIL" alone does not
// tell a seat that its lowercase spelling was the whole problem.
func checkOpenSets(body proto.Message) error {
	switch b := body.(type) {
	case *recordpb.Outcome:
		if b.Ended == nil {
			return nil // absent is legal; `ended` is optional
		}
		return checkWord("ended", flags.Ended, b.GetEnded(), endedValues())
	}
	return nil
}

// checkWord is the refusal itself: the value, the set that would have worked, and the consequence.
func checkWord(key, flag, got string, allowed []EnumValue) error {
	for _, want := range allowed {
		if got == want.Name {
			return nil
		}
	}
	detail := ""
	for _, want := range allowed {
		if strings.EqualFold(got, want.Name) {
			detail = fmt.Sprintf("%s differs from %s only in case, and ", jsonish(got), jsonish(want.Name))
		}
	}
	if got == "" {
		detail = "nothing was passed, and "
	}
	names := make([]string, 0, len(allowed))
	for _, want := range allowed {
		names = append(names, want.Name)
	}
	return fmt.Errorf("record: --%s must be one of %s (got %s) — %sthe word is what every later reader switches on",
		flag, strings.Join(names, "|"), jsonish(got), detail)
}

// endedValues is the `ended` set, read off the one declaration of it rather than re-typed.
func endedValues() []EnumValue {
	for _, ef := range EnumFields["outcome"] {
		if ef.Key == "ended" {
			return ef.Values
		}
	}
	return nil
}
