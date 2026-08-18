package main

import "testing"

// The gate's own contract. runSelftest (main.go) proves the matcher can FIRE — the property
// whose absence would make this report a clean board forever. These pin the parts a reader
// would otherwise have to infer: what counts as agent-facing, and that only ADDED lines count.

func TestOnlyAddedLinesCount(t *testing.T) {
	// A REMOVED archaeology line is the fix, not the defect. Scanning it would fail the very
	// commit that cleans a file, which would make the gate an argument against sweeping.
	diff := `+++ b/plugins/x/agents/seat.md
-the ` + "`blue confidence`" + ` verb is RETIRED (0.54.0).
 context line that used to say something
`
	if got := scan(diff); len(got) != 0 {
		t.Errorf("fired on a removed or context line, so cleaning a file would fail the gate: %v", got)
	}
}

func TestAddedArchaeologyIsCaught(t *testing.T) {
	diff := `+++ b/plugins/x/agents/seat.md
+It used to be armed only by passing --bin-dir.
`
	got := scan(diff)
	if len(got) != 1 {
		t.Fatalf("added archaeology produced %d finding(s), want 1: %v", len(got), got)
	}
	if got[0].file != "plugins/x/agents/seat.md" {
		t.Errorf("finding names %q, not the file the diff header set", got[0].file)
	}
	// The REASON travels with the finding. A gate that says only "line 12 is wrong" makes the
	// author guess at the rule, and the guess is usually "delete the sentence".
	if got[0].why == "" {
		t.Error("the finding carries no reason — the author is left to infer the rule")
	}
}

// SCOPE IS A DECISION, NOT AN ACCIDENT. Go comments are out of scope by design (see sweep.go);
// a gate that silently widened would fail changes for a rule nobody stated.
func TestScope(t *testing.T) {
	in := map[string]bool{
		"plugins/frank-exchange-of-views/agents/red-auditor.md":                      true,
		"plugins/frank-exchange-of-views/commands/research.md":                       true,
		"plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md":          true,
		"plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js": true,
		"plugins/prosthetic-conscience/skills/terse-communication/SKILL.md":          true,
		"CLAUDE.md": true,

		"plugins/frank-exchange-of-views/tools/internal/record/record.go": false,
		"scripts/check/gates.go": false,
		"plans/some-plan.md":     false,
		"README.md":              false,
	}
	for path, want := range in {
		if got := isAgentFacing(path); got != want {
			t.Errorf("isAgentFacing(%q) = %v, want %v", path, got, want)
		}
	}
}

// A phrase outside an agent-facing file is not this gate's business, even when it is textbook
// archaeology — which is what keeps the scope claim above honest rather than aspirational.
func TestArchaeologyInCodeIsIgnored(t *testing.T) {
	diff := `+++ b/scripts/check/gates.go
+// This used to be a per-PR rule and was removed.
`
	if got := scan(diff); len(got) != 0 {
		t.Errorf("fired outside the declared scope: %v", got)
	}
}

// The matcher must stay quiet on live prose, or authors learn to ignore it. This is the same
// assertion runSelftest makes; it lives here too because a false-positive regression is the
// failure most likely to get the gate turned off.
func TestLiveProseIsNotArchaeology(t *testing.T) {
	diff := `+++ b/plugins/x/agents/seat.md
+A realized risk is no longer a probability: it contributes 0 to mass.
+Substance leaves the report only through the retire verb.
+Read the whole list before your first act.
`
	if got := scan(diff); len(got) != 0 {
		t.Errorf("%d false positive(s) on live prose: %v", len(got), got)
	}
}

// The selftest is the gate on the gate. If it ever passes vacuously, everything above is decor.
func TestSelftestActuallyChecks(t *testing.T) {
	if err := runSelftest(); err != nil {
		t.Fatalf("selftest failed: %v", err)
	}
}
