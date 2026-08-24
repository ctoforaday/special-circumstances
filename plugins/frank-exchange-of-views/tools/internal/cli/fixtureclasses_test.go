package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"path/filepath"
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
	dir := t.TempDir()
	// The run's cached *sql.DB outlives the test otherwise; see record.CloseAll. Registered after
	// t.TempDir so LIFO runs it first and the directory is removable when cleanup reaches it.
	t.Cleanup(func() { _ = recordsql.Close(filepath.Join(dir, "records", "record.db")) })
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
	"anchor-as-fields", "anchor-visibility", "carry-vs-close",
	"cross-seat-visibility", "dispute-to-regrade", "grade-dispute-visibility",
	"halt-is-its-own-verb", "json-vs-markdown", "lineage-completion",
	"motion-test", "one-record", "overclaimed-independence",
	"payload-channel", "read-surface", "reference-integrity",
	"residue-carrying", "some-class", "state-checks",
}, shippedClasses...)

// shippedClasses mirrors feov-memory/class-registry.json.
var shippedClasses = []string{
	"citation-figure-misattribution", "figure-recount-fails", "incomplete-repair-propagation",
	"policy-without-mechanism", "derivation-status-overclaim", "cross-section-contradiction",
	"enumeration-non-exhaustive", "unhandled-degenerate-case", "false-universal",
	"unverified-composition", "self-attestation", "spec-underspecification",
	"claim-contradicts-own-record", "reader-channel-mismatch", "exhaustive-sweep-omits-case",
	"risk-coverage-omission", "live-source-drift", "within-source-condition-misattribution",
	"partial-control-coverage", "figure-miscomposition", "recurrence-detector-keying",
	"pricing-basis-drift", "metric-conflation", "citation-status-drift",
	"undecided-disjunction", "causal-narrative-fails-reproduction", "name-keying-vs-marker",
	"reflexivity-blindspot", "missing-root-invariant", "audited-artifact-sibling-halo",
	"artifact-preservation", "config-semantics-error", "vote-laundering",
	"verification-scope-blindspot", "doctrine-vs-implementation", "measurement-methodology-drift",
	"cross-corpus-id-collision", "negative-definition",
}
