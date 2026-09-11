package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/diagnostics"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/terms"
)

// EVERY MANUAL CARRIES THE WORDS ITS SURFACE USES, AND ONLY THOSE. The registry names which seats
// meet each term; the manual is where a seat is handed them, once, before any page.
func TestEveryManualCarriesTheWordsItsSurfaceUses(t *testing.T) {
	reg, err := terms.Load()
	if err != nil {
		t.Fatal(err)
	}
	for role, seatID := range everySurface() {
		out, pages := manualOf(t, seatID)
		head := strings.Index(out, diagnostics.ManualWordsHeading)
		if head < 0 {
			t.Errorf("%s: the manual has no %q section", role, diagnostics.ManualWordsHeading)
			continue
		}
		// Before the first page, so no page — and no expansion of one — carries it.
		if rule := strings.Index(out, diagnostics.ManualRule); rule >= 0 && head > rule {
			t.Errorf("%s: the words section comes after the first page", role)
		}
		for _, p := range pages {
			if strings.Contains(p.Body, diagnostics.ManualWordsHeading) {
				t.Errorf("%s: page %q carries the words section", role, strings.Join(p.Path, " "))
			}
		}
		mine := map[string]bool{}
		for _, e := range reg.ForSeat(role) {
			mine[e.Term] = true
			if !strings.Contains(out, "  - "+e.Definition) {
				t.Errorf("%s: the manual does not define %q", role, e.Term)
			}
		}
		for _, e := range reg.Entries {
			if !mine[e.Term] && strings.Contains(out, e.Definition) {
				t.Errorf("%s: the manual defines %q, which the registry does not deliver to this surface", role, e.Term)
			}
		}
	}
}

// wordsSection is the manual's WORDS THIS SURFACE USES block — its heading through the blank line
// that ends it — or "" when the manual has none.
func wordsSection(out string) string {
	i := strings.Index(out, diagnostics.ManualWordsHeading)
	if i < 0 {
		return ""
	}
	if j := strings.Index(out[i:], "\n\n"); j >= 0 {
		return out[i : i+j]
	}
	return out[i:]
}

// EVERY SEAT THE REGISTRY NAMES IS A SURFACE, AND EVERY SURFACE GETS WORDS. A seat value that is not
// a role delivers its term to nobody, and a surface with no entry prints no section at all — both
// look, from the registry, like a decision.
func TestEveryRegistrySeatIsASurfaceRole(t *testing.T) {
	reg, err := terms.Load()
	if err != nil {
		t.Fatal(err)
	}
	roles := everySurface()
	for _, s := range reg.Seats() {
		if _, ok := roles[s]; !ok {
			var known []string
			for r := range roles {
				known = append(known, r)
			}
			sort.Strings(known)
			t.Errorf("the terms registry delivers to %q, which is not a surface role (the roles are %s)", s, strings.Join(known, ", "))
		}
	}
	for role := range roles {
		if len(reg.ForSeat(role)) == 0 {
			t.Errorf("surface %q gets no words from the registry", role)
		}
	}
}
