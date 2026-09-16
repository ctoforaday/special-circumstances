package qualitygate

import (
	"context"
	"errors"
	"testing"
	"time"
)

// A QLTY THAT CRASHED OR TIMED OUT is the same broken machine state as a qlty that is MISSING — the
// gate is not running — and only the missing one was said. The crashed one reached the best-effort
// hook log alone.
func TestAQltyThatCannotRunIsSaid(t *testing.T) {
	dir, file := project(t, "a.go")
	e := healthy(dir)
	e.exec = func(context.Context, string, string, ...string) (string, int, error) {
		return "", 0, errors.New("signal: killed")
	}
	r := gate(e, "Edit", file, time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC))
	if r.Say == "" {
		t.Fatalf("a qlty that could not run was not said: %+v", r)
	}
	if r.Exit != 0 {
		t.Errorf("infrastructure must never fail the event: exit %d", r.Exit)
	}

	// And a qlty that runs clean says nothing.
	e.exec = func(context.Context, string, string, ...string) (string, int, error) { return "", 0, nil }
	if r := gate(e, "Edit", file, time.Date(2026, 9, 16, 12, 0, 0, 0, time.UTC)); r.Say != "" {
		t.Errorf("a clean run spoke: %q", r.Say)
	}
}
