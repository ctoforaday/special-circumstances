package cli

import (
	"regexp"
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

// A LABELLED CORROBORATION PUTS ONE SPAN OF RED'S TEXT INTO report.md: its title, in the source's note
// "[^N]: <title>. <url> (accessed <date>)" and, with no blue cite of the URL, its Bibliography line
// "- <title>. <url> (accessed <date>)". The url and the date are the source's identity, and the claim
// is blue's sentence. Red's reason, outcome, confidence and seat stay on the record — so the note and
// the line are exactly those three fields, and nothing else red wrote reaches the reader.
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
	entry := title + ". " + url + " (accessed 2026-09-01)"
	notes := regexp.MustCompile(`(?m)^\[\^\d+\]:.*$`).FindAllString(out, -1)
	if len(notes) != 1 || notes[0] != "[^1]: "+entry {
		t.Fatalf("the source's note is not exactly title, url and date:\n got %q\nwant %q", notes, "[^1]: "+entry)
	}
	i := strings.Index(out, "## Bibliography")
	if i < 0 {
		t.Fatalf("no Bibliography:\n%s", out)
	}
	lines := regexp.MustCompile(`(?m)^- .*$`).FindAllString(out[i:], -1)
	if len(lines) != 1 || lines[0] != "- "+entry {
		t.Fatalf("the Bibliography line is not exactly title, url and date:\n got %q\nwant %q", lines, "- "+entry)
	}
	for _, leak := range []string{"REDREASON", "red-lens-evidence", "supports_with_bridge", "corroborat"} {
		if strings.Contains(out, leak) {
			t.Errorf("report.md carries %q from the corroboration; only its source may reach the reader:\n%s", leak, out)
		}
	}
}

// THE AMBIGUOUS TELLS IN A TITLE ARE ADVICE, THROUGH THE REAL VERB: "this run" has a reading as part of a
// source's own title, so the corroboration lands and the advice names the note and Bibliography entry; a clean
// title says nothing extra.
func TestCorroborateAdvisesOnItsTitleThroughTheRealVerb(t *testing.T) {
	runDir := corroborateRun(t)
	out, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/voiced", "--title", "Standard as read for this run",
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the standard states it")
	if err != nil {
		t.Fatalf("an ambiguous tell refused the corroboration — it is advice: %v\n%s", err, out)
	}
	for _, want := range []string{`"this run"`, "Bibliography entry", "not a refusal"} {
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

// AN UNAMBIGUOUS TELL IN A TITLE THAT PRINTS IS REFUSED, THROUGH THE REAL VERB. A seat id has no reading as
// a source's title, and a supporting corroboration's title prints in the source's note and Bibliography entry: the refusal names
// the flag and the term, and nothing is recorded. The same title on a refutation lands — a refutation is
// no footnote, so its title stays on the record.
func TestCorroborateRefusesAnUnambiguousTellInItsTitleThroughTheRealVerb(t *testing.T) {
	runDir := corroborateRun(t)
	const voiced = "Standard found by red-lens-evidence"
	out, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/voiced", "--title", voiced,
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the standard states it")
	if err == nil {
		t.Fatalf("a title naming a seat id landed in the report:\n%s", out)
	}
	for _, want := range []string{"--title", `"red-lens-evidence"`, "Bibliography", "--reason"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %s: %v", want, err)
		}
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_VERIFY); n != 0 {
		t.Errorf("the refused corroboration recorded %d event(s)", n)
	}

	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/voiced", "--title", voiced,
		"--quote", corroborated, "--as", "refutes", "--confidence", "high",
		"--reason", "the standard says otherwise"); err != nil {
		t.Errorf("a refutation's title was refused; it prints nowhere in the report: %v", err)
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
		"--key", key, "--class", "overclaim",
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
