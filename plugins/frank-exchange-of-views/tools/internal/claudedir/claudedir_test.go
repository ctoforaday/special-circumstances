package claudedir

import (
	"path/filepath"
	"testing"
)

// EVERY ROW IS A VARIABLE THE CLI HONOURS. A resolver that joins $HOME/.claude answers the
// "neither" row correctly and every other row with the operator's own install — which is how a
// universe run recorded the host's four cached hook versions instead of its one.
func TestPluginCacheResolvesTheWayClaudeCodeDoes(t *testing.T) {
	home := filepath.FromSlash("/h")
	for _, tc := range []struct {
		name            string
		env             map[string]string
		config, plugins string
	}{
		{"neither", nil, "/h/.claude", "/h/.claude/plugins"},
		{"config dir", map[string]string{"CLAUDE_CONFIG_DIR": "/u/config"}, "/u/config", "/u/config/plugins"},
		{"plugin cache dir", map[string]string{"CLAUDE_CODE_PLUGIN_CACHE_DIR": "/u/plugins"}, "/h/.claude", "/u/plugins"},
		{"both: the plugin cache dir wins", map[string]string{"CLAUDE_CONFIG_DIR": "/u/config", "CLAUDE_CODE_PLUGIN_CACHE_DIR": "/u/plugins"}, "/u/config", "/u/plugins"},
		{"tilde", map[string]string{"CLAUDE_CODE_PLUGIN_CACHE_DIR": "~/p"}, "/h/.claude", "/h/p"},
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
			if got := PluginsDir(getenv, home); got != filepath.FromSlash(tc.plugins) {
				t.Errorf("PluginsDir = %q, want %q", got, tc.plugins)
			}
			if got, want := PluginCache(getenv, home), filepath.Join(filepath.FromSlash(tc.plugins), "cache"); got != want {
				t.Errorf("PluginCache = %q, want %q", got, want)
			}
		})
	}
}
