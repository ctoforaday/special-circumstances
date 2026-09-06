package cli

import (
	"strings"
	"testing"
)

// AN ABSENCE HAS NO SENTENCE TO QUOTE, and borrowing an innocent one as a handle is what the verb
// used to force. Measured in research/2026-09-02_quadratic-formula: a missing line of inquiry was
// pinned to the Diophantus sentence the finding itself called FINE, and a missing risk matrix —
// the word "risk" occurs zero times in that report — was pinned to the opening of section F.
// "A reader of the gap list will land on good prose and have to read three paragraphs to learn
// that the sentence is only a handle."
func TestAFindingAboutAnAbsenceNeedsNoBorrowedQuote(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L6",
		"--key", "F1", "--about-kind", "section", "--about", "Risk matrix",
		"--reason", "the template names a graded risk matrix and the report has no such section",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a finding about a MISSING section was refused: %v", err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-merge-r1", "findings")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Risk matrix") {
		t.Errorf("the findings view does not carry what the finding is about:\n%s", out)
	}
}

// AN ANCHOR IS REQUIRED, AND THE REFUSAL TEACHES THE ALTERNATIVE. A finding anchored to nothing
// cannot be found; one that only knows how to ask for a quote sends a lens back to borrowing one.
func TestAFindingWithNoAnchorIsRefusedAndOffersTheOtherKind(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L6",
		"--key", "F2", "--reason", "something is missing",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("a finding with neither --quote nor --about was accepted — it anchors to nothing")
	}
	for _, want := range []string{"--quote", "--about"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not offer %s:\n%v", want, err)
		}
	}
}

// TWO ANCHORS IS TWO SUBJECTS, and the gap this becomes would inherit the ambiguity.
func TestAFindingCannotClaimBothAnchors(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L6",
		"--key", "F3", "--quote", "§1", "--about-kind", "section", "--about", "Risk matrix",
		"--reason", "both", "--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil || !strings.Contains(err.Error(), "not both") {
		t.Errorf("a finding claiming a quote AND an about was accepted: %v", err)
	}
}

// THE REFERENCE IS CHECKED, WHICH A BORROWED QUOTE NEVER WAS. An avenue id either names a line
// this run proposed or it does not — that is the whole value of anchoring here rather than at
// whatever sentence happened to be nearby.
func TestAnAboutReferenceIsCheckedAgainstTheRecord(t *testing.T) {
	runDir := seatRun(t)
	_, err := run(t, "finding", "--run", runDir, "--seat-id", "red-lens-r1-L5",
		"--key", "F4", "--about-kind", "inquiry", "--about", "Q99",
		"--reason", "the decline reason is a category error",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("a finding anchored to an avenue id nothing proposed was accepted")
	}
}
