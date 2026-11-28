package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
)

// A GAP ABOUT AN ABSENCE ANCHORS BY WHERE THE MISSING THING BELONGS, with --about-kind and --about.
//
// It borrows no quote: telling a lens to quote the sentence where something MISSING should be is an
// instruction to anchor an omission on text that does not hold it. A SECTION GAP STANDS ON ITS
// HEADING (T-B52): the reference keys one heading of the report, and the gap's token lands at that
// heading's end; a reference keying no heading, or two, or a section the tool composes is refused.
func TestASectionGapStandsOnItsHeading(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nSeven is prime.\n\n## Method\n\nWe read.\n\n## Method\n\nAgain.\n")
	registerLensOnce(t, runDir)
	mint := func(key, about string) (string, error) {
		return runAt(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", key, "--class", "spec_underspecification", "--about-kind", "section", "--about", about,
			"--problem", "the section states no limits of the primality claim",
			"--check-kind", "document", "--check", "the section states the claim's limits",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	}
	if _, err := mint("G1", "h"); err != nil {
		t.Fatalf("a gap about a section the report holds was refused: %v", err)
	}
	id := handleText(t, runDir, "G1")
	if md := readReport(t, runDir); !strings.HasPrefix(md, "# H"+anchor.Token(id)+"\n") {
		t.Errorf("the section gap's token is not at its heading's end:\n%s", md)
	}
	for _, c := range []struct{ about, want string }{
		{"Results", "keys no heading of the report"},
		{"Method", "keys two headings of the report"},
		{"Risk matrix", "the tool composes"},
	} {
		if _, err := mint("bad-"+c.about, c.about); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("a gap about section %q: want a refusal saying %q, got %v", c.about, c.want, err)
		}
	}
	out, err := run(t, "show", "--run", runDir, "--seat-id", "red-chair", "board")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "the section states no limits") {
		t.Errorf("the board does not carry the section gap:\n%s", out)
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
