package toolchain

import "testing"

func st(name, tier string, found bool) Status {
	return Status{Tool: Tool{Name: name, Tier: tier, CheckCmd: name + " --version"}, Found: found}
}

// THE STRICTEST DECLARED NEED IS THE BOX'S OBLIGATION. Satisfying it satisfies every other
// declarer; nothing weaker does.
func TestMergeStrictestTakesTheStrongestObligation(t *testing.T) {
	own := []Status{st("jq", "optional", false)}
	sibling := []Status{st("jq", "required", false)}

	for _, order := range []struct {
		name   string
		groups [][]Status
	}{
		{"own first", [][]Status{own, sibling}},
		{"sibling first", [][]Status{sibling, own}},
	} {
		got := MergeStrictest(order.groups...)
		if len(got) != 1 {
			t.Fatalf("%s: one tool declared twice must merge to one obligation, got %d", order.name, len(got))
		}
		if got[0].Tier != "required" {
			t.Errorf("%s: merged tier = %q, want required — a weaker declaration cannot excuse a stronger one",
				order.name, got[0].Tier)
		}
	}
}

// ORDER OF DECLARATION MUST NOT DECIDE THE ANSWER. Before the merge the verdict came from one
// manifest and the rest were decoration; a fold that took first-or-last-wins would replace that
// with a result nobody chose, which is the same defect one layer along.
func TestMergeStrictestIsOrderIndependentForTheTier(t *testing.T) {
	a := []Status{st("git", "recommended", true)}
	b := []Status{st("git", "required", true)}
	c := []Status{st("git", "optional", true)}
	if MergeStrictest(a, b, c)[0].Tier != MergeStrictest(c, a, b)[0].Tier {
		t.Error("the merged tier depends on group order")
	}
}

// ABSENT BY DESIGN IN ONE PLUGIN CANNOT EXCUSE A PLUGIN THAT NEEDS IT HERE. `gh` is genuinely
// not applicable in a cloud session for prosthetic-conscience; a sibling that reaches for it
// there is still missing it, and folding to the winner's flag alone would hide that.
func TestMergeStrictestKeepsNotApplicableOnlyWhenEveryDeclarerSaysSo(t *testing.T) {
	na := st("gh", "recommended", false)
	na.NotApplicable = true
	needs := st("gh", "required", false) // applicable here

	got := MergeStrictest([]Status{na}, []Status{needs})
	if got[0].NotApplicable {
		t.Error("one plugin's not-applicable excused another plugin's real requirement")
	}

	bothNA := st("gh", "required", false)
	bothNA.NotApplicable = true
	if !MergeStrictest([]Status{na}, []Status{bothNA})[0].NotApplicable {
		t.Error("every declarer said out-of-scope and the merge still called it applicable")
	}
}

// A VERSION FACT IS ABOUT THE BINARY, NOT ABOUT WHO ASKED. Only one plugin need declare a
// minimum for the box to be below it.
func TestMergeStrictestCarriesVersionFaultsFromAnyDeclarer(t *testing.T) {
	old := st("go", "optional", true)
	old.TooOld = true
	plain := st("go", "required", true)

	if !MergeStrictest([]Status{plain}, []Status{old})[0].TooOld {
		t.Error("a minimum declared by the weaker entry was dropped by the merge")
	}

	unmeasured := st("node", "optional", true)
	unmeasured.VersionUnmeasured = "check command produced no version"
	if MergeStrictest([]Status{st("node", "required", true)}, []Status{unmeasured})[0].VersionUnmeasured == "" {
		t.Error("an unmeasurable minimum was dropped by the merge")
	}
}

// DEDUPLICATION KEYS OFF THE EXECUTABLE ACTUALLY LOOKED UP, which is what presence keys off —
// so two manifests naming one binary under different `name` values are one obligation.
func TestMergeStrictestKeysOnTheProbedExecutable(t *testing.T) {
	a := Status{Tool: Tool{Name: "jq (json filter)", Tier: "optional", CheckCmd: "jq --version"}}
	b := Status{Tool: Tool{Name: "jq", Tier: "required", CheckCmd: "jq --version"}}
	if got := MergeStrictest([]Status{a}, []Status{b}); len(got) != 1 {
		t.Fatalf("one binary declared under two names must merge to one obligation, got %d", len(got))
	}
}

// AN UNRECOGNISED TIER RANKS LOWEST. A typo in a manifest must not silently escalate a box to
// BLOCKED — the tier vocabulary is checked where manifests are authored, not here.
func TestUnknownTierDoesNotOutrankAKnownOne(t *testing.T) {
	if Stricter("requried", "optional") {
		t.Error("a misspelled tier outranked a real one; a typo must not escalate a verdict")
	}
	if !Stricter("optional", "requried") {
		t.Error("a real tier failed to outrank a misspelled one")
	}
}

// TOOLS DECLARED ONCE PASS THROUGH UNCHANGED AND IN ORDER, so the table stays deterministic.
func TestMergeStrictestPreservesFirstAppearanceOrder(t *testing.T) {
	got := MergeStrictest(
		[]Status{st("git", "required", true), st("go", "required", true)},
		[]Status{st("node", "required", false)},
	)
	want := []string{"git", "go", "node"}
	if len(got) != len(want) {
		t.Fatalf("got %d tools, want %d", len(got), len(want))
	}
	for i, n := range want {
		if got[i].Name != n {
			t.Errorf("position %d = %q, want %q", i, got[i].Name, n)
		}
	}
}
