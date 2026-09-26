package migrate_test

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/migrate"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatturn"
)

// A MIGRATION CARRIES THE PER-TURN MEASUREMENTS, and the loss it prevents is the silent kind.
//
// Turns are not events: they are measurements taken from outside, from transcripts that live in the
// client's project directory and are not part of a run. So an archived run holds the only copy, and a
// migration that dropped them would destroy it — while every downstream reader returned an empty
// decomposition, which is also what a run whose transcripts were never read returns. The failure and
// the healthy absence are the same bytes, which is why this is asserted rather than assumed.
//
// The source is built here rather than taken from the archive on purpose: no archived run carries turns
// (they all predate the ingest), so an archive-based test would pass on zero rows and prove nothing.
func TestMigrationCarriesThePerTurnMeasurements(t *testing.T) {
	fromDir := recordtest.TmpRun(t)
	src, err := record.NewRun(fromDir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][]seatturn.Turn{
		"agent-red-lens-logic": {
			{Index: 0, TSMillis: 1_000, Model: "claude-haiku", Input: 11, Output: 22, CacheRead: 33, CacheCreation: 44, Thinking: true, Tool: false},
			{Index: 1, TSMillis: 9_000, Model: "claude-haiku", Input: 55, Output: 66, CacheRead: 77, CacheCreation: 88, Thinking: false, Tool: true},
		},
		"agent-judge": {
			{Index: 0, TSMillis: 2_000, Model: "claude-sonnet", Input: 1, Output: 2, CacheRead: 3, CacheCreation: 4, Thinking: true, Tool: true},
		},
	}
	total := 0
	for agent, turns := range want {
		n, err := record.AppendSeatTurns(src, agent, turns)
		if err != nil {
			t.Fatal(err)
		}
		total += n
	}

	toDir := recordtest.TmpRun(t)
	m, err := migrate.Migrate(fromDir, toDir, migrate.Entries(), migrate.Options{})
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if m.SeatTurns != total {
		t.Errorf("the manifest reports %d per-turn measurement(s) carried, want %d — the count is on the "+
			"manifest so a dropped carry is a number that disagrees rather than an empty view downstream",
			m.SeatTurns, total)
	}

	dst := runtest.Open(t, toDir)
	if n, err := record.CountSeatTurns(dst); err != nil {
		t.Fatal(err)
	} else if n != total {
		t.Fatalf("the migrated record holds %d per-turn measurement(s), want %d", n, total)
	}

	// EVERY COLUMN, not just the count. A carry that wrote the right number of rows with the wrong
	// models or token figures reads as complete to a counter, and the columns are what the timing views
	// are made of.
	got, err := record.SeatTimingOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Measured {
		t.Error("the migrated record says it holds no per-turn measurement while holding them")
	}
	byAgent := map[string][]record.SeatTurnSpan{}
	for _, s := range got.Turns {
		byAgent[s.AgentID] = append(byAgent[s.AgentID], s)
	}
	for agent, turns := range want {
		spans := byAgent[agent]
		if len(spans) != len(turns) {
			t.Errorf("agent %s arrived with %d turn(s), want %d", agent, len(spans), len(turns))
			continue
		}
		for i, w := range turns {
			s := spans[i]
			if s.TurnIdx != w.Index || s.IsThinking != w.Thinking || s.IsTool != w.Tool || s.OutputTokens != int64(w.Output) {
				t.Errorf("agent %s turn %d arrived as %+v, want idx=%d thinking=%v tool=%v output=%d",
					agent, i, s, w.Index, w.Thinking, w.Tool, w.Output)
			}
		}
	}
	// The span is the gap to the turn before, so the carried timestamps must have survived: 9000-1000.
	if spans := byAgent["agent-red-lens-logic"]; len(spans) == 2 {
		if spans[1].SpanMillis == nil || *spans[1].SpanMillis != 8_000 {
			t.Errorf("the second turn's span came through as %v, want 8000 — a carry that dropped ts_ms "+
				"leaves every bucket at zero and the run reads as instantaneous", spans[1].SpanMillis)
		}
		if spans[0].SpanMillis != nil {
			t.Errorf("a seat's FIRST turn has no span, got %v", *spans[0].SpanMillis)
		}
	}
}

// AND THE STATED ZERO: a run whose transcripts were never read says so, in a field, on both surfaces.
// This is the arm that makes the one above mean something — without it, "measured" could be hardwired
// true and the carry test would still pass.
func TestAnUncapturedRunSaysItWasNotMeasured(t *testing.T) {
	m, dst := migrateArchive(t, "2026-09-11_is-91-prime-b9.tar.gz")
	if m.SeatTurns != 0 {
		t.Errorf("the archive carries no per-turn measurements, yet the manifest reports %d", m.SeatTurns)
	}
	got, err := record.SeatTimingOf(dst)
	if err != nil {
		t.Fatal(err)
	}
	if got.Measured {
		t.Error("a run holding no per-turn rows reports itself measured — the empty decomposition then " +
			"reads as a fast run rather than an absent measurement")
	}
	if len(got.Turns) != 0 || len(got.Buckets) != 0 {
		t.Errorf("an unmeasured run rendered %d turn(s) and %d bucket(s)", len(got.Turns), len(got.Buckets))
	}
}
