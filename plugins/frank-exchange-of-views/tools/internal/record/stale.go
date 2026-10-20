package record

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// StaleChange is one act on a gap that a seat's sitting opened before: its events.id, the seat
// that recorded it, and what it did, as one line.
type StaleChange struct {
	ID     int64
	SeatID string
	What   string
}

// Line is the change as a refusal shows it.
func (c StaleChange) Line() string { return fmt.Sprintf("event %d, %s: %s", c.ID, c.SeatID, c.What) }

// StaleError is the refusal of an act on a gap that changed after the acting seat's read point: the
// gap and what changed, in record order. Fresh returns it; the verb records the refusal (a Stale event
// through the last change) and hands the seat the changes.
type StaleError struct {
	Gap     string
	Changes []StaleChange
}

func (e *StaleError) Error() string {
	lines := make([]string, len(e.Changes))
	for i, c := range e.Changes {
		lines[i] = "  " + c.Line()
	}
	return fmt.Sprintf("%s changed after your sitting opened, and this act reads it as it stood then:\n%s\n\n"+
		"Read %s as it now stands and decide whether the act still holds. If it does, record it again: "+
		"this refusal is where your next act on %s reads from, so the same act is not refused for these changes twice",
		e.Gap, strings.Join(lines, "\n"), e.Gap, e.Gap)
}

// Through is the last change the refusal shows: the read point it moves the seat to.
func (e *StaleError) Through() int64 { return e.Changes[len(e.Changes)-1].ID }

// Fresh is the guard of an act seatID records about gapID (gblock ruling B, 2026-10-19): it refuses with a *StaleError while
// another seat's act on the gap is newer than the seat's read point. It reads the record the act
// lands on, under the write's lock, so a close that commits between a check and the write cannot
// pass between them. It is asked last (CheckLast): an act the write path refuses for itself is
// refused for that, and is not sent back to be recorded again.
//
// THE READ POINT IS THE SITTING'S OPENING EVENT. A seat takes its work list when its sitting opens,
// and lenses sit concurrently, so a gap can be closed, answered or regraded while a seat that read it
// open is still writing about it — on universe-m30 two lenses filed defects against a board that had
// moved under them. A stale refusal advances the point to the last act it showed (Stale.through), so
// the same act filed again once the seat has read them is not refused again.
func Fresh(seatID, gapID string) Guard {
	return CheckLast(func(l *Landing) error {
		evs, win, err := eventsAt(l.w.q)
		if err != nil {
			return err
		}
		if ch := staleOf(evs, win, seatID, gapID); len(ch) > 0 {
			return &StaleError{Gap: gapID, Changes: ch}
		}
		return nil
	})
}

// staleOf is StaleOn over a loaded stream.
//
// WHAT COUNTS AS A CHANGE is the work list's own news on a gap (gapNewsOf) — blue's answer, an
// accepted grade motion, the bench's remand, another seat's finding about it — and the acts that
// move the gap's state, which that fold leaves out because they are its minting lens's own or end
// the gap: a close, a regrade, and the bench's ruling on a docket motion on it. A seat's own acts are
// never news to it.
func staleOf(evs []*Event, win WindowIndex, seatID, gapID string) []StaleChange {
	sitting := win.LatestSittingOf(seatID)
	from := sitting
	if sitting == 0 || gapID == "" {
		return nil
	}
	ids := win.IDs(evs)
	seatAt := make(map[int64]string, len(evs))
	minter := ""
	for i, e := range evs {
		seatAt[ids[i]] = e.GetSeatId()
		if m := e.GetMint(); m != nil && m.GetGapId() == gapID {
			minter = e.GetSeatId()
		}
		if s := e.GetStale(); s != nil && s.GetGapId() == gapID && e.GetSeatId() == seatID && win.Of(e).SittingID == sitting {
			from = max(from, s.GetThrough())
		}
	}
	if minter == "" {
		return nil
	}
	seen := map[int64]bool{}
	var out []StaleChange
	add := func(at int64, what string) {
		if at > from && !seen[at] && seatAt[at] != seatID {
			seen[at] = true
			out = append(out, StaleChange{ID: at, SeatID: seatAt[at], What: what})
		}
	}
	for _, n := range gapNewsOf(evs, ids, win, []blockerGap{{id: gapID, mintedBy: minter}}) {
		add(n.at, n.item)
	}
	for i, e := range evs {
		switch {
		case e.GetClose() != nil && e.GetClose().GetGapId() == gapID:
			add(ids[i], "closed "+gapID+" as "+recordpb.Word(e.GetClose().GetClosureClass()))
		case e.GetRegrade() != nil && e.GetRegrade().GetGapId() == gapID:
			add(ids[i], "regraded "+gapID)
		}
	}
	for _, m := range motionsAt(evs, ids, win) {
		if m.Subject == "docket" && m.Ruled() && m.GapID == gapID {
			add(m.ruled, "the bench ruled "+gapID+" "+m.Ruling+" in "+m.ID)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
