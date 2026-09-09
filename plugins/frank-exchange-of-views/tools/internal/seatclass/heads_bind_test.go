package seatclass

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// THE NEEDLES ARE BOUND TO THE PROMPTS THE ENGINE RENDERS, not to a hand-typed sample of them.
//
// This table classified a transcript by heads such as "Red audit, round N" for weeks after the
// engine stopped writing them: every live prompt classified as `other`, the tier audit reported
// "seat could not be identified" as a WARN, and the unit test kept passing because its fixtures
// were the old heads. The simulator keeps a rendered golden per seat prompt, named for the seat;
// classifying each one is the check that the needles still describe what the engine writes.
func TestEveryRenderedPromptHeadClassifiesToItsSeat(t *testing.T) {
	dir, err := repotree.Plugin("tests", "simulator", "testdata")
	if err != nil {
		t.Fatal(err)
	}
	goldens, err := filepath.Glob(filepath.Join(dir, "prompt-*.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if len(goldens) < 8 {
		t.Fatalf("found %d rendered prompt goldens, expected at least 8 — the glob is reading the wrong place, and an empty set passes on nothing", len(goldens))
	}
	for _, g := range goldens {
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(g), "prompt-"), ".golden")
		want := seatOfGolden(name)
		b, err := os.ReadFile(g)
		if err != nil {
			t.Fatal(err)
		}
		// The golden's first line is the prompt's sha256; the prompt begins after it.
		body := string(b)
		if i := strings.Index(body, "\n"); i >= 0 {
			body = body[i+1:]
		}
		head := body
		if len(head) > 2000 {
			head = head[:2000]
		}
		if got := ClassifySeat(head).Seat; got != want {
			t.Errorf("%s: the rendered prompt classifies as %q, want %q — the head reads %q", filepath.Base(g), got, want, strings.TrimSpace(head[:min(80, len(head))]))
		}
	}
}

// seatOfGolden maps a golden's name to the seat class this table reports for it: the goldens are
// named for the SEAT INSTANCE (red-lens-evidence, blue-lane-1, red-lens-evidence-engaged) and the
// classifier reports the seat KIND (red-lens, blue-lane).
func seatOfGolden(name string) string {
	switch {
	case strings.HasPrefix(name, "red-lens-"):
		return "red-lens"
	case strings.HasPrefix(name, "blue-lane-"):
		return "blue-lane"
	case name == "judge-terminal":
		return "judge-terminal"
	}
	return name
}
