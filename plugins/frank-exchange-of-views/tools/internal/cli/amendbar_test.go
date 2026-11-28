package cli

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// T-B53 (--fix-new, the bar): AN AMENDED PRESCRIPTION OVER TEXT THE LENS'S OWN APPLIED PRESCRIPTION
// WROTE IS REFUSED BY THE ESTOPPEL BAR, as a mint's is. The lens holds two gaps on one sentence: G1
// prescribing a sentence that keeps G2's quoted words, and G2. Blue's accept of G1 applies it, so
// G2's location survives, now inside text the lens's own prescription wrote; the lens's amend of G2
// with a --fix-new over it is refused naming the estoppel, and records nothing. The same amend
// restating G2's problem alone lands: the bar is on the prescription.
func TestAnAmendedPrescriptionOverTheLenssAppliedTextIsBarred(t *testing.T) {
	runDir := acceptRun(t)
	g2, err := mintQuote(t, runDir, "B", "Costs rose sharply", "--class", "unsourced_claim")
	if err != nil {
		t.Fatalf("G2: %v", err)
	}
	g1, err := mintQuote(t, runDir, "P", acceptS, "--new", acceptS+" in every region")
	if err != nil {
		t.Fatalf("G1: %v", err)
	}
	if err := acceptEdit(t, runDir, "A", g1, "--accept"); err != nil {
		t.Fatalf("blue's accept of %s: %v", g1, err)
	}
	asSeat(t, runDir, lensSeat, "register")
	amends := countType(t, runDir, recordpb.EventType_EVENT_TYPE_AMEND)
	out, err := amendGap(t, runDir, lensSeat, g2, "--fix-new", "Costs rose", "--reason", "a narrower prescription")
	heldRefusal(t, "an amended prescription over the lens's applied text", out, err, "estoppel")
	if got := countType(t, runDir, recordpb.EventType_EVENT_TYPE_AMEND); got != amends {
		t.Errorf("amends on the record %d, want %d", got, amends)
	}
	if out, err := amendGap(t, runDir, lensSeat, g2, "--problem", "the claim names no source", "--reason", "restated"); err != nil {
		t.Errorf("the amend of %s's problem alone: %v %s", g2, err, out)
	}
}
