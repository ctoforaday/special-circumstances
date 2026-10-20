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
	if loc, err := LocateSourceSpan(run, doc, "Spot Welding 3.7"); err != nil || len(loc.Pages) != 1 || loc.Pages[0] != 2 {
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
// hyphen is read joined or kept, and every glyph else must match. The offsets name the source's
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
		{"it is \u201cprime\u201d", "it is \"prime\""},
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
			return text[f.at[i] : f.at[i+len(want)-1]+1], true
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
	loc, err := LocateSourceSpan(run, e, "a Fermat pseudoprime to base a, and a is a Fermat liar")
	if err != nil {
		t.Fatal(err)
	}
	if body[loc.Start:loc.End] != loc.Matched || loc.Matched != "a Fermat pseudo-\nprime to base a, and a is a Fermat liar" {
		t.Errorf("matched %q at [%d,%d)", loc.Matched, loc.Start, loc.End)
	}
	if loc.TextSha != e.TextSha || len(loc.Pages) != 0 {
		t.Errorf("text sha %q, pages %v: want the extraction's sha and no page", loc.TextSha, loc.Pages)
	}
	if len(loc.Before) > SpanWindow+3 || len(loc.After) > SpanWindow+3 || !strings.HasSuffix(loc.Before, "Filler words before. An odd composite n for which a^(n-1) is 1 mod n is ") ||
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
