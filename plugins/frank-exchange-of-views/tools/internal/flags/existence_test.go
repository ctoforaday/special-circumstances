package flags

import "testing"

func TestExistenceValueAcceptsCanonicalAxes(t *testing.T) {
	for _, s := range []string{"verified", "suspected", " verified ", "suspected\t"} {
		var v ExistenceValue
		if err := v.Set(s); err != nil {
			t.Errorf("Set(%q) errored: %v", s, err)
		}
		if !v.IsSet() {
			t.Errorf("Set(%q) left IsSet false", s)
		}
	}
}

func TestExistenceValueRejectsEverythingElse(t *testing.T) {
	for _, s := range []string{"probable", "VERIFIED", "yes", "true", "", "verified,suspected"} {
		var v ExistenceValue
		if err := v.Set(s); err == nil {
			t.Errorf("Set(%q) should have errored (enum is verified|suspected)", s)
		}
		if v.IsSet() {
			t.Errorf("a rejected Set(%q) must not mark the flag set", s)
		}
	}
}

// Unset stays empty so it never appears in the event — the pointer-set semantics the
// record format requires.
func TestExistenceValueUnsetIsEmpty(t *testing.T) {
	var v ExistenceValue
	if v.IsSet() || v.String() != "" {
		t.Errorf("a fresh ExistenceValue must be unset and empty, got IsSet=%v String=%q", v.IsSet(), v.String())
	}
	if v.Type() != "existence" {
		t.Errorf("Type = %q, want existence", v.Type())
	}
}
