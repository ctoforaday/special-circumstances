package modeltier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOfPicksTheTierBySubstring(t *testing.T) {
	for model, want := range map[string]string{
		"claude-fable-5":       "fable",
		"claude-opus-4-8":      "opus",
		"claude-sonnet-5":      "sonnet",
		"claude-haiku-4-5":     "haiku",
		"CLAUDE-OPUS-4-8":      "opus",
		"something-unheard-of": "fable",
		"":                     "fable",
	} {
		if got := Of(model); got != want {
			t.Errorf("Of(%q) = %q, want %q", model, got, want)
		}
	}
}

// Of flattens two different states onto `fable` — an unknown model and no model at all — and
// Recognized is the only thing that tells them from a real fable.
func TestRecognizedSeparatesAMeasurementFromTheFallback(t *testing.T) {
	if !Recognized("claude-fable-5") {
		t.Error("a real fable is recognized")
	}
	for _, m := range []string{"", "gpt-9", "claude-next"} {
		if Recognized(m) {
			t.Errorf("%q must not read as a recognized tier even though Of() answers fable", m)
		}
	}
}

// THE LADDER AND THE RANK ARE THE SAME LIST, which is the property that used to be an unstated
// coincidence between Order and the price table.
func TestDearerFollowsTheLadder(t *testing.T) {
	for i := 1; i < len(Order); i++ {
		if !Dearer(Order[i], Order[i-1]) {
			t.Errorf("%s must be dearer than %s", Order[i], Order[i-1])
		}
	}
	if Dearer("opus", "opus") {
		t.Error("a tier is not dearer than itself")
	}
	// A tier off the ladder ranks dearest, matching Of's over-report-never-under-report fallback.
	if !Dearer("unheard-of", "fable") {
		t.Error("an unknown tier must rank dearest")
	}
}

func TestConfigReadsTheRunsDeclaredTiers(t *testing.T) {
	run := t.TempDir()
	if err := os.MkdirAll(filepath.Join(run, "inputs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(run, "inputs", "run-config.json"),
		[]byte(`{"model":"claude-fable-5","judgmentModel":"claude-sonnet-5"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	m, j := Config(run)
	if m != "claude-fable-5" || j != "claude-sonnet-5" {
		t.Fatalf("got %q/%q", m, j)
	}
	// An absent or unparseable config yields two empty strings, and the caller decides what that
	// means — here it means "declared no tier", never "declared haiku".
	if m, j := Config(t.TempDir()); m != "" || j != "" {
		t.Errorf("absent config: got %q/%q", m, j)
	}
}
