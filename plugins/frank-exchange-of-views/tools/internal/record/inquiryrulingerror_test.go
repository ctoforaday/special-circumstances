package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"google.golang.org/protobuf/proto"
)

// A RULING THAT CANNOT BE READ IS AN ERROR, NOT "NEVER RULED". AvenueRuling's "" is the answer a
// caller skips its check on — the release gate's avenue oracle does exactly that — so a failed
// read answering "" turns the check off for that avenue and reports nothing.
//
// The fixture holds one avenue red ruled and one it did not, so the two answers a failure could
// be mistaken for are both on the record. The failure is made by dropping the view through the
// run's cached handle: every later read goes through that handle and fails as a busy or
// malformed database would.
func TestInquiryRulingReturnsTheReadItCannotMake(t *testing.T) {
	run := corrRun(t)
	blue := Identity{Run: run, SeatID: "blue-respond"}
	red := Identity{Run: run, SeatID: "red-chair"}
	for _, id := range []string{"Q1", "Q2"} {
		mustAppend(t, blue, &recordpb.Inquiry{InquiryId: proto.String(id), Question: proto.String("a line " + id), Status: recordpb.InquiryStatus_INQUIRY_STATUS_PROPOSED.Enum()})
	}
	mustAppend(t, red, &recordpb.MotionRule{
		MotionId: proto.String("Q1"),
		Subject:  recordtest.P(recordpb.MotionSubject_MOTION_SUBJECT_INQUIRY),
		Opinion:  proto.String("a real question, but not this run's"),
		Ruling:   &recordpb.MotionRule_Inquiry{Inquiry: recordpb.InquiryRuling_INQUIRY_RULING_OUT_OF_SCOPE},
	})

	// The healthy answers first: the schema's word verbatim, "" for the line red has not ruled,
	// and an error for an id that names no avenue.
	if got, err := InquiryRuling(run, "Q1"); err != nil || got != "out_of_scope" {
		t.Fatalf("InquiryRuling on a ruled inquiry = (%q, %v), want the schema's word out_of_scope", got, err)
	}
	if got, err := InquiryRuling(run, "Q2"); err != nil || got != "" {
		t.Fatalf("InquiryRuling on an inquiry red has not ruled = (%q, %v), want \"\" and no error", got, err)
	}
	if got, err := InquiryRuling(run, "Q9"); err == nil || got != "" || !strings.Contains(err.Error(), "Q9") {
		t.Errorf("InquiryRuling on an id no proposal created = (%q, %v), want an error naming Q9 — \"\" here reads as an inquiry red never ruled", got, err)
	}

	db, err := openRunForRead(run)
	if err != nil || db == nil {
		t.Fatalf("open the run's handle: (%v, %v)", db, err)
	}
	if _, err := db.Exec(`DROP VIEW "inquiry_state"`); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"Q1", "Q2"} {
		if got, err := InquiryRuling(run, id); err == nil || got != "" || !strings.Contains(err.Error(), "inquiry_state") {
			t.Errorf("InquiryRuling(%s) over a failing read = (%q, %v), want the read error naming inquiry_state — \"\" with no error is the never-ruled answer", id, got, err)
		}
	}
}
