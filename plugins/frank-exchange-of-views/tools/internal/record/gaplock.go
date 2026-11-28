package record

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// THE LOCK (plans/findings-fold.md §III.B5, picked 2026-11-25 "lock the anchor"): while a gap G2 about
// gap G is docketed, G is G2's subject before the bench, and its minter A changes it only by the act
// that concedes D(G2) — its accept of G2's raiser's statement, carrying answers = G2. A's amend,
// regrade and close of G, and its take of a proposal on G, are refused naming D(G2), at the docket's
// first or behind it. The bench's application of a ruling is the ruling's act and holds no lock.

// requireGapUnlocked refuses seatID's act on gapID while a docketing stands on an open gap about
// gapID, but where the act is the author's accept of that gap's statement (answers, accepted).
func requireGapUnlocked(q recordsql.Querier, seatID, gapID, answers string, accepted bool, act string) error {
	if gapID == "" || seatclass.RoleOfSeat(seatID) == "bench" {
		return nil
	}
	rows, err := q.Query(`SELECT m."gap_id", s."standing_docketing" FROM "mint" m JOIN "gap_state" s ON s."gap_id" = m."gap_id"
	  WHERE m."about_kind" = ? AND m."about_ref" = ? AND s."standing_docketing" IS NOT NULL ORDER BY m."event_id"`,
		recordpb.Word(recordpb.AboutKind_ABOUT_KIND_GAP), gapID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var g2, d string
		if err := rows.Scan(&g2, &d); err != nil {
			return err
		}
		if accepted && answers == g2 {
			continue
		}
		return feov.Errorf(feov.Conflict,
			"gap %s is locked: gap %s, about it, is docketed (%s), and %s's wording and grades are what the bench rules on there — your %s of it waits for that ruling. You concede by taking %s's raiser's statement (--accept %s); otherwise your position is your closing on %s, which the bench reads before it rules. Nothing was recorded",
			gapID, g2, d, gapID, act, g2, g2, g2)
	}
	return rows.Err()
}
