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

	// The prose payload, in its three interchangeable forms.
	File = "file"
	Text = "text"

	// The universal free-text field, in its three interchangeable forms.
	Comment      = "comment"
	CommentFile  = "comment-file"
	CommentStdin = "comment-stdin"

	// Identity and reference.
	ID        = "id"
	IDs       = "ids"
	Label     = "label"
	Claim     = "claim"
	Key       = "key"
	Location  = "location"
	Reference = "reference"
	Row       = "row"

	// Disposition and justification.
	//
	// Reason and Basis are TWO WORDS FOR TWO INTENTS, and the line between them is
	// whether anyone is being argued with:
	//   --reason ... why a thing happened or was not done, stated to a reader who is not
	//               contesting it (an avenue declined, a claim retired, an empty sample).
	//   --basis  ... the grounds you argue FROM in a contested exchange, where another
	//               seat may answer (dispute, dispute-respond, regrade, petition).
	// A third word, --rationale, existed on dispute-respond alone for exactly the intent
	// --basis already carried on the dispute it was answering; the two halves of one
	// argument used different words. It is gone.
	As     = "as"
	Reason = "reason"
	Basis  = "basis"
	Notes  = "notes"
	Status = "status"
	Into   = "into"
	None   = "none"

	// Grading.
	Severity   = "severity"
	Likelihood = "likelihood"
	Impact     = "impact"
	Complexity = "cx"
	Proposed   = "proposed"
	Dimension  = "dimension"
	// Confidence is one word for one intent: how sure the seat is. `blue confidence`
	// called it --grade and `lens cite` called it --confidence, for the same self-graded
	// high|medium|low. --grade is also the wrong word for it, because a GRADE elsewhere
	// in this vocabulary is severity/likelihood/impact, which a seat does not self-assign.
	Confidence = "confidence"

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
	Kind          = "kind"

	// Gap substance.
	Problem = "problem"
	Fix     = "fix"
	Check   = "check"

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

	// The bench's vocabulary.
	Principle  = "principle"
	Tension    = "tension"
	ReviewFlag = "review-flag"
	Petitioner = "petitioner"
	Relief     = "relief"

	// Blue's process record.
	Line        = "line"
	Method      = "method"
	ClaimCount  = "claim-count"
	AccessDate  = "access-date"
	Observation = "observation"
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
		Run, SeatID,
		File, Text,
		Comment, CommentFile, CommentStdin,
		ID, IDs, Label, Claim, Key, Location, Reference, Row,
		As, Reason, Basis, Notes, Status, Into, None,
		Severity, Likelihood, Impact, Complexity, Proposed, Dimension, Confidence,
		Class, PetitionClass, ClassNew, Definition, Neighbor, Distinguisher, Kind,
		Problem, Fix, Check,
		Supersedes, SupersededBy, Successor, FoundBy, CarriedFrom,
		AnchorSeat, AnchorTarget, AnchorTool,
		Principle, Tension, ReviewFlag, Petitioner, Relief,
		Line, Method, ClaimCount, AccessDate, Observation,
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
	"evidence":    Basis,
	"prose":       File,
}

// Canonical descriptions. The same word gets the same explanation wherever it appears —
// the audit found --file described two ways and --id three ways, which teaches a seat
// that they might be different things.
const (
	DescFile = "read the payload from a file — ALWAYS use this over --text for anything above ~2KB"
	DescText = "the payload, inline (short values only)"

	DescComment      = "free text this verb has no field for — recorded on the event, and a recurring one is a schema gap"
	DescCommentFile  = "read --comment from a file instead, for anything long or awkward to quote"
	DescCommentStdin = "read --comment from stdin, so nothing has to survive shell quoting"

	DescPetitionClass = "ethical | safety | integrity | constitutional"
)
