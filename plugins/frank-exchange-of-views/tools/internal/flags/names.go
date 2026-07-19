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
package flags

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
	ID       = "id"
	Label    = "label"
	Claim    = "claim"
	Key      = "key"
	Location = "location"

	// Disposition and justification.
	As     = "as"
	Reason = "reason"
	Basis  = "basis"
	Notes  = "notes"

	// Grading.
	Severity   = "severity"
	Likelihood = "likelihood"
	Impact     = "impact"
	Complexity = "cx"

	// Gap classification. PetitionClass is deliberately NOT `class`: a petition's class
	// is a fixed four-value enum, while a gap's class is a slug from a growing registry.
	// One word for two vocabularies is how a seat learns the wrong one.
	Class         = "class"
	ClassNew      = "class-new"
	PetitionClass = "petition-class"

	// Lineage.
	Supersedes  = "supersedes"
	Successor   = "successor"
	FoundBy     = "found-by"
	CarriedFrom = "carried-from"
)

// Canonical descriptions. The same word gets the same explanation wherever it appears —
// the audit found --file described two ways and --id three ways, which teaches a seat
// that they might be different things.
const (
	DescFile = "read the payload from a file — ALWAYS use this over --text for anything above ~2KB"
	DescText = "the payload, inline (short values only)"

	DescComment      = "free text this verb has no field for — recorded on the event, and a recurring one is a schema gap"
	DescCommentFile  = "read --comment from a file instead, for anything long or awkward to quote"
	DescCommentStdin = "read --comment from stdin, so nothing has to survive shell quoting"
)
