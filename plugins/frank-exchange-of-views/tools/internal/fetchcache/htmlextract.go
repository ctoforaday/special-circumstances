package fetchcache

import "fmt"

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
// styles, templates, comments, the no-script fallback, inline images, and the head (whose title
// is the Extraction's Title, as a PDF's metadata title is). It is read from the page's one parse
// (htmlpage.go) and makes no claim to render layout: a table's cells come out on one line,
// separated by spaces, and a `<pre>` block's own line breaks collapse like any other source
// whitespace.

// htmlExtractorID is this extraction's identity on the entry (Extractor). Bump it when the text it
// produces for the same bytes changes, so a text hash recorded under one rendering is never
// compared against another's.
const htmlExtractorID = "feov-html-text@2"

// extractHTML is the HTML arm of the cache's extraction.
func extractHTML(body []byte) Extraction {
	out := Extraction{Attempted: true, ExtractorID: htmlExtractorID}
	p := parseHTML(body)
	out.Title = p.title
	out.Text = htmlText(p)
	if out.Text == "" {
		// A REASON, because an empty extraction of a page reads exactly like a page with nothing
		// to say: the absence and its cause are the record, and no .txt is written.
		out.Reason = fmt.Sprintf("the page carries no text outside its markup, scripts and styles (%d bytes delivered) — "+
			"an app that renders client-side, or a page with nothing in it", len(body))
	}
	return out
}

// htmlText is the readable text of an HTML page: see this file's header for what it keeps.
func htmlText(p *htmlPage) string {
	return normalizeExtracted(p.text + "\n")
}
