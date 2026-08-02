package lens

import (
	"regexp"
	"strings"
)

// Finding-marker anchoring (slice 1b). A lens finding is anchored in blue/report.md by
// inserting an invisible footnote marker "[^<finding_id>]" at the quoted sentence the
// finding flags. The quote comes from --location ("section heading + quoted sentence").
// If the quote is not in the report, the finding is a MIS-QUOTE and is rejected — the
// marker is never placed on content that is not there.

// extractQuote pulls the locatable sentence out of a --location value. Prefer an
// explicit "…" span (the quoted sentence); otherwise use the whole trimmed value.
func extractQuote(location string) string {
	if i := strings.Index(location, `"`); i >= 0 {
		if j := strings.Index(location[i+1:], `"`); j >= 0 {
			if q := strings.TrimSpace(location[i+1 : i+1+j]); q != "" {
				return q
			}
		}
	}
	return strings.TrimSpace(location)
}

// locateEnd returns the byte offset just PAST the quote in report, or -1 if the quote
// is not present. Exact substring first, then a whitespace-flexible match (tokens
// separated by \s+) so trivially reflowed spacing still anchors. -1 == a mis-quote.
func locateEnd(report, quote string) int {
	if quote == "" {
		return -1
	}
	if i := strings.Index(report, quote); i >= 0 {
		return i + len(quote)
	}
	toks := strings.Fields(quote)
	if len(toks) == 0 {
		return -1
	}
	for k := range toks {
		toks[k] = regexp.QuoteMeta(toks[k])
	}
	re, err := regexp.Compile(strings.Join(toks, `\s+`))
	if err != nil {
		return -1
	}
	if loc := re.FindStringIndex(report); loc != nil {
		return loc[1]
	}
	return -1
}

// insertMarker splices marker into report at byte offset `at`.
func insertMarker(report []byte, at int, marker string) []byte {
	out := make([]byte, 0, len(report)+len(marker))
	out = append(out, report[:at]...)
	out = append(out, marker...)
	out = append(out, report[at:]...)
	return out
}

// insideFence reports whether byte offset `at` falls inside a ``` / ~~~ fenced code
// block. A marker must never land in code (it would ship literally / corrupt the fence).
func insideFence(report string, at int) bool {
	fence := false
	pos := 0
	for _, ln := range strings.SplitAfter(report, "\n") {
		start := pos
		pos += len(ln)
		trimmed := strings.TrimLeft(ln, " \t")
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = !fence
			continue
		}
		if at >= start && at < pos {
			return fence
		}
	}
	return false
}
