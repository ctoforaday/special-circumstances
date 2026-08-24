package record

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// THE MASS MAPPING LIVES IN TWO LANGUAGES, AND NOTHING MADE THEM AGREE.
//
// debate.js scores gaps to drive the engine's own decisions; this package scores them for the
// board and the telemetry. Same concept, two hand-written tables, no gate — and they carried
// DIFFERENT VERSION STAMPS for six releases (`v1` here, `v2` there) while a comment in this
// package explained the difference as real semantics deferred under an oracle freeze.
//
// It was not real. Measured 2026-08-16: eight keys, eight values, identical on both sides. The
// stamps disagreed about nothing, and every telemetry line said `v1` while the engine that
// produced the run called that mapping `v2` — so anything joining runs by mapping_version split
// one population in two. PR0 aligned the stamp; this test is what stops the pair drifting again.
//
// WHY A GUARD AND NOT GENERATION, because facts-are-fields requires that answer. Generating one
// side from the other is the better fix and it is not available: debate.js is loaded by node and
// by goja at runtime, with no build step in between, so there is nowhere to generate INTO. A
// generated file would have to be committed and kept fresh by exactly the check written here,
// one level up and with more moving parts. The guard is what you build when generation is
// impossible; this comment is the required statement of why it was.
//
// The parse is deliberately narrow. It reads the two declarations by shape and FAILS LOUDLY if
// it cannot find them — a regex that silently matches nothing would report agreement between a
// table and an empty map, which is the plausible zero this whole area exists to remove.

const debateJS = "../../../skills/research-protocol/scripts/debate.js"

var (
	reMassVersion = regexp.MustCompile(`(?m)^const MASS_MAPPING_VERSION\s*=\s*'([^']*)'`)
	reMassTable   = regexp.MustCompile(`(?m)^const MASS\s*=\s*\{([^}]*)\}`)
	// ANCHORED ON THE SEPARATOR, and the key charset includes `_`.
	//
	// It was `'?([a-z-]+)'?\s*:\s*([0-9.]+)` — letters and a HYPHEN, no underscore — so
	// `'low_medium': 1.5` did not fail to match. It matched the TAIL: `medium` : 1.5. The parse
	// then carried `medium=1.5` until the real `medium: 2` overwrote it, and reported
	// `low_medium` and `medium_high` ABSENT from a file that declares both. A regex that
	// half-matches is worse than one that misses: the miss is a hole, the half-match is a wrong
	// answer wearing a right one's clothes.
	//
	// Anchoring on `{` or `,` means a key must start where a key can start, so a partial match is
	// not a match at all.
	reMassEntry = regexp.MustCompile(`[{,]\s*'?([a-z_]+)'?\s*:\s*([0-9.]+)`)
)

func readDebateJS(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(debateJS)
	if err != nil {
		t.Fatalf("cannot read the engine at %s: %v\n\nThis test binds the Go mass mapping to the "+
			"JavaScript one. If debate.js moved, POINT THIS TEST AT IT rather than deleting the "+
			"test — the two tables drifted for six releases the last time nothing checked.", debateJS, err)
	}
	return string(b)
}

// TestMassMappingVersionMatchesTheEngine is the stamp half.
func TestMassMappingVersionMatchesTheEngine(t *testing.T) {
	m := reMassVersion.FindStringSubmatch(readDebateJS(t))
	if m == nil {
		t.Fatalf("no `const MASS_MAPPING_VERSION = '...'` in %s — the declaration this test reads "+
			"has changed shape. A no-match here must FAIL, never pass quietly: the whole point is "+
			"that a silent miss reads exactly like agreement.", debateJS)
	}
	if m[1] != MassMappingVersion {
		t.Errorf("mass mapping stamp disagrees: Go says %q, debate.js says %q.\n\n"+
			"These name the SAME mapping and must move together. They were %q and %q for six "+
			"releases while the tables underneath were identical, which put a meaningless split "+
			"through every telemetry join keyed on mapping_version.",
			MassMappingVersion, m[1], "v1", "v2")
	}
}

// TestMassTableMatchesTheEngine is the half that actually protects the scores.
//
// The stamp is a label; THIS is the thing a drift would corrupt. GapMass multiplies two entries
// out of the table below, so a single changed value silently re-ranks every gap in every run,
// and a key present on one side only makes that grade score zero on the other — the absent-reads-
// as-harmless failure that made likelihood and impact required fields in the first place.
func TestMassTableMatchesTheEngine(t *testing.T) {
	tbl := reMassTable.FindStringSubmatch(readDebateJS(t))
	if tbl == nil {
		t.Fatalf("no `const MASS = { ... }` in %s — see the sibling test on why this fails rather "+
			"than skips.", debateJS)
	}
	// The table body is comma-separated, so the entry count is knowable independently of the
	// pattern that reads them. A parse that finds FEWER is silently skipping a key, which is the
	// defect the anchor above exists to prevent — and it would otherwise be reported as a
	// divergence in the mapping rather than as a broken read.
	if want := strings.Count(tbl[1], ",") + 1; len(reMassEntry.FindAllStringSubmatch("{"+tbl[1], -1)) != want {
		t.Fatalf("the MASS table in %s has %d comma-separated entries and the parse found %d — "+
			"a partial parse reports the keys it could not see as ABSENT, which reads as a real "+
			"divergence between the two tables", debateJS, want, len(reMassEntry.FindAllStringSubmatch("{"+tbl[1], -1)))
	}
	entries := reMassEntry.FindAllStringSubmatch("{"+tbl[1], -1)
	if len(entries) == 0 {
		t.Fatalf("`const MASS` in %s parsed to ZERO entries. An empty parse compares equal to "+
			"nothing and would report success; failing instead.", debateJS)
	}

	js := map[string]float64{}
	for _, e := range entries {
		v, err := strconv.ParseFloat(e[2], 64)
		if err != nil {
			t.Fatalf("mass value %q for key %q in %s is not a number: %v", e[2], e[1], debateJS, err)
		}
		js[e[1]] = v
	}

	if diff := describeMassDiff(MASS, js); diff != "" {
		t.Errorf("the Go and JavaScript mass mappings have diverged:\n%s\n\n"+
			"GapMass multiplies two of these per gap, so a changed value re-ranks every gap in "+
			"every run and a missing key scores that grade as ZERO — which reads as a harmless "+
			"gap rather than an ungraded one. Change both tables, and bump MassMappingVersion.", diff)
	}
}

func describeMassDiff(goSide, jsSide map[string]float64) string {
	keys := map[string]bool{}
	for k := range goSide {
		keys[k] = true
	}
	for k := range jsSide {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var out []string
	for _, k := range sorted {
		g, inGo := goSide[k]
		j, inJS := jsSide[k]
		switch {
		case inGo && !inJS:
			out = append(out, fmt.Sprintf("  %-12s Go=%v  debate.js=ABSENT", k, g))
		case !inGo && inJS:
			out = append(out, fmt.Sprintf("  %-12s Go=ABSENT  debate.js=%v", k, j))
		case g != j:
			out = append(out, fmt.Sprintf("  %-12s Go=%v  debate.js=%v", k, g, j))
		}
	}
	return strings.Join(out, "\n")
}

// TestMassKeysAreGrades holds the table to the canonical grade set, so a new grade cannot be
// added to one vocabulary and forgotten in the other. `realized` is the documented exception:
// it is a grade that contributes zero mass by design, not a missing entry.
func TestMassKeysAreGrades(t *testing.T) {
	for k := range MASS {
		if !isGrade(k) {
			t.Errorf("MASS carries %q, which is not in the canonical grade set (flags.Grades) — "+
				"a mass entry nothing can be graded as is dead weight, and a grade with no mass "+
				"entry scores zero", k)
		}
	}
}
