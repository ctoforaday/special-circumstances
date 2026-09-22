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
// AN ACCESS CHALLENGE IS THE SAME PLAUSIBLE ZERO AND IS FAR MORE COMMON. Measured by fetching 22
// major scholarly sources through this tool: four answered with a wall rather than a paper, and
// only one of them — PubMed Central, by its mount-point shape — was caught. JSTOR, OpenReview and
// INSPIRE went through as documents. The rule had been fitted to the one case that had been seen.
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
	// shellProseFloor is the count of visible characters below which an HTML response is not
	// the document its url names, whatever its ratio says.
	//
	// THE RATIO ALONE MISSED THREE WALLS IN FOUR. Measured over 22 major scholarly sources
	// fetched through this tool: JSTOR answered 3,038 bytes carrying 303 characters titled
	// "Client Challenge", OpenReview 4,787 bytes carrying 288 titled "Verifying your browser",
	// and INSPIRE 2,896 bytes carrying 190 of "You need to enable JavaScript to run this app".
	// All three pass a 2% ratio — they are 10%, 6% and 7% — because a challenge page is mostly
	// PROSE about being a challenge page. Only PubMed Central was caught, by the mount-point
	// shape, and the ratio it would have needed to catch the others is low enough to flag real
	// documents.
	//
	// The floor separates them cleanly in that corpus: the four walls carry 165, 190, 288 and
	// 303 characters, and the smallest REAL document page carries 1,709 (a NeurIPS abstract
	// page), with CVF's at 2,447. 500 sits between with 1.6x headroom below the smallest
	// document and 5x above the largest wall. It is a floor on PROSE, never on bytes: a wall is
	// not small, it is empty, and those are different measurements.
	//
	// A RATIO CANNOT DO THIS JOB ALONE, and the numbers say why: the NeurIPS document's share is
	// 18.4% against JSTOR's wall at 10.0%, which is not a gap anything can be thresholded on.
	// The character count separates the same pair 1,709 to 303. So the count leads and the ratio
	// only excuses.
	shellProseFloor = 500
	// shellDenseRatio excuses a page that is SMALL rather than EMPTY. A wall spends most of its
	// bytes on machinery to run a check; a genuinely tiny page spends them on its few words, and
	// nothing was withheld from it. Measured on the same corpus the walls run 0.8% to 10.0% while
	// a 29-byte page carrying 16 characters runs 55%, so 25% sits between with 2.5x either way.
	// Without this a short page is refused for being short, which is the direction that
	// manufactures an absence.
	shellDenseRatio = 0.25
)

// challengeTitle and challengeBody name what a challenge SAYS about itself, so the refusal can
// state which wall this is.
//
// ATTRIBUTION IS A REFINEMENT OVER A SHAPE TEST, NEVER THE TEST ITSELF, and these are consulted
// ONLY for a response the shape has already refused. A vendor list is a pattern where no record
// exists — foreign bytes, nobody's schema — so it rots the moment a vendor rewords, and a rotted
// list must not turn a wall back into a document. The shape catches the CLASS; these only name
// the INSTANCE.
//
// ASKING THEM FIRST IS A MEASURED DEFECT, NOT A HYPOTHETICAL ONE. A draft of this file tested the
// markers before the shape and flagged bioRxiv's article page — 116,233 bytes carrying 8,619
// characters of real paper — as a wall, because Cloudflare serves its bot-management script
// INSIDE pages it has decided to deliver. `/cdn-cgi/challenge-platform` in a body means a
// challenge was considered, never that one was imposed. Refusing a readable source is the worse
// error of the two: a wall wrongly accepted is caught at the leaf when a seat reads it, while a
// document wrongly refused becomes a fabricated absence nothing downstream can question.
//
// What the name buys is the seat's next act, and it is not cosmetic — an app skeleton means
// nothing will render without a browser, while a challenge usually means the host publishes a
// sanctioned machine route to the same text, and those license opposite decisions.
var (
	challengeTitle = regexp.MustCompile(`(?is)<title[^>]*>[^<]*(client challenge|just a moment|verifying your browser|checking your browser|attention required|human verification|access denied|one moment)`)
	challengeBody  = regexp.MustCompile(`(?i)POW_CHALLENGE|/cdn-cgi/challenge-platform|__cf_chl|g-recaptcha|hcaptcha|turnstile`)
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
	if !strings.Contains(contentType, "html") || len(body) == 0 {
		return ""
	}
	text := visibleText(body)

	// THE SHAPE DECIDES, AND IT DECIDES FIRST. Three ways a response can fail to be the document
	// its url names, in the order of how certain each is.
	mount := rootDivOnly.Match(body)
	starved := len(text) < shellProseFloor && float64(len(text))/float64(len(body)) < shellDenseRatio
	// The ratio only has an opinion between the prose floor and the point where there is plainly
	// enough prose to read. A page that is almost entirely prose about not being a page passes it,
	// which is why it is no longer the only test.
	thin := len(text) < shellTextFloor && float64(len(text))/float64(len(body)) < shellTextRatio
	if !mount && !starved && !thin {
		return ""
	}

	// REFUSED. Now say which wall, where the page says so itself.
	if challengeTitle.Match(body) || challengeBody.Match(body) {
		return fmt.Sprintf("this is an ACCESS CHALLENGE, not the document — a wall asking for a browser to run its "+
			"script, solve its puzzle or set its cookie, which nothing here does (%d bytes delivered, %d characters "+
			"of text). It is a fact about how this host treats automated clients, NOT about whether the source exists "+
			"or is open: many hosts that challenge a browser path publish a sanctioned machine route to the same "+
			"text. Look for one before recording this source as unreachable, and never solve the challenge",
			len(body), len(text))
	}
	if mount {
		return fmt.Sprintf("the page body is an empty mount point with no prose in it — the markup a client-side app "+
			"renders into, and nothing here runs its JavaScript (%d bytes delivered, %d characters of text)", len(body), len(text))
	}
	if starved {
		return fmt.Sprintf("only %d characters of text are in this page, %.1f%% of the %d bytes delivered — too "+
			"little for the document this URL names. Something stands between this container and the source: a "+
			"wall, a login, or an app that renders client-side. These bytes are a record that the URL ANSWERED, "+
			"not that the source was read",
			len(text), float64(len(text))/float64(len(body))*100, len(body))
	}
	return fmt.Sprintf("only %d characters of text survived %d bytes of markup (%.1f%%) — too little to be "+
		"the document this URL names. Nothing here runs JavaScript and nothing here has a subscription, so "+
		"an app that renders client-side and a page behind a paywall arrive looking the same; either way "+
		"these bytes are a record that the source EXISTS, not its text",
		len(text), len(body), float64(len(text))/float64(len(body))*100)
}
