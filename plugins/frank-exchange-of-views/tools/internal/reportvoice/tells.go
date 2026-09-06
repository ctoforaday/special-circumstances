// Package reportvoice holds the one list of PROCESS-VOICE TELLS, and it is one list on purpose.
//
// The research report is addressed to a reader of its SUBJECT. Measured on the 2026-09-02
// quadratic-formula run, a 487-line report on a 4,000-year-old algebra question instead narrated
// its own construction: 161 "this run / this round / the debate", 24 inline lane-attribution tags
// in the research prose, 13 inlined access limits, 9 narrations of its own draft history, 2
// intrusions of its own verification apparatus. The report had, in its own words, "an account of
// itself" as a co-equal subject.
//
// TWO READERS, ONE SOURCE. A tell is checked in two places — an advisory when blue writes, and
// red's voice lens when it audits — and a second hand-kept copy of this list would drift the
// moment one was edited alone. That is the defect this package exists to avoid, not merely a
// tidiness: `flags.All()` is the precedent, enumerated once and read by the gate.
//
// WHAT THIS IS NOT. It is not a censor. Matching a tell is not proof of a leak: a quoted source
// may legitimately say "this round", and prose written for a human reader is never the violation.
// The advisory does not block, and the lens argues rather than enforces — a tell is where to LOOK.
package reportvoice

import "regexp"

// Class is what KIND of thing leaked, because the destinations differ. Process voice belongs on
// the record; an operational limit belongs on the operator channel; and the epistemic residue of
// either belongs in the report, RE-VOICED as a limit on the conclusion rather than a fact about
// the run. Separation, never deletion: "Savage is known only through the interested party's
// summary" stays, and "after four hosts refused this container" goes.
type Class string

const (
	// ProcessVoice is the report speaking about the run that made it.
	ProcessVoice Class = "process-voice"
	// LaneAttribution is a claim wearing the seat that produced it, inside research prose.
	LaneAttribution Class = "lane-attribution"
	// DraftHistory is the report narrating what it used to say.
	DraftHistory Class = "draft-history"
	// Apparatus is the report describing the machinery that checked it.
	Apparatus Class = "apparatus"
)

// A Tell is one pattern, what it is, and where the thing it caught belongs instead.
type Tell struct {
	Class    Class
	Pattern  *regexp.Regexp
	Redirect string
}

var tells = []Tell{
	{ProcessVoice, regexp.MustCompile(`(?i)\bthis (run|round|report)\b`),
		"the record already holds the run; a sentence about the subject does not need to name it"},
	{ProcessVoice, regexp.MustCompile(`(?i)\bthe debate\b`),
		"the record already holds the debate; say what is true of the subject"},
	{LaneAttribution, regexp.MustCompile(`\[(minority|lane-\d)[^\]]*\]`),
		"provenance is the record's; a claim in the report is the report's"},
	{DraftHistory, regexp.MustCompile(`(?i)an earlier version of this (sentence|bullet|paragraph)|corrected here`),
		"the change stack holds what the report used to say"},
	{Apparatus, regexp.MustCompile(`(?i)the checking (program|script)|measurement apparatus`),
		"the proof store holds the program; the report carries what it SHOWED"},
}

// Tells is the whole list, and the only way to get it.
func Tells() []Tell { return append([]Tell(nil), tells...) }

// Found is one tell that matched, with the text it matched on.
type Found struct {
	Tell
	Match string
}

// Find reports every tell present in a span of report prose. An empty result is not a clean bill:
// these are the LITERAL tells, and the leaks that matter most are the ones no pattern catches —
// which is why red's lens reads for voice rather than running this and stopping.
func Find(s string) []Found {
	var out []Found
	for _, t := range tells {
		if m := t.Pattern.FindString(s); m != "" {
			out = append(out, Found{Tell: t, Match: m})
		}
	}
	return out
}
