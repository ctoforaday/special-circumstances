package cli

import (
	"regexp"
	"strings"
	"testing"
)

// THE AMBIGUOUS TELLS ARE ADVICE, THROUGH THE REAL VERB. "this report" has a reading as subject prose,
// so the mint lands and the note rides back on its confirmation, naming where the text prints.
func TestMintAdvisesOnThisRun(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe error term is bounded by the second derivative.\n")
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	out, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "V1", "--class", "overclaim", "--problem", "The bound in this report omits its derivation.",
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

	clean, err := run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
		"--key", "V2", "--class", "overclaim", "--problem", "The error term's constant is unsourced.",
		"--fix", "Cite the constant.", "--check-kind", "document", "--check", "the constant is cited",
		"--severity", "medium", "--likelihood", "medium", "--impact", "medium")
	if err != nil {
		t.Fatalf("clean mint: %v", err)
	}
	if strings.Contains(clean, "NOTE") {
		t.Errorf("a clean mint carries an advisory:\n%s", clean)
	}
}

// A LABELLED CORROBORATION PUTS ONE SPAN OF RED'S TEXT INTO report.md: its title, in the Bibliography
// entry "[^N]: <title>. <url> (accessed <date>)". The url and the date are the source's identity, and
// the claim is blue's sentence. Red's reason, outcome, confidence and seat stay on the record — so the
// entry is exactly those three fields, and nothing else red wrote reaches the reader.
func TestCorroborationCarriesOnlyItsSourceIntoTheReport(t *testing.T) {
	runDir := corroborateRun(t)
	const (
		title  = "Po-Shen Loh, A Simple Proof of the Quadratic Formula"
		url    = "https://example.org/loh-2019"
		reason = "REDREASON the source states the method at section 2"
	)
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", url, "--title", title, "--access-date", "2026-09-01",
		"--quote", corroborated, "--as", "supports_with_bridge", "--confidence", "medium",
		"--reason", reason); err != nil {
		t.Fatalf("corroborate: %v", err)
	}
	out := assembled(t, runDir)
	bib := out[strings.Index(out, "## Bibliography"):]
	entries := regexp.MustCompile(`(?m)^\[\^\d+\]:.*$`).FindAllString(bib, -1)
	want := "[^1]: " + title + ". " + url + " (accessed 2026-09-01)"
	if len(entries) != 1 || entries[0] != want {
		t.Fatalf("the Bibliography entry is not exactly title, url and date:\n got %q\nwant %q", entries, want)
	}
	for _, leak := range []string{"REDREASON", "red-lens-evidence", "supports_with_bridge", "corroborat"} {
		if strings.Contains(out, leak) {
			t.Errorf("report.md carries %q from the corroboration; only its source may reach the reader:\n%s", leak, out)
		}
	}
}

// THE TITLE IS HELD TO THE REPORT'S VOICE, THROUGH THE REAL VERB: a voiced title lands (the refusal
// reaches the mint's problem and fix only) and the advice names the Bibliography entry; a clean title
// says nothing extra.
func TestCorroborateAdvisesOnItsTitleThroughTheRealVerb(t *testing.T) {
	runDir := corroborateRun(t)
	out, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/voiced", "--title", "Standard found by red-lens-evidence for this run",
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the standard states it")
	if err != nil {
		t.Fatalf("a voiced title refused the corroboration — it is advice: %v\n%s", err, out)
	}
	for _, want := range []string{`"red-lens-evidence"`, `"this run"`, "Bibliography entry", "not a refusal"} {
		if !strings.Contains(out, want) {
			t.Errorf("the corroboration's advice does not carry %s:\n%s", want, out)
		}
	}

	clean, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/clean", "--title", "ISO 80000-2:2019, Mathematical signs and symbols",
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the standard states it")
	if err != nil {
		t.Fatalf("clean corroboration: %v", err)
	}
	if strings.Contains(clean, "NOTE") {
		t.Errorf("a clean title carries an advisory:\n%s", clean)
	}
}
