package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// The readers migrate's Translation 3 asks of the record it is replaying (EventSchema 80): an
// archived docket the dispatch filed becomes the Docketing the dispatch writes now, and its ruling the
// ruling of that docketing. Each answers off the record as the replay has written it, by the rules the
// live write path holds — no reader here is a tolerance the live path lacks.

// DocketingCauseAt is the cause a docketing of gapID takes on q by the dispatch's rule: grade
// movement where the board's movement past the threshold carries movement on gapID the bench has not
// ruled on (gradeMovementOf's docketed, which routeOf reads), else impasse. It names the cause of a
// docketing written there, and does not ask whether the dispatch would write one: an archived dispatch
// docketed at impasse before the lens's re-audit that lensGoesFirst now asks for (capture's
// EarlyDockets names each). A PROPOSAL docketing is written at a proposal's window end, and no archived
// docket held a proposal.
func DocketingCauseAt(q recordsql.Querier, gapID string) (recordpb.DocketingCause, error) {
	r, err := planReadsAt(q)
	if err != nil || r.cast == nil {
		return recordpb.DocketingCause_DOCKETING_CAUSE_UNSPECIFIED, err
	}
	moved := gradeMovementOf(r.evs, r.win.IDs(r.evs), r.win, openSet(blockerGapsOfOpen(r.gaps, nil)))
	if moved.docketed(gapID) {
		return recordpb.DocketingCause_DOCKETING_CAUSE_GRADE_MOVEMENT, nil
	}
	return recordpb.DocketingCause_DOCKETING_CAUSE_IMPASSE, nil
}

// DocketingStands reports whether docketing id stands on q: written, unruled, and not ended by a
// Close of its gap (gap_state.standing_docketing). The trigger motion_rule_docket_answers_a_filing
// refuses a ruling naming any other.
func DocketingStands(q recordsql.Querier, id string) (bool, error) {
	var gap string
	found, err := queryRowAt(q, []any{&gap}, `SELECT "gap_id" FROM "docketing" WHERE "docketing_id" = ?`, id)
	if err != nil || !found {
		return false, err
	}
	standing, _, err := standingOnGap(q, gap)
	return standing == id, err
}

// GapOpen reports whether gap id is open on q; an unknown gap is not.
func GapOpen(q recordsql.Querier, id string) (bool, error) {
	var open bool
	found, err := queryRowAt(q, []any{&open}, `SELECT "open" FROM "gap" WHERE "gap_id" = ?`, id)
	if err != nil || !found {
		return false, err
	}
	return open, nil
}
