package record

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func hasItem(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if it.Blocks && strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

// listed is hasItem for an AFFORDANCE — availableOf's lines carry Blocks:false, open work rather
// than owed work, so a blocking-only match can never see them.
func listed(s SittingJSON, sub string) bool {
	for _, it := range s.Open {
		if strings.Contains(it.What, sub) {
			return true
		}
	}
	return false
}

func (b *stage) revision(seat string) *stage {
	return b.add(seat, &recordpb.Revision{Text: proto.String("the sitting's edits")})
}

func (b *stage) closeGap(lens, gap string) *stage {
	return b.add(lens, &recordpb.Close{GapId: proto.String(gap),
		ClosureClass: recordtest.P(recordpb.Disposition_DISPOSITION_REPAIRED), Prose: proto.String("verified at the leaf")})
}

// THE LOG IS OWED EVERY SITTING. It read the whole record, so a seat's first log entry discharged
// every later sitting's.
func TestTheLogIsOwedEverySitting(t *testing.T) {
	first := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).logNominal(evLens)
	}
	if hasItem(sittingOfRunT(t, first().seed(), "lens", evLens), "the log is open") {
		t.Fatal("a lens that logged this sitting is told its log is open")
	}
	again := first().register("red-chair").dispatch(2, evLens).register(evLens)
	if !hasItem(sittingOfRunT(t, again.seed(), "lens", evLens), "the log is open") {
		t.Fatal("a lens in its second sitting, with only its first sitting's log on the record, is not told the log is open")
	}
}

// BLUE-RESPOND OWES A REVISION FOR EACH SITTING THAT HAD SOMETHING TO ANSWER, and none for a
// sitting that found every engaged gap closed before it sat (gblock's ruling, 2026-09-11). In B9
// blue sat in epoch 5 with G4 open, filed no revision, and read complete because its epoch-4
// revision was on the record.
func TestBlueRespondsRevisionIsOwedPerSittingThatHadAGapOpen(t *testing.T) {
	sat := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").
			register("red-chair").dispatch(2, "blue-respond", "G1").register("blue-respond").revision("blue-respond")
	}
	const missing = "this sitting's revision is missing"
	if hasItem(sittingOfRunT(t, sat().seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue revised this sitting and is told its revision is missing")
	}
	again := sat().register("red-chair").dispatch(2, "blue-respond", "G1").register("blue-respond")
	if !hasItem(sittingOfRunT(t, again.seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue's second sitting on an open G1 has no revision, and the work list does not say so")
	}
	closedFirst := sat().register("red-chair").dispatch(2, "blue-respond", "G1").closeGap(evLens, "G1").register("blue-respond")
	if hasItem(sittingOfRunT(t, closedFirst.seed(), "blue", "blue-respond"), missing) {
		t.Fatal("blue's sitting found G1 closed before it sat and is still told a revision is owed")
	}
}

// THE CHAIR RE-SAMPLES THE ARCHIVE EVERY SITTING it is not empty — its item already said "this
// sitting has sampled none of it", and read the whole record.
func TestTheSpotCheckIsOwedEverySittingTheArchiveHoldsAClosure(t *testing.T) {
	sampled := func() *stage {
		return newStage(t).cast(evLens, "red-chair", "blue-respond", "judge").ingest().
			register("red-chair").dispatch(2, evLens).register(evLens).mint(evLens, "G1", "medium").closeGap(evLens, "G1").
			register("red-chair").add("red-chair", &recordpb.SpotCheck{Ids: []string{"G1"}, Reason: proto.String("the anchor still resolves")})
	}
	const unsampled = "this sitting has sampled none of it"
	if listed(sittingOfRunT(t, sampled().seed(), "chair", "red-chair"), unsampled) {
		t.Fatal("the chair sampled the archive this sitting and is told it has not")
	}
	if !listed(sittingOfRunT(t, sampled().register("red-chair").seed(), "chair", "red-chair"), unsampled) {
		t.Fatal("the chair's next sitting has sampled nothing, and the work list does not say so")
	}
}

// Every other blue seat sits once for its act, and owes that sitting's revision.
func TestTheSynthesizerOwesItsSittingsRevision(t *testing.T) {
	b := newStage(t).cast(evLens, "red-chair", "blue-respond", "blue-synthesize", "judge").register("blue-synthesize")
	if !hasItem(sittingOfRunT(t, b.seed(), "blue", "blue-synthesize"), "this sitting's revision is missing") {
		t.Fatal("the synthesizer filed no revision and is not told it is missing")
	}
}
