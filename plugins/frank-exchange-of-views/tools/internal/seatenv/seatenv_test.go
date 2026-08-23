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

// WHICH PATH SUPPLIED THE RUN DIRECTORY IS A FACT THE ORDER CANNOT ANSWER AFTERWARDS.
//
// Resolve returns a directory and nothing else, so a caller holding one cannot tell an injected
// value from an inferred one — and that difference is the whole of #512. The source is recorded
// at the write precisely so the next silent run is diagnosable from the record rather than from
// what is missing.
func TestResolveWithSourceNamesThePathThatWon(t *testing.T) {
	never := func() string { t.Helper(); t.Error("inference ran while an earlier source had a value"); return "" }
	for _, tc := range []struct {
		name    string
		env     string
		flag    string
		infer   func() string
		wantDir string
		wantVia RunSource
	}{
		{"the hook injected it", "/runs/a", "", never, "/runs/a", RunFromEnv},
		{"the seat typed it", "", "/runs/b", never, "/runs/b", RunFromFlag},
		{"the tool found the marker itself", "", "", func() string { return "/runs/c" }, "/runs/c", RunFromInference},
		{"nothing supplied one", "", "", func() string { return "" }, "", RunUnresolved},
		{"no inference function at all", "", "", nil, "", RunUnresolved},
		// AGREEMENT IS STILL THE INJECTED PATH. A seat that types the right --run has not made
		// the hook's injection stop being what supplied the value, and reporting `flag` here
		// would say the hook was absent on a run where it fired.
		{"both agree", "/runs/d", "/runs/d/", never, "/runs/d", RunFromEnv},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(Var, tc.env)
			t.Setenv(VarWrapper, "") // hermetic: this table is about Var, the flag and inference
			dir, via, err := ResolveWithSource(tc.flag, tc.infer)
			if err != nil {
				t.Fatalf("ResolveWithSource = %v, want no error", err)
			}
			if dir != tc.wantDir || via != tc.wantVia {
				t.Errorf("ResolveWithSource = (%q, %q), want (%q, %q)", dir, via, tc.wantDir, tc.wantVia)
			}
		})
	}
}

// A DISAGREEMENT RESOLVES TO NO SOURCE, not to whichever one the order would have picked. The
// refusal is the point: naming a winner here would let a run_via field claim a provenance for a
// directory the call never returned.
func TestResolveWithSourceRefusesADisagreementWithoutNamingAWinner(t *testing.T) {
	t.Setenv(Var, "/runs/real")
	t.Setenv(VarWrapper, "")
	dir, via, err := ResolveWithSource("/runs/typo", nil)
	if err == nil {
		t.Fatalf("ResolveWithSource = (%q, %q), want a refusal", dir, via)
	}
	if via != RunUnresolved {
		t.Errorf("a refused resolve reports source %q, want %q", via, RunUnresolved)
	}
}

// THE WRAPPER SUPPLIES THE RUN WHEN THE HOOK DOES NOT, and says so distinctly.
//
// `setup` bakes the run into <runDir>/.bin/feov-record, so a seat that retypes nothing still
// files against the right run. The source stays distinguishable from the hook's on purpose:
// the hook sets Var and Var wins, so reading back "wrapper" is a PRECISE statement that the
// PreToolUse hook did not reach this call (#512) — and unlike the old signal it needs no
// pairing with a missing identity to mean it.
func TestTheWrapperIsItsOwnSourceAndOutranksOnlyTheFlag(t *testing.T) {
	never := func() string { t.Helper(); t.Error("inference ran while an earlier source had a value"); return "" }
	for _, tc := range []struct {
		name    string
		env     string
		wrapper string
		flag    string
		infer   func() string
		wantDir string
		wantVia RunSource
	}{
		{"wrapper alone — the run is right and the hook is absent", "", "/runs/w", "", never, "/runs/w", RunFromWrapper},
		{"the hook outranks the wrapper", "/runs/h", "/runs/h", "", never, "/runs/h", RunFromEnv},
		{"the wrapper outranks a flag that agrees", "", "/runs/w", "/runs/w/", never, "/runs/w", RunFromWrapper},
		{"neither carrier — the flag still works for an operator", "", "", "/runs/f", never, "/runs/f", RunFromFlag},
		{"nothing at all falls to inference", "", "", "", func() string { return "/runs/i" }, "/runs/i", RunFromInference},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(Var, tc.env)
			t.Setenv(VarWrapper, tc.wrapper)
			dir, via, err := ResolveWithSource(tc.flag, tc.infer)
			if err != nil {
				t.Fatalf("ResolveWithSource = %v, want no error", err)
			}
			if dir != tc.wantDir || via != tc.wantVia {
				t.Errorf("ResolveWithSource = (%q, %q), want (%q, %q)", dir, via, tc.wantDir, tc.wantVia)
			}
		})
	}
}

// THE MARKER MOVED UNDER A LIVE RUN — the failure the second carrier exists to catch.
//
// The wrapper is baked at setup and cannot move; Var is re-derived per call from a singleton
// marker that can. They differ exactly when another run claimed the marker mid-run, which is
// how run 3's `assemble` seat came to write a foreign VERIFIED into a different run's record
// and had to be quarantined (#500). Nothing could see it at the time, because there was only
// one carrier and it was the one that had moved.
func TestAMovedMarkerIsRefusedAndNotBlamedOnTheSeat(t *testing.T) {
	t.Setenv(Var, "/runs/other")
	t.Setenv(VarWrapper, "/runs/mine")
	dir, via, err := ResolveWithSource("", nil)
	if err == nil {
		t.Fatalf("ResolveWithSource = (%q, %q), want a refusal", dir, via)
	}
	if via != RunUnresolved {
		t.Errorf("a refused resolve reports source %q, want %q", via, RunUnresolved)
	}
	for _, want := range []string{"/runs/other", "/runs/mine", "You did not cause this", "friction"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q: %v", want, err)
		}
	}
	// It must NOT read as a typo. A seat told to check its own flag looks somewhere it cannot fix.
	if strings.Contains(err.Error(), "--run") {
		t.Errorf("a moved marker is reported as a --run problem, which the seat cannot act on: %v", err)
	}
}

// AND A DISAGREEING --run IS STILL THE SEAT'S TYPO, whichever carrier delivered the dispatch.
// The wrapper must not create a hole where a hand-typed path silently wins.
func TestAFlagDisagreeingWithTheWrapperIsRefused(t *testing.T) {
	t.Setenv(Var, "")
	t.Setenv(VarWrapper, "/runs/mine")
	dir, via, err := ResolveWithSource("/runs/typo", nil)
	if err == nil {
		t.Fatalf("ResolveWithSource = (%q, %q), want a refusal — a typed --run must not outrank the wrapper", dir, via)
	}
	if via != RunUnresolved {
		t.Errorf("a refused resolve reports source %q, want %q", via, RunUnresolved)
	}
	for _, want := range []string{"/runs/typo", "/runs/mine", "omitting --run is correct"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not carry %q: %v", want, err)
		}
	}
}
