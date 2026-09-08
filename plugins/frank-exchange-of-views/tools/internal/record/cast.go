package record

import (
	"fmt"
	"sort"
	"strings"
)

// HarnessSeat is the seat the harness itself writes under — the cast at setup, the sitting span
// from the hooks. It is not a debate seat: nothing dispatches it and it never registers.
const HarnessSeat = "harness"

// DefaultCastAreas are the lens areas a run dispatches when the operator names none: the four
// that have always sat (debate.js DEFAULT_AREAS). The rest are opt-in per run.
var DefaultCastAreas = []string{"evidence", "logic", "dark-side", "voice"}

// CastFor is the run's admissible seats for a lens-area selection and a lane count
// (plans/roundless.md §III.B.1): a lens per area, the chair, the blue lanes, the synthesizer and
// responder, the frontier, the bench in its three sittings. Petition sittings are not listed —
// InCast admits `judge-petition-<s>` for any cast seat s.
func CastFor(areas []string, lanes int) []string {
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
	if lanes < 1 {
		lanes = 1
	}
	for i := 1; i <= lanes; i++ {
		out = append(out, fmt.Sprintf("blue-lane-%d", i))
	}
	return append(out, "blue-synthesize", "blue-respond", "frontier", "judge", "judge-terminal", "assemble")
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
	for _, s := range cast {
		if s == seatID || "judge-petition-"+s == seatID {
			return true, true, nil
		}
	}
	return false, true, nil
}
