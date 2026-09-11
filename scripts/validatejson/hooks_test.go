package main

import (
	"strings"
	"testing"
)

// The shape every plugin's hooks.json has today, once the prose moved out: a description, and
// entries with only documented keys.
const cleanHooks = `{
  "description": "one line",
  "hooks": {
    "PreToolUse": [{"matcher": "Bash", "hooks": [{"type": "command", "command": "x", "timeout": 5}]}],
    "Stop": [{"hooks": [{"type": "command", "command": "y"}]}]
  }
}`

func TestHookKeysAcceptTheDocumentedShape(t *testing.T) {
	if got := hookKeyProblems([]byte(cleanHooks)); len(got) != 0 {
		t.Fatalf("a documented hooks.json was refused: %q", got)
	}
}

// EVERY PLACE A `_comment` LIVED ON 2026-09-11 is refused, each named where it sits — the top level,
// a matcher group, and (never used, but the same warning) a hook entry.
func TestHookKeysRefuseUndocumentedKeys(t *testing.T) {
	for _, tc := range []struct {
		name, file, want string
	}{
		{"top level", `{"_comment": "why", "hooks": {}}`, `the top level: key "_comment"`},
		{"matcher group", `{"hooks": {"Stop": [{"_comment": "why", "hooks": [{"type": "command", "command": "y"}]}]}}`,
			`hooks.Stop[0]: key "_comment"`},
		{"hook entry", `{"hooks": {"Stop": [{"hooks": [{"type": "command", "command": "y", "comment": "why"}]}]}}`,
			`hooks.Stop[0].hooks[0]: key "comment"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := hookKeyProblems([]byte(tc.file))
			if len(got) != 1 || !strings.Contains(got[0], tc.want) {
				t.Fatalf("problems = %q, want exactly one naming %q", got, tc.want)
			}
		})
	}
}

// A file that is not the documented shape is a problem too, never a silent pass: a check that only
// looked for bad keys would call a file with no "hooks" at all clean.
func TestHookKeysRefuseTheWrongShape(t *testing.T) {
	for _, file := range []string{
		`[]`,
		`{"description": "no hooks"}`,
		`{"hooks": []}`,
		`{"hooks": {"Stop": [{"matcher": "x"}]}}`,
	} {
		if got := hookKeyProblems([]byte(file)); len(got) == 0 {
			t.Errorf("%s was reported clean", file)
		}
	}
}

func TestIsPluginHooks(t *testing.T) {
	for p, want := range map[string]bool{
		"plugins/gray-area/hooks/hooks.json":         true,
		"plugins/gray-area/hooks/other.json":         false,
		"plugins/gray-area/nested/hooks/hooks.json":  false,
		"plugins/gray-area/tools/x/hooks/hooks.json": false,
		".claude/settings.json":                      false,
		"hooks/hooks.json":                           false,
	} {
		if got := isPluginHooks(p); got != want {
			t.Errorf("isPluginHooks(%q) = %v, want %v", p, got, want)
		}
	}
}
