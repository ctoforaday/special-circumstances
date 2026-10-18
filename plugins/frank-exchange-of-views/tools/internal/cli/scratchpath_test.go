package cli

import (
	"strings"
	"testing"
)

// scratchFile is a file in the run's scratch/, as universe-m29's refuted fate named one in its method.
const scratchFile = "scratch/blue-respond-verify-91.py"

// A FILE IN THE RUN'S SCRATCH IS REFUSED IN EVERY FIELD OF RED'S THAT THE REPORT PRINTS: the risk
// matrix's problem and fix, a --new prescription blue may accept, and a labelled corroboration's title.
// Each refusal names the flag, the path and what to write instead, and nothing is recorded.
func TestRedsReportTextRefusesAScratchPath(t *testing.T) {
	const sentence = "The residues were recomputed in " + scratchFile + "."
	mint := func(runDir string, problem, fix, fixNew string) (string, error) {
		return run(t, "mint", "--run", runDir, "--seat-id", lensSeat,
			"--key", "S1", "--class", "overclaim",
			"--quote", "Five independent verification approaches agree.", "--problem", problem,
			"--fix", fix, "--check-kind", "document", "--check", "the section no longer claims independence",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium", "--complexity", "low",
			"--new", fixNew)
	}
	const problem, fix, fixNew = "The approaches share a definition.", "Drop the independence claim.", "Five verification approaches agree."
	for _, c := range []struct {
		flag string
		call func() (string, error)
	}{
		{"--problem", func() (string, error) { return mint(fixNewReport(t), sentence, fix, fixNew) }},
		{"--fix", func() (string, error) { return mint(fixNewReport(t), problem, sentence, fixNew) }},
		{"--new", func() (string, error) { return mint(fixNewReport(t), problem, fix, sentence) }},
		{"--title", func() (string, error) {
			return run(t, "corroborate", "--run", corroborateRun(t), "--seat-id", "red-lens-evidence",
				"--url", "https://example.org/residues", "--title", "Residue tables, "+scratchFile,
				"--quote", corroborated, "--as", "supports", "--confidence", "high", "--reason", "the tables state it")
		}},
	} {
		out, err := c.call()
		if err == nil {
			t.Errorf("%s carrying %s was recorded:\n%s", c.flag, scratchFile, out)
			continue
		}
		for _, want := range []string{c.flag, scratchFile, "never shipped", "recorded proof"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: the refusal does not name %q: %v", c.flag, want, err)
			}
		}
	}
}

// BLUE'S REPORT TEXT IS ADVISED, NEVER REFUSED: every field of an inquiry's row names the path in its
// advice and the act is recorded — the question, the hypothesis, and a fate's method and reason.
func TestBluesInquiryTextNamesAScratchPathInItsAdvice(t *testing.T) {
	const sentence = "the residues recomputed in " + scratchFile
	for _, c := range []struct{ name, question, hypothesis, reason, method string }{
		{"question", sentence, "91 has a factor below 10", "7 divides 91", "trial division"},
		{"hypothesis", "trial division up to 9", sentence, "7 divides 91", "trial division"},
		{"reason", "trial division up to 9", "91 has a factor below 10", sentence, "trial division"},
		{"method", "trial division up to 9", "91 has a factor below 10", "7 divides 91", sentence},
	} {
		runDir := newRun(t)
		registerChairOnce(t, runDir)
		seat := inquirySeat(t, runDir)
		propose, err := run(t, "inquiry", "propose", "--run", runDir, "--seat-id", seat,
			"--question", c.question, "--hypothesis", c.hypothesis)
		if err != nil {
			t.Fatalf("%s: propose: %v", c.name, err)
		}
		record, err := runAt(t, "inquiry", "record", "--run", runDir, "--seat-id", seat,
			"--id", "Q1", "--as", "refuted", "--reason", c.reason, "--method", c.method)
		if err != nil {
			t.Fatalf("%s: a scratch path in blue's text was refused — blue is advised: %v", c.name, err)
		}
		if out := propose + record; !strings.Contains(out, `"`+scratchFile+`" reads as apparatus`) {
			t.Errorf("%s: the advice does not name the path:\n%s", c.name, out)
		}
	}
}
