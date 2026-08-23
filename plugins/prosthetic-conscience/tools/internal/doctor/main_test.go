package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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
	// A SHA256SUMS that says nothing must verify nothing. Both shapes reach this
	// function when a release ships a truncated or malformed sums file, and the answer
	// has to be refusal — "no line matched" must never read as "no objection".
	if verifySHA256("", "sc-doctor_linux_arm64", "def456") {
		t.Error("an empty SHA256SUMS must verify nothing")
	}
	if verifySHA256("garbage-with-no-second-field\n", "sc-doctor_linux_arm64", "def456") {
		t.Error("an unparseable SHA256SUMS must verify nothing")
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

// naStatus is a tool the manifest declared out of scope for this environment.
func naStatus(name, tier string, found bool) toolchain.Status {
	s := status(name, tier, found)
	s.NotApplicable = true
	return s
}

// A cloud session has no gh and is never going to get one. Counting that DEGRADED made
// the verdict permanently wrong there — and a verdict that is always DEGRADED is one the
// operator stops reading, including on the axes where it is telling the truth.
func TestVerdictIgnoresNotApplicableTools(t *testing.T) {
	cases := []struct {
		name  string
		tools []toolchain.Status
		want  string
	}{
		{"n/a recommended does not degrade", []toolchain.Status{status("git", "required", true), naStatus("gh", "recommended", false)}, "READY"},
		{"n/a required does not block", []toolchain.Status{naStatus("gh", "required", false)}, "READY"},
		{"a genuinely missing tool still degrades", []toolchain.Status{naStatus("gh", "recommended", false), status("qlty", "recommended", false)}, "DEGRADED"},
		{"n/a does not rescue a missing binary", []toolchain.Status{naStatus("gh", "recommended", false)}, "DEGRADED"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var bins []binStatus
			if c.name == "n/a does not rescue a missing binary" {
				bins = []binStatus{{Name: "sc-quality-gate", Built: false}}
			}
			if got := verdict(c.tools, bins); got != c.want {
				t.Fatalf("verdict = %q; want %q", got, c.want)
			}
		})
	}
}

// The install string is the false alarm: it sends the operator after a tool the
// environment deliberately does not ship. The row must still appear — silence would
// leave them wondering whether the doctor even knows about gh.
func TestTableRendersNotApplicableWithoutInstallAdvice(t *testing.T) {
	t.Setenv("CLAUDE_CODE_REMOTE", "1")
	s := naStatus("gh", "recommended", false)
	s.Purpose = "GitHub PR / review flow; cloud sessions use MCP tools"
	out := table([]toolchain.Status{s}, nil)
	if !strings.Contains(out, "gh") || !strings.Contains(out, "n/a") {
		t.Fatalf("table must still name gh and mark it n/a:\n%s", out)
	}
	if strings.Contains(out, "install:") {
		t.Fatalf("table offered install advice for a by-design absence:\n%s", out)
	}
	if !strings.Contains(out, "claude-code-remote") {
		t.Fatalf("table must name WHICH environment exempted it:\n%s", out)
	}
	if !strings.Contains(out, "MCP") {
		t.Fatalf("table must carry the purpose text explaining why:\n%s", out)
	}
}

// The exemption lives in the manifest, not in Go. If someone drops the field while
// tidying the JSON, the false alarm returns silently — this is the tripwire.
func TestManifestExemptsGhInCloudSessions(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "requirements.json"))
	if err != nil {
		t.Fatalf("cannot read prosthetic-conscience/requirements.json: %v", err)
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	for _, tool := range req.Tools {
		if tool.Name != "gh" {
			continue
		}
		if tool.AppliesIn(toolchain.EnvRemote) {
			t.Fatal("gh lost its not_applicable_in exemption — every cloud session reports DEGRADED again and advises installing a CLI the environment does not ship")
		}
		if !tool.AppliesIn(toolchain.EnvLocal) {
			t.Fatal("gh must stay a real recommended tool on a local machine")
		}
		return
	}
	t.Fatal("gh is no longer declared in requirements.json")
}

// A REQUIRED TOOL BELOW ITS MINIMUM BLOCKS, exactly as a missing one does — the work it gates
// cannot run either way, and a verdict that said READY over it would be the presence probe
// again, one field along (#521).
func TestATooOldRequiredToolBlocks(t *testing.T) {
	got := verdict([]toolchain.Status{{
		Tool:   toolchain.Tool{Name: "go", Tier: "required", MinVersion: "1.25"},
		Found:  true,
		TooOld: true, Version: "1.22.2",
	}}, nil)
	if got != "BLOCKED" {
		t.Errorf("verdict with go1.22.2 against a 1.25 minimum = %q, want BLOCKED", got)
	}
}

// AND AN UNVERIFIABLE MINIMUM IS NOT READY. It may well be fine; the run cannot say so, and
// "could not tell" belongs on the degraded side of a preflight whose job is to make readiness
// legible before work starts.
func TestAnUnverifiableMinimumIsNotReady(t *testing.T) {
	got := verdict([]toolchain.Status{{
		Tool:              toolchain.Tool{Name: "go", Tier: "required", MinVersion: "1.25"},
		Found:             true,
		VersionUnmeasured: "no version number found in \"go: unknown subcommand\"",
	}}, nil)
	if got == "READY" {
		t.Errorf("verdict with an unverifiable required minimum = READY — the check could not run and said so")
	}
}

// THE ORDINARY CASE IS UNCHANGED. A tool that meets its minimum is just present, and a run
// where everything is present is still READY — a preflight that cried wolf on the healthy path
// would be discounted on the one axis it exists to be trusted on.
func TestAToolThatMeetsItsMinimumIsStillReady(t *testing.T) {
	got := verdict([]toolchain.Status{{
		Tool:    toolchain.Tool{Name: "go", Tier: "required", MinVersion: "1.25"},
		Found:   true,
		Version: "1.25.3",
	}}, nil)
	if got != "READY" {
		t.Errorf("verdict with go1.25.3 against a 1.25 minimum = %q, want READY", got)
	}
}

// THE ROW SAYS UPGRADE, NOT INSTALL. An operator sent to install what they are demonstrably
// holding looks in the wrong place — the same failure shape as telling a seat to pass a flag it
// already passed.
func TestTheTooOldRowNamesTheRightRemedy(t *testing.T) {
	out := table([]toolchain.Status{{
		Tool: toolchain.Tool{
			Name: "go", Tier: "required", MinVersion: "1.25",
			Install: map[string]string{runtime.GOOS: "see https://go.dev/dl/"},
		},
		Found: true, TooOld: true, Version: "1.22.2",
	}}, nil)
	for _, want := range []string{"go", "1.22.2", "below the minimum 1.25", "upgrade:", "https://go.dev/dl/"} {
		if !strings.Contains(out, want) {
			t.Errorf("the too-old row does not carry %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "✓") {
		t.Errorf("a too-old tool still renders a tick:\n%s", out)
	}
}
