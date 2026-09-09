package catalogue

import (
	"database/sql"
	"time"
)

// MetaRetainedOn is the UTC date retention last ran, as YYYY-MM-DD.
const MetaRetainedOn = "retained_on"

// DueForRetention reports whether the expensive half of the sweep should run, and is true at most
// once per UTC day.
//
// The marker lives in the STORE rather than in a file beside it, so it cannot drift from the data
// it governs: a store deleted and reprojected starts its day over, which is correct, and a marker
// file left behind would have claimed retention had already run against rows that no longer exist.
func DueForRetention(db *sql.DB, now time.Time) bool {
	today := now.UTC().Format("2006-01-02")
	var last string
	switch err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, MetaRetainedOn).Scan(&last); {
	case err == sql.ErrNoRows:
		return true
	case err != nil:
		return false // cannot tell: do the cheap thing, and let the next run try again
	}
	return last != today
}

// MarkRetained records that retention ran today.
func MarkRetained(db *sql.DB, now time.Time) error {
	_, err := db.Exec(
		`INSERT INTO meta(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
		MetaRetainedOn, now.UTC().Format("2006-01-02"))
	return err
}
