package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"google.golang.org/protobuf/proto"
)

// A SPOT-CHECK BEARS AN AREA'S CHANGES ONCE, WHICHEVER SITTING FILED IT. The chair spot-checks a stale
// area in one sitting; at its next, with nothing changed since, the area is not stale — no blocker, no
// plan line, no repeat owed — and the PASS stands. A change after the spot-check makes it stale again.
// The smoke's chair spot-checked at one epoch and was refused the PASS at the next with nothing
// changed: a spot-check counted only in the sitting that filed it, and a repeat stood on no delivery,
// since the chair had borne every change.
func TestASpotCheckBearsAnAreasChangesUntilTheReportMoves(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest()
	b.register("red-chair")
	b.sit(2, evLens).sit(2, evLens)
	b.ingest()
	b.dispatch(int64(b.n), evLens).register(evLens)
	b.ingest()
	openChairSitting(b)
	b.served("red-chair", b.head())
	if !holdsStale(t, b) {
		t.Fatal("fixture: the area is not stale before any spot-check, so this measures nothing")
	}
	b.add("red-chair", &recordpb.SpotCheck{Areas: []string{evLens}, Pins: []int64{2}, Changes: []int64{b.head()},
		Reason: proto.String("the changes hold against the area's duties")})

	// The chair's next sitting, nothing changed since its spot-check.
	openChairSitting(b)
	if holdsStale(t, b) {
		t.Error("an area the chair spot-checked, with nothing changed since, is stale again at its next sitting — a repeat spot-check is owed that no delivery can stand on")
	}
	if err := validateRun(fresh(b), "red-chair", recordpb.EventType_EVENT_TYPE_VERDICT, passGate); err != nil {
		t.Errorf("the PASS is refused at the chair's next sitting with nothing changed since its spot-check: %v", err)
	}

	// The report moves: the area holds a change no spot-check has borne.
	b.ingest()
	openChairSitting(b)
	if !holdsStale(t, b) {
		t.Error("a change after the spot-check left the area covered — the chair has not borne it")
	}
}

// holdsStale is whether the PASS gate's list and the plan both hold the stale area.
func holdsStale(t *testing.T, b *stage) bool {
	t.Helper()
	run := fresh(b)
	blockers, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	gate := false
	for _, bl := range blockers {
		if bl.Kind == BlockerStaleArea && bl.Subject == evLens {
			gate = true
		}
	}
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	planned := len(plan.StaleAreas) > 0
	if gate != planned {
		t.Errorf("the gate holds the stale area %v and the plan lists it %v — one fold, two answers", gate, planned)
	}
	return gate
}

// fresh seeds the stage's events, as they stand, into a run of their own: a stage is seeded once.
func fresh(b *stage) Run {
	b.runDir = newRun(b.t)
	return b.seed()
}
