package scorecard

import (
	"encoding/json"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func correctionBy(t *testing.T, seat, key string) *record.Event {
	return recordtest.At(t, seat, seat+":correction:"+key, &recordpb.Correction{
		Corrects: proto.String(key), Replacement: proto.String(key + "~1"), Why: proto.String("a word was lost")})
}

// EVERY CARD COUNTS ITS SEATS' CORRECTIONS — the only brake the owner left on an uncapped chain —
// per seat, and an unreadable record is "not measured", never a zero.
func TestEveryCardCountsItsSeatsCorrections(t *testing.T) {
	fam := famOfEventsT([]*record.Event{
		correctionBy(t, "blue-respond", "blue-respond:log:#1"),
		correctionBy(t, "blue-respond", "blue-respond:position:#1"),
		correctionBy(t, "red-chair", "red-chair:position:#1"),
		correctionBy(t, "red-lens-evidence", "red-lens-evidence:regrade:#1:G1"),
		correctionBy(t, "judge", "judge:outcome:#1"),
	})
	cards := Compute(record.Run{}, nil, fam)
	for card, want := range map[string]map[string]int{
		"blue":  {"blue-respond": 2},
		"red":   {"red-chair": 1, "red-lens-evidence": 1},
		"bench": {"judge": 1},
	} {
		r := rowByMetric(cards[card], "corrections")
		if r == nil {
			t.Errorf("the %s card carries no corrections row", card)
			continue
		}
		var got map[string]int
		v, _ := r.Value.(objJSON)
		if err := json.Unmarshal([]byte(v), &got); err != nil || len(got) != len(want) {
			t.Errorf("the %s card's corrections = %v, want %v", card, r.Value, want)
			continue
		}
		for seat, n := range want {
			if got[seat] != n {
				t.Errorf("the %s card counts %d correction(s) for %s, want %d", card, got[seat], seat, n)
			}
		}
	}

	for card, rows := range Compute(record.Run{}, nil, nil) {
		r := rowByMetric(rows, "corrections")
		if r == nil || r.Value != nil || !strings.Contains(r.Note, "not measured") {
			t.Errorf("an unreadable record must leave the %s card's corrections not measured, got %+v", card, r)
		}
	}

	none := famOfEventsT([]*record.Event{recordtest.Event(t, "blue-respond", &recordpb.Register{})})
	if r := rowByMetric(Compute(record.Run{}, nil, none)["blue"], "corrections"); r == nil || r.Value != objJSON("{}") {
		t.Errorf("a readable record with no correction is measured and none: %+v", r)
	}
}
