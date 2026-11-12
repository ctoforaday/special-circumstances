package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN EVENT A WRITE RETURNED IS READ BACK BY ITS KEY (EventID): the reading's result names its event.
func TestEventIDReadsTheWrittenEventBack(t *testing.T) {
	run := reopenedBoard(t).seed()
	ev, err := Append(Identity{Run: run, SeatID: evLens}, &recordpb.Verify{Anchor: proto.String("C-1"), Claim: proto.String("the claim C-1 backs"),
		Outcome: recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES.Enum(), Confidence: recordpb.Confidence_CONFIDENCE_HIGH.Enum(), Text: proto.String("t"),
		CopySubstitution: recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE.Enum()})
	if err != nil {
		t.Fatal(err)
	}
	id, err := EventID(run, ev)
	if err != nil {
		t.Fatal(err)
	}
	if want := lastEventID(t, run); id != want {
		t.Errorf("EventID = %d, the record's last event is %d", id, want)
	}
	if _, err := EventID(run, &Event{}); err == nil {
		t.Error("an event with no key read back an id")
	}
}
