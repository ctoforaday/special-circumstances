package record

import "testing"

// The role boundary is the engine's premise, so it gets a test per crossing
// rather than one happy-path assertion. Before this binding existed, every
// "refused" case below SUCCEEDED.
func TestSeatRoleBinding(t *testing.T) {
	allowed := map[string][]string{
		"lens":  {"red-lens-r1-L1", "red-lens-r12-L6"},
		"merge": {"red-merge-r1", "red-merge-r12"},
		"blue":  {"blue-lane-1", "blue-respond-r3", "blue-synthesize", "frontier"},
		"bench": {"judge-r1", "judge-petition", "judge-terminal", "assemble"},
	}
	for role, seats := range allowed {
		for _, s := range seats {
			if err := CheckSeatRole(role, s); err != nil {
				t.Errorf("%s should accept its own seat %q: %v", role, s, err)
			}
		}
	}
	// Every cross-role pairing is refused, not just the one that motivated this.
	for role := range allowed {
		for other, seats := range allowed {
			if other == role {
				continue
			}
			for _, s := range seats {
				err := CheckSeatRole(role, s)
				if err == nil {
					t.Errorf("%s accepted %s seat %q — the role boundary is not enforced", role, other, s)
					continue
				}
				if got := err.Error(); !contains(got, other) {
					t.Errorf("refusal for %q should name its real role %q: %s", s, other, got)
				}
			}
		}
	}
	if err := CheckSeatRole("blue", "totally-made-up"); err == nil {
		t.Error("a seat id no dispatch created was accepted")
	}
}

func contains(s, sub string) bool {
	return len(sub) > 0 && len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
