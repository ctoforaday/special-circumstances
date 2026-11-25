package claudedir

import (
	"path/filepath"
	"testing"
)

// EVERY ROW IS A VARIABLE THE CLI HONOURS. A resolver that joins $HOME/.claude answers the
// "unset" row correctly and the other with the operator's own install.
func TestConfigDirResolvesTheWayClaudeCodeDoes(t *testing.T) {
	home := filepath.FromSlash("/h")
	for _, tc := range []struct {
		name   string
		env    map[string]string
		config string
	}{
		{"unset", nil, "/h/.claude"},
		{"config dir", map[string]string{"CLAUDE_CONFIG_DIR": "/u/config"}, "/u/config"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(k string) string {
				if v, ok := tc.env[k]; ok && v[0] == '/' {
					return filepath.FromSlash(v)
				}
				return tc.env[k]
			}
			if got := ConfigDir(getenv, home); got != filepath.FromSlash(tc.config) {
				t.Errorf("ConfigDir = %q, want %q", got, tc.config)
			}
		})
	}
}
