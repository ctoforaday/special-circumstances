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
		{"pages without the reading", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: ocr.Enum(), Pages: []int32{3}}, "pages without the quote and the reading"},
		{"ocr with its pins", &recordpb.Cite{Label: proto.String("c-1"), SourceTextOrigin: ocr.Enum(), Pages: []int32{3},
			OcrQuote: proto.String("s"), OcrEngine: proto.String("e"), OcrTextSha: proto.String("t")}, ""},
	} {
		err := validate(run, "blue-respond", recordpb.EventType_EVENT_TYPE_CITE, tc.body)
		switch {
		case tc.want == "" && err != nil:
			t.Errorf("%s: refused: %v", tc.name, err)
		case tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)):
			t.Errorf("%s: err = %v, want a refusal saying %q", tc.name, err, tc.want)
		}
	}
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
	if err := validate(run, "red-lens-evidence", recordpb.EventType_EVENT_TYPE_VERIFY, half); err == nil || !strings.Contains(err.Error(), "written together") {
		t.Errorf("a page with one hash was not refused: %v", err)
	}
	whole := base()
	whole.Page, whole.PageRenderSha, whole.ReadingRenderSha = proto.Int32(2), proto.String("a"), proto.String("b")
	if err := validate(run, "red-lens-evidence", recordpb.EventType_EVENT_TYPE_VERIFY, whole); err != nil && strings.Contains(err.Error(), "written together") {
		t.Errorf("a page with both hashes was refused as partial: %v", err)
	}
}
