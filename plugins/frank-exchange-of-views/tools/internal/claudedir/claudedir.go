// Package claudedir resolves the directory Claude Code itself reads its configuration from, the way
// the CLI resolves it, so a tool that reports on it reports on the session it runs in rather than
// on the operator's default install.
//
// The convention is Claude Code's, not ours. Read from the CLI: the config directory is
// CLAUDE_CONFIG_DIR, else ~/.claude. An isolated universe (`scripts/universe.sh`) sets the
// variable, so a reader that joins $HOME/.claude reports the HOST's install instead of the run's.
package claudedir

import "path/filepath"

// ConfigDir is the Claude Code config directory: CLAUDE_CONFIG_DIR, else <home>/.claude.
func ConfigDir(getenv func(string) string, home string) string {
	if d := getenv("CLAUDE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(home, ".claude")
}
