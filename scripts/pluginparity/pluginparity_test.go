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

// A dev-only cmd/ directory is skipped by TWO readers — this checker's shipped-binary count and
// the cold-bootstrap build loop — and they must skip the same ones. Nothing makes that true by
// construction: they are a Go program and a shell script reading one marker by agreement.
//
// The failure if they drift is silent in the direction that matters. Bootstrap stops honouring
// the marker and every consumer quietly gains a binary that shells out to the `claude` CLI and
// reads this repo's agents/; the count still agrees, the check still passes, and the only
// symptom is a stray executable nobody looks at.
func TestDevOnlyMarkerIsHonouredByBothReaders(t *testing.T) {
	dir := t.TempDir()
	if devOnly(dir) {
		t.Fatal("an ordinary cmd/ directory was treated as dev-only — it would stop shipping")
	}
	if err := os.WriteFile(filepath.Join(dir, "DEV-ONLY"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !devOnly(dir) {
		t.Fatal("the marker was ignored: a development harness would ship to every consumer")
	}

	// The build script is the other reader. It cannot be called from here, so what is checked
	// is that it still consults the same marker at all.
	b, err := os.ReadFile(filepath.Join("..", "bootstrap-plugins.sh"))
	if err != nil {
		t.Fatalf("cannot read the bootstrap script: %v", err)
	}
	if !strings.Contains(string(b), "DEV-ONLY") {
		t.Fatal("bootstrap-plugins.sh no longer skips DEV-ONLY directories, so a cold bootstrap " +
			"builds a development harness into every consumer's plugin bin while this checker, " +
			"which excludes it from the count, goes on agreeing with the documented number")
	}

	// And the marker has to be on something: a marker nothing carries is a check that passes
	// by describing an empty set, which is how this class of guard usually dies.
	hits, _ := filepath.Glob(filepath.Join("..", "..", "plugins", "*", "tools", "cmd", "*", "DEV-ONLY"))
	if len(hits) == 0 {
		t.Fatal("no cmd/ directory carries a DEV-ONLY marker — if the last dev harness was " +
			"promoted or deleted, delete this mechanism too rather than leaving it passing vacuously")
	}
}
