package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/shared/buildid"

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
		{"all present", []toolchain.Status{status("git", "required", true)}, []binStatus{{Kind: "go", Name: "sc-quality-gate", Built: true}}, "READY"},
		{"required missing", []toolchain.Status{status("git", "required", false)}, nil, "BLOCKED"},
		{"recommended missing", []toolchain.Status{status("git", "required", true), status("qlty", "recommended", false)}, nil, "DEGRADED"},
		{"binary missing", []toolchain.Status{status("git", "required", true)}, []binStatus{{Kind: "go", Name: "sc-quality-gate", Built: false}}, "DEGRADED"},
		{"optional missing stays ready", []toolchain.Status{status("jq", "optional", false)}, nil, "READY"},
		{"blocked beats degraded", []toolchain.Status{status("git", "required", false), status("qlty", "recommended", false)}, []binStatus{{Kind: "go", Name: "x", Built: false}}, "BLOCKED"},
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
	out := table([]toolchain.Status{status("qlty", "recommended", false)}, []binStatus{{Kind: "go", Name: "sc-quality-gate", Built: false}})
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
	if err := os.WriteFile(filepath.Join(sib, "bin", buildid.ExeName("feov-record")), []byte("x"), 0o755); err != nil {
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
				bins = []binStatus{{Kind: "go", Name: "sc-quality-gate", Built: false}}
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
		Tool:        toolchain.Tool{Name: "go", Tier: "required", MinVersion: "1.25"},
		Found:       true,
		Version:     "1.25.3",
		VersionLine: "go version go1.25.3",
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

// THE TICK'S POSITIVE TWIN, in the same file as the negative above.
//
// That negative is the only thing in this package pinning the tool table's glyph, and a
// negative containment goes vacuous silently: change `%-18s ✓ %s` to spell the mark any other
// way and "a too-old tool still renders a tick" passes forever while checking nothing. The
// other ✓ assertions in this package are on "✓ built", a different renderer on a different
// row, so they would keep passing too.
//
// Pinning it in both directions is the house pattern — toolchainnudge/hook_test.go does
// exactly this with "LIVE" — and it means the pair moves together or fails loudly.
func TestAHealthyToolRendersTheTick(t *testing.T) {
	out := table([]toolchain.Status{{
		Tool:        toolchain.Tool{Name: "go", Tier: "required", MinVersion: "1.25"},
		Found:       true,
		Version:     "1.25.3",
		VersionLine: "go version go1.25.3",
	}}, nil)
	if !strings.Contains(out, "✓") {
		t.Errorf("a healthy tool renders no tick, so the negative above now proves nothing:\n%s", out)
	}
	// VersionLine, not Version: displayVersionOf reads the line the probe captured, and a
	// fixture that sets only Version renders a tick beside an empty column — which is how this
	// test first failed, on its own fixture rather than on the code.
	if !strings.Contains(out, "go") || !strings.Contains(out, "1.25.3") {
		t.Errorf("the healthy row lost its name or version:\n%s", out)
	}
}

// Flat layout sibling discovery: in Antigravity (~/.gemini/config/plugins/<plugin>) and
// dev checkouts (plugins/<plugin>), plugin directories sit directly beside each other
// without version subdirectories.
func TestSiblingBinariesAndRequirementsFlatLayout(t *testing.T) {
	pluginsDir := t.TempDir()
	self := filepath.Join(pluginsDir, "prosthetic-conscience")
	feov := filepath.Join(pluginsDir, "frank-exchange-of-views")
	gray := filepath.Join(pluginsDir, "gray-area")
	nonPlugin := filepath.Join(pluginsDir, "scratch-non-plugin")

	for _, d := range []string{
		filepath.Join(self, "tools", "cmd", "sc-doctor"),
		filepath.Join(feov, "tools", "cmd", "feov-record"),
		filepath.Join(feov, "bin"),
		filepath.Join(gray, "tools", "cmd", "telepathy"),
		filepath.Join(gray, "bin"),
		filepath.Join(gray, ".antigravity-plugin"),
		nonPlugin,
	} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Write manifests and requirements
	if err := os.WriteFile(filepath.Join(self, "plugin.json"), []byte(`{"name":"prosthetic-conscience","version":"0.52.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(feov, "plugin.json"), []byte(`{"name":"frank-exchange-of-views","version":"1.73.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(feov, "requirements.json"), []byte(`{"tools":[{"name":"node","tier":"required","check_cmd":"node --version"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gray, ".antigravity-plugin", "plugin.json"), []byte(`{"name":"gray-area","version":"0.18.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gray, "requirements.json"), []byte(`{"tools":[{"name":"ripgrep","tier":"required","check_cmd":"rg --version"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// 1. Sibling roots should discover feov and gray, ignoring nonPlugin
	roots := siblingPluginRoots(self)
	if len(roots) != 2 {
		t.Fatalf("expected 2 sibling roots, got %d: %+v", len(roots), roots)
	}
	foundFeov, foundGray := false, false
	for _, r := range roots {
		if r.Name == "frank-exchange-of-views" {
			foundFeov = true
			if r.Version != "1.73.0" {
				t.Errorf("feov version = %q, want 1.73.0", r.Version)
			}
			if r.Root != feov {
				t.Errorf("feov root = %q, want %q", r.Root, feov)
			}
		}
		if r.Name == "gray-area" {
			foundGray = true
			if r.Version != "0.18.0" {
				t.Errorf("gray-area version = %q, want 0.18.0", r.Version)
			}
			if r.Root != gray {
				t.Errorf("gray-area root = %q, want %q", r.Root, gray)
			}
		}
	}
	if !foundFeov || !foundGray {
		t.Errorf("missing expected siblings: feov=%v, gray=%v", foundFeov, foundGray)
	}

	// 2. Sibling binaries should list feov-record and telepathy
	bins := siblingBinaries(self)
	if len(bins) != 2 {
		t.Fatalf("expected 2 sibling binaries, got %d: %+v", len(bins), bins)
	}

	// Writing a binary flips Built
	if err := os.WriteFile(filepath.Join(feov, "bin", buildid.ExeName("feov-record")), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	rechecked := siblingBinaries(self)
	for _, b := range rechecked {
		if b.Name == "feov-record" && !b.Built {
			t.Errorf("feov-record reported not built after placing binary")
		}
		if b.Name == "telepathy" && b.Built {
			t.Errorf("telepathy reported built without binary")
		}
	}

	// 3. Sibling requirements should aggregate node and ripgrep
	reqs := siblingRequirements(self)
	if len(reqs) != 2 {
		t.Fatalf("expected 2 sibling requirements, got %d: %+v", len(reqs), reqs)
	}
	toolsFound := map[string]string{}
	for _, pr := range reqs {
		for _, tool := range pr.Tools {
			toolsFound[tool.Name] = pr.Plugin
		}
	}
	if toolsFound["node"] != "frank-exchange-of-views" {
		t.Errorf("node not found from feov: %+v", toolsFound)
	}
	if toolsFound["ripgrep"] != "gray-area" {
		t.Errorf("ripgrep not found from gray-area: %+v", toolsFound)
	}
}

func TestPluginIdentityResolution(t *testing.T) {
	tmp := t.TempDir()

	// 1. Manifest at root
	p1 := filepath.Join(tmp, "p1")
	if err := os.MkdirAll(p1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p1, "plugin.json"), []byte(`{"name":"plug-one","version":"1.2.3"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, v := pluginIdentity(p1); n != "plug-one" || v != "1.2.3" {
		t.Errorf("p1 = (%q, %q), want (plug-one, 1.2.3)", n, v)
	}

	// 2. Manifest in .claude-plugin
	p2 := filepath.Join(tmp, "p2")
	if err := os.MkdirAll(filepath.Join(p2, ".claude-plugin"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p2, ".claude-plugin", "plugin.json"), []byte(`{"name":"plug-two","version":"2.0.0"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, v := pluginIdentity(p2); n != "plug-two" || v != "2.0.0" {
		t.Errorf("p2 = (%q, %q), want (plug-two, 2.0.0)", n, v)
	}

	// 3. Fallback to .fetch/installed tag
	p3 := filepath.Join(tmp, "plug-three")
	if err := os.MkdirAll(filepath.Join(p3, ".fetch"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p3, ".fetch", "installed"), []byte("plug-three--v3.4.5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if n, v := pluginIdentity(p3); n != "plug-three" || v != "3.4.5" {
		t.Errorf("p3 = (%q, %q), want (plug-three, 3.4.5)", n, v)
	}

	// 4. Fallback for versioned directory without manifest
	p4 := filepath.Join(tmp, "plug-four", "4.5.6")
	if err := os.MkdirAll(p4, 0o755); err != nil {
		t.Fatal(err)
	}
	if n, v := pluginIdentity(p4); n != "plug-four" || v != "4.5.6" {
		t.Errorf("p4 = (%q, %q), want (plug-four, 4.5.6)", n, v)
	}
}
