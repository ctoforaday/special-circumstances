package record

import (
	"fmt"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

// AppendSeatTurns records a seat's per-turn measurements: what each API round cost and what kind
// of turn it was. Returns how many rows were newly written.
//
// # It is telemetry, and it is kept apart from the log for that reason
//
// Nothing here goes through Append. An event is an ACT BY A SEAT, written by that seat, and the
// adversarial process turns on the log holding only what the parties did. A turn is a MEASUREMENT
// ABOUT a seat, taken from outside after the fact by whoever read the transcript — the seat has no
// access to its own token accounting. Putting thousands of those a run into the same log would
// make "the record" a mixture of testimony and instrumentation, and the events triggers exist to
// say that the first kind cannot be rewritten.
//
// # INSERT OR IGNORE, and why that is the right shape rather than a convenience
//
// `capture` is re-runnable and a transcript GROWS between runs, so a second ingest legitimately
// arrives holding turns 1..121 again plus a new 122. IGNORE makes the already-recorded ones a
// no-op and adds only what is new. Nothing is ever rewritten — the same guarantee the events
// triggers give, obtained here from the primary key, so a re-ingest cannot quietly revise what an
// earlier one measured.
//
// UPDATE would be the wrong repair for a disagreement: if a turn already recorded now reads
// differently, the transcript changed under us, and silently taking the newer number is how a
// measurement stops being evidence.
func AppendSeatTurns(run Run, agentID string, turns []seatturn.Turn) (int, error) {
	if agentID == "" {
		// FILED AGAINST NOTHING IS WORSE THAN NOT FILED. A row keyed on an empty agent joins to
		// no register event, so it would sit in the table contributing to totals while belonging
		// to no seat — present in the count and absent from every per-seat view.
		return 0, fmt.Errorf("record: refusing to file %d seat turns against an empty agent id", len(turns))
	}
	if len(turns) == 0 {
		return 0, nil
	}
	db, err := openRun(run)
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO "seat_turn"
	  ("agent_id","turn_idx","ts_ms","model","input_tokens","output_tokens","cache_read","cache_creation","is_thinking","is_tool")
	  VALUES (?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return 0, fmt.Errorf("record: preparing the seat-turn insert: %w", err)
	}
	defer stmt.Close()

	written := 0
	for _, t := range turns {
		res, err := stmt.Exec(agentID, t.Index, t.TSMillis, t.Model,
			t.Input, t.Output, t.CacheRead, t.CacheCreation, boolInt(t.Thinking), boolInt(t.Tool))
		if err != nil {
			return 0, fmt.Errorf("record: writing turn %d for agent %s: %w", t.Index, agentID, err)
		}
		if n, aerr := res.RowsAffected(); aerr == nil {
			written += int(n)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("record: committing %d seat turns: %w", len(turns), err)
	}
	return written, nil
}

// boolInt is the STRICT-table encoding: the column is INTEGER, and SQLite's default affinity
// would have accepted a string into it and stored it as one, which is the typing this schema was
// migrated to prevent.
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// CountSeatTurns is how many per-turn rows the run holds. It exists for the tests and the
// capture line that reports the ingest; the analytical questions are views, not counters.
func CountSeatTurns(run Run) (int, error) {
	db, err := openRunForRead(run)
	if err != nil {
		return 0, err
	}
	if db == nil {
		return 0, nil
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "seat_turn"`).Scan(&n); err != nil {
		return 0, fmt.Errorf("record: counting seat turns: %w", err)
	}
	return n, nil
}
