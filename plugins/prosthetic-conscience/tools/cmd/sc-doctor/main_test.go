package main

import (
	"os"
	"path/filepath"
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
		{"all present", []toolchain.Status{status("git", "required", true)}, []binStatus{{Name: "sc-quality-gate", Built: true}}, "READY"},
		{"required missing", []toolchain.Status{status("git", "required", false)}, nil, "BLOCKED"},
		{"recommended missing", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", false)}, nil, "DEGRADED"},
		{"binary missing", []toolchain.Status{status("git", "required", true)}, []binStatus{{Name: "sc-quality-gate", Built: false}}, "DEGRADED"},
		{"optional missing stays ready", []toolchain.Status{status("jq", "optional", false)}, nil, "READY"},
		{"blocked beats degraded", []toolchain.Status{status("git", "required", false), status("qlty", "recommended", false)}, []binStatus{{Name: "x", Built: false}}, "BLOCKED"},
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
	out := table([]toolchain.Status{status("qlty", "recommended", false)}, []binStatus{{Name: "sc-quality-gate", Built: false}})
	for _, want := range []string{"qlty", "✗", "install:", "sc-quality-gate", "not built"} {
		if !strings.Contains(out, want) {
			t.Fatalf("table missing %q:\n%s", want, out)
		}
	}
}

func TestVerifySHA256(t *testing.T) {
	sums := "abc123  sc-quality-gate_windows_amd64.exe\ndef456  *sc-doctor_linux_arm64\n"
	cases := []struct {
		name, asset, digest string
		want                bool
	}{
		{"match", "sc-quality-gate_windows_amd64.exe", "abc123", true},
		{"match case-insensitive", "sc-quality-gate_windows_amd64.exe", "ABC123", true},
		{"binary-mode star prefix", "sc-doctor_linux_arm64", "def456", true},
		{"wrong digest refused", "sc-quality-gate_windows_amd64.exe", "def456", false},
		{"unknown asset refused", "nope.exe", "abc123", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := verifySHA256(sums, c.asset, c.digest); got != c.want {
				t.Fatalf("verifySHA256(%q,%q) = %v; want %v", c.asset, c.digest, got, c.want)
			}
		})
	}
}

// A sibling plugin's binaries are the doctor's business too. prosthetic-conscience
// shipped the only Go binaries for most of this suite's life, so "the binaries"
// and "our binaries" were the same set; frank-exchange-of-views now ships
// feov-record, and a seat reaching for a record tool nobody installed fails
// MID-ROUND — the expensive failure the preflight exists to move earlier.
func TestSiblingBinariesAreDiscoveredAndTagged(t *testing.T) {
	cache := t.TempDir()
	self := filepath.Join(cache, "prosthetic-conscience", "0.9.1")
	sib := filepath.Join(cache, "frank-exchange-of-views", "0.9.10")
	for _, d := range []string{
		filepath.Join(self, "tools", "cmd", "sc-doctor"),
		filepath.Join(sib, "tools", "cmd", "feov-record"),
		filepath.Join(sib, "bin"),
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	got := siblingBinaries(self)
	if len(got) != 1 {
		t.Fatalf("expected the sibling's one binary, got %d: %+v", len(got), got)
	}
	b := got[0]
	if b.Name != "feov-record" || b.Plugin != "frank-exchange-of-views" || b.Version != "0.9.10" {
		t.Errorf("sibling binary mis-identified: %+v", b)
	}
	if b.Built {
		t.Error("an absent binary reported as built")
	}
	// Root must point at the SIBLING, or --fix would build into the wrong plugin.
	if b.Root != sib {
		t.Errorf("root = %s, want %s", b.Root, sib)
	}
	// Writing the binary flips Built, which is what --fix re-probes.
	if err := os.WriteFile(filepath.Join(sib, "bin", "feov-record"+exeSuffix()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if again := siblingBinaries(self); !again[0].Built {
		t.Error("a present binary reported as missing")
	}
}
