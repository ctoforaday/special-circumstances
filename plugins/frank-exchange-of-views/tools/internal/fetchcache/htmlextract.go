package fetchcache

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

// AN HTML PAGE IS READ AS TEXT, AND THE RAW BYTES STAY THE SOURCE.
//
// The cache kept an HTML page as its bytes alone, so a seat searching what a source says searched
// markup: in smoke 07d nine sittings ran a pattern over tag soup and met the search tool's
// repetition limit before they met the sentence. A PDF already had its text beside it; a page now
// does too, through the same fold (foldExtraction) and the same `<sha>.txt` path, so a reader that
// knows where a PDF's text is knows where a page's is. The body's sha still certifies the bytes
// fetched; the text is a rendering of them, hashed separately on the entry.
//
// WHAT IT KEEPS IS WHAT A READER OF THE PAGE GETS: the text of the body, with a line break where
// the page breaks a block, entities decoded, and nothing a browser does not show — scripts,
// styles, templates, comments, inline images, and the head (whose title is the Extraction's Title,
// as a PDF's metadata title is). It is a regular-expression pass, not a parser, and makes no claim
// to render layout: a table's cells come out on one line, separated by spaces, and a `<pre>` block's
// own line breaks collapse like any other source whitespace.

// htmlExtractorID is this extraction's identity on the entry (Extractor). Bump it when the text it
// produces for the same bytes changes, so a text hash recorded under one rendering is never
// compared against another's.
const htmlExtractorID = "feov-html-text@1"

// attrs matches a tag's attributes up to its closing `>`, stepping over a `>` inside a quoted
// value — Wikipedia's `data-mw` attributes carry whole fragments of markup, and a plain `[^>]*`
// ends the tag inside one and leaves the rest of it in the text.
const attrs = `(?:[^>"']|"[^"]*"|'[^']*')*`

var (
	htmlComment = regexp.MustCompile(`(?s)<!--.*?-->`)
	// One expression per element, because Go's regexp has no back-reference to close the element
	// it opened: a shared alternation would let `<script>` end at a `</style>`.
	htmlHidden = []*regexp.Regexp{
		regexp.MustCompile(`(?is)<head\b` + attrs + `>.*?</head\s*>`),
		regexp.MustCompile(`(?is)<script\b` + attrs + `>.*?</script\s*>`),
		regexp.MustCompile(`(?is)<style\b` + attrs + `>.*?</style\s*>`),
		regexp.MustCompile(`(?is)<noscript\b` + attrs + `>.*?</noscript\s*>`),
		regexp.MustCompile(`(?is)<template\b` + attrs + `>.*?</template\s*>`),
		regexp.MustCompile(`(?is)<svg\b` + attrs + `>.*?</svg\s*>`),
	}
	htmlBlock = regexp.MustCompile(`(?i)</?(p|div|br|li|ul|ol|dl|dd|dt|h[1-6]|table|thead|tbody|tfoot|tr|caption|section|article|header|footer|nav|aside|main|blockquote|pre|figure|figcaption|hr|form|fieldset|address|details|summary)\b` + attrs + `>`)
	htmlCell  = regexp.MustCompile(`(?i)</?(td|th)\b` + attrs + `>`)
	htmlTag   = regexp.MustCompile(`<` + attrs + `>`)
	htmlTitle = regexp.MustCompile(`(?is)<title\b` + attrs + `>(.*?)</title\s*>`)
	// anySpace is HTML's own whitespace rule: a run of spaces, tabs and line breaks in the source
	// is one space on the page. Line breaks come from the blocks, never from the source's layout —
	// a pretty-printed formula otherwise lands one token to a line.
	anySpace = regexp.MustCompile(`[ \t\n\r\f\v]+`)
	hSpace   = regexp.MustCompile(`[ \t\f\v\x{00a0}]+`)
)

// extractHTML is the HTML arm of the cache's extraction.
func extractHTML(body []byte) Extraction {
	out := Extraction{Attempted: true, ExtractorID: htmlExtractorID}
	if m := htmlTitle.FindSubmatch(body); m != nil {
		out.Title = strings.Join(strings.Fields(html.UnescapeString(htmlTag.ReplaceAllString(string(m[1]), ""))), " ")
	}
	out.Text = htmlText(body)
	if out.Text == "" {
		// A REASON, because an empty extraction of a page reads exactly like a page with nothing
		// to say: the absence and its cause are the record, and no .txt is written.
		out.Reason = fmt.Sprintf("the page carries no text outside its markup, scripts and styles (%d bytes delivered) — "+
			"an app that renders client-side, or a page with nothing in it", len(body))
	}
	return out
}

// htmlText is the readable text of an HTML page: see this file's header for what it keeps.
func htmlText(body []byte) string {
	s := htmlComment.ReplaceAllString(string(body), " ")
	for _, re := range htmlHidden {
		s = re.ReplaceAllString(s, " ")
	}
	s = anySpace.ReplaceAllString(s, " ")
	s = htmlBlock.ReplaceAllString(s, "\n")
	s = htmlCell.ReplaceAllString(s, " ")
	s = htmlTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	var b strings.Builder
	blank := true // no leading blank line
	for _, line := range strings.Split(strings.ReplaceAll(s, "\r", "\n"), "\n") {
		line = strings.TrimSpace(hSpace.ReplaceAllString(line, " "))
		if line == "" {
			if !blank {
				b.WriteString("\n")
				blank = true
			}
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
		blank = false
	}
	return normalizeExtracted(strings.TrimRight(b.String(), "\n") + "\n")
}
