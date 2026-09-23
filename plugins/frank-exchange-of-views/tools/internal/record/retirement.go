package record

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordsql"
)

// PER-SEAT RETIREMENT (plans/feov-lens-bar.md III.5; gblock D1: re-arm ONCE per retirement).
//
// A lens used to be ready whenever the report head moved past its pin, so a lens that found nothing
// sat again on every head move for the whole run: b5's chair readied the voice lens alone eleven
// times at head 144. A lens now retires when it stops finding material, and a head move re-arms it
// once. The cost is bounded: at most two barren sittings per active spell, plus one per retirement.
//
//   - A SITTING is a register following a dispatch naming the lens (sittingFor), including the
//     sittings it is engaged for on its own gaps.
//   - A mint belongs to the sitting whose register most recently precedes it.
//   - A sitting is PRODUCTIVE when it minted a fresh gap (superseding nothing) that is material NOW.
//   - ACTIVE: fewer than two trailing barren sittings. A productive sitting returns a lens here,
//     with its re-arm unspent. RETIRED: two barren sittings, re-arm unspent. RETIRED FOR GOOD: its
//     re-arm sitting was barren.
//
// Readiness keeps the head precondition: nothing is ready before a report is ingested. An active
// lens is ready; a retired lens is ready once when the head has moved past its pin; a lens retired
// at the head, or for good, is not. A lens retired for good still answers its own open gaps.

// LensState is where a cast lens stands in its retirement fold.
type LensState int

const (
	LensActive LensState = iota
	LensRetired
	LensRetiredForGood
)

func (s LensState) String() string {
	switch s {
	case LensRetired:
		return "retired"
	case LensRetiredForGood:
		return "retired for good"
	}
	return "active"
}

// lensFold is one cast lens, folded.
type lensFold struct {
	seat     string
	state    LensState
	barren   int   // trailing barren sittings in the current active spell
	sittings int   // sittings on the record
	pin      int64 // pin of the last dispatch it sat for; 0 when it never sat
	ready    bool
	why      string
}

// StaleArea is a lens retired for good whose pin the report head has moved past: text it will not
// audit again changed after it last sat. The chair reads those changes against the area's duties
// and names the area in its spot-check before a PASS.
type StaleArea struct {
	SeatID string `json:"seat_id"`
	Pin    int64  `json:"pin"`
}

// lensStates folds every cast lens, in cast order. ids are the events' row ids, aligned with evs;
// head is the report head (0: nothing ingested); fresh is the set of gap ids minted fresh and
// material now.
func lensStates(evs []*Event, ids []int64, head int64, fresh map[string]bool) []lensFold {
	cast := castOfEvents(evs)
	ds, registers := dispatchLedger(evs, ids)
	var out []lensFold
	for _, seat := range cast {
		if !strings.HasPrefix(seat, "red-lens-") {
			continue
		}
		f := lensFold{seat: seat}
		sat := map[int64]bool{}
		var sittings []int64
		for _, d := range ds {
			if d.seat != seat {
				continue
			}
			r, ok := sittingFor(registers[seat], d)
			if !ok {
				continue
			}
			f.pin = d.pin
			if !sat[r] {
				sat[r] = true
				sittings = append(sittings, r)
			}
		}
		sort.Slice(sittings, func(a, b int) bool { return sittings[a] < sittings[b] })
		productive := map[int64]bool{}
		for i, e := range evs {
			if e.GetSeatId() != seat {
				continue
			}
			m, ok := recordpb.BodyAs[*recordpb.Mint](e)
			if !ok || !fresh[m.GetGapId()] {
				continue
			}
			// the sitting whose register most recently precedes the mint
			k := sort.Search(len(sittings), func(j int) bool { return sittings[j] >= ids[i] }) - 1
			if k >= 0 {
				productive[sittings[k]] = true
			}
		}
		for _, s := range sittings {
			f.sittings++
			if productive[s] {
				f.state, f.barren = LensActive, 0
				continue
			}
			switch f.state {
			case LensActive:
				f.barren++
				if f.barren >= 2 {
					f.state = LensRetired
				}
			case LensRetired:
				f.state = LensRetiredForGood
			}
		}
		switch {
		case head == 0:
			f.why = fmt.Sprintf("%s: %s — not ready, no report has been ingested", seat, f.state)
		case f.state == LensActive && f.sittings == 0:
			f.ready, f.why = true, fmt.Sprintf("%s: active, never sat — audits the report at head %d", seat, head)
		case f.state == LensActive:
			f.ready, f.why = true, fmt.Sprintf("%s: active (%d barren sitting(s) since it last minted fresh material; it retires at 2) — audits the report at head %d", seat, f.barren, head)
		case f.state == LensRetired && head > f.pin:
			f.ready, f.why = true, fmt.Sprintf("%s: retired, re-armed once — head %d moved past its pin %d", seat, head, f.pin)
		case f.state == LensRetired:
			f.why = fmt.Sprintf("%s: retired at the head (pin %d) — ready once when the head moves", seat, f.pin)
		default:
			f.why = fmt.Sprintf("%s: retired for good (pin %d) — not ready", seat, f.pin)
		}
		out = append(out, f)
	}
	return out
}

// staleAreasOf is every lens retired for good whose pin is behind the head, in cast order.
func staleAreasOf(folds []lensFold, head int64) []StaleArea {
	out := []StaleArea{}
	for _, f := range folds {
		if f.state == LensRetiredForGood && f.pin < head {
			out = append(out, StaleArea{SeatID: f.seat, Pin: f.pin})
		}
	}
	return out
}

// freshMaterialOf is the gap view's answer to "which gaps were minted fresh and are material now".
func freshMaterialOf(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT "gap_id" FROM "gap" WHERE "material" AND "supersedes_count" = 0`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its fresh material gaps: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// freshMaterialOfStates is the same answer off the work path's gap states, which carry the view's
// columns.
func freshMaterialOfStates(gaps []WorkGapState) map[string]bool {
	out := map[string]bool{}
	for _, g := range gaps {
		if g.Material && len(g.Supersedes) == 0 {
			out[g.ID] = true
		}
	}
	return out
}

// reportHeadOf is the report head off the stream: the last blue_edit or base_ingest, by row id.
func reportHeadOf(evs []*Event, ids []int64) int64 {
	var head int64
	for i, e := range evs {
		switch e.GetType() {
		case recordpb.EventType_EVENT_TYPE_BLUE_EDIT, recordpb.EventType_EVENT_TYPE_BASE_INGEST:
			if ids[i] > head {
				head = ids[i]
			}
		}
	}
	return head
}

// passLensGate is the PASS gate's lens conditions as ONE fold both halves read: the Gate write path
// refuses from it (requireNoCastLensReady, requirePassCoversStaleAreas), and the chair's work list
// states each of its items from it, so the list cannot say a PASS is open that the gate refuses.
// The dispatch readies from the same lensStates fold.
type passLensGate struct {
	cast    bool
	head    int64
	lenses  []lensFold
	stale   []StaleArea
	covered map[string]bool // areas a spot-check in the chair's current sitting named
}

func passLensGateOf(evs []*Event, ids []int64, fresh map[string]bool) passLensGate {
	g := passLensGate{cast: castOfEvents(evs) != nil, head: reportHeadOf(evs, ids), covered: map[string]bool{}}
	if !g.cast {
		return g
	}
	g.lenses = lensStates(evs, ids, g.head, fresh)
	g.stale = staleAreasOf(g.lenses, g.head)
	// THIS SITTING is the chair's: its spot-checks after the register that OPENED its latest
	// sitting. That is opensASitting's question — the window an act is attributed to — and asking
	// it there rather than on the event type is what keeps this level with the other attribution
	// readers. The chair cannot repair (checkRepair admits a blue role only), so the two readings
	// agree on every record a verb can write; this is the one that stays right if that gate moves.
	for i := len(evs) - 1; i >= 0; i-- {
		e := evs[i]
		if e.GetSeatId() != chairSeat {
			continue
		}
		if recordpb.OpensASitting(e) {
			break
		}
		if sc, ok := recordpb.BodyAs[*recordpb.SpotCheck](e); ok {
			for _, a := range sc.GetAreas() {
				g.covered[a] = true
			}
		}
	}
	return g
}

// ready is each cast lens the fold leaves ready.
func (g passLensGate) ready() []lensFold {
	var out []lensFold
	for _, f := range g.lenses {
		if f.ready {
			out = append(out, f)
		}
	}
	return out
}

// uncovered is each stale area no spot-check in the chair's current sitting names.
func (g passLensGate) uncovered() []StaleArea {
	var out []StaleArea
	for _, a := range g.stale {
		if !g.covered[a.SeatID] {
			out = append(out, a)
		}
	}
	return out
}

func readyReason(f lensFold) string {
	if f.state == LensRetired {
		return "re-arm owed"
	}
	return "active"
}

// statements is every refusal the gate would make, one item per reason, for the chair's list.
func (g passLensGate) statements() []string {
	if !g.cast {
		return nil
	}
	if g.head == 0 {
		return []string{"no report has been ingested — PASS is refused"}
	}
	var out []string
	for _, f := range g.ready() {
		out = append(out, fmt.Sprintf("lens %s is ready (%s) — PASS is refused while it is", f.seat, readyReason(f)))
	}
	for _, a := range g.uncovered() {
		out = append(out, fmt.Sprintf("area %s is behind its pin %d — PASS is refused until a spot-check this sitting names it", a.SeatID, a.Pin))
	}
	return out
}

// passLensGateOfRun reads the gate's inputs off the run.
func passLensGateOfRun(run Run) (passLensGate, bool, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return passLensGate{}, false, err
	}
	evs, _, err := recordsql.EventsW(db)
	if err != nil {
		return passLensGate{}, false, err
	}
	ids, err := eventIDs(db)
	if err != nil {
		return passLensGate{}, false, err
	}
	fresh, err := freshMaterialOf(db)
	if err != nil {
		return passLensGate{}, false, err
	}
	return passLensGateOf(evs, ids, fresh), true, nil
}

// requireNoCastLensReady refuses a PASS while the retirement fold leaves any cast lens ready — the
// same fold the dispatch readies from, so a chair that ignores pass_permitted cannot record a PASS
// over an active lens or an owed re-arm. It keeps the head-0 refusal. A record with no cast has no
// lenses to wait for.
func requireNoCastLensReady(run Run) error {
	g, ok, err := passLensGateOfRun(run)
	if err != nil || !ok || !g.cast {
		return err
	}
	if g.head == 0 {
		return fmt.Errorf("record: verdict PASS refused — no report has been ingested or edited, so there is nothing a lens could have audited")
	}
	ready := g.ready()
	if len(ready) == 0 {
		return nil
	}
	var why []string
	for _, f := range ready {
		why = append(why, f.why)
	}
	return fmt.Errorf("record: verdict PASS refused — %d cast lens(es) are ready: %s. A lens is ready while it is active (it minted fresh material within its last two sittings) or retired with its one re-arm owed; `dispatch next` readies them",
		len(ready), strings.Join(why, "; "))
}

// requirePassCoversStaleAreas refuses a PASS until a spot-check in the chair's current sitting names
// every stale area: text changed behind a lens retired for good, and the chair's read is what stands
// in for the sitting that lens will not take.
func requirePassCoversStaleAreas(run Run) error {
	g, ok, err := passLensGateOfRun(run)
	if err != nil || !ok || !g.cast {
		return err
	}
	un := g.uncovered()
	if len(un) == 0 {
		return nil
	}
	var names []string
	for _, a := range un {
		names = append(names, fmt.Sprintf("%s (pin %d)", a.SeatID, a.Pin))
	}
	return fmt.Errorf("record: verdict PASS refused — the report changed behind %d lens(es) retired for good: %s. Read the changes since each pin against that area's duties, and name the areas in a spot-check this sitting (--areas) before a PASS; a defect you find there goes in that spot-check",
		len(un), strings.Join(names, ", "))
}

// LastSittingJSON is a lens's last sitting, as its work view carries it: the latest sitting STRICTLY
// BEFORE the dispatch it is now sitting for. Kind is first | behind | unchanged | undispatched.
type LastSittingJSON struct {
	Kind string `json:"kind"`
	Pin  int64  `json:"pin"`
	Head int64  `json:"head"`
}

// lastSittingBefore is NOT read from lensPins. The lens reads its work view after registering, and
// sittingFor counts that register as having sat — so lensPins would report the current sitting and
// every lens would read `unchanged`. D is the latest dispatch naming the seat (the one it sits for,
// or owes); a prior sitting is a dispatch d before D whose sitting register is also before D, so an
// unsat earlier dispatch cannot borrow the current register.
func lastSittingBefore(evs []*Event, ids []int64, seatID string) LastSittingJSON {
	ds, registers := dispatchLedger(evs, ids)
	var cur *dispatchRow
	for i := len(ds) - 1; i >= 0; i-- {
		if ds[i].seat == seatID {
			cur = &ds[i]
			break
		}
	}
	if cur == nil {
		return LastSittingJSON{Kind: "undispatched"}
	}
	var prior *dispatchRow
	for i := range ds {
		d := ds[i]
		if d.seat != seatID || d.at >= cur.at {
			continue
		}
		if r, ok := sittingFor(registers[seatID], d); ok && r < cur.at {
			prior = &ds[i]
		}
	}
	ls := LastSittingJSON{Head: cur.pin}
	switch {
	case prior == nil:
		ls.Kind = "first"
	case prior.pin < cur.pin:
		ls.Kind, ls.Pin = "behind", prior.pin
	default:
		ls.Kind, ls.Pin = "unchanged", prior.pin
	}
	return ls
}
