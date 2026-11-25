package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A MINT SUPERSEDING AN OPEN GAP CLOSES IT (ruled by gblock 2026-11-04). The ancestor's purpose
// passes to the successor in the act that mints it, so the ancestor has no state of its own left to
// hold the board. universe-m38's architecture lens superseded G-29f07d26 to correct an acceptance
// check that could never pass, had no close for "moved to the successor", and the open ancestor
// refused every verdict of the run.
//
// THE CLOSING IS DERIVED, NEVER WRITTEN. The mint's `supersedes` is the one statement of the link;
// the gap state's refresh reads it as a closing arm (recordsql/gapstate.go, gap_refresh's supersede_event),
// and supersessionsOf is the same reading over a stream. A Close the tool wrote beside the mint
// would be a second copy of the fact, and an act the seat never stated.
//
// THE FIRST MINT CLOSES, AND ONLY AN OPEN GAP. A mint naming a gap a close or a bench ruling already
// closed carries the lineage forward and closes nothing; one naming a gap an earlier mint superseded
// is the same.

// supersessionsOf is, for each gap a mint closed by superseding it, the index in evs of that mint:
// the first mint naming the gap in `supersedes` while no close and no closing docket ruling of the
// gap precedes it. It is gap_refresh's supersede_event, read over a stream.
func supersessionsOf(evs []*Event) map[string]int {
	docketGap := DocketGapByMotion(evs)
	closed := map[string]bool{}
	out := map[string]int{}
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Close:
			closed[b.GetGapId()] = true
		case *recordpb.MotionRule:
			if d := b.GetDocket(); d != nil && recordpb.Closes(d.GetDisposition()) {
				closed[docketGap[b.GetMotionId()]] = true
			}
		case *recordpb.Mint:
			for _, anc := range b.GetSupersedes() {
				if !closed[anc] {
					closed[anc] = true
					out[anc] = i
				}
			}
		}
	}
	return out
}

// successorClosing is the successor whose mint closed gap id, "" where no supersession closed it —
// the gap view's `superseded_by`.
func successorClosing(q recordsql.Querier, id string) (string, error) {
	var by string
	_, err := queryRowAt(q, []any{&by}, `SELECT COALESCE("superseded_by", '') FROM "gap" WHERE "gap_id" = ?`, id)
	return by, err
}

// SupersededGap reports whether a successor's mint closed gap id by superseding it.
func SupersededGap(q recordsql.Querier, id string) (bool, error) {
	by, err := successorClosing(q, id)
	return by != "", err
}

// supersededRefusal words the refusal of an act on gap id, which the mint of successor closed by
// superseding it; flag is the flag that named the gap.
func supersededRefusal(flag, id, successor string) error {
	return feov.Errorf(feov.Conflict,
		"%s names gap %s, which closed when %s was minted superseding it: its purpose passed to %s, and %s holds no state of its own. Act on %s",
		flag, id, successor, successor, id, successor)
}

// requireSupersedable is the write path's check of a mint's `supersedes`, ancestor by ancestor: a
// mint superseding an OPEN gap closes it, so it is held to what a close is held to. Only the gap's
// minter closes it (requireOriginator's rule). A closed ancestor is lineage only, and any seat may
// name it.
func requireSupersedable(q recordsql.Querier, seatID string, m *recordpb.Mint) error {
	for _, anc := range m.GetSupersedes() {
		closed, err := gapState(q, anc)
		if err != nil {
			return err
		}
		if closed {
			continue
		}
		var mintedBy string
		if _, err := queryRowAt(q, []any{&mintedBy}, `SELECT COALESCE("minted_by", '') FROM "gap" WHERE "gap_id" = ?`, anc); err != nil {
			return err
		}
		if mintedBy != "" && mintedBy != seatID {
			return feov.Errorf(feov.Validation,
				"--supersedes names %s, which is open and was minted by %s. A mint superseding an open gap closes it, and only its minter closes a gap; %s cannot. "+
					"If %s is your defect, file a finding about it; if yours is a different defect, mint it without --supersedes. Nothing was recorded",
				anc, mintedBy, seatID, anc)
		}
	}
	return nil
}

// requireSupersededOffDocket refuses a mint superseding an open gap a docket motion stands unruled
// on, as a close of it is refused (requireNoStandingDocket): the bench disposes of a docketed gap. An
// act judgement (judgeAct): an archived lens that superseded a docketed gap superseded it.
func requireSupersededOffDocket(q recordsql.Querier, m *recordpb.Mint) error {
	for _, anc := range m.GetSupersedes() {
		var motion string
		if _, err := queryRowAt(q, []any{&motion}, `SELECT COALESCE("unruled_docket_motion", '') FROM "gap" WHERE "gap_id" = ? AND "open"`, anc); err != nil {
			return err
		}
		if motion != "" {
			return feov.Errorf(feov.Conflict,
				"--supersedes names %s, which is docketed for the bench: motion %s stands unruled on it, and a mint superseding an open gap closes it, while a docketed gap is the bench's to dispose of until it rules. "+
					"Nothing was recorded; what you found goes in your closing argument on %s, which the bench reads before it rules",
				anc, motion, anc)
		}
	}
	return nil
}

// requireNotSuperseded refuses an act on a gap a supersession closed, naming its successor.
func requireNotSuperseded(q recordsql.Querier, id, flag string) error {
	if id == "" {
		return nil
	}
	by, err := successorClosing(q, id)
	if err != nil || by == "" {
		return err
	}
	return supersededRefusal(flag, id, by)
}
