// Package flags owns the CLI's vocabulary.
//
// A flag name is a WORD IN A LANGUAGE the seats learn once and reuse everywhere. Two
// meanings for one word is a defect even when each is individually sensible, because a
// seat generalises from the verb it learned first — and two words for one meaning is the
// same defect from the other side: it teaches that the concepts differ when they do not.
//
// The second half is the one that keeps growing back. "The exact span quoted out of
// blue/report.md, matched literally" was spelled four ways — --location, --claim, --old and
// --fix-old — and two of the five --location descriptions contradicted each other, one of
// them in the same file as the code that proved it wrong. It is --quote now, once.
//
// So: names are constants here, registration goes through the helpers here, and a verb
// cannot invent a private spelling without editing this file. The constant is not for the
// compiler's benefit; it is a chokepoint that makes divergence require a deliberate act.
// vocabulary_test.go closes the loop in both directions — every registered flag is declared
// here, and every declared one is registered somewhere.
//
// FLAG WORDS ARE NOT PAYLOAD KEYS. The keys are the event schema and move on their own
// schedule. Keep them free to differ: renaming a word a seat types is cheap, rewriting
// recorded history is not.
package flags

import "strings"

// Flag names. One constant per word in the vocabulary.
const (
	// Seat context, on the root as persistent flags.
	Run    = "run"
	SeatID = "seat-id"

	// Output format, on the root as a persistent flag: --json emits a structured
	// result (and structured errors) for machine consumers; the default is human text.
	JSON = "json"

	// The prose payload — one word for the reasoning a seat argues from, inline
	// (--reason) or from a file/stdin (--reason-file). Every prose argument a verb
	// records lands here: an argument, a ruling's grounds, a spot-check's findings.
	Reason     = "reason"
	ReasonFile = "reason-file"

	// Identity and reference.
	ID  = "id"
	IDs = "ids"
	Key = "key"

	// Quote is THE EXACT TEXT, quoted out of blue/report.md and nothing else, matched
	// literally. Every verb that addresses a span of the report takes it: cite and prove
	// splice an anchor there, finding flags it, mint says the defect lives there, edit
	// replaces it, retire says it left. A section heading prepended to it makes it match
	// nothing, which is why the four words this replaced could not all be right.
	Quote = "quote"
	// New is what that span should BECOME — blue applying an edit, red proposing one,
	// blue naming the claim that replaces a retired one. Paired with --quote everywhere.
	New = "new"

	// The citation axis. --url is the source a `fetch` reads, a `blue cite` anchors and a
	// `lens verify --independent` corroborates from; --title is the human name it carries
	// into the composed bibliography. A source is named by these two, everywhere.
	URL   = "url"
	Title = "title"

	Format = "format" // operator output selector
	// Sitting names WHOSE sitting a diagnostic is about, and is deliberately not --seat-id:
	// that one says who is ASKING and selects the tree the command is even on.
	Sitting = "sitting"
	// Trajectory is the captured transcript a diagnostic reads. The record says what a seat
	// WROTE; only the trajectory says what it was SHOWN.
	Trajectory = "trajectory"

	// Window sizes `show report --quote`'s sibling read `--anchor <id> --window N`: how many
	// paragraphs of content either side of the anchor a live read carries.
	//
	// The address is --anchor, never a line number: the report is edited every round, so a
	// line captured in round 2 points at different text in round 3. Line numbers come back
	// as OUTPUT, for saying what you read.
	Window = "window"

	// Answers is an act's PROVENANCE: the gap id it is a response to. A distinct word from
	// --id (which names the gap an act is ABOUT) because the relation differs.
	//
	// It exists because the link was a CONVENTION: 19 of 26 edits in one smoke happened to
	// open --reason with the gap id, and 7 did not. Prose that usually carries a fact is not
	// a join key.
	Answers = "answers"

	// A line of inquiry's ABSTRACT: what would be true if it pays off. Distinct from
	// --problem (a defect) and --reason (the argument for an act) because it is a forward
	// claim, and it is what makes a later abandonment checkable.
	Hypothesis = "hypothesis"

	// The proof axis. --script is the program that settles a claim; --cites names the METHOD
	// citation it applies — the source that says trial division decides primality. The method
	// is cited; the instance is computed.
	Script = "script"
	Cites  = "cites"

	// Disposition and justification.
	//
	// As is the one word for "which value from a closed set describes what became of this" —
	// a closure's class, a bench disposition, a verification's outcome, a ruling, a run's
	// verdict, a line of inquiry's fate. Every set is closed, and each verb documents its own
	// through enumhelp.
	As    = "as"
	None  = "none"
	Ended = "ended"

	// Confidence is how sure a seat is of a determination it just made — orthogonal to WHAT
	// the determination was (--as). One word, one question, and the question is "how sure are
	// you", never "how good is the source".
	Confidence = "confidence"

	// Grading. All four axes ARE their flag names, which is what lets a seat learn the
	// pattern from any one of them.
	Severity   = "severity"
	Likelihood = "likelihood"
	Impact     = "impact"
	Complexity = "complexity"
	Proposed   = "proposed"
	Dimension  = "dimension"

	// Class is what KIND of thing this is, from a closed set the verb documents: a gap's
	// class slug from the growing registry, a petition's class from a fixed four. Same
	// arrangement as --as — one word, a value space per verb.
	//
	// A mint that supplies --definition is COINING the slug; there is no separate boolean
	// saying so, because a flag that means "I also passed three other flags" is those three
	// flags read twice.
	Class         = "class"
	Definition    = "definition"
	Neighbor      = "neighbor"
	Distinguisher = "distinguisher"

	// Gap substance.
	Problem = "problem"
	Fix     = "fix"
	Check   = "check"
	// CheckKind says what would SETTLE the acceptance check — read a document, run a
	// computation, or verify a source. It is what lets red demand evidence prose cannot fake.
	CheckKind = "check-kind"

	// Lineage, in two directions on one relation: what this gap replaces, and what replaces
	// it. Symmetric words, because the reader has to hold both at once.
	Supersedes   = "supersedes"
	SupersededBy = "superseded-by"
	FoundBy      = "found-by"
	CarriedFrom  = "carried-from"

	// Closure evidence: who checked it, with what, against what. NOT --anchor-*: --anchor is
	// a document anchor id everywhere else in this vocabulary, and a seat that learned it
	// there read these as anchors too.
	VerifiedBy      = "verified-by"
	VerifiedWith    = "verified-with"
	VerifiedAgainst = "verified-against"

	// Anchor is the DOCUMENT anchor — the id inside a `<!--cite:c-…-->`, `<!--fx:f-…-->` or
	// `<!--proof:p-…-->` token — naming a place in the report by identity rather than by
	// quoting it. Where --quote says the text, this says the marker.
	//
	// Two verbs take it and each narrows the CLASS it accepts, which is the per-verb
	// validator's job rather than a second word's.
	Anchor = "anchor"

	// The bench's vocabulary. Binds is the addressee of granted relief — set on a RULING,
	// never on the filing. What a petitioner asks for and what the bench orders are different
	// facts, and only the second binds anyone.
	Binds      = "binds"
	Principle  = "principle"
	Tension    = "tension"
	ReviewFlag = "review-flag"
	// Settled and ReopensOn are what a ruling BARS and what would undo it — the two things a
	// losing party needs and a disposition cannot supply (#502).
	Settled   = "settled"
	ReopensOn = "reopens-on"
	// Final is ReopensOn's assertable empty case, on the `friction --none` pattern: "nothing
	// would reopen this" is a positive answer to the question, not a skipped field.
	Final  = "final"
	Relief = "relief"

	// Blue's process record.
	Method     = "method"
	AccessDate = "access-date"

	// The operator `setup` command: it builds a research run's blackboard. Operator flags,
	// not seat-verb payload keys.
	Topic         = "topic"
	Model         = "model"
	JudgmentModel = "judgment-model"
	Cite          = "cite"
	MaxRounds     = "max-rounds"
	Lanes         = "lanes"
	BinDir        = "bin-dir"
	MemoryDir     = "memory-dir"
	RunID         = "run-id"
	ScriptPath    = "script-path"

	// The operator `scorecard` command: a chair's in-run self-read.
	Chair = "chair"

	// The operator `dashboard` command. --watch regenerates on a timer; --now injects the
	// clock (test determinism); --serve hosts it over HTTP, rendered fresh per request.
	Watch = "watch"
	Now   = "now"
	Serve = "serve"
)

// All is the declared vocabulary, enumerated.
//
// Go cannot reflect over constants, so the only way a test can ask "is every registered
// flag a declared one, and does every declared one get registered" is if the set is
// listed. Both questions are asked in internal/cli/vocabulary_test.go, and the second is
// what would have caught this package's first day, when 20 of 24 constants were orphans
// and the CLI spoke a vocabulary this file did not describe.
func All() []string {
	return []string{
		Run, SeatID, JSON,
		Reason, ReasonFile,
		ID, IDs, Key, Quote, New, Answers, URL, Title, Format, Window,
		Sitting, Trajectory,
		As, None, Ended, Confidence,
		Severity, Likelihood, Impact, Complexity, Proposed, Dimension,
		Class, Definition, Neighbor, Distinguisher,
		Problem, Fix, Check, CheckKind, Hypothesis, Script, Cites,
		Supersedes, SupersededBy, FoundBy, CarriedFrom,
		VerifiedBy, VerifiedWith, VerifiedAgainst, Anchor,
		Principle, Tension, ReviewFlag, Relief, Binds,
		Settled, ReopensOn, Final,
		Method, AccessDate,
		Topic, Model, JudgmentModel, Cite, MaxRounds, Lanes, BinDir, MemoryDir, RunID, ScriptPath,
		Chair, Watch, Now, Serve,
	}
}

// ForPayloadKey maps a stored payload key back to the flag a seat types to set it.
//
// The two are NOT the same word and must not be derived from each other. Validation used to
// build its error message by replacing the first underscore with a hyphen — gap_id became
// "--gap-id" — which was true until the surface renamed that flag to --id. The message then
// taught a spelling the parser rejects, and a wrong one costs a whole turn.
//
// Unknown keys fall back to that transform: a payload key with no flag is one a verb sets
// internally, and the hyphenated form is the best available guess for a field a seat cannot
// type anyway.
//
// LIMIT, stated because it is easy to over-trust this: the map is GLOBAL and payload keys
// are NOT globally unique. A caller on a verb-specific key needs a per-verb lookup, not
// another line here — adding one would make this function quietly wrong for the other verb.
func ForPayloadKey(key string) string {
	if name, ok := payloadFlag[key]; ok {
		return name
	}
	return strings.ReplaceAll(key, "_", "-")
}

// Only the keys whose flag is a DIFFERENT WORD need an entry. review_flag, principle,
// tension and the rest are spelled by the fallback and would be noise here.
var payloadFlag = map[string]string{
	"gap_id": ID,
	// The line-of-inquiry id a support verdict joins on. Typed as --id, the word every verb
	// uses for the thing it is acting on, and stored as `inquiry_id` because the event schema
	// names WHICH id it is — a payload carrying a bare `id` beside a gap's would be two facts
	// wearing one key.
	"inquiry_id":    ID,
	"disposition":   As,
	"outcome":       As,
	"status":        As,
	"closure_class": As,
	"verdict":       As,
	"soundness":     As,
	// Every span a verb addresses in blue/report.md is typed as --quote; the schema keeps the
	// key that says WHICH span it is.
	"location":         Quote,
	"claim":            Quote,
	"old":              Quote,
	"fix_new":          New,
	"superseded_by":    New,
	"successor":        SupersededBy,
	"acceptance_check": Check,
	"complexity_cost":  Complexity,
	"anchor_seat":      VerifiedBy,
	"anchor_tool":      VerifiedWith,
	"anchor_target":    VerifiedAgainst,
	// Every prose payload key funnels through the one --reason flag. The keys stay distinct in
	// the event schema — a dispute stores `evidence`, an opinion `rationale`, a close `prose`
	// — while the WORD a seat types is `reason`.
	// `line` is a line of inquiry's own statement and `row` a manifest receipt; both are the
	// verb's prose, and both are typed as --reason like every other prose field.
	"line":      Reason,
	"row":       Reason,
	"prose":     Reason,
	"text":      Reason,
	"evidence":  Reason,
	"rationale": Reason,
	"opinion":   Reason,
	"statement": Reason,
	"basis":     Reason,
	"notes":     Reason,
}

// Canonical descriptions. The same word gets the same explanation wherever it appears — the
// audit that produced this file found --file described two ways and --id three ways, which
// teaches a seat that they might be different things.
const (
	DescReason     = "your reasoning or argument for this act — the substance the report renders and the other side answers; --reason-file for anything long or from stdin with -"
	DescReasonFile = "read --reason from a file, or from stdin with `-` — the same field as --reason, for anything long or that would fight shell quoting"

	// DescQuote is the whole contract of --quote, and it is stated once because it was the
	// contradiction: quote the text and NOTHING else. A section heading, a dash or a pipe
	// prepended to it makes the match fail, and the match is what places the anchor.
	DescQuote = "the EXACT text from blue/report.md, quoted verbatim and NOTHING else — no section heading, no dash, no pipe: this whole string is matched against the report, so anything you prepend makes it match nothing. Name the section in --reason, where prose belongs"
	DescKey   = "a stable local handle of your own (C1, F2, P3 …) making a retried call idempotent — a repeat under the same handle returns the first result instead of acting twice"
	DescURL   = "the source's http/https URL — fetched once and cached, so both sides read the same bytes"
	DescTitle = "the source's name, as it appears in the composed bibliography"
)
