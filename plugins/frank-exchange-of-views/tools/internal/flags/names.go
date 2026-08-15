// Package flags owns the CLI's vocabulary.
//
// THE PROBLEM THIS SOLVES. Every verb declared its own flags with its own wording, and
// the vocabulary drifted the way vocabularies do when thirty files each get a vote. The
// 2026-07-18 run measured what that cost: `merge close` named its payload --prose-file
// while eight other verbs called the same thing --file, and seats typed --file there and
// were refused. An audit of the whole surface then found 8 of 57 flag names carrying more
// than one description, including two genuine collisions — --class meaning both a
// petition class and a gap-registry slug, and --label meaning a finding label on one verb
// and "the claim this attaches to" on another.
//
// A flag name is a WORD IN A LANGUAGE the seats have to learn once and reuse everywhere.
// Two meanings for one word is a defect even when each is individually sensible, because
// the seat generalises from the verb it learned first.
//
// So: names are constants here, registration goes through the helpers here, and a verb
// cannot invent a private spelling without editing this file — which is the point. The
// constant is not for the compiler's benefit; it is a chokepoint that makes divergence
// require a deliberate act.
//
// THAT CLAIM WAS FALSE FOR A DAY. The file shipped 2026-07-18 with 24 constants, 20 of
// them referenced NOWHERE: every verb went on registering literals, so the chokepoint
// gated nothing and both collisions above survived the file that was written to prevent
// them. Fixed 2026-07-19 — all 55 registered names now resolve through these constants,
// --class/--petition-class and --label/--claim are separated, --rationale collapsed into
// --basis, and --grade into --confidence. An unused single source of truth is a comment,
// not a constraint; if you add a verb, register through here or this comment starts lying
// again.
//
// FLAG WORDS ARE NOT PAYLOAD KEYS. The keys are the event schema and move on their own
// schedule, so `blue confidence --claim X --confidence med` still stores label/grade. Keep
// them free to differ: renaming a word a seat types is cheap, rewriting recorded history
// is not.
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
	// (--reason) or from a file/stdin (--reason-file). It replaced the split
	// --file/--text payload channel, the universal --comment junk drawer, and the
	// contested-exchange --basis; every prose argument a verb records now lands here.
	Reason     = "reason"
	ReasonFile = "reason-file"

	// Identity and reference.
	ID        = "id"
	IDs       = "ids"
	Claim     = "claim"
	Key       = "key"
	Location  = "location"
	Reference = "reference"
	// The citation axis (bibliography core). --url is the source a `fetch` reads and a
	// `blue cite` anchors; --title is the human name that source carries into the composed
	// bibliography. One canonical word each — a citation is tool-managed, so these are the
	// only two facts a seat supplies about a source.
	URL   = "url"
	Title = "title"
	Row   = "row"
	// View is RETIRED (0.56.0). `show --view <name>` became `show <name>`: cobra models
	// commands and flags, never a flag's VALUE space, so every projection's description had to
	// be crammed into one usage string with no --help and no completion of its own. Making the
	// flag optional made it worse — a seat had no reason to discover it at all.
	Format = "format" // operator output selector (graph: mermaid | dot)
	// Candidate is the merge's near-match query: the problem text of a gap it is about
	// to mint, screened against the board (open + closed) before minting so a
	// near-duplicate surfaces as a reopen rather than a fresh gap. One word, read-only.
	Candidate = "candidate"

	// `blue edit` is the ONLY write path to blue/report.md for a response seat: it replaces
	// the exact current span --old with --new, preserving red's finding-markers. The words
	// mirror the Edit tool's old_string/new_string so a seat reaches for what it knows.
	Old = "old"
	New = "new"
	// Answers is the edit's PROVENANCE: the gap id this edit is blue's response to.
	// A distinct word from --id (which names the gap an act is ABOUT) because the
	// relation differs — the same reason --into, --successor, --superseded-by and
	// --carried-from each get their own word rather than overloading one.
	//
	// It exists because the link was a CONVENTION: 19 of 26 edits in the 2026-08-04 smoke
	// happened to open --reason with the gap id, and 7 did not. Prose that usually carries
	// a fact is not a join key, and every measurement in #267 needs one.
	Answers = "answers"
	// An avenue's ABSTRACT: what would be true if this line of inquiry pays off (#246).
	// Distinct from --problem (a defect) and --reason (the argument for an act) because it
	// is a forward claim, and it is what makes a later abandonment checkable — an avenue
	// abandoned against its own stated hypothesis is evidence of choosing; one abandoned on
	// a shrug is not.
	Hypothesis = "hypothesis"
	// The proof axis (#277). --script is the program that settles a claim; --cites names the
	// METHOD citation it applies. Separate words from --url/--title (a document's identity)
	// because a computation and a source are different things a claim can rest on — and
	// deliberately NOT --method, which already means "the source class an avenue belonged to".
	Script = "script"
	Cites  = "cites"
	// Red's fate for a proposed direction. A separate word from --as (an observation's
	// disposal) and --status (an avenue's own lifecycle state) because it is a different
	// party ruling on a different question: is this worth the run's time at all.
	// A CONCRETE proposed fix (#267 stage 3): the exact span red says is wrong and the exact
	// text it should become. Distinct words from --old/--new, which are the edit blue APPLIES;
	// these are what red PROPOSES, and one word per act keeps a seat from typing red's
	// proposal into blue's write verb by muscle memory.
	//
	// Optional, and bounded: a proposal that grows the span past bluedoc.MaxProposalGrowth is
	// refused as authoring. Their presence is what DERIVES `fix_basis: verified` — the basis is
	// never a word a seat types about its own work.
	FixOld = "fix-old"
	FixNew = "fix-new"

	// Disposition and justification.
	//
	// --reason (declared above with the prose payload) is now the ONE word for every
	// prose argument, whether or not it is contested. It absorbed --basis (the grounds
	// argued FROM in a dispute/regrade/dispute-respond/petition) and the --rationale
	// that briefly split one argument's two halves: a seat learns a single word and
	// reaches for it everywhere, and the event schema keeps its own distinct keys
	// (evidence, basis, rationale) underneath.
	As     = "as"
	Notes  = "notes"
	Status = "status"
	None   = "none"

	// Confidence is how sure red is of a determination it just made — orthogonal to WHAT the
	// determination was (--as), and the field the original plan specified: "for each statement ↔
	// reference pair it assigns a confidence that the source actually corroborates the statement
	// … low confidence → needs more evidence, blue digs further, not an automatic fail".
	//
	// The word was surrendered in #341 to a collision with `blue confidence` — blue self-grading
	// its own claims — and the field went out as --trust. That verb was DELETED in 0.54.0, so the
	// collision has not existed since, and the substitute cost more than it saved: `trust` reads
	// as a property of the SOURCE, which pulled the value descriptions into a support scale and
	// then read as a positive-only outcome. One word, one question — and the question is "how
	// sure are you", never "how good is the source".
	Confidence = "confidence"

	// Grading.
	Severity   = "severity"
	Likelihood = "likelihood"
	Impact     = "impact"
	Complexity = "cx"
	Proposed   = "proposed"
	Dimension  = "dimension"

	// Gap classification. A gap's --class is a slug from a GROWING REGISTRY; a petition's
	// --petition-class is a FIXED four-value enum. They shared the word --class until
	// 2026-07-19, which is the collision this package was created to prevent and then
	// carried anyway: the constant existed, was referenced nowhere, and both verbs went
	// on registering the literal. Two vocabularies, two words.
	Class         = "class"
	PetitionClass = "petition-class"
	ClassNew      = "class-new"
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

	// Lineage. Supersedes (a minted gap's ancestors) and SupersededBy (the claim that
	// replaces a retired one) are different directions of a different relation on
	// different objects — two words, deliberately.
	Supersedes   = "supersedes"
	SupersededBy = "superseded-by"
	Successor    = "successor"
	FoundBy      = "found-by"
	CarriedFrom  = "carried-from"

	// Closure anchoring: who, with what, against what.
	AnchorSeat   = "anchor-seat"
	AnchorTarget = "anchor-target"
	AnchorTool   = "anchor-tool"

	// Anchor is the DOCUMENT anchor — the c-<hex> inside a `<!--cite:c-…-->` token — naming
	// which citation a verification adjudicates. Distinct from the AnchorSeat/Target/Tool trio
	// above, which anchor a CLOSURE (who checked what, with which tool); this one is the join
	// key between red's verification and the source blue actually cited (#382).
	Anchor = "anchor"
	// Independent is the explicit negative beside it: a source red found itself, which has no
	// anchor to name. Stated rather than left as an empty --anchor, because an empty field
	// cannot distinguish "corroboration" from "I never looked it up" — the same argument that
	// gave `friction` its --none.
	Independent = "independent"

	// The bench's vocabulary.
	// Binds is the addressee of granted relief — set on a petition RULING, never on the filing.
	// What a petitioner asks for and what the bench orders are different facts, and only the
	// second binds anyone. Without it, relief was threaded into one hardcoded prompt and relief
	// for any other seat reached nothing (#360).
	Binds = "binds"

	Principle  = "principle"
	Tension    = "tension"
	ReviewFlag = "review-flag"
	Relief     = "relief"

	// The run's terminal determination (bench outcome). The verdict rides on --as; these
	// two say HOW a non-pass ended, and are booleans because each is present-or-not.
	Deadlocked = "deadlocked"
	Exhausted  = "exhausted"

	// Blue's process record. (claim_count was --claim-count here until #70 moved the
	// count to the deterministic `count-claims` root command; no flag types it now.)
	Line       = "line"
	Method     = "method"
	AccessDate = "access-date"

	// The operator `setup` command (ported from setup-research-run.mjs): it builds a
	// research run's blackboard. These are operator flags, not seat-verb payload keys.
	Topic         = "topic"
	Model         = "model"
	JudgmentModel = "judgment-model"
	Cite          = "cite"
	MaxRounds     = "max-rounds"
	Lanes         = "lanes"
	BinDir        = "bin-dir"
	MemoryDir     = "memory-dir"

	// The operator `scorecard` command (ported from scorecards.mjs): a chair's in-run
	// self-read. --run is the shared persistent flag; --chair selects the chair.
	Chair = "chair"

	// The operator `dashboard` command (ported from render-run-dashboard.mjs): the live run
	// dashboard. --watch regenerates on a timer; --now injects the clock (test determinism);
	// --serve hosts it over HTTP, rendered fresh per request from the record.
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
//
// Yes, this duplicates the block above; the duplication is load-bearing and the tests
// close it. Adding a constant without adding it here fails the round-trip.
func All() []string {
	return []string{
		Run, SeatID, JSON,
		Reason, ReasonFile,
		ID, IDs, Claim, Key, Location, Reference, URL, Title, Row, Format, Candidate, Old, New, Answers,
		As, Notes, Status, None,
		Severity, Likelihood, Impact, Complexity, Proposed, Dimension, Confidence,
		Class, PetitionClass, ClassNew, Definition, Neighbor, Distinguisher,
		Problem, Fix, Check, CheckKind, FixOld, FixNew, Hypothesis, Script, Cites,
		Supersedes, SupersededBy, Successor, FoundBy, CarriedFrom,
		AnchorSeat, AnchorTarget, AnchorTool, Anchor, Independent,
		Principle, Tension, ReviewFlag, Relief, Binds,
		Deadlocked, Exhausted,
		Line, Method, AccessDate,
		Topic, Model, JudgmentModel, Cite, MaxRounds, Lanes, BinDir, MemoryDir,
		Chair,
		Watch, Now, Serve,
	}
}

// ForPayloadKey maps a stored payload key back to the flag a seat types to set it.
//
// The two are NOT the same word and must not be derived from each other. Validation
// used to build its error message by replacing the first underscore with a hyphen —
// gap_id became "--gap-id" — which was true until the command-surface audit renamed
// that flag to --id and --disposition to --as. The message then taught a spelling the
// parser rejects, and the seat's only teacher is the error text, so a wrong one costs a
// whole turn. Derivation was a guess that happened to be right; this is a statement.
//
// Unknown keys fall back to the old transform: a payload key with no flag is one a verb
// sets internally, and the hyphenated form is the best available guess for a field a
// seat cannot type anyway.
//
// LIMIT, stated because it is easy to over-trust this: the map is GLOBAL and payload keys
// are NOT globally unique. Key "label" is typed as --label on `lens finding` but as
// --claim on `blue confidence`; key "grade" is --confidence there and nothing elsewhere.
// A global map cannot express that, so those keys deliberately have no entry and fall
// through. Today the only caller is `opinion` validation, whose four keys are unambiguous.
// A second caller on a verb-specific key needs a per-verb lookup, not another line here —
// adding one would make this function quietly wrong for the other verb.
func ForPayloadKey(key string) string {
	if name, ok := payloadFlag[key]; ok {
		return name
	}
	return strings.ReplaceAll(key, "_", "-")
}

// Only the keys whose flag is a DIFFERENT WORD need an entry. review_flag, principle,
// tension and the rest are spelled by the fallback and would be noise here.
var payloadFlag = map[string]string{
	"gap_id":      ID,
	"disposition": As,
	// A verification's verdict — what the source did for the claim. Typed as --as, the word this
	// vocabulary already uses for every verdict a seat renders (`close --as`, `reproduce --as`,
	// `verdict --as`), and stored as `outcome` because the event schema names the fact rather
	// than the syntax that carried it.
	"outcome":          As,
	"acceptance_check": Check,
	// Every prose payload key now funnels through the one --reason flag, so each maps
	// to it. The keys stay distinct in the event schema — a dispute stores `evidence`,
	// a regrade `basis`, an opinion `rationale`, a close `prose` — while the WORD a
	// seat types collapsed to `reason`. None of these keys is spelled by any other flag,
	// so the global map stays unambiguous (see the LIMIT note above).
	"prose":     Reason,
	"text":      Reason,
	"evidence":  Reason,
	"rationale": Reason,
	"opinion":   Reason,
	"statement": Reason,
	"basis":     Reason,
}

// Canonical descriptions. The same word gets the same explanation wherever it appears —
// the audit found --file described two ways and --id three ways, which teaches a seat
// that they might be different things.
const (
	DescReason     = "your reasoning or argument for this act — the substance the report renders and the other side answers; --reason-file for anything long or from stdin with -"
	DescReasonFile = "read --reason from a file, or from stdin with `-` — the same field as --reason, for anything long or that would fight shell quoting"

	DescPetitionClass = "ethical | safety | integrity | constitutional"
)
