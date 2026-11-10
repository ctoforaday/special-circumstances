package doctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// agyInstalledRefFile is the marker written in each Antigravity plugin directory
// recording the exact commit the plugin directory's contents were installed from.
// It matches the marker written by ~/.local/bin/agy-update-plugins (#1796).
const agyInstalledRefFile = ".sc-installed-ref"

type marketplacePin struct {
	Name string
	Path string
	Ref  string
}

type marketplaceManifest struct {
	Plugins []struct {
		Name   string          `json:"name"`
		Source json.RawMessage `json:"source"`
	} `json:"plugins"`
}

type marketplaceSourceObj struct {
	Path string `json:"path"`
	Ref  string `json:"ref"`
}

func parseMarketplacePins(data []byte) ([]marketplacePin, error) {
	var m marketplaceManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	var out []marketplacePin
	for _, p := range m.Plugins {
		if p.Name == "" {
			continue
		}
		var strSource string
		if err := json.Unmarshal(p.Source, &strSource); err == nil {
			relPath := strings.TrimPrefix(strSource, "./")
			out = append(out, marketplacePin{
				Name: p.Name,
				Path: relPath,
				Ref:  "origin/main",
			})
			continue
		}
		var objSource marketplaceSourceObj
		if err := json.Unmarshal(p.Source, &objSource); err == nil && objSource.Path != "" {
			ref := objSource.Ref
			if ref == "" {
				ref = "origin/main"
			}
			out = append(out, marketplacePin{
				Name: p.Name,
				Path: strings.TrimPrefix(objSource.Path, "./"),
				Ref:  ref,
			})
		}
	}
	return out, nil
}

// isAntigravityInstallRoot reports whether root is an Antigravity flat installed plugin
// directory under $HOME/.gemini/config/plugins/<plugin> (or an install root paired with
// an explicit SC_MARKETPLACE_REPO override), as opposed to a Claude versioned cache
// or a git checkout.
func isAntigravityInstallRoot(root string) bool {
	if root == "" || isSemver(filepath.Base(root)) {
		return false
	}
	if HeadCommit(root) != "" {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(root))
	if h := homeDir(); h != "" {
		agyPluginsDir := filepath.ToSlash(filepath.Clean(filepath.Join(h, ".gemini", "config", "plugins"))) + "/"
		if strings.HasPrefix(clean+"/", agyPluginsDir) {
			return true
		}
	}
	if strings.Contains(clean, "/.gemini/config/plugins/") {
		return true
	}
	return isAntigravity() && os.Getenv("SC_MARKETPLACE_REPO") != ""
}

// findMarketplaceRepo locates a local git repository containing .claude-plugin/marketplace.json.
// Checks SC_MARKETPLACE_REPO first (for isolated tests), then the current working directory's
// git root, then ~/.claude/plugins/marketplaces/special-circumstances.
func findMarketplaceRepo() string {
	if envRepo := os.Getenv("SC_MARKETPLACE_REPO"); envRepo != "" {
		return envRepo
	}
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		cmd := exec.Command("git", "-C", cwd, "rev-parse", "--show-toplevel")
		if out, err := cmd.Output(); err == nil {
			top := strings.TrimSpace(string(out))
			if _, err := os.Stat(filepath.Join(top, ".claude-plugin", "marketplace.json")); err == nil {
				return top
			}
		}
	}
	if h := homeDir(); h != "" {
		cand := filepath.Join(h, ".claude", "plugins", "marketplaces", "special-circumstances")
		if _, err := os.Stat(filepath.Join(cand, ".claude-plugin", "marketplace.json")); err == nil {
			return cand
		}
	}
	return ""
}

// readMarketplaceManifest reads .claude-plugin/marketplace.json from origin/main in repo,
// falling back to HEAD or the working tree copy when origin/main is not present (e.g. local test repos).
func readMarketplaceManifest(repo string) ([]byte, error) {
	if repo == "" {
		return nil, fmt.Errorf("no marketplace repository found")
	}
	for _, rev := range []string{"origin/main:.claude-plugin/marketplace.json", "HEAD:.claude-plugin/marketplace.json"} {
		cmd := exec.Command("git", "-C", repo, "show", rev)
		if out, err := cmd.Output(); err == nil && len(out) > 0 {
			return out, nil
		}
	}
	return os.ReadFile(filepath.Join(repo, ".claude-plugin", "marketplace.json"))
}

func resolveRefCommit(repo, ref string) (string, error) {
	candidates := []string{ref + "^{commit}"}
	if ref == "origin/main" {
		candidates = append(candidates, "main^{commit}", "HEAD^{commit}")
	}
	for _, c := range candidates {
		cmd := exec.Command("git", "-C", repo, "rev-parse", "-q", "--verify", c)
		if out, err := cmd.Output(); err == nil {
			if sha := strings.TrimSpace(string(out)); sha != "" {
				return sha, nil
			}
		}
	}
	return "", fmt.Errorf("ref %s not found in %s", ref, repo)
}

func findAgyBin() string {
	if agy := os.Getenv("AGY"); agy != "" {
		return agy
	}
	if p, err := exec.LookPath("agy"); err == nil {
		return p
	}
	if h := homeDir(); h != "" {
		cand := filepath.Join(h, ".local", "bin", "agy")
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return cand
		}
	}
	return "agy"
}

func allInstalledPluginRoots(root string) []pluginRef {
	name, ver := pluginIdentity(root)
	out := []pluginRef{{Name: name, Version: ver, Root: root}}
	out = append(out, siblingPluginRoots(root)...)
	return out
}

// agyContentWarnings checks whether each Antigravity-installed plugin's contents match
// the commit pinned by its marketplace entry (#1796).
//
// It checks two things:
//  1. Whether .sc-installed-ref matches the commit of the tag pinned in .claude-plugin/marketplace.json.
//  2. Whether hooks.json (when present in the tag) matches the tag's hooks.json byte-for-byte,
//     so a hand-copied or tampered file is caught even if .sc-installed-ref was left behind.
var agyContentWarnings = func(root string) []string {
	if !isAntigravityInstallRoot(root) {
		return nil
	}
	repo := findMarketplaceRepo()
	return agyContentWarningsFromRepo(root, repo)
}

func agyContentWarningsFromRepo(root, repo string) []string {
	if !isAntigravityInstallRoot(root) {
		return nil
	}
	manifestBytes, err := readMarketplaceManifest(repo)
	if err != nil {
		return []string{fmt.Sprintf("PLUGIN CONTENTS NOT VERIFIED: cannot read .claude-plugin/marketplace.json (%v) — run `doctor --fix` with a marketplace checkout available.", err)}
	}
	pins, err := parseMarketplacePins(manifestBytes)
	if err != nil {
		return []string{fmt.Sprintf("PLUGIN CONTENTS NOT VERIFIED: malformed .claude-plugin/marketplace.json (%v).", err)}
	}
	pinByName := make(map[string]marketplacePin, len(pins))
	for _, p := range pins {
		pinByName[p.Name] = p
	}

	var warnings []string
	for _, plug := range allInstalledPluginRoots(root) {
		pin, ok := pinByName[plug.Name]
		if !ok {
			continue
		}
		wantCommit, err := resolveRefCommit(repo, pin.Ref)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("PLUGIN CONTENTS NOT VERIFIED: %s pins %s, which could not be resolved (%v).", plug.Name, pin.Ref, err))
			continue
		}
		installedBytes, err := os.ReadFile(filepath.Join(plug.Root, agyInstalledRefFile))
		installedCommit := strings.TrimSpace(string(installedBytes))
		if err != nil || installedCommit == "" {
			warnings = append(warnings, fmt.Sprintf("PLUGIN CONTENTS STALE: %s has no %s marker (pinned %s at %s) — installed contents were not verified against the tag; run `doctor --fix` to reinstall.",
				plug.Name, agyInstalledRefFile, pin.Ref, short(wantCommit)))
			continue
		}
		if installedCommit != wantCommit {
			warnings = append(warnings, fmt.Sprintf("PLUGIN CONTENTS STALE: %s is at %s, marketplace pins %s (%s) — run `doctor --fix` to reinstall.",
				plug.Name, short(installedCommit), pin.Ref, short(wantCommit)))
			continue
		}
		// Even when .sc-installed-ref agrees, verify hooks.json matches the tag if the tag carries one,
		// so a half-updated or hand-edited install is caught by content.
		if tagHooks, err := gitShowFile(repo, wantCommit, pin.Path+"/hooks.json"); err == nil {
			diskHooks, diskErr := os.ReadFile(filepath.Join(plug.Root, "hooks.json"))
			if diskErr != nil || !bytes.Equal(bytes.TrimSpace(tagHooks), bytes.TrimSpace(diskHooks)) {
				warnings = append(warnings, fmt.Sprintf("PLUGIN CONTENTS STALE: %s hooks.json differs from %s (%s) — run `doctor --fix` to reinstall.",
					plug.Name, pin.Ref, short(wantCommit)))
			}
		}
	}
	return warnings
}

func gitShowFile(repo, commit, relPath string) ([]byte, error) {
	cmd := exec.Command("git", "-C", repo, "show", commit+":"+filepath.ToSlash(relPath))
	return cmd.Output()
}

// reinstallAgyPlugins checks all installed Antigravity plugins against the marketplace pins
// and reinstalls any mismatched plugin whole via `git archive` + `agy plugin install <dir>`.
var reinstallAgyPlugins = func(root string) []string {
	if !isAntigravityInstallRoot(root) {
		return nil
	}
	repo := findMarketplaceRepo()
	return reinstallAgyPluginsFromRepo(root, repo, findAgyBin())
}

func reinstallAgyPluginsFromRepo(root, repo, agyBin string) []string {
	if !isAntigravityInstallRoot(root) || repo == "" {
		return nil
	}
	manifestBytes, err := readMarketplaceManifest(repo)
	if err != nil {
		return []string{fmt.Sprintf("agy plugin reinstall skipped: %v", err)}
	}
	pins, err := parseMarketplacePins(manifestBytes)
	if err != nil {
		return []string{fmt.Sprintf("agy plugin reinstall skipped: %v", err)}
	}
	pinByName := make(map[string]marketplacePin, len(pins))
	for _, p := range pins {
		pinByName[p.Name] = p
	}

	var report []string
	for _, plug := range allInstalledPluginRoots(root) {
		pin, ok := pinByName[plug.Name]
		if !ok {
			continue
		}
		wantCommit, err := resolveRefCommit(repo, pin.Ref)
		if err != nil {
			report = append(report, fmt.Sprintf("%s: cannot resolve %s: %v", plug.Name, pin.Ref, err))
			continue
		}
		if agyPluginMatchesCommit(plug.Root, repo, pin.Path, wantCommit) {
			continue
		}
		if err := archiveAndAgyInstall(repo, wantCommit, pin.Path, plug.Root, agyBin); err != nil {
			report = append(report, fmt.Sprintf("%s: reinstall from %s failed: %v", plug.Name, pin.Ref, err))
			continue
		}
		report = append(report, fmt.Sprintf("%s: reinstalled plugin contents from %s (%s)", plug.Name, pin.Ref, short(wantCommit)))
	}
	return report
}

func agyPluginMatchesCommit(plugRoot, repo, relPath, wantCommit string) bool {
	installedBytes, err := os.ReadFile(filepath.Join(plugRoot, agyInstalledRefFile))
	if err != nil || strings.TrimSpace(string(installedBytes)) != wantCommit {
		return false
	}
	if tagHooks, err := gitShowFile(repo, wantCommit, relPath+"/hooks.json"); err == nil {
		diskHooks, diskErr := os.ReadFile(filepath.Join(plugRoot, "hooks.json"))
		if diskErr != nil || !bytes.Equal(bytes.TrimSpace(tagHooks), bytes.TrimSpace(diskHooks)) {
			return false
		}
	}
	return true
}

func archiveAndAgyInstall(repo, commit, relPath, destPluginRoot, agyBin string) error {
	work, err := os.MkdirTemp("", "sc-doctor-agy-install-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	tarPath := filepath.Join(work, "plugin.tar")
	archCmd := exec.Command("git", "-C", repo, "archive", "--format=tar", "-o", tarPath, commit, filepath.ToSlash(relPath))
	if out, err := archCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git archive %s:%s: %s: %w", short(commit), relPath, strings.TrimSpace(string(out)), err)
	}
	srcDir := filepath.Join(work, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		return err
	}
	tarCmd := exec.Command("tar", "-xf", tarPath, "-C", srcDir)
	if out, err := tarCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tar extract: %s: %w", strings.TrimSpace(string(out)), err)
	}
	extractedPluginDir := filepath.Join(srcDir, filepath.FromSlash(relPath))
	instCmd := exec.Command(agyBin, "plugin", "install", extractedPluginDir)
	if h := homeDir(); h != "" {
		instCmd.Dir = h
	}
	if out, err := instCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("agy plugin install: %s: %w", strings.TrimSpace(string(out)), err)
	}
	if err := os.WriteFile(filepath.Join(destPluginRoot, agyInstalledRefFile), []byte(commit+"\n"), 0o644); err != nil {
		return err
	}
	return nil
}
