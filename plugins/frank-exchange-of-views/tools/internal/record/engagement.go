package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// BLUE SITS AGAIN IN THE LAST EPOCH IT ANSWERS IN, ON THE GAPS MINTED AFTER ITS DISPATCH
// (plans/findings-fold.md X22a, §III.B4 blueSitsAfter; R2 2026-10-23, generalized and narrowed).
//
// The lenses sit before blue in an epoch, so a gap of blue's minted after the chair's dispatch named
// blue's gaps is on no dispatch row blue sits for that epoch. Where a later sitting of blue's follows
// (blueSitsAfter), that sitting answers it, with a lens sitting after to audit the answer. Where none
// follows — the last epoch blue answers in, the one before the epoch that holds blue out — the gap
// would wait for a sitting that never comes: on universe-m31 a repair landed in the final epoch and no
// lens sat after it. So there, and only there, blue's sitting owes `engage` — a dispatch row naming
// blue-respond on those gaps, written by blue — and the engine sits it once more that epoch.
//
// THE SET IS THE RECORD'S, NEVER BLUE'S: every open gap blue authors (authorOf) minted after the
// dispatch row engaging the sitting and engaged by no dispatch row of this epoch. The bound is one
// more sitting an epoch: once blue's own `engage` row stands this epoch, nothing more is owed.

// OwedEngagement is one gap blue's sitting owes an `engage` on.
type OwedEngagement struct {
	GapID string
}

// owedEngagements is what a blue sitting owes an `engage` on, and where it owes none, why.
type owedEngagements struct {
	owed []OwedEngagement
	none string // the reason nothing is owed, in words; "" where something is
}

// engagementsOwed is what blue-respond's current sitting owes an `engage` on (owedEngagementsOf). ok
// is false where blue-respond is not sitting a sitting for a dispatch — before it registers, or
// between sittings — so the act has no sitting to follow from.
func engagementsOwed(evs []*Event, win WindowIndex, open func(string) bool) (owed owedEngagements, ok bool) {
	s, sitting := blueSittingNow(evs, win)
	if !sitting {
		return owedEngagements{}, false
	}
	return owedEngagementsOf(evs, win, s, open), true
}

// engagementsOwedOnTheWorkList is what blue-respond's work list states it owes an `engage` on: the
// sitting it owes for its latest dispatch (owedSitting), read off that dispatch (engagementOf), or
// else the sitting it is in now.
//
// THE LIST IS READ BEFORE THE REGISTER, and blue's register opens its sitting: the hook's bracket
// cannot name blue-respond, because blue-researcher's agent type is every blue seat's. Read only off
// the sitting blue is in, the owed set was empty on the list blue reads first and appeared at its
// first act (#1601 smoke, ev362).
func engagementsOwedOnTheWorkList(evs []*Event, win WindowIndex, open func(string) bool) []OwedEngagement {
	if d, owed := owedSitting(evs, win, blueRespondSeat); owed {
		return owedEngagementsOf(evs, win, engagementOf(evs, d, int64(len(evs))), open).owed
	}
	owed, _ := engagementsOwed(evs, win, open)
	return owed.owed
}

// owedEngagementsOf is what blue-respond's sitting s owes an `engage` on: where blue sits again in no
// later epoch but answers in this one (blueSittingLast), and blue has written no `engage` this
// epoch, each open gap blue authors minted after s's dispatch row that no dispatch row naming
// blue-respond in this epoch engages, in mint order.
//
// THE BOUND IS THE EPOCH'S ROWS. The chair's row engages blue on its gaps, and blue's own `engage`
// row on the rest; a gap either names is engaged this epoch and owed no further sitting, and blue's
// own row is its one more sitting.
func owedEngagementsOf(evs []*Event, win WindowIndex, s BlueSitting, open func(string) bool) owedEngagements {
	if why, last := blueSittingLast(evs); !last {
		return owedEngagements{none: why}
	}
	epoch := win.LastEpoch(evs)
	for _, e := range evs {
		if isEngageRow(e) && win.Of(e).Epoch == epoch {
			return owedEngagements{none: "your one more sitting this epoch is already engaged — blue sits at most twice in an epoch"}
		}
	}
	engaged := blueEngagedThisEpoch(evs, win)
	minter := map[string]string{}
	var out []OwedEngagement
	for i, e := range evs {
		m, ok := recordpb.BodyAs[*recordpb.Mint](e)
		if !ok {
			continue
		}
		g := m.GetGapId()
		minter[g] = e.GetSeatId()
		if int64(i) <= s.dispatched || engaged[g] || !open(g) {
			continue
		}
		if authorOf(m.GetAboutKind(), m.GetAboutRef(), func(gap string) string { return minter[gap] }) == blueRespondSeat {
			out = append(out, OwedEngagement{GapID: g})
		}
	}
	if len(out) == 0 {
		return owedEngagements{none: "no open gap of yours was minted after your dispatch that this epoch's dispatch rows do not already engage you on"}
	}
	return owedEngagements{owed: out}
}

// blueSittingLast reports whether the latest stored plan's epoch is the last blue answers in: blue is
// dispatched in it not for its closings only, and in no epoch after it but for its closings
// (blueSitsAfter fails). Where it is not, why is the reason nothing is owed. A record with no stored
// plan states no limit, and a run held to none has a later sitting after every epoch.
func blueSittingLast(evs []*Event) (why string, last bool) {
	params, epoch, ok := latestPlanTermsOf(evs)
	const later = "a later sitting of yours answers every gap minted after your dispatch, with a lens sitting after it, so the gap waits on your work list for that sitting"
	switch {
	case !ok, blueSitsAfter(params, epoch):
		return later, false
	case closingsOnlyFor(params, epoch, "blue"):
		return "this epoch holds your sittings to their closings, and no sitting of yours answers a gap again before the run ends", false
	}
	return "", true
}

// latestPlanTermsOf is the latest stored plan's epoch and limit on evs, as rulingSittingAt reads them
// off the record; ok is false where no plan is stored or the stored plan does not read.
func latestPlanTermsOf(evs []*Event) (Params, int, bool) {
	for i := len(evs) - 1; i >= 0; i-- {
		b, ok := recordpb.BodyAs[*recordpb.Plan](evs[i])
		if !ok {
			continue
		}
		p, _, err := ReadPlan([]byte(b.GetPrinted()))
		if err != nil {
			return Params{}, 0, false
		}
		return Params{MaxEpochs: p.MaxEpochs}, p.Epoch, true
	}
	return Params{}, 0, false
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

// engagementsOwedAt is engagementsOwed read off the record through q, with the stream it read.
func engagementsOwedAt(q recordsql.Querier) (owedEngagements, bool, []*Event, WindowIndex, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return owedEngagements{}, false, nil, win, err
	}
	var openErr error
	open := func(id string) bool {
		closed, err := gapState(q, id)
		if err != nil && openErr == nil {
			openErr = err
		}
		return !closed
	}
	owed, sitting := engagementsOwed(evs, win, open)
	return owed, sitting, evs, win, openErr
}

// EngagementOwesNone reports whether b, written by writer, is blue's engagement in a sitting that
// owes none: requireEngage's "owes none" refusal, and that refusal alone. Migrate writes no event for
// an archived engagement its sitting owes none of under the rule of this record (X22a).
func EngagementOwesNone(q recordsql.Querier, writer string, b *recordpb.Dispatch) (bool, error) {
	if writer != blueRespondSeat || b.GetSeatId() != blueRespondSeat {
		return false, nil
	}
	owed, sitting, _, _, err := engagementsOwedAt(q)
	return sitting && len(owed.owed) == 0, err
}

// isEngageRow reports whether e is a dispatch row blue-respond wrote — an `engage`, which engages
// blue's next sitting in the epoch the chair's dispatch opened rather than opening a dispatch of its
// own.
func isEngageRow(e *Event) bool {
	_, ok := recordpb.BodyAs[*recordpb.Dispatch](e)
	return ok && e.GetSeatId() == blueRespondSeat
}

// requireEngage is the write path's check of blue's `engage`: its row names blue-respond, and its
// gaps are exactly the set blue's current sitting owes (engagementsOwed) — none it does not owe, none
// left out. An act naming part of the set is refused rather than completed: the seat states the set,
// the record holds it to the set.
func requireEngage(q recordsql.Querier, writer string, b *recordpb.Dispatch) error {
	if writer != blueRespondSeat {
		return feov.Errorf(feov.Validation, "only %s engages a next sitting — it is the blue seat dispatched onto gaps; %s is dispatched by no row", blueRespondSeat, writer)
	}
	if b.GetSeatId() != blueRespondSeat {
		return feov.Errorf(feov.Validation, "blue engages its own next sitting and no other seat's: this row names %q", b.GetSeatId())
	}
	owed, sitting, _, _, err := engagementsOwedAt(q)
	if err != nil {
		return err
	}
	if !sitting {
		return feov.Errorf(feov.Validation, "you are not sitting a sitting for a dispatch in this epoch — an engagement follows from the sitting you are in")
	}
	if len(owed.owed) == 0 {
		return feov.Errorf(feov.Validation, "this sitting owes none: %s — your work list's engaged.owed is empty", owed.none)
	}
	want := make([]string, len(owed.owed))
	for i, o := range owed.owed {
		want[i] = o.GapID
	}
	var problems []string
	seen := map[string]bool{}
	for _, g := range b.GetGapIds() {
		switch {
		case seen[g]:
			problems = append(problems, fmt.Sprintf("%s is named twice", g))
		case slices.Contains(want, g):
		default:
			problems = append(problems, fmt.Sprintf("%s is not owed: it is closed, not yours, minted before your dispatch, or already engaged this epoch", g))
		}
		seen[g] = true
	}
	if len(problems) > 0 {
		return feov.Errorf(feov.Validation, "%s. This sitting owes exactly %s — your work list's engaged.owed",
			strings.Join(problems, "; "), strings.Join(want, ","))
	}
	return nil
}

// requireEngageNamesEveryOwed refuses an engagement that leaves out a gap its sitting owes: one
// `engage` names the whole set (requireEngage holds each name to it). An act judgement (judgeAct):
// an archived blue engaged what its sitting owed under the rule of its day.
func requireEngageNamesEveryOwed(q recordsql.Querier, b *recordpb.Dispatch) error {
	owed, _, _, _, err := engagementsOwedAt(q)
	if err != nil {
		return err
	}
	var missing, want []string
	for _, o := range owed.owed {
		want = append(want, o.GapID)
		if !slices.Contains(b.GetGapIds(), o.GapID) {
			missing = append(missing, o.GapID)
		}
	}
	if len(missing) > 0 {
		return feov.Errorf(feov.Validation, "it leaves out %s — one `engage` names every gap the sitting owes. This sitting owes exactly %s — your work list's engaged.owed",
			strings.Join(missing, ", "), strings.Join(want, ","))
	}
	return nil
}

// ReportHead is the report head on the landing record — the pin a dispatch row records.
func (l *Landing) ReportHead() (int64, error) { return ReportHeadAt(l.w.q) }
