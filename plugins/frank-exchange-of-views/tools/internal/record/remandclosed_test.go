package record

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A REMAND OF A CLOSED GAP IS REFUSED, AND THE REFUSAL NAMES THE CLOSE (gblock, 2026-10-24). A
// docketing can stand on a closed gap — written after the close, as a proposed successor on it is
// docketed (F8); a close written after a docketing ends it — and the bench rules it on the record as it
// stands, but a remand of it owes an exchange no seat can have, and the record does not reopen it. On
// universe-m33 the bench remanded a gap red had already closed; the remand was stored, never acted on,
// and the run passed. A disposition that closes is admitted on the same docketing, and under a
// migration the archived remand is what the bench did.
func TestARemandOfAClosedGapIsRefusedNamingTheClose(t *testing.T) {
	st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
		register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", "high")
	st.register(evLens).closeGap(evLens, "G1")
	closedAt := int64(st.n)
	st.register("red-chair").docketMotion("red-chair", "D1", "G1")
	// The chair's bench row on G1 convenes the bench's docket sitting, which holds D1 (benchHeld).
	run := st.dispatch(2, "judge", "G1").register("judge").seed()

	err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("D1", remandDirection))
	if err == nil {
		t.Fatal("a remand of a closed gap was admitted")
	}
	for _, want := range []string{"G1", "CLOSED", evLens, fmt.Sprintf("event %d", closedAt), "Nothing was recorded"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q: %v", want, err)
		}
	}
	if got := feov.CodeOf(err); got != string(feov.Conflict) {
		t.Errorf("the refusal is coded %q, want %q", got, feov.Conflict)
	}
	if err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, docketRuling("D1", recordpb.Disposition_DISPOSITION_NOT_A_DEFECT)); err != nil {
		t.Errorf("a disposition that closes is refused on the closed gap's docketing: %v", err)
	}
	if err := wouldReplay(t, run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("D1", remandDirection)); err == nil {
		t.Error("the replay admitted the remand of a closed gap; no archive holds one, and the write refuses it whoever replays it")
	}
}

// A REMAND OF AN OPEN GAP IS ADMITTED, material or not: the refusal is the closed gap's alone.
func TestARemandOfAnOpenGapIsAdmitted(t *testing.T) {
	for _, grade := range []string{"high", "low"} {
		st := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).sitClosed(evLens).mint(evLens, "G1", grade)
		run := st.register("red-chair").docketMotion("red-chair", "D1", "G1").dispatch(2, "judge", "G1").register("judge").seed()
		if err := validateRun(run, "judge", recordpb.EventType_EVENT_TYPE_MOTION_RULE, remandRuling("D1", remandDirection)); err != nil {
			t.Errorf("a remand of an open gap graded %s is refused: %v", grade, err)
		}
	}
}
