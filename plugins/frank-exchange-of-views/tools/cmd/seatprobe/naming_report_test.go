package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE SURVIVING-NAME COUNT IS PRINTED WITH THE RESULT, and these pin the line.
//
// It used to guard an ARM: a redactor that had stopped matching would produce a `none` run
// byte-identical to `partial`, and both would report the same behaviour — a null manufactured by
// the instrument. The arms are gone, and the check that outlived them is the stronger one: the
// SHIPPED constitution must name no verb, so a name surviving is a regression rather than a
// treatment that failed to land.

func TestNamingTreatmentSaysNotMeasuredRatherThanZero(t *testing.T) {
	// "0 names survived" and "I could not open the constitution" are different facts, and only
	// one of them means the treatment worked. Reporting the second as the first is the exact
	// shape this whole report exists to avoid.
	got := namingTreatment("lens", "/nonexistent-constitutions", seatprobe.NewSurface(cli.CommandPaths()))
	if !strings.Contains(got, "NOT MEASURED") {
		t.Errorf("an unreadable constitution reported %q — it must say NOT MEASURED, never a count", got)
	}
}

// THE SHIPPED CONSTITUTIONS NAME NO VERB, checked before a seat is ever dispatched. A partial
// list in front of a seat satisfies its need to know what exists and stops it looking — measured
// at 58% of the surface seen, against 95% with the list removed — so a name that has crept back
// in is the defect, not an arm.
func TestTheShippedConstitutionsNameNoVerbAtDispatch(t *testing.T) {
	sf := seatprobe.NewSurface(cli.CommandPaths())
	// The default now resolves through repotree and would be correct here too; this stays
	// explicit so the gate names the directory it is asserting about rather than inheriting it
	// from the code under test.
	agents, err := repotree.Plugin("agents")
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		src, err := constitutionFor(role, agents)
		if err != nil {
			t.Fatalf("no constitution for %s: %v", role, err)
		}
		line := namingTreatment(role, agents, sf)
		if strings.Contains(line, "NOT MEASURED") {
			t.Errorf("%s (%s): %s", role, src, line)
			continue
		}
		if strings.Contains(line, "NAMES") {
			t.Errorf("%s: %s", role, line)
		}
	}
}
