package view

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE CHECK OF AN INQUIRY LIVES IN THE INQUIRIES' OWN DOCUMENT, beside the fate it judged: a reader
// following one inquiry meets its path, the fate's reason, and the logic lens's check with its reason.
func TestAnInquiryCheckRendersBesideTheFateItJudged(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_CONFIRMED), Reason: proto.String("the adjacent literature is what the question turns on")}),
		recordtest.Event(t, record.InquiryOwner(), &recordpb.InquiryCheck{InquiryId: proto.String("Q1"), Reason: proto.String("the fate says what the survey found")}),
	}
	got := string(inquiryMD(inputT(t, evs...)))
	for _, want := range []string{
		"checked by " + record.InquiryOwner(),
		"the fate says what the survey found",
		"the adjacent literature is what the question turns on",
		"proposed -> e0 confirmed",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the inquiry's document does not carry %q:\n%s", want, got)
		}
	}
}

// THE SEAT'S VIEW PRINTS EACH VALUE AS THE RECORD HOLDS IT. A seat quotes what it reads back into a
// write that matches the record, so a value of several lines prints with its lines, as written.
func TestTheInquiriesViewPrintsTheRecordsText(t *testing.T) {
	q, why, hyp, method := "census the plan.\n\nThen argue.", "none exists.\n\nThe plan names no function.", "one exists\nand it is explicit", "read it\n```\ngrep x\n```"
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String(q), Hypothesis: proto.String(hyp)}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Reason: proto.String(why), Method: proto.String(method)}),
		recordtest.Event(t, record.InquiryOwner(), &recordpb.InquiryCheck{InquiryId: proto.String("Q1"), Reason: proto.String("in scope.\n\nWholly.")}),
	}
	got := string(inquiryMD(inputT(t, evs...)))
	for _, want := range []string{q, why, hyp, method, "in scope.\n\nWholly."} {
		if !strings.Contains(got, want) {
			t.Errorf("the view does not print %q as written:\n%s", want, got)
		}
	}
}
