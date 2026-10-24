package surface

import (
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/cli/seat"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
)

// A POSITION ASKS ONLY FOR WHAT ITS ADDRESSEE CAN READ.
//
// A position is rendered in one place — the transcript — and the transcript is on two surfaces: the
// bench's and the chair's. So every kind of content the chair is told to put in its position has to
// be addressed to a seat that holds that projection. Asked of the archive, an older wording failed
// it: chair positions carried instructions to the lenses, written, on the record, and read by no
// seat they addressed.
//
// THE TABLE IS THE CLAIM, AND THE RENDERED TEXT AND THE COMMAND TREE ARE WHAT IT IS HELD TO. Each
// row is one kind of content: the phrase that asks for it, the seat asked, and the seat it is for.
// The phrase is in ONE of that seat's carriers and not the other — the chair's constitution, which
// carries the duty, and not its prompt. The addressee's own command tree holds the transcript, and
// the position verb is on the chair's tree and no other: a lens and blue argue a gap in their
// closings.
var positionKinds = []struct {
	seat, phrase, addressee string
	carrier, other          string
}{
	{"red-chair", "it says why you withheld it", "judge", "red-chair.md", "prompt-red-chair.golden"},
	{"red-chair", "a step the record offered you left untaken", "judge", "red-chair.md", "prompt-red-chair.golden"},
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

	holdsPosition := func(seatID string) bool {
		for _, c := range cli.NewRootFor(seatID).Commands() {
			if c.Name() == "position" {
				return true
			}
		}
		return false
	}
	for _, k := range positionKinds {
		if !holdsPosition(k.seat) {
			t.Errorf("%s is asked for %q and its surface holds no position verb", k.seat, k.phrase)
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

	// THE POSITION IS THE CHAIR'S ALONE, and no seat that reads no transcript is named as a reader
	// by it: the limit its sentence states ("It instructs no lens and not blue") is true of the tree.
	for _, seatID := range []string{"red-lens-logic", "blue-respond", "blue-lane-1", "frontier", "blue-synthesize", "judge"} {
		if holdsPosition(seatID) {
			t.Errorf("%s holds the position verb: the position is the chair's argument, and a party argues a gap in its closing", seatID)
		}
	}
	for _, seatID := range []string{"red-lens-logic", "blue-respond", "blue-lane-1", "frontier", "blue-synthesize"} {
		if holdsTranscript(seatID) {
			t.Errorf("%s holds the transcript: the position sentences say no lens and no blue seat reads a position, and the tree disagrees", seatID)
		}
	}
}
