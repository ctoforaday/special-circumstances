package main

// A VERSION IS A FACT WITH MORE THAN ONE CARRIER, and nothing checked that the carriers agreed.
//
// This repository holds two version facts per plugin, not one, and the distinction is
// deliberate — an earlier reading of this as "four numbers that drifted" was wrong (#405):
//
//	TOOL version    tools/internal/cli/root.go `const Version`   +   requirements.json `recordToolVersion`
//	PLUGIN version  .claude-plugin/plugin.json `version`         +   the git tag <plugin>--v<version>
//
// The tool line is documented by a 67-entry changelog in root.go, and a `register` event
// stamping the TOOL version is right: it is a fact about the binary that wrote the record.
//
// What is NOT right is that each fact is typed into two files by hand. They happen to agree
// today because the change that moved them moved both; nothing makes the next one do so, and a
// disagreement is invisible until `setup` refuses a run at the preflight — in the field, on
// someone else's machine.
//
// # Why this file also SWEEPS
//
// A guard with a hand-kept list of what to check drifts exactly like the list it replaces —
// which is the whole finding this repo has been working through. So `carriers` is the
// allowlist, `notVersions` is the explicit denial, and the sweep FAILS on any version-shaped
// declaration in neither: a new one has to be classified by a person, not defaulted into
// invisibility. Same shape as mjsparity and scripts/check's parity test, both written after
// the same class recurred.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// carrier is one place a version fact is written down.
type carrier struct {
	path string         // repo-relative, under plugins/<plugin>/
	re   *regexp.Regexp // must capture the version in group 1
}

// versionFact is one fact and every carrier that restates it.
type versionFact struct {
	name     string
	plugin   string
	carriers []carrier
	why      string
}

var (
	reGoConstVersion  = regexp.MustCompile(`(?m)^const Version = "([^"]+)"`)
	reRecordToolVer   = regexp.MustCompile(`"recordToolVersion"\s*:\s*"([^"]+)"`)
	reManifestVersion = regexp.MustCompile(`"version"\s*:\s*"([^"]+)"`)
)

// facts enumerates what must agree. Adding a plugin with a versioned tool means adding it
// here — and the sweep below is what makes forgetting to a failure rather than a silence.
var facts = []versionFact{{
	name:   "feov-record tool version",
	plugin: "frank-exchange-of-views",
	carriers: []carrier{
		{"plugins/frank-exchange-of-views/tools/internal/cli/root.go", reGoConstVersion},
		{"plugins/frank-exchange-of-views/requirements.json", reRecordToolVer},
	},
	why: "setup preflights the running binary against requirements.json; a disagreement refuses " +
		"the run at dispatch, in the field, rather than here",
}}

// notVersions are version-SHAPED declarations that are not a version fact, each with the
// reason. Being on this list is a decision somebody made; being on neither list is a failure.
var notVersions = map[string]string{
	"MassMappingVersion": "a pinned SEMANTICS version for the mass mapping — changing it is a " +
		"schema change, and it is deliberately not tied to any release",
	"fetchUAVersion": "the User-Agent this tool sends; it identifies the fetcher to a server " +
		"and has nothing to do with what is shipped",
	"ToolVersion": "the mutable variable const Version is assigned INTO at init; the constant " +
		"is the carrier, this is the destination",
}

// staleBinaryVersions are the per-binary `const version` declarations, one per hook binary,
// each frozen at the value it was born with. They are NOT denied as "not a version" — they are
// versions, and the value they assert is wrong: a binary shipped inside prosthetic-conscience
// 0.37.0 answering `-version` with 0.1.0 tells an operator diagnosing a bad hook the one thing
// they cannot use.
//
// MEASURED 2026-08-15: 15 of them across two plugins, at 0.1.0 or 0.2.0, in plugins at 0.37.0
// and 0.7.1. Only two are even reachable (`-version` is parsed in secretsgate and
// toolchainnudge); the rest assert a version nothing can read, which is why nobody noticed.
//
// TOLERATED, NOT HIDDEN, and the distinction is the point. Failing on all 15 would block every
// pull request until the version is derived at build time (#405), which is a change with a
// decision in front of it. So they are counted and printed on every run, and a SIXTEENTH one
// fails — the class cannot grow while it waits.
var staleBinaryVersions = map[string]bool{
	"plugins/gray-area/tools/cmd/gray-area-capture/main.go":                   true,
	"plugins/gray-area/tools/cmd/gray-area/main.go":                           true,
	"plugins/prosthetic-conscience/tools/internal/checkpointrestore/main.go":  true,
	"plugins/prosthetic-conscience/tools/internal/checkpointseal/main.go":     true,
	"plugins/prosthetic-conscience/tools/internal/doctor/main.go":             true,
	"plugins/prosthetic-conscience/tools/internal/filechangedrearm/main.go":   true,
	"plugins/prosthetic-conscience/tools/internal/postcompactobserve/main.go": true,
	"plugins/prosthetic-conscience/tools/internal/posttooluse/main.go":        true,
	"plugins/prosthetic-conscience/tools/internal/pretooluse/main.go":         true,
	"plugins/prosthetic-conscience/tools/internal/pushfreezeguard/main.go":    true,
	"plugins/prosthetic-conscience/tools/internal/qualitygate/main.go":        true,
	"plugins/prosthetic-conscience/tools/internal/secretsgate/main.go":        true,
	"plugins/prosthetic-conscience/tools/internal/sessionstart/main.go":       true,
	"plugins/prosthetic-conscience/tools/internal/strikecounter/main.go":      true,
	"plugins/prosthetic-conscience/tools/internal/toolchainnudge/main.go":     true,
}

// reVersionShaped finds declarations that look like a version so the sweep can demand each be
// classified. Deliberately broad: a false positive costs one line on a list, and a false
// NEGATIVE is a carrier nobody is checking.
//
// THE `const`/`var` KEYWORD IS OPTIONAL, and that is the whole point. The first version of this
// required it, which meant it could only see SINGLE declarations — every name inside a grouped
// `const ( … )` block was invisible, and grouped blocks are the idiomatic form. It found two
// declarations and missed `fetchUAVersion`, which sits in exactly such a block. A sweep blind to
// a whole syntactic form reports a clean board for the half of the tree it cannot see.
//
// Bare `Name = "v"` inside a block is matched by allowing the keyword to be absent. A qualified
// assignment (`x.Version = "y"`) is excluded by anchoring the identifier to the start of the
// declaration, and `:=` by requiring a plain `=`.
//
// AND THE PREFIX IS OPTIONAL TOO. It was not, which meant the identifier had to have at least
// one character before "Version" — so bare `const Version`, the single most important carrier
// in the tree, never matched. The live run still passed, because a later branch skips that name
// as an already-checked carrier; the skip was dead code hiding the hole. A test asserting the
// pattern against the forms that actually occur is what found it.
var reVersionShaped = regexp.MustCompile(`(?m)^\s*(?:(?:const|var)\s+)?((?:[A-Za-z][A-Za-z0-9_]*)?[Vv]ersion[A-Za-z0-9_]*)\s*=\s*"([^"]*)"`)

func readCarrier(root string, c carrier) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, c.path))
	if err != nil {
		return "", fmt.Errorf("%s: %w", c.path, err)
	}
	m := c.re.FindSubmatch(b)
	if m == nil {
		// A carrier that stopped matching is NOT "no problem found" — it is a carrier this
		// guard can no longer see, which is the failure mode it exists to prevent.
		return "", fmt.Errorf("%s: no version matched %q — the declaration moved or changed shape, "+
			"so this carrier is no longer being checked", c.path, c.re)
	}
	return string(m[1]), nil
}

// checkCarriers reports any fact whose carriers disagree.
func checkCarriers(root string) []string {
	var problems []string
	for _, f := range facts {
		seen := map[string][]string{}
		for _, c := range f.carriers {
			v, err := readCarrier(root, c)
			if err != nil {
				problems = append(problems, fmt.Sprintf("%s: %v", f.name, err))
				continue
			}
			seen[v] = append(seen[v], c.path)
		}
		if len(seen) <= 1 {
			continue
		}
		var lines []string
		for v, paths := range seen {
			sort.Strings(paths)
			lines = append(lines, fmt.Sprintf("    %s = %s", strings.Join(paths, ", "), v))
		}
		sort.Strings(lines)
		problems = append(problems, fmt.Sprintf(
			"%s: carriers DISAGREE — one fact, %d different values:\n%s\n  %s",
			f.name, len(seen), strings.Join(lines, "\n"), f.why))
	}
	return problems
}

// sweepUnclassified fails on any version-shaped declaration that is neither a known carrier nor
// explicitly denied. This is what keeps `facts` from quietly falling behind the tree.
//
// It returns the tolerated count alongside the problems so the caller can PRINT it. A tolerated
// item that is never mentioned again is a denied item wearing a different word.
func sweepUnclassified(root string) []string {
	problems, _ := sweepWithTolerated(root)
	return problems
}

func sweepWithTolerated(root string) ([]string, int) {
	known := map[string]bool{}
	for _, f := range facts {
		for _, c := range f.carriers {
			known[filepath.ToSlash(c.path)] = true
		}
	}
	var problems []string
	tolerated := 0
	err := filepath.WalkDir(filepath.Join(root, "plugins"), func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		for _, m := range reVersionShaped.FindAllSubmatch(b, -1) {
			name := string(m[1])
			if notVersions[name] != "" {
				continue
			}
			if known[rel] && name == "Version" {
				continue // a declared carrier, already checked by checkCarriers
			}
			if name == "version" && staleBinaryVersions[rel] {
				tolerated++
				continue
			}
			problems = append(problems, fmt.Sprintf(
				"%s:%s = %q is version-shaped and CLASSIFIED NOWHERE.\n"+
					"  Add it to `facts` if it carries a version somebody else also writes down, or to "+
					"`notVersions` with the reason it is not one. Defaulting it into invisibility is how "+
					"the carriers stopped agreeing in the first place.", rel, name, string(m[2])))
		}
		return nil
	})
	if err != nil {
		problems = append(problems, fmt.Sprintf("sweeping for version declarations: %v", err))
	}
	sort.Strings(problems)
	return problems, tolerated
}

// pluginVersionOf reads a plugin's manifest version, for the tag check below.
func pluginVersionOf(root, plugin string) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, "plugins", plugin, ".claude-plugin", "plugin.json"))
	if err != nil {
		return "", err
	}
	var m struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(b, &m) != nil || m.Version == "" {
		return "", fmt.Errorf("no version in %s's manifest", plugin)
	}
	return m.Version, nil
}
