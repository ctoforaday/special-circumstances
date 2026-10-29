package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/view"
	"google.golang.org/protobuf/proto"
)

// inquiriesDoc assembles a run of the events and returns its inquiries.md, with the inquiries view
// a seat reads beside it.
func inquiriesDoc(t *testing.T, evs ...*record.Event) (doc, seatView string) {
	t.Helper()
	dir := recordtest.TmpRun(t)
	recordtest.Seed(t, dir, evs...)
	run := runtest.Open(t, dir)
	// The document as composed, under no masthead: it is compared with the view whole.
	docs, _, _, err := stdReaders().assemble(run, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if d.File == FileInquiries {
			doc = d.Body
		}
	}
	fam, err := record.FamilyOf(run)
	if err != nil {
		t.Fatal(err)
	}
	return doc, view.InquiryBody(fam.Events, fam.At, false, view.AsWritten)
}

// EACH REVIEW THE CHAIR RECORDED SHIPS WITH ITS REASON, in the inquiries' own document, after the
// inquiries it read: the epoch it was made in, the seat, and the reason whole — its paragraphs and
// its list intact, quoted as the chair wrote it. It is the chair's account of a read, so it carries
// no verdict word; and the view a seat reads does not carry it.
func TestEachInquiryReviewShipsWithItsReason(t *testing.T) {
	const (
		first  = "Read at head 62, once, whole.\n\nBody coverage, by inquiry:\n- Q1: answered in \"Trial division\"."
		second = "One read of the report at head 80: Q1 still answered in \"Trial division\"."
	)
	doc, seatView := inquiriesDoc(t,
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("trial division to the square root")}),
		recordtest.Event(t, "red-chair", &recordpb.InquiryReview{Reason: proto.String(first)}),
		recordtest.Event(t, "red-chair", &recordpb.Register{}),
		recordtest.Event(t, "red-chair", &recordpb.InquiryReview{Reason: proto.String(second)}),
	)
	want := "## Inquiry reviews (2)\n\n" +
		"_Each is one read of the report against the inquiries on the record, and the reason is the chair's account of that read, in its words: where the body delivers the research each inquiry claims._\n\n" +
		"### Epoch 1 — `red-chair`\n\n" +
		"> Read at head 62, once, whole.\n>\n> Body coverage, by inquiry:\n> - Q1: answered in \"Trial division\".\n\n" +
		"### Epoch 2 — `red-chair`\n\n" +
		"> One read of the report at head 80: Q1 still answered in \"Trial division\".\n"
	if !strings.HasSuffix(doc, want) {
		t.Errorf("inquiries.md does not end with the chair's reviews:\n%s\nwant it to end with\n%s", doc, want)
	}
	if i := strings.Index(doc, "trial division to the square root"); i < 0 || i > strings.Index(doc, "## Inquiry reviews") {
		t.Errorf("the inquiries do not stand before the reviews of them:\n%s", doc)
	}
	if strings.Contains(seatView, "Read at head 62") || strings.Contains(seatView, "Inquiry reviews") {
		t.Errorf("the view a seat reads carries the chair's reviews:\n%s", seatView)
	}
}

// A RUN WITH NO REVIEW SHIPS NO REVIEW HEADING, and the inquiries are the document whole.
func TestARunWithNoReviewShipsTheInquiriesAlone(t *testing.T) {
	doc, seatView := inquiriesDoc(t,
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("trial division to the square root")}),
	)
	if strings.Contains(doc, "Inquiry reviews") {
		t.Errorf("a run with no review ships a heading for them:\n%s", doc)
	}
	if strings.TrimSpace(doc) != strings.TrimSpace(stdReaders().escapeNotes(seatView)) {
		t.Errorf("with no review, inquiries.md is not the inquiries view whole:\n%s\nview:\n%s", doc, seatView)
	}
}

// THE DOCUMENT KEEPS A VALUE'S PARAGRAPHS INSIDE ITS ITEM, AND THE VIEW PRINTS IT AS WRITTEN. One
// rendering serves both: inquiries.md indents each later line of a value to its item's content
// column (inItem), so a blank line or a fence stays in the row, and `show inquiries` prints the
// record's text, which a seat quotes back.
func TestAnInquiryValueKeepsItsParagraphsInTheDocument(t *testing.T) {
	doc, seatView := inquiriesDoc(t,
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("census the plan.\n\nThen argue.")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Reason: proto.String("none.\n\nThe plan names no function."), Method: proto.String("read it\n```\ngrep x")}))
	for _, want := range []string{"census the plan.\n\n  Then argue.", "none.\n\n  The plan names no function.", "  - method: read it\n    ```\n    grep x\n    ```"} {
		if !strings.Contains(doc, want) {
			t.Errorf("inquiries.md does not keep %q inside its item:\n%s", want, doc)
		}
	}
	for _, want := range []string{"census the plan.\n\nThen argue.", "none.\n\nThe plan names no function.", "  - method: read it\n```\ngrep x"} {
		if !strings.Contains(seatView, want) {
			t.Errorf("the seat's view does not print %q as written:\n%s", want, seatView)
		}
	}
}
