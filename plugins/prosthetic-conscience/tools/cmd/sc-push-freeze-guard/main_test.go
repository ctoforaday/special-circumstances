package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/runlive"
)

func TestDecide(t *testing.T) {
	live := &runlive.Marker{RunDir: "research/x", PinnedPaths: []string{"ideas/backlog.md", "research/old"}, Started: "2026-07-16T00:00:00Z"}
	cases := []struct {
		name    string
		m       *runlive.Marker
		cmd     string
		wantHit bool
	}{
		{"no marker is silent even on push", nil, "git push origin main", false},
		{"live + push warns", live, "cd repo && git push", true},
		{"live + non-push is silent", live, "git status", false},
		{"live + commit without push is silent", live, "git commit -m x", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decide(c.m, c.cmd)
			if (got != "") != c.wantHit {
				t.Fatalf("decide(%v, %q) = %q; wantHit=%v", c.m != nil, c.cmd, got, c.wantHit)
			}
			if c.wantHit {
				for _, want := range []string{"research/x", "ideas/backlog.md", "FROZEN"} {
					if !strings.Contains(got, want) {
						t.Fatalf("warning missing %q: %q", want, got)
					}
				}
			}
		})
	}
}
