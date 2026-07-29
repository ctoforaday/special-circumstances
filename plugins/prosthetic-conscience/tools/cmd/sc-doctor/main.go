// sc-doctor is the prosthetic-conscience environment preflight (Go binary).
//
// Contract (Design by Contract):
//
//	It MUST produce a deterministic table + verdict (READY / DEGRADED / BLOCKED).
//	Plain run is read-only. `-fix` rebuilds missing hook binaries (go build) when Go
//	is present, else prints the release-asset fetch instructions — it MUST NOT
//	install external tools (that stays consent-gated at the agent layer).
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/toolchain"
)

const version = "0.1.0"

type requirements struct {
	Tools []toolchain.Tool `json:"tools"`
}

type binStatus struct {
	Name  string
	Built bool
	// Plugin/Root/Version identify WHICH plugin owns this binary. The doctor used
	// to know only its own: prosthetic-conscience shipped the only Go binaries, so
	// "the binaries" and "our binaries" were the same set. frank-exchange-of-views
	// now ships feov-record, and a seat that reaches for a record tool nobody
	// installed fails MID-ROUND — the expensive failure this preflight exists to
	// move to setup time.
	Plugin  string
	Root    string
	Version string
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// binariesOf lists every command under a plugin root's tools/cmd and whether
// bin/<name> exists.
func binariesOf(root, plugin, version string) []binStatus {
	entries, err := os.ReadDir(filepath.Join(root, "tools", "cmd"))
	if err != nil {
		return nil
	}
	var out []binStatus
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		bin := filepath.Join(root, "bin", e.Name()+exeSuffix())
		_, err := os.Stat(bin)
		out = append(out, binStatus{
			Name: e.Name(), Built: err == nil,
			Plugin: plugin, Root: root, Version: version,
		})
	}
	return out
}

// hookBinaries lists this plugin's own commands.
func hookBinaries(root string) []binStatus {
	return binariesOf(root, filepath.Base(filepath.Dir(root)), filepath.Base(root))
}

// siblingBinaries walks the marketplace cache for every OTHER installed SC plugin
// that ships a Go module, so `--fix` provisions the whole suite rather than only
// the plugin the doctor happens to live in. Each plugin owns its own bin/ and its
// own release tag; nothing here assumes a shared one.
func siblingBinaries(root string) []binStatus {
	marketDir := filepath.Dir(filepath.Dir(root))
	self := filepath.Base(filepath.Dir(root))
	entries, err := os.ReadDir(marketDir)
	if err != nil {
		return nil
	}
	var out []binStatus
	for _, e := range entries {
		if !e.IsDir() || e.Name() == self {
			continue
		}
		v := newestVersionDir(filepath.Join(marketDir, e.Name()))
		if v == "" {
			continue
		}
		out = append(out, binariesOf(filepath.Join(marketDir, e.Name(), v), e.Name(), v)...)
	}
	return out
}

// verdict computes READY / DEGRADED / BLOCKED from tool + binary status.
func verdict(tools []toolchain.Status, bins []binStatus) string {
	degraded := false
	for _, t := range tools {
		if !t.Found {
			// Absent BY DESIGN is not absent. A cloud session has no gh and is not
			// going to get one; counting it DEGRADED made the verdict permanently
			// wrong there, which is worse than useless — it teaches the operator to
			// discount the verdict on the one axis it exists to be trusted on.
			if t.NotApplicable {
				continue
			}
			if t.Tier == "required" {
				return "BLOCKED"
			}
			if t.Tier == "recommended" {
				degraded = true
			}
		}
	}
	for _, b := range bins {
		if !b.Built {
			degraded = true
		}
	}
	if degraded {
		return "DEGRADED"
	}
	return "READY"
}

// table renders the preflight report.
func table(tools []toolchain.Status, bins []binStatus) string {
	var sb strings.Builder
	for _, t := range tools {
		switch {
		case t.Found:
			fmt.Fprintf(&sb, "%-18s ✓ %s\n", t.Name, versionOf(t.CheckCmd))
		case t.NotApplicable:
			// No install string: printing one here is the false alarm this branch
			// exists to remove.
			fmt.Fprintf(&sb, "%-18s – n/a in %s (%s)\n", t.Name, toolchain.Environment(), t.Purpose)
		default:
			fmt.Fprintf(&sb, "%-18s ✗ (%s) install: %s\n", t.Name, t.Tier, t.Install[runtime.GOOS])
		}
	}
	for _, b := range bins {
		mark := "✓ built"
		if !b.Built {
			mark = "✗ not built (run -fix)"
		}
		fmt.Fprintf(&sb, "%-18s %s\n", b.Name, mark)
	}
	return sb.String()
}

// versionOf best-effort runs a check command and returns its first output line.
func versionOf(checkCmd string) string {
	fields := strings.Fields(checkCmd)
	if len(fields) == 0 {
		return ""
	}
	out, err := exec.Command(fields[0], fields[1:]...).CombinedOutput()
	if err != nil {
		return ""
	}
	if i := strings.IndexByte(string(out), '\n'); i >= 0 {
		return strings.TrimSpace(string(out[:i]))
	}
	return strings.TrimSpace(string(out))
}

// ---- Cross-plugin aggregation + dance-state (efficiency phase PR-C) ----

// pluginReq is one sibling plugin's requirements, tagged with its origin.
type pluginReq struct {
	Plugin string
	Tools  []toolchain.Tool
}

// semverLess compares dotted numeric versions (missing parts = 0).
func semverLess(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, y := 0, 0
		if i < len(pa) {
			fmt.Sscanf(pa[i], "%d", &x)
		}
		if i < len(pb) {
			fmt.Sscanf(pb[i], "%d", &y)
		}
		if x != y {
			return x < y
		}
	}
	return false
}

// newestVersionDir returns the highest-semver subdirectory name, "" if none.
func newestVersionDir(pluginDir string) string {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return ""
	}
	best := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if best == "" || semverLess(best, e.Name()) {
			best = e.Name()
		}
	}
	return best
}

// siblingRequirements walks the marketplace cache (root/../..) and aggregates every OTHER
// installed SC plugin's newest requirements.json — each plugin owns its manifest; the
// doctor reports the whole suite (backlog: doctor cross-plugin aggregation).
func siblingRequirements(root string) []pluginReq {
	marketDir := filepath.Dir(filepath.Dir(root)) // <plugin>/<version> -> cache root
	selfPlugin := filepath.Base(filepath.Dir(root))
	entries, err := os.ReadDir(marketDir)
	if err != nil {
		return nil
	}
	var out []pluginReq
	for _, e := range entries {
		if !e.IsDir() || e.Name() == selfPlugin {
			continue
		}
		v := newestVersionDir(filepath.Join(marketDir, e.Name()))
		if v == "" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(marketDir, e.Name(), v, "requirements.json"))
		if err != nil {
			continue
		}
		var req requirements
		if err := json.Unmarshal(raw, &req); err != nil {
			continue
		}
		out = append(out, pluginReq{Plugin: e.Name(), Tools: req.Tools})
	}
	return out
}

// danceWarnings computes update-dance state from observable cache facts: a newer version
// dir than the one running means the dance is mid-flight; a newest dir with an empty bin/
// is the empty-bin window (the 0.7.0 crash-storm class). Deterministic verdicts, not
// remembered rituals.
func danceWarnings(root string) []string {
	pluginDir := filepath.Dir(root)
	current := filepath.Base(root)
	newest := newestVersionDir(pluginDir)
	var out []string
	if newest != "" && semverLess(current, newest) {
		out = append(out, fmt.Sprintf("DANCE INCOMPLETE: this doctor runs from %s but the cache holds %s — finish the dance (/plugin update -> /reload-plugins -> /reload-skills -> doctor --fix).", current, newest))
	}
	if newest != "" {
		binDir := filepath.Join(pluginDir, newest, "bin")
		entries, err := os.ReadDir(binDir)
		if err != nil || len(entries) == 0 {
			out = append(out, fmt.Sprintf("EMPTY-BIN WINDOW: cache %s has no hook binaries — hooks degrade to guard warnings until doctor --fix runs there.", newest))
		}
	}
	return out
}

// releaseRepo is where CI publishes the cross-compiled hook binaries.
const releaseRepo = "ctoforaday/special-circumstances"

// verifySHA256 checks that sums (the SHA256SUMS file content) records digest for asset.
func verifySHA256(sums, asset, digest string) bool {
	for _, line := range strings.Split(sums, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == asset {
			return strings.EqualFold(fields[0], digest)
		}
	}
	return false
}

// fetchRelease downloads the CI-built asset for this platform, verifies its
// checksum against the release SHA256SUMS, and installs it into bin/.
func fetchRelease(b binStatus) error {
	if !toolchain.Present("gh") {
		return fmt.Errorf("gh not on PATH")
	}
	root, name := b.Root, b.Name
	asset := fmt.Sprintf("%s_%s_%s%s", name, runtime.GOOS, runtime.GOARCH, exeSuffix())
	// The release is PINNED to the owning plugin's tag. Falling back to "latest"
	// would resolve to whichever plugin tagged most recently, whose assets do not
	// contain this binary at all — a download failure that reads like a network
	// problem rather than the version confusion it is.
	tag := ""
	if b.Plugin != "" && b.Version != "" {
		tag = fmt.Sprintf("%s--v%s", b.Plugin, b.Version)
	}
	tmp, err := os.MkdirTemp("", "sc-doctor-fetch-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	args := []string{"release", "download"}
	if tag != "" {
		args = append(args, tag)
	}
	args = append(args, "--repo", releaseRepo, "--pattern", asset, "--pattern", "SHA256SUMS", "--dir", tmp)
	dl := exec.Command("gh", args...)
	if msg, err := dl.CombinedOutput(); err != nil {
		return fmt.Errorf("release download failed: %s", strings.TrimSpace(string(msg)))
	}
	data, err := os.ReadFile(filepath.Join(tmp, asset))
	if err != nil {
		return err
	}
	sums, err := os.ReadFile(filepath.Join(tmp, "SHA256SUMS"))
	if err != nil {
		return err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(data))
	if !verifySHA256(string(sums), asset, digest) {
		return fmt.Errorf("checksum mismatch for %s — refusing to install", asset)
	}
	dst := filepath.Join(root, "bin", name+exeSuffix())
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

// fix provisions every missing binary: CI-built release asset first (checksum-
// verified); a from-source build is only the dev-convenience fallback.
func fix(bins []binStatus) []string {
	var report []string
	goPresent := toolchain.Present("go")
	for _, b := range bins {
		if b.Built {
			continue
		}
		fetchErr := fetchRelease(b)
		if fetchErr == nil {
			report = append(report, fmt.Sprintf("%s (%s): fetched CI-built release asset (checksum verified)", b.Name, b.Plugin))
			continue
		}
		if goPresent {
			out := filepath.Join(b.Root, "bin", b.Name+exeSuffix())
			cmd := exec.Command("go", "build", "-C", filepath.Join(b.Root, "tools"), "-o", out, "./cmd/"+b.Name)
			if msg, err := cmd.CombinedOutput(); err != nil {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); BUILD FAILED: %s", b.Name, fetchErr, strings.TrimSpace(string(msg))))
			} else {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); built from source (dev fallback)", b.Name, fetchErr))
			}
		} else {
			report = append(report, fmt.Sprintf("%s: fetch failed (%v) and Go not found — install gh or Go, or place %s_%s_%s%s into %s manually",
				b.Name, fetchErr, b.Name, runtime.GOOS, runtime.GOARCH, exeSuffix(), filepath.Join(b.Root, "bin")))
		}
	}
	return report
}

// resolveRoot picks the plugin root: the -root flag wins, then CLAUDE_PLUGIN_ROOT,
// then the binary's own grandparent (bin/sc-doctor -> plugin root). executable is a
// parameter so the fallback is reachable in a test.
func resolveRoot(rootFlag, envRoot string, executable func() (string, error)) string {
	if rootFlag != "" {
		return rootFlag
	}
	if envRoot != "" {
		return envRoot
	}
	if exe, err := executable(); err == nil {
		return filepath.Dir(filepath.Dir(exe))
	}
	return ""
}

// run is the doctor with its process boundary passed in. fixFn is a parameter because
// the real one reaches the network and the local Go toolchain; a preflight's REPORTING
// must be testable without either.
func run(args []string, stdout io.Writer, envRoot string, executable func() (string, error), fixFn func([]binStatus) []string) int {
	fs := flag.NewFlagSet("sc-doctor", flag.ContinueOnError)
	fs.SetOutput(stdout)
	doFix := fs.Bool("fix", false, "build/fetch missing hook binaries")
	showVersion := fs.Bool("version", false, "print version and exit")
	rootFlag := fs.String("root", "", "plugin root (default: CLAUDE_PLUGIN_ROOT or the binary's parent dir)")
	if err := fs.Parse(args); err != nil {
		return 0
	}
	if *showVersion {
		fmt.Fprintln(stdout, "sc-doctor", version)
		return 0
	}

	root := resolveRoot(*rootFlag, envRoot, executable)

	raw, err := os.ReadFile(filepath.Join(root, "requirements.json"))
	if err != nil {
		fmt.Fprintf(stdout, "sc-doctor: cannot read requirements.json under %q: %v\n", root, err)
		return 0
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		fmt.Fprintf(stdout, "sc-doctor: malformed requirements.json: %v\n", err)
		return 0
	}

	tools := toolchain.Probe(req.Tools)
	// Own binaries AND every sibling plugin's: the suite is provisioned as a
	// whole, because a seat blocked by another plugin's missing binary fails just
	// as hard as one blocked by ours.
	allBins := func() []binStatus { return append(hookBinaries(root), siblingBinaries(root)...) }
	bins := allBins()

	if *doFix {
		for _, line := range fixFn(bins) {
			fmt.Fprintln(stdout, line)
		}
		bins = allBins() // re-probe
	}

	fmt.Fprint(stdout, table(tools, bins))

	// Cross-plugin aggregation: every installed SC plugin's own requirements.json,
	// probed and reported here — one preflight for the whole suite.
	for _, pr := range siblingRequirements(root) {
		fmt.Fprintf(stdout, "-- %s --\n", pr.Plugin)
		fmt.Fprint(stdout, table(toolchain.Probe(pr.Tools), nil))
	}

	warnings := danceWarnings(root)
	for _, w := range warnings {
		fmt.Fprintln(stdout, w)
	}
	v := verdict(tools, bins)
	if v == "READY" && len(warnings) > 0 {
		v = "DEGRADED"
	}
	fmt.Fprintln(stdout, "VERDICT:", v)
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Getenv("CLAUDE_PLUGIN_ROOT"), os.Executable, fix))
}
