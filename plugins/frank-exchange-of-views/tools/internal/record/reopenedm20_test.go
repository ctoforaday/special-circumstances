package record

import (
	"slices"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// m20Archive is universe-m20: 463 events at the epoch this binary reads. Blue's edit at event 414
// reopened proof P-9e198e9a, last judged at event 383.
const m20Archive = "2026-10-11_is-91-prime-m20.tar.gz"

// THE PROOF'S OWNER IS READIED FOR A REOPENED PROOF, AND ANY LENS'S LATER CHECK CLEARS IT.
//
// In m20 two lenses opened sittings after the edit (the dark-side lens at event 427, the evidence
// lens at 428). The dark-side lens judged the proof at event 430 and the evidence lens at 434.
// Every surface said the PASS waited "until red-lens-evidence re-runs it" while the fold had
// cleared the anchor three events earlier, on another seat's check. The fold is the rule: a check
// is a reading from a sitting opened after the edit, whichever lens makes it. What the surfaces
// state is who owes it and is readied, and what clears it — both, in the same sentence. The proof is
// the computation lens's (anchorOwnerOf), and neither lens that sat owes it.
//
// The event ids are this archive's rows, read by SQL outside the fold (blue_edit_reopened, and
// reproduce joined to events on the proof's anchor).
func TestM20AnotherLensClearsWhatTheOwnerIsReadiedFor(t *testing.T) {
	whole := archivedEvents(t, m20Archive)
	if len(whole) != 463 || whole[413].GetType() != recordpb.EventType_EVENT_TYPE_BLUE_EDIT ||
		whole[429].GetSeatId() != "red-lens-dark-side" || whole[429].GetType() != recordpb.EventType_EVENT_TYPE_REPRODUCE ||
		whole[433].GetSeatId() != evLens || whole[433].GetType() != recordpb.EventType_EVENT_TYPE_REPRODUCE {
		t.Fatalf("the archive is not the m20 this test reads: %d events, 414 a %s, 430 %s's %s, 434 %s's %s", len(whole),
			whole[413].GetType(), whole[429].GetSeatId(), whole[429].GetType(), whole[433].GetSeatId(), whole[433].GetType())
	}
	const fact = "P-9e198e9a's sentence changed at edit 414 and no sitting opened since has checked it (its last check is event 383)"

	// After both lenses opened their sittings, before either re-ran the proof.
	run := archiveAt(t, m20Archive, whole, 429)
	var owners []string
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerUnchecked {
			owners = append(owners, b.Subject+" "+b.Owner)
		}
	}
	// P-337c380d is a proof nobody has judged yet, and is unchecked as well (gblock, 2026-10-25).
	if !slices.Equal(owners, []string{"P-9e198e9a " + computationSeat, "P-337c380d " + computationSeat}) {
		t.Fatalf("before any re-run the unchecked blockers are %q, want P-9e198e9a and P-337c380d owned by %s", owners, computationSeat)
	}
	// The plan readies the owner, and the chair's item says, in the words a seat reads, who owes the
	// check and that the owner is not the only seat whose check clears.
	plan, err := PlanDispatch(run)
	if err != nil {
		t.Fatal(err)
	}
	if _, ready := partyOf(plan, computationSeat); !ready {
		t.Errorf("the plan readies %+v, want %s for the proof it owns", plan.Parties, computationSeat)
	}
	chair := strings.Join(blockingItems(sittingOfRunT(t, run, "chair", "red-chair")), "\n")
	for _, want := range []string{fact + " — PASS is refused until a lens checks it from a sitting opened after the edit, by judging whether its script establishes THIS sentence",
		computationSeat + " owes that check and is the seat readied for it", "another lens's check clears it as well"} {
		if !strings.Contains(chair, want) {
			t.Errorf("the chair's work list does not say %q:\n%s", want, chair)
		}
	}
	// Neither lens open since the edit owes it, so neither's work list blocks on it.
	for _, seat := range []string{evLens, "red-lens-dark-side"} {
		if other := strings.Join(blockingItems(sittingOfRunT(t, run, "lens", seat)), "\n"); strings.Contains(other, "P-9e198e9a") {
			t.Errorf("%s's work list blocks on an anchor it does not owe:\n%s", seat, other)
		}
	}

	// The dark-side lens's judgement, event 430: the anchor is checked, four events before the
	// evidence lens's own.
	for _, b := range mustBlockers(t, archiveAt(t, m20Archive, whole, 430)) {
		if b.Kind == BlockerUnchecked && b.Subject == "P-9e198e9a" {
			t.Errorf("%s still holds the PASS after the dark-side lens re-ran it from a sitting opened after the edit", b.Subject)
		}
	}
}

func mustBlockers(t *testing.T, run Run) []Blocker {
	t.Helper()
	bs, err := PassBlockers(run)
	if err != nil {
		t.Fatal(err)
	}
	return bs
}
