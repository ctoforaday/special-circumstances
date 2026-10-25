package record

import (
	"fmt"
	"math"
	"sort"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/feov"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// A BENCH REMAND SENDS ITS GAP BACK TO THE DEBATE FOR ONE MORE EXCHANGE (gblock, 2026-09-29: a
// remand sends the gap back with a stated research direction, and the dispatch readies the minting
// lens and blue for one more exchange on it).
//
// WHAT COUNTS AS A REMAND. A bench sitting that remanded the gap WHILE IT WAS AT IMPASSE, once per
// sitting however many of its docket motions that sitting ruled. A remand of a docket a party filed
// early, ruled while the gap was below its limits, sends nothing anywhere the debate was not already
// going, so it spends nothing; two motions remanded in one sitting are one hearing of the gap.
//
// WHAT THE REMAND'S EXCHANGE DECIDES. The exchange count stays monotone. After a counted remand the
// gap stays at impasse until the first exchange that begins after the ruling: if that exchange
// leaves it unmoved it is still where the bench found it, and the dispatch dockets it again; if it
// moves the gap, the gap leaves impasse and the run's terms (k, kMax) are counted afresh from the
// ruling — the route by which the gap first reached impasse does not matter, so a gap that reached
// the bench on kMax and moves on its remand is not docketed again for a count it can never undo.
// A second counted remand leaves the gap open at its limit, and those gaps are what CEILING is made
// of. So a gap reaches the bench at most twice through the dispatch, and the run cannot cycle
// between the bench and the debate.

// remandRow is one row of the "remand" view: a remand ruling on a gap.
type remandRow struct {
	eventID, sitting int64
	direction        string
}

// remandRulingsOf reads the "remand" view, by gap, in record order.
func remandRulingsOf(db recordsql.Querier) (map[string][]remandRow, error) {
	rows, err := db.Query(`SELECT "gap_id", "event_id", "sitting_id", COALESCE("direction", '') FROM "remand" ORDER BY "event_id"`)
	if err != nil {
		return nil, fmt.Errorf("asking the record for the bench's remands: %w", err)
	}
	defer rows.Close()
	out := map[string][]remandRow{}
	for rows.Next() {
		var g string
		var r remandRow
		if err := rows.Scan(&g, &r.eventID, &r.sitting, &r.direction); err != nil {
			return nil, err
		}
		out[g] = append(out[g], r)
	}
	return out, rows.Err()
}

// remandSittingsOf groups a gap's remand rulings by the bench sitting that ruled them, one place
// per sitting — its latest remand ruling's — in record order.
func remandSittingsOf(rows []remandRow) []int64 {
	at := map[int64]int64{}
	for _, r := range rows {
		at[r.sitting] = max(at[r.sitting], r.eventID)
	}
	out := make([]int64, 0, len(at))
	for _, p := range at {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// countedRemands is the remand sittings that found the gap at impasse, each judged with the remands
// counted before it — the second remand is heard at the impasse the first one's exchange left.
func countedRemands(rounds []exchangeRound, sittings []int64, p Params) []int64 {
	var counted []int64
	for _, at := range sittings {
		if impasseOf(rounds, counted, p, at) {
			counted = append(counted, at)
		}
	}
	return counted
}

// impasseOf is the run's terms applied to the exchanges that ended before upto, under the remands
// counted before it. With no remand it is Stalled >= K or Exchanges >= KMax. After one, it is the
// remand's exchange first: none had since the ruling, or the first one left the gap unmoved, and the
// gap is still at impasse; that exchange moved it, and the terms are counted from the ruling.
func impasseOf(rounds []exchangeRound, remands []int64, p Params, upto int64) bool {
	ruled, remanded := int64(math.MinInt64), false
	for _, at := range remands {
		if at < upto {
			ruled, remanded = at, true
		}
	}
	var since []exchangeRound
	for _, r := range rounds {
		if r.end < upto && r.start > ruled {
			since = append(since, r)
		}
	}
	if remanded && (len(since) == 0 || !since[0].moved) {
		return true
	}
	stalled := 0
	for _, r := range since {
		if r.moved {
			stalled = 0
		} else {
			stalled++
		}
	}
	return stalled >= p.K || len(since) >= p.KMax
}

// latestRemandOwed reports whether the latest of a gap's remand sittings has had no exchange begun
// since its ruling: the one statement of "its remand's exchange is owed", which the counted stage
// (remandStageOf) and a remand of a gap that is not material (routeOf) both read.
func latestRemandOwed(rounds []exchangeRound, sittings []int64) bool {
	if len(sittings) == 0 {
		return false
	}
	ruled := sittings[len(sittings)-1]
	for _, r := range rounds {
		if r.start > ruled {
			return false
		}
	}
	return true
}

// remandsOwedOf is each gap whose latest remand's exchange is owed (latestRemandOwed), off the
// stream's exchanges and the "remand" view's rows — what the PASS gate's list reads where it holds
// no run terms, which the question does not need: whether an exchange has begun since a ruling is
// the same under every k.
func remandsOwedOf(evs []*Event, ids []int64, win WindowIndex, remands map[string][]remandRow) map[string]bool {
	out := map[string]bool{}
	for g, x := range exchangeFold(evs, ids, win, WhileRunning) {
		if latestRemandOwed(x.rounds, remandSittingsOf(remands[g])) {
			out[g] = true
		}
	}
	return out
}

// remandedClosedGapSQL is the gap a docket motion is about when that gap is CLOSED, with the act
// that closed it last — the lens's close or the bench's closing ruling, whichever is later — and
// that act's seat; no row when the gap is open.
const remandedClosedGapSQL = `SELECT md."gap_id", e."id", e."seat_id"
  FROM "motion" mo
  JOIN "motion_docket" md ON md."event_id" = mo."event_id"
  JOIN "gap" g ON g."gap_id" = md."gap_id"
  JOIN "events" e ON e."id" = max(COALESCE(g."close_event", 0), COALESCE(g."bench_close_event", 0))
  WHERE mo."motion_id" = ? AND NOT g."open"`

// requireRemandOfOpenGap refuses a remand of a docket motion whose gap is closed (gblock,
// 2026-10-24). A gap may close while a docket motion on it stands — its lens closed it, and the bench
// rules the motion on the record as it stands (closingowed.go) — but a remand sends the gap back for
// one more exchange between its minting lens and blue, and a closed gap has none to have: the
// obligation would stand with no seat able to discharge it, as the remand of a run's closed gap did
// on universe-m33, stored and never acted on while the run passed. The tool does not reopen the gap
// either; the refusal names the close, and the bench rules with a disposition that closes.
func requireRemandOfOpenGap(q recordsql.Querier, motionID string) error {
	if motionID == "" {
		return nil
	}
	var gapID, seat string
	var closed int64
	found, err := queryRowAt(q, []any{&gapID, &closed, &seat}, remandedClosedGapSQL, motionID)
	if err != nil || !found {
		return err
	}
	return feov.Errorf(feov.Conflict,
		"motion %s is about gap %s, which is CLOSED — %s closed it at event %d. A remand sends a gap back for one more exchange between its minting lens and blue, and a closed gap has none to have, so nobody could discharge it; the record does not reopen a gap. "+
			"Nothing was recorded; rule the motion with a disposition that closes, on the record as it stands",
		motionID, gapID, seat, closed)
}

// remandStage is where a gap stands against the bench's counted remands.
type remandStage int

const (
	notRemanded   remandStage = iota
	remandOwed                // remanded once at impasse, and no exchange has begun since the ruling
	remandSpent               // remanded once at impasse, and its exchange has begun since
	remandAtLimit             // remanded at impasse twice: never sent to the bench again
)

// remandStageWords is the stage as the work list's `remand` field states it; notRemanded has none.
var remandStageWords = map[remandStage]string{remandOwed: "owed", remandSpent: "spent", remandAtLimit: "at_limit"}

// remandStageOf reads the stage off the fold's counted remands and rounds.
func remandStageOf(x *GapExchanges) remandStage {
	switch n := len(x.remandSittings); {
	case n == 0:
		return notRemanded
	case n > 1:
		return remandAtLimit
	}
	if latestRemandOwed(x.rounds, x.remandSittings) {
		return remandOwed
	}
	return remandSpent
}

// gapRoute is what the dispatch does with one open gap — THE ONE PREDICATE the plan's readiness, the
// remand items on blue's and the minting lens's work lists, and the chair's row all read, so no
// list offers an exchange the dispatch does not ready, nor withholds one it does.
type gapRoute int

const (
	routeNobody        gapRoute = iota // not material and not stranded, nothing docketed, no remand's exchange owed: readies nobody
	routeBench                         // a docket motion on it stands unruled: the bench's
	routeGradeMovement                 // its grade movement, with the open gaps', is past the threshold: docketed for the bench
	routeDebate                        // below its limits: its minting lens and blue
	routeDocket                        // at impasse, never remanded there: docketed for the bench
	routeRemandOwed                    // at impasse and remanded, or not material and remanded: its minting lens and blue, once, on the direction
	routeRedocket                      // at impasse after its remand's exchange: docketed for the bench again
	routeAtLimit                       // at impasse and remanded twice: open at its limit, readies nobody
)

// routeOf is the route for an open gap: material is the class's answer, stranded whether a successor
// names it, unruledDocket whether a docket motion on it stands unruled, moved whether its grade
// movement is docketed (gradeMovement.docketed) — whatever its materiality, because a gap regraded
// below medium is the movement the bench rules on — and x its exchanges.
//
// THE BENCH'S REMAND OUTRANKS MATERIALITY (gblock, 2026-10-24). A gap that is not material readies
// nobody — until the bench remands it: then its minting lens and blue are readied for the one more
// exchange the remand grants, as for any remand, and once that exchange has begun it readies nobody
// again. Whether the bench ruled it at impasse does not matter here: a gap that is not material is
// never docketed for impasse, so its remand came from a party's motion or its grade movement.
func routeOf(material, stranded, unruledDocket, moved bool, x *GapExchanges) gapRoute {
	switch {
	case unruledDocket:
		return routeBench
	case moved:
		return routeGradeMovement
	case !material && !stranded:
		if x.latestRemandOwed {
			return routeRemandOwed
		}
		return routeNobody
	case !x.Impasse:
		return routeDebate
	}
	switch x.Remand {
	case remandOwed:
		return routeRemandOwed
	case remandSpent:
		return routeRedocket
	case remandAtLimit:
		return routeAtLimit
	}
	return routeDocket
}
