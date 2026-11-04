package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// THE AMBIGUOUS TELLS ARE ADVICE, THROUGH THE REAL VERB. "this report" has a reading as subject prose,
// so the mint lands and the note rides back on its confirmation, naming where the text prints.
func TestMintAdvisesOnThisRun(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe error term is bounded by the second derivative.\n")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	out, err := run(t, "mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
		"--key", "V1", "--class", "spec_underspecification", "--problem", "The bound in this report omits its derivation.",
		"--fix", "State the derivation the red team asked for.", "--check-kind", "document", "--check", "the derivation is stated",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	if err != nil {
		t.Fatalf("an ambiguous tell refused the mint — it is advice: %v\n%s", err, out)
	}
	if !strings.Contains(out, "minted G") {
		t.Errorf("the advisory replaced the mint's confirmation:\n%s", out)
	}
	for _, want := range []string{`"this report"`, `"red team"`, "risk matrix", "not a refusal"} {
		if !strings.Contains(out, want) {
			t.Errorf("the mint's advice does not carry %s:\n%s", want, out)
		}
	}

	clean, err := run(t, "mint", "--about-kind", "section", "--about", "Findings", "--run", runDir, "--seat-id", lensSeat,
		"--key", "V2", "--class", "spec_underspecification", "--problem", "The error term's constant is unsourced.",
		"--fix", "Cite the constant.", "--check-kind", "document", "--check", "the constant is cited",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	if err != nil {
		t.Fatalf("clean mint: %v", err)
	}
	if strings.Contains(clean, "NOTE") {
		t.Errorf("a clean mint carries an advisory:\n%s", clean)
	}
}

// fixNewReport is a report with one sentence red can prescribe a replacement for.
func fixNewReport(t *testing.T) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nFive independent verification approaches agree.\n")
	registerLensOnce(t, runDir)
	return runDir
}

func mintWithNew(t *testing.T, runDir, key, fixNew string) (string, error) {
	t.Helper()
	return run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", key, "--class", "spec_underspecification",
		"--quote", "Five independent verification approaches agree.", "--problem", "The approaches share a definition, so they are not independent.",
		"--fix", "Drop the independence claim.", "--check-kind", "document", "--check", "the section no longer claims independence",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low",
		"--new", fixNew)
}

// A --new PRESCRIPTION IS REPORT TEXT ONCE BLUE ACCEPTS IT, so an unambiguous tell in it is refused at the
// write, through the real verb, naming the flag and the term — and no gap is minted.
func TestMintRefusesAnUnambiguousTellInFixNewThroughTheRealVerb(t *testing.T) {
	runDir := fixNewReport(t)
	_, err := mintWithNew(t, runDir, "V1", "Five verification approaches agree, as blue-respond conceded.")
	if err == nil {
		t.Fatal("a prescription naming a seat id landed; blue's acceptance prints it in the report")
	}
	for _, want := range []string{"--new", `"blue-respond"`, "--reason"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s: %v", want, err)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
		t.Errorf("the refused prescription minted %d gap(s)", n)
	}
}

// The ambiguous tells in a --new prescription are advice: the gap is minted and the note rides back on
// its confirmation.
func TestMintAdvisesOnAnAmbiguousTellInFixNewThroughTheRealVerb(t *testing.T) {
	runDir := fixNewReport(t)
	out, err := mintWithNew(t, runDir, "V1", "Five verification approaches in this report agree.")
	if err != nil {
		t.Fatalf("an ambiguous tell refused the prescription — it is advice: %v\n%s", err, out)
	}
	for _, want := range []string{"minted G", `"this report"`, "not a refusal"} {
		if !strings.Contains(out, want) {
			t.Errorf("the mint's confirmation does not carry %s:\n%s", want, out)
		}
	}
}
