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
