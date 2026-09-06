package blue

import (
	"strings"
	"testing"
)

func TestVerifyReproductionPassesOnAMatch(t *testing.T) {
	if err := VerifyReproduction("same bytes", "same bytes"); err != nil {
		t.Errorf("identical strings should verify, got: %v", err)
	}
}

// The mismatch error is read by a model that ACTS on text. It must stop blue, not steer it into a
// repair — so it must carry the safety directives and must NOT read as an edit instruction.
func TestVerifyReproductionErrorTellsBlueToStopNotEdit(t *testing.T) {
	err := VerifyReproduction(
		"The value is stable and holds across the range.",
		"The value is steady and holds across the range.", // tool "lost" a word — a render defect
	)
	if err == nil {
		t.Fatal("a mismatch must error")
	}
	msg := err.Error()

	// MUST carry: stop, file safe, do-not-edit, escalate-as-friction, byte offset.
	for _, want := range []string{"STOP", "PRESERVED", "Do not edit", "no diff for you to apply", "friction", "byte "} {
		if !strings.Contains(msg, want) {
			t.Errorf("mismatch error is missing the safety directive %q:\n%s", want, msg)
		}
	}
	// The divergence window must be labelled as diagnostic, not a target.
	if !strings.Contains(msg, "NOT an instruction") || !strings.Contains(msg, "FOR THE MAINTAINER") {
		t.Errorf("the divergence is not framed as maintainer-only diagnostic:\n%s", msg)
	}
	// MUST NOT read as an actionable edit: no old→new / change-X-to-Y framing a model would splice.
	for _, forbidden := range []string{"change to", "replace with", "should say", "should be", "--old", "--new", "→"} {
		if strings.Contains(msg, forbidden) {
			t.Errorf("mismatch error contains edit-instruction language %q that blue could apply blindly:\n%s", forbidden, msg)
		}
	}
	// The directive must come BEFORE any shown text, so a model reading only the top still stops.
	if iStop, iDiag := strings.Index(msg, "STOP"), strings.Index(msg, "as you wrote it"); iStop < 0 || iDiag < 0 || iStop > iDiag {
		t.Errorf("the STOP directive must precede the divergence window")
	}
}

// A prefix relationship (render truncated the tail) still diverges and still reports where.
func TestVerifyReproductionCatchesTruncation(t *testing.T) {
	err := VerifyReproduction("full report text with a tail", "full report text")
	if err == nil || !strings.Contains(err.Error(), "byte 16") {
		t.Errorf("truncation should be caught and located at byte 16, got: %v", err)
	}
}
