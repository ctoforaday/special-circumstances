package surface

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// A POSITION ASKS ONLY FOR WHAT ITS ADDRESSEE CAN READ.
//
// A position is rendered in one place — the transcript — and the transcript is on two surfaces: the
// bench's and the chair's. So every kind of content a seat is told to put in its position has to be
// addressed to a seat that holds that projection. Asked of the archive, the old wording failed it:
// chair positions carried instructions to the lenses and a synthesizer's carried one to red, each
// written, on the record, and read by no seat it addressed.
//
// THE TABLE IS THE CLAIM, AND THE RENDERED TEXT AND THE COMMAND TREE ARE WHAT IT IS HELD TO. Each
// row is one kind of content: the phrase that asks for it, the seat asked, and the seat it is for.
// The phrase is in ONE of that seat's carriers and not the other — the chair's constitution, because
// every chair sitting owes a position; blue's response prompt, because the lanes and the frontier
// share blue's constitution and owe none. The addressee's own command tree holds the transcript;
// and the number of rows for a seat is the number its sentence states ("two things"), so a third
// kind added to the sentence fails here until a row names its reader.
var positionKinds = []struct {
	seat, phrase, addressee string
	carrier, other          string
}{
	{"red-chair", "what you ask the bench to hold", "judge", "red-chair.md", "prompt-red-chair.golden"},
	{"red-chair", "why you left untaken a step the record offered you", "judge", "red-chair.md", "prompt-red-chair.golden"},
	{"blue-respond", "what you ask the bench to weigh", "judge", "prompt-blue-respond.golden", "blue-researcher.md"},
	{"blue-respond", "why you left untaken a route the record offered you", "judge", "prompt-blue-respond.golden", "blue-researcher.md"},
}

// surfacePath is where a carrier named in positionKinds lives: a prompt golden or an agent definition.
// Both reads drop the generated surface: the position verb's own page states what a position holds,
// and that page is the tool's, not a carrier of the duty.
func surfacePath(surface string) []string {
	if strings.HasSuffix(surface, ".golden") {
		return []string{"tests", "simulator", "testdata", surface}
	}
	return []string{"agents", surface}
}

// positionCount reads the count a position sentence states of itself.
var positionCount = regexp.MustCompile(`YOUR POSITION ARGUES TO THE BENCH\*{0,2}[^.]*\. It holds (\w+) things no act of the sitting holds`)

var countWords = map[string]int{"one": 1, "two": 2, "three": 3, "four": 4, "five": 5}

func TestAPositionAsksOnlyForWhatItsAddresseeCanRead(t *testing.T) {
	read := func(parts ...string) string {
		t.Helper()
		path, err := repotree.Plugin(parts...)
		if err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return strings.Join(strings.Fields(string(b)), " ")
	}
	holdsTranscript := func(seatID string) bool {
		for _, g := range cli.NewRootFor(seatID).Commands() {
			if g.Name() != seat.GroupOf("debate") {
				continue
			}
			for _, v := range g.Commands() {
				if v.Name() == "debate" {
					return true
				}
			}
		}
		return false
	}

	rows := map[string]int{}
	carriers := map[string]string{}
	for _, k := range positionKinds {
		rows[k.seat]++
		carriers[k.seat] = k.carrier
		if !record.SeatOwesPosition(k.seat) {
			t.Errorf("%s is asked for %q and owes no position", k.seat, k.phrase)
		}
		if !strings.Contains(withoutGeneratedSurface(read(surfacePath(k.carrier)...)), k.phrase) {
			t.Errorf("%s: %s is not asked for %q there — the table names a kind of content the surface does not ask for", k.carrier, k.seat, k.phrase)
		}
		if strings.Contains(withoutGeneratedSurface(read(surfacePath(k.other)...)), k.phrase) {
			t.Errorf("%s: %s is asked for %q there as well as in %s — a seat holds each duty in one carrier", k.other, k.seat, k.phrase, k.carrier)
		}
		if !holdsTranscript(k.addressee) {
			t.Errorf("%s is told to put %q in its position for %s, whose surface holds no transcript: it is written, and read by no seat it addresses", k.seat, k.phrase, k.addressee)
		}
	}

	// THE SENTENCE'S OWN COUNT IS THE NUMBER OF ROWS, on its carrier.
	for seatID, n := range rows {
		surface := carriers[seatID]
		m := positionCount.FindStringSubmatch(read(surfacePath(surface)...))
		if m == nil {
			t.Errorf("%s: no position sentence stating its count — the matcher reads nothing, so a kind added there is unheld", surface)
			continue
		}
		if countWords[m[1]] != n {
			t.Errorf("%s: the position sentence asks %s for %q things and the table names %d reader(s) — each kind of content needs a row naming the seat that reads it", surface, seatID, m[1], n)
		}
	}

	// EVERY SEAT THAT OWES A POSITION IS IN THE TABLE, and no seat that reads no transcript is
	// named as a reader by it: the two limits the sentences state ("instructs no lens", "answers
	// no lens") are true of the tree.
	for _, seatID := range []string{"red-chair", "blue-respond"} {
		if rows[seatID] == 0 {
			t.Errorf("%s owes a position and the table holds no kind of content for it", seatID)
		}
	}
	for _, seatID := range []string{"red-lens-logic", "blue-respond", "blue-lane-1", "frontier", "blue-synthesize"} {
		if holdsTranscript(seatID) {
			t.Errorf("%s holds the transcript: the position sentences say no lens and no blue seat reads a position, and the tree disagrees", seatID)
		}
	}
}
