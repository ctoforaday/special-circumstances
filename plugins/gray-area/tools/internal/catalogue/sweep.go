package catalogue

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
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

// Retain deletes what is past the window, by TRANSCRIPT time throughout. It runs at most once a
// UTC day — the caller decides, via LastRetention — because it is the expensive half: measured
// 325 ms to delete a day's rows from a million-row table, against 2.4 ms for a closure tail read.
//
// A SESSION ROW EXPIRES WITH ITS LAST RECORD, never by when the store read it. Keyed on
// ingested_last, a backfill — which stamps every session it reads with the wall clock — kept a
// month of rows whose every act, word and thought had already gone: sessions the store listed and
// could say nothing about. The rule is FindLost's last sign of life: the newest parsed timestamp
// across a session's records, and for a session with none, when it registered. The session rows
// go FIRST, while the records that date them are still there to be read.
//
// A LIVE SESSION IS NEVER EXPIRED, whatever its records say. A session resumed after a month idle
// registers at SessionStart holding nothing newer than the window, and the retention pass that
// follows in the same sweep would delete the row registration just reopened — and with it the
// cloud id, the one fact no transcript holds. `live` is the caller's set: the sessions the client
// advertises, and the hook's own.
//
// No VACUUM: SQLite reuses freed pages, so at steady state the file plateaus rather than growing.
func Retain(db *sql.DB, lim SweepLimits, now time.Time, live map[string]bool) (int64, error) {
	cut := now.Add(-lim.Window).Unix()
	expired, err := db.Query(`
		SELECT session_id FROM session
		WHERE session_id NOT IN (
		        SELECT session_id FROM act     WHERE ts >= ?1 UNION
		        SELECT session_id FROM word    WHERE ts >= ?1 UNION
		        SELECT session_id FROM thought WHERE ts >= ?1 UNION
		        SELECT session_id FROM skip    WHERE at >= ?1)
		  AND (ingested_first < ?1 OR session_id IN (
		        SELECT session_id FROM act     WHERE ts > 0 UNION
		        SELECT session_id FROM word    WHERE ts > 0 UNION
		        SELECT session_id FROM thought WHERE ts > 0 UNION
		        SELECT session_id FROM skip    WHERE at > 0))`, cut)
	if err != nil {
		return 0, fmt.Errorf("catalogue: retention (session): %w", err)
	}
	var gone []string
	for expired.Next() {
		var sid string
		if err := expired.Scan(&sid); err != nil {
			expired.Close()
			return 0, fmt.Errorf("catalogue: retention (session): %w", err)
		}
		if !live[sid] {
			gone = append(gone, sid)
		}
	}
	if err := expired.Close(); err != nil {
		return 0, fmt.Errorf("catalogue: retention (session): %w", err)
	}
	for _, sid := range gone {
		if _, err := db.Exec(`DELETE FROM session WHERE session_id = ?`, sid); err != nil {
			return 0, fmt.Errorf("catalogue: retention (session %s): %w", sid, err)
		}
	}
	var total int64
	for _, q := range []string{
		`DELETE FROM act WHERE ts < ?`,
		`DELETE FROM word WHERE ts < ?`,
		`DELETE FROM thought WHERE ts < ?`,
		`DELETE FROM skip WHERE at < ?`,
	} {
		res, err := db.Exec(q, cut)
		if err != nil {
			return total, fmt.Errorf("catalogue: retention: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}
	// AN OFFSET ROW BELONGS TO A SESSION ROW. Left behind, it names a transcript the store can say
	// nothing about: Closure's queue is file_offset rows whose session has no closed_at — which an
	// absent session row satisfies — so the expired session is pending at every SessionStart with
	// nothing to read and nothing to mark, and `find` stats and counts its path for ever.
	if _, err := db.Exec(`DELETE FROM file_offset WHERE session_id NOT IN (SELECT session_id FROM session)`); err != nil {
		return total, fmt.Errorf("catalogue: retention (file_offset): %w", err)
	}
	return total, nil
}

// MarkGone looks for every transcript the store names and records the ones no longer on disk,
// returning how many are gone now.
//
// THE CLIENT EXPIRES TRANSCRIPT FILES ON ITS OWN CLOCK (cleanupPeriodDays, by a file's last
// write), and the rows read from a file outlive it for as long as their own timestamps allow. That
// is kept on purpose — the rows are then the only record — and this is what says so: gone_at is
// when the store first found the file missing, so "still on disk" and "these rows are all there
// is" stop being the same row.
//
// ONLY A FILE THAT DOES NOT EXIST IS GONE. Any other stat failure leaves the row as it was: an
// unreadable directory is not evidence the transcript went. A file found again is cleared, so a
// corpus restored from a copy does not stay marked.
func MarkGone(db *sql.DB, now time.Time) (int, error) {
	rows, err := db.Query(`SELECT path, gone_at IS NOT NULL FROM file_offset`)
	if err != nil {
		return 0, fmt.Errorf("catalogue: gone transcripts: %w", err)
	}
	var went, back []string
	gone := 0
	for rows.Next() {
		var p string
		var marked bool
		if err := rows.Scan(&p, &marked); err != nil {
			rows.Close()
			return 0, fmt.Errorf("catalogue: gone transcripts: %w", err)
		}
		_, statErr := os.Stat(p)
		switch {
		case errors.Is(statErr, fs.ErrNotExist):
			gone++
			if !marked {
				went = append(went, p)
			}
		case statErr == nil && marked:
			back = append(back, p)
		case statErr != nil && marked:
			gone++ // could not look; the row keeps what the last look found
		}
	}
	if err := rows.Close(); err != nil {
		return 0, fmt.Errorf("catalogue: gone transcripts: %w", err)
	}
	for _, p := range went {
		if _, err := db.Exec(`UPDATE file_offset SET gone_at = ? WHERE path = ?`, now.Unix(), p); err != nil {
			return gone, fmt.Errorf("catalogue: marking %s gone: %w", p, err)
		}
	}
	for _, p := range back {
		if _, err := db.Exec(`UPDATE file_offset SET gone_at = NULL WHERE path = ?`, p); err != nil {
			return gone, fmt.Errorf("catalogue: clearing %s: %w", p, err)
		}
	}
	return gone, nil
}
