package scorecard

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func inquiry(t *testing.T, id, line string, st recordpb.InquiryStatus, reason string) *record.Event {
	t.Helper()
	return recordtest.Event(t, "blue-synthesize", &recordpb.Inquiry{
		InquiryId: proto.String(id), Question: proto.String(line),
		Status: &st, Reason: proto.String(reason),
	})
}

// THE RECORD HELD THE LINES AND THE SCORECARD SAID THERE WERE NONE.
//
// This row counted `avenues` arrays out of the seat ENVELOPES. When those did not arrive in the
// shape it expected it saw nothing and rendered "no avenues recorded — think-around-problem is
// back to self-attested", which reads as a measured finding ABOUT THE RUN rather than as a failed
// read. Measured in research/2026-09-02_quadratic-formula: the record held 35 distinct lines with
// 113 pursued-moves while this row reported none, and a seat noticed only because
// `show avenues` rendered 23 in the same sitting.
//
// The scorecard is HARVESTED INTO feov-memory, so the zero did not stay in the run — it became a
// cross-run memory row asserting no alternatives were explored in a run that explored 35. A wrong
// number is worse than a missing one precisely here, because the next run inherits it.
func TestInquiriesComeFromTheRecordNotTheEnvelopes(t *testing.T) {
	board := famSeededT(t, []*record.Event{
		inquiry(t, "Q1", "the ring-theoretic generalisation", recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY, ""),
		inquiry(t, "Q2", "non-English prior art", recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED,
			"the originality claim is scoped to English sources, so this is out of scope for the question asked"),
	})
	// The envelopes carry NOTHING — exactly the state that produced the false zero.
	rows := blueRows(record.Run{}, []map[string]any{{}}, nil, board, record.WhileRunning)

	r := rowByMetric(rows, "inquiries")
	if r == nil {
		t.Fatal("no inquiries row at all")
	}
	if r.Value == nil {
		t.Fatalf("the record holds two lines and the scorecard reported none: %q", r.Note)
	}
	got := string(r.Value.(objJSON))
	for _, want := range []string{"underway", "declined"} {
		if !strings.Contains(got, want) {
			t.Errorf("the status breakdown is missing %q: %s", want, got)
		}
	}
}

// A LINE THAT MOVED IS COUNTED ONCE, UNDER THE STATUS IT NOW HOLDS. Counting events would report a
// line declined in one round and pursued in another as two lines, inflating the very breadth
// number this row exists to keep honest.
func TestALineThatMovedIsCountedOnceUnderItsCurrentStatus(t *testing.T) {
	board := famSeededT(t, []*record.Event{
		inquiry(t, "Q1", "the Medium essay", recordpb.InquiryStatus_INQUIRY_STATUS_DECLINED, "no access channel exists for it"),
		func() *record.Event {
			st := recordpb.InquiryStatus_INQUIRY_STATUS_UNDERWAY
			return recordtest.Event(t, "blue-respond", &recordpb.Inquiry{
				InquiryId: proto.String("Q1"), Status: &st, SupersedesStatus: proto.String("declined"),
			})
		}(),
	})
	rows := blueRows(record.Run{}, nil, nil, board, record.WhileRunning)
	got := string(rowByMetric(rows, "inquiries").Value.(objJSON))
	if !strings.Contains(got, `"underway":1`) {
		t.Errorf("a line that moved declined->underway is not counted under its CURRENT status: %s", got)
	}
	if strings.Contains(got, `"declined":1`) {
		t.Errorf("the superseded status was counted as a second line: %s", got)
	}
}

// NOT MEASURED IS NOT ZERO. Without a board this row has not been computed, and saying "no
// avenues recorded" would be the same defect one layer up — a failed read wearing the words of
// a finding.
func TestAnUnreadableRecordSaysNotMeasuredRatherThanNone(t *testing.T) {
	r := rowByMetric(blueRows(record.Run{}, nil, nil, nil, record.WhileRunning), "inquiries")
	if r == nil || r.Value != nil {
		t.Fatalf("expected an uncomputed row, got %+v", r)
	}
	if !strings.Contains(r.Note, "NOT MEASURED") {
		t.Errorf("an unread record renders as a statement about the run: %q", r.Note)
	}
}
