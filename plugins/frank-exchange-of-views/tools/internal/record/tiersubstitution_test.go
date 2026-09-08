package record

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/modeltier"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/servedmodel"
)

// THE RUN THIS GATE IS NAMED FOR.
//
// 2026-08-23: `claude-fable-5` configured for the bulk tier, `claude-opus-4-8` served to all 44
// bulk seats, ~$379 spent, and the one component that noticed graded it a WARNING — because opus
// is CHEAPER than fable and cheaper had been filed as "verification may be discounted". A research
// debate does not buy a tier for its price; it buys the strength of the party arguing each side.
// So the substitution stops the run, and it stops it in either direction.

func runWithTiers(t *testing.T, cfg string) string {
	t.Helper()
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "inputs", "run-config.json"), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return run
}

// configuredFor is the class -> configured-model join SeatModels makes, so these tests drive
// TierSubstitution with the same input the read path builds.
func configuredFor(t *testing.T, run string, seatID string) string {
	t.Helper()
	model, judgment := modeltier.Config(run)
	if TierClassOfSeat(seatID) == "judgment" {
		return judgment
	}
	if TierClassOfSeat(seatID) == "" {
		return ""
	}
	return model
}

const fableSonnet = `{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`

func TestTheGateStopsTheRunTheRetrospectiveMeasured(t *testing.T) {
	run := runWithTiers(t, fableSonnet)
	got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "blue-lane-1"), servedmodel.Observation{
		Served: "claude-opus-4-8", Requested: "claude-fable-5", Declared: true})
	if got == "" {
		t.Fatal("a bulk seat answered by opus against a configured fable must be reported")
	}
	for _, want := range []string{"claude-fable-5", "claude-opus-4-8", "declared the substitution"} {
		if !strings.Contains(got, want) {
			t.Errorf("the report must name %q; got:\n%s", want, got)
		}
	}
}

// The judgment tier was served as configured in that same run, and must not be swept up with it.
func TestASeatAnsweredByItsConfiguredTierPasses(t *testing.T) {
	run := runWithTiers(t, fableSonnet)
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "red-chair-r1"), servedmodel.Observation{Served: "claude-sonnet-5"}); got != "" {
		t.Fatalf("judgment seat on its configured sonnet: %s", got)
	}
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "blue-lane-2"), servedmodel.Observation{Served: "claude-fable-5"}); got != "" {
		t.Fatalf("bulk seat on its configured fable: %s", got)
	}
}

// A tier swap NOBODY DECLARED still refuses: the served model is measured either way, and the
// declaration is evidence about the swap rather than the swap itself.
func TestAnUndeclaredMismatchAlsoRefuses(t *testing.T) {
	run := runWithTiers(t, fableSonnet)
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "judge-r1"), servedmodel.Observation{Served: "claude-haiku-4-5"}); got == "" {
		t.Fatal("a judgment seat answered by haiku against a configured sonnet must be reported")
	}
}

// NOT MEASURED IS NOT A PASS AND NOT A FAILURE. The gate says nothing, because the stderr line at
// the call site already said the measurement did not happen; asserting soundness here would be
// the exact substitution of a miss for a clean board this whole change is about.
func TestAnUnmeasuredSeatIsNotJudged(t *testing.T) {
	run := runWithTiers(t, fableSonnet)
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "blue-lane-1"), servedmodel.Observation{}); got != "" {
		t.Fatalf("nothing measured, nothing to judge: %s", got)
	}
}

// A run that declared no tier for a class cannot hold a seat to one.
func TestARunWithNoDeclaredTierDoesNotRefuse(t *testing.T) {
	run := runWithTiers(t, `{}`)
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "blue-lane-1"), servedmodel.Observation{Served: "claude-opus-4-8"}); got != "" {
		t.Fatalf("no configured tier: %s", got)
	}
}

// The operator is not a debating seat and rides no tier.
func TestSeatsThatRideNoTierAreNotGated(t *testing.T) {
	run := runWithTiers(t, fableSonnet)
	if got := TierClassOfSeat(OperatorRole); got != "" {
		t.Fatalf("the operator has no tier class, got %q", got)
	}
	if got := TierSubstitution(mustRun(t, run), configuredFor(t, run, OperatorRole), servedmodel.Observation{Served: "claude-haiku-4-5"}); got != "" {
		t.Fatalf("operator: %s", got)
	}
}

// THE OVERRIDE IS THE OPERATOR'S AND IT IS A FIELD ON THE RUN, never a flag a seat could type —
// the seat is the party whose adversary strength is in question.
func TestTheOperatorsStandingConsentLetsTheRunProceed(t *testing.T) {
	run := runWithTiers(t, `{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5","allowModelSubstitution":true}`)
	got := TierSubstitution(mustRun(t, run), configuredFor(t, run, "blue-lane-1"), servedmodel.Observation{
		Served: "claude-opus-4-8", Requested: "claude-fable-5", Declared: true})
	if !strings.Contains(got, "ALLOWED BY THIS RUN'S CONFIG") {
		t.Fatalf("a consented substitution must say so rather than read as an unconsented one: %q", got)
	}
}

// ABSENT MEANS NO. A run whose config predates the field never consented to anything, and the
// failing direction of a gate has to be the safe one.
func TestConsentIsNotInferredFromAnAbsentField(t *testing.T) {
	if allowSubstitution(mustRun(t, runWithTiers(t, fableSonnet))) {
		t.Error("an absent allowModelSubstitution must not read as consent")
	}
	if allowSubstitution(mustRun(t, t.TempDir())) {
		t.Error("an unreadable run-config must not read as consent")
	}
	if allowSubstitution(mustRun(t, runWithTiers(t, `{ not json`))) {
		t.Error("an unparseable run-config must not read as consent")
	}
}

// The seat-id grammar and the tier-class table are two lists, and TierClassOfSeat is the join.
// A shape whose base is not a key in seatclass would silently make that whole seat class ungated.
func TestEverySeatShapeJoinsToATierClass(t *testing.T) {
	for _, s := range seatShapes {
		if s.base == "" {
			continue // the operator, deliberately
		}
		if got := TierClassOfSeat(s.sample); got != "bulk" && got != "judgment" {
			t.Errorf("seat %s (base %q) has no tier class — that class of seat would never be gated", s.sample, s.base)
		}
	}
	// The petition sitting is named for the seat that petitioned and is handled apart from the
	// table; it is a judgment seat like the rest of the bench.
	if got := TierClassOfSeat("judge-petition-red-chair-r1"); got != "judgment" {
		t.Errorf("petition sitting: got %q", got)
	}
}
