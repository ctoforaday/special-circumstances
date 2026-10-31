package fetchcache

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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
// styles, templates, comments, the no-script fallback, inline images, and the head (whose title
// is the Extraction's Title, as a PDF's metadata title is). It is read from the page's one parse
// (htmlpage.go) and makes no claim to render layout: a table's cells come out on one line,
// separated by spaces, and a `<pre>` block's own line breaks collapse like any other source
// whitespace. The inline markup whose loss changes what the text says stays in the text
// (TextMarkup): the exponent of 3<sup>n</sup> is why "3n ≡ 3" is false and "3^n ≡ 3" is true.

// htmlExtractorID is this extraction's identity on the entry (Extractor). Bump it when the text it
// produces for the same bytes changes, so a text hash recorded under one rendering is never
// compared against another's.
const htmlExtractorID = "feov-html-text@3"

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

// TextMarkup is how a page's text carries the markup a reader of the page sees, in a seat's words:
// the one statement of it that fetch's, cite's and `show source`'s --help render.
const TextMarkup = "in an HTML page's text a list item opens with \"- \", a superscript reads ^n, or ^(n-1) when it is more than one number or one letter, " +
	"a subscript reads _n or _(n+1), and a linked footnote mark reads [1]; quote these as the text prints them"

// A SCRIPT IS PART OF WHAT THE TEXT SAYS, and the keyboard forms are what a seat can type and a
// search can find. A compound script is bracketed, because "a^n-1" and "a^(n-1)" are different
// numbers. Unicode's superscript characters cover too little to carry this, and SpanNormalisation
// has no row for them: a seat quotes "3^n" from the text it reads.
//
// WHAT READS AS PRINTED, from the marks on m37's cached pages and the labelled bodies in testdata
// (Wikipedia, Nature, Springer, PMC, ScienceDirect, Highwire):
//   - A REFERENCE MARK IS NOT AN EXPONENT. A span boundary reads past Wikipedia's
//     `<sup class="reference">[9]</sup>` after a sentence's end (pastReferences), and "^[9]" would
//     cut every such sentence. Springer Nature's linked `<sup><a href="#Aff1">1</a></sup>` gets the
//     same reading by being bracketed, each mark its own: "Smith1," is a word nobody wrote.
//   - An ordinal suffix is spelling: "4th" is what the page says, and what a seat types.
//   - A mark with no letter or digit, "*" or "†", sits on its word as printed.
//
// An unlinked `<sup>1</sup>` is an exponent or a footnote, and the markup does not say which. It
// keeps its caret: a reader told "cells^1." loses nothing, and one told "x2" for x^2 loses the
// number.

// superscript is the text of a <sup> whose content walked to inner.
func superscript(n *html.Node, inner string) string {
	text := strings.Join(strings.Fields(inner), " ")
	switch {
	case text == "":
		return ""
	case referenceClass(n) || strings.HasPrefix(text, "["):
		return text
	case linkedMark(n):
		parts := strings.Split(strings.ReplaceAll(text, " ", ""), ",")
		return "[" + strings.Join(parts, "][") + "]"
	case ordinalSuffix[text] || !hasLetterOrDigit(text):
		return text
	}
	return "^" + scriptOperand(text)
}

// subscript is the text of a <sub> whose content walked to inner.
func subscript(inner string) string {
	text := strings.Join(strings.Fields(inner), " ")
	if text == "" {
		return ""
	}
	return "_" + scriptOperand(text)
}

// scriptOperand brackets a script that is more than one number or one letter.
func scriptOperand(text string) string {
	if isDigits(text) || len([]rune(text)) == 1 && unicode.IsLetter([]rune(text)[0]) {
		return text
	}
	return "(" + text + ")"
}

var ordinalSuffix = map[string]bool{"st": true, "nd": true, "rd": true, "th": true}

func isDigits(s string) bool {
	return strings.Trim(s, "0123456789") == ""
}

func hasLetterOrDigit(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) })
}

// referenceClass reports the class MediaWiki sets on its footnote marks and page pointers.
func referenceClass(n *html.Node) bool {
	for _, a := range n.Attr {
		if a.Namespace == "" && a.Key == "class" && has(strings.Fields(a.Val), "reference") {
			return true
		}
	}
	return false
}

// linkedMark reports a <sup> whose words are all links to a place in a page, with at most commas
// and spaces between them: the link is what says it is a mark and not a power.
func linkedMark(n *html.Node) bool {
	links, other := 0, false
	var visit func(*html.Node)
	visit = func(c *html.Node) {
		switch {
		case c.Type == html.TextNode:
			other = other || strings.Trim(c.Data, ", \t\r\n\f") != ""
		case c.Type == html.ElementNode && c.DataAtom == atom.A && linkHasFragment(c):
			links++
		default:
			for k := c.FirstChild; k != nil; k = k.NextSibling {
				visit(k)
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		visit(c)
	}
	return links > 0 && !other
}

func linkHasFragment(a *html.Node) bool {
	for _, at := range a.Attr {
		if at.Namespace == "" && at.Key == "href" {
			return strings.Contains(at.Val, "#")
		}
	}
	return false
}

// listItem is the text of an <li> whose content walked to inner. The marker is what a reader sees,
// and it is what makes the item a unit of a text that holds no markup: a list item may open with a
// lowercase word, which after a blank line alone reads as a sentence carried on from before it
// (SpanBoundary). An item with no text has no line.
func listItem(inner string) string {
	text := strings.TrimSpace(inner)
	if text == "" {
		return ""
	}
	return "\n- " + text + "\n"
}
