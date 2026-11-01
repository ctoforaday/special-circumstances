package record

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// WHEN THE CHAIR HOLDS BACK A PASS ITS PLAN PERMITS, THE BENCH HEARS THE CHAIR'S POSITION (gblock,
// 2026-10-30). Before this, nobody was readied, the position went unheard, and the run ended
// UNVERIFIED on a board that permitted VERIFIED.
//
//   - Holding back a permitted PASS — a chair sitting whose last stored plan permitted PASS, that
//     records none, and that files the owed position (heardPositionsOf) — readies the bench on that
//     position. The position is a motion of subject `position` whose filing is the position itself,
//     so it holds the PASS as an unruled motion and the record's read reports the hearing owed
//     (RecordRead.BenchOwed), and the engine convenes the position sitting, as for a petition.
//   - UPHELD: the bench names the work, and the debate continues on it. A lens or open gaps, chosen
//     per ruling (gblock, 2026-10-30): the lens owes an audit of its area on the ruling's direction
//     (an upheld_position blocker it owns, which readies it until it is served the report in a
//     sitting opened after the ruling); each gap goes back for one more exchange on the direction —
//     a remand, read off the same "remand" view as the docket's, so it is the gap's one remand at
//     impasse and the second-remand refusal stays one rule.
//   - OVERRULED: the chair's next sitting decides on the PASS. Its plan still permits the PASS; its
//     work list holds the verdict as a blocking item (OverruledVerdictOwed); a position filed there
//     while the PASS is permitted repeats the decided objection and is refused
//     (requireNoRepeatedPosition); and a sitting that closes with no verdict is short
//     (PositionVerdictMissing, which capture's record-parity audit fails).
//   - AT THE EPOCH LIMIT the chair's final sitting convenes nobody, so the terminal bench hears the
//     position at the exit and the run ends CEILING however it rules (epochLimitHeldTheBench).
//
// Either way the objection is decided, not dropped.

// PositionWork is what an upheld ruling on a position names: a lens, or open gaps, and the direction
// the work owes. Zero on every other ruling.
type PositionWork struct {
	Lens      string
	GapIDs    []string
	Direction string
}

// requirePositionRef refuses a position ruling whose id names no position the bench hears.
func requirePositionRef(q recordsql.Querier, id string) error {
	evs, win, err := eventsAt(q)
	if err != nil {
		return err
	}
	heard := heardPositionsOf(evs, win)
	for _, e := range evs {
		p, ok := recordpb.BodyAs[*recordpb.Position](e)
		if !ok || p.GetPositionId() != id {
			continue
		}
		if heard[e] {
			return nil
		}
		return fmt.Errorf("--id names position %s, which %s filed at a sitting that withheld no PASS its plan permitted — the bench hears a position only where the chair held back a PASS the plan permitted, and the transcript carries this one to the bench at the exit", id, e.GetSeatId())
	}
	return fmt.Errorf("--id names position %s, which no position on the record holds — a dangling reference is accepted here and dropped at replay", id)
}

// requirePositionHearing refuses a ruling on a position that names the wrong work (gblock,
// 2026-10-30): upheld names a cast lens or open gaps — one or the other — with a direction; overruled
// names none. A named gap the bench has already remanded at impasse is refused as a second remand
// at impasse is (remandPastLimit), and one with a docket motion standing unruled is before the bench
// already. Off under a migration, as a remand's checks are: an archived ruling is what the bench did.
func requirePositionHearing(w writeView, seatID string, h *recordpb.PositionHearing) error {
	lens, gaps := strings.TrimSpace(h.GetLens()), h.GetGapIds()
	switch h.GetRuling() {
	case recordpb.PositionRuling_POSITION_RULING_OVERRULED:
		if h.Lens != nil || len(gaps) > 0 || h.Direction != nil {
			return fmt.Errorf("--as %s names no work: the objection does not hold and the PASS goes back to the chair's next sitting, so --lens, --ids and --direction are an upheld ruling's", positionOverruled)
		}
		return nil
	case recordpb.PositionRuling_POSITION_RULING_UPHELD:
	default:
		return nil // CheckRequired refuses an absent word
	}
	if (lens != "") == (len(gaps) > 0) {
		return fmt.Errorf("--as %s names the work the debate continues on, and names it one way: --lens for a lens to audit its area, or --ids for open gaps, comma-separated, sent back for one more exchange — not both, and not neither", positionUpheld)
	}
	if strings.TrimSpace(h.GetDirection()) == "" {
		return fmt.Errorf("--as %s requires --direction: the work the ruling names owes it — what the lens audits its area for, or what each gap's one more exchange owes blue and its minting lens", positionUpheld)
	}
	if Migrating {
		return nil
	}
	if lens != "" {
		cast, err := castAt(w.q)
		if err != nil {
			return err
		}
		if seatclass.RoleOfSeat(lens) != "lens" || !slices.Contains(cast, lens) {
			var lenses []string
			for _, s := range cast {
				if seatclass.RoleOfSeat(s) == "lens" {
					lenses = append(lenses, s)
				}
			}
			return fmt.Errorf("--lens %q is no lens of this run's cast; the lenses are %s", lens, strings.Join(lenses, ", "))
		}
		return nil
	}
	seen := map[string]bool{}
	for _, g := range gaps {
		if seen[g] {
			return fmt.Errorf("--ids names %s twice; each gap goes back for one more exchange once", g)
		}
		seen[g] = true
		if err := requireGap(w.q, g, "--ids"); err != nil {
			return err
		}
		if err := requireOpenGap(w.q, g, "--ids",
			"a remand sends a gap back for one more exchange, begun by blue's act on it and audited by its minting lens after, and a closed gap has none to have"); err != nil {
			return err
		}
		var filed int64
		if found, err := queryRowAt(w.q, []any{&filed}, `SELECT COALESCE("unruled_docket_filed", 0) FROM "gap" WHERE "gap_id" = ?`, g); err != nil {
			return err
		} else if found && filed > 0 {
			return fmt.Errorf("--ids %s stands docketed with a motion you have not ruled: it is before you already, and its docket motion is where its fate is decided", g)
		}
	}
	p, err := w.files.params()
	if err != nil {
		return err
	}
	evs, win, err := eventsAt(w.q)
	if err != nil {
		return err
	}
	remands, err := remandRulingsOf(w.q)
	if err != nil {
		return err
	}
	exch := exchangesOf(evs, win.IDs(evs), win, p, WhileRunning, remands)
	for _, g := range gaps {
		if at, past := remandPastLimit(exch[g], remands[g], win.LatestSittingOf(seatID)); past {
			return feov.Errorf(feov.Conflict,
				"--ids %s names a gap the bench already remanded at impasse (event %d) — "+RemandLimitRule+". Nothing was recorded; name the work as a lens, or name other gaps",
				g, at)
		}
	}
	return nil
}

// requireNoRepeatedPosition refuses the chair's position at its first sitting after the bench
// overruled its last one, while that sitting's plan permits the PASS (gblock, 2026-10-30): the
// objection is decided, and a second position would hold the PASS back on it again. Refused whether
// or not the sitting has recorded its verdict yet: a FAIL there does not reopen the objection, and a
// position after it would be heard again on the same withholding.
func requireNoRepeatedPosition(w writeView, seatID string) error {
	if seatID != chairSeat || Migrating {
		return nil
	}
	evs, win, err := eventsAt(w.q)
	if err != nil {
		return err
	}
	if id, permitted, _ := overruledFollowerAt(evs, win); permitted {
		return feov.Errorf(feov.Conflict,
			"the bench overruled your position %s, and this is your first sitting since: the objection is decided, and a position here would hold the PASS back on it again. Nothing was recorded. "+
				"The sitting decides on the PASS with its verdict — the PASS, or a FAIL where the board as it stands still stops you", id)
	}
	return nil
}

// OverruledVerdictOwed is the position the bench overruled where the chair's latest sitting is its
// first since that ruling, the sitting's latest stored plan permits the PASS, and the sitting has
// recorded no verdict: that sitting owes the verdict (gblock, 2026-10-30). "" and false otherwise.
func OverruledVerdictOwed(evs []*Event, win WindowIndex) (string, bool) {
	return overruledVerdictOwed(evs, win)
}

func overruledVerdictOwed(evs []*Event, win WindowIndex) (string, bool) {
	id, permitted, verdict := overruledFollowerAt(evs, win)
	return id, permitted && !verdict
}

// overruledFollowerAt reads the chair's latest sitting where it is the chair's first after the bench
// overruled a position: that position's id, whether the sitting's latest stored plan permits the
// PASS, and whether the sitting has recorded a verdict. "", false, false for any other sitting.
func overruledFollowerAt(evs []*Event, win WindowIndex) (id string, permitted, verdict bool) {
	latest := win.LatestSittingOf(chairSeat)
	if latest == 0 {
		return "", false, false
	}
	ids := win.IDs(evs)
	_, openings := dispatchLedger(evs, ids, win)
	id, ok := overruledFollowers(motionsAt(evs, ids, win), openings[chairSeat])[latest]
	if !ok {
		return "", false, false
	}
	for _, e := range thisSitting(evs, win, chairSeat) {
		switch b := mustBodyOf(e).(type) {
		case *recordpb.Gate:
			verdict = true
		case *recordpb.Plan:
			if p, _, err := ReadPlan([]byte(b.GetPrinted())); err == nil {
				permitted = p.PassPermitted
			}
		}
	}
	return id, permitted, verdict
}

// upheldPositionBlockers is the gate list's upheld_position arm: each position the bench upheld
// naming a lens, owned by that lens while it has been served the report in no sitting opened after
// the ruling. ids is evs's places, aligned with it.
func upheldPositionBlockers(evs []*Event, ids []int64, win WindowIndex) []Blocker {
	var out []Blocker
	var reads map[string][]served
	for _, m := range motionsAt(evs, ids, win) {
		if m == nil || m.Subject != positionSubject || m.Ruling != positionUpheld || m.Work.Lens == "" {
			continue
		}
		if reads == nil {
			reads = servedOf(evs, ids, win)
		}
		if slices.ContainsFunc(reads[m.Work.Lens], func(r served) bool { return r.opening > m.ruled }) {
			continue
		}
		out = append(out, Blocker{Kind: BlockerUpheldPosition, Subject: m.ID, Owner: m.Work.Lens, Detail: m.Work.Direction, since: m.ruled})
	}
	return out
}

// overruledVerdictItem is the chair's work-list item in its first sitting after the bench overruled
// its position.
func overruledVerdictItem(positionID string) string {
	return fmt.Sprintf("the bench overruled your position %s on the PASS you withheld: the objection is decided, and this sitting decides on the PASS with the verdict your plan permits — the PASS, or a FAIL where the board as it stands still stops you. A position here would withhold the PASS on the same objection again, and is refused", positionID)
}

// upheldAuditItem is the named lens's work-list item for an upheld position: its audit of its area.
func upheldAuditItem(b Blocker) string {
	return fmt.Sprintf("the bench upheld the chair's position %s on the PASS it withheld and named your area: audit it on the ruling's direction — %q. Read the report as it now stands; a defect you find there is a finding and a gap. PASS is refused until you have been served the report since the ruling", b.Subject, b.Detail)
}

// upheldPositionGaps is the open gaps an upheld position ruling names — each one remanded, as a
// docket's remand is — and nil for any other motion or ruling.
func upheldPositionGaps(m *Motion) []string {
	if m == nil || m.Subject != positionSubject || m.Ruling != positionUpheld {
		return nil
	}
	return m.Work.GapIDs
}
