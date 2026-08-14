package seatenv

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
)

func TestEnvWinsOverFlagAndInference(t *testing.T) {
	t.Setenv(Var, "/runs/live")
	got, err := Resolve("/runs/live", func() string { return "/runs/inferred" })
	if err != nil || got != "/runs/live" {
		t.Errorf("got %q, %v — want the injected value", got, err)
	}
}

// THE MEASURED FAILURE, refused. blue-respond-r1 typed `special circumstances` where the path
// has a hyphen; the tool obeyed the flag, wrote nothing the run could see, and the seat
// concluded the RULE was broken — abandoning five manifest receipts and filing a false bug.
// Obeying a --run that contradicts the dispatch is what made that possible.
func TestATypoedRunIsRefusedNotObeyed(t *testing.T) {
	t.Setenv(Var, "/c/Users/gb/Projects/special-circumstances/research/2026-08-05_smoke")
	_, err := Resolve("/c/Users/gb/Projects/special circumstances/research/2026-08-05_smoke", nil)
	if err == nil {
		t.Fatal("a --run disagreeing with the injected run directory was obeyed")
	}
	if feov.CodeOf(err) != string(feov.Conflict) {
		t.Errorf("code %q, want conflict — two sources disagree, the value is not ill-formed", feov.CodeOf(err))
	}
	// BOTH values must appear: the operator reading this has to see which one is the typo.
	for _, want := range []string{"special circumstances", "special-circumstances"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not name %q, so the typo is invisible:\n%s", want, err)
		}
	}
}

// An AGREEING --run is not an error. Prompts still emit it at 16 sites, and refusing the
// correct value would break every one of them for no gain.
func TestAnAgreeingFlagIsFine(t *testing.T) {
	t.Setenv(Var, "/runs/live")
	for _, flag := range []string{"/runs/live", "/runs/live/", "/runs/live\\"} {
		if got, err := Resolve(flag, nil); err != nil || got != "/runs/live" {
			t.Errorf("--run %q: got %q, %v — a trailing separator is the same directory", flag, got, err)
		}
	}
}

// Nothing else is normalised. Case and separator STYLE can be genuinely different paths, and
// a guess attaches a seat's events to the wrong run — the exact outcome this prevents.
func TestOnlyTrailingSeparatorsAreForgiven(t *testing.T) {
	t.Setenv(Var, "/runs/Live")
	if _, err := Resolve("/runs/live", nil); err == nil {
		t.Error("a case difference was smoothed over; this tool does not guess which path was meant")
	}
}

func TestFallsBackThroughFlagThenInference(t *testing.T) {
	t.Setenv(Var, "")
	if got, _ := Resolve("/runs/flag", func() string { return "/runs/inferred" }); got != "/runs/flag" {
		t.Errorf("got %q, want the flag when nothing is injected", got)
	}
	if got, _ := Resolve("", func() string { return "/runs/inferred" }); got != "/runs/inferred" {
		t.Errorf("got %q, want inference when neither is present", got)
	}
	if got, _ := Resolve("", nil); got != "" {
		t.Errorf("got %q, want empty — the caller owns the '--run is required' message", got)
	}
}
