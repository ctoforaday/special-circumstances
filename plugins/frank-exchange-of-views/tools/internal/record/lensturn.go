package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EarlyLensDispatch is one dispatch row that engaged a lens on a gap whose turn had not come: blue
// had recorded no act on the gap since the lens's last act there, its mint included, and since the
// bench's latest remand of it (lensTurnOn), read on the record as it stood at the row.
type EarlyLensDispatch struct {
	At     int64  // the dispatch row's events.id
	SeatID string // the lens
	GapID  string
}

// EarlyLensDispatches is every dispatch row of the run that engaged a lens before its turn on a gap
// (ruled by gblock 2026-10-30) — what capture's lens-turn audit names. The plan engages a lens only
// once its turn has come, so a row here is a plan defect, or a run dispatched by a binary that did
// not hold the rule: universe-m36 holds seven, each a lens that sat first and audited an answer that
// did not exist yet. A gap a successor names is not counted: its lens owes the close the supersession
// promised, whatever blue did.
func EarlyLensDispatches(run Run) ([]EarlyLensDispatch, error) {
	var evs []*Event
	var win WindowIndex
	var remands map[string][]remandRow
	absent := false
	if err := readSnapshot(run, func(q recordsql.Querier) error {
		if absent = noRecord(q); absent {
			return nil
		}
		var err error
		if evs, win, err = eventsAt(q); err != nil {
			return err
		}
		remands, err = remandRulingsOf(q)
		return err
	}); err != nil || absent {
		return nil, err
	}
	return earlyLensDispatchesOf(evs, win, remands), nil
}

// earlyLensDispatchesOf is EarlyLensDispatches over a read stream and the "remand" view's rows.
func earlyLensDispatchesOf(evs []*Event, win WindowIndex, remands map[string][]remandRow) []EarlyLensDispatch {
	ids := win.IDs(evs)
	acts, mintedAt := gapActsOf(evs, ids), mintPlaces(evs, ids)
	superseded := map[string]int64{} // gap → the place of the first mint that names it superseded
	for i, e := range evs {
		if m := e.GetMint(); m != nil {
			for _, s := range m.GetSupersedes() {
				if _, seen := superseded[s]; !seen {
					superseded[s] = ids[i]
				}
			}
		}
	}
	var out []EarlyLensDispatch
	for i, e := range evs {
		d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
		if !ok || seatclass.RoleOfSeat(d.GetSeatId()) != "lens" {
			continue
		}
		at := ids[i]
		for _, g := range d.GetGapIds() {
			if s, named := superseded[g]; named && s < at {
				continue
			}
			if !lensTurnOn(acts[g], d.GetSeatId(), mintedAt[g], latestBefore(remandPlaces(remands[g]), at), at) {
				out = append(out, EarlyLensDispatch{At: at, SeatID: d.GetSeatId(), GapID: g})
			}
		}
	}
	return out
}
