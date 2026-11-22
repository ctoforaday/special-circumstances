package main

// THE PARITY CHECKS ABOVE READ FILES; THESE ASK THE HARNESSES.
//
// pluginparity proves the manifests agree with each other. That is not the same claim as "Claude
// Code and Antigravity load all four plugins, every skill and every agent" — a manifest can agree
// with its twin and still be skipped by the loader that reads it (Antigravity drops an agent whose
// `tools:` is a scalar). So these tests
// stage the checkout into a temp tree, point the real `claude` and `agy` binaries at it under an
// isolated HOME, and compare what each harness reports loading with what is on disk.
//
// Each harness test also runs a negative control: the defect it guards, planted in a copy of the
// stage, must change what the harness reports. A check that cannot see the defect is not a check.
//
// CI installs both binaries in the `scripts` job; locally, a missing binary skips the test, naming
// the binary, while on CI (`CI=true`) a missing binary fails the test outright.

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// harnessCLI resolves a harness binary from PATH, then from ~/.local/bin, where both installers
// put it and where a non-interactive shell's PATH often does not reach.
func harnessCLI(t *testing.T, name string) string {
	t.Helper()
	if testing.Short() {
		t.Skipf("-short: not invoking the %s binary", name)
	}
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		p := filepath.Join(home, ".local", "bin", name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	if os.Getenv("CI") == "true" {
		t.Fatalf("%s is not installed (PATH or ~/.local/bin) on CI: real harness validation must run on every PR", name)
	}
	t.Skipf("%s is not installed (PATH or ~/.local/bin): its load test did not run", name)
	return ""
}

// stagedPlugin is what the checkout says one plugin carries.
type stagedPlugin struct {
	name, version  string
	skills, agents []string
	hooks          int
}

// stage copies the files a harness loads — the marketplace and every plugin minus its
// Go tools tree — into a temp directory, and rewrites each marketplace source to its relative path
// so both harnesses install THIS checkout rather than the release the manifest pins.
func stage(t *testing.T) (string, []stagedPlugin) {
	t.Helper()
	root, err := gitx.Root()
	if err != nil {
		t.Fatal(err)
	}
	out, err := gitx.Run(root, "ls-files", "-co", "--exclude-standard", "--", ".claude-plugin", "plugins")
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	copied := 0
	for _, rel := range gitx.Lines(out) {
		if strings.Contains(rel, "/tools/") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
		if os.IsNotExist(err) {
			continue // deleted in the worktree, still in the index
		}
		if err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(dst, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, b, 0o644); err != nil {
			t.Fatal(err)
		}
		copied++
	}
	if copied == 0 {
		t.Fatal("staged zero files — the listing is wrong, not the tree")
	}

	mpath := filepath.Join(dst, ".claude-plugin", "marketplace.json")
	mb, err := os.ReadFile(mpath)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(mb, &m); err != nil {
		t.Fatal(err)
	}
	entries, _ := m["plugins"].([]any)
	if len(entries) == 0 {
		t.Fatal("marketplace.json lists no plugins")
	}
	var plugins []stagedPlugin
	for _, e := range entries {
		entry := e.(map[string]any)
		name := entry["name"].(string)
		entry["source"] = "./plugins/" + name
		plugins = append(plugins, inventory(t, dst, name))
	}
	mb, err = json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mpath, mb, 0o644); err != nil {
		t.Fatal(err)
	}
	return dst, plugins
}

// inventory reads one staged plugin's version, skill directories and agent files.
func inventory(t *testing.T, dst, name string) stagedPlugin {
	t.Helper()
	dir := filepath.Join(dst, "plugins", name)
	b, err := os.ReadFile(filepath.Join(dir, ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var pj struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &pj); err != nil {
		t.Fatal(err)
	}
	sp := stagedPlugin{name: name, version: pj.Version}
	skills, _ := filepath.Glob(filepath.Join(dir, "skills", "*", "SKILL.md"))
	for _, s := range skills {
		sp.skills = append(sp.skills, filepath.Base(filepath.Dir(s)))
	}
	agents, _ := filepath.Glob(filepath.Join(dir, "agents", "*.md"))
	for _, a := range agents {
		sp.agents = append(sp.agents, strings.TrimSuffix(filepath.Base(a), ".md"))
	}
	sort.Strings(sp.skills)
	sort.Strings(sp.agents)
	if _, err := os.Stat(filepath.Join(dir, "hooks.json")); err == nil {
		sp.hooks = 1
	}
	return sp
}

// copyTree duplicates a staged tree so a negative control can break its copy.
func copyTree(t *testing.T, src string) string {
	t.Helper()
	dst := t.TempDir()
	err := filepath.Walk(src, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if fi.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

// isolatedEnv is the process environment with every config, cache and data root pointed at home,
// so a harness reads and writes nothing of the developer's own installation.
func isolatedEnv(home string) []string {
	var env []string
	for _, kv := range os.Environ() {
		k, _, _ := strings.Cut(kv, "=")
		switch {
		case k == "HOME", k == "USERPROFILE", k == "CLAUDE_CONFIG_DIR",
			strings.HasPrefix(k, "XDG_"), strings.HasPrefix(k, "FEOV_"):
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"HOME="+home,
		"USERPROFILE="+home,
		"CLAUDE_CONFIG_DIR="+filepath.Join(home, ".claude"),
		"XDG_CONFIG_HOME="+filepath.Join(home, ".config"),
		"XDG_CACHE_HOME="+filepath.Join(home, ".cache"),
		"XDG_DATA_HOME="+filepath.Join(home, ".local", "share"),
		"CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1",
		// Each client's own updater switch. A client's updater replaces the running executable
		// wherever it lives, whatever HOME says: `agy agents` alone swaps the developer's agy for
		// the latest build. agy honours its variable only at the exact value "true".
		"DISABLE_AUTOUPDATER=1",
		"AGY_CLI_DISABLE_AUTO_UPDATE=true",
	)
}

// tempHome creates an isolated HOME directory and retries removal on cleanup so a harness child
// process (such as Antigravity's language server) flushing its logs as it exits does not fail
// t.TempDir()'s RemoveAll. A home that still will not go fails the test.
//
// A DETACHED UPDATER IN THE HOME FAILS THE TEST. agy without its updater switch starts
// `agy --bg-updater` in the background; it outlives the test, writes
// .gemini/antigravity-cli/updater/ after the removal, and the directory is left in TMPDIR — 56 of
// them on 2026-11-22. Its lock file is there by the time the command that started it exits, so
// the check at cleanup sees what the removal would otherwise race.
func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Cleanup(func() {
		if updaterStarted(home) {
			t.Errorf("agy started its background updater in the test HOME %s; it outlives the test and "+
				"re-creates the directory after removal — isolatedEnv must switch it off", home)
		}
		var err error
		for range 20 {
			if err = os.RemoveAll(home); err == nil {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
		t.Errorf("the test HOME survived its removal: %v", err)
	})
	return home
}

// updaterStarted reports whether agy's background updater has run under home.
func updaterStarted(home string) bool {
	_, err := os.Stat(filepath.Join(home, ".gemini", "antigravity-cli", "updater"))
	return err == nil
}

// run executes a harness command and returns its stdout; a non-zero exit fails the test with
// both streams.
func run(t *testing.T, home, bin string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = isolatedEnv(home)
	cmd.Dir = home
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s: %v\nstdout:\n%s\nstderr:\n%s", filepath.Base(bin), strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

// lines returns the sorted non-empty lines of a listing.
func lines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	sort.Strings(out)
	return out
}

func sameSet(a, b []string) bool {
	x, y := append([]string(nil), a...), append([]string(nil), b...)
	sort.Strings(x)
	sort.Strings(y)
	return strings.Join(x, "\x00") == strings.Join(y, "\x00")
}

func allAgents(plugins []stagedPlugin) []string {
	var out []string
	for _, p := range plugins {
		out = append(out, p.agents...)
	}
	sort.Strings(out)
	return out
}

// detailsLine matches one row of `claude plugin details`' component inventory.
var detailsLine = regexp.MustCompile(`^\s*(Skills|Agents) \((\d+)\)\s*(.*)$`)

// claudeValidateAgents parses the agent names `claude plugin validate --strict` reports under
// "Provides: ... • agents: <plugin>:<name>, ...". Claude Code's `plugin details` only counts the
// default agents/ directory and prints Agents (0) when `.claude-plugin/plugin.json` declares an
// explicit `"agents"` array (such as `./agents-claude/*.md`), whereas `plugin validate --strict`
// resolves and lists every declared agent.
func claudeValidateAgents(out, pluginName string) []string {
	ansi := regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	clean := ansi.ReplaceAllString(out, "")
	clean = strings.ReplaceAll(clean, "\r", "")
	prefix := pluginName + ":"
	re := regexp.MustCompile(`agents:\s*([^\n]+)`)
	var names []string
	for _, m := range re.FindAllStringSubmatch(clean, -1) {
		for _, item := range strings.Split(m[1], ",") {
			item = strings.TrimSpace(item)
			if strings.HasPrefix(item, prefix) {
				if name := strings.TrimPrefix(item, prefix); name != "" {
					names = append(names, name)
				}
			}
		}
	}
	sort.Strings(names)
	return names
}

// claudeInventory parses the skill and agent names `claude plugin details` reports.
func claudeInventory(t *testing.T, out string) (skills, agents []string) {
	t.Helper()
	seen := map[string]bool{}
	for _, l := range strings.Split(out, "\n") {
		m := detailsLine.FindStringSubmatch(strings.TrimRight(l, "\r"))
		if m == nil {
			continue
		}
		seen[m[1]] = true
		var names []string
		for _, n := range strings.Split(m[3], ",") {
			if n = strings.TrimSpace(n); n != "" {
				names = append(names, n)
			}
		}
		if m[1] == "Skills" {
			skills = names
		} else {
			agents = names
		}
	}
	if !seen["Skills"] || !seen["Agents"] {
		t.Fatalf("`claude plugin details` printed no Skills/Agents inventory rows — the format moved, so this test read nothing:\n%s", out)
	}
	return skills, agents
}

// Claude Code: the marketplace and every plugin validate strictly, install from THIS checkout at
// the manifest's version, and load every skill and agent on disk.
func TestClaudeLoadsMarketplaceAndAllFourPlugins(t *testing.T) {
	claude := harnessCLI(t, "claude")
	repo, plugins := stage(t)
	home := tempHome(t)

	run(t, home, claude, "plugin", "validate", "--strict", repo)
	validatedAgents := map[string][]string{}
	for _, p := range plugins {
		vout := run(t, home, claude, "plugin", "validate", "--strict", filepath.Join(repo, "plugins", p.name))
		validatedAgents[p.name] = claudeValidateAgents(vout, p.name)
	}

	run(t, home, claude, "plugin", "marketplace", "add", repo)
	for _, p := range plugins {
		run(t, home, claude, "plugin", "install", p.name+"@special-circumstances")
	}

	var listed []struct {
		ID             string `json:"id"`
		Version        string `json:"version"`
		Enabled        bool   `json:"enabled"`
		ReadFromFolder string `json:"readFromFolder"`
	}
	if err := json.Unmarshal([]byte(run(t, home, claude, "plugin", "list", "--json")), &listed); err != nil {
		t.Fatal(err)
	}
	byID := map[string]int{}
	for i, l := range listed {
		byID[l.ID] = i
	}
	for _, p := range plugins {
		id := p.name + "@special-circumstances"
		i, ok := byID[id]
		if !ok {
			t.Errorf("claude plugin list does not show %s", id)
			continue
		}
		l := listed[i]
		if !l.Enabled || l.Version != p.version {
			t.Errorf("%s: enabled=%v version=%s, want enabled at %s", id, l.Enabled, l.Version, p.version)
		}
		if want := filepath.Join(repo, "plugins", p.name); filepath.Clean(l.ReadFromFolder) != want {
			t.Errorf("%s was read from %q, not the staged checkout %q — the test measured a published release", id, l.ReadFromFolder, want)
		}

		skills, _ := claudeInventory(t, run(t, home, claude, "plugin", "details", id))
		if !sameSet(skills, p.skills) {
			t.Errorf("%s: Claude loaded skills %v, disk has %v", p.name, skills, p.skills)
		}
		if !sameSet(validatedAgents[p.name], p.agents) {
			t.Errorf("%s: Claude validated agents %v, disk has %v", p.name, validatedAgents[p.name], p.agents)
		}
	}

	t.Run("strict validate rejects broken agents entry", func(t *testing.T) {
		broken := copyTree(t, repo)
		manifest := filepath.Join(broken, "plugins", "prosthetic-conscience", ".claude-plugin", "plugin.json")
		b, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatal(err)
		}
		mutated := bytes.Replace(b, []byte("./agents-claude/probe.md"), []byte("./agents-claude/does-not-exist.md"), 1)
		if bytes.Equal(b, mutated) {
			t.Fatalf("%s has no ./agents-claude/probe.md entry to break", manifest)
		}
		if err := os.WriteFile(manifest, mutated, 0o644); err != nil {
			t.Fatal(err)
		}
		runMustFail(t, home, claude, "plugin", "validate", "--strict", filepath.Join(broken, "plugins", "prosthetic-conscience"))
	})
}

type agyFinding struct {
	Severity string `json:"severity"`
	File     string `json:"file"`
	Message  string `json:"message"`
}

// agyTolerated is the only finding a plugin may carry in this step: the logo note is
// informational. Anything else — a manifest key, a hook warning, a skipped agent, a skill — fails.
func agyTolerated(f agyFinding) bool {
	return f.Severity == "note" && f.File == ".antigravity-plugin/plugin.json" && strings.HasPrefix(f.Message, "no logo")
}

// runMustFail executes a harness command that must exit non-zero and returns its stdout.
func runMustFail(t *testing.T, home, bin string, args ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = isolatedEnv(home)
	cmd.Dir = home
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err == nil {
		t.Fatalf("%s %s succeeded, want non-zero exit\nstdout:\n%s\nstderr:\n%s", filepath.Base(bin), strings.Join(args, " "), stdout.String(), stderr.String())
	}
	return stdout.String()
}

// Antigravity: every plugin validates strictly against .antigravity-plugin/plugin.json and root
// hooks.json, installing each plugin directory enables all four at the manifest's version, and every
// skill and agent on disk loads. Antigravity runs installed copies only: no workspace registration
// loads a checkout's plugins (#1792), so every subtest installs into an isolated HOME.
func TestAntigravityValidatesInstallsAndDiscoversAllFourPlugins(t *testing.T) {
	agy := harnessCLI(t, "agy")
	repo, plugins := stage(t)
	want := allAgents(plugins)

	t.Run("validate", func(t *testing.T) {
		home := tempHome(t)
		for _, p := range plugins {
			var v struct {
				Manifest   string `json:"manifest"`
				Name       string `json:"name"`
				Components struct {
					Agents int `json:"agents"`
					Hooks  int `json:"hooks"`
					Skills int `json:"skills"`
				} `json:"components"`
				Findings []agyFinding `json:"findings"`
			}
			if err := json.Unmarshal([]byte(run(t, home, agy, "plugin", "validate", "--strict", "--json", filepath.Join(repo, "plugins", p.name))), &v); err != nil {
				t.Fatal(err)
			}
			if v.Manifest != ".antigravity-plugin/plugin.json" || v.Name != p.name {
				t.Errorf("%s: agy read manifest %q naming %q, want .antigravity-plugin/plugin.json naming %q", p.name, v.Manifest, v.Name, p.name)
			}
			if v.Components.Agents != len(p.agents) || v.Components.Skills != len(p.skills) || v.Components.Hooks != p.hooks {
				t.Errorf("%s: agy counts %d agents / %d skills / %d hooks, disk has %d / %d / %d", p.name, v.Components.Agents, v.Components.Skills, v.Components.Hooks, len(p.agents), len(p.skills), p.hooks)
			}
			for _, f := range v.Findings {
				if !agyTolerated(f) {
					t.Errorf("%s: agy finding %s %s: %s", p.name, f.Severity, f.File, f.Message)
				}
			}
		}
	})

	t.Run("strict validate rejects hook defects", func(t *testing.T) {
		home := tempHome(t)
		cases := []struct {
			name   string
			mutate func(t *testing.T, pluginDir string)
			file   string
		}{
			{
				name: "PLUGIN_ROOT in hooks.json command",
				mutate: func(t *testing.T, pluginDir string) {
					p := filepath.Join(pluginDir, "hooks.json")
					b, err := os.ReadFile(p)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(p, bytes.Replace(b, []byte("${PWD}"), []byte("${PLUGIN_ROOT}"), 1), 0o644); err != nil {
						t.Fatal(err)
					}
				},
				file: "hooks.json",
			},
			{
				name: "Claude tool name in hooks.json matcher",
				mutate: func(t *testing.T, pluginDir string) {
					p := filepath.Join(pluginDir, "hooks.json")
					b, err := os.ReadFile(p)
					if err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(p, bytes.Replace(b, []byte("run_command"), []byte("Bash"), 1), 0o644); err != nil {
						t.Fatal(err)
					}
				},
				file: "hooks.json",
			},
			{
				name: "legacy hooks/hooks.json without root hooks.json",
				mutate: func(t *testing.T, pluginDir string) {
					if err := os.Remove(filepath.Join(pluginDir, "hooks.json")); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(filepath.Join(pluginDir, "hooks", "hooks.json"), []byte("{}"), 0o644); err != nil {
						t.Fatal(err)
					}
				},
				file: "hooks/hooks.json",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				broken := copyTree(t, repo)
				pluginDir := filepath.Join(broken, "plugins", "prosthetic-conscience")
				tc.mutate(t, pluginDir)
				out := runMustFail(t, home, agy, "plugin", "validate", "--strict", "--json", pluginDir)
				var v struct {
					Findings []agyFinding `json:"findings"`
				}
				if err := json.Unmarshal([]byte(out), &v); err != nil {
					t.Fatalf("unmarshal validate output: %v\n%s", err, out)
				}
				var matched bool
				for _, f := range v.Findings {
					if f.File == tc.file && !agyTolerated(f) {
						matched = true
						break
					}
				}
				if !matched {
					t.Errorf("expected non-tolerated finding on %s, got %+v", tc.file, v.Findings)
				}
			})
		}
	})

	t.Run("a scalar tools field drops the agent", func(t *testing.T) {
		// Negative control for the frontmatter gate: plant the scalar form in one agent and the
		// harness must stop listing exactly that agent.
		broken := copyTree(t, repo)
		probe := filepath.Join(broken, "plugins", "prosthetic-conscience", "agents", "probe.md")
		b, err := os.ReadFile(probe)
		if err != nil {
			t.Fatal(err)
		}
		re := regexp.MustCompile(`(?m)^tools: \[(.*)\]$`)
		if !re.Match(b) {
			t.Fatalf("%s has no `tools: [...]` line to break", probe)
		}
		if err := os.WriteFile(probe, re.ReplaceAll(b, []byte("tools: $1")), 0o644); err != nil {
			t.Fatal(err)
		}
		home := tempHome(t)
		installEach(t, home, agy, broken, plugins)
		got := lines(run(t, home, agy, "agents"))
		for _, a := range got {
			if a == "probe" {
				t.Fatalf("agy still loads probe with a scalar tools field — the frontmatter rule guards nothing: %v", got)
			}
		}
		if len(got) != len(want)-1 {
			t.Errorf("agy listed %d agents with one broken, want %d: %v", len(got), len(want)-1, got)
		}
	})

	t.Run("install each plugin", func(t *testing.T) {
		home := tempHome(t)
		installEach(t, home, agy, repo, plugins)
		var listed []struct {
			Name    string   `json:"name"`
			Enabled bool     `json:"enabled"`
			Version string   `json:"version"`
			Source  string   `json:"source"`
			Skills  []string `json:"skills"`
			Agents  []string `json:"agents"`
		}
		if err := json.Unmarshal([]byte(run(t, home, agy, "plugin", "list", "--json")), &listed); err != nil {
			t.Fatal(err)
		}
		byName := map[string]int{}
		for i, l := range listed {
			byName[l.Name] = i
		}
		for _, p := range plugins {
			i, ok := byName[p.name]
			if !ok {
				t.Errorf("agy plugin list does not show %s", p.name)
				continue
			}
			l := listed[i]
			if !l.Enabled || l.Source != "antigravity" || l.Version != p.version {
				t.Errorf("%s: enabled=%v source=%q version=%s, want enabled antigravity at %s", p.name, l.Enabled, l.Source, l.Version, p.version)
			}
			if !sameSet(l.Skills, p.skills) {
				t.Errorf("%s: agy loaded skills %v, disk has %v", p.name, l.Skills, p.skills)
			}
			if !sameSet(l.Agents, p.agents) {
				t.Errorf("%s: agy loaded agents %v, disk has %v", p.name, l.Agents, p.agents)
			}
		}
		if len(listed) != len(plugins) {
			t.Errorf("agy lists %d plugins, the marketplace declares %d", len(listed), len(plugins))
		}
		if got := lines(run(t, home, agy, "agents")); !sameSet(got, want) {
			t.Errorf("agy agents after install: %v, want %v", got, want)
		}
	})
}

// installEach installs every staged plugin directory into home with `agy plugin install`, one at a
// time — the way sc-doctor -fix and a consumer install them.
func installEach(t *testing.T, home, agy, root string, plugins []stagedPlugin) {
	t.Helper()
	for _, p := range plugins {
		run(t, home, agy, "plugin", "install", filepath.Join(root, "plugins", p.name))
	}
}

// TestLiveBothCLIsInstantiateAgentsWithValidTools exercises real executor construction on both
// `agy -p --agent` and `claude -p --plugin-dir --agent`. Static `agy plugin validate --strict` and
// `agy agents` only check YAML shape and never resolve `tools:` against the live tool registry;
// this test proves every distinct toolset across both plugins instantiates cleanly on both CLIs
// and includes a negative control planting a Claude tool name in an Antigravity agent.
func TestLiveBothCLIsInstantiateAgentsWithValidTools(t *testing.T) {
	if os.Getenv("HOOKCORE_LIVE") != "1" {
		t.Skip("skipping live dual-CLI agent instantiation test without HOOKCORE_LIVE=1")
	}
	agy := harnessCLI(t, "agy")
	claude := harnessCLI(t, "claude")
	repo, plugins := stage(t)

	realHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	agyToken := filepath.Join(realHome, ".gemini", "antigravity-cli", "antigravity-oauth-token")
	if _, err := os.Stat(agyToken); err != nil {
		t.Skipf("no agy credentials at %s", agyToken)
	}
	seedAgyHome := func(t *testing.T) string {
		t.Helper()
		home := tempHome(t)
		dstDir := filepath.Join(home, ".gemini", "antigravity-cli")
		if err := os.MkdirAll(dstDir, 0o700); err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"antigravity-oauth-token", "antigravity_state.pbtxt"} {
			src := filepath.Join(realHome, ".gemini", "antigravity-cli", name)
			if b, err := os.ReadFile(src); err == nil {
				if err := os.WriteFile(filepath.Join(dstDir, name), b, 0o600); err != nil {
					t.Fatal(err)
				}
			}
		}
		return home
	}

	t.Run("agy instantiates agents across all distinct toolsets", func(t *testing.T) {
		home := seedAgyHome(t)
		installEach(t, home, agy, repo, plugins)
		for _, agent := range []string{"probe", "plan-auditor", "blue-lane"} {
			out := run(t, home, agy, "-p", "Reply with the single word OK.", "--agent", agent, "--model", "gemini-3.8-flash-low", "--output-format", "text")
			if strings.Contains(out, "not found in registry") {
				t.Errorf("agy -p --agent %s reported missing tool:\n%s", agent, out)
			}
		}
	})

	t.Run("agy negative control rejects Claude tool name in agents/probe.md", func(t *testing.T) {
		broken := copyTree(t, repo)
		probe := filepath.Join(broken, "plugins", "prosthetic-conscience", "agents", "probe.md")
		b, err := os.ReadFile(probe)
		if err != nil {
			t.Fatal(err)
		}
		mutated := bytes.Replace(b, []byte("read_file"), []byte("Read"), 1)
		if bytes.Equal(b, mutated) {
			t.Fatalf("%s has no read_file to replace with Read", probe)
		}
		if err := os.WriteFile(probe, mutated, 0o644); err != nil {
			t.Fatal(err)
		}
		home := seedAgyHome(t)
		installEach(t, home, agy, broken, plugins)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cmd := exec.CommandContext(ctx, agy, "-p", "Reply OK.", "--agent", "probe", "--model", "gemini-3.8-flash-low", "--output-format", "text")
		cmd.Env = isolatedEnv(home)
		cmd.Dir = home
		out, err := cmd.CombinedOutput()
		if err == nil || !strings.Contains(string(out), `tool "Read" not found in registry`) {
			t.Fatalf("expected agy -p --agent probe with Read to fail with registry error, got err=%v out:\n%s", err, string(out))
		}
	})

	t.Run("claude instantiates agents-claude with Claude built-in tools", func(t *testing.T) {
		cases := []struct {
			plugin, agent string
			wantTools     []string
		}{
			{"prosthetic-conscience", "probe", []string{"Read", "Write", "Bash"}},
			{"prosthetic-conscience", "plan-auditor", []string{"Read", "Bash", "WebSearch", "WebFetch"}},
			{"frank-exchange-of-views", "blue-lane", []string{"Read", "Write", "Edit", "Bash"}},
		}
		for _, tc := range cases {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			cmd := exec.CommandContext(ctx, claude,
				"-p", "Reply with the single word OK.",
				"--plugin-dir", filepath.Join(repo, "plugins", tc.plugin),
				"--agent", tc.plugin+":"+tc.agent,
				"--model", "haiku",
				"--output-format", "stream-json",
				"--verbose",
			)
			cmd.Dir = t.TempDir()
			out, err := cmd.CombinedOutput()
			cancel()
			if err != nil {
				t.Fatalf("claude -p --agent %s:%s failed: %v\n%s", tc.plugin, tc.agent, err, string(out))
			}
			var initTools []string
			for _, line := range strings.Split(string(out), "\n") {
				if !strings.Contains(line, `"subtype":"init"`) {
					continue
				}
				var ev struct {
					Tools []string `json:"tools"`
				}
				if json.Unmarshal([]byte(line), &ev) == nil {
					initTools = ev.Tools
					break
				}
			}
			have := map[string]bool{}
			for _, tl := range initTools {
				have[tl] = true
			}
			for _, want := range tc.wantTools {
				if !have[want] {
					t.Errorf("claude %s:%s init tools = %v, missing %q", tc.plugin, tc.agent, initTools, want)
				}
			}
		}
	})
}
