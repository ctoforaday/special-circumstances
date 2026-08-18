package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

// naming.go's type doc says the surviving-name count "is printed with the result, and the tests
// refuse a redaction that removed nothing from an input that named something." The printing half
// was not true: no caller printed it, so an arm whose redactor had stopped matching would have
// produced a `none` run byte-identical to `partial` and both arms would have reported the same
// behaviour — a null manufactured by the instrument. These pin the report line that closes it.

func TestNamingTreatmentSaysNotMeasuredRatherThanZero(t *testing.T) {
	// "0 names survived" and "I could not open the constitution" are different facts, and only
	// one of them means the treatment worked. Reporting the second as the first is the exact
	// shape this whole report exists to avoid.
	got := namingTreatment("lens", "/nonexistent-constitutions", seatprobe.NewSurface(cli.CommandPaths()), seatprobe.NamingNone, false)
	if !strings.Contains(got, "NOT MEASURED") {
		t.Errorf("an unreadable constitution reported %q — it must say NOT MEASURED, never a count", got)
	}
}

// The shipped constitutions must actually BE treatable, or every naming experiment run against
// them is comparing two identical prompts. This is the assertion that would have caught a
// redactor gone stale before an arm was ever dispatched.
func TestTheShippedConstitutionsAreActuallyRedacted(t *testing.T) {
	sf := seatprobe.NewSurface(cli.CommandPaths())
	// EXPLICIT, because the default resolves relative to the working directory — which is the
	// package dir under `go test`, and is why a probe launched from the wrong directory failed
	// all nine boards at once.
	agents := filepath.Join("..", "..", "..", "agents")
	for _, role := range []string{"blue", "lens", "merge", "bench"} {
		src, err := constitutionFor(role, agents)
		if err != nil {
			t.Fatalf("no constitution for %s: %v", role, err)
		}
		line := namingTreatment(role, agents, sf, seatprobe.NamingNone, false)
		if strings.Contains(line, "NOT MEASURED") {
			t.Errorf("%s (%s): %s", role, src, line)
			continue
		}
		if strings.Contains(line, "REMOVED NOTHING") {
			t.Errorf("%s: the `none` arm is the `partial` arm wearing another label — %s", role, line)
		}
	}
}
