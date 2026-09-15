package fetchcache

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

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
