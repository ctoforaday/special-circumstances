package record

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A READING'S GAP THAT CANNOT BE PLACED IS REFUSED BY THE ONE-GAP RULE, NEVER PASSED. The rule asks
// which sentence the act's Anchor ends to say whether an open gap of the Mint's class already stands
// there; a sentence_of placement the report refuses is a refusal in the rule's own words, since
// passing it would report "no gap of this class here" about a sentence it never found. Driven with no
// guards — the act's own placement guard refuses the same placement first in live use — while an open
// gap of the class (G-1) stands, so the rule reaches the placement.
func TestAReadingGapTheOneGapRuleCannotPlaceIsRefused(t *testing.T) {
	run := foundAgainstBoard(t).seed()
	before := lastEventID(t, run)
	_, err := AppendReading(Identity{Run: run, SeatID: evLens}, liveRefutes(), readingGap("G-2"),
		&recordpb.Anchor{Id: proto.String("G-2"), SentenceOf: proto.String("C-9")})
	mustRefuse(t, err, "gap G-2 is placed at the end of the sentence", "refuses that placement")
	if after := lastEventID(t, run); after != before {
		t.Errorf("the refused reading wrote events %d..%d", before+1, after)
	}
}
