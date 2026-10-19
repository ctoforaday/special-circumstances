package fetchcache

import (
	"net/url"
	"strings"
	"testing"
)

// AN ATTRIBUTE IS READ BY ITS NAME, WHEREVER IT SITS. Each pattern this parse replaced read one
// attribute order, so the same tag written content-first was a silent miss: a book's sales page
// cited as an abstract, a rights reservation recorded as none, a pointer to the full text unread.
func TestMetaTagsAreReadInEitherAttributeOrder(t *testing.T) {
	book := []byte(`<html><head><meta content="book" property="og:type"><meta content="t" name="citation_title"></head><body><p>x</p></body></html>`)
	if got, reason := Completeness("text/html", book, true); got != CompletenessNotTheWork {
		t.Errorf("a content-first og:type book got %q (%s)", got, reason)
	}
	tdm := []byte(`<html><head><meta content=1 name=tdm-reservation><meta content='https://pub.example/tdm' name="tdm-policy"></head></html>`)
	if reserved, policy := TDMReservation("text/html", tdm); !reserved || policy != "https://pub.example/tdm" {
		t.Errorf("a content-first, unquoted reservation read as reserved=%v policy=%q", reserved, policy)
	}
	base, _ := url.Parse("https://pub.example/article/1")
	land := []byte(`<html><head><meta content="/article/1/fulltext" name="citation_fulltext_html_url"></head></html>`)
	if got := LandingPageFullText("text/html", land, "", base); got != "https://pub.example/article/1/fulltext" {
		t.Errorf("a content-first full-text pointer read as %q", got)
	}
	// A value carrying an apostrophe is the value: the pattern stopped at it.
	apos := []byte(`<html><head><meta name="citation_pdf_url" content="/pdfs/o'neil.pdf"></head></html>`)
	if got := LandingPageFullText("text/html", apos, "", base); got != "https://pub.example/pdfs/o'neil.pdf" {
		t.Errorf("a pointer with an apostrophe read as %q", got)
	}
}

// A <link> IS READ BY ITS OWN ATTRIBUTES: `data-rel` is not `rel`, and a `>` inside a quoted value
// does not end the element.
func TestALinkElementIsReadByItsOwnAttributes(t *testing.T) {
	page := []byte(`<html><head>` +
		`<link data-rel="item" rel="stylesheet" href="/style.css">` +
		`<link title="a > b" rel="item" href="/paper.pdf" type="application/pdf">` +
		`</head></html>`)
	if got := linkElements(page); got != `</paper.pdf>; rel="item"` {
		t.Errorf("linkElements = %q, want only the item link", got)
	}
}

// THE PAGE'S TEXT IS WHAT A READER SEES. Comments, scripts and an inline image's title are not;
// text a pattern pass lost is: a "<!--" inside a script string ate the page up to the next "-->",
// and a bare "<" in prose ate it up to the next ">".
func TestThePageTextIsWhatAReaderSees(t *testing.T) {
	page := []byte(`<html><head><title>The Title</title></head><body>` +
		`<script>var s = "<!--";</script><p>first paragraph</p><!-- -->` +
		`<p>a < n is prime when gcd(n, a!) = 1</p>` +
		`<svg><title>Just a moment</title></svg>` +
		`<math><mi>x</mi></math>` +
		`<p>line one<br>line two</p><p>last -->paragraph</p></body></html>`)
	text := htmlText(page)
	for _, want := range []string{"first paragraph", "a < n is prime when gcd(n, a!) = 1", "x", "line one\nline two", "last -->paragraph"} {
		if !strings.Contains(text, want) {
			t.Errorf("the text lost %q:\n%s", want, text)
		}
	}
	for _, not := range []string{"var s", "Just a moment", "The Title"} {
		if strings.Contains(text, not) {
			t.Errorf("the text carries %q, which a reader does not see:\n%s", not, text)
		}
	}
	// A title the page put in its body is still its title, not its text.
	if text := htmlText([]byte(`<html><body><p>body prose</p><title>Stray Title</title></body></html>`)); strings.Contains(text, "Stray Title") {
		t.Errorf("a title outside the head was read as text:\n%s", text)
	}
	if ex := extractHTML(page); ex.Title != "The Title" {
		t.Errorf("Title = %q, want the document's own title and not the inline image's", ex.Title)
	}
	// An inline image's title is not a challenge page's title.
	if r := ShellReason("text/html", append([]byte(strings.Repeat("<p>prose of a real document</p>", 100)), page...)); r != "" {
		t.Errorf("a document whose inline svg says 'Just a moment' was refused: %s", r)
	}
}

// A BODY THAT BEGINS AND ENDS WITH A SCRIPT IS NOT A MOUNT POINT. The pattern this replaced let
// its script alternative run lazily from the first <script> to the last </script>, so any page
// opening and closing its body with one — pubs.acs.org, science.org, journals.sagepub.com in the
// captured runs, 12,000 to 700,000 characters of article — was refused as an empty app shell.
func TestABodyBetweenTwoScriptsIsNotAMountPoint(t *testing.T) {
	doc := []byte(`<html><body><script>var a=1;</script><main>` + strings.Repeat("<p>The article's own prose, sentence after sentence.</p>", 60) +
		`</main><script src="/app.js"></script></body></html>`)
	if p := readHTML(doc); p.mount {
		t.Error("a body carrying an article between two scripts was read as a mount point")
	}
	if r := ShellReason("text/html", doc); r != "" {
		t.Errorf("an article was refused: %s", r)
	}
	shell := []byte(`<html><body><div id="root"></div><noscript>You need to enable JavaScript to run this app.</noscript><script src="/app.js"></script></body></html>`)
	if p := readHTML(shell); !p.mount {
		t.Error("an empty mount point was not recognised")
	}
}

// ONE PARSE PER PAGE: the same bytes asked twice are parsed once, and different bytes of the same
// length are not mistaken for them.
func TestAPageIsParsedOnceWhileInUse(t *testing.T) {
	a := []byte(`<html><head><title>A</title></head></html>`)
	b := []byte(`<html><head><title>B</title></head></html>`)
	if readHTML(a) != readHTML(a) {
		t.Error("the same body was parsed twice")
	}
	if readHTML(b).title != "B" || readHTML(a).title != "A" {
		t.Error("two bodies of one length shared a parse")
	}
}

// A REFRESH DIRECTIVE IS READ BY THE STANDARD'S STEPS, not by one spelling of it.
func TestARefreshDirectiveIsReadByTheStandardsSteps(t *testing.T) {
	for content, want := range map[string]string{
		"2; url='/retrieve/pii/S1'": "/retrieve/pii/S1",
		`0;URL="/next page"`:        "/next page",
		"0; https://ex.org/b":       "https://ex.org/b",
		"0, /c":                     "/c",
		"5 url=/d":                  "/d",
		"  1.5 ; url = /e  ":        "/e",
		"3":                         "",
		"3x; url=/f":                "",
	} {
		if got := refreshURL(content); got != want {
			t.Errorf("refreshURL(%q) = %q, want %q", content, got, want)
		}
	}
	base, _ := url.Parse("https://ex.org/a")
	page := []byte(`<html><head><meta content="0; url=/next" http-equiv="Refresh"></head><body></body></html>`)
	if got := metaRefreshTarget("text/html", page, base); got != "https://ex.org/next" {
		t.Errorf("a content-first refresh read as %q", got)
	}
}
