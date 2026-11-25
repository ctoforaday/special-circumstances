// Package pluginroot holds the shared plugin-root resolution core and directory-marker predicate
// used by hookfailures.Ensure, hookcore.PluginRoot, and sc-doctor across Claude Code and
// Antigravity. Callers that need extra fallbacks extend Resolve rather than duplicating it:
// hookcore.PluginRoot adds a cwd fallback, and sc-doctor adds -root, Antigravity home-install,
// and unconditional-grandparent fallbacks.
package pluginroot

import (
	"os"
	"path/filepath"
)

// Resolve resolves a plugin's root directory using the cross-harness precedence order:
//
//  1. envClaude (CLAUDE_PLUGIN_ROOT) — always wins over PLUGIN_ROOT so a Claude Code session
//     spawned inside an Antigravity shell (which inherits PLUGIN_ROOT) never resolves the wrong
//     plugin.
//  2. envPluginRoot (PLUGIN_ROOT) — Antigravity's native plugin-root hook variable.
//  3. The binary's own grandparent (bin/<hook> -> <plugin>) when IsPluginDir confirms it holds a
//     plugin marker.
//
// executable is injectable so tests can pin the grandparent fallback without relocating binaries.
func Resolve(envClaude, envPluginRoot string, executable func() (string, error)) string {
	if envClaude != "" {
		return envClaude
	}
	if envPluginRoot != "" {
		return envPluginRoot
	}
	if executable != nil {
		if exe, err := executable(); err == nil && exe != "" {
			grandparent := filepath.Dir(filepath.Dir(exe))
			if IsPluginDir(grandparent) {
				return grandparent
			}
		}
	}
	return ""
}

// IsPluginDir reports whether dir holds one of sc-doctor's five plugin-root markers, kept as one
// shared set so sc-doctor's behaviour is unchanged while hookcore and hookfailures share the same
// predicate:
//   - .antigravity-plugin/plugin.json: Antigravity plugin manifest
//   - .claude-plugin/plugin.json: Claude Code plugin manifest
//   - plugin.json: root manifest accepted by Antigravity
//   - requirements.json: toolchain manifest read by sc-doctor and sc-toolchain-nudge
//   - tools/cmd: Go hook-binary checkout layout
func IsPluginDir(dir string) bool {
	for _, sub := range []string{
		filepath.Join(".antigravity-plugin", "plugin.json"),
		filepath.Join(".claude-plugin", "plugin.json"),
		"plugin.json",
		"requirements.json",
		filepath.Join("tools", "cmd"),
	} {
		if _, err := os.Stat(filepath.Join(dir, sub)); err == nil {
			return true
		}
	}
	return false
}
