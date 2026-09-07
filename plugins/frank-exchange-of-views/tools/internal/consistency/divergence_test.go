package consistency

import (
	"strings"
	"testing"
)

// EVERY OTHER TEST IN THIS PACKAGE ASSERTS ZERO VIOLATIONS. That proves the oracle agrees
// with a healthy record; it proves nothing about whether the oracle can still DISAGREE.
// An oracle that has quietly lost a rule reports the same empty slice a clean board does
// — the plausible zero, in the one component whose entire job is refusing it. So the
// detectors are exercised here against records built to violate.
//
// findCycle is tested at the DETECTOR rather than through a seeded run, and deliberately:
// its own comment records that the write path cannot construct a cycle (an ancestor must
// already exist), so the rule exists for records built by something other than the tool.
// A test that seeded one through the store would be testing a state the store refuses;
// this one tests the thing the rule is actually for.
func TestTheCycleDetectorDisagreesWhenItShould(t *testing.T) {
	gap := func(supersedes ...string) *gtGap { return &gtGap{supersedes: supersedes} }

	t.Run("a two-gap loop is reported", func(t *testing.T) {
		gt := &groundTruth{gaps: map[string]*gtGap{
			"A-1": gap("B-1"),
			"B-1": gap("A-1"),
		}}
		cyc := findCycle(gt)
		if cyc == "" {
			t.Fatal("a supersedes cycle went unreported — the lineage rule cannot fire")
		}
		if !strings.Contains(cyc, "A-1") || !strings.Contains(cyc, "B-1") {
			t.Errorf("cycle = %q, want it to name both gaps in the loop", cyc)
		}
	})

	t.Run("a longer loop is reported", func(t *testing.T) {
		gt := &groundTruth{gaps: map[string]*gtGap{
			"A-1": gap("B-1"), "B-1": gap("C-1"), "C-1": gap("A-1"),
		}}
		if findCycle(gt) == "" {
			t.Error("a three-gap cycle went unreported")
		}
	})

	t.Run("a self-loop is reported", func(t *testing.T) {
		gt := &groundTruth{gaps: map[string]*gtGap{"A-1": gap("A-1")}}
		if findCycle(gt) == "" {
			t.Error("a gap superseding itself went unreported")
		}
	})

	// And the other direction, which is what keeps the rule usable: a DAG — including a
	// diamond, where one gap is reached twice by different paths — is not a cycle.
	t.Run("a diamond is not a cycle", func(t *testing.T) {
		gt := &groundTruth{gaps: map[string]*gtGap{
			"D-1": gap("B-1", "C-1"),
			"B-1": gap("A-1"),
			"C-1": gap("A-1"),
			"A-1": gap(),
		}}
		if cyc := findCycle(gt); cyc != "" {
			t.Errorf("a diamond was reported as a cycle (%q); every re-merged lineage would be a violation", cyc)
		}
	})
}
