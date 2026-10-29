package view

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// THE DEBATE OVER AN INQUIRY LIVES IN THE INQUIRIES' OWN DOCUMENT. Red's ruling is an argument
// beside the inquiry, and the fate blue recorded against it is the inquiry's own path: a reader
// following one inquiry meets the ruling, its opinion, and the fate blue recorded with that ruling
// in front of it.
func TestAnInquiryRulingRendersBesideTheFateRecordedAgainstIt(t *testing.T) {
	evs := []*record.Event{
		recordtest.Event(t, "blue-r0", &recordpb.Inquiry{InquiryId: proto.String("Q1"), Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED), Question: proto.String("survey the adjacent literature")}),
		recordtest.Event(t, "red-chair", &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Opinion:  proto.String("a real question, not this one's"),
			Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE},
		}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY), Reason: proto.String("the adjacent literature is what the question turns on")}),
	}
	got := string(inquiryMD(inputT(t, evs...)))
	for _, want := range []string{
		"RED RULED **out_of_scope**",
		"a real question, not this one's",
		"the adjacent literature is what the question turns on",
		"proposed -> e0 underway",
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
		recordtest.Event(t, "red-chair", &recordpb.MotionRule{
			MotionId: proto.String("Q1"),
			Opinion:  proto.String("in scope.\n\nWholly."),
			Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE},
		}),
		recordtest.Event(t, "blue-r1", &recordpb.Inquiry{InquiryId: proto.String("Q1"), SupersedesStatus: proto.String("1"),
			Status: recordtest.P(recordpb.InquiryStatus_INQUIRY_STATUS_REFUTED), Reason: proto.String(why), Method: proto.String(method)}),
	}
	got := string(inquiryMD(inputT(t, evs...)))
	for _, want := range []string{q, why, hyp, method, "in scope.\n\nWholly."} {
		if !strings.Contains(got, want) {
			t.Errorf("the view does not print %q as written:\n%s", want, got)
		}
	}
}
