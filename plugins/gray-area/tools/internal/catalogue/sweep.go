package catalogue

import (
	"database/sql"
	"fmt"
	"time"
)

// SweepLimits bound one invocation. The cap is on the unit that COSTS — files and bytes — not on
// sessions: offsets are per file, and on the measured corpus the eight largest non-live sessions
// hold 94 files, so a per-session cap bounds nothing.
type SweepLimits struct {
	MaxFiles int
	MaxBytes int64
	// Window is how long rows are kept. Retention is a DELETE rather than an unlink because the
	// store is ONE database: SQLite cannot join across connections and ATTACH caps at 10, so a
	// day-partitioned store could not serve a month-wide query at all.
	Window time.Duration
}

// DefaultLimits: 12 file-tails or 4 MB, whichever binds first, and a one-month window.
//
// Derived rather than asserted: a tail read beyond a stored offset measured 2.4 ms, so 12 tails
// is ~29 ms against the 50 ms steady-state budget, leaving room for the writes. ESTIMATED for the
// combined worst case — SweepStats reports what an invocation actually did, so the budget is
// observable in production rather than only in a test.
func DefaultLimits() SweepLimits {
	return SweepLimits{MaxFiles: 12, MaxBytes: 4 << 20, Window: 30 * 24 * time.Hour}
}

// SweepStats is what one invocation did. Reported so "deferred, not dropped" is observable.
type SweepStats struct {
	FilesRead    int
	BytesRead    int64
	SessionsShut int
	Deferred     int
	RowsExpired  int64
	Elapsed      time.Duration
}

// Closure brings non-live sessions up to date and marks them shut.
//
// # Why it ingests before it decides
//
// A session's final turn is never hook-ingested — there is no later Stop — but the transcript
// DOES receive it: measured, 406 of 409 completed transcripts hold assistant text for their final
// turn. So a sweep that only marked sessions shut would leave that text on disk unread. Closure
// reads first, and only then records the session as settled.
//
// The queue is a QUERY, not a table: non-live sessions holding an offset record with closed_at
// NULL, oldest first. A missing marker therefore reads as "not yet closed", the safe direction,
// because it means the session is still eligible.
func Closure(db *sql.DB, files []TranscriptFile, live map[string]bool, lim SweepLimits) (SweepStats, error) {
	var st SweepStats
	start := time.Now()
	defer func() { st.Elapsed = time.Since(start) }()

	byPath := map[string]TranscriptFile{}
	for _, f := range files {
		byPath[f.Path] = f
	}

	rows, err := db.Query(`
		SELECT DISTINCT fo.session_id
		FROM file_offset fo
		LEFT JOIN session s ON s.session_id = fo.session_id
		WHERE s.closed_at IS NULL
		ORDER BY fo.session_id`)
	if err != nil {
		return st, fmt.Errorf("catalogue: pending queue: %w", err)
	}
	var pending []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err == nil && !live[sid] {
			pending = append(pending, sid)
		}
	}
	rows.Close()

	for _, sid := range pending {
		// Bound on FILES and BYTES, checked before each session rather than after, so one large
		// session cannot overshoot by its whole file count.
		if st.FilesRead >= lim.MaxFiles || st.BytesRead >= lim.MaxBytes {
			st.Deferred = len(pending) - st.SessionsShut
			break
		}
		for _, f := range files {
			if f.SessionID != sid {
				continue
			}
			r, err := IngestFile(db, f)
			if err != nil {
				return st, err
			}
			st.FilesRead++
			st.BytesRead += r.BytesRead
		}
		if _, err := db.Exec(`UPDATE session SET closed_at = ? WHERE session_id = ?`,
			time.Now().Unix(), sid); err != nil {
			return st, fmt.Errorf("catalogue: closing %s: %w", sid, err)
		}
		st.SessionsShut++
	}
	return st, nil
}

// Retain deletes rows older than the window. It runs at most once a UTC day — the caller decides,
// via LastRetention — because it is the expensive half: measured 325 ms to delete a day's rows
// from a million-row table, against 2.4 ms for a closure tail read.
//
// No VACUUM: SQLite reuses freed pages, so at steady state the file plateaus rather than growing.
func Retain(db *sql.DB, lim SweepLimits, now time.Time) (int64, error) {
	cut := now.Add(-lim.Window).Unix()
	var total int64
	for _, q := range []string{
		`DELETE FROM act WHERE ts < ?`,
		`DELETE FROM word WHERE ts < ?`,
		`DELETE FROM thought WHERE ts < ?`,
		`DELETE FROM provisional_skip WHERE at < ?`,
	} {
		res, err := db.Exec(q, cut)
		if err != nil {
			return total, fmt.Errorf("catalogue: retention: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}
	// A session with nothing left is itself expired: keeping the row would report a session that
	// the store can say nothing about, which is worse than not listing it.
	if _, err := db.Exec(`DELETE FROM session WHERE last_seen < ?`, cut); err != nil {
		return total, fmt.Errorf("catalogue: retention (session): %w", err)
	}
	return total, nil
}
