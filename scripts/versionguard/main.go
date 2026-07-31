// Command version-guard fails when a plugin's version goes BACKWARDS, or when its content
// changed without the version moving.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// WHY IT EXISTS. CLAUDE.md states the rule — "Every PR that changes a plugin's content MUST
// bump that plugin's version in its plugin.json; /plugin update is version-gated and ships
// nothing without it" — and nothing enforced it. On 2026-07-31 two pull requests merged
// hours apart:
//
//	#198 set prosthetic-conscience to 0.22.0 and added the strike counter
//	#195, branched earlier, carried 0.21.0 and merged ON TOP
//
// Git resolved the version like any other line: last writer won. main kept BOTH features'
// content and REVERTED the version to 0.21.0 — a version already shipped. Every consumer
// who had pulled 0.21.0 would therefore never receive the strike counter, and nothing
// anywhere would have said so. No test failed. No guard fired. The plugin was simply
// undeliverable, silently.
//
// That is policy-without-mechanism, on a rule the repository's own guide states in its
// first bullet. This is the mechanism.
//
// It compares against the merge-base rather than the branch tip, so it asks "did THIS
// branch move the version" rather than "is this version the newest anyone has used".
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/scripts/internal/gitx"
)

type manifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// semver compares dotted numeric versions. Missing parts count as zero, so 0.22 and 0.22.0
// are the same version rather than an ordering surprise.
func semver(a, b string) int {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < 3; i++ {
		x, y := 0, 0
		if i < len(pa) {
			x, _ = strconv.Atoi(strings.TrimSpace(pa[i]))
		}
		if i < len(pb) {
			y, _ = strconv.Atoi(strings.TrimSpace(pb[i]))
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func versionAt(root, ref, path string) (string, bool) {
	out, err := gitx.Run(root, "show", ref+":"+path)
	if err != nil {
		return "", false // absent at the base: a new plugin, nothing to compare
	}
	var m manifest
	if json.Unmarshal([]byte(out), &m) != nil {
		return "", false
	}
	return m.Version, m.Version != ""
}

// check reports every version problem on this branch, and how many plugins it compared.
// Separated from main so the failure it exists to catch can be reproduced in a scratch
// repository instead of being argued about.
func check(root, base string) (problems []string, checked int, err error) {
	// The merge-base, not the tip: the question is whether THIS branch moved the version,
	// not whether it happens to match whatever landed on main since it branched.
	mb, err := gitx.Run(root, "merge-base", base, "HEAD")
	if err != nil {
		return nil, 0, fmt.Errorf("cannot find the merge-base with %s: %w", base, err)
	}
	changed, err := gitx.Run(root, "diff", "--name-only", mb+"...HEAD")
	if err != nil {
		return nil, 0, err
	}
	manifests, err := filepath.Glob(filepath.Join(root, "plugins", "*", ".claude-plugin", "plugin.json"))
	if err != nil {
		return nil, 0, err
	}
	if len(manifests) == 0 {
		return nil, 0, fmt.Errorf("no plugin manifests found at all — the glob is wrong, not the tree")
	}
	sort.Strings(manifests)

	for _, abs := range manifests {
		rel, _ := filepath.Rel(root, abs)
		rel = filepath.ToSlash(rel)
		plugin := strings.Split(rel, "/")[1]

		b, readErr := os.ReadFile(abs)
		if readErr != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", rel, readErr))
			continue
		}
		var now manifest
		if json.Unmarshal(b, &now) != nil || now.Version == "" {
			problems = append(problems, fmt.Sprintf("%s: no version to check", rel))
			continue
		}
		was, ok := versionAt(root, mb, rel)
		if !ok {
			continue // new plugin: nothing to compare against
		}
		checked++

		touched := false
		for _, f := range gitx.Lines(changed) {
			if strings.HasPrefix(filepath.ToSlash(f), "plugins/"+plugin+"/") {
				touched = true
				break
			}
		}

		switch {
		case semver(now.Version, was) < 0:
			problems = append(problems, fmt.Sprintf(
				"%s: version went BACKWARDS, %s -> %s. `/plugin update` is version-gated, so a consumer already on %s "+
					"receives NOTHING from this branch — the content ships and the version says it did not. "+
					"This is what happens when two branches bump independently and the later merge wins the line.",
				plugin, was, now.Version, was))
		case touched && semver(now.Version, was) == 0:
			problems = append(problems, fmt.Sprintf(
				"%s: content changed under plugins/%s/ and the version stayed at %s. "+
					"`/plugin update` ships nothing without a bump (CLAUDE.md).", plugin, plugin, was))
		}
	}
	return problems, checked, nil
}

func main() {
	base := "origin/main"
	if len(os.Args) > 2 && os.Args[1] == "-base" {
		base = os.Args[2]
	}
	root, err := gitx.Root()
	if err != nil {
		fmt.Fprintln(os.Stderr, "version-guard:", err)
		os.Exit(1)
	}
	problems, checked, err := check(root, base)
	if err != nil {
		fmt.Fprintln(os.Stderr, "version-guard:", err)
		os.Exit(1)
	}
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
	fmt.Printf("plugin versions move forward: %d checked against %s\n", checked, base)
}
