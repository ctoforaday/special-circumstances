// Package claudedir resolves the directories Claude Code itself reads — its config directory and its
// plugin cache — the way the CLI resolves them, so a tool that reports on them reports on the
// session it runs in rather than on the operator's default install.
//
// The convention is Claude Code's, not ours. Read from the CLI: the config directory is
// CLAUDE_CONFIG_DIR, else ~/.claude; the plugins directory is CLAUDE_CODE_PLUGIN_CACHE_DIR (a leading
// `~` expanded), else <config dir>/plugins; installed plugins live under <plugins dir>/cache/
// <marketplace>/<plugin>/<version>. An isolated universe (`scripts/universe.sh`) sets both
// variables, so a reader that joins $HOME/.claude reports the HOST's install instead of the run's.
//
// Not followed: a Cowork host reads `cowork_plugins` instead of `plugins`, chosen by host settings
// this process cannot see. A reader there finds no cache and says where it looked.
package claudedir

import (
	"path/filepath"
	"strings"
)

// ConfigDir is the Claude Code config directory: CLAUDE_CONFIG_DIR, else <home>/.claude.
func ConfigDir(getenv func(string) string, home string) string {
	if d := getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(home, ".claude")
}

// PluginsDir is the directory Claude Code keeps plugins in: CLAUDE_CODE_PLUGIN_CACHE_DIR, else
// <config dir>/plugins. It holds installed_plugins.json, marketplaces/ and cache/.
func PluginsDir(getenv func(string) string, home string) string {
	if d := getenv("CLAUDE_CODE_PLUGIN_CACHE_DIR"); d != "" {
		if d == "~" {
			return home
		}
		if strings.HasPrefix(d, "~/") {
			return filepath.Join(home, d[2:])
		}
		return d
	}
	return filepath.Join(ConfigDir(getenv, home), "plugins")
}

// PluginCache is the versioned install cache: <plugins dir>/cache, holding
// <marketplace>/<plugin>/<version>.
func PluginCache(getenv func(string) string, home string) string {
	return filepath.Join(PluginsDir(getenv, home), "cache")
}
