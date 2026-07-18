package record

// THE HELP IS THE CONTRACT.
//
// A seat is not a human operator browsing a man page. It reads the help, and
// whatever it finds there is what it believes it may do. That makes two things
// true at once:
//
//	if a seat can SEE it, it must be able to DO it — a visible-but-forbidden
//	command is a plausible action that fails, which costs turns and invites the
//	seat to improvise a workaround;
//
//	if a seat CANNOT see it, that is the answer — the capability does not exist
//	for it, and a missing capability is FRICTION, not a puzzle to solve.
//
// Which is why an undocumented flag is a real defect and not a cosmetic one. The
// verbs shipped with rich Help strings, but every flag was registered with an
// empty description, so `close --help` listed `--anchor-seat string` with nothing
// beside it. A seat reading that has to infer the anchor's meaning from the verb
// blurb, and inference is exactly what these tools exist to remove: the whole
// point was that the structure never has to be got right from memory.
//
// Docs are keyed "verb:flag" first, then "flag", because a few names are
// deliberately overloaded across verbs (--as is a closure class at close, a
// disposition at dispose, a verdict at verdict, a ruling at petition-rule) and a
// single shared gloss for those would be worse than none.
var flagDocs = map[string]string{
	// identity and payload, shared by every verb
	"run":     "the run directory (the engine passes it; it is in your prompt)",
	"seat-id": "your seat id, assigned by the engine and bound to this role's namespace",
	"file":    "read the prose payload from a file — ALWAYS use this over --text for anything above ~2KB",
	"text":    "the prose payload, inline (short values only)",

	// lens
	"label":       "your lens-scoped finding label (L3-F1). Stable R<round>-N ids are the merge's to assign, never a lens's",
	"location":    "where the defect lives: a section heading plus a quoted sentence",
	"kind":        "note | checked-held — an observation's flavour; the merge must still dispose it",
	"claim":       "the claim being verified, quoted from the report",
	"reference":   "the source the claim rests on",
	"confidence":  "high | medium | low — your confidence the source SUPPORTS the claim",
	"access-date": "YYYY-MM-DD you actually fetched it; drives the staleness re-fetch trigger",

	// grading
	"severity":   "low | low-medium | medium | medium-high | high | certain | realized | trivial",
	"likelihood": "how likely the CONSEQUENCE is (v2 grades consequence only, never existence)",
	"impact":     "how bad the consequence is if it lands",
	"cx":         "complexity_cost — what fixing it costs, graded on the same scale",

	// merge: minting and lineage
	"key":           "a stable local label (e.g. the source lens finding) making a retried mint idempotent",
	"class":         "the gap's class slug from the registry — what KIND of defect this is",
	"class-new":     "mint a new class slug; requires --definition, --neighbor and --distinguisher",
	"definition":    "what the new class is, in one line",
	"neighbor":      "the existing class this one sits closest to",
	"distinguisher": "the tie-break question that tells the two apart",
	"problem":       "what is wrong (or pass it via --file)",
	"fix":           "the required fix",
	"check":         "the acceptance check red will RUN at re-audit — the pre-agreed contract, not a description",
	"supersedes":    "comma-separated ancestor ids this gap replaces; lineage is never dropped",
	"found-by":      "comma-separated lens findings that surfaced it (L5-F3,L6-F2)",
	"id":            "the gap id",
	"successor":     "the gap id carrying the unresolved remainder forward",

	// merge: closure
	"close:as":           "closed | closed_with_regression | ... — the closure class",
	"anchor-seat":        "WHO verified the closure (the seat)",
	"anchor-tool":        "WITH WHAT it was verified (the tool or command)",
	"anchor-target":      "AGAINST WHAT — the exact file, ref or URL read",
	"carried-from":       "the round this closure was carried from, when it is not a fresh act",
	"prose-file":         "the full closure record, from a file",
	"dispose:as":         "minted-as | folded-into | declined | banked — the observation's fate",
	"observation":        "the lens observation being disposed, by label or key",
	"into":               "the gap id it was folded or minted into",
	"reason":             "why, when the fate is declined or banked",
	"basis":              "why the grade moved — grade movement is recorded with its reason",
	"dispute-respond:as": "accepted | rejected — your answer to blue's grade dispute",
	"rationale":          "the reasoning behind the answer",
	"ids":                "comma-separated archived closures you re-verified this round",
	"notes":              "what the spot-check found",
	"verdict:as":         "PASS | FAIL — the seat's terminal act",

	// blue
	"claim-count": "FOOTNOTED declarative claims in the report (a footnoted sentence counts once)",
	"row":         "the correctness-manifest receipt: what you checked and what it showed",
	"dimension":   "severity | likelihood | impact | complexity_cost — the axis you contest",
	"proposed":    "the grade you say it should be",
	"evidence":    "why, citing the section",
	"grade":       "high | medium | low — your confidence in the claim",

	// bench
	"gap-id":           "the gap being ruled on",
	"disposition":      "carried | closed | rebuttal_sustained | risk_accepted | routed_to_infrastructure | ...",
	"principle":        "the principle applied — a ruling is an OPINION, not a disposition",
	"tension":          "the values in tension (e.g. correctness vs economy)",
	"review-flag":      "why a human should, or should not, look at this",
	"petitioner":       "the seat that filed the petition",
	"petition-rule:as": "granted | denied — a halt is its own verb",
	"petition:class":   "ethical | safety | integrity | constitutional",
	"relief":           "the relief sought, stated as it would bind the coming seats",
}

// flagDoc resolves a flag's documentation, most specific first.
func flagDoc(verb, flag string) string {
	if d, ok := flagDocs[verb+":"+flag]; ok {
		return d
	}
	return flagDocs[flag]
}

// frictionFooter closes the loop the help opens. A seat that needs something the
// contract does not offer must not improvise around it — the gap in the tooling
// is itself a finding, and friction is the channel that carries it to the human
// who can retool the seat.
const frictionFooter = `
If you need a verb or a flag that is not listed here, it does not exist for you:
do not improvise around it, and do not hand-write the artifact. Record what you
needed and what you would have done with the 'friction' verb — a missing
capability is a finding about the tooling, and that channel is how it gets fixed.`
