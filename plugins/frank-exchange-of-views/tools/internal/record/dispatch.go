package record

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// Party is one seat the chair engages and the gaps it is engaged on. A lens engaged with no gaps is
// ready by its retirement state, and sits to audit the current report.
type Party struct {
	SeatID string   `json:"seat_id"`
	GapIDs []string `json:"gap_ids"`
}

// Plan is what `feov-record dispatch next` computes FROM THE BOARD (plans/roundless.md §III.B.1).
// The chair relays it; the workflow dispatches what it says; nothing in it is a seat's assertion.
type Plan struct {
	Head    int64   `json:"head"`    // events.id of the latest blue_edit or base_ingest — the report the parties audit
	Parties []Party `json:"parties"` // who to engage; empty is the termination signal, and empty is [], never null
	// Docket names the gaps that reached impasse with no docket motion yet: the verb files one for
	// each under the chair's authorship the moment impasse is first computed, so "docketed" and
	// "at impasse" are one fact and no seat's discretion sits between a stalled gap and the bench.
	Docket        []string `json:"docket"`
	PassPermitted bool     `json:"pass_permitted"` // a report ingested, no material gap open (IsMaterial: by its class, else graded medium or above), no cast lens ready, every docket ruled, nobody dispatched
	// StaleAreas is every lens retired for good whose pin the head has moved past. The chair reads
	// the changes since each pin against that area's duties and names the areas in its spot-check;
	// a PASS is refused until it has.
	StaleAreas []StaleArea `json:"stale_areas"`
	// Ceiling is the run at its limit, for one of two reasons EpochLimitReached tells apart: every
	// open material gap is at impasse and has had its bench ruling (remanded), or the run's epoch
	// limit is reached with parties still ready.
	Ceiling bool `json:"ceiling"`
	// MaxEpochs is the run's epoch limit (Params.MaxEpochs); 0 when the run is held to none.
	MaxEpochs int `json:"max_epochs"`
	// EpochLimitReached: this chair sitting opens the run's last epoch with parties the board
	// readies, and they are not dispatched — the run ends CEILING. A last epoch with nobody ready
	// is not at the limit: it ends as any empty plan does.
	EpochLimitReached bool     `json:"epoch_limit_reached"`
	Why               []string `json:"why"` // the readiness of each source, in words a reader can check against the board
}

// material is the severity mass at or above which a `by_grade` gap is material: GRADE_MEDIUM and
// up. The class decides for `always` and `never`. A gap that is not material stays on the board,
// and blue may answer it when engaged for something else, but it alone cannot spin the cycle
// (gblock, 2026-09-08). Read only by the two carriers of the definition: IsMaterial, which the
// Go family fold calls, and the gap view's "material" column, which spells the same floor in SQL.
const material = 2.0

// IsMaterial is THE definition of a material gap: its class is `always`, or its class goes
// `by_grade` and its current severity is medium or above. A `never` class is not material at any
// grade. Every reader of materiality reads this, or the gap view's column that states it in SQL.
func IsMaterial(cm recordpb.ClassMaterial, currentSeverity recordpb.Grade) bool {
	switch cm {
	case recordpb.ClassMaterial_CLASS_MATERIAL_ALWAYS:
		return true
	case recordpb.ClassMaterial_CLASS_MATERIAL_BY_GRADE:
		return recordpb.GradeMass(currentSeverity) >= material
	}
	return false
}

type openGap struct {
	id, mintedBy, severity string
	// material is the gap view's column — the one definition — and classMaterial the class's
	// default, which the reason names when the class alone makes the gap not material.
	material         bool
	classMaterial    string
	dockets, rulings int
	// supersededBy is the successor that names this gap as an ancestor, when one does and this gap
	// is still open — the gap view's `stranded`. The PASS gate refuses a verdict over one.
	supersededBy string
}

// PlanDispatch computes readiness from three sources — each cast lens's retirement state, each
// open material gap below its limits, and each docketed gap awaiting the bench — and the two
// derived facts the termination turns on. It writes nothing.
func PlanDispatch(run Run) (Plan, error) {
	// ARRAYS, NEVER null, AT THE SOURCE. The chair relays this JSON and the engine type-checks every
	// field it reads; a nil slice marshals as null, which the relay refuses as a missing array. B9's
	// chair relayed a PASS-permitted plan verbatim, "parties": null, and the engine aborted the run
	// at the sitting that should have ended it.
	plan := Plan{Parties: []Party{}, Docket: []string{}, Why: []string{}, StaleAreas: []StaleArea{}}
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

	// Source 1: each cast lens by its retirement state (retirement.go) — the same fold the PASS
	// gate refuses from. A lens engaged that never registered has not sat, so its state has not
	// moved and it stays ready.
	evs, _, err := recordsql.EventsW(db)
	if err != nil {
		return plan, err
	}
	ids, err := eventIDs(db)
	if err != nil {
		return plan, err
	}
	_, satIDs := lensPins(evs, ids)
	fresh, err := freshMaterialOf(db)
	if err != nil {
		return plan, err
	}
	folds := lensStates(evs, ids, plan.Head, fresh)
	plan.StaleAreas = staleAreasOf(folds, plan.Head)
	parties := map[string][]string{}
	order := []string{}
	engage := func(seat string, gaps ...string) {
		if _, seen := parties[seat]; !seen {
			order = append(order, seat)
			parties[seat] = []string{}
		}
		parties[seat] = append(parties[seat], gaps...)
	}
	noLensReady := true
	for _, f := range folds {
		if f.ready {
			engage(f.seat)
			noLensReady = false
		}
		plan.Why = append(plan.Why, f.why)
	}
	_ = cast // the cast is the fold's roster (castOfEvents): the same Cast event CastOf reads

	// Source 2 and 3: the open gaps, with their materiality, their exchanges and their docket.
	gaps, err := openGaps(db)
	if err != nil {
		return plan, err
	}
	exch := exchangesOf(evs, ids, params, WhileRunning)
	materialOpen, materialSettled := 0, 0
	unruledDocket := false
	for _, g := range gaps {
		// A STRANDED GAP IS READY WORK WHATEVER ITS CLASS OR GRADE. Superseding is a promise to
		// replace, and the PASS gate refuses a verdict while the ancestor is open (refs.go) — so a
		// non-material ancestor nobody is dispatched to close would leave the plan saying "pass
		// permitted" and the gate saying no, forever, and the run ends UNVERIFIED with nobody ready.
		// Found by the release sweep. Held as material here: its minter and blue are engaged, its
		// exchanges count, and at impasse it reaches the bench like any other.
		stranded := g.supersededBy != ""
		trifle := !g.material && !stranded
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
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open and not material (%s) — readies nobody", g.id, notMaterialBecause(g.classMaterial, g.severity)))
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
			plan.Why = append(plan.Why, fmt.Sprintf("%s: open, material, %s — below its limits", g.id, x.Counted()))
			continue
		}
		if g.dockets == 0 {
			plan.Docket = append(plan.Docket, g.id)
			engage("judge", g.id)
			plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse (%s) — docketed for the bench", g.id, x.Counted()))
			continue
		}
		materialSettled++ // ruled and still open: remanded
		plan.Why = append(plan.Why, fmt.Sprintf("%s: at impasse, ruled remanded — at its limit", g.id))
	}
	for _, s := range order {
		plan.Parties = append(plan.Parties, Party{SeatID: s, GapIDs: parties[s]})
	}
	plan.PassPermitted = plan.Head > 0 && materialOpen == 0 && noLensReady && !unruledDocket && len(plan.Parties) == 0
	plan.Ceiling = len(plan.Parties) == 0 && materialOpen > 0 && materialSettled == materialOpen

	// THE EPOCH LIMIT IS A TERM OF THE RUN, read like k and kMax. The chair sitting that opens the
	// last epoch dispatches nobody: the parties the board readies stay undispatched and the run
	// ends CEILING. A board that permits PASS there readied nobody, so it is not at the limit and
	// the chair records the PASS as ever. The docket stands — filing a motion dispatches no seat,
	// and the terminal bench rules what stays unruled at the exit.
	plan.MaxEpochs = params.MaxEpochs
	if params.MaxEpochs > 0 && len(plan.Parties) > 0 && epochOf(evs) >= params.MaxEpochs {
		plan.Why = append(plan.Why, fmt.Sprintf("epoch limit %d reached — this chair sitting opens the run's last epoch, so the %d party(ies) above are not dispatched", params.MaxEpochs, len(plan.Parties)))
		plan.Parties = []Party{}
		plan.EpochLimitReached, plan.Ceiling = true, true
	}
	return plan, nil
}

// epochOf is the epoch the record's clock has reached — one per chair register, the clock every
// reader of "epoch" uses.
func epochOf(evs []*Event) int {
	var clk Clock
	epoch := 0
	for _, e := range evs {
		epoch = clk.Advance(e).Epoch
	}
	return epoch
}

// dispatchRow is one dispatch event: where it sits in the stream, the head it pinned, the seat it
// named and the gaps it engaged that seat on.
type dispatchRow struct {
	at, pin int64
	seat    string
	gaps    []string
}

// dispatchLedger reads the stream once for what "did this seat sit" is decided from: every
// dispatch row, and each seat's registers, in stream order. seq gives each event's place — the
// events."id" where the reader has them, the position where it holds only the stream. The
// predicate compares order and nothing else, so either answers it the same.
//
// THE REGISTERS ARE THE ONES THAT OPEN A SITTING FOR A DISPATCH, which is opensASitting's question
// and is asked there: a register naming the sitting it repairs opens none of those — the seat is
// handed a prompt, so it IS sitting and Clock counts the turn, but it sits for no dispatch and ends
// no sitting; sittingCloser adds its acts to the sitting it repairs, and ActClock gives them that
// sitting's number.
func dispatchLedger(evs []*Event, seq []int64) ([]dispatchRow, map[string][]int64) {
	var ds []dispatchRow
	registers := map[string][]int64{}
	for i, e := range evs {
		// A SITTING OPENS EITHER WAY, and the seat is whichever one this event opens a sitting for
		// — the registering seat, or the seat a hook's sitting_open resolves to through its
		// configuration. The second arm is how a no-op sitting costs nothing: every reader of "has
		// this seat sat" comes through this map, so a hook-opened sitting satisfies the dispatch,
		// the pin, the retirement fold and the work list without the seat running a command.
		if seat, opens := recordpb.SeatOpeningSitting(e); opens {
			registers[seat] = append(registers[seat], seq[i])
		}
		if b, ok := recordpb.BodyAs[*recordpb.Dispatch](e); ok {
			ds = append(ds, dispatchRow{at: seq[i], pin: b.GetPin(), seat: b.GetSeatId(), gaps: b.GetGapIds()})
		}
	}
	return ds, registers
}

// firstAfter is the first of an ascending sequence that follows at.
func firstAfter(xs []int64, at int64) (int64, bool) {
	k := sort.Search(len(xs), func(i int) bool { return xs[i] > at })
	if k == len(xs) {
		return 0, false
	}
	return xs[k], true
}

// sittingFor IS THE ONE ANSWER TO "HAS THIS SEAT SAT FOR WHAT IT WAS DISPATCHED FOR": its sitting
// for a dispatch is its first register after the dispatch row, and a seat that has not registered
// since has not sat. Every reader of the question asks it here — the lens's pin (lensPins), the
// lens's retirement fold (lensStates) and its last sitting (lastSittingBefore), the bench's sitting
// for a docketing (benchSatFor, off lensPins' sittings), the exchange count (exchangesOf), and the
// seat's own work list (owedSitting). An unsat dispatch is why dispatch
// readies the seat again, and it is why the seat's work list is not complete.
func sittingFor(registers []int64, d dispatchRow) (int64, bool) {
	return firstAfter(registers, d.at)
}

// sittingCloser IS THE ONE ANSWER TO "WHERE DID THIS SEAT'S SITTING END", as sittingFor is to where
// it began. Every reader that bounds a seat's sitting asks it here — the exchange count
// (exchangesOf) and blue's sittings (BlueSittings, behind revisionOwed, the manifest's owed set and
// capture's record-parity audit) — so a change to what closes a sitting reaches all of them at once.
//
// ONLY FACTS ABOUT THIS SEAT CLOSE IT, AND THE EARLIER OF TWO DOES. A sitting ends at whichever
// comes first after it began:
//
//   - the seat's next register. A register is a seat's first act of a sitting, so the next one is
//     proof the sitting before it ended — written by that seat.
//   - the stop of the agent that sat it: a sitting_close whose agent_id is the one on THIS
//     sitting's register. The SubagentStop hook writes it when that agent returns, so it is the
//     harness's observation of this seat, not an act of another seat.
//
// A REPAIR OF THE SITTING IS PART OF IT. A register naming the sitting it repairs (repairs_sitting)
// opens no span of its own: from it to where that register's own sitting would end — the seat's
// next opening register, or the stop of the agent on the repair — its acts are the repaired
// sitting's. Under the shipped Workflow engine the first agent's stop has already closed the
// sitting when the re-prompt registers, so the repair is a second span, and the sitting ends where
// its last span does.
//
// NOTHING ANOTHER SEAT WRITES BOUNDS IT. A read keyed on another seat's act — the chair's next
// register, the chair's next dispatch row — depends on a rule enforced at that seat's write path,
// and a warm chair registers once per run: on five archived runs a chair-keyed exchange fold read
// every party sitting as still open and no gap could reach the bench by impasse (#1002).
//
// A STOP THAT JOINS NO REGISTER CLOSES NOTHING. The hook records every typed subagent in a project
// whose run marker is live, a developer's own subagents included, and a register with no agent_id
// (a run the PreToolUse hook never reached) has nothing a stop can join. A headless seat run as a
// `claude -p` main session fires no SubagentStop at all, so its sittings close by register alone.
//
// WITH NEITHER PAST IT, WHAT THE SITTING IS DEPENDS ON WHEN THE RECORD IS READ (ReadWhen), and the
// reader says which — the closer never guesses it:
//
//   - WhileRunning: the sitting is UNRESOLVED, not complete and not absent. It may be in flight, or
//     it may have ended with nothing recorded after it, and the record holds no fact that tells the
//     two apart. Each reader states that case as its own answer.
//   - AfterTheRun: the end of the record closes it. A finished run has no sitting in flight, so the
//     last act on the record is the last act of every sitting still open — that is the premise of
//     the read, not an act of another seat.
type sittingCloser struct {
	registers map[string][]int64 // seat -> its opening registers' places, ascending
	agentOf   map[int64]string   // a register's place -> the agent_id it carries
	stops     map[string][]int64 // agent_id -> the places of its sitting_close events, ascending
	repairs   map[int64][]int64  // an opening register's place -> the places of the registers repairing its sitting, ascending
	recordEnd int64              // one past the last place on the record
	when      ReadWhen
}

// ReadWhen is when a reader reads the record: while the run can still be sitting, or after it
// ended. It decides one thing — whether a sitting nothing about its seat has closed is unresolved
// or closed by the end of the record (sittingCloser) — and every reader that bounds a sitting
// takes it from where it runs, never from a default.
type ReadWhen int

const (
	// WhileRunning is a read made while a sitting may be in flight: a seat's work list, the
	// chair's dispatch plan, a live view.
	WhileRunning ReadWhen = iota
	// AfterTheRun is a read of a finished run, which is capture's alone: capture runs once the run
	// has ended, so no sitting is still in flight.
	AfterTheRun
)

// sittingCloserOf reads the stream once for every fact that can close a sitting. seq and registers
// are dispatchLedger's: the same places the sitting's start was read at.
func sittingCloserOf(evs []*Event, seq []int64, registers map[string][]int64, when ReadWhen) sittingCloser {
	c := sittingCloser{registers: registers, agentOf: map[int64]string{}, stops: map[string][]int64{}, repairs: map[int64][]int64{}, when: when}
	if n := len(seq); n > 0 {
		c.recordEnd = seq[n-1] + 1
	}
	type opening struct {
		seat  string
		place int64
	}
	openedBy := map[string]opening{} // an opening register's key -> its seat and place
	for i, e := range evs {
		switch b := mustBody(e).(type) {
		case *recordpb.Register:
			if a := b.GetAgentId(); a != "" {
				c.agentOf[seq[i]] = a
			}
		case *recordpb.SittingClose:
			if a := b.GetAgentId(); a != "" {
				c.stops[a] = append(c.stops[a], seq[i])
			}
		}
		// WHICH REGISTERS OPEN A SPAN IS opensASitting's, the same predicate dispatchLedger read
		// for where a sitting begins — so the spans this bounds and the sittings that exist cannot
		// come from two different answers. An agent id is read above and a span opened here
		// because they are different facts: a body nothing can read carries no agent, and still
		// opens a sitting.
		switch repaired, repairs := recordpb.SittingRepairedBy(e); {
		case recordpb.OpensASitting(e):
			openedBy[e.GetKey()] = opening{seat: e.GetSeatId(), place: seq[i]}
		case repairs:
			if o, ok := openedBy[repaired]; ok && o.seat == e.GetSeatId() {
				c.repairs[o.place] = append(c.repairs[o.place], seq[i])
			}
		}
	}
	return c
}

// span is one stretch of a sitting's acts: the places from its register up to, not including, the
// act that closed it.
type span struct{ from, to int64 }

// holds reports whether a place falls in any of the spans.
func holds(spans []span, at int64) bool {
	for _, s := range spans {
		if at >= s.from && at < s.to {
			return true
		}
	}
	return false
}

// bounds is the seat's sitting that began at start: the spans its acts fall in — its own, then one
// per register repairing it — where the last of them ends, and whether every one is closed.
func (c sittingCloser) bounds(seat string, start int64) (spans []span, end int64, closed bool) {
	closed = true
	for _, from := range append([]int64{start}, c.repairs[start]...) {
		to, ok := c.end(seat, from)
		spans = append(spans, span{from: from, to: to})
		end, closed = max(end, to), closed && ok
	}
	return spans, end, closed
}

// end is where the stretch a register at from opened ended — the place of the act that closed it,
// which is outside it. With nothing about this seat past it, end is the end of the record, and
// closed is false while the run is read as running and true after it.
func (c sittingCloser) end(seat string, from int64) (int64, bool) {
	end, closed := firstAfter(c.registers[seat], from)
	if agent := c.agentOf[from]; agent != "" {
		if stop, stopped := firstAfter(c.stops[agent], from); stopped && (!closed || stop < end) {
			end, closed = stop, true
		}
	}
	if !closed {
		return c.recordEnd, c.when == AfterTheRun
	}
	return end, true
}

// DispatchGroup is one chair sitting's dispatch as the record delimits it: the dispatch rows
// written with NO REGISTER BETWEEN THEM, by any seat. Somebody registering is somebody sitting —
// the chair opening a sitting, or a party sitting for what the chair dispatched — and the workflow
// sequences the two, so a dispatch row after a register belongs to a later chair sitting, and one
// before any register belongs to the same.
//
// NOT THE CLOCK'S CHAIR-SITTING COUNT, and the difference is measured. The clock counts the
// chair's registers, and a warm chair — a later sitting resuming the same session — registered
// ONCE per run on all four archived is-91-prime B runs (B3–B6), so every chair sitting read as
// sitting 1 and every dispatch of the run as one. Nor one row set per `dispatch next`: the B5 and
// B6 chairs each ran the verb twice in their first sitting, once for the prose and once for
// --json, and wrote the plan twice, 4 s apart, with nobody sitting in between.
type DispatchGroup struct {
	First, Last int      // stream positions of the group's first and last dispatch row
	Parties     []string // each seat the group names, in the order first named
	// Sat is each party's sitting for the group, by sittingFor off the last row naming it: the
	// stream position of its first register after that row. A party absent here has not sat.
	Sat map[string]int
	// PartyRows is each party's LAST row in the group — the row Sat reads, and the one a relayed
	// plan is compared against: a docket plan is never standing, so a chair that writes the plan
	// twice leaves two rows for one party, and the later is the dispatch.
	PartyRows map[string]PartyRow
	rows      []dispatchRow
}

// PartyRow is one party's dispatch row as recorded: the head it was pinned to and the gaps it
// engaged the party on.
type PartyRow struct {
	Pin    int64
	GapIDs []string
}

// DispatchGroups is the record's dispatches, grouped as DispatchGroup says, in stream order.
func DispatchGroups(evs []*Event) []DispatchGroup {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, registers := dispatchLedger(evs, seq)
	var anyone []int64
	for _, rs := range registers {
		anyone = append(anyone, rs...)
	}
	sort.Slice(anyone, func(a, b int) bool { return anyone[a] < anyone[b] })
	var groups []DispatchGroup
	for i, d := range ds {
		if i == 0 || registeredBetween(anyone, ds[i-1].at, d.at) {
			groups = append(groups, DispatchGroup{First: int(d.at), Sat: map[string]int{}, PartyRows: map[string]PartyRow{}})
		}
		g := &groups[len(groups)-1]
		g.Last = int(d.at)
		g.rows = append(g.rows, d)
	}
	for k := range groups {
		g := &groups[k]
		last := map[string]dispatchRow{}
		for _, d := range g.rows {
			if _, seen := last[d.seat]; !seen {
				g.Parties = append(g.Parties, d.seat)
			}
			last[d.seat] = d
		}
		for p, d := range last {
			g.PartyRows[p] = PartyRow{Pin: d.pin, GapIDs: append([]string{}, d.gaps...)}
			if r, ok := sittingFor(registers[p], d); ok {
				g.Sat[p] = int(r)
			}
		}
	}
	return groups
}

// registeredBetween reports whether any of the ascending registers falls strictly between a and b.
func registeredBetween(registers []int64, a, b int64) bool {
	r, ok := firstAfter(registers, a)
	return ok && r < b
}

// chairSeat is the seat whose registers the clock counts as epochs.
const chairSeat = "red-chair"

// blueRespondSeat is blue's responding seat — the one the chair dispatches onto gaps.
const blueRespondSeat = "blue-respond"

// unopenedChairSitting is the chair's latest dispatch when a party has SAT for it (sittingFor)
// and the chair has not registered since recording it. The workflow comes back to the chair only
// after the parties sit, so that is a new chair sitting no register opened: the clock would read
// it as the previous epoch, and the dispatch groups would have only the parties' registers to
// split on. Every seat's sitting opens with a register; this is the chair's owed one.
//
// THE EXCHANGE COUNT NO LONGER DEPENDS ON IT. exchangesOf closed a party's sitting at the chair's
// next register, so a chair that skipped this register silently zeroed the fold; it now closes a
// sitting at the sitting seat's OWN next register or its own agent's stop, whichever is first
// (#1002). This item stands on its own ground —
// the clock and the groups — and the fold stands on the parties'.
func unopenedChairSitting(evs []*Event) (DispatchGroup, bool) {
	groups := DispatchGroups(evs)
	if len(groups) == 0 {
		return DispatchGroup{}, false
	}
	g := groups[len(groups)-1]
	if len(g.Sat) == 0 {
		return DispatchGroup{}, false // nobody has sat for it: still the sitting that recorded it
	}
	for i := len(evs) - 1; i > g.Last; i-- {
		if evs[i].GetType() == recordpb.EventType_EVENT_TYPE_REGISTER && evs[i].GetSeatId() == chairSeat {
			return DispatchGroup{}, false
		}
	}
	return g, true
}

// dispatchEventsOf reads the stream the dispatch folds read, or nil for a run with no record.
func dispatchEventsOf(run Run) ([]*Event, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	evs, _, err := recordsql.EventsW(db)
	return evs, err
}

// RequireChairSittingOpened refuses `dispatch next` in a chair sitting no register opened (see
// unopenedChairSitting). It is the write-time half of the chair's "register for this sitting" work
// item: the verb is the chair's first act every sitting, so refusing it there puts the register
// ahead of every act the sitting records.
func RequireChairSittingOpened(run Run) error {
	evs, err := dispatchEventsOf(run)
	if err != nil {
		return err
	}
	g, owed := unopenedChairSitting(evs)
	if !owed {
		return nil
	}
	return fmt.Errorf("record: dispatch refused — %s", chairRegisterOwed(g))
}

// chairRegisterOwed is the one wording of the chair's owed register, for the refusal and the work
// list alike.
func chairRegisterOwed(g DispatchGroup) string {
	sat := make([]string, 0, len(g.Sat))
	for p := range g.Sat {
		sat = append(sat, p)
	}
	sort.Strings(sat)
	return fmt.Sprintf("%s sat for the dispatch you recorded against report head %d, and you have not registered since. The workflow came back to you, so this is a new sitting, and a sitting is opened by a register — even one that resumes your earlier session. Register for this sitting, then ask again",
		strings.Join(sat, ", "), g.rows[0].pin)
}

// DispatchStands reports whether plan is already the record's standing dispatch: nobody has
// registered since the latest dispatch rows, and those rows name exactly plan's parties, on the
// same gaps, at plan's head. `dispatch next` then records nothing new — asking twice in one
// sitting, for the prose and then for --json, is one decision, and the B5 and B6 chairs each wrote
// it twice. A plan that differs is a new decision and is recorded.
func DispatchStands(run Run, plan Plan) (bool, error) {
	if len(plan.Parties) == 0 || len(plan.Docket) > 0 {
		return false, nil
	}
	evs, err := dispatchEventsOf(run)
	if err != nil || evs == nil {
		return false, err
	}
	groups := DispatchGroups(evs)
	if len(groups) == 0 {
		return false, nil
	}
	g := groups[len(groups)-1]
	for _, e := range evs[g.Last+1:] {
		if e.GetType() == recordpb.EventType_EVENT_TYPE_REGISTER {
			return false, nil
		}
	}
	key := func(seat string, pin int64, gaps []string) string {
		gs := append([]string{}, gaps...)
		sort.Strings(gs)
		return fmt.Sprintf("%s@%d:%s", seat, pin, strings.Join(gs, ","))
	}
	standing := map[string]bool{}
	for _, d := range g.rows {
		standing[key(d.seat, d.pin, d.gaps)] = true
	}
	asked := map[string]bool{}
	for _, p := range plan.Parties {
		asked[key(p.SeatID, plan.Head, p.GapIDs)] = true
	}
	if len(asked) != len(standing) {
		return false, nil
	}
	for k := range asked {
		if !standing[k] {
			return false, nil
		}
	}
	return true, nil
}

// owedSitting is the latest dispatch naming seatID that the seat has not sat for, by sittingFor.
// The stream's positions stand in for events.id: evs is in id order.
func owedSitting(evs []*Event, seatID string) (dispatchRow, bool) {
	seq := make([]int64, len(evs))
	for i := range seq {
		seq[i] = int64(i)
	}
	ds, registers := dispatchLedger(evs, seq)
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].seat != seatID {
			continue
		}
		_, sat := sittingFor(registers[seatID], ds[i])
		return ds[i], !sat
	}
	return dispatchRow{}, false
}

// lensPins is, per seat, the pin of the last dispatch it SAT for (registered after), and the ids
// of the registers that followed a dispatch naming the seat (its sittings for dispatches).
func lensPins(evs []*Event, ids []int64) (map[string]int64, map[string][]int64) {
	ds, registers := dispatchLedger(evs, ids)
	pins := map[string]int64{}
	sat := map[string][]int64{}
	for _, d := range ds {
		r, ok := sittingFor(registers[d.seat], d)
		if !ok {
			continue
		}
		pins[d.seat] = d.pin
		sat[d.seat] = append(sat[d.seat], r)
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
	    g."material", COALESCE(g."class_material", ''),
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
		if err := rows.Scan(&g.id, &g.mintedBy, &g.severity, &g.material, &g.classMaterial, &g.supersededBy, &g.dockets, &g.rulings); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// notMaterialBecause is the one wording of WHY an open gap is not material: its class says never,
// or its class goes by grade and its current grade is below the floor. The dispatch's reasons and
// the chair's work list both say it.
func notMaterialBecause(classMaterial, severity string) string {
	if classMaterial == recordpb.Word(recordpb.ClassMaterial_CLASS_MATERIAL_NEVER) {
		return "its class is never material"
	}
	if severity == "" {
		return "ungraded"
	}
	return "graded " + severity
}
