package record

import (
	"database/sql"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// EVERY STATEMENT STANDING ON A GAP HOLDS UNTIL ITS RULING (#1695 §III.B14, L45). A docketing is
// answered by applying the statement of the side the bench names, so each statement standing on G must
// still apply when the bench sits: red's (the assignee's) closing's road, from the closing until the
// answer of the docketing or proposal it argues. Blue's edit and retire, and the text the bench's answer
// for red applies, re-plan each standing road on the report they leave (StandingRoads, planned by the
// blue verbs with the accept's own planner, asked of the record with the act on it), and an act after
// which a road no longer plans is refused naming its docketing or proposal.
//
// A raw row appended below the verbs is protected as Stale and Plan are, by no verb mounting it.

// Road is a statement of text standing on a gap until its ruling (L45): red's closing's road — its
// pair, or the retire of the anchor the gap is about. By names the docketing or proposal it argues,
// which a refusal names.
type Road struct {
	By, GapID        string
	Location, FixNew string
	// Retire is the anchor a closing's retire road retires; "" for a pair.
	Retire string
}

// StandingRoads is every road standing on the record the write lands on: red's closing's road on each
// gap a docketing or a proposal stands on, from the closing until the answer.
func (l *Landing) StandingRoads() ([]Road, error) { return standingRoadsAt(l.w.q) }

// standingRoadsAt is StandingRoads asked of q.
func standingRoadsAt(q recordsql.Querier) ([]Road, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return nil, err
	}
	ids := win.IDs(evs)
	var out []Road
	seen := map[string]bool{} // gaps a standing docketing's road was read for
	for _, k := range docketingsAt(evs, ids) {
		if !k.Standing() {
			continue
		}
		p, found, err := positionsAt(q, k.ID)
		if err != nil || !found {
			return nil, err
		}
		seen[k.GapID] = true
		if r, ok := roadOf(p.Road, k.ID, k.GapID, p.AboutRef); ok {
			out = append(out, r)
		}
	}
	// A PROPOSAL IN ITS ASSIGNEE'S WINDOW: the assignee's closing argues it from its filing (L30).
	assignee := assigneeOf(evs)
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != "docket" || m.GapID == "" || m.Ruled() || seen[m.GapID] {
			continue
		}
		_, proposal, err := standingOnGap(q, m.GapID)
		if err != nil {
			return nil, err
		}
		if proposal != m.ID {
			continue
		}
		var road *recordpb.Closing
		for i, e := range evs {
			if ids[i] <= m.filed || e.GetSeatId() != assignee[m.GapID] {
				continue
			}
			if c, ok := mustBody(e).(*recordpb.Closing); ok && c.GetGapId() == m.GapID && road == nil {
				road = c
			}
		}
		var kind, ref sql.NullString
		if road.GetRetire() {
			if _, err := queryRowAt(q, []any{&kind, &ref}, `SELECT "about_kind", "about_ref" FROM "mint" WHERE "gap_id" = ?`, m.GapID); err != nil {
				return nil, err
			}
		}
		if r, ok := roadOf(road, m.ID, m.GapID, ref.String); ok {
			out = append(out, r)
		}
	}
	return out, nil
}

// roadOf is the road a closing states, false where it states none.
func roadOf(c *recordpb.Closing, by, gapID, aboutRef string) (Road, bool) {
	switch {
	case c == nil:
	case c.Location != nil && c.FixNew != nil:
		return Road{By: by, GapID: gapID, Location: c.GetLocation(), FixNew: c.GetFixNew()}, true
	case c.GetRetire() && aboutRef != "":
		return Road{By: by, GapID: gapID, Retire: aboutRef}, true
	}
	return Road{}, false
}

// roadEstoppel refuses a closing's road whose quote a standing estoppel would refuse its raiser's mint
// of (L43): the raiser states the road, and estoppel binds only the lens that prescribed the applied
// text (X11), so the road is asked as a mint of gapID's minter. evs is the stream it is asked of.
func roadEstoppel(f Family, evs []*Event, gapID, loc string) error {
	f.Events = evs
	raiser := mintersOf(evs)[gapID]
	est, ok := EstoppelConflict(f, raiser, loc)
	if !ok {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"red's text on %s quotes text its raiser %s prescribed on %s and blue applied verbatim — %q — and a lens is bound by the fix it prescribed: its mint of that quote is refused, and so is this road. "+
			"State a road outside that text, or file --no-text where every road you can state is refused. Nothing was recorded",
		gapID, raiser, est.Gap, est.Text)
}
