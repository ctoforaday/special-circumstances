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
// 401, against a sentence that inverted the definition; blue rewrote it at 431, reopening the
// citation; the adversary and the evidence lens read the new sentence as supported at 447 and 450.
// Red finds against C-4f4b8cfa from 401 until the edit, and not after it — while every citation of
// the same url (C-3b8d7782, C-f0ae536b) is never found against at all.
func TestM30RefutationIsAnsweredByTheEditAndTheRecheck(t *testing.T) {
	whole := archivedEvents(t, m30Archive)
	if v := whole[400].GetVerify(); v.GetAnchor() != "C-4f4b8cfa" || v.GetOutcome() != recordpb.SourceOutcome_SOURCE_OUTCOME_REFUTES {
		t.Fatalf("event 401 is not the adversary's refutation of C-4f4b8cfa: %v", whole[400])
	}
	if e := whole[430].GetBlueEdit(); !slices.Contains(e.GetReopened(), "C-4f4b8cfa") {
		t.Fatalf("event 431 is not the edit reopening C-4f4b8cfa: %v", whole[430])
	}
	at := func(n int) []string {
		run := archiveAt(t, m30Archive, whole, n)
		fam, err := FamilyOf(run)
		if err != nil {
			t.Fatal(err)
		}
		return CitationsFoundAgainst(fam.Events, fam.At)
	}
	if got := at(400); len(got) != 0 {
		t.Errorf("before the refutation, red finds against %v, want none", got)
	}
	if got := at(430); !slices.Equal(got, []string{"C-4f4b8cfa"}) {
		t.Errorf("after the refutation and before the edit, red finds against %v, want [C-4f4b8cfa]", got)
	}
	if got := at(len(whole)); len(got) != 0 {
		t.Errorf("at the end of the run, red finds against %v, want none — the refutation was of a sentence the edit replaced, and the new one was read as supported", got)
	}
}
