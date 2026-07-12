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
