package record

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// ONE LENS OWNS A PROOF NOBODY JUDGED (anchorOwnerOf): the computation lens. Its work list is the
// only one offering the proof; the proof stays on the record for every lens, `not_judged`.
func TestAProofIsOfferedToItsOwnerAlone(t *testing.T) {
	for _, c := range []struct {
		name  string
		cast  []string
		owner string
	}{
		{"computation cast", []string{compLens, evLens, outsideLens}, compLens},
	} {
		t.Run(c.name, func(t *testing.T) {
			seats := append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")
			b := newStage(t).cast(seats...).ingest().register("red-chair").proof("P-1", "s1")
			for _, lens := range c.cast {
				b.register(lens)
			}
			run := b.seed()
			for _, lens := range c.cast {
				_, listed := itemNaming(sittingOfRunT(t, run, "lens", lens), "proof P-1 is recorded and nobody has re-run it")
				if want := lens == c.owner; listed != want {
					t.Errorf("%s's work list offers the unjudged proof = %v, want %v", lens, listed, want)
				}
			}
			// The proof is on the record for every lens: `show evidence` lists it, not judged.
			raw, err := EvidenceJSONBytes(run)
			if err != nil {
				t.Fatal(err)
			}
			var ev EvidenceJSON
			if err := json.Unmarshal(raw, &ev); err != nil || len(ev.Proofs) != 1 || ev.Proofs[0].Soundness != SoundnessNotJudged {
				t.Errorf("show evidence proofs = %+v (%v), want P-1 listed not judged", ev.Proofs, err)
			}
		})
	}
}

// ONE LENS OWNS A CITATION NOBODY VERIFIED (anchorOwnerOf, gblock 2026-10-25): the evidence lens. Its
// work list is the only one offering the citation, and the citation keeps it from retiring
// (lensOwesOf); `show evidence` lists it with no verification, for every lens.
func TestACitationIsOfferedToItsOwnerAlone(t *testing.T) {
	for _, c := range []struct {
		name  string
		cast  []string
		owner string
	}{
		{"evidence cast", []string{compLens, evLens, outsideLens}, evLens},
	} {
		t.Run(c.name, func(t *testing.T) {
			seats := append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")
			b := newStage(t).cast(seats...).ingestCited().register("red-chair").cite("C-1")
			for _, lens := range c.cast {
				b.register(lens)
			}
			run := b.seed()
			fam, err := FamilyOf(run)
			if err != nil {
				t.Fatal(err)
			}
			ids := fam.At.IDs(fam.Events)
			debts := lensOwesOf(fam.Events, ids, fam.At, nil, owedBlockersOf(fam.Events, ids, fam.At, nil))
			for _, lens := range c.cast {
				_, listed := itemNaming(sittingOfRunT(t, run, "lens", lens), "citation C-1 is on the record and nobody has verified it")
				want := lens == c.owner
				if listed != want {
					t.Errorf("%s's work list offers the unverified citation = %v, want %v", lens, listed, want)
				}
				if owes := slices.Contains(debts[lens], "unverified citation C-1"); owes != want {
					t.Errorf("%s owes the unverified citation = %v, want %v: %q", lens, owes, want, debts[lens])
				}
			}
			raw, err := EvidenceJSONBytes(run)
			if err != nil {
				t.Fatal(err)
			}
			var ev EvidenceJSON
			if err := json.Unmarshal(raw, &ev); err != nil || len(ev.Sources) != 1 || ev.Sources[0].Anchor != "C-1" || len(ev.Sources[0].Verified) != 0 {
				t.Errorf("show evidence sources = %+v (%v), want C-1 listed with no verification", ev.Sources, err)
			}
		})
	}
}

// ONE DEFINITION OF WHO OWES AN ANCHOR'S CHECK. For each anchor kind, the lens a never-checked
// anchor's work-list item goes to, the lens it keeps from retiring, and the owner of the blocker a
// reopened one raises are one seat.
func TestOneOwnerPerAnchorKind(t *testing.T) {
	for _, c := range []struct {
		name                 string
		cast                 []string
		citation, proofOwner string
	}{
		{"every lens", []string{compLens, evLens, outsideLens}, evLens, compLens},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := anchorOwnerOf(false); got != c.citation {
				t.Errorf("anchorOwnerOf(citation) = %q, want %q", got, c.citation)
			}
			if got := anchorOwnerOf(true); got != c.proofOwner {
				t.Errorf("anchorOwnerOf(proof) = %q, want %q", got, c.proofOwner)
			}
			seats := append(append([]string{}, c.cast...), "red-chair", "blue-respond", "judge")
			// C-1 and P-1 nobody checked; C-2 and P-2 checked by a lens that owns neither, then reopened.
			b := newStage(t).cast(seats...).ingestCited().register("red-chair")
			for _, lens := range c.cast {
				b.sit(2, lens)
			}
			b.cite("C-1").proof("P-1", "s1").cite("C-2").supports(outsideLens, "C-2").
				proof("P-2", "s2").reproduce(outsideLens, "P-2", "s2", recordpb.Rerun_RERUN_REPRODUCED, recordpb.Soundness_SOUNDNESS_SOUND).reopen("C-2", "P-2")
			run := b.seed()
			fam, err := FamilyOf(run)
			if err != nil {
				t.Fatal(err)
			}
			ids := fam.At.IDs(fam.Events)
			debts := lensOwesOf(fam.Events, ids, fam.At, nil, owedBlockersOf(fam.Events, ids, fam.At, nil))
			reopened := map[string]string{}
			for _, bl := range reopenedOf(t, run) {
				reopened[bl.Subject] = bl.Owner
			}
			for _, k := range []struct {
				owner, never, item, debt, again string
			}{
				{c.citation, "C-1", "citation C-1 is on the record and nobody has verified it", "unverified citation C-1", "C-2"},
				{c.proofOwner, "P-1", "proof P-1 is recorded and nobody has re-run it", "unjudged proof P-1", "P-2"},
			} {
				if got := reopened[k.again]; got != k.owner {
					t.Errorf("reopened %s's blocker is owned by %q, want %q", k.again, got, k.owner)
				}
				for _, lens := range c.cast {
					want := lens == k.owner
					if _, listed := itemNaming(sittingOfRunT(t, run, "lens", lens), k.item); listed != want {
						t.Errorf("%s's work list offers %s = %v, want %v", lens, k.never, listed, want)
					}
					if owes := slices.Contains(debts[lens], k.debt); owes != want {
						t.Errorf("%s owes %s = %v, want %v: %q", lens, k.never, owes, want, debts[lens])
					}
				}
			}
		})
	}
}

// compLens is the computation lens: the owner of every unjudged proof.
const compLens = "red-lens-computation"
