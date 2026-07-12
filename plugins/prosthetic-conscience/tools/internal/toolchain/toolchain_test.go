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
