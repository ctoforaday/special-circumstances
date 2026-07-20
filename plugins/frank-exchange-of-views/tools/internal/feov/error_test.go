package feov

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorfCarriesCodeAndMessage(t *testing.T) {
	e := Errorf(NotFound, "no gap %s", "R1-9")
	if e.Error() != "no gap R1-9" {
		t.Errorf("message = %q", e.Error())
	}
	if e.Code != NotFound {
		t.Errorf("code = %q, want not_found", e.Code)
	}
}

func TestCodeOfWalksTheChainAndDefaults(t *testing.T) {
	// A plain error has no code: the edge renders it as the generic "error".
	if got := CodeOf(errors.New("plain")); got != "error" {
		t.Errorf("plain error code = %q, want error", got)
	}
	// A coded error still reports its code after fmt.Errorf %w wraps it — which is exactly
	// what the CLI edge does when it prefixes the role, so the code must survive that.
	wrapped := fmt.Errorf("merge: %w", Errorf(RoleViolation, "wrong role"))
	if got := CodeOf(wrapped); got != string(RoleViolation) {
		t.Errorf("wrapped code = %q, want role_violation", got)
	}
}

func TestWrapKeepsTheCause(t *testing.T) {
	cause := errors.New("disk gone")
	w := Wrap(Conflict, cause, "could not close %s", "R1-1")
	if !errors.Is(w, cause) {
		t.Error("Wrap lost the cause — errors.Is could not find it")
	}
	if CodeOf(w) != string(Conflict) {
		t.Errorf("Wrap code = %q, want conflict", CodeOf(w))
	}
}
