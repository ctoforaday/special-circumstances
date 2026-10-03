package record

import (
	"fmt"
	"sort"
)

// Keys is every struck act's key, sorted — the Go side of the question the `struck` view answers.
func (x StruckIndex) Keys() []string {
	out := make([]string, 0, len(x.byKey))
	for k := range x.byKey {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// StruckKeys is the SQL side: the keys of the acts the `struck` view names, sorted. A record that
// holds no correction answers the empty list, which is the fact.
func StruckKeys(run Run) ([]string, error) {
	return correctionQuery(run, `SELECT e."key" FROM "struck" s JOIN "events" e ON e."id" = s."event_id" ORDER BY e."key"`)
}

// LiveKeys is the SQL order of the acts that stand: every row of live_event by pos, as event keys
// ("" for an event with none) — the order Live gives in Go, answered by the view instead.
func LiveKeys(run Run) ([]string, error) {
	return correctionQuery(run, `SELECT COALESCE(e."key", '') FROM "live_event" l JOIN "events" e ON e."id" = l."event_id"
	  ORDER BY l."pos", l."event_id"`)
}

// correctionQuery reads one column of keys through the correction views. The open held the record
// to this binary's schema, so the views are there; a read that fails is the caller's error.
func correctionQuery(run Run, q string) ([]string, error) {
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return nil, err
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("record: asking the correction views: %w", err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}
