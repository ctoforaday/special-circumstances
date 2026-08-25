package report

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// newRun makes a run directory whose gap-class vocabulary is DECLARED.
//
// Every fixture here used to be exempt from the class check by accident: `knownClasses` returned
// nil when no registry was staged, `validateClass` read nil as "accept anything", and so
// `--class anything-at-all` succeeded in every test in this package — while the flag's own help
// said the registry constrained it. The tolerance was the only thing between those two facts.
//
// The exemption is gone, so a fixture states its vocabulary like a real run does. It stages the
// SHIPPED registry's slugs plus the placeholders these tests mint with, and deliberately NOT the
// ones tests use to prove an unknown class is refused.
func newRun(t *testing.T) string {
	t.Helper()
	dir := recordtest.TmpRun(t)
	if err := record.StageForRun(dir, fixtureClasses...); err != nil {
		t.Fatalf("stage the class registry: %v", err)
	}
	return dir
}

// fixtureClasses is the shipped registry plus this package's placeholders. Kept here rather than
// read from feov-memory so a test binary does not depend on locating the repository root, and so
// a slug a test needs is visible beside the tests that need it.
var fixtureClasses = append([]string{
	// Placeholders the fixtures mint with. Short on purpose: a test about closure lineage should
	// not have to pick a real taxonomy entry to say "some class".
	"x", "c", "a", "g", "scope-creep", "safety", "integrity", "correctness", "overclaim",
	"resolved", "fuzzcls", "amends_prior", "attestation-inflation", "citation-drift",
	"unverified-arithmetic", "sibling-halo",
	// Descriptive placeholders: a fixture names the thing it is about, which reads better in
	// a failure than "x" and costs nothing — the point is that the vocabulary is DECLARED.
}, recordtest.ShippedClasses...)
