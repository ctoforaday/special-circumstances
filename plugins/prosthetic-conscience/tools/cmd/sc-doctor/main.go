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
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// hookBinaries lists every command under tools/cmd and whether bin/<name> exists.
func hookBinaries(root string) []binStatus {
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
		out = append(out, binStatus{Name: e.Name(), Built: err == nil})
	}
	return out
}

// verdict computes READY / DEGRADED / BLOCKED from tool + binary status.
func verdict(tools []toolchain.Status, bins []binStatus) string {
	degraded := false
	for _, t := range tools {
		if !t.Found {
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
		if t.Found {
			fmt.Fprintf(&sb, "%-18s ✓ %s\n", t.Name, versionOf(t.CheckCmd))
		} else {
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
func fetchRelease(root, name string) error {
	if !toolchain.Present("gh") {
		return fmt.Errorf("gh not on PATH")
	}
	asset := fmt.Sprintf("%s_%s_%s%s", name, runtime.GOOS, runtime.GOARCH, exeSuffix())
	tmp, err := os.MkdirTemp("", "sc-doctor-fetch-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	dl := exec.Command("gh", "release", "download",
		"--repo", releaseRepo, "--pattern", asset, "--pattern", "SHA256SUMS", "--dir", tmp)
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
func fix(root string, bins []binStatus) []string {
	var report []string
	goPresent := toolchain.Present("go")
	for _, b := range bins {
		if b.Built {
			continue
		}
		fetchErr := fetchRelease(root, b.Name)
		if fetchErr == nil {
			report = append(report, fmt.Sprintf("%s: fetched CI-built release asset (checksum verified)", b.Name))
			continue
		}
		if goPresent {
			out := filepath.Join(root, "bin", b.Name+exeSuffix())
			cmd := exec.Command("go", "build", "-C", filepath.Join(root, "tools"), "-o", out, "./cmd/"+b.Name)
			if msg, err := cmd.CombinedOutput(); err != nil {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); BUILD FAILED: %s", b.Name, fetchErr, strings.TrimSpace(string(msg))))
			} else {
				report = append(report, fmt.Sprintf("%s: fetch failed (%v); built from source (dev fallback)", b.Name, fetchErr))
			}
		} else {
			report = append(report, fmt.Sprintf("%s: fetch failed (%v) and Go not found — install gh or Go, or place %s_%s_%s%s into %s manually",
				b.Name, fetchErr, b.Name, runtime.GOOS, runtime.GOARCH, exeSuffix(), filepath.Join(root, "bin")))
		}
	}
	return report
}

func main() {
	doFix := flag.Bool("fix", false, "build/fetch missing hook binaries")
	showVersion := flag.Bool("version", false, "print version and exit")
	rootFlag := flag.String("root", "", "plugin root (default: CLAUDE_PLUGIN_ROOT or the binary's parent dir)")
	flag.Parse()
	if *showVersion {
		fmt.Println("sc-doctor", version)
		return
	}

	root := *rootFlag
	if root == "" {
		root = os.Getenv("CLAUDE_PLUGIN_ROOT")
	}
	if root == "" {
		if exe, err := os.Executable(); err == nil {
			root = filepath.Dir(filepath.Dir(exe)) // bin/sc-doctor -> plugin root
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, "requirements.json"))
	if err != nil {
		fmt.Printf("sc-doctor: cannot read requirements.json under %q: %v\n", root, err)
		os.Exit(0)
	}
	var req requirements
	if err := json.Unmarshal(raw, &req); err != nil {
		fmt.Printf("sc-doctor: malformed requirements.json: %v\n", err)
		os.Exit(0)
	}

	tools := toolchain.Probe(req.Tools)
	bins := hookBinaries(root)

	if *doFix {
		for _, line := range fix(root, bins) {
			fmt.Println(line)
		}
		bins = hookBinaries(root) // re-probe
	}

	fmt.Print(table(tools, bins))

	// Cross-plugin aggregation: every installed SC plugin's own requirements.json,
	// probed and reported here — one preflight for the whole suite.
	for _, pr := range siblingRequirements(root) {
		fmt.Printf("-- %s --\n", pr.Plugin)
		fmt.Print(table(toolchain.Probe(pr.Tools), nil))
	}

	warnings := danceWarnings(root)
	for _, w := range warnings {
		fmt.Println(w)
	}
	v := verdict(tools, bins)
	if v == "READY" && len(warnings) > 0 {
		v = "DEGRADED"
	}
	fmt.Println("VERDICT:", v)
}
