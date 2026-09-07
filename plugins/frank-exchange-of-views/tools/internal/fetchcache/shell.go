package fetchcache

import (
	"fmt"
	"regexp"
	"strings"
)

// SHELL DETECTION — the first rung of #668, and deliberately not the second.
//
// A single-page app answers `fetch` with a near-empty HTML skeleton, because nothing here runs
// JavaScript. Red can then "verify" a citation against that skeleton and the miss is
// indistinguishable from an honest check: bytes arrived, a 200 was returned, a sha was stored.
// That is the plausible zero, on the surface that decides whether a source says what a claim
// says it says.
//
// The remedy here is NOT to render. Headless rendering is a chromium-class dependency and a
// decision to price once the flag says how often shells actually bite — which is what this
// produces. Until then the fix is that the system SAYS it could not read the page, rather than
// presenting the skeleton as the source.
//
// # Why the ratio and not a byte threshold
//
// A short page is not a shell. A 400-byte article extract is a short article; 400 bytes of text
// under 180KB of script is a shell. The signal is the RELATIONSHIP between what was delivered and
// what could be read, so the test is a ratio with a floor — and the floor exists because the
// ratio is meaningless on genuinely tiny documents, where a 90% script share is one inline
// handler.
//
// # Why it is three-state
//
// `NotRenderable` is a pointer for the same reason `TextExtracted` is: nil means nobody asked —
// an older index line, or a content type this question does not apply to — and false means asked
// and answered no. A plain bool collapses those into the reading that flatters.
const (
	// shellTextFloor is the extracted-text size below which the ratio is consulted at all.
	// Above it a page has enough prose to be worth reading whatever the markup weighs.
	shellTextFloor = 2000
	// shellTextRatio is the share of the delivered bytes that must survive as text. A page
	// yielding less than this is mostly machinery.
	shellTextRatio = 0.02
)

// rootDivOnly matches a body whose entire content is one empty mount point — the canonical
// React/Vue/Angular skeleton, which carries no prose at all and is not caught by a ratio when
// the framework loads its script from elsewhere.
var rootDivOnly = regexp.MustCompile(`(?is)<body[^>]*>\s*(<div[^>]*>\s*</div>\s*|<noscript>.*?</noscript>\s*|<script[^>]*>.*?</script>\s*)+</body>`)

// visibleText strips markup and the two elements that carry no prose, so the ratio measures what
// a READER would get rather than what the parser was handed.
//
// IT STRIPS ITS OWN TEXT RATHER THAN TAKING THE EXTRACTOR'S, and that is not a duplicated
// responsibility: there is no HTML extractor. DefaultExtractor is a PDF extractor and returns
// Attempted=false for HTML on purpose ("for an HTML page [an empty result] would be a lie about a
// document that is entirely text"). A first cut of this keyed on that extraction and would have
// flagged EVERY HTML page as a skeleton — measured on two real captures, both yielding 0
// characters, one of them a page dense with prose. The synthetic tests passed it because they
// handed the text in directly.
var (
	markupTag   = regexp.MustCompile(`(?s)<[^>]*>`)
	scriptStyle = regexp.MustCompile(`(?is)<(script|style)[^>]*>.*?</(script|style)>`)
)

func visibleText(body []byte) string {
	s := scriptStyle.ReplaceAllString(string(body), " ")
	s = markupTag.ReplaceAllString(s, " ")
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// ShellReason reports why a response looks like an unrendered app skeleton rather than a
// document, or "" when it does not.
//
// It answers only for HTML. A PDF that extracts badly is a different fact with its own field
// (TextExtracted / TextReason), and folding the two would make "we could not read this scan" and
// "this page needs a browser" the same sentence to a seat who has to act differently on each.
func ShellReason(contentType string, body []byte) string {
	if !strings.Contains(contentType, "html") {
		return ""
	}
	text := visibleText(body)
	if rootDivOnly.Match(body) {
		return fmt.Sprintf("the page body is an empty mount point with no prose in it — the markup a client-side app "+
			"renders into, and nothing here runs its JavaScript (%d bytes delivered, %d characters of text)", len(body), len(text))
	}
	if len(text) >= shellTextFloor || len(body) == 0 {
		return ""
	}
	if ratio := float64(len(text)) / float64(len(body)); ratio < shellTextRatio {
		return fmt.Sprintf("only %d characters of text survived %d bytes of markup (%.1f%%) — too little to be "+
			"the document this URL names. Nothing here runs JavaScript and nothing here has a subscription, so "+
			"an app that renders client-side and a page behind a paywall arrive looking the same; either way "+
			"these bytes are a record that the source EXISTS, not its text",
			len(text), len(body), ratio*100)
	}
	return ""
}
