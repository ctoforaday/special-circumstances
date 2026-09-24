package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
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
	// tmpRun already releases this run's handle; newRun only adds the class vocabulary.
	dir := recordtest.TmpRun(t)
	if err := StageForRun(mustRun(t, dir), fixtureClasses...); err != nil {
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

// newRunWithLanes is newRun plus the cast setup writes for a run with this many lanes.
//
// A LANE IS A LANE BECAUSE THE CAST SAYS SO — its id carries no such fact, since how many lanes a
// run seats is a run parameter and no compiled table can name them. So a fixture that registers a
// lane writes the cast; one testing the NO-CAST case still gets a bare run from newRun.
func newRunWithLanes(t *testing.T, lanes int) string {
	t.Helper()
	dir := newRun(t)
	seats, laneSeats := CastFor(nil, lanes)
	if _, err := Append(Identity{Run: mustRun(t, dir), SeatID: HarnessSeat},
		&recordpb.Cast{SeatIds: seats, LaneSeatIds: laneSeats}); err != nil {
		t.Fatalf("seed the cast: %v", err)
	}
	return dir
}

// castBody is a Cast naming these seats and these lanes. NOT `castWith`, which castcoverage_test
// already uses for a whole RUN built around a cast — the same word for two things reads as
// confirmation to anyone checking.
func castBody(seats, lanes []string) *recordpb.Cast {
	return &recordpb.Cast{SeatIds: seats, LaneSeatIds: lanes}
}
