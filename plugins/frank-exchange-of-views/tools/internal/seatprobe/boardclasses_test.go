package seatprobe

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

// BOARDCLASSES IS A SUBSET OF WHAT SHIPS, and nothing checked that.
//
// It is production code: build.go stages it as a probe run's class registry, so a slug here
// that the shipped registry does not carry is a vocabulary the probe mints under and no real
// run would accept. It was a hand-written 13-slug list beside a hand-written 38-slug mirror
// of the registry, with no comparison between either of them and the registry itself.
//
// The subset is deliberate — a probe board does not need the whole taxonomy — so this asserts
// containment rather than equality. What it refuses is the drift: rename a class in the
// registry and this fails, instead of the probe quietly minting under a name production has
// stopped accepting.
func TestBoardClassesAreAllShipped(t *testing.T) {
	shipped := map[string]bool{}
	for _, s := range recordtest.ShippedClasses {
		shipped[s] = true
	}
	if len(shipped) == 0 {
		t.Fatal("the shipped vocabulary is EMPTY — an empty set would make every check below " +
			"pass without comparing anything")
	}
	if len(BoardClasses) == 0 {
		t.Fatal("BoardClasses is empty, so the boards stage no vocabulary at all")
	}
	for _, c := range BoardClasses {
		if !shipped[c] {
			t.Errorf("BoardClasses carries %q, which the shipped registry does not — a probe "+
				"board would mint under a class no real run accepts", c)
		}
	}
}
