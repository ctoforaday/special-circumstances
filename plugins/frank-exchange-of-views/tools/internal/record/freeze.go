package record

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"google.golang.org/protobuf/proto"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchortext"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// EVERY STATEMENT STANDING ON A GAP HOLDS UNTIL ITS RULING (#1695 §III.B14, L45, L48). A docketing is
// answered by applying the statement of the side the bench names, so each statement standing on G must
// still apply when the bench sits: red's (the assignee's) closing's road, from the closing until the
// answer of the docketing or proposal it argues, and a standing supersede proposal's successor, from
// its filing. Two arms hold them:
//
//   - THE ROADS (L45): blue's edit and retire, and the text the bench's answer for red applies, re-plan
//     each standing road on the report they leave (StandingRoads, planned by the blue verbs with the
//     accept's own planner, asked of the record with the act on it), and an act after which a road no
//     longer plans is refused naming its docketing or proposal;
//   - THE SUCCESSOR (L48): a supersede proposal's successor is checked mintable against G's assignee
//     at the filing (requireSuccessorMintable, called from requireDocketFiling), and every act that can
//     move one of those checks' inputs — blue's edit and retire, the bench's applied edit and retire,
//     G's assignee's verify, reproduce and mint — is judged again with the act on the record
//     (judgeLanded), refused naming the proposal and the check where the successor is no longer one
//     the bench could write.
//
// A raw row appended below the verbs is protected as Stale and Plan are, by no verb mounting it.

// requireSuccessorMintable is every check the bench's write of the successor m runs at a for-blue
// (for-filer) answer, asked now against G's assignee — the acting seat the bench's mint is checked as
// (mintActingSeat, L5): judgeAct's Mint arm (judgeMint), validate's Mint arm that turns on the acting
// seat (requireSupersedable, requireSuccessorKeepsBenchGrades), the placing of its marker at its quote,
// and the estoppel guard on its quote. motionID is the proposal that states m, whose own standing (and
// its PROPOSAL docketing's) the supersede does not meet: "" at the filing, which is before it stands.
// The fix the successor prescribes is planned by the verbs (`blue.PlanAccept`), the one check the
// record cannot ask.
func requireSuccessorMintable(w writeView, m *recordpb.Mint, motionID string) error {
	q := w.q
	g := m.GetSupersedes()
	var assignee string
	if _, err := queryRowAt(q, []any{&assignee}, `SELECT COALESCE("assignee", '') FROM "gap" WHERE "gap_id" = ?`, g); err != nil {
		return err
	}
	if assignee == "" {
		return nil // an unknown gap is requireGap's refusal
	}
	// EACH CHECK IS NAMED IN WORDS, NOT BY ITS OWN REFUSAL: that refusal is worded for the minting seat
	// and names its verbs, and the seat told here is the filer, or the seat whose act moved an input.
	checks := []struct {
		what string
		ask  func() error
	}{
		{"the anchor it is about is retired", func() error { return requireAboutAnchorLive(q, m.GetAboutKind(), m.GetAboutRef()) }},
		{"its class stands on a reading of its subject, and " + assignee + ", the gap's assignee, holds no such standing reading", func() error { return requireOwnStandingReading(w, assignee, m) }},
		{"an open gap of its class, or a ruling's tombstone of one, stands on its subject, and one class on one subject is one gap", func() error { return requireOneGapPerClassAndSubject(w, assignee, m) }},
		{"another proposal or a docketing stands on " + g, func() error { return requireSupersededOffDocket(q, m, motionID) }},
		{"it drops the live answer " + g + " demands", func() error { return requireSuccessorCarriesDemand(q, m) }},
		{"it is written in the run's own voice", func() error { return refuseMintReportVoice(m) }},
		{g + " is open and " + assignee + "'s to close", func() error { return requireSupersedable(q, assignee, m) }},
		{"it moves a grade a ruling holds locked on " + g, func() error { return requireSuccessorKeepsBenchGrades(w, m) }},
		{"its quote places nowhere on the report as it stands", func() error {
			if m.GetLocation() == "" {
				return nil
			}
			report, err := w.report()
			if err != nil {
				return err
			}
			_, err = anchortext.Attach(report, m.GetGapId(), m.GetLocation())
			return err
		}},
	}
	for _, c := range checks {
		if err := c.ask(); err != nil {
			return successorCheck{what: c.what, err: err}
		}
	}
	if loc := m.GetLocation(); loc != "" {
		fam, err := familyAt(q)
		if err != nil {
			return err
		}
		if est, ok := EstoppelConflict(fam, loc); ok && est.Gap != g {
			return successorCheck{what: fmt.Sprintf("its quote is text %s prescribed on %s and blue applied verbatim, which estops a mint of it", est.PrescribedBy, est.Gap)}
		}
	}
	return nil
}

// successorCheck is a check the successor fails: what it is, in words, and the check's own refusal.
type successorCheck struct {
	what string
	err  error
}

func (e successorCheck) Error() string { return e.what }
func (e successorCheck) Unwrap() error { return e.err }

// successorRefusal is the filing's refusal of a supersede whose successor the bench could not write.
func successorRefusal(g string, m *recordpb.Mint, err error) error {
	return feov.Wrap(feov.Validation, err,
		"the successor you propose on %s is one its assignee's mint could not write, and the bench writes it as the assignee's on siding with you: %v. "+
			"Nothing was recorded. Your other roads: a close proposal on %s, an edit of the report, or a supersede the assignee's own readings and gaps admit",
		g, err, g)
}

// standingSupersede is a supersede proposal standing on a gap: the motion and the successor it states.
type standingSupersede struct {
	motionID, gapID, assignee string
	mint                      *recordpb.Mint
}

// standingSupersedesAt is every supersede proposal standing on q's record — unanswered, in its
// assignee's window or held by its PROPOSAL docketing.
func standingSupersedesAt(q recordsql.Querier) ([]standingSupersede, error) {
	evs, win, err := eventsAt(q)
	if err != nil {
		return nil, err
	}
	var out []standingSupersede
	for _, m := range motionsAt(evs, win.IDs(evs), win) {
		if m == nil || m.Subject != "docket" || m.GapID == "" || m.Ruled() {
			continue
		}
		dp, found, err := docketProposalAt(q, m.ID)
		if err != nil {
			return nil, err
		}
		s, ok := StoredAct(dp.Motion).(*recordpb.Mint)
		if !found || !ok {
			continue
		}
		docketing, proposal, err := standingOnGap(q, m.GapID)
		if err != nil {
			return nil, err
		}
		if proposal != m.ID {
			if docketing == "" {
				continue
			}
			k, _, err := docketingAt(q, docketing)
			if err != nil {
				return nil, err
			}
			if k.MotionID != m.ID {
				continue
			}
		}
		out = append(out, standingSupersede{motionID: m.ID, gapID: m.GapID, assignee: dp.Assignee, mint: s})
	}
	return out, nil
}

// judgedLanded reports whether the freeze's successor arm judges body written by seatID, and whether
// only the supersedes standing on a gap seatID is the assignee of: blue's edit or retire, the bench's
// applied edit or retire, and a lens's verify, reproduce or mint.
func judgedLanded(seatID string, body proto.Message) (judged, assigneeOnly bool) {
	role := seatclass.RoleOfSeat(seatID)
	switch body.(type) {
	case *recordpb.BlueEdit, *recordpb.Retire:
		return role == "blue" || role == seatclass.Lane.Role || role == "bench", false
	case *recordpb.Verify, *recordpb.Reproduce, *recordpb.Mint:
		return role == "lens", true
	}
	return false, false
}

// unwritableSuccessors is the standing supersede proposals whose successor the bench could not write
// on the record a write begins on, read only where the arm judges one of its acts: the motions
// judgeLanded leaves to the act that moved their inputs.
//
// ANOTHER SEAT'S ACT CAN MOVE ONE. The one-gap rule reads every open gap of the class and each ruling's
// tombstone, so another lens's mint, or the bench's ruling, can leave M's successor unwritable while
// the arm judges neither; the assignee's next act that moves nothing of M's is still its own.
func unwritableSuccessors(w writeView, seatID string, acts []Act) (map[string]bool, error) {
	if !slices.ContainsFunc(acts, func(a Act) bool { judged, _ := judgedLanded(seatID, a.Body); return judged }) {
		return nil, nil
	}
	standing, err := standingSupersedesAt(w.q)
	if err != nil {
		return nil, err
	}
	var out map[string]bool
	for _, s := range standing {
		err := requireSuccessorMintable(w, s.mint, s.motionID)
		var c successorCheck
		switch {
		case err == nil:
			continue
		case !errors.As(err, &c):
			return nil, err
		}
		if out == nil {
			out = map[string]bool{}
		}
		out[s.motionID] = true
	}
	return out, nil
}

// judgeLanded is the freeze's successor arm (L48), asked of the record with the act on it: an act that
// can move an input of a standing supersede's checks — blue's edit or retire, the bench's applied edit
// or retire, or G's assignee's own
// verify, reproduce or mint — is refused where the successor is no longer one the bench could write.
// unwritable is the motions whose successor already failed when the write began
// (unwritableSuccessors), which no act of the write moved.
func judgeLanded(w writeView, seatID string, body proto.Message, unwritable map[string]bool) error {
	judged, assigneeOnly := judgedLanded(seatID, body)
	if !judged {
		return nil
	}
	standing, err := standingSupersedesAt(w.q)
	if err != nil {
		return err
	}
	for _, s := range standing {
		if assigneeOnly && s.assignee != seatID || unwritable[s.motionID] {
			continue
		}
		if err := requireSuccessorMintable(w, s.mint, s.motionID); err != nil {
			return feov.Wrap(feov.Conflict, err,
				"proposal %s stands on %s, and the successor it states (%s) is one the bench writes as %s's on siding with it: after this %s it could not be written — %v. "+
					"Every statement standing on %s holds until its ruling, so nothing was recorded; argue %s in your closing instead",
				s.motionID, s.gapID, s.mint.GetGapId(), s.assignee, recordpb.Word(typeOf(body)), err, s.gapID, s.motionID)
		}
	}
	return nil
}

// typeOf is body's event type, zero where it is none.
func typeOf(body proto.Message) recordpb.EventType {
	t, _ := recordpb.SetBody(&Event{}, body)
	return t
}

// Road is a statement of text standing on a gap until its ruling (L45): red's closing's road — its
// pair, or the retire of the anchor the gap is about — or a supersede proposal's stored quote and fix.
// By names the docketing or proposal it argues, which a refusal names.
type Road struct {
	By, GapID string
	// Successor is the gap id a supersede proposal's road is the successor of; "" for a closing's.
	Successor        string
	Location, FixNew string
	// Retire is the anchor a closing's retire road retires; "" for a pair.
	Retire string
}

// StandingRoads is every road standing on the record the write lands on: red's closing's road on each
// gap a docketing or a proposal stands on, from the closing until the answer, and each standing
// supersede proposal's stored quote.
func (l *Landing) StandingRoads() ([]Road, error) { return standingRoadsAt(l.w.q) }

// AnchorLive reports whether ref is an anchor no retire has taken, on the record the write lands on.
func (l *Landing) AnchorLive(ref string) (bool, error) {
	err := requireAboutAnchorLive(l.w.q, recordpb.AboutKind_ABOUT_KIND_ANCHOR, ref)
	return err == nil, nil
}

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
	supersedes, err := standingSupersedesAt(q)
	if err != nil {
		return nil, err
	}
	for _, s := range supersedes {
		if s.mint.GetLocation() == "" {
			continue
		}
		out = append(out, Road{By: s.motionID, GapID: s.gapID, Successor: s.mint.GetGapId(), Location: s.mint.GetLocation(), FixNew: s.mint.GetFixNew()})
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

// roadEstoppel refuses a closing's road whose quote a standing estoppel would refuse a red mint of
// (L43): the prescription it relitigates, named with its author. evs is the stream it is asked of.
func roadEstoppel(f Family, evs []*Event, gapID, loc string) error {
	f.Events = evs
	est, ok := EstoppelConflict(f, loc)
	if !ok {
		return nil
	}
	return feov.Errorf(feov.Conflict,
		"red's text on %s quotes text %s prescribed on %s and blue applied verbatim — %q — and red is bound by the fix it prescribed: a mint of that quote is refused, and so is this road. "+
			"State a road outside that text, or file --no-text where every road you can state is refused. Nothing was recorded",
		gapID, est.PrescribedBy, est.Gap, est.Text)
}
