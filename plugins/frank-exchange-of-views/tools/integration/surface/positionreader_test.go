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
// The phrase is in that seat's dispatch and in its constitution; the addressee's own command tree
// holds the transcript; and the number of rows for a seat is the number its sentence states ("two
// things"), so a third kind added to the sentence fails here until a row names its reader.
var positionKinds = []struct {
	seat, phrase, addressee string
	prompt, constitution    string
}{
	{"red-chair", "what you ask the bench to hold", "judge", "prompt-red-chair.golden", "red-chair.md"},
	{"red-chair", "why you left untaken a step the record offered you", "judge", "prompt-red-chair.golden", "red-chair.md"},
	{"blue-respond", "what you ask the bench to weigh", "judge", "prompt-blue-respond.golden", "blue-researcher.md"},
	{"blue-respond", "why you left untaken a route the record offered you", "judge", "prompt-blue-respond.golden", "blue-researcher.md"},
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
	carriers := map[string][2]string{}
	for _, k := range positionKinds {
		rows[k.seat]++
		carriers[k.seat] = [2]string{k.prompt, k.constitution}
		if !record.SeatOwesPosition(k.seat) {
			t.Errorf("%s is asked for %q and owes no position", k.seat, k.phrase)
		}
		for surface, text := range map[string]string{
			k.prompt:       read("tests", "simulator", "testdata", k.prompt),
			k.constitution: read("agents", k.constitution),
		} {
			if !strings.Contains(text, k.phrase) {
				t.Errorf("%s: %s is not asked for %q there — the table names a kind of content the surface does not ask for", surface, k.seat, k.phrase)
			}
		}
		if !holdsTranscript(k.addressee) {
			t.Errorf("%s is told to put %q in its position for %s, whose surface holds no transcript: it is written, and read by no seat it addresses", k.seat, k.phrase, k.addressee)
		}
	}

	// THE SENTENCE'S OWN COUNT IS THE NUMBER OF ROWS, on both carriers.
	for seatID, n := range rows {
		for _, surface := range carriers[seatID] {
			dir := []string{"agents", surface}
			if strings.HasSuffix(surface, ".golden") {
				dir = []string{"tests", "simulator", "testdata", surface}
			}
			m := positionCount.FindStringSubmatch(read(dir...))
			if m == nil {
				t.Errorf("%s: no position sentence stating its count — the matcher reads nothing, so a kind added there is unheld", surface)
				continue
			}
			if countWords[m[1]] != n {
				t.Errorf("%s: the position sentence asks %s for %q things and the table names %d reader(s) — each kind of content needs a row naming the seat that reads it", surface, seatID, m[1], n)
			}
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
