package main

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

func status(name, tier string, found bool) toolchain.Status {
	return toolchain.Status{Tool: toolchain.Tool{Name: name, Tier: tier, Install: map[string]string{"windows": "winget install " + name, "darwin": "brew install " + name, "linux": "apt install " + name}}, Found: found}
}

func TestVerdict(t *testing.T) {
	cases := []struct {
		name  string
		tools []toolchain.Status
		bins  []binStatus
		want  string
	}{
		{"all present", []toolchain.Status{status("git", "required", true)}, []binStatus{{"sc-quality-gate", true}}, "READY"},
		{"required missing", []toolchain.Status{status("git", "required", false)}, nil, "BLOCKED"},
		{"recommended missing", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", false)}, nil, "DEGRADED"},
		{"binary missing", []toolchain.Status{status("git", "required", true)}, []binStatus{{"sc-quality-gate", false}}, "DEGRADED"},
		{"optional missing stays ready", []toolchain.Status{status("jq", "optional", false)}, nil, "READY"},
		{"blocked beats degraded", []toolchain.Status{status("git", "required", false), status("qlty", "recommended", false)}, []binStatus{{"x", false}}, "BLOCKED"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := verdict(c.tools, c.bins); got != c.want {
				t.Fatalf("verdict = %q; want %q", got, c.want)
			}
		})
	}
}

func TestTableShowsInstallForMissing(t *testing.T) {
	out := table([]toolchain.Status{status("qlty", "recommended", false)}, []binStatus{{"sc-quality-gate", false}})
	for _, want := range []string{"qlty", "✗", "install:", "sc-quality-gate", "not built"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q:\n%s", want, out)
		}
	}
}
