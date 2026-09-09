package record

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Party is one seat the chair engages and the gaps it is engaged on. Empty GapIDs is rule 1's
// dispatch: the head moved past the lens's pin, so it sits to audit the current report.
type Party struct {
	SeatID string   `json:"seat_id"`
	GapIDs []string `json:"gap_ids"`
}

// Plan is what `feov-record dispatch next` computes FROM THE BOARD (plans/roundless.md §III.B.1).
// The chair relays it; the workflow dispatches what it says; nothing in it is a seat's assertion.
type Plan struct {
	Head    int64   `json:"head"`    // events.id of the latest blue_edit or base_ingest — the report the parties audit
	Parties []Party `json:"parties"` // who to engage; empty is the termination signal
	// Docket names the gaps that reached impasse with no docket motion yet: the verb files one for
	// each under the chair's authorship the moment impasse is first computed, so "docketed" and
	// "at impasse" are one fact and no seat's discretion sits between a stalled gap and the bench.
	Docket        []string `json:"docket"`
	PassPermitted bool     `json:"pass_permitted"` // no material gap open, every cast lens sat against the head, every docket ruled
	Ceiling       bool     `json:"ceiling"`        // every open material gap is at impasse and has had its bench ruling (carried)
	Why           []string `json:"why"`            // the readiness of each source, in words a reader can check against the board
}

// material is the severity mass at or above which a gap holds the gate and readies its parties:
// GRADE_MEDIUM and up. Below it a gap is on the board, blue may answer it when engaged for
// something else, but a trifle alone cannot spin the cycle (gblock, 2026-09-08).
const material = 2.0

type openGap struct {
	id, mintedBy, severity string
	dockets, rulings       int
	// supersededBy is the successor that names this gap as an ancestor, when one does and this gap
	// is still open — the gap view's `stranded`. The PASS gate refuses a verdict over one.
	supersededBy string
}

// PlanDispatch computes readiness from three sources — the report head against each lens's pin,
// each open material gap below its limits, and each docketed gap awaiting the bench — and the two
// derived facts the termination turns on. It writes nothing.
func PlanDispatch(run Run) (Plan, error) {
	var plan Plan
	cast, err := CastOf(run)
	if err != nil {
		return plan, err
	}
	if cast == nil {
		return plan, fmt.Errorf("record: dispatch refused — the record holds no cast. setup writes the run's admissible seats before any seat registers; a run without one has nothing to dispatch against")
	}
	params, err := RunParams(run)
	if err != nil {
		return plan, err
	}
	db, err := openRunForRead(run)
	if err != nil {
		return plan, err
	}
	if db == nil {
		return plan, fmt.Errorf("record: dispatch refused — the run has no record")
	}
	if _, err := queryRow(run, []any{&plan.Head},
		`SELECT COALESCE(MAX("id"), 0) FROM "events" WHERE "type" IN ('blue_edit', 'base_ingest')`); err != nil {
		return plan, err
	}

	// Source 1: a lens is ready when the head is past its last pin. The pin is the dispatch the
	// lens SAT for — its register follows the dispatch — not the dispatch row: a lens engaged
	// that never registered has not sat, and stays ready.
	evs, _, err := recordsql.EventsW(db)
	if err != nil {
		return plan, err
	}
	ids, err := eventIDs(db)
	if err != nil {
		return plan, err
	}
	pins, satIDs := lensPins(evs, ids)
	var lenses []string
	for _, s := range cast {
		if strings.HasPrefix(s, "red-lens-") {
			lenses = append(lenses, s)
		}
	}
	parties := map[string][]string{}
	order := []string{}
	engage := func(seat string, gaps ...string) {
		if _, seen := parties[seat]; !seen {
			order = append(order, seat)
			parties[seat] = []string{}
		}
		parties[seat] = append(parties[seat], gaps...)
	}
	allLensesSat := plan.Head > 0
	for _, l := range lenses {
		if pins[l] < plan.Head {
			engage(l)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: head %d is past its pin %d", l, plan.Head, pins[l]))
			allLensesSat = false
		}
	}

	// Source 2 and 3: the open gaps, with their materiality, their exchanges and their docket.
	gaps, err := openGaps(db)
	if err != nil {
		return plan, err
	}
	exch := exchangesOf(evs, ids, params)
	materialOpen, materialSettled := 0, 0
	unruledDocket := false
	for _, g := range gaps {
		// A STRANDED GAP IS READY WORK WHATEVER ITS GRADE. Superseding is a promise to replace, and
		// the PASS gate refuses a verdict while the ancestor is open (refs.go) — so a sub-material
		// ancestor nobody is dispatched to close would leave the plan saying "pass permitted" and
		// the gate saying no, forever, and the run ends UNVERIFIED with nobody ready. Found by the
		// release sweep. Held as material here: its minter and blue are engaged, its exchanges
		// count, and at impasse it reaches the bench like any other.
		stranded := g.supersededBy != ""
		trifle := MASS[g.severity] < material && !stranded
		// A DOCKET MOTION IS THE ESCALATION ROUTE, and it readies the bench whether the gap is at
		// impasse or not. The dispatch files one at impasse; a party may file one earlier (`motion
		// docket file` is a red and blue verb, kept so the route to the bench is not the record's
		// discretion alone), and a trifle may be escalated too. Either way the motion stands until
		// the bench rules it, PASS is refused while it stands, and a bench that sat for it and
		// ruled nothing is not re-readied — the run cannot end in a verdict while it stands.
		if g.rulings < g.dockets {
			unruledDocket = true
			if !trifle {
				materialOpen++
			}
			if benchSatFor(evs, ids, g.id, satIDs) {
				plan.Why = append(plan.Why, fmt.Sprintf("%s: docketed, the bench sat and ruled nothing — one bench sitting per docketing, so this gap is not re-readied; the run cannot end in a verdict while it stands", g.id))
				continue
			}
			engage("judge", g.id)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: docketed and unruled — the bench is ready", g.id))
			continue
		}
		if trifle {
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open but below material (%s) — readies nobody", g.id, g.severity))
			continue
		}
		if stranded {
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open and superseded by %s — held as material until its minter closes it", g.id, g.supersededBy))
		}
		materialOpen++
		x := exch[g.id]
		if x == nil {
			x = &GapExchanges{GapID: g.id}
		}
		if !x.Impasse {
			if g.mintedBy != "" {
				engage(g.mintedBy, g.id)
			}
			engage("blue-respond", g.id)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open, material, %d exchange(s) (%d stalled) — below its limits", g.id, x.Exchanges, x.Stalled))
			continue
		}
		if g.dockets == 0 {
			plan.Docket = append(plan.Docket, g.id)
			engage("judge", g.id)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse (%d exchange(s), %d stalled) — docketed for the bench", g.id, x.Exchanges, x.Stalled))
			continue
		}
		materialSettled++ // ruled and still open: carried
		plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse, ruled carried — at its limit", g.id))
	}
	for _, s := range order {
		plan.Parties = append(plan.Parties, Party{SeatID: s, GapIDs: parties[s]})
	}
	plan.PassPermitted = materialOpen == 0 && allLensesSat && !unruledDocket && len(plan.Parties) == 0
	plan.Ceiling = len(plan.Parties) == 0 && materialOpen > 0 && materialSettled == materialOpen
	return plan, nil
}

// lensPins is, per seat, the pin of the last dispatch it SAT for (registered after), and the ids
// of the registers that followed a dispatch naming the seat (its sittings for dispatches).
func lensPins(evs []*Event, ids []int64) (map[string]int64, map[string][]int64) {
	registers := map[string][]int64{}
	type d struct {
		id, pin int64
		seat    string
		gaps    []string
	}
	var ds []d
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Register:
			registers[e.GetSeatId()] = append(registers[e.GetSeatId()], ids[i])
		case *recordpb.Dispatch:
			ds = append(ds, d{id: ids[i], pin: b.GetPin(), seat: b.GetSeatId(), gaps: b.GetGapIds()})
		}
	}
	pins := map[string]int64{}
	sat := map[string][]int64{}
	for _, x := range ds {
		rs := registers[x.seat]
		k := sort.Search(len(rs), func(i int) bool { return rs[i] > x.id })
		if k == len(rs) {
			continue
		}
		pins[x.seat] = x.pin
		sat[x.seat] = append(sat[x.seat], rs[k])
	}
	return pins, sat
}

// benchSatFor reports whether the bench registered after the latest dispatch engaging it on gapID
// — it had its sitting for this docketing.
func benchSatFor(evs []*Event, ids []int64, gapID string, sat map[string][]int64) bool {
	var last int64
	for i, e := range evs {
		if b, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok && b.GetSeatId() == "judge" {
			for _, g := range b.GetGapIds() {
				if g == gapID {
					last = ids[i]
				}
			}
		}
	}
	if last == 0 {
		return false
	}
	for _, r := range sat["judge"] {
		if r > last {
			return true
		}
	}
	return false
}

// openGaps reads the open gaps with what the plan needs of each: who minted it, its current
// severity, and whether it has been docketed and ruled. All off the gap view and the motion
// tables — the same fold every reader uses.
func openGaps(db *sql.DB) ([]openGap, error) {
	rows, err := db.Query(`SELECT g."gap_id", COALESCE(g."minted_by", ''), COALESCE(g."current_severity", ''),
	    CASE WHEN g."stranded" THEN COALESCE(g."superseded_by", '') ELSE '' END,
	    (SELECT count(*) FROM "motion_docket" md WHERE md."gap_id" = g."gap_id"),
	    (SELECT count(*) FROM "motion_rule" mr JOIN "motion" m ON m."motion_id" = mr."motion_id"
	       JOIN "motion_docket" md ON md."event_id" = m."event_id" WHERE md."gap_id" = g."gap_id")
	  FROM "gap" g WHERE g."open" ORDER BY g."minted_event"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its open gaps: %w", err)
	}
	defer rows.Close()
	var out []openGap
	for rows.Next() {
		var g openGap
		if err := rows.Scan(&g.id, &g.mintedBy, &g.severity, &g.supersededBy, &g.dockets, &g.rulings); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// requireEveryCastLensSatAgainstHead is the direct replacement for the round loop's "every lens
// sits every round" (plans/roundless.md §III.B.1): PASS is refused until every cast lens has SAT
// against the current report head — its register FOLLOWING the dispatch pinned at the head, not
// the dispatch row. A record with no cast has no lenses to wait for; that is the fixtures' world
// and a migrated archive's, where the property held by construction.
func requireEveryCastLensSatAgainstHead(run Run) error {
	cast, err := CastOf(run)
	if err != nil || cast == nil {
		return err
	}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return err
	}
	var head int64
	if _, err := queryRow(run, []any{&head},
		`SELECT COALESCE(MAX("id"), 0) FROM "events" WHERE "type" IN ('blue_edit', 'base_ingest')`); err != nil {
		return err
	}
	evs, _, err := recordsql.EventsW(db)
	if err != nil {
		return err
	}
	ids, err := eventIDs(db)
	if err != nil {
		return err
	}
	pins, _ := lensPins(evs, ids)
	var behind []string
	for _, s := range cast {
		if strings.HasPrefix(s, "red-lens-") && pins[s] < head {
			behind = append(behind, fmt.Sprintf("%s (pin %d)", s, pins[s]))
		}
	}
	if len(behind) == 0 && head > 0 {
		return nil
	}
	if head == 0 {
		return fmt.Errorf("record: verdict PASS refused — no report has been ingested or edited, so there is nothing a lens could have audited")
	}
	return fmt.Errorf("record: verdict PASS refused — the report head is %d and %d cast lens(es) have not sat against it: %s. Every lens sits against the text it passes; `dispatch next` readies them", head, len(behind), strings.Join(behind, ", "))
}
