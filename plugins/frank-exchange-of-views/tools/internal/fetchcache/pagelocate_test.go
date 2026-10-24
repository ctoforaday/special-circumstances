package fetchcache

import (
	"errors"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// A LINE THE ENGINE READ AND REMOVED IS REFUSED AS REMOVED (#1030). A quote of "ee ee Fe ee!" — a
// dashed rule the recognizer read as letters — is not a misquote, and the refusal must not read
// like one: it names the page and the line, and a quote of nothing at all still gets the absent
// sentence.
func TestAQuoteOfARemovedLineIsRefusedAsRemovedNotAsAbsent(t *testing.T) {
	withEngine(t, &fakeEngine{perCall: func(n int) (tessocr.PageResult, error) {
		if n == 2 {
			return tessocr.PageResult{Text: "Spot Welding 3.7 (2-6)", RemovedLines: []string{"mom me ee | ee ee Fe ee!"}}, nil
		}
		return tessocr.PageResult{Text: "SENSORS"}, nil
	}})
	run := runtest.New(t, t.TempDir())
	sha := "doc-under-test"
	rd := fakeRender(t, run, sha, [][]byte{[]byte("p1"), []byte("p2")}, tessocr.RenderDPI)
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatal(err)
	}
	doc := Entry{Sha: sha, ContentType: "application/pdf"}
	if loc, err := LocateSourceSpan(run, doc, "Spot Welding 3.7 (2-6)"); err != nil || len(loc.Pages) != 1 || loc.Pages[0] != 2 {
		t.Fatalf("a quote of the reading: %v, %v", loc, err)
	}
	_, err := LocateSourceSpan(run, doc, "ee  ee Fe")
	if !errors.Is(err, ErrSpanRemoved) {
		t.Fatalf("a quote of a removed line: %v, want ErrSpanRemoved", err)
	}
	for _, want := range []string{"page 2", `"mom me ee | ee ee Fe ee!"`, "removed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
	if _, err := LocateSourceSpan(run, doc, "Arc Welding"); !errors.Is(err, ErrSpanAbsent) {
		t.Errorf("a quote of nothing: %v, want ErrSpanAbsent", err)
	}
}

// EMBEDDED ONLY ON POSITIVE EVIDENCE. Each rule of TextOrigin, one case each; the OCR case, which
// needs a reading on disk, is driven end to end in internal/cli/ocrcite_test.go.
func TestTextOriginNeedsEvidenceBeforeItSaysEmbedded(t *testing.T) {
	yes, no := true, false
	cases := []struct {
		name  string
		entry Entry
		want  recordpb.SourceTextOrigin
	}{
		{"a text-layer PDF", Entry{ContentType: "application/pdf", TextExtracted: &yes}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED},
		{"html nobody extracted", Entry{ContentType: "text/html"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED},
		{"json", Entry{ContentType: "application/ld+json"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED},
		{"an image", Entry{ContentType: "image/png"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE},
		{"a legacy PDF line with no reading", Entry{ContentType: "application/pdf"}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE},
		{"an unread scan", Entry{ContentType: "application/pdf", TextExtracted: &no}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE},
		{"an app skeleton", Entry{ContentType: "text/html", NotRenderable: &yes}, recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := runtest.New(t, t.TempDir())
			tc.entry.URL = "https://ex/" + tc.name
			e, err := Store(run, tc.entry, []byte(tc.name))
			if err != nil {
				t.Fatal(err)
			}
			got, _, err := TextOrigin(run, e)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("origin = %v, want %v", got, tc.want)
			}
		})
	}
}

// THE ONE DECLARED NORMALISATION, each rule one case: whitespace runs are one space, a line-end
// hyphen is read joined or kept, and every glyph typoFold does not name must match. The offsets name the source's
// own bytes, so the matched text is the page's, not the span as typed.
func TestASpanIsFoundAfterTheDeclaredNormalisationAndPlacedInTheSourcesBytes(t *testing.T) {
	for _, tc := range []struct {
		text, span, matched string
	}{
		{"Annexes A\nthrough  J apply.", "Annexes A through J", "Annexes A\nthrough  J"},
		{"the photo-\nsynthesis rate", "photosynthesis rate", "photo-\nsynthesis rate"},
		{"a well-\n  known result", "well-known result", "well-\n  known result"},
		{"tool-\nsupported", "tool- supported", "tool-\nsupported"},
		{"naïve café", "naïve café", "naïve café"},
	} {
		got, ok := locateIn(tc.text, tc.span)
		if !ok || got != tc.matched {
			t.Errorf("text %q, span %q: matched %q (%v), want %q", tc.text, tc.span, got, ok, tc.matched)
		}
	}
	for _, tc := range []struct{ text, span string }{
		{"Annexes A through J", "annexes a through j"},
		{"it is \u00abprime\u00bb", "it is \"prime\""},
		{"x \u2212 1", "x - 1"},
		{"\u00e9tude", "etude"},
		{"a well - known result", "a wellknown result"},
		{"the end-\n\nNext", "endNext"},
	} {
		if got, ok := locateIn(tc.text, tc.span); ok {
			t.Errorf("text %q holds span %q as %q; the normalisation does not declare that", tc.text, tc.span, got)
		}
	}
}

func locateIn(text, span string) (string, bool) {
	for _, mode := range []hyphenMode{hyphenAsIs, hyphenJoined, hyphenKept} {
		f, want := fold(text, mode), fold(span, mode).s
		if i := strings.Index(f.s, want); i >= 0 && want != "" {
			return text[f.at[i]:f.end[i+len(want)-1]], true
		}
	}
	return "", false
}

// A SPAN IN AN HTML PAGE'S TEXT is found once, its offsets name the extraction's bytes, and the
// text either side is the page's, bounded. A span the text holds twice is refused with the count,
// and one it does not hold names the file blue copies from.
func TestASpanInAPagesTextIsFoundOnceWithTheTextAroundIt(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	body := strings.Repeat("Filler words before. ", 30) +
		"An odd composite n for which a^(n-1) is 1 mod n is a Fermat pseudo-\nprime to base a, and a is a Fermat liar. " +
		"Twice said. Twice said. " + strings.Repeat("Filler words after. ", 30)
	e, err := Store(run, Entry{URL: "https://ex/fermat", ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+body+"</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if e.TextSha, err = StoreText(run, e.Sha, []byte(body)); err != nil {
		t.Fatal(err)
	}
	loc, err := LocateSourceSpan(run, e, "An odd composite n for which a^(n-1) is 1 mod n is a Fermat pseudoprime to base a, and a is a Fermat liar")
	if err != nil {
		t.Fatal(err)
	}
	if body[loc.Start:loc.End] != loc.Matched || loc.Matched != "An odd composite n for which a^(n-1) is 1 mod n is a Fermat pseudo-\nprime to base a, and a is a Fermat liar" {
		t.Errorf("matched %q at [%d,%d)", loc.Matched, loc.Start, loc.End)
	}
	if loc.TextSha != e.TextSha || len(loc.Pages) != 0 {
		t.Errorf("text sha %q, pages %v: want the extraction's sha and no page", loc.TextSha, loc.Pages)
	}
	if len(loc.Before) > SpanWindow+3 || len(loc.After) > SpanWindow+3 || !strings.HasSuffix(loc.Before, "Filler words before. Filler words before. ") ||
		!strings.HasPrefix(loc.After, ". Twice said.") {
		t.Errorf("window: before %q, after %q", loc.Before, loc.After)
	}
	if _, err := LocateSourceSpan(run, e, "Twice said."); !errors.Is(err, ErrSpanRepeated) || !strings.Contains(err.Error(), "2 times") {
		t.Errorf("a span the text holds twice: %v, want ErrSpanRepeated naming the count", err)
	}
	_, err = LocateSourceSpan(run, e, "a Fermat liar is a composite")
	if !errors.Is(err, ErrSpanAbsent) || !strings.Contains(err.Error(), TextPath(run, e.Sha)) {
		t.Errorf("a span the text does not hold: %v, want ErrSpanAbsent naming %s", err, TextPath(run, e.Sha))
	}
	img, err := Store(run, Entry{URL: "https://ex/figure", ContentType: "image/png"}, []byte("png"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LocateSourceSpan(run, img, "anything"); !errors.Is(err, ErrNoText) {
		t.Errorf("a span in an image: %v, want ErrNoText", err)
	}
}

// A TYPOGRAPHIC CHARACTER IS ITS KEYBOARD FORM, both ways: a span typed in keyboard characters is
// found in a page that sets the typographic ones, and a span copied with the typographic ones is
// found in a page that has the keyboard ones. Either way the match is the page's own bytes. One
// row per character class, so dropping a class from typoFold fails its rows.
func TestATypographicCharacterIsFoundAsItsKeyboardFormBothWays(t *testing.T) {
	for _, tc := range []struct {
		class, typographic, keyboard string
	}{
		{"left single quote", "the \u2018sieve\u2019 test", "the 'sieve' test"},
		{"apostrophe", "Fermat\u2019s little theorem", "Fermat's little theorem"},
		{"low single quote", "the \u201asieve\u2019 test", "the 'sieve' test"},
		{"reversed single quote", "the \u201bsieve\u2019 test", "the 'sieve' test"},
		{"double quotes", "a \u201cliar\u201d base", "a \"liar\" base"},
		{"low double quote", "a \u201eliar\u201d base", "a \"liar\" base"},
		{"reversed double quote", "a \u201fliar\u201d base", "a \"liar\" base"},
		{"hyphen", "a well\u2010known result", "a well-known result"},
		{"non-breaking hyphen", "a well\u2011known result", "a well-known result"},
		{"figure dash", "call 555\u20121234 now", "call 555-1234 now"},
		{"en dash", "pages 10\u201320 hold it", "pages 10-20 hold it"},
		{"em dash", "the test\u2014a weak one\u2014fails", "the test-a weak one-fails"},
		{"horizontal bar", "the test\u2015a weak one", "the test-a weak one"},
		{"ellipsis", "and so on\u2026 until n", "and so on... until n"},
		{"soft hyphen", "pseudo\u00adprime to base a", "pseudoprime to base a"},
		{"ff ligature", "the e\ufb00ect holds", "the effect holds"},
		{"fi ligature", "the \ufb01rst base", "the first base"},
		{"fl ligature", "a \ufb02oor of n", "a floor of n"},
		{"ffi ligature", "su\ufb03cient for n", "sufficient for n"},
		{"ffl ligature", "the ba\ufb04ing case", "the baffling case"},
		{"no-break space", "10\u00a0kg of it", "10 kg of it"},
		{"thin space", "10\u2009kg of it", "10 kg of it"},
		{"narrow no-break space", "10\u202fkg of it", "10 kg of it"},
		{"figure space", "10\u2007kg of it", "10 kg of it"},
	} {
		page := "Before it. " + tc.typographic + ". After it."
		if got, ok := locateIn(page, tc.keyboard); !ok || got != tc.typographic {
			t.Errorf("%s: keyboard span %q in a typographic page matched %q (%v), want the page's %q", tc.class, tc.keyboard, got, ok, tc.typographic)
		}
		page = "Before it. " + tc.keyboard + ". After it."
		if got, ok := locateIn(page, tc.typographic); !ok || got != tc.keyboard {
			t.Errorf("%s: typographic span %q in a keyboard page matched %q (%v), want the page's %q", tc.class, tc.typographic, got, ok, tc.keyboard)
		}
	}
	// A soft hyphen at a line end breaks a word: it reads joined, never as a compound.
	if got, ok := locateIn("a pseudo\u00ad\nprime base", "pseudoprime base"); !ok || got != "pseudo\u00ad\nprime base" {
		t.Errorf("soft hyphen at a line end: matched %q (%v)", got, ok)
	}
	if got, ok := locateIn("a pseudo\u00ad\nprime base", "pseudo-prime base"); ok {
		t.Errorf("soft hyphen at a line end read as a compound: matched %q", got)
	}
	// An en dash at a line end is a hyphen there like any other.
	if got, ok := locateIn("pages 10\u2013\n20 hold it", "10-20 hold"); !ok || got != "10\u2013\n20 hold" {
		t.Errorf("en dash at a line end: matched %q (%v)", got, ok)
	}
	// A span that ends or starts inside a folded character is widened to the whole character.
	if got, ok := locateIn("the \ufb01rst base", "the f"); !ok || got != "the \ufb01" {
		t.Errorf("span ending inside a ligature: matched %q (%v)", got, ok)
	}
}

// THROUGH THE VERB'S LOCATOR, a keyboard span in a page set in typographic characters records the
// page's own bytes: the offsets index the extraction, and the stored matched and surrounding text
// are what the source holds, not the span as typed.
func TestATypographicPagesSpanIsStoredAsThePagesOwnBytes(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	yes := true
	body := "Before \u201cthe sieve\u201d. Fermat\u2019s test \u2014 the \ufb01rst one \u2014 holds for 10\u00a0bases\u2026 After\u2019s end."
	e, err := Store(run, Entry{URL: "https://ex/typo", ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+body+"</html>"))
	if err != nil {
		t.Fatal(err)
	}
	if e.TextSha, err = StoreText(run, e.Sha, []byte(body)); err != nil {
		t.Fatal(err)
	}
	loc, err := LocateSourceSpan(run, e, "Fermat's test - the first one - holds for 10 bases...")
	if err != nil {
		t.Fatal(err)
	}
	want := "Fermat\u2019s test \u2014 the \ufb01rst one \u2014 holds for 10\u00a0bases\u2026"
	if loc.Matched != want || body[loc.Start:loc.End] != want {
		t.Errorf("matched %q at [%d,%d), want the page's %q", loc.Matched, loc.Start, loc.End, want)
	}
	if loc.Before != "Before \u201cthe sieve\u201d. " || loc.After != " After\u2019s end." {
		t.Errorf("window: before %q, after %q — want the page's own bytes", loc.Before, loc.After)
	}
	if _, err := LocateSourceSpan(run, e, "Fermats test"); !errors.Is(err, ErrSpanAbsent) || !strings.Contains(err.Error(), SpanNormalisation) {
		t.Errorf("a span the text does not hold: %v, want ErrSpanAbsent stating the normalisation", err)
	}
}
