// Package reportheading is the report's headings as every reader compares them, and the three
// sections assembly composes from the inquiries on the record, with the fates each holds.
//
// IT IS A LEAF so the record can ask what the report prints. Assembly (internal/report) lays the
// inquiry sections out from InquirySections and parts the report's text by HeadingKey; the record
// (internal/record) reads the same two to say which gaps a fate blue changed is an act on: a gap
// anchored on the section that prints that inquiry (ruled by gblock 2026-11-05). internal/report
// imports the record, so neither could ask the other.
package reportheading

import (
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// NoteOpen is what opens a footnote reference or definition.
const NoteOpen = "[^"

// HEADINGS ARE READ ONE WAY, by every reader: the composer that parts the report's text, the refusals
// at the write, the view that slices the assembled document, and the record's reading of a gap
// anchored on a section.
//
// HeadingKey folds a heading to its comparison key: two headings name the same section when they agree
// after their anchors and note references are taken out, case is folded, runs of whitespace are single
// spaces and a trailing "(in full)" is dropped. The note references are read as text (NoteRefsIn): a
// heading is read one line at a time, with no definition in reach to resolve them.
func HeadingKey(h string) string {
	h = anchor.Replace(h, func(string, string) string { return "" })
	for _, sp := range slices.Backward(NoteRefsIn(h)) {
		h = h[:sp[0]] + h[sp[1]:]
	}
	h = strings.ToLower(strings.Join(strings.Fields(h), " "))
	return strings.TrimSpace(strings.TrimSuffix(h, " (in full)"))
}

// NoteRefsIn is the span of each `[^label]` s holds as text: not escaped, its label one or more
// characters up to the first "]", with no "[" and no line break in it. It is how a reference reads
// where no definition is in reach to resolve it — a run the parser left as prose, or one line read
// alone.
func NoteRefsIn(s string) (spans [][2]int) {
	for i := 0; ; {
		j := strings.Index(s[i:], NoteOpen)
		if j < 0 {
			return spans
		}
		at := i + j
		i = at + len(NoteOpen)
		if EscapedAt(s, at) {
			continue
		}
		end := strings.IndexAny(s[i:], "]\n[")
		if end <= 0 || s[i+end] != ']' {
			continue
		}
		spans = append(spans, [2]int{at, i + end + 1})
	}
}

// EscapedAt reports whether the byte at i of s is escaped, and so is text: an odd run of backslashes
// stands before it. It is the one reading of the escape, for a `[^` and for a `<`.
func EscapedAt[T ~string | ~[]byte](s T, i int) bool {
	n := 0
	for j := i - 1; j >= 0 && s[j] == '\\'; j-- {
		n++
	}
	return n%2 == 1
}

// InquirySection is one section assembly composes from the inquiries: its heading, and the fates it
// holds, by the status word.
type InquirySection struct {
	Heading string
	Holds   func(status string) bool
}

// THE LIFECYCLE OF A RESEARCH TOPIC, in the three descriptive areas a reader gets:
//
//	Research areas               the approaches the research RAN: `underway` (running), `confirmed`
//	                             (ran, the hypothesis held) and `refuted` (ran, the hypothesis was
//	                             refuted) — a refutation is an answer, and it stands beside the
//	                             confirmations. AND `proposed`: an inquiry put forward and not yet
//	                             resolved is research this run has taken on, and its row carries
//	                             the undecided state so the heading does not have to lie about it.
//	Future research directions   `deferred` — "worth taking, and not by THIS run". KEPT, not set
//	                             aside. --reason says what the question still leaves unsettled,
//	                             --method what would settle it, and it reaches the report as a
//	                             proposal a human selects.
//	Alternatives considered      `declined` (weighed, never run) and `abandoned` (started, and no
//	                             answer came).
//
// # Every fate has a section
//
// Taking the set-aside inquiries as the complement of `underway` would put `deferred` AND
// `proposed` under "Alternatives considered" — wrong in a way no reader can detect, since only the
// `[deferred]` tag on the row contradicts the heading above it. Narrowing the pair instead makes
// them vanish from the report entirely, which is the failure the status enum's own Why warns about
// ("it silently vanishes from the section that exists to show the roads not taken").
//
// # `proposed` is a research area
//
// Excluding `proposed` from every section would argue that a heading announces a fate and an
// undecided inquiry has none; but the fuzz invariant `prose absent from report` holds, and a
// seat's recorded reasoning must reach the reader. A fourth section is no answer either. Three is
// the decision — an inquiry blue put forward IS an area this run is researching, and the PASS it
// holds while open is what stops it sitting there undecided, rather than a heading that describes
// the omission.
//
// setAside stays a COMPLEMENT so a status no predicate names cannot silently match nothing. The
// fate `refuted` is NOT set aside: it is named in Ran, with the other approaches that ran.
var InquirySections = []InquirySection{
	{Heading: "Research areas", Holds: taken},
	{Heading: "Future research directions", Holds: deferred},
	{Heading: "Alternatives considered", Holds: setAside},
}

// InquirySectionOf is the heading of the section that prints an inquiry whose fate is status.
func InquirySectionOf(status string) string {
	for _, s := range InquirySections {
		if s.Holds(status) {
			return s.Heading
		}
	}
	return "" // unreachable: setAside is the complement of the other two
}

func taken(status string) bool { return Ran(status) || status == "proposed" }

// Ran is the fates of an approach that RAN: its reason is what it FOUND, where the reason of any
// other is WHY it stands where it does.
func Ran(status string) bool {
	return status == "underway" || status == "confirmed" || status == "refuted"
}

func deferred(status string) bool { return status == "deferred" }

func setAside(status string) bool { return !taken(status) && !deferred(status) }
