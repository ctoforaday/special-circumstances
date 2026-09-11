// Package reportvoice holds the one list of PROCESS-VOICE TELLS, and it is one list on purpose.
//
// The research report is addressed to a reader of its SUBJECT. Measured on the 2026-09-02
// quadratic-formula run, a 487-line report on a 4,000-year-old algebra question instead narrated
// its own construction: 161 "this run / this epoch / the debate", 24 inline lane-attribution tags
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
// may legitimately say "this epoch", and prose written for a human reader is never the violation.
// The advisory does not block, and the lens argues rather than enforces — a tell is where to LOOK.
package reportvoice

import (
	"regexp"
	"sort"
	"strings"
)

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
	{ProcessVoice, regexp.MustCompile(`(?i)\bthis (run|round|report|sitting)\b`),
		"the record already holds the run; a sentence about the subject does not need to name it"},
	{ProcessVoice, regexp.MustCompile(`(?i)\b(the|this) debate\b`),
		"the record already holds the debate; say what is true of the subject"},
	{LaneAttribution, regexp.MustCompile(`\[(minority|lane-\d)[^\]]*\]`),
		"provenance is the record's; a claim in the report is the report's"},
	{LaneAttribution, regexp.MustCompile(`(?i)\bresearch lanes?\b`),
		"which lane found it is the record's; the report says what was found"},
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

// Find reports WHICH TELLS ARE PRESENT in a span of report prose — at most one per tell, the
// first match. An empty result is not a clean bill: these are the LITERAL tells, and the leaks
// that matter most are the ones no pattern catches — which is why red's lens reads for voice
// rather than running this and stopping.
//
// PRESENCE IS THE RIGHT SHAPE FOR A SPAN, and the wrong one for a document. `blue edit` passes one
// edit's --new text, where "this span carries a lane tag" is the whole useful signal and a second
// occurrence twelve characters later tells the author nothing new. Hand it a whole report and the
// same shape becomes a defect: measured on the #861 arm-A base, this returns 3 findings for
// 16,931 bytes carrying sixteen occurrences, and an author told once that a lane tag exists has no
// way to reach the other five. Use FindAll when the input is a document (#873).
func Find(s string) []Found {
	var out []Found
	for _, t := range tells {
		if m := t.Pattern.FindString(s); m != "" {
			out = append(out, Found{Tell: t, Match: m})
		}
	}
	return out
}

// An Occurrence is one match, with enough to go and look at it.
type Occurrence struct {
	Tell
	Match string
	// Offset is the byte offset into the text that was searched, and Line is its 1-indexed line.
	// The line is what an author needs and the offset is what a caller needs to compute anything
	// else; deriving the line here rather than at each call site keeps one definition of it.
	Offset int
	Line   int
}

// FindAll reports EVERY occurrence of every tell, in the order they appear in the text.
//
// This is the census half of the surface, for callers whose input is a whole document rather than
// one edit — today `blue ingest`, which writes the largest authored artifact in a run. It is
// still ADVICE and still refuses nothing; what it adds over Find is that the count and the
// locations are real, so "six lane tags at lines 54, 70, 74, 76, 82, 96" can be acted on where
// "a lane tag is present" cannot.
//
// Ordered by position, NOT grouped by tell: an author reads a document top to bottom, and the
// grouping a caller wants for a summary is one line of code away from this while the ordering is
// not recoverable once discarded.
func FindAll(s string) []Occurrence {
	var out []Occurrence
	for _, t := range tells {
		for _, loc := range t.Pattern.FindAllStringIndex(s, -1) {
			out = append(out, Occurrence{
				Tell:   t,
				Match:  s[loc[0]:loc[1]],
				Offset: loc[0],
				Line:   1 + strings.Count(s[:loc[0]], "\n"),
			})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Offset < out[j].Offset })
	return out
}

// CountByClass summarises a census, for a caller that wants the shape before the detail.
func CountByClass(occ []Occurrence) map[Class]int {
	by := map[Class]int{}
	for _, o := range occ {
		by[o.Class]++
	}
	return by
}
