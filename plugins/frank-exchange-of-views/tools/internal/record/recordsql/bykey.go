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

// HasTable answers whether the record's database holds a table. A run's schema is fixed when its
// database is created, so a table added since is absent from an older run — and for a table whose
// rows only this binary can write (the correction), its absence is the true answer "none were
// written", not a miss.
func HasTable(q queryRower, table string) (bool, error) {
	var n int
	if err := q.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&n); err != nil {
		return false, fmt.Errorf("recordsql: asking whether the record has a %q table: %w", table, err)
	}
	return n > 0, nil
}

// RequireTable refuses a write that needs a table an older run's database does not have, in the
// words olderSchema uses for the same fact on the read path: the cause, and the migration that is
// the way out.
func RequireTable(q queryRower, table string) error {
	ok, err := HasTable(q, table)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("recordsql: this run's record has no %q table — %s", table, olderRunAdvice)
	}
	return nil
}
