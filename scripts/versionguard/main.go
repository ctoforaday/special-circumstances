// Command version-guard polices the two ways a plugin version can be WRONG: going backwards on
// a branch, and disagreeing with the release tag that publishes it.
//
// Dev tooling for this repository only. Nothing here ships to an installing project.
//
// It no longer demands a bump per pull request. Versions move at a RELEASE BOUNDARY, which is a
// human call; see releaseTagMatchesManifest and the note inside check for why that rule was
// removed rather than relaxed (#405).
//
// WHY IT EXISTS. CLAUDE.md stated a version rule and nothing enforced it. On 2026-07-31 two
// pull requests merged hours apart:
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

// This tool no longer has a DID-NOT-MEASURE state, and so no longer declares exit code 3.
// That code existed because the retired must-bump rule read committed history on one side and
// the manifest on disk on the other, so an uncommitted change made the question unanswerable.
// Both surviving checks read the tree, so both always answer. `scripts/check` still understands
// exit 3 — rulesweep emits it — and that is the right place for the contract to live.
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

		// THE PER-PR BUMP RULE IS GONE, AND ITS ABSENCE IS THE POINT (#405).
		//
		// This used to fail any PR that changed plugin content without bumping the version, on
		// the reasoning that `/plugin update` is version-gated. True, and it produced the wrong
		// system: versions became a per-commit counter (1.45 -> 1.57 in one afternoon) that told
		// a consumer nothing about whether anything they cared about had changed, while the tags
		// never moved. That is exactly why NO plugin has a tag matching its version and
		// `sc-doctor -fix` asks for releases that do not exist.
		//
		// The intent was always to tag at a RELEASE BOUNDARY — a human call about when the
		// binary/text contract has really moved — and to bump there. So an ordinary PR may leave
		// the version alone, a release bumps and tags in one act, and `releaseTagMatchesManifest`
		// below is what makes the tag and the manifest unable to disagree.
		//
		// WHAT SURVIVES: a version may never go BACKWARDS. That arm caught two real concurrent
		// bumps this month, where two branches took the same number and the later merge won the
		// line — and under the new model it is the only version motion left to police.
		if semver(now.Version, was) < 0 {
			problems = append(problems, fmt.Sprintf(
				"%s: version went BACKWARDS, %s -> %s. `/plugin update` is version-gated, so a consumer already on %s "+
					"receives NOTHING from this branch — the content ships and the version says it did not. "+
					"This is what happens when two branches bump independently and the later merge wins the line.",
				plugin, was, now.Version, was))
		}
	}
	return problems, checked, nil
}

// releaseTagMatchesManifest is the release-boundary invariant, and it is what replaces the
// per-PR bump rule (#405).
//
// The original intent was to tag when the binary/text contract really moves — a human call —
// and to bump the plugin there. The per-PR rule turned the version into a commit counter and
// left the tags behind: measured 2026-08-15, NO plugin had a tag matching its own version, so
// `sc-doctor -fix` pinned every download to a release that does not exist.
//
// A tag cannot be DERIVED from the manifest, nor the manifest from the tag: the plugin system
// reads plugin.json out of the checkout, so the right version has to be IN the commit the tag
// points at. The non-circular answer is that one release act writes both, and this refuses the
// tag if they disagree. Run by the release job before it publishes anything.
func releaseTagMatchesManifest(root, tag string) error {
	plugin, version, ok := strings.Cut(tag, "--v")
	if !ok || plugin == "" || version == "" {
		return fmt.Errorf("tag %q is not <plugin>--v<version>; the release job resolves the plugin from the tag, so a malformed one builds the wrong thing or nothing", tag)
	}
	rel := filepath.Join("plugins", plugin, ".claude-plugin", "plugin.json")
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return fmt.Errorf("tag %q names plugin %q, which has no manifest at %s: %w", tag, plugin, rel, err)
	}
	var m manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	if m.Version != version {
		return fmt.Errorf("tag %s says %s, but %s says %s at this commit.\n"+
			"A release is ONE act: bump the manifest and tag that commit. A tag whose manifest "+
			"disagrees publishes assets a consumer's `/plugin update` will never ask for — which "+
			"is the state #405 measured across all four plugins.", tag, version, rel, m.Version)
	}
	return nil
}

func main() {
	base := "origin/main"
	if len(os.Args) > 1 && os.Args[1] == "-base" {
		if len(os.Args) < 3 {
			die("-base needs a ref")
		}
		base = os.Args[2]
	}
	root, err := gitx.Root()
	if err != nil {
		die(err)
	}
	if len(os.Args) > 1 && os.Args[1] == "-tag" {
		// A MISSING VALUE MUST NOT FALL THROUGH. `-tag` with nothing after it used to reach the
		// branch check below and exit 0 — the release job would have read that as "the tag and
		// the manifest agree" on a run where no tag was ever examined. That is the plausible
		// zero this repository keeps finding: the absent case and the healthy case, same bytes.
		if len(os.Args) < 3 || os.Args[2] == "" {
			die("-tag needs the tag to check; refusing rather than reporting a pass on a tag nobody named")
		}
		if err := releaseTagMatchesManifest(root, os.Args[2]); err != nil {
			die(err)
		}
		fmt.Printf("release tag %s matches the manifest\n", os.Args[2])
		return
	}
	problems, checked, err := check(root, base)
	if err != nil {
		die(err)
	}
	// The forward-motion check above asks whether THIS BRANCH moved the version. These two ask
	// whether the version is COHERENT at all — they read the tree, not the diff, so they fail on
	// a disagreement no matter which branch introduced it.
	sweepProblems := sweepUnclassified(root)
	problems = append(problems, sweepProblems...)
	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, p)
		}
		os.Exit(1)
	}
	fmt.Printf("plugin versions move forward: %d checked against %s\n", checked, base)
}

func die(v any) {
	fmt.Fprintln(os.Stderr, "version-guard:", v)
	os.Exit(1)
}
