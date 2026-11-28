package fuzz

import (
	"strings"
	"testing"
)

// RED'S ROAD KEEPS THE COST SENTENCE'S PHRASING. A ruling for red applies its text over the words it
// quotes; text without the phrasing the edit drives swap leaves every later gap of the run nothing to
// propose or counter-edit over (redRoadText).
func TestRedsRoadKeepsTheSwappedPhrasing(t *testing.T) {
	for _, c := range []struct{ words, want string }{
		{"The cost is rising over time<!--gap:G-1-->", "rising over time"},
		{"The cost is climbing sharply", "climbing sharply"},
		{"A § fuzz sentence to anchor findings", ""},
	} {
		got := redRoadText("G-1", c.words)
		if c.want != "" && !strings.Contains(got, c.want) {
			t.Errorf("red's text over %q is %q, which drops %q", c.words, got, c.want)
		}
		if c.want == "" && got != "fuzz: red's text for G-1" {
			t.Errorf("red's text over %q is %q, want its own text alone", c.words, got)
		}
	}
}
