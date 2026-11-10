package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// CLOSURES ARE HISTORY (gblock, 2026-11-01). A retire moves no closure: a closed gap whose answers
// are all retired stays closed for every reader, and what the retire adds is one hop to its minting
// lens's work list — the gap view's `answers_retired_event`, read through the WORK query. The item
// blocks nothing, so `complete` and PASS are as they were: no BlockerKind, no lensOwesOf row and no
// gapNewsOf row reads the column (gapNewsOf holds open gaps).
//
// WHAT REMOVES IT, each off the record and none a new event or a seen-marker:
//
//   - the lens has sat through a sitting opened after the retire (satThrough, as lensOwesOf's `owe`
//     reads it): it is listed for the whole of that sitting, so the sitting that clears it is one that
//     read it, and a sitting stopped at the call limit, or never closed, does not clear it;
//   - the lens has minted a gap naming it in `supersedes` after the retire;
//   - a closing act other than a carry follows the retire: the column is then NULL.

// answersRetiredItems is seatID's items for the closed gaps of gaps it minted whose answers are all
// retired. ids are evs's places.
func answersRetiredItems(evs []*Event, ids []int64, win WindowIndex, gaps []WorkGapState, seatID string) []Item {
	var held []WorkGapState
	for _, g := range gaps {
		if !g.Open && g.AnswersRetiredEvent != 0 && g.MintedBy == seatID {
			held = append(held, g)
		}
	}
	if len(held) == 0 {
		return nil
	}
	_, openings := dispatchLedger(evs, ids, win)
	ends := satThroughOf(evs, ids, win, openings)
	var out []Item
	for _, g := range held {
		n := g.AnswersRetiredEvent
		if _, sat := firstAfter(ends.through[seatID], n); sat {
			continue
		}
		if supersededBySeatAfter(evs, ids, seatID, g.ID, n) {
			continue
		}
		what := fmt.Sprintf("the answers to gap %s, which you closed, have all been retired, the last at event %d (%s) — the closure stands", g.ID, n, strings.Join(retiredAnswersAt(evs, ids, g.ID, n), ", "))
		if g.benchClosedLast {
			what = fmt.Sprintf("the answers to gap %s, which the bench closed, have all been retired, the last at event %d (%s) — the closure stands, and what its ruling settled stays barred to you", g.ID, n, strings.Join(retiredAnswersAt(evs, ids, g.ID, n), ", "))
		}
		what += "; if the defect it closed is still in the report, that is a new gap: mint it, naming " + g.ID + " as the gap it supersedes"
		out = append(out, gapItem(ItemNews, g.ID, what))
	}
	return out
}

// supersededBySeatAfter is whether seatID minted a gap naming gap in `supersedes` after the place at.
func supersededBySeatAfter(evs []*Event, ids []int64, seatID, gap string, at int64) bool {
	for i, e := range evs {
		if m, ok := recordpb.BodyAs[*recordpb.Mint](e); ok && ids[i] > at && e.GetSeatId() == seatID && slices.Contains(m.GetSupersedes(), gap) {
			return true
		}
	}
	return false
}

// retiredAnswersAt is the anchors placed as gap's answer that the retire at the place at took out.
func retiredAnswersAt(evs []*Event, ids []int64, gap string, at int64) []string {
	answering := map[string]bool{}
	var out []string
	for i, e := range evs {
		if id, g, ok := placedAnswer(e); ok && g == gap {
			answering[id] = true
		}
		if r, ok := recordpb.BodyAs[*recordpb.Retire](e); ok && ids[i] == at {
			for _, a := range r.GetAnchors() {
				if answering[a] {
					out = append(out, a)
				}
			}
		}
	}
	return out
}
