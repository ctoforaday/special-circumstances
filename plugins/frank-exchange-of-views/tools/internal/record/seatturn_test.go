package record

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

func turns(n int) []seatturn.Turn {
	out := make([]seatturn.Turn, n)
	for i := range out {
		out[i] = seatturn.Turn{Index: i, TSMillis: int64(1000 + i), Model: "m", Input: 10, Output: 2, Thinking: i%2 == 0}
	}
	return out
}

func countTurns(t *testing.T, run Run) int {
	t.Helper()
	db, err := openRun(run)
	if err != nil {
		t.Fatal(err)
	}
	var n int
	if err := db.QueryRow(`SELECT count(*) FROM "seat_turn"`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// RE-INGEST IS A NO-OP ON WHAT IS ALREADY THERE, AND PICKS UP WHAT IS NEW.
//
// capture is re-runnable and a transcript GROWS between runs, so the second ingest legitimately
// arrives holding every earlier turn again. Without the key this would double the run's measured
// cost each time it was re-run — a number that looks like a fact and is an artefact of how often
// the operator pressed the button.
func TestSeatTurnsIngestIsIdempotentAndAppends(t *testing.T) {
	run := runFor(t)

	n, err := AppendSeatTurns(run, "a1", turns(5))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 || countTurns(t, run) != 5 {
		t.Fatalf("first ingest wrote %d, table holds %d, want 5 and 5", n, countTurns(t, run))
	}

	// The same five again, as a re-run of capture on an unchanged transcript.
	n, err = AppendSeatTurns(run, "a1", turns(5))
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 || countTurns(t, run) != 5 {
		t.Errorf("re-ingest wrote %d rows and the table holds %d — a re-run must add nothing", n, countTurns(t, run))
	}

	// The transcript grew: seven turns now, two of them new.
	n, err = AppendSeatTurns(run, "a1", turns(7))
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || countTurns(t, run) != 7 {
		t.Errorf("a grown transcript added %d rows to a table of %d, want 2 and 7", n, countTurns(t, run))
	}
}

// TURNS ARE PER AGENT. Two seats' turn 0 are different turns, and a key that collided would make
// the second seat's ingest silently vanish into IGNORE.
func TestSeatTurnsAreScopedToTheAgent(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "a1", turns(3)); err != nil {
		t.Fatal(err)
	}
	if _, err := AppendSeatTurns(run, "a2", turns(3)); err != nil {
		t.Fatal(err)
	}
	if got := countTurns(t, run); got != 6 {
		t.Errorf("two agents' three turns each gave %d rows, want 6 — the key is colliding across seats", got)
	}
}

// A ROW FILED AGAINST NO AGENT WOULD COUNT TOWARDS EVERY TOTAL AND BELONG TO NO SEAT.
func TestSeatTurnsRefuseAnEmptyAgent(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "", turns(3)); err == nil {
		t.Error("turns filed against an empty agent id were accepted")
	}
	if got := countTurns(t, run); got != 0 {
		t.Errorf("the refusal still wrote %d rows", got)
	}
}

func TestSeatTurnsOfNothingIsNotAnError(t *testing.T) {
	run := runFor(t)
	if n, err := AppendSeatTurns(run, "a1", nil); err != nil || n != 0 {
		t.Errorf("no turns gave (%d, %v), want (0, nil)", n, err)
	}
}

func runFor(t *testing.T) Run {
	t.Helper()
	dir := recordtest.TmpRun(t)
	run, err := NewRun(dir)
	if err != nil {
		t.Fatal(err)
	}
	return run
}

// queryOne is a scalar read for the view assertions below.
func queryOne(t *testing.T, run Run, q string, dest ...any) {
	t.Helper()
	db, err := openRun(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(q).Scan(dest...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

// AN UNSTAMPED TURN MUST NOT DRAG THE SEAT'S START TO THE EPOCH.
//
// The parser records 0 for a line that carried no timestamp, so a MIN() that let it in would put
// the seat's first turn in 1970 and report a span of fifty-six years. NULLIF is the contract
// honoured in SQL, and this is the assertion that says so.
func TestSeatMetricsIgnoresUnstampedTurnsInTheSpan(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "AG", []seatturn.Turn{
		{Index: 0, TSMillis: 1000, Model: "m", Output: 3, Thinking: true},
		{Index: 1, TSMillis: 5500, Model: "m", Output: 20},
		{Index: 2, TSMillis: 0, Model: "m", Output: 7}, // no timestamp on the line
	}); err != nil {
		t.Fatal(err)
	}
	var turns, wall, first int64
	queryOne(t, run, `SELECT turns, wall_ms, first_ts_ms FROM seat_metrics`, &turns, &wall, &first)
	if turns != 3 {
		t.Errorf("turns = %d, want 3 — an unstamped turn is still a turn", turns)
	}
	if first != 1000 {
		t.Errorf("first_ts_ms = %d, want 1000", first)
	}
	if wall != 4500 {
		t.Errorf("wall_ms = %d, want 4500 (5500-1000); an unstamped turn reached MIN()", wall)
	}
}

// A SEAT WITH NO TIMESTAMPED TURN IS NOT MEASURED, AND THAT IS NOT ZERO. Reporting 0 would put a
// seat that ran for twenty minutes into a throughput table as instantaneous.
func TestSeatMetricsSaysNotMeasuredRatherThanZero(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "AG", []seatturn.Turn{{Index: 0, Model: "m", Output: 5}}); err != nil {
		t.Fatal(err)
	}
	var wall *int64
	queryOne(t, run, `SELECT wall_ms FROM seat_metrics`, &wall)
	if wall != nil {
		t.Errorf("wall_ms = %d for a seat with no timestamped turn, want NULL", *wall)
	}
}

// TURNS FROM AN AGENT THAT NEVER REGISTERED STILL COUNT.
//
// The join to the seat is a LEFT join for this reason: an inner join would drop every row whose
// agent has no register event, and report a cheaper run than happened. A crashed seat is exactly
// the one whose cost you want to see.
func TestSeatMetricsKeepsTurnsFromAnUnregisteredAgent(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "NEVER-REGISTERED", []seatturn.Turn{
		{Index: 0, TSMillis: 1, Model: "m", Output: 11},
	}); err != nil {
		t.Fatal(err)
	}
	var out int64
	var seat *string
	queryOne(t, run, `SELECT output_tokens, seat_id FROM seat_metrics`, &out, &seat)
	if out != 11 {
		t.Errorf("output_tokens = %d, want 11 — the row was dropped by the join", out)
	}
	if seat != nil {
		t.Errorf("seat_id = %q for an agent that never registered, want NULL", *seat)
	}
}

// A SEAT'S FIRST TURN HAS NO PREDECESSOR, so its span is NULL and not 0. Zero would say
// "instant", and a bucket summing it would under-report every seat by its opening turn.
func TestSeatTurnSpanLeavesTheFirstTurnUnmeasured(t *testing.T) {
	run := runFor(t)
	if _, err := AppendSeatTurns(run, "AG", []seatturn.Turn{
		{Index: 0, TSMillis: 1000, Model: "m"},
		{Index: 1, TSMillis: 3000, Model: "m"},
	}); err != nil {
		t.Fatal(err)
	}
	var span *int64
	queryOne(t, run, `SELECT span_ms FROM seat_turn_span WHERE turn_idx = 0`, &span)
	if span != nil {
		t.Errorf("the first turn's span_ms = %d, want NULL", *span)
	}
	var second int64
	queryOne(t, run, `SELECT span_ms FROM seat_turn_span WHERE turn_idx = 1`, &second)
	if second != 2000 {
		t.Errorf("the second turn's span_ms = %d, want 2000", second)
	}
}
