package fuzz

import (
	"strings"
	"testing"
)

// A SECOND LENS'S READING OF A SENTENCE ANOTHER LENS'S GAP STANDS ON OWES NOTHING, AND ITS MINT IS
// REFUSED: the sibling's reading of a source it found refutes the sentence and its
// claim_contradicts_source mint quoting the sentence lands on that reading; the lens's own refuting
// reading of another source for the sentence is recorded with no debt — the open gap of the class
// on the sentence discharges it (other_seat_gap) — and the lens's mint quoting the sentence is
// refused by the one-gap rule naming the sibling's gap, its point filed as a finding about it. Run
// outside the release gate, on the drive's own sittings.
func TestASecondLenssReadingIsDischargedByTheFirstsGap(t *testing.T) {
	const (
		sentence = "A § fuzz sentence to anchor findings."
		lens     = "red-lens-logic"
		sibling  = "red-lens-adversary"
	)
	r := stagedRunner(t)
	if err := r.stagePart(lanePartText("blue-lane-1")); err != nil {
		t.Fatal(err)
	}
	r.sit("lens", sibling)
	at := r.readFound(sibling, foundSourceURL(sibling), sentence, "refutes")
	if at == 0 {
		t.Fatalf("the sibling's reading: %s", r.actErr)
	}
	res := r.mintOnReading(sibling, classContradicts, quoteArgs(sentence), at, "quote")
	if res.gap == "" || !r.debtDischargedBy(sibling, sentence, "mint", "the sibling's mint on its reading") {
		t.Fatalf("the sibling's mint on its reading: %q (%s)", res.out, r.actErr)
	}
	r.sit("lens", lens)
	if r.readFound(lens, foundSourceURL(lens), sentence, "refutes") == 0 {
		t.Fatalf("the lens's reading: %s", r.actErr)
	}
	if b, stands := r.debtOf(lens, sentence); stands {
		t.Fatalf("the lens's reading left a debt over the sibling's open gap %s: %s", res.gap, b.WorkItem())
	}
	second := r.mintOnReading(lens, classContradicts, quoteArgs(sentence), 0, "quote")
	if second.refusedBy != "one_gap" || !strings.Contains(second.out, res.gap) || r.actErr != "" {
		t.Fatalf("the lens's mint over the sibling's gap: refused by %q, naming %s %v (%s)", second.refusedBy, res.gap, strings.Contains(second.out, res.gap), r.actErr)
	}
	if r.oneGapRefused != 1 || r.oneGapFindings != 1 || r.reading.debtsDischarged["other_seat_gap"] != 1 {
		t.Errorf("one-gap refused %d, findings %d, debts discharged by another seat's gap %d — want 1 each", r.oneGapRefused, r.oneGapFindings, r.reading.debtsDischarged["other_seat_gap"])
	}
}

// The RUNSTAT line, which the job log keeps, names a refused act's cause: the envelope exec puts on
// the lines below the exit status is carried onto it.
func TestRunstatNamesTheRefusal(t *testing.T) {
	err := "an act the drive issued: refused, and the drive expects it admitted: verify --seat-id L: exit status 2\n" +
		`  {"verb":"verify","role":"lens","ok":false,"code":"validation","error":"L holds 1 gap(s) open and its mint budget admits no more."}`
	got := runstatErr(err)
	for _, want := range []string{"exit status 2", "refused (validation)", "its mint budget admits no more"} {
		if !strings.Contains(got, want) {
			t.Errorf("RUNSTAT err %q does not say %q", got, want)
		}
	}
	if strings.Contains(got, "\n") {
		t.Errorf("RUNSTAT err spans lines: %q", got)
	}
	if got := runstatErr("forced unsound computation: no plan readied"); got != "forced unsound computation: no plan readied" {
		t.Errorf("an error with no envelope = %q, want its first line", got)
	}
}
