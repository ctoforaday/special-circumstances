package doctor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ReferenceCommit answers the question a staleness check needs and could not previously ask of
// an installed plugin: which commit SHOULD this binary have been built from?
//
// In a checkout the answer is HEAD. In an INSTALL it is not available that way at all — a plugin
// lives under .claude/plugins/cache/<marketplace>/<plugin>/<version> or ~/.gemini/config/plugins/<plugin>,
// which is a copied tree with no .git — so `git rev-parse HEAD` failed there for every binary in every
// real installation, and the whole staleness arm was inert in the only place it runs in production.
//
// The client already records the answer. installed_plugins.json carries `gitCommitSha` per
// install, and the release binaries are built from exactly that commit — measured: the recorded
// sha and the binaries' vcs.revision agreed byte for byte on this machine. Additionally, an
// installed plugin may carry `.fetch/commit` stamped at install time.
//
// Matched on installPath first, then on recorded .fetch/commit, then on installed_plugins.json
// by plugin name and version.
func ReferenceCommit(b binStatus) string {
	if head := HeadCommit(b.Root); head != "" {
		return head
	}
	if commitBytes, err := os.ReadFile(filepath.Join(b.Root, ".fetch", "commit")); err == nil {
		if s := strings.TrimSpace(string(commitBytes)); s != "" {
			return s
		}
	}
	if sha := installedCommit(b.Root); sha != "" {
		return sha
	}
	if sha := installedCommitByName(b.Root, b.Plugin, b.Version); sha != "" {
		return sha
	}
	return ""
}

// installedPlugins is the shape this reads, and ONLY the shape it reads. The client owns this
// file; a decoder that insisted on the rest of it would break on a field the client added.
type installedPlugins struct {
	Plugins map[string][]struct {
		InstallPath  string `json:"installPath"`
		GitCommitSha string `json:"gitCommitSha"`
		Version      string `json:"version"`
	} `json:"plugins"`
}

// installedCommit finds the record whose installPath IS this plugin root.
//
// It walks upward for the file rather than counting directories to it. The depth is a property of
// the client's layout, not of anything this package controls, and re-deriving it from a path is
// how a reader breaks silently the next time that layout gains a level.
func installedCommit(root string) string {
	dir := root
	for range 8 {
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
		b, err := os.ReadFile(filepath.Join(dir, "installed_plugins.json"))
		if err != nil {
			continue
		}
		var ip installedPlugins
		if json.Unmarshal(b, &ip) != nil {
			return "" // present and unreadable: say nothing rather than guess
		}
		for _, entries := range ip.Plugins {
			for _, e := range entries {
				if e.InstallPath == root && e.GitCommitSha != "" {
					return e.GitCommitSha
				}
			}
		}
		return ""
	}
	return ""
}

// installedCommitByName matches a plugin entry in installed_plugins.json by plugin name
// and version, supporting cross-harness installs (such as Antigravity importing plugins).
func installedCommitByName(root, plugin, version string) string {
	if plugin == "" || version == "" {
		return ""
	}
	var candidateFiles []string
	dir := root
	for range 8 {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
		candidateFiles = append(candidateFiles, filepath.Join(dir, "installed_plugins.json"))
	}
	if h := homeDir(); h != "" {
		candidateFiles = append(candidateFiles, filepath.Join(h, ".claude", "plugins", "installed_plugins.json"))
	}

	seen := make(map[string]bool)
	for _, path := range candidateFiles {
		if seen[path] {
			continue
		}
		seen[path] = true
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var ip installedPlugins
		if json.Unmarshal(b, &ip) != nil {
			continue
		}
		for key, entries := range ip.Plugins {
			name := key
			if at := strings.IndexByte(key, '@'); at >= 0 {
				name = key[:at]
			}
			if name != plugin {
				continue
			}
			for _, e := range entries {
				if e.Version == version && e.GitCommitSha != "" {
					return e.GitCommitSha
				}
			}
		}
	}
	return ""
}

// NoReferenceCount reports how many BUILT and STAMPED binaries had nothing to be checked against.
//
// Counted apart from UnstampedCount because the two name different parties: an unstamped binary is
// a build problem the operator can fix, and a missing reference is this program having no second
// operand. Reporting the second in the first's words sent every reader after the wrong thing.
func NoReferenceCount(bins []binStatus) int {
	n := 0
	for _, b := range bins {
		if !b.Built {
			continue
		}
		if v, _ := stampVerdict(b); v == NoReference {
			n++
		}
	}
	return n
}
