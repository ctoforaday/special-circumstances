package recordsql

import (
	"database/sql"
	"errors"
)

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
