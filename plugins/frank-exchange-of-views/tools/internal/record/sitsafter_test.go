package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// TestPartiesSitAfterKeysOnTheAuthor holds the three predicates on their reasons (#1695 §III.B4), for
// each epoch of a run limited to 6 and for a run held to none: a lens sits after a docket sitting up to
// the held-out epoch's plan, blue up to the plan before the one that holds it out, and a remand needs
// both — so in the epoch before the held-out one it is open with a lens the author and closed with
// blue. No plan follows the terminal or the assembly sitting.
func TestPartiesSitAfterKeysOnTheAuthor(t *testing.T) {
	docket := recordpb.Occasion_OCCASION_DOCKET
	limited := Params{MaxEpochs: 6}
	for _, c := range []struct {
		epoch                        int
		lens, blue, lensAut, blueAut bool
	}{
		{3, true, true, true, true},
		{4, true, false, true, false},
		{5, false, false, false, false},
		{6, false, false, false, false},
	} {
		if got := lensSitsAfter(docket, limited, c.epoch); got != c.lens {
			t.Errorf("lensSitsAfter at epoch %d of 6 = %v, want %v", c.epoch, got, c.lens)
		}
		if got := blueSitsAfter(limited, c.epoch); got != c.blue {
			t.Errorf("blueSitsAfter at epoch %d of 6 = %v, want %v", c.epoch, got, c.blue)
		}
		if got := partiesSitAfter(docket, limited, c.epoch, "lens"); got != c.lensAut {
			t.Errorf("partiesSitAfter with a lens the author at epoch %d of 6 = %v, want %v", c.epoch, got, c.lensAut)
		}
		if got := partiesSitAfter(docket, limited, c.epoch, "blue"); got != c.blueAut {
			t.Errorf("partiesSitAfter with blue the author at epoch %d of 6 = %v, want %v", c.epoch, got, c.blueAut)
		}
	}
	for _, o := range []recordpb.Occasion{recordpb.Occasion_OCCASION_TERMINAL, recordpb.Occasion_OCCASION_ASSEMBLE} {
		if lensSitsAfter(o, Params{}, 3) || partiesSitAfter(o, Params{}, 3, "lens") {
			t.Errorf("a lens or the parties sit after the %s sitting", recordpb.Word(o))
		}
	}
	if !lensSitsAfter(docket, Params{}, 40) || !blueSitsAfter(Params{}, 40) || !partiesSitAfter(docket, Params{}, 40, "blue") {
		t.Error("a run held to no epoch limit has a sitting after every docket sitting")
	}
}
