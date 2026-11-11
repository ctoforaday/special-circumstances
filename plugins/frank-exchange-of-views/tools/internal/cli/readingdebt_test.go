package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A READING RECORDS, A GAP HOLDS (plans/gap-is-the-one-holder.md, phase 1b), driven through the
// verbs as a seat drives them.

// T1. Every gap flag is an unknown flag on `verify`, and a refused reading writes nothing.
func TestVerifyTakesNoGapFlag(t *testing.T) {
	b := readingRun(t)
	for _, f := range []string{"--class", "--problem", "--check", "--check-kind", "--severity", "--likelihood", "--impact", "--fix", "--new", "--complexity"} {
		_, err := b.verify(lensL, citeReading(b.c1, "refutes", f, "x")...)
		mustRefuse(t, err, "", "unknown flag: "+f)
	}
	if c := b.counts(); c != [3]int{} {
		t.Errorf("the refusals wrote verify/mint/anchor %v", c)
	}
}

// T2. A refutes reading of C writes one Verify and nothing else, names its event, and leaves no
// failed check on C.
func TestAReadingWritesOneEvent(t *testing.T) {
	b := readingRun(t)
	r := b.mustRead(lensL, citeReading(b.c1, "refutes")...)
	_, at := b.lastVerify()
	if r.Event != at || r.Event == 0 {
		t.Errorf("the result names event %d, the record holds the reading at %d", r.Event, at)
	}
	if c := b.counts(); c != [3]int{1, 0, 0} {
		t.Errorf("the reading wrote verify/mint/anchor %v, want one reading", c)
	}
	if fs := failedOf(t, b.runDir, b.c1); len(fs) != 0 {
		t.Errorf("a refutes reading left failed checks on C: %+v", fs)
	}
	if !strings.Contains(r.Debt, "reading of citation "+b.c1) {
		t.Errorf("the result's debt = %q, want this reading's own", r.Debt)
	}
	out, err := b.verify(lensM, citeReading(b.c2, "supports")...)
	if err != nil || !strings.Contains(out, "recorded as event") {
		t.Errorf("the human text does not name the event: %v\n%s", err, out)
	}
}

// T9. The debt holds the PASS by its own kind, and the mint discharges it: after the gap is graded
// low and closed not_a_defect, the PASS is admitted; `show evidence`'s counts carry no
// sources_refuted key throughout.
func TestAReadingDebtHoldsThePassUntilItsGap(t *testing.T) {
	b := readingRunCast(t, rdReport, 0, rdCited, rdBacked)
	counts := func() map[string]any {
		raw, err := record.EvidenceJSONBytes(runtest.Open(t, b.runDir))
		if err != nil {
			t.Fatal(err)
		}
		var ev struct {
			Counts map[string]any `json:"counts"`
		}
		if err := json.Unmarshal(raw, &ev); err != nil {
			t.Fatal(err)
		}
		return ev.Counts
	}
	// Every citation read, so nothing but the debt holds.
	b.mustRead(lensL, citeReading(b.c1, "refutes")...)
	b.mustRead(lensK, citeReading(b.c2, "supports")...)
	if _, ok := counts()["sources_refuted"]; ok {
		t.Error("show evidence's counts carry sources_refuted")
	}
	// The gate's list: the debt holds the PASS by its own kind, naming X and C, and no failed check
	// stands on C.
	holds := func() (debt, failed, material bool) {
		bs, err := record.PassBlockers(runtest.Open(t, b.runDir))
		if err != nil {
			t.Fatal(err)
		}
		for _, bl := range bs {
			switch bl.Kind {
			case record.BlockerReadingDebt:
				debt = debt || (bl.Owner == lensL && bl.Subject == b.c1)
			case record.BlockerFailedCheck:
				failed = failed || bl.Subject == b.c1
			case record.BlockerMaterialGap:
				material = true
			}
		}
		return
	}
	if debt, failed, _ := holds(); !debt || failed {
		t.Fatalf("after the refutes reading: reading_debt %v, failed_check on C %v — want the debt alone", debt, failed)
	}
	if _, err := run(t, "register", "--run", b.runDir, "--seat-id", "red-chair"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "verdict", "--run", b.runDir, "--seat-id", "red-chair", "--as", "PASS")
	mustRefuse(t, err, "", "reading of citation "+b.c1, lensL)
	out, err := run(t, append([]string{"mint", "--run", b.runDir, "--seat-id", lensL, "--class", "claim_contradicts_source",
		"--problem", "the source says otherwise than the sentence", "--fix", "the fix", "--check-kind", "document",
		"--check", "the sentence states what the source says", "--severity", "low", "--likelihood", "low", "--impact", "low"}, about(b.c1)...)...)
	if err != nil {
		t.Fatalf("the gap on the reading: %v\n%s", err, out)
	}
	g := gapID(out)
	// The class is material whatever its grade: from the mint on, the gap's materiality holds the PASS.
	if debt, _, material := holds(); debt || !material {
		t.Errorf("after the mint: reading_debt %v, material gap %v — want the debt discharged and the gap holding", debt, material)
	}
	if _, err := run(t, "close", "--run", b.runDir, "--seat-id", lensL, "--id", g, "--as", "not_a_defect", "--reason", "blue's argument held",
		"--verified-by", lensL, "--verified-with", "a reading of the source", "--verified-against", rdCited); err != nil {
		t.Fatalf("the close: %v", err)
	}
	if debt, failed, material := holds(); debt || failed || material {
		t.Errorf("after the close: reading_debt %v, failed_check %v, material %v — want none", debt, failed, material)
	}
	if _, ok := counts()["sources_refuted"]; ok {
		t.Error("show evidence's counts carry sources_refuted")
	}
}

// T22/T23's mint halves: a found-source reading whose sentence blue reworded stands behind no mint
// until the lens reads the new sentence; a found source refuting text blue applied from a red fix is
// estopped, and mints with --supersedes.
func TestAMintOnALostReadingWaitsForTheReRead(t *testing.T) {
	cc := "claim_contradicts_source"
	const newText = "Volume recovered much later in the year."
	b := readingRun(t)
	b.mustRead(lensL, urlReading(rdU3, rdFree1, "refutes")...)
	if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", rdFree1, "--new", newText, "--reason", "reworded"); err != nil {
		t.Fatal(err)
	}
	_, out, err := b.mintSource(lensL, cc, "--quote", newText)
	mustRefuse(t, err, out, standingRefusal, "(none)")
	b.sit(lensL)
	b.mustRead(lensL, urlReading(rdU3, newText, "refutes")...)
	if _, out, err := b.mintSource(lensL, cc, "--quote", newText); err != nil {
		t.Fatalf("the mint on the re-read: %v\n%s", err, out)
	}
}

// T28. The retry key reads the sentence: a reworded sentence, same url, same seat, same sitting is
// still a retry; the next sitting admits it; the sentence cut and another read admits; an identical
// claim is refused as before.
func TestAFoundSourceRetryIsKeyedOnTheSentence(t *testing.T) {
	const reworded = "Volume recovered much later in the year."
	b := readingRun(t)
	b.mustRead(lensL, urlReading(rdU3, rdFree1, "refutes")...)
	_, first := b.lastVerify()
	_, err := b.verify(lensL, urlReading(rdU3, rdFree1, "refutes")...)
	mustRefuse(t, err, "", fmt.Sprintf("at event %d", first))
	if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", rdFree1, "--new", reworded, "--reason", "reworded"); err != nil {
		t.Fatal(err)
	}
	editAt := countType(t, b.runDir, recordpb.EventType_EVENT_TYPE_BLUE_EDIT)
	_ = editAt
	_, err = b.verify(lensL, urlReading(rdU3, reworded, "supports")...)
	mustRefuse(t, err, "", fmt.Sprintf("at event %d", first), "reworded by blue's edit at event")
	if c := b.counts(); c != [3]int{1, 0, 0} {
		t.Errorf("the refused retry wrote verify/mint/anchor %v", c)
	}
	b.sit(lensL)
	b.mustRead(lensL, urlReading(rdU3, reworded, "supports")...)
	// A fragment edit at the floor, then the sentence cut and another sentence read: no retry.
	b.sit(lensL)
	b.mustRead(lensL, urlReading(rdU4, rdFree2, "refutes")...)
	if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", rdFree2+" ", "--new", "", "--reason", "cut"); err != nil {
		t.Fatal(err)
	}
	if _, err := b.verify(lensL, urlReading(rdU4, rdFree3, "refutes")...); err != nil {
		t.Errorf("the same url read against another sentence after the first was cut: %v", err)
	}
}

// T22, the sub-floor edit, through the verbs: blue swaps a phrase shorter than the estoppel floor
// inside the sentence a found source was read against. The debt stands naming no edit and the
// sentence the report now holds; a 32 quoting that sentence is refused until the lens reads it; the
// lens's `supports` of the same url against it clears the debt; its `refutes` is a fresh debt located
// on it, which the 32 quoting it clears.
func TestASubFloorEditLeavesADebtTheCurrentSentenceClears(t *testing.T) {
	const cc = "claim_contradicts_source"
	const phrase, swapped = "recovered later", "recovered much later"
	const current = "Volume recovered much later in the year."
	if len(phrase) >= 40 {
		t.Fatal("fixture: the phrase is at the estoppel floor")
	}
	debts := func(b readingBoard) []record.Blocker {
		t.Helper()
		bs, err := record.PassBlockers(runtest.Open(t, b.runDir))
		if err != nil {
			t.Fatal(err)
		}
		var out []record.Blocker
		for _, bl := range bs {
			if bl.Kind == record.BlockerReadingDebt && bl.Owner == lensL {
				out = append(out, bl)
			}
		}
		return out
	}
	stage := func(t *testing.T) readingBoard {
		t.Helper()
		b := readingRun(t)
		b.mustRead(lensL, urlReading(rdU3, rdFree1, "refutes")...)
		if _, err := run(t, "edit", "--run", b.runDir, "--seat-id", blueSeat, "--quote", phrase, "--new", swapped, "--reason", "a phrase sharpened"); err != nil {
			t.Fatal(err)
		}
		ds := debts(b)
		if len(ds) != 1 || ds[0].Subject != rdFree1 || ds[0].Lost() != 0 || ds[0].Reworded() != strings.TrimSuffix(current, ".") || !ds[0].Unlocated() {
			t.Fatalf("after the sub-floor edit the debts = %+v, want one on %q naming no edit and the sentence %q", ds, rdFree1, current)
		}
		if item := ds[0].WorkItem(); !strings.Contains(item, fmt.Sprintf("changed that sentence to %q", strings.TrimSuffix(current, "."))) || !strings.Contains(item, "against the sentence as it now stands") {
			t.Errorf("the item: %s", item)
		}
		_, out, err := b.mintSource(lensL, cc, "--quote", current)
		mustRefuse(t, err, out, standingRefusal)
		b.sit(lensL)
		return b
	}
	t.Run("supports clears", func(t *testing.T) {
		b := stage(t)
		b.mustRead(lensL, urlReading(rdU3, current, "supports")...)
		if ds := debts(b); len(ds) != 0 {
			t.Errorf("after the supports of the current sentence the debts = %+v", ds)
		}
	})
	t.Run("refutes is a fresh debt its gap clears", func(t *testing.T) {
		b := stage(t)
		b.mustRead(lensL, urlReading(rdU3, current, "refutes")...)
		ds := debts(b)
		if len(ds) != 1 || ds[0].Subject != current || ds[0].Lost() != 0 || ds[0].Reworded() != "" || ds[0].Unlocated() {
			t.Fatalf("after the refutes of the current sentence the debts = %+v, want one located on %q", ds, current)
		}
		if _, out, err := b.mintSource(lensL, cc, "--quote", current); err != nil {
			t.Fatalf("the mint on the re-read: %v\n%s", err, out)
		}
		if ds := debts(b); len(ds) != 0 {
			t.Errorf("after the gap quoting the current sentence the debts = %+v", ds)
		}
	})
}
