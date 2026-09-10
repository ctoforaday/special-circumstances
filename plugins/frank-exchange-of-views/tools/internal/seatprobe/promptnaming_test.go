package seatprobe

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE PROMPTS NAME NO VERB EITHER, AND NOW SOMETHING HOLDS THEM TO IT.
//
// TestTheShippedConstitutionsNameNoVerb holds the four agent .md files to this rule, on a
// measured finding: naming a slice of the surface does not under-inform a seat, it SATISFIES it —
// 58% surface exposure against 95% with the names removed. The prompt engine is the other half of
// what a seat reads, and nothing held it to anything.
//
// Measured 2026-09-07 before writing this: the emitted prompts already name zero verbs. Every
// hit in debate.js is a `//` comment, and the bare occurrences in prompt text are ordinary
// English ("a successor you mint"). So this pins a property that HOLDS rather than fixing one
// that does not — which is the whole reason to write it now, while the answer is zero and the
// diff is empty. `debate.test.mjs` carries 307 assertions and not one of them is this rule; every
// prompt property there is a hand-written string match, which is the shape that let
// /`opinion` cannot carry it/ go on passing against a verb that had been deleted.
//
// # Why the goldens rather than the running engine
//
// The prompt goldens ARE the emitted prompts — prompts.test.mjs records each seat class's prompt
// in full precisely so a wave's effect arrives as a reviewable diff. Reading them here means the
// detector runs against the recorded product surface with no second harness, no goja, and no
// verb list restated in JavaScript. A prompt that gains a verb name updates its golden, and this
// fails on the same commit.
//
// It also means the check lives where the detector does. The alternative — asserting this in
// debate.test.mjs — needs the tool's command surface in JS, which is either a hand-kept list (the
// thing facts-are-fields refuses) or a generator for a property that needs neither.
func TestTheSeatPromptsNameNoVerb(t *testing.T) {
	dir, err := repotree.Plugin("tests", "simulator", "testdata")
	if err != nil {
		t.Fatal(err)
	}
	goldens, err := filepath.Glob(filepath.Join(dir, "prompt-*.golden"))
	if err != nil {
		t.Fatal(err)
	}
	// ANTI-VACUITY, and it is not decoration: this test reads files by glob, so a rename of the
	// golden convention would leave it sweeping an empty set and reporting a clean surface in the
	// same words it uses for a real one. That is the exact defect the suite exists to catch, and
	// a check that walks nothing is the easiest way to write it.
	if len(goldens) < 10 {
		t.Fatalf("found %d prompt goldens under %s — this check has stopped reading the prompts and "+
			"would now pass over any of them", len(goldens), dir)
	}

	sf := NewSurface(cli.CommandPaths())
	for _, g := range goldens {
		b, err := os.ReadFile(g)
		if err != nil {
			t.Fatal(err)
		}
		named := NamesSurviving(string(b), sf)
		if len(named) == 0 {
			continue
		}
		var left []string
		for v, n := range named {
			left = append(left, v+"×"+itoa(n))
		}
		sort.Strings(left)
		t.Errorf("%s names %d verb(s): %s\n\nThe prompt is not where a seat learns the tool's "+
			"vocabulary — `<verb> --help` is, generated from the command that enforces it, so it "+
			"cannot offer a flag the write path refuses. A prompt naming a slice of the surface "+
			"SATISFIES a seat rather than under-informing it (58%% exposure against 95%%, measured "+
			"2026-08-15). Name the ACT; let the manual — every command's own help page — name the verb.",
			filepath.Base(g), len(named), strings.Join(left, ", "))
	}
}
