package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/capture"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// A SEAT'S WORDS REACH THE SET AS TEXT, THROUGH THE REAL ASSEMBLY. A lens files a finding that
// names the report's own proof footnote — a label report.md defines and the board document does
// not — and logs a regex whose character class is the two characters that open a footnote. The
// set that assembles carries both as typed and escaped, and capture's audit of it passes: the one
// reference in the set is the weave's, in the document that defines it.
func TestASeatsWordsOpenNoFootnoteInTheSet(t *testing.T) {
	runDir := actFixture(t)
	const sentence = "the parser accepts an empty body in this line."
	script := filepath.Join(t.TempDir(), "blue-respond-empty.sh")
	if err := os.WriteFile(script, []byte("echo 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "prove", "--run", runDir, "--seat-id", "blue-respond", "--quote", sentence,
		"--script", script, "--reason", "an empty body parses to nothing, which the class [^\\s] never matches"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "avenue", "propose", "--run", runDir, "--seat-id", "blue-respond",
		"--reason", "whether a tokenizer built on [^,]+ splits an empty field", "--hypothesis", "it does not"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "finding", "--run", runDir, "--seat-id", "red-lens-evidence", "--quote", sentence,
		"--reason", "the one footnote [^P1] cites types its constants instead of computing them",
		"--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	if _, err := runAt(t, "log", "--run", runDir, "--seat-id", "red-lens-evidence", "--type", "friction",
		"--reason", "grep with bounded repetition, the form grep -o '[^<>]{0,120}term[^<>]{0,120}', fails on the cached page"); err != nil {
		t.Fatal(err)
	}

	docs, err := report.AssembleAll(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	set := t.TempDir()
	body := map[string]string{}
	for _, d := range docs {
		body[d.File] = d.Body
		if err := os.WriteFile(filepath.Join(set, d.File), []byte(d.Body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(body[report.FileReport], "[^P1]") || !strings.Contains(body[report.FileReport], "\n[^P1]: ") {
		t.Fatalf("the fixture's report does not carry the weave's own proof note and its definition:\n%s", body[report.FileReport])
	}
	if d := body[report.FileDocket]; !strings.Contains(d, `the one footnote \[^P1] cites`) {
		t.Errorf("the board document does not carry the finding's words escaped:\n%s", d)
	}
	if r := body[report.FileRun]; !strings.Contains(r, `grep -o '\[^<>]{0,120}term\[^<>]{0,120}'`) {
		t.Errorf("the run document does not carry the log entry's words escaped:\n%s", r)
	}
	// THE THREE WRITERS OUTSIDE THE TABLE: the note the weave defines, which carries blue's words
	// after the tool's label; the evidence document's heading for the same proof; and the avenues
	// document.
	if rep := body[report.FileReport]; !strings.Contains(rep, `which the class \[^\s] never matches`) {
		t.Errorf("the proof note's definition carries blue's words as a second reference:\n%s", rep)
	}
	if ev := body[report.FileEvidence]; !strings.Contains(ev, `\[^\s]`) || strings.Contains(strings.ReplaceAll(ev, `\[^`, ""), "[^\\s]") {
		t.Errorf("the evidence document carries the proof's note unescaped:\n%.600s", ev)
	}
	if av := body[report.FileAvenues]; !strings.Contains(av, `built on \[^,]+ splits`) {
		t.Errorf("the avenues document carries the avenue's words unescaped:\n%s", av)
	}
	if a := capture.FootnoteIntegrity(runtest.Open(t, set)); a.Verdict != "PASS" {
		t.Errorf("the set a finding and a log entry were printed into fails the footnote audit: %s", a.Detail)
	}
}
