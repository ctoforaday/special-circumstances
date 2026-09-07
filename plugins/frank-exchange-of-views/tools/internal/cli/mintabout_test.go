package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A GAP ABOUT AN ABSENCE ANCHORS THE WAY THE FINDING DID, with the same two flags.
//
// `merge mint` is the next step of the act `lens finding` begins, and its help used to send an
// omission back to the handle the finding had just been taught to stop using: "for a gap about
// something MISSING, quote the sentence where it SHOULD be — that is how a lens finding anchors an
// omission". Once that stopped being how a finding anchors an omission, the sentence was a
// cross-verb instruction to do the wrong thing.
func TestAGapAboutAnAbsenceNeedsNoBorrowedQuote(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	if _, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G1", "--class", "scope-creep", "--about-kind", "section", "--about", "Risk matrix",
		"--problem", "the template names a graded risk matrix and the report has no such section",
		"--check-kind", "document", "--check", "the report carries a graded risk matrix",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a gap about a MISSING section was refused: %v", err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", "board")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Risk matrix") {
		t.Errorf("the board does not carry what the gap is about:\n%s", out)
	}
}

// ONE SUBJECT, ON BOTH VERBS. A gap that claims a quote AND an about is claiming two, which is the
// same refusal `lens finding` gives — the point of sharing the flags is that they behave the same.
func TestAGapCannotClaimBothAnchors(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G2", "--class", "scope-creep",
		"--quote", "Seven is prime.", "--about-kind", "section", "--about", "Risk matrix",
		"--problem", "p", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("a gap claiming a quote AND an about was accepted")
	}
	if !strings.Contains(err.Error(), "one subject") {
		t.Errorf("the refusal does not say why two anchors is wrong:\n%v", err)
	}
}

// THE REFERENCE IS CHECKED HERE TOO. The whole advantage over a borrowed quote is that the record
// can verify the target, and a gap that skipped the check would be the weaker half of one act.
func TestAGapsAboutReferenceIsCheckedAgainstTheRecord(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	_, err := run(t, "mint", "--run", runDir, "--seat-id", "red-merge-r1",
		"--key", "G3", "--class", "scope-creep", "--about-kind", "inquiry", "--about", "Q99",
		"--problem", "p", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("a gap anchored to an avenue id nothing proposed was accepted")
	}
}

// THE TWO VERBS OFFER THE SAME VOCABULARY, and this is the assertion that keeps them offering it.
//
// The flags could stay spelled alike while the closed sets behind them drifted — a value added for
// findings and not for gaps would leave a seat able to anchor a finding somewhere it cannot then
// anchor the gap, which is the half-state that reads as done. The sets are compared rather than
// re-listed, so a value added to one and not the other fails here rather than in a run.
func TestAFindingAndAGapAnchorToTheSameThings(t *testing.T) {
	finding, gap := record.MustEnum("finding", "about_kind"), record.MustEnum("mint", "about_kind")
	words := func(e record.EnumField) []string {
		var out []string
		for _, v := range e.Values {
			out = append(out, v.Name)
		}
		return out
	}
	f, g := strings.Join(words(finding), ","), strings.Join(words(gap), ",")
	if f != g {
		t.Errorf("`lens finding` and `merge mint` no longer anchor to the same things:\n  finding: %s\n  mint:    %s", f, g)
	}
	if finding.Flag != gap.Flag {
		t.Errorf("one anchor, two flag names: finding uses %q and mint uses %q", finding.Flag, gap.Flag)
	}
}
