package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// WHERE A CITE'S TEXT CAME FROM IS STAMPED OR THE WRITE IS REFUSED, and the OCR pins belong to OCR
// text alone. validate is the single write path, so these hold for every caller, not only the verb.
func TestACiteOriginAndItsPinsAreCheckedAtTheWrite(t *testing.T) {
	run := mustRun(t, t.TempDir())
	ocr := recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_OCR
	embedded := recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED
	for _, tc := range []struct {
		name string
		body *recordpb.Cite
		want string
	}{
		{"unstamped", &recordpb.Cite{Label: proto.String("c-1")}, "carries no source_text_origin"},
		{"pages on embedded text", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum(), Pages: []int32{3}}, "whose text origin is embedded"},
		{"pages without the reading", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: ocr.Enum(), Pages: []int32{3}}, "pages without the span and the reading"},
		{"ocr with its pins", located(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: ocr.Enum(), Pages: []int32{3},
			OcrEngine: proto.String("e"), OcrTextSha: proto.String("t")}), ""},
		{"ocr on two pages", located(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: ocr.Enum(), Pages: []int32{3, 4},
			OcrEngine: proto.String("e"), OcrTextSha: proto.String("t")}), "carries 2 pages"},
		// A SOURCE WITH TEXT IS CITED AT A LOCATED SPAN, and one without at none.
		{"embedded with its span", located(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum()}), ""},
		{"embedded without a span", stamped(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum()}), "carries no located span"},
		{"a span without the source's text", stamped(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum(),
			SourceSpan: proto.String("s"), SpanStart: proto.Int32(0), SpanEnd: proto.Int32(1), SpanTextSha: proto.String("t")}), "carries no located span"},
		{"a span on a source with no text", located(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()}), "the tool holds no text"},
		{"no text, no span", stamped(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()}), ""},
		// THE SECOND THING THE TOOL STAMPS, refused on the same terms. A cite with a perfect
		// origin and no word on whether the work still stands renders as an ordinary source.
		{"no work status", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum()}, "carries no work_status"},
		// AND THE THIRD: a cite of an abstract with nothing saying so renders like a cite of the paper.
		{"no completeness", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: embedded.Enum(),
			WorkStatus: recordpb.WorkStatus_WORK_STATUS_STANDING.Enum()}, "carries no source_completeness"},
		// AND THE FOURTH: whose bytes the citation read. Unstamped, a recovered copy renders as the
		// typed url's own; a copy url on the url's own bytes names a place nothing read; a followed
		// pointer or a fuller copy is always fetched from a url, so either without one lost it. A
		// recovery whose body is the tool's own statement was read from no url, and says so.
		{"no copy substitution", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()}, 0, ""), "carries no copy_substitution"},
		{"a copy url on the url's own bytes", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()},
			recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE, "https://elsewhere"), "a copy url names where substituted bytes came from"},
		{"a followed pointer with no url", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()},
			recordpb.CopySubstitution_COPY_SUBSTITUTION_FOLLOWED_LINK, ""), "followed_link and no copy_url"},
		{"a fuller copy with no url", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()},
			recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY, ""), "fuller_copy and no copy_url"},
		{"a recovered copy and its url", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()},
			recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED, "https://web.archive.org/x"), ""},
		{"a recovery read from no url", copied(&recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE.Enum()},
			recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED, ""), ""},
	} {
		err := validateRun(run, "blue-respond", recordpb.EventType_EVENT_TYPE_CITE, tc.body)
		switch {
		case tc.want == "" && err != nil:
			t.Errorf("%s: refused: %v", tc.name, err)
		case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
			t.Errorf("%s: err = %v, want a refusal saying %q", tc.name, err, tc.want)
		}
	}
}

func stamped(c *recordpb.Cite) *recordpb.Cite {
	c.WorkStatus = recordpb.WorkStatus_WORK_STATUS_STANDING.Enum()
	c.SourceCompleteness = recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL.Enum()
	c.CopySubstitution = recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()
	return c
}

// copied is stamped with the copy fields set as given: a zero kind leaves it unstamped.
func copied(c *recordpb.Cite, kind recordpb.CopySubstitution, copyURL string) *recordpb.Cite {
	stamped(c)
	c.CopySubstitution = nil
	if kind != 0 {
		c.CopySubstitution = kind.Enum()
	}
	if copyURL != "" {
		c.CopyUrl = proto.String(copyURL)
	}
	return c
}

func located(c *recordpb.Cite) *recordpb.Cite {
	c.SourceSpan, c.SpanStart, c.SpanEnd, c.SpanTextSha = proto.String("s"), proto.Int32(4), proto.Int32(5), proto.String("t")
	c.SourceBefore, c.SourceMatched, c.SourceAfter = proto.String("the "), proto.String("s"), proto.String(" said")
	return stamped(c)
}

func TestAVerifyPageAndItsHashesAreOneFact(t *testing.T) {
	run := mustRun(t, t.TempDir())
	base := func() *recordpb.Verify {
		return &recordpb.Verify{Claim: proto.String("c"), Anchor: proto.String("c-1"), Text: proto.String("r"),
			Outcome:    recordpb.SourceOutcome_SOURCE_OUTCOME_SUPPORTS.Enum(),
			Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum()}
	}
	half := base()
	half.Page = proto.Int32(2)
	half.PageRenderSha = proto.String("a")
	if err := validateRun(run, "red-lens-evidence", recordpb.EventType_EVENT_TYPE_VERIFY, half); err == nil || !strings.Contains(err.Error(), "written together") {
		t.Errorf("a page with one hash was not refused: %v", err)
	}
	whole := base()
	whole.Page, whole.PageRenderSha, whole.ReadingRenderSha = proto.Int32(2), proto.String("a"), proto.String("b")
	if err := validateRun(run, "red-lens-evidence", recordpb.EventType_EVENT_TYPE_VERIFY, whole); err != nil && strings.Contains(err.Error(), "written together") {
		t.Errorf("a page with both hashes was refused as partial: %v", err)
	}
}
