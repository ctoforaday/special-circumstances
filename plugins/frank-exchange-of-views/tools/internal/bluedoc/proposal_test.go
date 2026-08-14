package bluedoc

import (
	"strings"
	"testing"
)

const report = "# H\n\nFive independent verification approaches agree that 7 is prime.\n\nThe sieve is fast<!--fx:f-abc123--> and simple.\n"

// A concrete proposal states an exact, unique, applyable span. Stating one is the forced
// re-read: it cannot be written from memory of what the report probably says.
func TestValidateProposalAcceptsAnApplyableTextualFix(t *testing.T) {
	if err := ValidateProposal("merge mint", report, "Five independent verification", "Five verification"); err != nil {
		t.Fatalf("a legal textual proposal was refused: %v", err)
	}
}

// THE LINE BETWEEN AN AUDIT AND AUTHORSHIP. A substantive addition is blue's to write.
func TestValidateProposalRefusesAuthoring(t *testing.T) {
	big := "Five verification approaches agree. " + strings.Repeat("Furthermore the residual risks are enumerated as follows. ", 4)
	err := ValidateProposal("merge mint", report, "Five independent verification", big)
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
	atLimit := old + strings.Repeat("x", MaxProposalGrowth)
	if err := ValidateProposal("merge mint", report, old, atLimit); err != nil {
		t.Errorf("a proposal exactly at the +%d bound was refused: %v", MaxProposalGrowth, err)
	}
	overLimit := old + strings.Repeat("x", MaxProposalGrowth+1)
	if err := ValidateProposal("merge mint", report, old, overLimit); err == nil {
		t.Errorf("a proposal one character past the +%d bound was accepted", MaxProposalGrowth)
	}
}

// Shrinking is always fine: overclaims and contradictions are usually removed, not added to.
func TestValidateProposalAllowsAnyShrink(t *testing.T) {
	if err := ValidateProposal("merge mint", report, "Five independent verification approaches agree that 7 is prime", "7 is prime"); err != nil {
		t.Fatalf("a large DELETION was refused: %v", err)
	}
}

// A span red cannot locate is a span red did not read. That is the check that forces the
// re-read whose absence caused all three of the smoke's round-2 gaps.
func TestValidateProposalRefusesAQuoteThatIsNotThere(t *testing.T) {
	err := ValidateProposal("merge mint", report, "text that was never in the report", "anything")
	if err == nil {
		t.Fatal("a proposal against text the report does not contain was accepted")
	}
	if !strings.Contains(err.Error(), "merge mint") {
		t.Errorf("the message must name the verb the seat actually typed: %v", err)
	}
}

// Red may not smuggle text past the anchor invariant either — a proposal that drops blue's
// anchor is a proposal blue could not legally apply.
func TestValidateProposalRefusesAnAnchorDrop(t *testing.T) {
	// Quoted ACROSS the anchor: matching skips annotations, so the located span CONTAINS
	// f-abc123 even though the quote does not mention it. 71% of anchored quotes in the smoke
	// had their anchor mid-span, so this is the common shape.
	if err := ValidateProposal("merge mint", report, "The sieve is fast and simple", "The sieve is quick and simple"); err == nil {
		t.Fatal("a proposal that drops a finding-marker was accepted")
	}
}

// An anchor CARRIED across is legal, for red exactly as for blue — the invariant is "which
// anchors exist", not "red may not touch one".
func TestValidateProposalAllowsAnchorTransit(t *testing.T) {
	if err := ValidateProposal("merge mint", report, "The sieve is fast and simple", "The sieve is quick<!--fx:f-abc123--> and simple"); err != nil {
		t.Fatalf("carrying an anchor across a proposal was refused: %v", err)
	}
}

// A no-op is not a proposal.
func TestValidateProposalRefusesAnIdenticalPair(t *testing.T) {
	if err := ValidateProposal("merge mint", report, "The sieve is fast", "The sieve is fast"); err == nil {
		t.Fatal("a proposal that changes nothing was accepted")
	}
}
