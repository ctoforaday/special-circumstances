package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EarlyLensDispatch is one dispatch row that engaged a lens on a gap whose turn had not come: blue
// had recorded no act on the gap and changed no sentence of it since the lens's last turn there, its
// mint included, and since the bench's latest remand of it (lensTurnOn), read on the record as it
// stood at the row.
type EarlyLensDispatch struct {
	At     int64  // the dispatch row's events.id
	SeatID string // the lens
	GapID  string
}

// EarlyLensDispatches is every dispatch row of the run that engaged a lens before its turn on a gap
// (ruled by gblock 2026-10-30) — what capture's lens-turn audit names. The plan engages a lens only
// once its turn has come, so a row here is a plan defect, or a run dispatched by a binary that did
// not hold the rule: universe-m36 holds seven, each a lens that sat first and audited an answer that
// did not exist yet.
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
	acts, mintedAt := gapActsOf(evs, ids, win), mintPlaces(evs, ids)
	minted := assigneeOf(evs)
	turns := lensTurnActsOf(evs, ids, acts, gapSittingsOf(evs, ids, win, AfterTheRun, minted), minted)
	var out []EarlyLensDispatch
	for i, e := range evs {
		d, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
		if !ok || seatclass.RoleOfSeat(d.GetSeatId()) != "lens" {
			continue
		}
		at := ids[i]
		for _, g := range d.GetGapIds() {
			if !lensTurnOn(turns[g], d.GetSeatId(), mintedAt[g], latestBefore(remandPlaces(remands[g]), at), at) {
				out = append(out, EarlyLensDispatch{At: at, SeatID: d.GetSeatId(), GapID: g})
			}
		}
	}
	return out
}

// EarlyDocket is one impasse docketing the dispatch wrote on a gap whose assignee still had its
// audit to come (lensGoesFirst), read on the record as it stood at the docketing.
type EarlyDocket struct {
	At        int64  // the docketing's events.id
	Docketing string // the docketing's id
	GapID     string
}

// EarlyDockets is every impasse docketing on the record written before its lens's re-audit (ruled by
// gblock 2026-10-31) — what capture's docket-order audit names. The plan never writes one, so a row
// here is a plan defect, or a run dispatched by a binary that did not hold the rule: universe-m37 holds
// six, five of them closed by the lens minutes after they were written. Only cause IMPASSE is counted:
// routeLensFirst orders the lens's audit before the impasse docketing alone, grade movement is docketed
// whatever the lens's turn, and a PROPOSAL docketing is a party's.
func EarlyDockets(run Run) ([]EarlyDocket, error) {
	p, err := RunParams(run)
	if err != nil {
		return nil, err
	}
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
	ids := win.IDs(evs)
	var out []EarlyDocket
	for i, e := range evs {
		d, ok := recordpb.BodyAs[*recordpb.Docketing](e)
		if !ok || d.GetCause() != recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE || d.GetGapId() == "" {
			continue
		}
		g := d.GetGapId()
		before := map[string][]remandRow{}
		for _, r := range remands[g] {
			if r.eventID < ids[i] {
				before[g] = append(before[g], r)
			}
		}
		if x := exchangesOf(evs[:i], ids[:i], win, p, WhileRunning, before)[g]; x != nil && (x.lensFirst || x.lensAudit) {
			out = append(out, EarlyDocket{At: ids[i], Docketing: d.GetDocketingId(), GapID: g})
		}
	}
	return out, nil
}
