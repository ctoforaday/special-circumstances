package record

import (
	"strings"
	"testing"
)

// Every entry must say something. A key with an empty value is the absence this table exists
// to remove, wearing a present key.
func TestEveryDeltaCarriesText(t *testing.T) {
	if len(capabilityDeltas) == 0 {
		t.Fatal("the table is empty, so every other check here passes without reading anything")
	}
	for v, d := range capabilityDeltas {
		if len(strings.TrimSpace(d)) < 10 {
			t.Errorf("version %s has a delta of %q — too short to answer \"should I care?\"", v, d)
		}
	}
}

// THE ANSWER A REFUSED PREFLIGHT NEEDS: what this binary cannot do, oldest first.
func TestDeltasBetweenListsOnlyWhatIsMissing(t *testing.T) {
	got := DeltasBetween("0.69.0", "0.71.0")
	if len(got) != 2 {
		t.Fatalf("0.69.0 -> 0.71.0 should list 0.70.0 and 0.71.0, got %d: %v", len(got), got)
	}
	if !strings.HasPrefix(got[0], "0.70.0:") || !strings.HasPrefix(got[1], "0.71.0:") {
		t.Errorf("oldest first, each naming its version: %v", got)
	}
	// The running version's OWN delta is not missing — it already has it.
	for _, g := range got {
		if strings.HasPrefix(g, "0.69.0:") {
			t.Errorf("the running version's own delta is not something it lacks: %v", got)
		}
	}
}

// A binary NEWER than the manifest is a different problem, and listing deltas it already has
// would be noise dressed as diagnosis.
func TestDeltasBetweenIsSilentWhenNotBehind(t *testing.T) {
	for _, tc := range [][2]string{{"0.71.0", "0.71.0"}, {"0.71.0", "0.70.0"}, {"", "0.71.0"}, {"0.69.0", ""}} {
		if got := DeltasBetween(tc[0], tc[1]); got != nil {
			t.Errorf("DeltasBetween(%q, %q) = %v, want nothing", tc[0], tc[1], got)
		}
	}
}

// Ordering is numeric, not lexicographic. "0.9.0" < "0.10.0" is the case a string compare gets
// backwards, and it decides which deltas a refusal prints.
func TestVersionOrderingIsNumeric(t *testing.T) {
	if compareVersions("0.9.0", "0.10.0") >= 0 {
		t.Error("0.9.0 must sort before 0.10.0; a string compare gets this backwards")
	}
	if compareVersions("0.71.0", "0.8.0") <= 0 {
		t.Error("0.71.0 must sort after 0.8.0")
	}
	if compareVersions("1.0.0", "1.0.0") != 0 {
		t.Error("equal versions must compare equal")
	}
}
