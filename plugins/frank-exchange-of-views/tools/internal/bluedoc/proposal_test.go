package bluedoc

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// THE LINE BETWEEN AN AUDIT AND AUTHORSHIP. A substantive addition is blue's to write.
func TestRefuseAuthoringRefusesASubstantiveAddition(t *testing.T) {
	big := "Five verification approaches agree. " + strings.Repeat("Furthermore the residual risks are enumerated as follows. ", 4)
	err := RefuseAuthoring("Five independent verification", big)
	if err == nil {
		t.Fatal("red was allowed to hand blue a substantive addition as concrete text")
	}
	for _, want := range []string{"ADDITION", "--fix"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal must say why and where the substance belongs (%q): %v", want, err)
		}
	}
}

// The bound is a THRESHOLD, so both sides of it are pinned — otherwise a later refactor can
// move it silently and only the loose direction is ever noticed.
func TestProposalGrowthBoundIsExactlyWhereItSays(t *testing.T) {
	const old = "Five independent verification"
	if err := RefuseAuthoring(old, old+strings.Repeat("x", MaxProposalGrowth)); err != nil {
		t.Errorf("a proposal exactly at the +%d bound was refused: %v", MaxProposalGrowth, err)
	}
	if err := RefuseAuthoring(old, old+strings.Repeat("x", MaxProposalGrowth+1)); err == nil {
		t.Errorf("a proposal one character past the +%d bound was accepted", MaxProposalGrowth)
	}
}

// Shrinking is always fine: overclaims and contradictions are usually removed, not added to.
func TestRefuseAuthoringAllowsAnyShrink(t *testing.T) {
	if err := RefuseAuthoring("Five independent verification approaches agree that 7 is prime", "7 is prime"); err != nil {
		t.Fatalf("a large DELETION was refused: %v", err)
	}
}

// A replacement keeps the tokens of the span it replaces; they are not words it adds. Measured with
// them, a 70-character rewording of a sentence five citations back reads as 180 and is refused.
func TestRefuseAuthoringCountsWordsNotTokens(t *testing.T) {
	const words = "Costs rose sharply in the first quarter."
	marked := "Costs rose sharply in the first quarter"
	for _, id := range []string{"C-0000000a", "C-0000000b", "C-0000000c", "C-0000000d", "C-0000000e"} {
		marked += anchor.Token(id)
	}
	marked += "."
	reworded := "Read again, with the measured figures for every month of the quarter, " + marked
	if err := RefuseAuthoring(words, reworded); err != nil {
		t.Errorf("a %d-character rewording carrying its sentence's five tokens was refused: %v",
			utf8.RuneCountInString(reworded)-utf8.RuneCountInString(marked), err)
	}
	if err := RefuseAuthoring(marked, words+strings.Repeat("x", MaxProposalGrowth+1)); err == nil {
		t.Error("a proposal past the bound in words was accepted because its old span carried tokens")
	}
}
