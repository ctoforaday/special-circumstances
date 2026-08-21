// Package repotree locates the repository's own files for the gates that read them, and REFUSES
// to hand back an empty set.
//
// # The defect this exists for
//
// A gate that discovers its subjects at runtime — glob the constitutions, walk the source tree,
// read the goldens directory — and then asserts a NEGATIVE over them ("no constitution names a
// verb", "no literal names a retired surface", "no golden carries a home directory") passes
// perfectly when it discovers nothing. The miss and the clean board are the same result.
//
// Three live instances, found 2026-08-21:
//
//	TestTheDutyDoesNotSmuggleAVerbListBackIn   `paths, _ := filepath.Glob(…)` — error discarded,
//	                                          no length check, and it sits BETWEEN two tests in
//	                                          the same file that glob the same pattern and do
//	                                          guard it, one of them saying why in a comment.
//	TestNoStringLiteralNamesARetiredSurface    walks ../../internal; a moved package scans nothing.
//	TestGoldensAreMachineIndependent           t.Skip("no goldens yet") on a missing directory,
//	                                          which reads green rather than absent.
//
// # And the relative paths were a second copy of the same fact
//
// Each caller hand-wrote its own `filepath.Join("..", "..", "..", …)`, so the depth was a property
// of where the test file happened to live. Moving a test between packages silently repointed it —
// at nothing, which the negative assertions would not notice. Callers name what they want here;
// the root is found by looking for it.
package repotree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Root is the repository root, found by walking up for the directories that only it has.
func Root() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		plugins := filepath.Join(dir, "plugins")
		scripts := filepath.Join(dir, "scripts")
		if isDir(plugins) && isDir(scripts) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repotree: no repository root above the working directory — " +
				"looked for a directory holding both plugins/ and scripts/")
		}
		dir = parent
	}
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// Glob returns the files matching a pattern under the repository root, and ERRORS when it matches
// nothing.
//
// The empty case is the whole point: a caller that globs a moved path gets a refusal naming the
// pattern, not an empty slice a negative assertion will sail through.
func Glob(pattern ...string) ([]string, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}
	full := filepath.Join(append([]string{root}, pattern...)...)
	matches, err := filepath.Glob(full)
	if err != nil {
		return nil, fmt.Errorf("repotree: %s: %w", full, err)
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("repotree: %s matched NOTHING. A gate that scans an empty set and "+
			"asserts a negative over it passes on every run — check whether the path moved before "+
			"believing the result", full)
	}
	return matches, nil
}

// GoSources returns every non-test .go file under a directory relative to the repository root,
// erroring when it finds none.
func GoSources(dir ...string) ([]string, error) {
	root, err := Root()
	if err != nil {
		return nil, err
	}
	base := filepath.Join(append([]string{root}, dir...)...)
	var out []string
	err = filepath.Walk(base, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		out = append(out, p)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repotree: walk %s: %w", base, err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("repotree: no Go source under %s — a walk that reads nothing reports "+
			"every file clean", base)
	}
	return out, nil
}

// Constitutions are the seat constitutions the harness dispatches under. Named here because three
// packages read them and each had its own relative path to them.
func Constitutions() ([]string, error) {
	return Glob("plugins", "frank-exchange-of-views", "agents", "*.md")
}

// Plugin is the frank-exchange-of-views plugin directory — the parent of agents/, skills/,
// commands/ and docs/. Six test packages computed it by counting `..` from wherever their own file
// sat, which made the depth a property of where a file lived rather than of what it wanted.
// Anchoring it here means moving a test between packages cannot silently repoint it.
//
// It returns an error rather than a best-effort relative path for the reason the package exists:
// a path that is wrong but plausible is the failure mode, not a missing one.
func Plugin(rel ...string) (string, error) {
	root, err := Root()
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{root, "plugins", "frank-exchange-of-views"}, rel...)...), nil
}

// ToolSources is every non-test .go file in the tool's own internal tree — the subject of the
// gates that sweep for a string a seat must not be handed.
func ToolSources() ([]string, error) {
	return GoSources("plugins", "frank-exchange-of-views", "tools", "internal")
}

// DebateJS is the orchestrator script three gates parse. One address, because it moved once
// already and each caller held its own route to it.
func DebateJS() (string, error) {
	return Plugin("skills", "research-protocol", "scripts", "debate.js")
}
