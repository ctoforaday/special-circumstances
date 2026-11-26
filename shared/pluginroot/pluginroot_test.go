package pluginroot_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ctoforaday/special-circumstances/shared/pluginroot"
)

func TestResolvePrecedenceAndFallback(t *testing.T) {
	noExe := func() (string, error) { return "", errors.New("no exe") }

	t.Run("CLAUDE_PLUGIN_ROOT beats PLUGIN_ROOT when both are set", func(t *testing.T) {
		got := pluginroot.Resolve("/claude/plugin", "/agy/plugin", noExe)
		if got != "/claude/plugin" {
			t.Fatalf("Resolve = %q, want /claude/plugin", got)
		}
	})

	t.Run("PLUGIN_ROOT wins when CLAUDE_PLUGIN_ROOT is empty", func(t *testing.T) {
		got := pluginroot.Resolve("", "/agy/plugin", noExe)
		if got != "/agy/plugin" {
			t.Fatalf("Resolve = %q, want /agy/plugin", got)
		}
	})

	t.Run("grandparent fallback matches every plugin marker", func(t *testing.T) {
		markers := []string{
			filepath.Join(".antigravity-plugin", "plugin.json"),
			filepath.Join(".claude-plugin", "plugin.json"),
			"plugin.json",
			"requirements.json",
			filepath.Join("tools", "cmd"),
		}
		for _, m := range markers {
			t.Run(m, func(t *testing.T) {
				root := t.TempDir()
				target := filepath.Join(root, m)
				if m == filepath.Join("tools", "cmd") {
					if err := os.MkdirAll(target, 0o755); err != nil {
						t.Fatal(err)
					}
				} else {
					if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
						t.Fatal(err)
					}
					if err := os.WriteFile(target, []byte(`{}`), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				exe := func() (string, error) { return filepath.Join(root, "bin", "hook"), nil }
				if got := pluginroot.Resolve("", "", exe); got != root {
					t.Fatalf("Resolve for marker %s = %q, want %q", m, got, root)
				}
			})
		}
	})

	t.Run("grandparent without plugin marker returns empty", func(t *testing.T) {
		dir := t.TempDir()
		exe := func() (string, error) { return filepath.Join(dir, "bin", "plain"), nil }
		if got := pluginroot.Resolve("", "", exe); got != "" {
			t.Fatalf("Resolve for non-plugin grandparent = %q, want empty", got)
		}
	})
}
