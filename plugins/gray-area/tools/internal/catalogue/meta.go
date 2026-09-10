package catalogue

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

// MetaRetainedOn is the UTC date retention last ran, as YYYY-MM-DD.
const MetaRetainedOn = "retained_on"

// The rebuild markers. A rebuilt store is EMPTY, and an empty store answers every question with a
// plausible zero — so the fact that it was rebuilt, and whether anything has re-read the corpus
// since, is kept where every read verb can find it.
//
// The notice Open prints at the rebuild is not that signal: the first Open after an upgrade is
// almost always a hook's, and a hook's exit-0 stderr goes to the client's debug log and nowhere
// else. These fields are.
//
// Both timestamps are Unix seconds from the WALL clock, never an injected one: they are compared
// with each other, and a frozen test clock on one side would make the comparison meaningless.
const (
	MetaRebuiltAt    = "rebuilt_at"    // when Open rebuilt the store, Unix seconds
	MetaRebuiltFrom  = "rebuilt_from"  // the stamp the store carried before the rebuild
	MetaBackfilledAt = "backfilled_at" // when `telepathy backfill` last completed, Unix seconds
)

// execer is a *sql.DB or a *sql.Tx.
type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func setMeta(x execer, key, value string) error {
	_, err := x.Exec(
		`INSERT INTO meta(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`,
		key, value)
	return err
}

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
	return setMeta(db, MetaRetainedOn, now.UTC().Format("2006-01-02"))
}

// MarkRebuilt records that the store was rebuilt from stamp `from` at `at`. Open calls it inside
// the rebuild's own transaction, so a rebuild that commits always carries its markers.
func MarkRebuilt(x execer, from int, at time.Time) error {
	if err := setMeta(x, MetaRebuiltAt, strconv.FormatInt(at.Unix(), 10)); err != nil {
		return err
	}
	return setMeta(x, MetaRebuiltFrom, strconv.Itoa(from))
}

// MarkBackfilled records that a backfill completed at `at`, which clears RebuildPending for every
// rebuild at or before that second.
func MarkBackfilled(db *sql.DB, at time.Time) error {
	return setMeta(db, MetaBackfilledAt, strconv.FormatInt(at.Unix(), 10))
}

// RebuildPending reports whether the store was rebuilt and not backfilled since: pending iff
// `rebuilt_at` exists AND `backfilled_at` is absent or STRICTLY older. Equal seconds count as
// backfilled — a backfill that finishes in the rebuild's own second did read the corpus after it.
//
// A marker that exists and does not parse is an ERROR, not "not pending": the caller must be able
// to say "not measured" rather than fold an unreadable marker into a clean board.
func RebuildPending(db *sql.DB) (rebuiltAt time.Time, rebuiltFrom int, pending bool, err error) {
	at, ok, err := metaInt(db, MetaRebuiltAt)
	if err != nil || !ok {
		return time.Time{}, 0, false, err
	}
	rebuiltAt = time.Unix(at, 0).UTC()
	from, ok, err := metaInt(db, MetaRebuiltFrom)
	if err != nil {
		return rebuiltAt, 0, false, err
	}
	if !ok {
		return rebuiltAt, 0, false, fmt.Errorf("%s is set but %s is not", MetaRebuiltAt, MetaRebuiltFrom)
	}
	rebuiltFrom = int(from)
	back, ok, err := metaInt(db, MetaBackfilledAt)
	if err != nil {
		return rebuiltAt, rebuiltFrom, false, err
	}
	return rebuiltAt, rebuiltFrom, !ok || back < at, nil
}

// metaInt reads one integer marker; ok is false when the key is absent.
func metaInt(db *sql.DB, key string) (n int64, ok bool, err error) {
	var s string
	switch err := db.QueryRow(`SELECT value FROM meta WHERE key = ?`, key).Scan(&s); {
	case err == sql.ErrNoRows:
		return 0, false, nil
	case err != nil:
		return 0, false, fmt.Errorf("reading meta.%s: %w", key, err)
	}
	n, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("meta.%s = %q is not an integer", key, s)
	}
	return n, true, nil
}
