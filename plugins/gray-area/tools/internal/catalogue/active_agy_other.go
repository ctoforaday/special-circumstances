//go:build !linux

package catalogue

import (
	"os"
	"path/filepath"
	"strings"
)

// ActiveAntigravitySessions off Linux falls back to inspecting active WAL files in the default
// Antigravity conversations directory.
func ActiveAntigravitySessions() map[string]int {
	active := map[string]int{}
	home, err := os.UserHomeDir()
	if err != nil {
		return active
	}
	pattern := filepath.Join(home, ".gemini", "antigravity-cli", "conversations", "*-wal")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return active
	}
	for _, m := range matches {
		base := strings.TrimSuffix(filepath.Base(m), "-wal")
		base = strings.TrimSuffix(base, ".db")
		if base != "" && base != "conversation_summaries" {
			active[base] = 0
		}
	}
	return active
}
