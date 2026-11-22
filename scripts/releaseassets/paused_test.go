package main

import (
	"slices"
	"strings"
	"testing"
)

// TestAPausedHelperBuildsNoLeg: a helper the paused list names leaves the helper set, the others
// stay in order, and a name that is no helper is refused rather than pausing nothing.
func TestAPausedHelperBuildsNoLeg(t *testing.T) {
	helpers := []string{"feov-embed", "feov-ocr"}
	kept, paused, err := Unpaused(helpers, " feov-embed ")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(kept, []string{"feov-ocr"}) || !slices.Equal(paused, []string{"feov-embed"}) {
		t.Fatalf("Unpaused = %v, %v; want [feov-ocr], [feov-embed]", kept, paused)
	}
	kept, paused, err = Unpaused(helpers, "")
	if err != nil || !slices.Equal(kept, helpers) || len(paused) != 0 {
		t.Fatalf("an empty list = %v, %v, %v; want every helper kept", kept, paused, err)
	}
	if _, _, err := Unpaused(helpers, "feov-embd"); err == nil || !strings.Contains(err.Error(), `"feov-embd"`) {
		t.Fatalf("a misspelt name = %v; want it refused by name", err)
	}
	// The pause reaches the matrix: DecideScope given the kept set builds no feov-embed leg.
	targets := []Target{{Name: "linux-amd64", OS: "linux", Arch: "amd64", Runner: "ubuntu-24.04"}}
	kept, _, _ = Unpaused(helpers, "feov-embed")
	scope, err := DecideScope("schedule", "refs/heads/main", nil, NativeSet{}, targets, nil, kept)
	if err != nil {
		t.Fatal(err)
	}
	for _, leg := range scope.Matrix.Include {
		if leg.Helper == "feov-embed" {
			t.Fatalf("the matrix builds a paused helper: %+v", leg)
		}
	}
}
