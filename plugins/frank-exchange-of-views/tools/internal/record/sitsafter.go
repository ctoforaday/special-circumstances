package record

import (
	"database/sql"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// WHO SITS AFTER A RULING (#1695 §III.B4). A raiser's statement the bench applies is a change of the
// report no party has made, so it stands only where a lens sits afterwards to audit it; a remand grants
// one more exchange, so it is open only where both of the gap's parties sit afterwards to have it. Each
// predicate is keyed on that reason, read from what the ruling's write sees: the bench's sitting's
// occasion, and the latest stored plan's epoch and limit — the plan that convened it.
//
//   - lensSitsAfter: the bench rules in a sitting the dispatch convened, and a lens is dispatched, not
//     for its closings only (closingsOnlyFor), in the epoch after the plan's. False in the terminal
//     and the assembly sittings, which no plan follows.
//   - blueSitsAfter: blue is dispatched, not for its closings only, in the epoch after.
//   - partiesSitAfter: lensSitsAfter, and the gap's author (authorOf) dispatched, not for its closings
//     only, in the epoch after.

// lensSitsAfter reports whether a lens sits after a ruling in a sitting of occasion, under the plan
// whose epoch and limit params and epoch state.
func lensSitsAfter(occasion recordpb.Occasion, params Params, epoch int) bool {
	switch occasion {
	case recordpb.Occasion_OCCASION_TERMINAL, recordpb.Occasion_OCCASION_ASSEMBLE:
		return false
	}
	return !closingsOnlyFor(params, epoch+1, "lens")
}

// blueSitsAfter reports whether blue is dispatched, not for its closings only, in an epoch after epoch.
func blueSitsAfter(params Params, epoch int) bool {
	return !closingsOnlyFor(params, epoch+1, "blue")
}

// partiesSitAfter reports whether both of a gap's parties sit after a ruling: a lens, and its author,
// whose role is authorRole.
func partiesSitAfter(occasion recordpb.Occasion, params Params, epoch int, authorRole string) bool {
	return lensSitsAfter(occasion, params, epoch) && !closingsOnlyFor(params, epoch+1, authorRole)
}

// rulingSitting is the bench's current sitting's occasion and the latest stored plan's terms, as a
// ruling's write sees them. A record with no stored plan states no limit.
type rulingSitting struct {
	occasion recordpb.Occasion
	params   Params
	epoch    int
}

// rulingSittingAt reads the bench's current occasion and the latest stored plan off q's record.
func rulingSittingAt(q recordsql.Querier, evs []*Event, win WindowIndex) (rulingSitting, error) {
	s := rulingSitting{occasion: benchOccasionNow(evs, win.IDs(evs), win)}
	var printed sql.NullString
	found, err := queryRowAt(q, []any{&printed}, `SELECT "printed" FROM "plan" ORDER BY "event_id" DESC LIMIT 1`)
	if err != nil || !found {
		return s, err
	}
	p, _, err := ReadPlan([]byte(printed.String))
	if err != nil {
		return s, fmt.Errorf("reading the latest stored plan: %w", err)
	}
	s.params, s.epoch = Params{MaxEpochs: p.MaxEpochs}, p.Epoch
	return s, nil
}

// benchOccasionNow is the occasion the bench's current sitting's register states; unspecified where
// the bench has not sat or its sitting states none.
func benchOccasionNow(evs []*Event, ids []int64, win WindowIndex) recordpb.Occasion {
	opening := win.LatestSittingOf(benchSeat)
	if opening == 0 {
		return recordpb.Occasion_OCCASION_UNSPECIFIED
	}
	for o, at := range benchOccasionSittings(evs, ids, win) {
		for _, p := range at {
			if p == opening {
				return o
			}
		}
	}
	return recordpb.Occasion_OCCASION_UNSPECIFIED
}

// requireSitsAfter refuses a bench ruling whose answer needs a sitting no plan gives it: an applied
// statement of the raiser's where no lens sits after (lensSitsAfter), a remand where the gap's parties
// do not both sit after (partiesSitAfter). applied is whether the ruling's transaction applies the
// raiser's statement; an answer that applies nothing — the statement already standing, or a side
// whose statement is the gap as it stands — lands in every sitting.
func requireSitsAfter(q recordsql.Querier, p Positions, d *recordpb.DocketRuling, applied bool) error {
	remand := d.GetDisposition() == recordpb.Disposition_DISPOSITION_REMANDED
	if !remand && !(applied && d.GetSidedWith() == recordpb.SidedWith_SIDED_WITH_RAISER) {
		return nil
	}
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	s, err := rulingSittingAt(q, evs, win)
	if err != nil {
		return err
	}
	if !remand {
		if lensSitsAfter(s.occasion, s.params, s.epoch) {
			return nil
		}
		return feov.Errorf(feov.Conflict,
			"no lens sits after this ruling to audit what --for-raiser would apply on %s: %s. Answer for the author, or remand where both parties sit after. Nothing was recorded",
			p.GapID, s.noLensWhy())
	}
	author, _, err := authorOfGap(q, p.GapID)
	if err != nil {
		return err
	}
	role := seatclass.RoleOfSeat(author)
	if partiesSitAfter(s.occasion, s.params, s.epoch, role) {
		return nil
	}
	why := s.noLensWhy()
	if lensSitsAfter(s.occasion, s.params, s.epoch) {
		why = fmt.Sprintf("%s, the author of %s's subject, sits after this ruling for its closings only, so the one more exchange a remand grants cannot be had", author, p.GapID)
	}
	return feov.Errorf(feov.Conflict,
		"a remand grants one more exchange, and both of %s's parties sit after the ruling to have it only where a plan follows: %s. Answer for a side. Nothing was recorded",
		p.GapID, why)
}

// noLensWhy is why no lens sits after a ruling in s, in words.
func (s rulingSitting) noLensWhy() string {
	switch s.occasion {
	case recordpb.Occasion_OCCASION_TERMINAL, recordpb.Occasion_OCCASION_ASSEMBLE:
		return fmt.Sprintf("this is the %s sitting, which no plan follows", recordpb.Word(s.occasion))
	}
	return fmt.Sprintf("the plan that convened this sitting is epoch %d under an epoch limit of %d, and no lens is dispatched after it but for its closings", s.epoch, s.params.MaxEpochs)
}
