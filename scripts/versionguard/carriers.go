package main

// The per-binary version constants, and a gate that stops the class GROWING.
//
// # What this is, and what it deliberately is not
//
// It is not a guard over two hand-written copies of a version agreeing. That is what this
// file did first, and it was wrong for a reason worth writing down: `const Version` and
// requirements.json's recordToolVersion ALREADY AGREE, and the reason releases are
// unreachable is that nothing has tagged since 0.50.0 (#405). Guarding the agreement policed
// a thing that was not broken while the broken thing sat next to it — `facts-are-fields`
// clause 2, "before blaming the format, find what actually produced the number", driven past
// in the same change that cited the rule.
//
// The right fix for two carriers of one fact is to GENERATE one from the other and gate
// staleness, the way scripts/golden does. A guard is what you build when generation is
// impossible; it was not impossible here, and a guard whose own allowlist is hand-kept has
// reproduced the defect one level up.
//
// # What survived, because it stands on its own evidence
//
// Fifteen per-binary `const version` declarations across two plugins, every one frozen at the
// value it was born with — 0.1.0 or 0.2.0 — while the plugins ship at 0.37.0 and 0.7.1. Only
// two are reachable at all (`-version` is parsed in secretsgate and toolchainnudge); the rest
// assert a version nothing can read, which is why nobody noticed.
//
// That is a known defect class with a count. This file exists to keep it from growing: the
// fifteen are listed, counted and PRINTED on every run, and a sixteenth FAILS. No doctrine
// required — a bug class that cannot grow while it waits for its real fix is worth a gate on
// its own terms.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// notVersions are version-SHAPED declarations that are not a version of anything shipped,
// each with the reason. Being on this list is a decision somebody made; being on neither list
// is a failure.
var notVersions = map[string]string{
	"MassMappingVersion": "a pinned SEMANTICS version for the mass mapping — changing it is a " +
		"schema change, and it is deliberately not tied to any release",
	"fetchUAVersion": "the User-Agent this tool sends; it identifies the fetcher to a server " +
		"and has nothing to do with what is shipped",
	"ToolVersion": "the mutable variable const Version is assigned INTO at init; the constant " +
		"is the source, this is the destination",
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
			if name == "Version" && strings.HasSuffix(rel, "/internal/cli/root.go") {
				// feov-record's own version, the one requirements.json restates. Derivation
				// is the fix (#405); until then it is not a stale-at-birth constant and does
				// not belong in the class below.
				continue
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
