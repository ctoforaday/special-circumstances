package surface

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// EVERY PROJECTION A PROMPT SENDS A SEAT TO IS ON THAT SEAT'S SURFACE.
//
// TestEveryViewNamedInAPromptExists asks whether a named view exists ANYWHERE. That is not the
// question a seat meets: `debate`, `motions` and `telemetry` exist, on the bench's and the chair's
// own read group, and a prompt that sends blue to one of them names a real view blue cannot open.
// Wave A (2026-10-09): blue-respond was told to "pull your working set — the board and work
// projections and the transcript", found no such read among its own, and went looking for a
// `debate.md` at the run root. The views had moved off a working seat's surface and the prompt
// went on sending blue to them.
//
// IT READS THE RENDERED PROMPTS, each of which states the seat it is for, and asks that seat's
// real command tree — never a copy of which role holds which group.
//
// TWO MATCHERS, AND THE SECOND IS NARROW ON PURPOSE. A prompt names an act, never a command, so a
// view arrives in prose. One idiom is regular — "the board, work and motions projections", "the
// `avenues` projection" — and is read in full. The transcript is the exception: it is the one
// view a prompt calls by a word that is not its name, and that word also appears where nothing is
// being read ("your closings in the transcript"). So the gate looks for the word in
// a sentence that tells the seat to READ or PULL, and nowhere else; a prompt that sent a seat to
// it in other words would pass. Both matchers are held to a floor below so that a reworded idiom
// fails this test instead of emptying it.
var (
	promptProjections = regexp.MustCompile("((?:`?[a-z]+`?(?:, | and )?)+) projections?\\b")
	promptSeatID      = regexp.MustCompile(`SEAT_ID: ([a-z0-9-]+)`)
	// The sentence is bounded by full stops and by the colon a clause's heading ends in.
	readsTranscript = regexp.MustCompile(`(?i)\b(?:read|re-read|pull)\b[^.:]*\btranscript\b|\btranscript\b[^.:]*\b(?:in full|for context)\b`)
)

func TestEveryProjectionAPromptNamesIsOnThatSeatsSurface(t *testing.T) {
	real := map[string]bool{}
	for _, v := range cli.ViewNames() {
		real[v] = true
	}
	goldens, err := repotree.Glob("plugins", "frank-exchange-of-views", "tests", "simulator", "testdata", "prompt-*.golden")
	if err != nil {
		t.Fatal(err)
	}

	holds := func(root *cobra.Command, view string) bool {
		for _, g := range root.Commands() {
			if g.Name() != seat.GroupOf(view) {
				continue
			}
			for _, v := range g.Commands() {
				if v.Name() == view {
					return true
				}
			}
		}
		return false
	}

	var missing []string
	named, transcriptReaders := 0, 0
	for _, path := range goldens {
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := strings.Join(strings.Fields(string(b)), " ")
		// THE SEAT IS THE ONE THE PROMPT STATES. A repair prompt states none — it goes to a seat
		// already sitting — and is named for that seat instead. Either way the tree must be a
		// seat's: an id the roster does not know gets a tree with no read group, and every view
		// would then be "missing" for a reason that has nothing to do with the prompt.
		seatID := strings.TrimSuffix(strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), "prompt-"), ".golden"), "-sitting-record")
		if m := promptSeatID.FindStringSubmatch(text); m != nil {
			seatID = m[1]
		}
		root := cli.NewRootFor(seatID)
		if !holds(root, "work") {
			t.Errorf("%s: %q is no seat with a read surface — the prompt's seat cannot be asked, and skipping it would pass it unread", filepath.Base(path), seatID)
			continue
		}

		for _, list := range promptProjections.FindAllStringSubmatch(text, -1) {
			for _, word := range regexp.MustCompile(`[a-z]+`).FindAllString(list[1], -1) {
				if !real[word] {
					continue // "your", "the", "and": the list's own grammar
				}
				named++
				if !holds(root, word) {
					missing = append(missing, filepath.Base(path)+": tells "+seatID+" to read the "+word+" projection, which is not on its surface")
				}
			}
		}
		if readsTranscript.MatchString(text) {
			transcriptReaders++
			if !holds(root, "debate") {
				missing = append(missing, filepath.Base(path)+": tells "+seatID+" to read the transcript — "+
					readsTranscript.FindString(text)+" — and the debate projection is not on its surface")
			}
		}
	}

	// THE FLOORS. Six goldens name a projection by the idiom today and three send the bench to the
	// transcript; a count of zero means the wording moved and this gate is reading nothing.
	if named < 4 {
		t.Errorf("the projection idiom matched %d view name(s) across %d rendered prompts — the prompts name more than that, so the matcher has stopped reading them", named, len(goldens))
	}
	if transcriptReaders == 0 {
		t.Error("no rendered prompt sends any seat to the transcript — the bench's does, so the matcher has stopped reading them")
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("%d read(s) a prompt asks for that the seat cannot make. The seat is refused, guesses a file, and carries on without it:\n  %s\n\n"+
			"Say where that seat DOES read what the sentence is about, or drop the read. Whether the seat should hold the projection is a ruling, not a prompt edit.",
			len(missing), strings.Join(missing, "\n  "))
	}
}
