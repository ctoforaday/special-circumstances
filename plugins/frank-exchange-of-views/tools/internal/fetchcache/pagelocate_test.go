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
	rec, err := ReadRenderedPages(run, sha, rd)
	if err != nil {
		t.Fatal(err)
	}
	if loc, err := LocateSpan(run, rec, "Spot Welding 3.7"); err != nil || len(loc.Pages) != 1 || loc.Pages[0] != 2 {
		t.Fatalf("a quote of the reading: %v, %v", loc, err)
	}
	_, err = LocateSpan(run, rec, "ee  ee Fe")
	if !errors.Is(err, ErrSpanRemoved) {
		t.Fatalf("a quote of a removed line: %v, want ErrSpanRemoved", err)
	}
	for _, want := range []string{"page 2", `"mom me ee | ee ee Fe ee!"`, "removed"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q: %v", want, err)
		}
	}
	if _, err := LocateSpan(run, rec, "Arc Welding"); !errors.Is(err, ErrSpanAbsent) {
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

func TestLocateSpanComparesWhitespaceAsOneSpaceAndGlyphsExactly(t *testing.T) {
	for _, tc := range []struct {
		page, span string
		want       bool
	}{
		{"Annexes A\nthrough  J", "Annexes A through J", true},
		{"Annexes A through J", "annexes a through j", false},
		{"tool-\nsupported", "toolsupported", false},
		{"tool-\nsupported", "tool- supported", true},
	} {
		got := strings.Contains(normalizeSpace(tc.page), normalizeSpace(tc.span))
		if got != tc.want {
			t.Errorf("page %q holds span %q = %v, want %v", tc.page, tc.span, got, tc.want)
		}
	}
}
