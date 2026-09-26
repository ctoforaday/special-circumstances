package record

import "fmt"

// WHERE A SEAT'S WALL CLOCK WENT, at the grain the aggregate hides.
//
// seat_metrics answers this per SEAT — turns, tokens, wall_ms — and the cost surface renders it. What
// it cannot answer is where a seat's wall clock GOES, because a total is the one shape that cannot
// show a stall: sixty turns averaging four seconds and fifty-nine fast turns beside one that took
// four minutes are the same number.
//
// The two views these read have existed, unwired, since the analysis that motivated them (#684 F11)
// decomposed a run BY HAND. That analysis is the standing reason to have this: the way to make a run
// faster is to find which turns are slow and what those turns were doing.
//
// THE BUCKETS ARE FACTS, NOT JUDGEMENTS, and the views say why: a threshold like "70-83 tok/s is
// healthy generation" was chosen once, from one run, and freezing it into the schema would apply it to
// every future run by readers who never saw it chosen. So the split is by what the turn CONTAINED —
// thinking, a tool call, both, neither — and span_ms and output_tokens come through per turn so a
// reader applies its own cutoffs in the open, where the next reader can disagree.

// SeatTurnSpan is one turn of one seat, with how long it took and what it held.
type SeatTurnSpan struct {
	AgentID string `json:"agent_id"`
	TurnIdx int    `json:"turn_idx"`
	// SpanMillis is the gap to the turn before it, so the FIRST turn of a seat has none. Null rather
	// than zero: zero would say "instant", and anything summing it under-reports every seat by its
	// opening turn.
	SpanMillis   *int64 `json:"span_ms"`
	IsThinking   bool   `json:"is_thinking"`
	IsTool       bool   `json:"is_tool"`
	OutputTokens int64  `json:"output_tokens"`
}

// SeatTimeBucket is a seat's turns grouped by what they contained.
type SeatTimeBucket struct {
	AgentID string `json:"agent_id"`
	Bucket  string `json:"bucket"`
	Turns   int    `json:"turns"`
	// SpanMillis is null when NO turn in the bucket has a span — a bucket whose only member is the
	// seat's opening turn, which has no predecessor to be measured against. That is a bucket with a
	// real turn count and no measurable duration, and it is not a zero: zero would put a seat that
	// spoke once into a total as instantaneous. Found by the first bucket of real data this was ever
	// rendered on, having read clean on a record that held no turns at all.
	SpanMillis   *int64 `json:"span_ms"`
	OutputTokens int64  `json:"output_tokens"`
}

// SeatTiming is both grains together: the decomposition a reader scans, and the per-turn spans it
// needs to apply a cutoff of its own to anything the decomposition makes it curious about.
type SeatTiming struct {
	// Measured says whether the run holds per-turn measurements AT ALL, and it is here because the
	// absence and the honest zero are otherwise the same bytes: a run whose transcripts were never read
	// renders the empty pair below, and so would a reader that had silently stopped being fed. The two
	// call for different acts — one is a capture that has not run, the other is a defect — so the
	// distinction is a field rather than something a reader infers from emptiness.
	Measured bool             `json:"measured"`
	Buckets  []SeatTimeBucket `json:"buckets"`
	Turns    []SeatTurnSpan   `json:"turns"`
}

// SeatTimingOf reads both views. An absent seat_turn table is not an error — a run captured before it
// existed, or one whose transcripts could not be read, has nothing to report, and the caller renders
// that absence rather than a table of zeroes.
func SeatTimingOf(run Run) (SeatTiming, error) {
	out := SeatTiming{Buckets: []SeatTimeBucket{}, Turns: []SeatTurnSpan{}}
	db, err := openRunForRead(run)
	if err != nil || db == nil {
		return out, err
	}
	brows, err := db.Query(`SELECT "agent_id", "bucket", "turns", "span_ms", "output_tokens"
	  FROM "seat_time_decomposition" ORDER BY "agent_id", "bucket"`)
	if err != nil {
		return out, fmt.Errorf("record: reading the seat time decomposition: %w", err)
	}
	for brows.Next() {
		var b SeatTimeBucket
		if err := brows.Scan(&b.AgentID, &b.Bucket, &b.Turns, &b.SpanMillis, &b.OutputTokens); err != nil {
			brows.Close()
			return out, fmt.Errorf("record: scanning a time bucket: %w", err)
		}
		out.Buckets = append(out.Buckets, b)
	}
	brows.Close()
	if err := brows.Err(); err != nil {
		return out, err
	}

	// Asked of the TABLE, not of the views: a run can hold turns whose spans no bucket admits, and
	// "measured" is about whether the measurement was taken, never about what it found.
	var rows int
	if err := db.QueryRow(`SELECT count(*) FROM "seat_turn"`).Scan(&rows); err != nil {
		return out, fmt.Errorf("record: counting the per-turn measurements: %w", err)
	}
	out.Measured = rows > 0

	trows, err := db.Query(`SELECT "agent_id", "turn_idx", "span_ms", "is_thinking", "is_tool", "output_tokens"
	  FROM "seat_turn_span" ORDER BY "agent_id", "turn_idx"`)
	if err != nil {
		return out, fmt.Errorf("record: reading the seat turn spans: %w", err)
	}
	defer trows.Close()
	for trows.Next() {
		var t SeatTurnSpan
		if err := trows.Scan(&t.AgentID, &t.TurnIdx, &t.SpanMillis, &t.IsThinking, &t.IsTool, &t.OutputTokens); err != nil {
			return out, fmt.Errorf("record: scanning a turn span: %w", err)
		}
		out.Turns = append(out.Turns, t)
	}
	return out, trows.Err()
}
