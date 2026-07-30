package toolchain

import "testing"

func TestPresentAbsent(t *testing.T) {
	if Present("sc-definitely-not-a-real-binary-xyz") {
		t.Fatal("absent binary reported present")
	}
}

func TestProbeUsesCheckCmdFirstField(t *testing.T) {
	tools := []Tool{{Name: "phantom", CheckCmd: "sc-definitely-not-real-xyz --version"}}
	got := Probe(tools)
	if len(got) != 1 {
		t.Fatalf("want 1 status, got %d", len(got))
	}
	if got[0].Found {
		t.Fatal("phantom tool should be absent")
	}
	if got[0].Name != "phantom" {
		t.Fatalf("status lost its tool identity: %q", got[0].Name)
	}
}

// The fallback to Name when CheckCmd yields no first field. Every fixture in the suite
// carried a non-empty CheckCmd, so `len(fields) == 0` was never true and a mutation
// sweep flipped it with nothing failing. A manifest entry with no CheckCmd — or one
// that is whitespace — must probe the tool's own Name, not the empty string, because
// Present("") is a lookup that can only ever report absent, and a tool reported absent
// while installed is a DEGRADED verdict the operator cannot act on.
func TestProbeFallsBackToNameWithoutACheckCmd(t *testing.T) {
	for _, cmd := range []string{"", "   ", "\t\n"} {
		got := ProbeIn([]Tool{{Name: "go", CheckCmd: cmd}}, "local")
		if len(got) != 1 {
			t.Fatalf("want 1 status, got %d", len(got))
		}
		if !got[0].Found {
			t.Errorf("CheckCmd=%q: probed something other than the tool's Name — go is on PATH in every environment that can run this test", cmd)
		}
	}
	// And the first field still wins when there is one.
	got := ProbeIn([]Tool{{Name: "not-a-real-binary-xyz", CheckCmd: "go version"}}, "local")
	if !got[0].Found {
		t.Error("CheckCmd's first field must override Name")
	}
}
