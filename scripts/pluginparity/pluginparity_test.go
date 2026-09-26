package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
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

// A development harness wearing a shipped binary's directory is found by what its code
// imports, and only there: the same import is legal in a test and under tools/devcmd/.
func TestShippedHarnessIsFoundByItsImport(t *testing.T) {
	const imp = "package main\n\nimport _ \"example.com/x/internal/repotree\"\n"
	write := func(t *testing.T, root, rel, body string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		name, rel string
		want      bool
	}{
		{"a shipped command", "plugins/p/tools/cmd/probe/main.go", true},
		{"an internal package a shipped command can reach", "plugins/p/tools/internal/lib/lib.go", true},
		{"a development harness", "plugins/p/tools/devcmd/probe/main.go", false},
		{"a test", "plugins/p/tools/cmd/hook/main_test.go", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			root := t.TempDir()
			write(t, root, "plugins/p/tools/cmd/hook/main.go", "package main\n")
			write(t, root, c.rel, imp)
			got := shippedHarnessProblems(root)
			if c.want != (len(got) == 1) || len(got) > 1 {
				t.Fatalf("want a finding: %v; got %v", c.want, got)
			}
		})
	}
	if got := shippedHarnessProblems(t.TempDir()); len(got) != 1 || !strings.Contains(got[0], "read no Go file") {
		t.Fatalf("an empty tree must be refused, not passed: %v", got)
	}
}

// THE GUARD WAS REAL AND POINTED AT THE WRONG ALTITUDE.
//
// committedBinaries scanned `ls-files plugins`, while its own reasoning — "binaries reach a
// consumer through Releases or a local build, never through git, so a tracked executable is
// always an accident" — is about GIT, not about plugins/. A 3.1 MB `scripts/check.exe` was
// committed and reached main one directory outside what the guard read, and the guard
// reported a clean board throughout. `invariant-at-wrong-level`: the enforcement was real,
// the altitude was wrong.
//
// This builds a throwaway repository so the check runs against a tree whose contents are the
// test's own, rather than against whatever happens to be tracked here.
func TestCommittedBinariesSeesOutsidePlugins(t *testing.T) {
	if _, err := gitx.Root(); err != nil {
		t.Skipf("not a git checkout: %v", err)
	}
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := gitx.Run(root, args...); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "t@example.com")
	run("config", "user.name", "t")

	// A source file, a plugin-side executable, and one OUTSIDE plugins/ — the case that got
	// through. Named `.txt` deliberately: the check is on CONTENT, so a name that reveals
	// nothing must not help it hide.
	write := func(rel string, body []byte) {
		t.Helper()
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, body, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("plugins/x/tools/main.go", []byte("package main\n"))
	write("plugins/x/tools/feov-record", []byte{0x7f, 'E', 'L', 'F', 0, 0})
	write("scripts/notes.txt", []byte{'M', 'Z', 0x90, 0, 0})
	run("add", "-A")

	got := committedBinaries(root)
	var sawPlugin, sawScripts bool
	for _, p := range got {
		if strings.Contains(p, "plugins/x/tools/feov-record") {
			sawPlugin = true
		}
		if strings.Contains(p, "scripts/notes.txt") {
			sawScripts = true
		}
		if strings.Contains(p, "main.go") {
			t.Errorf("a Go source file was reported as an executable: %s", p)
		}
	}
	if !sawPlugin {
		t.Error("the original plugins/ case must still be caught")
	}
	if !sawScripts {
		t.Error("a tracked executable OUTSIDE plugins/ was not reported — this is the exact " +
			"miss that put a 3.1 MB scripts/check.exe on main while the guard read clean")
	}
}

// A check that reads nothing must SAY so. An empty repository and a clean one are otherwise
// the same output, which is the failure mode this whole file is written against.
func TestCommittedBinariesRefusesToPassOnAnEmptyListing(t *testing.T) {
	root := t.TempDir()
	if _, err := gitx.Run(root, "init"); err != nil {
		t.Skip("git unavailable")
	}
	got := committedBinaries(root)
	if len(got) != 1 || !strings.Contains(got[0], "without reading anything") {
		t.Errorf("a zero-file listing must be reported as unmeasured, not as a pass: %v", got)
	}
}
