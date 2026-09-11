package verify

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// A CORRECTION THAT DOES NOT RESOLVE IS A VIOLATION, and one that does is held. Both directions:
// a check that only ever passes would read the same over a record whose pairs are broken.
func TestCorrectionsResolve(t *testing.T) {
	pos := func(s string) *recordpb.Position { return &recordpb.Position{Text: proto.String(s)} }
	corr := func(seat, k, r string) *record.Event {
		return recordtest.At(t, seat, seat+":correction:"+k, &recordpb.Correction{
			Corrects: proto.String(k), Replacement: proto.String(r), Why: proto.String("a word was lost")})
	}
	k := "blue-respond:position:#1"
	held := &boardT{Events: []*record.Event{
		recordtest.Event(t, "blue-respond", &recordpb.Register{}),
		recordtest.At(t, "blue-respond", k, pos("the report is  now")),
		recordtest.At(t, "blue-respond", k+"~1", pos("the report is sound now")),
		corr("blue-respond", k, k+"~1"),
	}}
	if got := find(t, Run(held.fam()), "corrections-resolve"); !got.OK || got.NA {
		t.Errorf("a correction whose act and replacement are both on the record must hold: %+v", got)
	}

	broken := &boardT{Events: []*record.Event{
		recordtest.Event(t, "blue-respond", &recordpb.Register{}),
		recordtest.At(t, "blue-respond", k, pos("the report is  now")),
		corr("blue-respond", k, k+"~1"), // the replacement never landed
	}}
	got := find(t, Run(broken.fam()), "corrections-resolve")
	if got.OK || len(got.Violations) != 1 || !strings.Contains(got.Violations[0], "names replacement "+k+"~1") {
		t.Errorf("a correction naming a replacement the record does not carry must be a violation, named: %+v", got)
	}
}
