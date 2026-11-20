package installshape

// What the clients install: this checkout's marketplace and plugins, tools/ and native/ included
// (fetch-bin derives its binary set from them, and a real install carries both), bin/ holding only
// .gitkeep. Each plugin's .claude-plugin/hooks.json gains a ROOT PROBE on SessionStart,
// UserPromptSubmit and SessionEnd that writes the CLAUDE_PLUGIN_ROOT the client handed it; that
// line, and nothing the test computes, is the root every assertion is checked against.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

// probeEvents are the events a no-login `claude -p` fires.
var probeEvents = []string{"SessionStart", "UserPromptSubmit", "SessionEnd"}

// stageTree copies the installable part of the checkout into dst.
func stageTree(t *testing.T, repo, dst string) {
	t.Helper()
	out, err := gitx.Run(repo, "ls-files", "-co", "--exclude-standard", "--", ".claude-plugin", ".agents", "plugins")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, rel := range gitx.Lines(out) {
		parts := strings.Split(rel, "/")
		if len(parts) >= 4 && parts[0] == "plugins" && parts[2] == "bin" && parts[3] != ".gitkeep" {
			continue
		}
		src := filepath.Join(repo, filepath.FromSlash(rel))
		fi, err := os.Lstat(src)
		if os.IsNotExist(err) {
			continue // deleted in the worktree, still in the index
		}
		if err != nil {
			t.Fatal(err)
		}
		if !fi.Mode().IsRegular() {
			continue
		}
		copyFile(t, src, filepath.Join(dst, filepath.FromSlash(rel)), fi.Mode().Perm())
		n++
	}
	if n == 0 {
		t.Fatal("staged zero files: the listing is wrong, not the tree")
	}
}

// stagedPlugins lists the marketplace's plugins, in its order, as the staged tree carries them.
func stagedPlugins(t *testing.T, stage string) []plugin {
	t.Helper()
	var m struct {
		Plugins []struct{ Name string } `json:"plugins"`
	}
	b, err := os.ReadFile(filepath.Join(stage, ".claude-plugin", "marketplace.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	var out []plugin
	for _, e := range m.Plugins {
		out = append(out, readPlugin(t, filepath.Join(stage, "plugins", e.Name)))
	}
	if len(out) == 0 {
		t.Fatal("the marketplace lists no plugins")
	}
	return out
}

// probeCommand is the root-probe hook command for one plugin and event.
func probeCommand(plugin, event, rootsLog string) string {
	return "printf '%s\\t%s\\t%s\\n' " + shellQuote(plugin) + " " + shellQuote(event) + ` "$CLAUDE_PLUGIN_ROOT" >> ` + shellQuote(rootsLog)
}

// injectProbes adds the root probe to every plugin's .claude-plugin/hooks.json and returns the
// probe commands, which the hook drive skips: they are the test's, not the plugin's.
func injectProbes(t *testing.T, stage string, plugins []plugin, rootsLog string) map[string]bool {
	t.Helper()
	probes := map[string]bool{}
	for _, p := range plugins {
		path := filepath.Join(stage, "plugins", p.name, ".claude-plugin", "hooks.json")
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue // a plugin with no hooks has no root to observe
		}
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		hooks, _ := doc["hooks"].(map[string]any)
		if hooks == nil {
			t.Fatalf("%s has no hooks object", path)
		}
		for _, ev := range probeEvents {
			c := probeCommand(p.name, ev, rootsLog)
			probes[c] = true
			arr, _ := hooks[ev].([]any)
			hooks[ev] = append(arr, map[string]any{"hooks": []any{map[string]any{"type": "command", "command": c}}})
		}
		writeJSON(t, path, doc)
	}
	return probes
}

// setVersion rewrites a staged plugin's version in both manifests. fetch-bin reads the version
// with sed at the two-space indent, which MarshalIndent keeps.
func setVersion(t *testing.T, dir, version string) {
	t.Helper()
	for _, m := range []string{".claude-plugin", ".antigravity-plugin"} {
		path := filepath.Join(dir, m, "plugin.json")
		b, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		var doc map[string]any
		if err := json.Unmarshal(b, &doc); err != nil {
			t.Fatal(err)
		}
		doc["version"] = version
		writeJSON(t, path, doc)
	}
}

// setSources rewrites the staged marketplace's plugin sources.
func setSources(t *testing.T, stage string, source func(name string) any) {
	t.Helper()
	path := filepath.Join(stage, ".claude-plugin", "marketplace.json")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	for _, e := range doc["plugins"].([]any) {
		entry := e.(map[string]any)
		entry["source"] = source(entry["name"].(string))
	}
	writeJSON(t, path, doc)
}

// plantDecoy puts every plugin into the HOST cache, $HOME/.claude/plugins/cache, at version
// decoyVersion — the tree the 09-08 hook observation read while CLAUDE_CONFIG_DIR pointed
// elsewhere. Its bin/ and its fetch-bin.sh are tripwires, stamped as installed, and one binary of
// each plugin is left out, so a reader that lists it answers differently from the loaded root.
func plantDecoy(t *testing.T, stage, home string, plugins []plugin, decoyVersion, trace string) []string {
	t.Helper()
	var roots []string
	for _, p := range plugins {
		dst := filepath.Join(home, ".claude", "plugins", "cache", "special-circumstances", p.name, decoyVersion)
		copyTree(t, filepath.Join(stage, "plugins", p.name), dst)
		setVersion(t, dst, decoyVersion)
		if p.hasBinaries() {
			fillTripwires(t, dst, p, trace)
			removeOne(t, dst, p)
			write(t, filepath.Join(dst, ".fetch", "installed"), p.tag(decoyVersion)+"\n", 0o644)
		}
		roots = append(roots, dst)
	}
	return roots
}

// fillTripwires replaces a root's bin/ and hooks/fetch-bin.sh with tripwires, keeping the real
// bin/ beside it as bin.release.
func fillTripwires(t *testing.T, root string, p plugin, trace string) {
	t.Helper()
	bin := filepath.Join(root, "bin")
	if _, err := os.Stat(bin); err == nil {
		if err := os.Rename(bin, filepath.Join(root, "bin.release")); err != nil {
			t.Fatal(err)
		}
	}
	mkdir(t, bin)
	for name := range p.installed {
		write(t, filepath.Join(bin, name), tripwire(trace), 0o755)
	}
	write(t, filepath.Join(root, "hooks", "fetch-bin.sh"), tripwire(trace), 0o755)
}

// removeOne deletes the first command's binary from a non-loaded tree's bin/.
func removeOne(t *testing.T, root string, p plugin) {
	t.Helper()
	if len(p.cmds) == 0 {
		return
	}
	if err := os.Remove(filepath.Join(root, "bin", p.cmds[0])); err != nil {
		t.Fatal(err)
	}
}

// pluginTrees finds every plugin tree under base: a directory holding a plugin manifest.
func pluginTrees(t *testing.T, base string) []string {
	t.Helper()
	var out []string
	err := filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if d.Name() == ".git" || d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		for _, m := range []string{".claude-plugin", ".antigravity-plugin"} {
			if _, err := os.Stat(filepath.Join(p, m, "plugin.json")); err == nil {
				out = append(out, p)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(out)
	return out
}

// clientBookkeeping are files the CLIENT writes into a cache version directory; they say nothing
// about what the suite installed there.
var clientBookkeeping = map[string]bool{".in_use": true, ".orphaned_at": true}

// treeHash hashes every file under a plugin tree: path, mode and content.
func treeHash(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Dir(rel) == "." && clientBookkeeping[d.Name()] {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		io.WriteString(h, rel+"\x00"+fi.Mode().String()+"\x00")
		if fi.Mode().IsRegular() {
			f, err := os.Open(p)
			if err != nil {
				return err
			}
			_, err = io.Copy(h, f)
			f.Close()
			if err != nil {
				return err
			}
		} else if fi.Mode()&fs.ModeSymlink != 0 {
			target, _ := os.Readlink(p)
			io.WriteString(h, target)
		}
		io.WriteString(h, "\x00")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// snapshot hashes every plugin tree under base.
func snapshot(t *testing.T, base string) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, tree := range pluginTrees(t, base) {
		out[tree] = treeHash(t, tree)
	}
	return out
}

// --- small filesystem helpers ---

func mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func write(t *testing.T, path, content string, mode os.FileMode) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, mode); err != nil {
		t.Fatal(err)
	}
}

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		t.Fatal(err)
	}
	write(t, path, buf.String(), 0o644)
}

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

func copyFile(t *testing.T, src, dst string, mode os.FileMode) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	write(t, dst, string(b), mode)
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return os.MkdirAll(target, 0o755)
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		copyFile(t, p, target, fi.Mode().Perm())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
