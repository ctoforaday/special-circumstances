package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

func status(name, tier string, found bool) toolchain.Status {
	return toolchain.Status{Tool: toolchain.Tool{Name: name, Tier: tier}, Found: found}
}

func TestNudge(t *testing.T) {
	cases := []struct {
		name     string
		statuses []toolchain.Status
		want     []string // substrings; empty slice = expect silence
	}{
		{"all present is silent", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", true)}, nil},
		{"missing recommended named", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", false)}, []string{"qlty", "doctor"}},
		{"missing optional is silent", []toolchain.Status{status("jq", "optional", false)}, nil},
		{"multiple missing listed", []toolchain.Status{status("gh", "recommended", false), status("qlty", "recommended", false)}, []string{"gh", "qlty"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := nudge(c.statuses)
			if len(c.want) == 0 {
				if got != "" {
					t.Fatalf("want silence, got %q", got)
				}
				return
			}
			for _, w := range c.want {
				if !strings.Contains(got, w) {
					t.Fatalf("nudge() = %q; missing %q", got, w)
				}
			}
			if strings.Count(got, "\n") > 0 {
				t.Fatalf("nudge must be one line: %q", got)
			}
		})
	}
}

// SessionStart is the most expensive place in the suite for a warning nobody can act
// on: it is the first thing every session shows, every time. A cloud session ships no
// gh by design, so nudging about it there is pure noise — and noise at line one is how
// an operator learns to skip the line that eventually matters.
func TestNudgeSkipsNotApplicableTools(t *testing.T) {
	na := func(name, tier string) toolchain.Status {
		s := status(name, tier, false)
		s.NotApplicable = true
		return s
	}
	if got := nudge([]toolchain.Status{status("git", "required", true), na("gh", "recommended")}); got != "" {
		t.Fatalf("want silence for a by-design absence, got %q", got)
	}
	got := nudge([]toolchain.Status{na("gh", "recommended"), status("qlty", "recommended", false)})
	if !strings.Contains(got, "qlty") {
		t.Fatalf("a genuinely missing tool must still be named: %q", got)
	}
	if strings.Contains(got, "gh") {
		t.Fatalf("the exempt tool leaked into the nudge: %q", got)
	}
}
