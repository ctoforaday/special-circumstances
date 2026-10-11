package record

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// m20Archive is universe-m20: 437 events at the epoch this binary reads. Blue's edit at event 389
// reopened proof P-9e198e9a, last re-run at event 359.
const m20Archive = "2026-10-11_is-91-prime-m20.tar.gz"

// THE EVIDENCE LENS IS READIED FOR A REOPENED ANCHOR, AND ANY LENS'S LATER CHECK CLEARS IT.
//
// In m20 both lenses opened sittings after the edit (the dark-side lens at event 403, the evidence
// lens at 404). The dark-side lens re-ran the proof at event 406 and the evidence lens at 409.
// Every surface said the PASS waited "until red-lens-evidence re-runs it" while the fold had
// cleared the anchor three events earlier, on another seat's check. The fold is the rule: a check
// is a reading from a sitting opened after the edit, whichever lens makes it. What the surfaces
// state is who owes it and is readied, and what clears it — both, in the same sentence.
//
// The event ids are this archive's rows, read by SQL outside the fold (blue_edit_reopened, and
// reproduce joined to events on the proof's sha).
func TestM20AnotherLensClearsWhatTheEvidenceLensIsReadiedFor(t *testing.T) {
	archived, err := MergedEvents(mustRun(t, recordtest.RequireArchive(t, m20Archive)))
	if err != nil {
		t.Fatal(err)
	}
	whole := archived.Events
	if len(whole) != 437 || whole[388].GetType() != recordpb.EventType_EVENT_TYPE_BLUE_EDIT ||
		whole[405].GetSeatId() != "red-lens-dark-side" || whole[405].GetType() != recordpb.EventType_EVENT_TYPE_REPRODUCE ||
		whole[408].GetSeatId() != evLens || whole[408].GetType() != recordpb.EventType_EVENT_TYPE_REPRODUCE {
		t.Fatalf("the archive is not the m20 this test reads: %d events, 389 a %s, 406 %s's %s, 409 %s's %s", len(whole),
			whole[388].GetType(), whole[405].GetSeatId(), whole[405].GetType(), whole[408].GetSeatId(), whole[408].GetType())
	}
	const fact = "P-9e198e9a's sentence changed at edit 389 and no sitting opened since has checked it (its last check is event 359)"

	// After both lenses opened their sittings, before either re-ran the proof.
	run := archiveAt(t, m20Archive, whole, 405)
	var owners []string
	for _, b := range mustBlockers(t, run) {
		if b.Kind == BlockerReopened {
			owners = append(owners, b.Subject+" "+b.Owner)
		}
	}
	if len(owners) != 1 || owners[0] != "P-9e198e9a "+evLens {
		t.Fatalf("before any re-run the reopened blockers are %q, want P-9e198e9a owned by %s", owners, evLens)
	}
	// The evidence lens's work list blocks on it; the dark-side lens's, open since the same edit, does not.
	owed := strings.Join(blockingItems(sittingOfRunT(t, run, "lens", evLens)), "\n")
	for _, want := range []string{fact + " — PASS is refused until a lens checks it from a sitting opened after the edit, by re-running it (`reproduce --id ea091047a1ce", reopenedOwed} {
		if !strings.Contains(owed, want) {
			t.Errorf("the evidence lens's work list does not say %q:\n%s", want, owed)
		}
	}
	if other := strings.Join(blockingItems(sittingOfRunT(t, run, "lens", "red-lens-dark-side")), "\n"); strings.Contains(other, "P-9e198e9a") {
		t.Errorf("the dark-side lens's work list blocks on an anchor it does not owe:\n%s", other)
	}

	// The dark-side lens's re-run, event 406: the anchor is checked, three events before the
	// evidence lens's own.
	for _, b := range mustBlockers(t, archiveAt(t, m20Archive, whole, 406)) {
		if b.Kind == BlockerReopened {
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
