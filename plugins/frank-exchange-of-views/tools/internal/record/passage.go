package record

import (
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/reportheading"
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
// is the opposite of a hunk. The whole report, and a lens's full re-read, arrive on the work list as
// report text; `show report` stays a read that records nothing.
//
// THE BOUND IS GENEROUS ON PURPOSE (gblock's call). Measured on that run's report — 12,348
// characters over 21 sections — the median section is 372 characters and the largest is 2,795. At
// 6,000 the bound never bites on real content; it exists so that one pathological section cannot
// hand an auditor the whole document under another name. Clipping to save bytes would reintroduce
// exactly the defect the constitution warns about, which is why the trimmed case SAYS it was
// trimmed rather than looking like a complete passage.
const passageLimit = 6000

// PassageAround is the report section offset at sits in — a gap's anchor: from the heading above it
// to the next heading at the same level or shallower. A gap whose anchor stands at no prose has no
// passage, and its reader reads the report.
func PassageAround(report string, at int) string {
	from, level := reportheading.SectionStart(report, at)
	to := reportheading.SectionEnd(report, at, level)
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
	lo := at - from - passageLimit/2
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
