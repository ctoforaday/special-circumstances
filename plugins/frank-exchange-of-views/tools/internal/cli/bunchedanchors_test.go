package cli

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// A RUN CARRIED TO THE END OF A LONGER REPLACEMENT IS REFUSED WITH WHAT EACH ANCHOR MARKS, through
// the verbs: one sentence holding an anchor of every kind is split in two with the whole run typed
// after the second. The refusal prints, beside each token, the text the record holds about it — the
// finding's argument, the citation's title and argument, the computation's argument, the gap's problem —
// and the replacement's sentences, numbered; it never prints the sentence the run stood on, which
// is every anchor's stored location and tells them apart by nothing. Nothing reaches the record,
// and the same split with the anchors written on the sentences they belong to applies.
func TestABunchedRunIsRefusedWithWhatEachAnchorMarks(t *testing.T) {
	runDir := acceptRun(t)
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", blueSeat, "--quote", acceptS+".",
		"--url", "https://costs/1", "--title", "Quarterly cost tables", "--reason", "table two gives the first quarter's rise"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "finding", "--run", runDir, "--seat-id", lensSeat, "--key", "F1", "--quote", acceptS+".",
		"--reason", "the word sharply has\nno figure behind it", "--severity", "low", "--likelihood", "low", "--impact", "low"); err != nil {
		t.Fatal(err)
	}
	gap, err := mintQuote(t, runDir, "G", acceptS+".", "--problem", "the quarter is not dated")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "compute", "--run", runDir, "--seat-id", blueSeat, "--quote", acceptS+".",
		"--script", script(t, runDir, "rise.js", "console.log(3-1)"), "--reason", "the difference of the two quarters is computed"); err != nil {
		t.Fatal(err)
	}
	before := readReport(t, runDir)
	m := regexp.MustCompile(regexp.QuoteMeta(acceptS) + `((?:<!--[a-z]+:[A-Z]-[0-9a-f]{8}-->){4})\.`).FindStringSubmatch(before)
	if m == nil {
		t.Fatalf("the fixture does not hold four anchors at the sentence's end:\n%s", before)
	}
	toks := regexp.MustCompile(`<!--[a-z]+:[A-Z]-[0-9a-f]{8}-->`).FindAllString(m[1], -1)
	const one, two = "Costs rose in the first quarter", "The rise was sharp"

	err = acceptEdit(t, runDir, "E", gap, "--quote", acceptS+m[1]+".", "--new", one+". "+two+m[1]+".")
	if err == nil {
		t.Fatalf("the split applied with the whole run after its last sentence:\n%s", readReport(t, runDir))
	}
	for _, want := range append(toks,
		"the word sharply has no figure behind it",
		"Quarterly cost tables — table two gives the first quarter's rise",
		"the difference of the two quarters is computed",
		"the quarter is not dated",
		"holds 2 sentences where the text it replaces holds 1",
		"  1. "+one+".\n", "  2. "+two+".\n",
		"Write each anchor after the sentence that now holds its claim",
	) {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lacks %q:\n%v", want, err)
		}
	}
	if strings.Contains(err.Error(), acceptS) {
		t.Errorf("the refusal prints the sentence the run stood on, every anchor's stored location:\n%v", err)
	}
	if got := readReport(t, runDir); got != before {
		t.Fatalf("the refused edit reached the record:\n%s", got)
	}

	// THE SAME SPLIT, PLACED: two anchors on each sentence.
	placed := fmt.Sprintf("%s%s%s. %s%s%s.", one, toks[0], toks[1], two, toks[2], toks[3])
	if err := acceptEdit(t, runDir, "E", gap, "--quote", acceptS+m[1]+".", "--new", placed); err != nil {
		t.Fatalf("the split with its anchors placed on both sentences was refused: %v", err)
	}
	if got := readReport(t, runDir); !strings.Contains(got, placed) {
		t.Errorf("the placed split is not in the report:\n%s", got)
	}
}
