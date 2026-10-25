package surface

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// A CONTRADICTION HAS ONE ANSWERING GAP, and every lens is told which.
//
// The record answers a contradiction with the gap minted from the finding that raised it, and with
// no other gap (internal/record/contradictionref_test.go). The lens definitions carry that model in
// prose authored once, at scripts/agentgen/src/fragments/lens-audit-and-mint.md, and a sentence
// there that offers the lens a second answering gap sends it after an act the tool does not have:
// the verb gate cannot see it, because the sentence names an act, not a backticked verb. What this
// holds is that every lens definition states the one answer, and that the tool's own `finding`
// page states the same one.
func TestEveryLensIsToldTheOneGapThatAnswersAContradiction(t *testing.T) {
	const lensSays = "the finding naming it raises it, and only the gap you mint from that finding answers it."
	const helpSays = "the gap minted from this finding answers it"

	help, err := repotree.Plugin("tools", "internal", "cli", "seat", "help", "finding.md")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(help)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), helpSays) {
		t.Fatalf("the `finding` help page no longer says %q — the lens definitions are held to the model "+
			"that page teaches, so this end has to hold first", helpSays)
	}

	paths, err := repotree.Constitutions()
	if err != nil {
		t.Fatal(err)
	}
	lenses := 0
	for _, p := range paths {
		if !strings.HasPrefix(filepath.Base(p), "red-lens-") {
			continue
		}
		lenses++
		text, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(text), lensSays) {
			t.Errorf("%s does not state the one gap that answers a contradiction (%q): a lens told of "+
				"another answering gap reaches for an act the tool does not have", filepath.Base(p), lensSays)
		}
	}
	if lenses == 0 {
		t.Fatal("no red-lens-* definition found — the census measured nothing")
	}
}
