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
	"extractorFallbackVersion": "a DEPENDENCY's version, not this repo's — the PDF extractor " +
		"identity that #629 D3 keys cached extractions on. The shipping binary never reads it: " +
		"extractorIdentity() takes the real version out of the linked module graph, and this " +
		"constant answers only where that graph is empty (a `go test` binary carries deps=0). " +
		"It is checked against go.mod itself by TestExtractorFallbackMatchesGoMod, so it cannot " +
		"drift from the module it names",
}

// THE PER-BINARY `const version` CLASS IS GONE, AND THAT IS WHY THERE IS NO ALLOWLIST HERE.
//
// There used to be one: 15 declarations across two plugins, each frozen at the value it was born
// with, so a binary shipped inside prosthetic-conscience 0.39.0 answered `-version` with 0.1.0.
// They were TOLERATED and counted on every run rather than denied, explicitly so the class could
// not grow while it waited for a decision (#405 item 3).
//
// The decision landed: `internal/buildid` reads the commit the running binary was built from —
// a fact the Go toolchain already stamps into every binary with no flags (#450) — and all 15
// constants were deleted. With the allowlist removed, a NEW `const version = "0.1.0"` now falls
// through to the unclassified branch below and FAILS, which is the correct answer: the fact is
// available from the build, so hand-writing it again is a regression rather than a carrier to
// register.

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
func sweepUnclassified(root string) []string {
	var problems []string
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
	return problems
}
