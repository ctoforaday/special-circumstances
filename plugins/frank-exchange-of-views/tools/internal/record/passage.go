package record

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
)

// THE GAP ARRIVES WITH ITS PASSAGE, SO THE AUDITOR IS NOT NAVIGATING (#1091).
//
// A gap's `location` is the challenged SENTENCE. Reading it tells an auditor what is disputed and
// nothing about whether the dispute survives its surroundings — so every auditor rendered the whole
// report to find out. Measured on the 2026-09-20 run: 166 full renders, 562,968 characters, 34% of
// every byte agents read all run, and 121 of those renders repeated a view the same agent already
// had.
//
// THIS IS NOT A DIFF, AND THE DISTINCTION IS THE WHOLE DESIGN. The adversary constitution says a
// change-summary is a navigation hint and never the audit surface, because "decontextualized diffs
// mislead on research prose" — a paragraph that reads as an overclaim alone may be qualified two
// paragraphs later, and a citation that looks unsupported may be carried by the preceding sentence.
// That rule is correct. A SECTION is what it actually asks for: the passage IN ITS CONTEXT, which
// is the opposite of a hunk. `show report` stays available and unchanged for the full re-read.
//
// THE BOUND IS GENEROUS ON PURPOSE (gblock's call). Measured on that run's report — 12,348
// characters over 21 sections — the median section is 372 characters and the largest is 2,795. At
// 6,000 the bound never bites on real content; it exists so that one pathological section cannot
// hand an auditor the whole document under another name. Clipping to save bytes would reintroduce
// exactly the defect the constitution warns about, which is why the trimmed case SAYS it was
// trimmed rather than looking like a complete passage.
const passageLimit = 6000

// PassageAround is the report section the quote sits in: from the heading above it to the next
// heading at the same level or shallower.
//
// EMPTY WHEN THE QUOTE CANNOT BE FOUND, and that is the honest answer rather than a guess at a
// neighbourhood. A gap can be anchored to something that is not report text at all (about_kind), a
// quote can have been edited away, and a run can have no ingested report yet. In each case the
// auditor gets no passage and reads the report as it does today — never a passage built around a
// location the text does not actually hold.
func PassageAround(report, quote string) string {
	if report == "" || strings.TrimSpace(quote) == "" {
		return ""
	}
	start, end := anchortext.LocateSpan(report, quote)
	if start < 0 || end <= start || end > len(report) {
		return ""
	}
	from, level := sectionStart(report, start)
	to := sectionEnd(report, end, level)
	sec := strings.TrimSpace(report[from:to])
	if len(sec) <= passageLimit {
		return sec
	}
	// A section past the bound: keep the heading and the quote's own neighbourhood, and say so.
	// A silently shortened passage is the decontextualized hunk this whole change refuses.
	head := sec
	if i := strings.IndexByte(sec, '\n'); i >= 0 {
		head = sec[:i]
	}
	lo := start - from - passageLimit/2
	if lo < 0 {
		lo = 0
	}
	hi := lo + passageLimit
	if hi > len(sec) {
		hi, lo = len(sec), len(sec)-passageLimit
	}
	return head + "\n\n[…this section runs to " + itoaLen(len(sec)) +
		" characters and is shown in part, around the quoted text; the whole report is still one read away…]\n\n" +
		strings.TrimSpace(sec[lo:hi])
}

// sectionStart is the offset of the heading line governing at, and that heading's level. Level 0
// with offset 0 means the quote sits before any heading — the report's preamble is its section.
func sectionStart(report string, at int) (int, int) {
	for i := lineStart(report, at); ; {
		if lvl := headingLevel(report, i); lvl > 0 {
			return i, lvl
		}
		if i == 0 {
			return 0, 0
		}
		i = lineStart(report, i-1)
	}
}

// sectionEnd is the offset of the next heading at the same level or shallower, or the end.
func sectionEnd(report string, from, level int) int {
	// LEVEL 0 IS "NO HEADING GOVERNS THIS", which is text before the document's first heading. It
	// ends at the NEXT heading of any level — defaulting it to 1 made a `##` fail to stop it, so a
	// quote in an unheaded preamble swallowed the first real section.
	if level == 0 {
		level = 6
	}
	for i := from; i < len(report); {
		nl := strings.IndexByte(report[i:], '\n')
		if nl < 0 {
			return len(report)
		}
		i += nl + 1
		if lvl := headingLevel(report, i); lvl > 0 && lvl <= level {
			return i
		}
	}
	return len(report)
}

// headingLevel is the number of leading '#' on the line beginning at i, or 0 if it is not an ATX
// heading. A '#' not followed by a space is not one — `#1091` in prose is not a section.
func headingLevel(report string, i int) int {
	n := 0
	for i+n < len(report) && report[i+n] == '#' {
		n++
	}
	if n == 0 || n > 6 || i+n >= len(report) || report[i+n] != ' ' {
		return 0
	}
	return n
}

func lineStart(s string, at int) int {
	if at > len(s) {
		at = len(s)
	}
	if i := strings.LastIndexByte(s[:at], '\n'); i >= 0 {
		return i + 1
	}
	return 0
}

func itoaLen(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
