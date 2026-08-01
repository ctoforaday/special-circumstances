package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The always-on import list decides which rules load in EVERY session of every consuming
// project, and nothing checked it before #214. Deleting one line silently disables a rule:
// the SKILL.md still exists, still passes every check, and is never loaded again.
func TestAlwaysOnImportParity(t *testing.T) {
	// A scratch tree standing in for the repo: two always-on skills, one ordinary one.
	build := func(t *testing.T, claude string, marks map[string]bool) string {
		t.Helper()
		dir := t.TempDir()
		for name, always := range marks {
			desc := "description: Something ordinary."
			if always {
				desc = "description: Always-on discipline — does a thing."
			}
			p := filepath.Join(dir, "plugins", "pc", "skills", name, "SKILL.md")
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(p, []byte("---\nname: "+name+"\n"+desc+"\n---\n# "+name+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(claude), 0o644); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	imp := func(n string) string { return "@plugins/pc/skills/" + n + "/SKILL.md\n" }

	cases := []struct {
		name, claude string
		marks        map[string]bool
		want         string
	}{
		{"agreeing", imp("a") + imp("b"), map[string]bool{"a": true, "b": true, "c": false}, ""},
		{"always-on skill not imported", imp("a"), map[string]bool{"a": true, "b": true}, "does not load, in any session"},
		{"import pointing at nothing", imp("a") + imp("ghost"), map[string]bool{"a": true}, "which does not exist"},
		{"imported but not marked always-on", imp("a") + imp("c"), map[string]bool{"a": true, "c": false}, "does not describe itself as Always-on"},
		// The flattering failure: read nothing, report nothing wrong.
		{"no imports at all", "# just a guide\n", map[string]bool{"a": true}, "declares no @ imports at all"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := alwaysOnProblems(build(t, c.claude, c.marks))
			if c.want == "" {
				if len(got) != 0 {
					t.Fatalf("a healthy tree must be silent: %v", got)
				}
				return
			}
			joined := strings.Join(got, "\n")
			if !strings.Contains(joined, c.want) {
				t.Errorf("missing %q in:\n%s", c.want, joined)
			}
		})
	}
}
