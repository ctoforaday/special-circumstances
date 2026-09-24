package record

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatclass"
)

// HarnessSeat is the seat the harness itself writes under — the cast at setup, the sitting span
// from the hooks. It is not a debate seat: nothing dispatches it and it never registers.
const HarnessSeat = "harness"

// DefaultCastAreas are the lens areas a run dispatches when the operator names none: every area
// (debate.js DEFAULT_AREAS). An operator narrows the cast only with a reason; retirement bounds
// what a lens that finds nothing costs.
var DefaultCastAreas = LensAreas

// CastFor is the run's admissible seats for a lens-area selection and a lane count
// (plans/roundless.md §III.B.1): a lens per area, the chair, the blue lanes, the synthesizer and
// responder, the frontier, the bench in its three sittings. Petition sittings are not listed —
// InCast admits exactly the seat ids the cast names.
func CastFor(areas []string, lanes int) (seats, laneSeats []string) {
	if len(areas) == 0 {
		areas = DefaultCastAreas
	}
	sorted := append([]string{}, areas...)
	sort.Strings(sorted)
	out := []string{}
	for _, a := range sorted {
		out = append(out, "red-lens-"+strings.TrimPrefix(a, "red-lens-"))
	}
	out = append(out, "red-chair")
	laneSeats = LaneSeatIDs(lanes)
	out = append(out, laneSeats...)
	return append(out, "blue-synthesize", "blue-respond", "frontier", "judge"), laneSeats
}

// LaneSeatIDs are the lane seat ids a run with this many lanes seats, and this is the ONE place
// that composes one.
//
// A LANE ID IS GENERATED, SO IT IS COMPARED RATHER THAN PARSED. How many lanes a run has is a run
// parameter, so no table compiled into this binary can enumerate them — which is why the roster
// carried a pattern for lanes and the tier join, the register gate and the coverage audit each
// asked that pattern a question. A pattern bounds a shape and answers nothing about membership,
// and the coverage audit went further and read the lane's INDEX back out of its name.
//
// Generating the set closes both. "Is this a lane" is `slices.Contains`, and "which lane is
// missing" is a NAME the caller already holds rather than a number recovered from one. Nothing in
// the tree matches a seat id against a pattern any more.
//
// A run that declares no lanes still seats one: a debate with no blue lane has nothing to audit.
func LaneSeatIDs(lanes int) []string {
	if lanes < 1 {
		lanes = 1
	}
	out := make([]string, 0, lanes)
	for i := 1; i <= lanes; i++ {
		out = append(out, fmt.Sprintf("%s%d", seatclass.LaneSeatPrefix, i))
	}
	return out
}

// LanesDeclared is the lane count the run's config declares, and whether it declares one.
//
// THE THREE ANSWERS ARE KEPT APART: a declared count, a run that declared none (`--lanes` is
// optional), and a config that could not be read at all. Only the first can be checked, and the
// other two must not read as agreement — which is why this returns a second value rather than 0.
//
// It lives beside LaneSeatIDs because the two are one question: how many lanes, and therefore
// which ids. The coverage audit in internal/capture had its own copy of this reader beside its own
// copy of the lane pattern.
func LanesDeclared(run Run) (n int, declared bool) {
	b, err := os.ReadFile(filepath.Join(run.Dir(), "inputs", "run-config.json"))
	if err != nil {
		return 0, false
	}
	var cfg map[string]any
	if json.Unmarshal(b, &cfg) != nil {
		return 0, false
	}
	// setup writes it as a STRING (ptrOrNil over the flag), so that is what is read first; a
	// number is accepted too rather than silently missed if the writer ever changes.
	switch v := cfg["lanes"].(type) {
	case string:
		if v == "" {
			return 0, false
		}
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil || n <= 0 {
			return 0, false
		}
		return n, true
	case float64:
		if v <= 0 {
			return 0, false
		}
		return int(v), true
	}
	return 0, false
}

// LaneSeatsOf is the lane seat ids this run's CAST names, or none.
//
// THE CAST AND NOT run-config, and the difference is which way the answer FAILS. run-config's
// `lanes` is optional and absent on an older run; keying lane-ness on it made a lane whose count
// was never declared undispatchable — the register gate refused a seat that had done nothing
// wrong. The cast is written by setup before any seat registers and is the record's own statement
// of which seats exist, so a run that has lanes says so.
func LaneSeatsOf(run Run) []string {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil
	}
	rows, err := db.Query(`SELECT v."value" FROM "cast_lane_seat_ids" v
	  WHERE v."event_id" = (SELECT MAX("event_id") FROM "cast")
	  ORDER BY v."ord"`)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if rows.Scan(&s) != nil {
			return nil
		}
		out = append(out, s)
	}
	return out
}

// laneSeatsOfEvents is LaneSeatsOf read off a stream already in hand.
func laneSeatsOfEvents(evs []*Event) []string {
	var out []string
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Cast](e); ok {
			out = append([]string{}, c.GetLaneSeatIds()...)
		}
	}
	return out
}

// IsLaneSeat reports whether seatID is one of the lane seats a run with this many lanes seats.
func IsLaneSeat(seatID string, lanes int) bool {
	for _, id := range LaneSeatIDs(lanes) {
		if id == seatID {
			return true
		}
	}
	return false
}

// castOfEvents is CastOf read off a stream already in hand: the seats of the LAST cast event, or
// nil when the stream holds none — the same answer as the cast table, for a fold that holds only
// the events.
func castOfEvents(evs []*Event) []string {
	var out []string
	for _, e := range evs {
		if c, ok := recordpb.BodyAs[*recordpb.Cast](e); ok {
			out = append([]string{}, c.GetSeatIds()...)
		}
	}
	return out
}

// CastOf is the run's admissible seats — the Cast event setup wrote under `harness` before any
// seat registered (plans/roundless.md §III.B.1). Nil with no error when the record holds no cast:
// that is a fact the callers say something about (register and the dispatch verb refuse to act
// without one), not one to invent a default for. The LAST cast written wins, and setup writes one.
func CastOf(run Run) ([]string, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(`SELECT v."value" FROM "cast_seat_ids" v
	  WHERE v."event_id" = (SELECT MAX("event_id") FROM "cast")
	  ORDER BY v."ord"`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its cast: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// InCast reports whether seatID is admissible: a member of the cast, or a petition sitting by a
// member (`judge-petition-<s>`, the shape roster.go admits for a cast seat s). The second value is
// whether the record HAS a cast at all, so a caller can tell "not in it" from "nothing to be in".
func InCast(run Run, seatID string) (member, hasCast bool, err error) {
	cast, err := CastOf(run)
	if err != nil {
		return false, false, err
	}
	if cast == nil {
		return false, false, nil
	}
	// THE CAST ADMITS A SEAT ID AND NOTHING DERIVED FROM ONE. It used to also admit
	// "judge-petition-"+s, because a petition sitting was its own seat id named for the petitioner.
	// The bench is one seat now: a petition is a question put to `judge`, and who filed is on the
	// petition rather than in the identity of the seat ruling it.
	for _, s := range cast {
		if s == seatID {
			return true, true, nil
		}
	}
	return false, true, nil
}
