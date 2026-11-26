package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// BLUE SITS AGAIN ON THE SUCCESSOR OF A GAP IT FOUND CLOSED (gblock's ruling R2, 2026-10-23).
//
// The lenses sit before blue in an epoch, and a lens can close a gap blue was dispatched onto and
// mint the gap that carries its remainder forward in the same sitting. Blue's dispatch names the
// gap it closed, so blue sat on nothing and the successor waited a whole epoch for an answer: on
// universe-m31 the repair of G-1ac85c33 landed in the final epoch and no lens sat after it.
//
// So blue's sitting that finds an engaged gap closed reads the successor the record names, and
// records `engage` — a dispatch row naming blue-respond on those successors, written by blue — and
// the engine sits it a second time that epoch. THE SUCCESSOR IS THE RECORD'S, NEVER BLUE'S: the
// write refuses a gap the record names as no successor, and one already engaged this epoch, which
// is the bound — at most one sitting per successor per epoch.

// successorsOf is each gap's successors as the record names them, in record order: the gap its
// close names in `superseded_by`.
func successorsOf(evs []*Event) map[string][]string {
	out := map[string][]string{}
	add := func(of, succ string) {
		if succ != "" && !slices.Contains(out[of], succ) {
			out[of] = append(out[of], succ)
		}
	}
	for _, e := range evs {
		if b, ok := mustBody(e).(*recordpb.Close); ok {
			add(b.GetGapId(), b.GetSuccessor())
		}
	}
	return out
}

// OwedSuccessor is one successor blue's sitting owes an `engage` on: the gap, and the engaged gap
// it found closed whose successor the record names it.
type OwedSuccessor struct {
	GapID string
	Of    string
}

// successorsOwed is what blue-respond's current sitting owes an `engage` on (owedSuccessorsOf). ok
// is false where blue-respond is not sitting a sitting for a dispatch — before it registers, or
// between sittings — so the act has no sitting to follow from.
func successorsOwed(evs []*Event, win WindowIndex, open func(string) bool) (owed []OwedSuccessor, ok bool) {
	s, sitting := blueSittingNow(evs, win)
	if !sitting {
		return nil, false
	}
	return owedSuccessorsOf(evs, win, s, open), true
}

// successorsOwedOnTheWorkList is what blue-respond's work list states it owes an `engage` on: the
// sitting it owes for its latest dispatch (owedSitting), split by the closes landed since the row
// (engagementOf), or else the sitting it is in now.
//
// THE LIST IS READ BEFORE THE REGISTER, and blue's register opens its sitting: the hook's bracket
// cannot name blue-respond, because blue-researcher's agent type is every blue seat's. Read only off
// the sitting blue is in, the owed set was empty on the list blue reads first and appeared at its
// first act (#1601 smoke, ev362).
func successorsOwedOnTheWorkList(evs []*Event, win WindowIndex, open func(string) bool) []OwedSuccessor {
	if d, owed := owedSitting(evs, win, blueRespondSeat); owed {
		return owedSuccessorsOf(evs, win, engagementOf(evs, d, int64(len(evs))), open)
	}
	owed, _ := successorsOwed(evs, win, open)
	return owed
}

// owedSuccessorsOf is what blue-respond's sitting s owes an `engage` on: for each engaged gap it
// found closed, in engaged order, each successor the record names that is still open and that no
// dispatch row naming blue-respond in this epoch engages.
//
// THE BOUND IS THE EPOCH'S ROWS. The chair's row engages blue on its gaps, and blue's own `engage`
// row on the successors; a successor either names is engaged this epoch and owed no further sitting.
// So once blue records the act its sitting owes nothing more, and a successor the chair's plan
// already engaged blue on is never sat twice.
func owedSuccessorsOf(evs []*Event, win WindowIndex, s BlueSitting, open func(string) bool) (owed []OwedSuccessor) {
	succ := successorsOf(evs)
	engaged := blueEngagedThisEpoch(evs, win)
	seen := map[string]bool{}
	for _, c := range s.FoundClosed {
		g := c.GapID
		for _, sg := range succ[g] {
			if seen[sg] || engaged[sg] || !open(sg) {
				continue
			}
			seen[sg] = true
			owed = append(owed, OwedSuccessor{GapID: sg, Of: g})
		}
	}
	return owed
}

// blueSittingNow is blue-respond's latest sitting for a dispatch, when it is the sitting blue is in
// now: its opening is blue's latest, in the epoch the record has reached. A sitting of an earlier
// epoch is not the sitting an `engage` follows from — the chair has sat since, and its plan, not
// blue's, says who sits on what.
func blueSittingNow(evs []*Event, win WindowIndex) (BlueSitting, bool) {
	ss := BlueSittings(evs, win, WhileRunning)
	if len(ss) == 0 {
		return BlueSitting{}, false
	}
	s := ss[len(ss)-1]
	at := win.Of(evs[s.opened])
	if at.SittingID != win.LatestSittingOf(blueRespondSeat) || at.Epoch != win.LastEpoch(evs) {
		return BlueSitting{}, false
	}
	return s, true
}

// blueEngagedThisEpoch is every gap a dispatch row naming blue-respond engages it on in the epoch
// the record has reached.
func blueEngagedThisEpoch(evs []*Event, win WindowIndex) map[string]bool {
	epoch := win.LastEpoch(evs)
	out := map[string]bool{}
	for _, e := range evs {
		if b, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok && b.GetSeatId() == blueRespondSeat && win.Of(e).Epoch == epoch {
			for _, g := range b.GetGapIds() {
				out[g] = true
			}
		}
	}
	return out
}

// successorsOwedAt is successorsOwed read off the record through q, with the stream it read.
func successorsOwedAt(q recordsql.Querier) ([]OwedSuccessor, bool, []*Event, WindowIndex, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return nil, false, nil, win, err
	}
	var openErr error
	open := func(id string) bool {
		closed, err := gapState(q, id)
		if err != nil && openErr == nil {
			openErr = err
		}
		return !closed
	}
	owed, sitting := successorsOwed(evs, win, open)
	return owed, sitting, evs, win, openErr
}

// EngagementOwesNone reports whether b, written by writer, is blue's engagement in a sitting that
// owes no successor: requireEngage's "owes none" refusal, and that refusal alone. Migrate writes no
// event for an archived engagement it reports, whose successors a dropped lineage link left unnamed.
func EngagementOwesNone(q recordsql.Querier, writer string, b *recordpb.Dispatch) (bool, error) {
	if writer != blueRespondSeat || b.GetSeatId() != blueRespondSeat {
		return false, nil
	}
	owed, sitting, _, _, err := successorsOwedAt(q)
	return sitting && len(owed) == 0, err
}

// isEngageRow reports whether e is a dispatch row blue-respond wrote — an `engage`, which engages
// blue's next sitting in the epoch the chair's dispatch opened rather than opening a dispatch of its
// own.
func isEngageRow(e *Event) bool {
	_, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
	return ok && e.GetSeatId() == blueRespondSeat
}

// requireEngage is the write path's check of blue's `engage`: its row names blue-respond, and its
// gaps are exactly the successors blue's current sitting owes (successorsOwed) — none the record
// does not name, none already engaged this epoch, none left out. An act naming part of the set is
// refused rather than completed: the seat states the set, the record holds it to the set.
func requireEngage(q recordsql.Querier, writer string, b *recordpb.Dispatch) error {
	if writer != blueRespondSeat {
		return feov.Errorf(feov.Validation, "only %s engages a next sitting — it is the blue seat dispatched onto gaps; %s is dispatched by no row", blueRespondSeat, writer)
	}
	if b.GetSeatId() != blueRespondSeat {
		return feov.Errorf(feov.Validation, "blue engages its own next sitting and no other seat's: this row names %q", b.GetSeatId())
	}
	owed, sitting, evs, win, err := successorsOwedAt(q)
	if err != nil {
		return err
	}
	if !sitting {
		return feov.Errorf(feov.Validation, "you are not sitting a sitting for a dispatch in this epoch — an engagement follows a gap the sitting you are in found closed to the successor the record names")
	}
	if len(owed) == 0 {
		return feov.Errorf(feov.Validation, "this sitting owes none: no gap it found closed has an open successor on the record that is not already engaged this epoch — your work list's engaged.successors_owed is empty")
	}
	want := make([]string, len(owed))
	for i, o := range owed {
		want[i] = o.GapID
	}
	gave := b.GetGapIds()
	succ := successorsOf(evs)
	engaged := blueEngagedThisEpoch(evs, win)
	var problems []string
	seen := map[string]bool{}
	for _, g := range gave {
		switch {
		case seen[g]:
			problems = append(problems, fmt.Sprintf("%s is named twice", g))
		case slices.Contains(want, g):
		case engaged[g]:
			problems = append(problems, fmt.Sprintf("%s is already engaged this epoch — blue sits once on it, and this epoch's dispatch rows already name it", g))
		case !isSuccessorOfAny(succ, g):
			problems = append(problems, fmt.Sprintf("%s is no successor the record names: no gap your sitting found closed names it in its close's superseded_by", g))
		default:
			problems = append(problems, fmt.Sprintf("%s is not owed: it is closed, or it succeeds no gap this sitting found closed", g))
		}
		seen[g] = true
	}
	if len(problems) > 0 {
		return feov.Errorf(feov.Validation, "%s. This sitting owes exactly %s — your work list's engaged.successors_owed",
			strings.Join(problems, "; "), strings.Join(want, ","))
	}
	return nil
}

// requireEngageNamesEveryOwed refuses an engagement that leaves out a successor its sitting owes:
// one `engage` names the whole set (requireEngage holds each name to it). An act judgement (judgeAct):
// an archived blue engaged the successors its sitting owed under the rule of its day.
func requireEngageNamesEveryOwed(q recordsql.Querier, b *recordpb.Dispatch) error {
	owed, _, _, _, err := successorsOwedAt(q)
	if err != nil {
		return err
	}
	var missing, want []string
	for _, o := range owed {
		want = append(want, o.GapID)
		if !slices.Contains(b.GetGapIds(), o.GapID) {
			missing = append(missing, fmt.Sprintf("%s (successor of %s)", o.GapID, o.Of))
		}
	}
	if len(missing) > 0 {
		return feov.Errorf(feov.Validation, "it leaves out %s — one `engage` names every successor the sitting owes. This sitting owes exactly %s — your work list's engaged.successors_owed",
			strings.Join(missing, ", "), strings.Join(want, ","))
	}
	return nil
}

// isSuccessorOfAny reports whether the record names g as any gap's successor.
func isSuccessorOfAny(succ map[string][]string, g string) bool {
	for _, ss := range succ {
		if slices.Contains(ss, g) {
			return true
		}
	}
	return false
}

// ReportHead is the report head on the landing record — the pin a dispatch row records.
func (l *Landing) ReportHead() (int64, error) { return ReportHeadAt(l.w.q) }
