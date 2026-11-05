package fuzz

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// A MINT SUPERSEDING AN OPEN GAP CLOSES IT (ruled by gblock 2026-11-04). The drive supersedes only
// what the record admits — a gap the minting lens holds open and may close — and holds the record to
// the closing: the ancestor leaves the open board, closed, naming the successor.

// holdClosing marks gap id as one whose close the drive is driving (on), or clears it: a mint drawn
// in between does not supersede it, since the supersession would close it before that close.
func (r *runner) holdClosing(id string, on bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closingNow == nil {
		r.closingNow = map[string]bool{}
	}
	if on {
		r.closingNow[id] = true
	} else {
		delete(r.closingNow, id)
	}
}

// supersedable is each open gap seatID may supersede, in board order: one its work list says is its
// to close — it minted it, and no docket motion stands on it — and no close of which is being driven.
func (r *runner) supersedable(seatID string) []string {
	w, err := record.WorkOfSeat(r.run(), "lens", seatID)
	if err != nil {
		r.noteActErr(seatID + "'s work list could not be read to draw a gap to supersede: " + err.Error())
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, g := range w.Open {
		if g.YoursToClose && !r.closingNow[g.ID] {
			out = append(out, g.ID)
		}
	}
	return out
}

// supersessionClosed holds the record to a landed mint of successor that superseded ancestor, open
// when it was drawn: the board holds the ancestor closed, and names successor as the gap whose mint
// closed it.
func (r *runner) supersessionClosed(ancestor, successor string) {
	if ancestor == "" {
		return
	}
	bj, err := record.BoardJSONOfRun(r.run())
	if err != nil {
		r.noteActErr("the board could not be read after a mint superseding " + ancestor + ": " + err.Error())
		return
	}
	for _, g := range bj.Closed {
		if g.ID == ancestor {
			if g.SupersededBy != successor {
				r.noteActErr("gap " + ancestor + " closed, and the board names " + g.SupersededBy + " as the successor whose mint closed it, not " + successor + ", which superseded it open")
			}
			return
		}
	}
	r.noteActErr("gap " + ancestor + " is still open after " + successor + " was minted superseding it: a mint superseding an open gap closes it")
}
