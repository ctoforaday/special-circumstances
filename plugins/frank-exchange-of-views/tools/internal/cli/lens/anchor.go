package lens

import (
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

// trailingPunct is the run of punctuation trimmed from the END of a quote (only the
// end): a quote may omit or include a terminal period the report has, or vice versa.
// INTERNAL punctuation is CONTENT — never collapsed — so two sentences that differ only
// by an internal mark ("costs rise, sharply" vs "costs rise sharply") stay distinct.
const trailingPunct = ".,;:!?\"'…"

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// annotationLen returns the byte length of an annotation span starting at s[i:] — an
// invisible finding-marker "<!--fx:…-->" or a footnote reference "[^label]" — or 0 if
// none starts there. These spans are the "invisible layer": the matcher skips them so a
// quote of the semantic prose still anchors even when markers/footnotes are spliced in.
func annotationLen(s string, i int) int {
	if strings.HasPrefix(s[i:], "<!--fx:") {
		if j := strings.Index(s[i:], "-->"); j >= 0 {
			return j + 3
		}
	}
	if i+1 < len(s) && s[i] == '[' && s[i+1] == '^' {
		if j := strings.IndexByte(s[i+2:], ']'); j >= 0 {
			if !strings.ContainsRune(s[i+2:i+2+j], '\n') { // a real footnote ref is single-line
				return 2 + j + 1
			}
		}
	}
	return 0
}

// normalizeQuote reduces a quote to its matchable skeleton: annotation spans dropped,
// whitespace runs collapsed to a single space, TRAILING punctuation and whitespace
// trimmed. Internal punctuation survives as content. Returns "" for an empty/all-
// punctuation quote (which locateEnd rejects — never a false match at offset 0).
func normalizeQuote(q string) string {
	var out []byte
	for i := 0; i < len(q); {
		if n := annotationLen(q, i); n > 0 {
			i += n
			continue
		}
		if isSpace(q[i]) {
			if len(out) > 0 && out[len(out)-1] != ' ' {
				out = append(out, ' ')
			}
			i++
			continue
		}
		out = append(out, q[i])
		i++
	}
	s := strings.TrimRight(string(out), " ")
	s = strings.TrimRight(s, trailingPunct)
	return strings.TrimRight(s, " ")
}

// locateEnd returns the byte offset in report just past the last CONTENT char of the
// quote, or -1 if the quote is not present. It is an EXACT match run against the report
// minus its invisible annotation layer — the same exact-match semantics the Edit tool
// uses, with markers/footnotes skipped and whitespace runs treated as a single
// separator. A single streaming pass: the report cursor IS the raw insert offset (no
// offset-map). -1 == a mis-quote. No fuzzy/edit-distance matching (see docs/finding-
// markers.md and the #248 spec).
func locateEnd(report, quote string) int {
	nq := normalizeQuote(quote)
	if nq == "" { // empty or all-trailing-punctuation → reject, never match at 0
		return -1
	}
	for start := 0; start < len(report); {
		if n := annotationLen(report, start); n > 0 { // never start inside an annotation
			start += n
			continue
		}
		if end, ok := matchFrom(report, start, nq); ok {
			return end
		}
		start++
	}
	return -1
}

// matchFrom attempts to match the normalized quote nq against report beginning at
// `start`. Returns the offset just past the last matched content char and true, or 0,
// false. Annotation spans in the report are skipped (consume no quote char); a quote
// separator matches a run of report whitespace; a quote content char must match exactly.
// A whitespace run containing a blank line (paragraph break) is a hard boundary — a
// single quoted sentence never spans one — so the match fails there.
func matchFrom(report string, start int, nq string) (int, bool) {
	ri := start
	lastContentEnd := -1
	for qi := 0; qi < len(nq); {
		if n := annotationLen(report, ri); n > 0 {
			ri += n
			continue
		}
		if nq[qi] == ' ' {
			// The quote wants a separator; the report must have whitespace here.
			if ri >= len(report) || !isSpace(report[ri]) {
				return 0, false // else-branch: report content where the quote wants a space → fail
			}
			newlines := 0
			for ri < len(report) {
				if n := annotationLen(report, ri); n > 0 {
					ri += n
					continue
				}
				if !isSpace(report[ri]) {
					break
				}
				if report[ri] == '\n' {
					newlines++
				}
				ri++
			}
			if newlines >= 2 {
				return 0, false // crossed a paragraph break
			}
			qi++
			continue
		}
		// A content char (letters, digits, internal punctuation) must match exactly.
		if ri >= len(report) || report[ri] != nq[qi] {
			return 0, false
		}
		ri++
		qi++
		lastContentEnd = ri
	}
	return lastContentEnd, true
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
