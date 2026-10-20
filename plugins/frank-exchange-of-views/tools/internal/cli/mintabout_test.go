package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
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

// ONE SUBJECT, ON BOTH VERBS. A gap that claims a quote AND an about is claiming two, which is the
// same refusal `lens finding` gives — the point of sharing the flags is that they behave the same.
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

// THE TWO VERBS OFFER ONE VOCABULARY, and this is the assertion that keeps them offering it: a gap
// takes every kind a finding takes that the kind's `gap_may_name` facet admits, and nothing else.
//
// The flags could stay spelled alike while the closed sets behind them drifted — a value added for
// findings and not for gaps would leave a seat able to anchor a finding somewhere it cannot then
// anchor the gap, which is the half-state that reads as done. The one value a gap refuses is a
// contradiction: the finding raises it, and the gap is minted from that finding. The sets are
// compared rather than re-listed, so a value added to one and not the other fails here.
func TestAFindingAndAGapAnchorToTheSameThings(t *testing.T) {
	finding, gap := record.MustEnum("finding", "about_kind"), record.MustEnum("mint", "about_kind")
	words := func(e record.EnumField) []string {
		var out []string
		for _, v := range e.Values {
			out = append(out, v.Name)
		}
		return out
	}
	var admitted []string
	for _, w := range words(finding) {
		if w != recordpb.Word(recordpb.AboutKind_ABOUT_KIND_CONTRADICTION) {
			admitted = append(admitted, w)
		}
	}
	f, g := strings.Join(admitted, ","), strings.Join(words(gap), ",")
	if f != g || len(words(finding)) != len(admitted)+1 {
		t.Errorf("`lens finding` and `lens mint` no longer anchor to the same things:\n  finding: %s\n  mint:    %s", f, g)
	}
	if finding.Flag != gap.Flag {
		t.Errorf("one anchor, two flag names: finding uses %q and mint uses %q", finding.Flag, gap.Flag)
	}
}

// AN INQUIRY THAT CONTRADICTS WHAT THE READER IS TOLD IS ANSWERED THROUGH A GAP, AND ONLY A GAP.
//
// The inquiries print in the report beside the body, so an approach printed as declined that a
// section carries is a contradiction on the reader's page. A finding about an inquiry is on no
// seat's work list, and blue's responding seat answers its list. The gap anchored to the same
// inquiry is on blue's list with what it is about, and `finding`'s own page says which of the two
// reaches anybody.
func TestAGapAboutAnInquiryReachesBlueAndAFindingAboutOneReachesNoWorkList(t *testing.T) {
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
	const objection = "a section carries this approach and the inquiry says it was never run"
	if _, err := runAt(t, "finding", "--run", runDir, "--seat-id", lens, "--key", "F1", "--about-kind", "inquiry", "--about", inquiry,
		"--reason", objection, "--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a finding about an inquiry was refused: %v", err)
	}
	work := func(seat string) string {
		out, err := run(t, "show", "--run", runDir, "--seat-id", seat, "--json", "work")
		if err != nil {
			t.Fatalf("%s's work list: %v", seat, err)
		}
		return out
	}
	seats := []string{"blue-respond", "red-chair", lens, lensSeat, "judge"}
	for _, seat := range seats {
		if w := work(seat); strings.Contains(w, objection) || strings.Contains(w, `"about_ref": "`+inquiry+`"`) {
			t.Errorf("a finding about an inquiry is on %s's work list; `finding --help` says it reaches none:\n%s", seat, w)
		}
	}
	if h := help(t, "finding", "--help", "--seat-id", lens); !strings.Contains(h, "A finding about an inquiry reaches no seat's work list; a gap anchors to an inquiry the same way, and a material one is on blue's") {
		t.Errorf("`finding --help` does not say that a finding about an inquiry reaches no work list and that the gap is what does:\n%s", h)
	}
	if _, err := runAt(t, "mint", "--run", runDir, "--seat-id", lens, "--key", "G1", "--class", "spec_underspecification", "--about-kind", "inquiry", "--about", inquiry,
		"--problem", "The trial-division approach is listed as never run and a section reports its result.", "--check-kind", "document", "--check", "the inquiry's fate and the section agree",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium"); err != nil {
		t.Fatalf("a gap anchored to an inquiry was refused: %v", err)
	}
	if w := work("blue-respond"); !strings.Contains(w, `"about_kind": "inquiry"`) || !strings.Contains(w, `"about_ref": "`+inquiry+`"`) {
		t.Errorf("a gap anchored to an inquiry is not on blue's work list with what it is about:\n%s", w)
	}
}
