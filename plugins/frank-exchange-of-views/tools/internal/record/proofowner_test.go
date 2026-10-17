package record

import (
	"encoding/json"
	"testing"
)

// ONE LENS OWNS A PROOF NOBODY JUDGED (proofOwnerOf): the computation lens where the cast seats it,
// else the evidence lens, else nobody. Its work list is the only one offering the proof; the proof
// stays on the record for every lens, `not_judged`, whoever owns it.
func TestAProofIsOfferedToItsOwnerAlone(t *testing.T) {
	for _, c := range []struct {
		name  string
		cast  []string
		owner string
	}{
		{"computation cast", []string{compLens, evLens, outsideLens}, compLens},
		{"no computation lens", []string{evLens, outsideLens}, evLens},
		{"neither lens", []string{outsideLens}, ""},
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
			// The proof is on the record whoever owns it: `show evidence` lists it, not judged.
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

// THE RULE ITSELF, one row per arm: a proof falls to the evidence lens only where the cast seats no
// computation lens, and to nobody where it seats neither.
func TestProofOwnerOf(t *testing.T) {
	for _, c := range []struct {
		cast []string
		want string
	}{
		{[]string{compLens, evLens}, compLens},
		{[]string{evLens, compLens}, compLens},
		{[]string{evLens}, evLens},
		{[]string{compLens}, compLens},
		{[]string{outsideLens}, ""},
		{nil, ""},
	} {
		if got := proofOwnerOf(c.cast); got != c.want {
			t.Errorf("proofOwnerOf(%q) = %q, want %q", c.cast, got, c.want)
		}
	}
}

// compLens is the computation lens: the owner of every unjudged proof where the cast seats it.
const compLens = "red-lens-computation"
