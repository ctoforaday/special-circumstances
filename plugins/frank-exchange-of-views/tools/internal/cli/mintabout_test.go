package cli

import (
	"strings"
	"testing"
)

// A GAP ABOUT AN ABSENCE ANCHORS THE WAY THE FINDING DID, with the same two flags.
//
// `lens mint` is the next step of the act `lens finding` begins, and its help used to send an
// omission back to the handle the finding had just been taught to stop using: "for a gap about
// something MISSING, quote the sentence where it SHOULD be — that is how a lens finding anchors an
// omission". Once that stopped being how a finding anchors an omission, the sentence was a
// cross-verb instruction to do the wrong thing.
func TestAGapAboutAnAbsenceNeedsNoBorrowedQuote(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	registerLensOnce(t, runDir)
	if _, err := runAt(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "G1", "--class", "spec_underspecification", "--about-kind", "section", "--about", "Risk matrix",
		"--problem", "the template names a graded risk matrix and the report has no such section",
		"--check-kind", "document", "--check", "the report carries a graded risk matrix",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a gap about a MISSING section was refused: %v", err)
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "board")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Risk matrix") {
		t.Errorf("the board does not carry what the gap is about:\n%s", out)
	}
}

// ONE SUBJECT. A gap that claims a quote AND an about is claiming two, and is refused.
func TestAGapCannotClaimBothAnchors(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n")
	registerLensOnce(t, runDir)
	_, err := runAt(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "G2", "--class", "spec_underspecification",
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
	registerLensOnce(t, runDir)
	_, err := runAt(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "G3", "--class", "spec_underspecification", "--about-kind", "inquiry", "--about", "Q99",
		"--problem", "p", "--check-kind", "document", "--check", "c",
		"--severity", "low", "--likelihood", "low", "--impact", "low")
	if err == nil {
		t.Fatal("a gap anchored to an inquiry id nothing proposed was accepted")
	}
}

// AN INQUIRY THAT CONTRADICTS WHAT THE READER IS TOLD IS ANSWERED THROUGH A GAP.
//
// The inquiries print in the report beside the body, so an approach printed as declined that a
// section carries is a contradiction on the reader's page. The gap anchored to the inquiry is on
// blue's list with what it is about.
func TestAGapAboutAnInquiryReachesBlue(t *testing.T) {
	runDir := seatRun(t)
	if _, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", "blue-respond", "--question", "test primality by trial division"); err != nil {
		t.Fatal(err)
	}
	inquiry := ""
	for _, e := range events(t, runDir) {
		if a := e.GetInquiry(); a != nil {
			inquiry = a.GetInquiryId()
		}
	}
	const lens = "red-lens-logic"
	registerSeatOnce(t, runDir, lens)
	if _, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", "blue-respond", "--id", inquiry, "--as", "confirmed", "--reason", "trial division found 7"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "mint", "--run", runDir, "--seat-id", lens, "--key", "G1", "--class", "spec_underspecification", "--about-kind", "inquiry", "--about", inquiry,
		"--problem", "The trial-division approach is listed as never run and a section reports its result.", "--check-kind", "document", "--check", "the inquiry's fate and the section agree",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a gap anchored to an inquiry was refused: %v", err)
	}
	w, err := run(t, "show", "--run", runDir, "--seat-id", "blue-respond", "--json", "work")
	if err != nil {
		t.Fatalf("blue's work list: %v", err)
	}
	if !strings.Contains(w, `"about_kind": "inquiry"`) || !strings.Contains(w, `"about_ref": "`+inquiry+`"`) {
		t.Errorf("a gap anchored to an inquiry is not on blue's work list with what it is about:\n%s", w)
	}
}
