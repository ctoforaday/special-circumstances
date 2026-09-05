package record

import "fmt"

// AnchorIDs returns the distinct finding-marker ids EXPECTED in blue/report.md — the id of
// every `anchor` event on the record, in first-seen order. The blue-report lockdown's
// PostToolUse backstop compares this to the ids actually present to catch a dropped marker.
func AnchorIDs(run Run) ([]string, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return nil, nil
	}
	// First-seen order, distinct, empties dropped — the same three rules the fold applied.
	rows, err := db.Query(`SELECT "id" FROM "anchor" WHERE COALESCE("id", '') != ''
	  GROUP BY "id" ORDER BY MIN("event_id")`)
	if err != nil {
		return nil, fmt.Errorf("record: asking the record for its anchor ids: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ExistingBlueEditByKey reports whether this seat already recorded a `blue_edit` event
// under --key. A crash-retried `blue edit` uses it to RECONCILE (re-apply the write
// idempotently) rather than append a second stack op — the event-first ordering keeps the
// stack durable across the crash window between the event append and the report write.
func ExistingBlueEditByKey(run Run, seatID, key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	// key is non-empty (guarded above), so a row whose edit_key is NULL can never match it —
	// the same property the fold's absent-key zero value had, held by the comparison itself.
	return recordHas(run, `SELECT 1 FROM "blue_edit" b JOIN "events" e ON e."id" = b."event_id"
	  WHERE e."seat_id" = ? AND b."edit_key" = ? LIMIT 1`, seatID, key)
}
