package goldentest

import (
	"os"
	"testing"
)

// THE GATE WHOSE FAILURE IS GREEN. Assert rewrites the golden when Update is set and
// compares when it is not; an inverted gate therefore rewrites every golden on every
// ordinary run and passes, which no test run can notice from the inside. Pinning the
// predicate is the only place the polarity can be caught.
func TestTheUpdateGateIsExactAndOffByDefault(t *testing.T) {
	for _, c := range []struct {
		env  string
		want bool
	}{
		{"1", true},
		{"", false},
		{"0", false},
		{"true", false},
		{"yes", false},
		{"11", false},
		{" 1", false},
	} {
		if got := updateFrom(c.env); got != c.want {
			t.Errorf("updateFrom(%q) = %v, want %v", c.env, got, c.want)
		}
	}
	// And the package-level value this suite is running under: unless the operator asked
	// for a regeneration, every Assert in the tree must be COMPARING.
	if asked := os.Getenv("UPDATE_GOLDENS") == "1"; Update != asked {
		t.Errorf("Update = %v while UPDATE_GOLDENS=%q — a run that was not asked to "+
			"regenerate would rewrite every golden it touched and pass", Update, os.Getenv("UPDATE_GOLDENS"))
	}
}
