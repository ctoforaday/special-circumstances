package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// A GAP HAS A SUBJECT. A mint with an empty or blank --quote and no --about-kind/--about is
// refused at the write and records nothing, whatever stage the run is at; one that names its
// subject lands. A smoke drove `mint --quote ""` by hand and the record admitted a gap nothing
// locates.
func TestAMintWithNoSubjectIsRefused(t *testing.T) {
	runDir := newRun(t)
	registerChairOnce(t, runDir)
	registerLensOnce(t, runDir)
	mint := func(key string, subject ...string) error {
		t.Helper()
		_, err := runMint(t, runDir, append([]string{"--run", runDir, "--seat-id", lensSeat, "--key", key,
			"--class", "spec_underspecification", "--problem", "the defect " + key, "--fix", "the fix",
			"--check-kind", "document", "--check", "the acceptance check",
			"--severity", "medium", "--likelihood", "medium", "--impact", "medium"}, subject...)...)
		return err
	}
	for _, c := range []struct {
		name    string
		subject []string
		want    string
	}{
		{"no --quote and no --about", nil, "a gap needs a subject"},
		{"an empty --quote", []string{"--quote", ""}, "a gap needs a subject"},
		{"a blank --quote", []string{"--quote", " \t\n"}, "--quote is blank"},
		{"a blank --quote beside an --about", []string{"--quote", "  ", "--about-kind", "section", "--about", "Findings"}, "--quote is blank"},
	} {
		t.Run(c.name, func(t *testing.T) {
			err := mint("k-"+strings.ReplaceAll(c.name, " ", "-"), c.subject...)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("a mint with %s must be refused with %q: %v", c.name, c.want, err)
			}
			if !strings.Contains(err.Error(), "--about-kind/--about") {
				t.Errorf("the refusal does not name the other way to state a subject: %v", err)
			}
			if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 0 {
				t.Fatalf("the refused mint recorded %d mint(s)", n)
			}
		})
	}
	if err := mint("k-about", "--about-kind", "section", "--about", "Findings"); err != nil {
		t.Fatalf("a mint about a section was refused: %v", err)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_MINT); n != 1 {
		t.Fatalf("the mint about a section recorded %d mint(s), want 1", n)
	}
}
