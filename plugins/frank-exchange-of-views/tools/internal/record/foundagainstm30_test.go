package record

import (
	"slices"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// m30Archive is universe-m30, whose capture failed its assembly screen on a citation the debate had
// already answered.
const m30Archive = "2026-10-19_is-91-prime-m30.tar.gz"

// M30'S REFUTATION WAS ANSWERED, AND THE STANDING SAYS SO. The adversary refuted C-4f4b8cfa at event
// 427, against a sentence that inverted the definition; blue rewrote it at 462, reopening the
// citation; the adversary and the evidence lens read the new sentence as supported at 479 and 482.
// Red finds against C-4f4b8cfa from 427 until the edit, and not after it — while every citation of
// the same url (C-3b8d7782, C-f0ae536b) is never found against at all.
func TestM30RefutationIsAnsweredByTheEditAndTheRecheck(t *testing.T) {
	whole := archivedEvents(t, m30Archive)
	if v := whole[426].GetVerify(); v.GetAnchor() != "C-4f4b8cfa" || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES {
		t.Fatalf("event 427 is not the adversary's refutation of C-4f4b8cfa: %v", whole[426])
	}
	if e := whole[461].GetBlueEdit(); !slices.Contains(e.GetReopened(), "C-4f4b8cfa") {
		t.Fatalf("event 462 is not the edit reopening C-4f4b8cfa: %v", whole[461])
	}
	at := func(n int) []string {
		run := archiveAt(t, m30Archive, whole, n)
		fam, err := FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		return CitationsFoundAgainst(fam.Events, fam.At)
	}
	if got := at(426); len(got) != 0 {
		t.Errorf("before the refutation, red finds against %v, want none", got)
	}
	if got := at(461); !slices.Equal(got, []string{"C-4f4b8cfa"}) {
		t.Errorf("after the refutation and before the edit, red finds against %v, want [C-4f4b8cfa]", got)
	}
	if got := at(len(whole)); len(got) != 0 {
		t.Errorf("at the end of the run, red finds against %v, want none — the refutation was of a sentence the edit replaced, and the new one was read as supported", got)
	}
}

// SOURCES_REFUTED IS THE STANDING, AS THE ASSEMBLY SCREEN READS IT (gblock, 2026-10-19). The same
// m30 points: after the refutation and before the edit, `show evidence` counts C-4f4b8cfa against;
// at the end of the run — the refuted sentence replaced and the new one read as supported — it
// counts nothing, though the refuting row is still in the verified list.
func TestM30SourcesRefutedCountsOnlyTheStanding(t *testing.T) {
	whole := archivedEvents(t, m30Archive)
	at := func(n int) EvidenceJSON {
		fam, err := FamilyOf(archiveAt(t, m30Archive, whole, n))
		if err != nil {
			t.Fatal(err)
		}
		return EvidenceJSONOf(fam.Events, fam.At)
	}
	if got := at(461).Counts.SourcesRefuted; got != 1 {
		t.Errorf("after the refutation and before the edit, sources_refuted = %d, want 1 (C-4f4b8cfa)", got)
	}
	end := at(len(whole))
	if got := end.Counts.SourcesRefuted; got != 0 {
		t.Errorf("at the end of the run, sources_refuted = %d, want 0 — the refutation was of a sentence the edit replaced, and the re-check backs the new one", got)
	}
	refutedRow := false
	for _, s := range end.Sources {
		if s.Anchor == "C-4f4b8cfa" {
			refutedRow = slices.ContainsFunc(s.Verified, EvidenceVerificationJSON.Refuted)
		}
	}
	if !refutedRow {
		t.Error("C-4f4b8cfa's refuting row is gone from the verified list — the count reads the standing, the rows keep the history")
	}
}
