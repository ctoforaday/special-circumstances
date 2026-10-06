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

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
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

// IsSpace is the whitespace a quote's separator matches.
func IsSpace(b byte) bool {
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

// Visible is a span of report text with the invisible layer removed — the skeleton any comparison
// of two such spans must run on.
//
// IT IS THE SAME REDUCTION LocateSpan MATCHES THROUGH, exported rather than reimplemented. Three
// other sites compared two quotes of the report on RAW BYTES and each stopped matching the moment a
// sentence gained an anchor: the `gap_edit` view, CurrentLocation's fold, and EstoppelConflict.
// Minting PLACES an anchor at the location a gap names, so those comparisons broke on the act that
// created the thing being compared — and broke silently, because "different spans" is also the
// honest answer for two unrelated sentences.
//
// LocateSpan answers "where in this report is this quote" and must return RAW offsets, so it walks
// the layer in a streaming pass. This answers "are these two quotes the same text", where offsets
// are meaningless and a reduced string is the whole answer. One definition of the layer
// (annotationLen), two questions.
//
// IT IS NOT THE RIGHT REDUCTION FOR EVERY READER, and the boundary is the trailing-punctuation trim.
// Dropping it is right when comparing two quotes of one sentence, because a seat may or may not carry
// the full stop. It is wrong where the text is matched against a seat's own PATTERN — the view
// selectors keep their own, narrower reduction for that reason (internal/cli/seat/selector.go), since
// `--match "negligible[.]"` is a pattern that names the punctuation this would remove.
func Visible(s string) string { return normalizeQuote(s) }

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
		if IsSpace(q[i]) {
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
			if ri >= len(report) || !IsSpace(report[ri]) {
				return 0, 0, false // else-branch: report content where the quote wants a space → fail
			}
			newlines := 0
			for ri < len(report) {
				if n := annotationLen(report, ri); n > 0 {
					ri += n
					continue
				}
				if !IsSpace(report[ri]) {
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

// LocateOnce is the one write-time matcher: the span of the quote's only occurrence in doc. It
// tries the refusals in a fixed order — absent, then a second occurrence, then (StopAtParagraph)
// a blank line inside the one match, then a split word — so a quote that also occurs inside one
// paragraph is told it repeats, never that it may not cross.
func LocateOnce(doc, quote string, scope SpanScope) (start, end int, err error) {
	start, end = locate(doc, quote, CrossParagraphs)
	if start < 0 {
		return -1, -1, ErrMisQuote
	}
	if s2, _ := locate(doc[end:], quote, CrossParagraphs); s2 >= 0 {
		return -1, -1, ErrAmbiguous
	}
	if _, _, ok := matchFrom(doc, start, normalizeQuote(quote), scope); !ok {
		return -1, -1, ErrCrossesParagraph
	}
	if !SpanBoundaryOK(doc, start, end) {
		return -1, -1, ErrSplitsWord
	}
	return start, end, nil
}

// SpanBoundaryOK rejects only a span that SPLITS A WORD.
//
// It is deliberately not a whitespace rule: normalizeQuote trims trailing punctuation, so a
// strict whitespace boundary would reject every sentence-final edit — measured, not feared.
func SpanBoundaryOK(s string, start, end int) bool {
	word := func(b byte) bool {
		return b == '_' || (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}
	splits := func(i int) bool { return i > 0 && i < len(s) && word(s[i-1]) && word(s[i]) }
	return !splits(start) && !splits(end)
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

// LocateOnce's refusals after ErrMisQuote. Callers map them to their own text.
var (
	ErrAmbiguous        = errors.New("the quote occurs more than once")
	ErrCrossesParagraph = errors.New("the quote crosses a blank line")
	ErrSplitsWord       = errors.New("the quote starts or ends inside a word")
)

// InsertAnchor is the shared invisible-anchor placement behind lens finding and blue cite:
// it extracts the quote from `location`, locates it in report across the invisible
// annotation layer, and returns report with `marker` spliced at the quote's end — or
// ErrMisQuote / ErrInFence. marker is the full token, "<!--fx:f-…-->" for a finding or
// "<!--cite:c-…-->" for a citation. Both axes place their immortal anchor by exactly this
// rule, so a citation and a finding are located, fenced, and spliced identically.
func InsertAnchor(report []byte, location, marker string) ([]byte, error) {
	// THE WHOLE LOCATION FIRST, AND THE EXTRACTED SPAN ONLY IF IT IS NOT THERE.
	//
	// extractQuote prefers the text between the first two double quotes, which is what makes
	// `§ Foundations: "the scheduler is preemptive"` locatable — a section label followed by its
	// sentence. Applied unconditionally it also fires on a location that IS the sentence and
	// merely CONTAINS a quoted phrase, and then the anchor is placed on the phrase somewhere else
	// in the document:
	//
	//	location: Blue wrote that the cost is "climbing sharply" and gave no source for it.
	//	anchored: The cost is climbing sharply<!--fx:f-1-->.        <- a different paragraph
	//
	// Silently, exit 0, on the surface whose whole job is to say WHERE the evidence is (#552,
	// measured on the 2026-08-23 research-loop-counterparts run: a long exact uniquely-matching
	// quote landed on an unrelated shorter sentence while a shorter quote from the same target
	// landed correctly — which is this, from the other side).
	//
	// So the choice is not made on the SHAPE of the string. It is made on what the report
	// actually contains: the literal location is tried first, and the extracted span is a
	// FALLBACK for the decorated form, which by construction is not in the report literally.
	// A miss is still ErrMisQuote — the marker never lands on content that is not there.
	quote := strings.TrimSpace(location)
	end := locateEnd(string(report), quote)
	if end < 0 {
		if q := extractQuote(location); q != quote {
			quote = q
			end = locateEnd(string(report), quote)
		}
	}
	if end < 0 {
		return nil, ErrMisQuote
	}
	if insideFence(string(report), end) {
		return nil, ErrInFence
	}
	return insertMarker(report, end, marker), nil
}

// Attach is the one write-time placement: doc with the anchor id's token placed at the end of
// the quote, by InsertAnchor's rule. Every placing verb validates its anchor through it and builds
// no token of its own. It returns InsertAnchor's sentinels unmapped, so each verb keeps its own
// refusal text.
func Attach(doc, id, quote string) (string, error) {
	out, err := InsertAnchor([]byte(doc), quote, anchor.Token(id))
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// insideFence reports whether byte offset `at` falls inside a fenced code block, its opener and
// closer lines included, as anchor.Blocks reads one. A marker must never land in code (it would
// ship literally / corrupt the fence).
func insideFence(report string, at int) bool {
	for _, b := range anchor.Blocks(report) {
		if b.Kind == anchor.Fence && b.Start <= at && at <= b.End {
			return true
		}
	}
	return false
}

// OrphanAnchorAt (torn-splice adoption) lived here: the three splicing verbs mutated
// blue/report.md first and appended their event second, so a crash between the two left an anchor
// no event backed, and a retry adopted it rather than splice a rival. Under report-as-record (#709)
// a marker exists ONLY as its event — there is no file to tear from the append — so the orphan
// state cannot arise and the adoption is gone with its callers.
