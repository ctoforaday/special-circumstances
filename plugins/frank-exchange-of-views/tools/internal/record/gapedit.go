package record

import (
	"database/sql"
	"strings"
)

// A GAP'S LOCATION IS CARRIED THROUGH THE EDITS, not frozen at mint and not merely flagged.
//
// The report is a frozen base plus an ordered stack of tool-made ops (#709), and every op is on
// the record with its exact old span and replacement. So a location captured at mint can be
// REPLAYED forward: the same fold that reconstructs the report reconstructs where the sentence
// went. Nothing needs a marker in the text and nothing needs a second copy of the location — the
// answer is derived from acts already recorded.
//
// The alternative shipped first and was wrong: a `location_stale` flag, which tells a reader the
// pointer is broken without repairing it, and leaves red re-auditing a gap to GUESS what blue
// changed underneath it. The flag is a measurement of the defect, not a fix for it.
//
// WHY THE HISTORY IS KEPT AND NOT JUST THE ANSWER. A relocated pointer alone would let blue move
// red's gap silently: red returns in the next round, sees a sentence it never audited, and cannot
// tell whether the text drifted or was rewritten under it. The edits ARE the answer to that, and
// they are already on the record — see the gap_edit view, which states the join rule once where
// every reader can see it.

// GapEdit is one edit that touched a gap's sentence, in the order it happened.
type GapEdit struct {
	Epoch    int    `json:"epoch"`
	EditedBy string `json:"edited_by"`
	// Old and New are the exact span replaced and what replaced it — the same pair blue typed
	// and the tool matched against the report, so a reader can see the change rather than be
	// told one happened.
	Old string `json:"old"`
	New string `json:"new"`
}

// GapEdits returns every edit that touched each gap's location, keyed by gap id, in record order.
//
// A gap with no location (one anchored by --about) and a gap nothing has edited both answer with
// no entry — which is the same answer, and correctly so: neither has a change history.
func GapEdits(run Run) (map[string][]GapEdit, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return map[string][]GapEdit{}, err
	}
	rows, err := db.Query(`SELECT "gap_id", "epoch", "edited_by", "old", "new"
	  FROM "gap_edit" ORDER BY "event_id"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]GapEdit{}
	for rows.Next() {
		var id string
		var ge GapEdit
		var by, old, new sql.NullString
		if err := rows.Scan(&id, &ge.Epoch, &by, &old, &new); err != nil {
			return nil, err
		}
		ge.EditedBy, ge.Old, ge.New = by.String, old.String, new.String
		out[id] = append(out[id], ge)
	}
	return out, rows.Err()
}

// CurrentLocation replays a gap's minted location through the edits that touched it and returns
// where that text is now.
//
// THE TWO SHAPES ARE DIFFERENT ACTS AND ARE FOLDED DIFFERENTLY:
//
//   - the edit REWROTE the sentence (old contains location) — the location becomes the
//     replacement, because that is where the text the gap points at now lives;
//   - the edit changed a FRAGMENT inside it (location contains old) — the location keeps its
//     shape with the fragment substituted, so it still names one sentence rather than widening
//     to whatever else the edit touched.
//
// Anything else is not in the view: gap_edit only joins edits whose span contains the location or
// is contained by it, and containment either way is what "this edit moved that text" means when
// both spans are exact quotes the tool matched.
//
// An empty location stays empty. Replaying nothing onto nothing is not a relocation.
func CurrentLocation(minted string, edits []GapEdit) string {
	loc := minted
	if strings.TrimSpace(loc) == "" {
		return loc
	}
	for _, e := range edits {
		switch {
		case e.Old == "":
			continue
		case strings.Contains(e.Old, loc):
			loc = e.New
		case strings.Contains(loc, e.Old):
			loc = strings.Replace(loc, e.Old, e.New, 1)
		}
	}
	return loc
}
