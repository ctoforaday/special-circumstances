package fuzz

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE SHIPPED PROSE IS A SURFACE TOO, AND NOTHING WAS WALKING IT.
//
// Three sweeps cover the ways a retired command name reaches a reader, and until now they covered
// two: TestNoStringLiteralNamesARetiredSurface walks Go string literals — text a program EMITS at a
// seat — and the prompt gates walk skills/, agents/ and commands/. The root README and docs/ sit
// outside both.
//
// MEASURED 2026-08-22. README.md:67 described the projections as rendered on read "(`show --view
// …`)". `--view` stopped existing in 0.56.0 when `show` became a group. It had been wrong for
// sixteen releases, in the FIRST document anyone reads about this system, while three gates
// watched everything else.
//
// A human following it types a command that exits 2. That is the same failure the literal gate was
// built for — "a seat that reads one and follows it is refused for a mistake the tool made" — with
// a human in the seat.
func TestNoShippedDocNamesARetiredSurface(t *testing.T) {
	docs := shippedDocs(t)
	var hits []string
	for _, path := range docs {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(b), "\n") {
			// The retired list is deliberately shared with the literal gate — see
			// retiredsurfaces_test.go. One memory, two audiences.
			for _, re := range retiredSurfaces {
				if m := re.FindString(line); m != "" {
					hits = append(hits, filepath.Base(path)+":"+strconv.Itoa(i+1)+": "+strings.TrimSpace(m))
					break
				}
			}
		}
	}
	sort.Strings(hits)
	if len(hits) > 0 {
		t.Errorf("%d line(s) of shipped documentation name a surface that no longer exists:\n  %s\n\n"+
			"A reader who follows one types a command that exits 2. Fix the spelling; if the mention is\n"+
			"deliberately historical, say so in the sentence — this gate reads what a person is told to run.",
			len(hits), strings.Join(hits, "\n  "))
	}
}

// shippedDocs are the markdown a CONSUMER reads: the root README, the plugin READMEs, and docs/.
// It REFUSES an empty discovery, because every assertion above is a negative and a sweep that
// reads nothing reports every file clean.
func shippedDocs(t *testing.T) []string {
	t.Helper()
	var out []string
	for _, pattern := range [][]string{
		{"README.md"},
		{"plugins", "*", "README.md"},
		{"plugins", "frank-exchange-of-views", "docs", "*.md"},
	} {
		m, err := repotree.Glob(pattern...)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, m...)
	}
	if len(out) < 4 {
		t.Fatalf("found only %d shipped doc(s) — too few to be this repository's documentation", len(out))
	}
	return out
}
