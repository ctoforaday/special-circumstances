package record

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
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
// red's gap silently: red returns in the next epoch, sees a sentence it never audited, and cannot
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
//
// THE WALK IS TRANSITIVE, AND THAT IS WHY IT IS HERE RATHER THAN IN THE VIEW. Each edit is matched
// against the location AS IT STANDS after the edits before it, not against the text the gap was
// minted on. A sentence rewritten twice is the ordinary case — measured on universe-m11, G4 went
// Eight -> Seven -> Six — and matching everything against the minted text attributes only the first,
// leaving the location one edit behind the report and `edited_since` reporting one change of several.
//
// IT CANNOT BE THE VIEW'S JOB. Following a chain means choosing, at each step, the NEXT edit that
// overlaps the current location — which in SQL is a recursive CTE whose recursive term needs
// ORDER BY … LIMIT 1, and SQLite allows neither there; a plain UNION ALL enumerates orderings
// instead. The `gap_edit` view remains the SQL statement of the single-step rule, and of the
// visible-text comparison both sides share, but the ordered fold belongs to a language with a cursor.
func GapEdits(run Run) (map[string][]GapEdit, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return map[string][]GapEdit{}, err
	}
	// Each gap's minted location and the event that minted it: an edit BEFORE the mint is part of the
	// text red minted against, not a change to it.
	mints, err := db.Query(`SELECT m."gap_id", COALESCE(m."location", ''), me."id"
	  FROM "mint" m JOIN "events" me ON me."id" = m."event_id"
	  WHERE COALESCE(m."location", '') != '' ORDER BY me."id"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking which gaps have a location: %w", err)
	}
	type minted struct {
		gap  string
		loc  string
		when int64
	}
	var gaps []minted
	for mints.Next() {
		var m minted
		if err := mints.Scan(&m.gap, &m.loc, &m.when); err != nil {
			mints.Close()
			return nil, err
		}
		gaps = append(gaps, m)
	}
	mints.Close()
	if err := mints.Err(); err != nil {
		return nil, err
	}

	// Every edit that STANDS, in the order Live gives — a corrected edit is replaced by its
	// successor in the struck act's place, so the chain follows what the record now says happened.
	rows, err := db.Query(`SELECT e."id", e."epoch", e."seat_id", COALESCE(b."old", ''), COALESCE(b."new", '')
	  FROM "blue_edit" b
	  JOIN "live_event" l ON l."event_id" = b."event_id"
	  JOIN "events_w" e ON e."id" = b."event_id"
	  ORDER BY l."pos", b."event_id"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking for the edits: %w", err)
	}
	defer rows.Close()
	type edit struct {
		id  int64
		ge  GapEdit
		old string
	}
	var edits []edit
	for rows.Next() {
		var e edit
		var by sql.NullString
		if err := rows.Scan(&e.id, &e.ge.Epoch, &by, &e.old, &e.ge.New); err != nil {
			return nil, err
		}
		e.ge.EditedBy, e.ge.Old = by.String, e.old
		edits = append(edits, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := map[string][]GapEdit{}
	for _, g := range gaps {
		loc := g.loc
		for _, e := range edits {
			if e.id <= g.when || e.old == "" {
				continue
			}
			next, touched := relocate(loc, e.old, e.ge.New)
			if !touched {
				continue
			}
			out[g.gap] = append(out[g.gap], e.ge)
			loc = next
		}
	}
	return out, nil
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
		if next, touched := relocate(loc, e.Old, e.New); touched {
			loc = next
		}
	}
	return loc
}

// relocate is the ONE per-step rule: does this edit's replaced span and this location name the same
// text, and if so where does the location now live. GapEdits uses it to choose the next edit in a
// chain; CurrentLocation uses it to replay the chain it chose. Two callers, one definition, so the
// attribution and the relocation cannot disagree about what "this edit moved that text" means.
//
// THROUGH THE ANNOTATION LAYER, BOTH SIDES. A gap's stored location is the sentence as red quoted it;
// the edit's span is the sentence as the report holds it, anchor included — placed there by this
// gap's own mint. Compared raw, the two stop matching at the moment the gap exists.
// internal/record/locatorclass_test.go holds the whole class to this.
func relocate(loc, old, new string) (string, bool) {
	vLoc, vOld := anchortext.Visible(loc), anchortext.Visible(old)
	switch {
	case old == "" || vOld == "" || vLoc == "":
		return loc, false
	case strings.Contains(vOld, vLoc):
		// The edit REWROTE the sentence: the location becomes the replacement, because that is where
		// the text the gap points at now lives.
		return new, true
	case strings.Contains(vLoc, vOld):
		// A FRAGMENT INSIDE it changed: the location keeps its shape with the fragment substituted, so
		// it still names one sentence rather than widening to whatever else the edit touched. Try the
		// raw substitution first so a location that never met an anchor is returned byte for byte;
		// fall back to the visible form, which is what the seat has to quote anyway.
		if next := strings.Replace(loc, old, new, 1); next != loc {
			return next, true
		}
		return strings.Replace(vLoc, vOld, anchortext.Visible(new), 1), true
	}
	return loc, false
}
