package hookinvocation

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// declines lists the binaries whose packages call hookenv.Explain — the ones that have
// declared, in a message they print, that they will do NOTHING rather than guess a project
// root from the working directory.
//
//	grep -rln "hookenv.Explain" --include=*.go internal/ | grep -v _test
var declines = []string{
	"sc-filechanged-rearm", "sc-postcompact-observe", "sc-precompact",
	"sc-sessionend", "sc-sessionstart", "sc-stop", "sc-strike-counter",
	"sc-subagentstop",
}

// A HOOK WITH NO PROJECT ROOT MUST NOT FALL BACK TO ITS WORKING DIRECTORY.
//
// hookenv.Explain states the contract in the message it prints — "doing nothing rather
// than guessing from the working directory" — and one binary was doing exactly the
// guessing that message disclaims: sc-stop's Main passed os.Getwd() where its siblings
// pass CLAUDE_PROJECT_DIR. Because Explain prefers its first argument and a working
// directory is never empty, the refusal could not fire and the payload fallback beneath it
// was unreachable.
//
// The client's own launch directory is not the project. A hook that reads a CHECKPOINT.md
// found there is reading somebody else's note, and writing beside it is worse.
//
// THE ASSERTION IS A FILESYSTEM OBSERVATION, not a search for the message. Whether the
// refusal is *worded* correctly is a different question from whether it HAPPENED, and a
// substring that stops matching after a reword reads exactly like a pass.
func TestNoHookGuessesAProjectRootFromItsWorkingDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("builds every hook binary")
	}
	root := build(t)

	for _, name := range declines {
		t.Run(name, func(t *testing.T) {
			// BAIT: a working directory that looks exactly like a project root, complete
			// with a note worth acting on. A binary that guesses will find it.
			bait := t.TempDir()
			if err := os.MkdirAll(filepath.Join(bait, ".claude", "checkpoints"), 0o755); err != nil {
				t.Fatal(err)
			}
			note := filepath.Join(bait, ".claude", "checkpoints", "CHECKPOINT.md")
			body := "---\nschema: 3\nwritten_at: 2026-08-23T00:00:00Z\n---\n## Validation loop\n1. x\n"
			if err := os.WriteFile(note, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			before := treeOf(t, bait)

			bin := filepath.Join(root, "bin", name+exeSuffix())
			cmd := exec.Command(bin)
			cmd.Dir = bait
			cmd.Env = withoutProjectRoot(os.Environ())
			// A payload carrying NO cwd: the second and last legitimate source of a root.
			cmd.Stdin = strings.NewReader(`{"session_id":"s1"}`)
			out, err := cmd.Output()
			if err != nil {
				t.Errorf("exited non-zero (%v) with no project root; a hook must degrade, not fail the event", err)
			}
			if len(out) != 0 {
				t.Errorf("wrote to STDOUT with no project root: %q", out)
			}

			if after := treeOf(t, bait); after != before {
				t.Errorf("touched the working directory it was told not to guess from:\n before: %s\n after:  %s", before, after)
			}
			if got, err := os.ReadFile(note); err == nil && string(got) != body {
				t.Errorf("rewrote a CHECKPOINT.md it found in its working directory")
			}
		})
	}
}

// withoutProjectRoot removes every variable a hook may legitimately read a root from, so
// the only remaining source is the payload — which carries none in this test.
func withoutProjectRoot(env []string) []string {
	out := env[:0:0]
	for _, kv := range env {
		if strings.HasPrefix(kv, "CLAUDE_PROJECT_DIR=") || strings.HasPrefix(kv, "CLAUDE_PLUGIN_ROOT=") {
			continue
		}
		out = append(out, kv)
	}
	return out
}

// treeOf renders a directory as a sorted path list, so an added, removed or renamed file
// is a string difference rather than something a test has to go looking for.
func treeOf(t *testing.T, dir string) string {
	t.Helper()
	var paths []string
	err := filepath.Walk(dir, func(p string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	return strings.Join(paths, " ")
}
