package recordsql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// EventByKey reads the one event carrying key, with its record id. found=false is the honest
// answer for a key no event carries; any other failure is an error, never folded into "absent".
func EventByKey(db *sql.DB, key string) (ev *recordpb.Event, id int64, found bool, err error) {
	if err := db.QueryRow(`SELECT "id" FROM "events" WHERE "key" = ?`, key).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, false, nil
		}
		return nil, 0, false, fmt.Errorf("recordsql: reading the event keyed %q: %w", key, err)
	}
	evs, _, err := eventsWhere(db, false, ` WHERE id = ?`, id)
	if err != nil {
		return nil, 0, false, err
	}
	if len(evs) != 1 {
		return nil, 0, false, fmt.Errorf("recordsql: event %d (key %q) read back as %d events", id, key, len(evs))
	}
	return evs[0], id, true, nil
}

// viewSQL is the text SQLite stored for this view, or "" if the database has no such view. It
// returns the DEFINITION rather than a bool because a view that is present and DIFFERENT is its own
// answer — see requireDeclaredSchema.
func viewSQL(q queryRower, view string) (string, error) {
	var sqlText sql.NullString
	err := q.QueryRow(`SELECT "sql" FROM sqlite_master WHERE type = 'view' AND "name" = ?`, view).Scan(&sqlText)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return sqlText.String, nil
}
