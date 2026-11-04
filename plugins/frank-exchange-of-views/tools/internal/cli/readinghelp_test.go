package cli

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// 30. THE READING VERB DESCRIBES ITS GAP FLAGS BY ITS SUBJECT: what --new replaces is the sentence
// the anchor's token sits in or the quote, --class offers the classes a reading admits and no other,
// --reason says the text is the gap's argument, and the seven gap flags are REQUIRED unless the
// reading mints nothing. mint's own flag text is unchanged.
func TestTheReadingVerbsDescribeTheirGapFlagsByTheirSubject(t *testing.T) {
	v := cmdAt(NewRootFor(lensSeat), []string{"verify"})
	page := seat.HelpText(v)
	if strings.Contains(page, "unclassified") {
		t.Error("verify's page offers unclassified")
	}
	for _, f := range []string{"anchor", "url"} {
		if u := v.Flags().Lookup(f).Usage; !strings.Contains(u, "Give exactly one of --anchor") {
			t.Errorf("--%s does not say verify takes exactly one of --anchor and --url: %s", f, u)
		}
	}
	if u := v.Flags().Lookup("new").Usage; !strings.Contains(u, "the sentence your --anchor's token sits in, or your --quote") {
		t.Errorf("--new does not name the reading's subject: %s", u)
	}
	// --class lists exactly the four classes the two facets admit.
	want := []string{"citation_figure_misattribution", "within_source_condition_misattribution", "unsourced_claim", "claim_contradicts_source"}
	_, block, _ := strings.Cut(page, "\n  --class\n")
	block, _, _ = strings.Cut(block, "\n\n")
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^    ([a-z_]+) `).FindAllStringSubmatch(block, -1) {
		got = append(got, m[1])
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("--class lists %v, want %v", got, want)
	}
	reason := v.Flags().Lookup("reason").Usage
	if !strings.HasPrefix(reason, "REQUIRED — "+seat.ReasonIs["verify"][:40]) || !strings.Contains(reason, "that gap's argument") {
		t.Errorf("--reason does not open with its ReasonIs text and name the gap's argument: %s", reason)
	}
	for _, f := range []string{"class", "problem", "check", "check-kind", "severity", "likelihood", "impact"} {
		if u := v.Flags().Lookup(f).Usage; !strings.HasPrefix(u, "REQUIRED unless") {
			t.Errorf("--%s does not open \"REQUIRED unless\": %s", f, u)
		}
	}
	// A class outside the two facets is refused at the parse, and nothing is recorded.
	b := readingRun(t)
	_, err := b.verify(lensL, citeReading(b.c1, "refutes", "spec_underspecification")...)
	mustRefuse(t, err, "", "spec_underspecification")
	if c := b.counts(); c != [3]int{} {
		t.Errorf("a refused class wrote verify/mint/anchor %v", c)
	}

	// mint's flag text is the lens manual's, byte for byte: the generated evidence-lens definition
	// inlines its seat's manual, mint's page included (the difftest manual golden walks the bench's,
	// blue's, a lane's and the operator's surfaces, which hold no mint). The manual prints each flag
	// as pflag renders it, and a text another page already printed as a pointer to its SHARED section.
	path, err := repotree.Plugin("agents", "red-lens-evidence.md")
	if err != nil {
		t.Fatal(err)
	}
	agent, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, mintPage, _ := strings.Cut(string(agent), "$ feov-record mint --help\n")
	mintPage, _, _ = strings.Cut(mintPage, "$ feov-record ")
	lines := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^      --([a-z-]+) (?:[^ ]+ )? *(.*)$`).FindAllStringSubmatch(mintPage, -1) {
		lines[m[1]] = strings.TrimSpace(m[2])
	}
	cmdAt(NewRootFor(lensSeat), []string{"mint"}).Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" || f.Hidden {
			return
		}
		_, usage := pflag.UnquoteUsage(f)
		got, ok := lines[f.Name]
		if shared, _, cut := strings.Cut(got, "→ SHARED"); cut {
			ok = ok && strings.HasPrefix(usage, strings.TrimSpace(shared))
		} else {
			ok = ok && got == usage
		}
		if !ok {
			t.Errorf("mint --%s's text is not the lens manual's:\n  now:    %s\n  manual: %s", f.Name, usage, got)
		}
	})
	_, err = run(t, "corroborate", "--run", b.runDir, "--seat-id", lensL)
	mustRefuse(t, err, "", `no command named "corroborate" exists`)
}
