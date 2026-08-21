package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// EVERY DOCUMENT LOADS, HERE, RATHER THAN AT FIRST USE.
//
// Rendering moved out of package init because init order could not supply the values (measured: the
// first version panicked on {{.NearMatchTopN}}, because package seat initializes before the package
// that supplies it). The cost of that move is that a malformed document now fails when its verb is
// first constructed rather than at startup — so this forces all of them, and CI sees a broken one
// without anybody having to run the verb it belongs to.
func TestEveryHelpDocumentLoads(t *testing.T) {
	if len(seat.HelpNames()) == 0 {
		t.Fatal("no help documents embedded at all — the //go:embed pattern matched nothing, and every verb would panic on construction")
	}
	for _, n := range seat.HelpNames() {
		func() {
			menu, detail, err := seat.LoadHelp(n)
			if err != nil {
				t.Errorf("help/%s.md: %v", n, err)
				return
			}
			if strings.TrimSpace(menu) == "" || strings.TrimSpace(detail) == "" {
				t.Errorf("help/%s.md parsed to an empty section", n)
			}
		}()
	}
}

// THE MENU LINE IS A MENU LINE.
//
// Cobra prints Short IN FULL inside the parent's Available Commands block, so a paragraph there is
// paid by every seat that opens the listing whether or not it cares about the verb. `finding` ran to
// 961 characters, and the measured failure is a seat reading the first clause and stopping — the
// first clause being the flag list.
func TestMenuLinesStayScannable(t *testing.T) {
	for _, n := range seat.HelpNames() {
		menu, _, err := seat.LoadHelp(n)
		if err != nil {
			t.Errorf("help/%s.md: %v", n, err)
			continue
		}
		if len(menu) > seat.MenuLimit {
			t.Errorf("help/%s.md menu is %d chars, over %d", n, len(menu), seat.MenuLimit)
		}
		if strings.Contains(menu, "\n") {
			t.Errorf("help/%s.md menu spans lines — it is printed inside a listing", n)
		}
		// A MENU LINE NAMES NO FLAGS. Flags are what you need once you have CHOSEN, and cobra
		// prints them structurally on the verb's own page. Putting them in the line a seat reads
		// while deciding is what buried the discriminator underneath the mechanics.
		if strings.Contains(menu, "--") {
			t.Errorf("help/%s.md menu names a flag — the menu says what the verb is FOR and when to reach for it; flags belong in the detail, which cobra also prints as a Flags block", n)
		}
	}
}

// A DOCUMENT NOBODY CLAIMS IS AS MUCH A DEFECT AS A VERB WITHOUT ONE.
//
// The lookup is by NAME, which is a pattern standing in for a schema: nothing structural ties a file
// to the verb it serves. A verb with no document panics loudly at construction, which covers one
// direction. This covers the other — an orphan renders nowhere and reads, to anyone auditing the
// help, exactly like a verb that was never written.
func TestEveryHelpDocumentIsClaimedByAVerb(t *testing.T) {
	// BUILD EVERY ROLE'S TREE FIRST. The claim is recorded when a verb is constructed, so a gate
	// that read it without building would find nothing claimed and report every document orphaned.
	for _, role := range []string{"lens", "merge", "blue", "bench"} {
		if root := NewRootFor(record.SampleSeatOf(role)); root == nil {
			t.Fatalf("no command tree for %s", role)
		}
	}
	claimed := map[string]bool{}
	for _, k := range seat.RegisteredHelpKeys() {
		claimed[k] = true
	}
	if len(claimed) == 0 {
		t.Fatal("no verb registered a help key — this gate would pass vacuously and report every document as orphaned or none")
	}
	for _, n := range seat.HelpNames() {
		if !claimed[n] {
			t.Errorf("help/%s.md is claimed by no verb — it renders nowhere", n)
		}
	}
	for k := range claimed {
		if _, ok := seat.HelpSourceFor(k); !ok {
			t.Errorf("a verb claims help key %q and no help/%s.md exists", k, k)
		}
	}
}
