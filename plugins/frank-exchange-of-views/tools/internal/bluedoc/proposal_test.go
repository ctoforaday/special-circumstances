package bluedoc

import (
	"strings"
	"testing"
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
