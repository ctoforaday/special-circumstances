package record

import (
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// AN AMEND OF A GAP IS A CHANGE TO IT, to a seat whose sitting opened before it (staleOf): blue sits,
// the lens amends G1, and the amend is among the changes to G1 blue has not read. A sitting opened
// after the amend has read it.
func TestAnAmendIsAChangeToItsGap(t *testing.T) {
	amend := &recordpb.Amend{GapId: proto.String("G1"), AcceptanceCheck: proto.String("the check names an outcome"), Basis: proto.String("restated")}
	amended := func(ch []StaleChange) bool {
		return slices.ContainsFunc(ch, func(c StaleChange) bool { return strings.Contains(c.What, "amended G1") })
	}
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		bracket(evLens, "lens-a").mint(evLens, "G1", "high").stop("lens-a").
		bracket("blue-respond", "blue-a").add(evLens, amend)
	evs, win := mustEvents(t, b.seed())
	if ch := staleOf(evs, win, "blue-respond", "G1"); !amended(ch) {
		t.Errorf("changes to G1 since blue sat = %+v, want the lens's amend", ch)
	}
	b2 := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		bracket(evLens, "lens-a").mint(evLens, "G1", "high").add(evLens, amend).stop("lens-a").
		bracket("blue-respond", "blue-a")
	evs, win = mustEvents(t, b2.seed())
	if ch := staleOf(evs, win, "blue-respond", "G1"); amended(ch) {
		t.Errorf("changes to G1 since blue sat = %+v, want none: the amend preceded its sitting", ch)
	}
}
