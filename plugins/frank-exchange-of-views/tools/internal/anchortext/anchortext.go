// Package anchortext is the report-text geometry of immortal anchors: how a quoted span is
// LOCATED across the invisible annotation layer (LocateSpan and its scoped/unique variants) and
// how a marker is PLACED at that span (InsertAnchor). It is the sibling of
// internal/anchor — that leaf owns the anchor VOCABULARY (Token, Label, the class grammar), this
// one owns where an anchor SITS in the document.
//
// It lived in internal/cli/lens, which made it unreachable to any package cli/lens imports. Under
// report-as-record (#709) the report is REPLAYED from the record by internal/reportproj, which
// must re-place every marker — and cli/lens must read the replayed report to locate its own
// splice, so cli/lens now imports reportproj. Geometry in cli/lens would have closed that loop
// into a cycle. It depends on nothing but the standard library and the internal/anchor leaf, so
// reportproj, cli/lens, cli/blue and bluedoc can all share the one locator.
package anchortext

import (
	"errors"
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

// TrailingPunct is the same cutset, exported for the one other place that must skip the run a
// quote is allowed to omit: bluedoc, deciding whether a span abuts an anchor. Exported rather
// than restated there — two copies of "which marks a quote may drop" is precisely the pair that
// drifts, and the drift would show as an edit refused or permitted for no reason a reader could
// find.
const TrailingPunct = trailingPunct

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// annotationLen returns the byte length of an annotation span starting at s[i:] — an
// invisible HTML-comment anchor ("<!--fx:…-->" finding, "<!--cite:…-->" citation, or any
// other) or a footnote reference "[^label]" — or 0 if none starts there. These spans are
// the "invisible layer": the matcher skips them so a quote of the semantic prose still
// anchors even when markers/footnotes are spliced in. The whole HTML-comment class is
// skipped (not just fx) so a sentence already carrying a citation anchor is still locatable
// for a finding, and vice versa — the two anchor axes never block each other.
func annotationLen(s string, i int) int {
	if strings.HasPrefix(s[i:], "<!--") {
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
// quote, or -1 if not present — the insert-point half of LocateSpan, used by the
// finding-marker anchor (which needs only the end).
func locateEnd(report, quote string) int {
	_, end := LocateSpan(report, quote)
	return end
}

// LocateSpan returns the raw [start,end) byte span in report matched by quote — an EXACT
// match run against the report minus its invisible annotation layer (markers/footnotes
// skipped, whitespace runs treated as a single separator) — or (-1,-1) if the quote is not
// present. `start` is the first matched CONTENT byte; `end` is just past the last matched
// content byte (before any trailing annotation). A single streaming pass: the report cursor
// IS the raw offset (no offset-map). `blue edit` replaces report[start:end], leaving the
// invisible marker layer intact; a span that INTERNALLY contains a marker is the caller's to
// reject (edit around it). No fuzzy/edit-distance matching.
func LocateSpan(report, quote string) (int, int) {
	return locate(report, quote, StopAtParagraph)
}

// SpanScope says whether a quote may cross a blank line.
//
// ONE MATCHER SERVED TWO OPPOSITE REQUIREMENTS AND THAT WAS THE BUG. A finding ANCHORS a
// sentence, so a paragraph break is correctly a hard boundary — a marker that spanned one would
// sit in two places at once. An EDIT REPLACES a span, and a span across paragraphs is an ordinary
// thing to replace.
//
// Sharing the boundary made `blue edit`'s two rules jointly unsatisfiable, which is measurable
// rather than theoretical. Blue is told to propagate a correction to every site stating a claim,
// so the SAME SENTENCE IN TWO SECTIONS is the expected shape of a real report. Quote it alone and
// the tool says "appears MORE THAN ONCE — quote more surrounding context"; take that advice and
// the context crosses a blank line and the tool says "not found". A haiku seat hit that pair ten
// times in one sitting and gave up with the repair unmade, correctly reporting that multi-line
// spans were rejected "even with exact headers and context".
type SpanScope int

const (
	// StopAtParagraph refuses a quote that crosses a blank line. For anchors.
	StopAtParagraph SpanScope = iota
	// CrossParagraphs allows it. For replacements.
	CrossParagraphs
)

// LocateSpanScoped is LocateSpan with the boundary rule stated by the caller.
func LocateSpanScoped(report, quote string, scope SpanScope) (int, int) {
	return locate(report, quote, scope)
}

func locate(report, quote string, scope SpanScope) (int, int) {
	nq := normalizeQuote(quote)
	if nq == "" { // empty or all-trailing-punctuation → reject, never match at 0
		return -1, -1
	}
	for start := 0; start < len(report); {
		if n := annotationLen(report, start); n > 0 { // never start inside an annotation
			start += n
			continue
		}
		if s, e, ok := matchFrom(report, start, nq, scope); ok {
			return s, e
		}
		start++
	}
	return -1, -1
}

// matchFrom attempts to match the normalized quote nq against report beginning at `start`.
// Returns (spanStart, end, true) — spanStart = first matched content byte, end = just past
// the last matched content byte — or (0,0,false). Annotation spans in the report are skipped
// (consume no quote char); a quote separator matches a run of report whitespace; a quote
// content char must match exactly. A whitespace run containing a blank line (paragraph
// break) is a hard boundary — a single quoted sentence never spans one — so the match fails.
func matchFrom(report string, start int, nq string, scope SpanScope) (int, int, bool) {
	ri := start
	firstContent := -1
	lastContentEnd := -1
	for qi := 0; qi < len(nq); {
		if n := annotationLen(report, ri); n > 0 {
			ri += n
			continue
		}
		if nq[qi] == ' ' {
			// The quote wants a separator; the report must have whitespace here.
			if ri >= len(report) || !isSpace(report[ri]) {
				return 0, 0, false // else-branch: report content where the quote wants a space → fail
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
			if newlines >= 2 && scope == StopAtParagraph {
				return 0, 0, false // crossed a paragraph break, and this caller anchors
			}
			qi++
			continue
		}
		// A content char (letters, digits, internal punctuation) must match exactly.
		if ri >= len(report) || report[ri] != nq[qi] {
			return 0, 0, false
		}
		if firstContent < 0 {
			firstContent = ri
		}
		ri++
		qi++
		lastContentEnd = ri
	}
	return firstContent, lastContentEnd, true
}

// LocateSpanUniqueScoped reports whether the quote matches at MORE THAN ONE place, with the
// boundary rule stated by the caller — the AMBIGUITY verdict `blue edit`'s uniqueness guard needs
// (LocateSpan takes the first match and says nothing, which is silent mis-targeting). The
// boundary-free wrapper LocateSpanUnique went with report-as-record's torn-splice removal, its only
// caller; bluedoc still calls this scoped form.
func LocateSpanUniqueScoped(report, quote string, scope SpanScope) (start, end int, ambiguous bool) {
	start, end = LocateSpanScoped(report, quote, scope)
	if start < 0 {
		return start, end, false
	}
	// Look for a SECOND match beyond the first one's end.
	if rest := end; rest < len(report) {
		if s2, _ := LocateSpanScoped(report[rest:], quote, scope); s2 >= 0 {
			return start, end, true
		}
	}
	return start, end, false
}

// insertMarker splices marker into report at byte offset `at`.
func insertMarker(report []byte, at int, marker string) []byte {
	out := make([]byte, 0, len(report)+len(marker))
	out = append(out, report[:at]...)
	out = append(out, marker...)
	out = append(out, report[at:]...)
	return out
}

// ErrMisQuote and ErrInFence are the two ways an anchor placement is refused: the quote is
// not present (a mis-quote — never place a marker on content that is not there), or it
// resolves inside a code fence (a marker there would ship literally / corrupt the fence).
// Callers map them to a verb-specific message; the shared machinery only classifies.
var (
	ErrMisQuote = errors.New("the quoted content was not found in report.md")
	ErrInFence  = errors.New("the quote resolves inside a code fence")
)

// InsertAnchor is the shared invisible-anchor placement behind lens finding and blue cite:
// it extracts the quote from `location`, locates it in report across the invisible
// annotation layer, and returns report with `marker` spliced at the quote's end — or
// ErrMisQuote / ErrInFence. marker is the full token, "<!--fx:f-…-->" for a finding or
// "<!--cite:c-…-->" for a citation. Both axes place their immortal anchor by exactly this
// rule, so a citation and a finding are located, fenced, and spliced identically.
func InsertAnchor(report []byte, location, marker string) ([]byte, error) {
	quote := extractQuote(location)
	end := locateEnd(string(report), quote)
	if end < 0 {
		return nil, ErrMisQuote
	}
	if insideFence(string(report), end) {
		return nil, ErrInFence
	}
	return insertMarker(report, end, marker), nil
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

// OrphanAnchorAt (torn-splice adoption) lived here: the three splicing verbs mutated
// blue/report.md first and appended their event second, so a crash between the two left an anchor
// no event backed, and a retry adopted it rather than splice a rival. Under report-as-record (#709)
// a marker exists ONLY as its event — there is no file to tear from the append — so the orphan
// state cannot arise and the adoption is gone with its callers.
