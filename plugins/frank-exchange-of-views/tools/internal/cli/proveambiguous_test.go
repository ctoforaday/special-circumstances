package cli

import (
	"strings"
	"testing"
)

// PROVE'S UNIQUENESS REFUSAL NAMES WHERE THE QUOTE STANDS (m35, blue ev462: "the quote occurs more
// than once" twice, no location, and blue's own anchored search for the second copy missed it). The
// refusal names each occurrence by its section and sentence, so the seat can quote the one it means.
func TestProveRefusesAnAmbiguousQuoteNamingEachOccurrence(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Method\n\nWe ran the sieve. 91 is 7 times 13.\n\n## Check\n\nBy hand, 91 is 7 times 13.\n")
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "blue-respond"); err != nil {
		t.Fatal(err)
	}
	_, err := run(t, "prove", "--run", runDir, "--seat-id", "blue-respond", "--quote", "91 is 7 times 13",
		"--script", script(t, runDir, "f.js", "console.log(7*13)"), "--reason", "r")
	if err == nil {
		t.Fatal("prove placed a proof on a quote that occurs twice")
	}
	for _, want := range []string{
		"occurs more than once in the report",
		"it stands 2 times",
		`under the heading "Method", paragraph 1, in the sentence "91 is 7 times 13"`,
		`under the heading "Check", paragraph 1, in the sentence "By hand, 91 is 7 times 13"`,
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not say %q:\n%v", want, err)
		}
	}
}
