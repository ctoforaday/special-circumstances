package freshness

import (
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
)

func ts(min int) time.Time { return time.Date(2026, 8, 23, 0, min, 0, 0, time.UTC) }

// A note written at minute 10, seen when the session held 100k tokens.
func baseState() State {
	return State{
		WrittenAtSeen:   ts(10),
		TokensAtWrite:   100_000,
		DroppedAtWrite:  0,
		HasWriteReading: true,
	}
}

func TestGrowthIsTheDifferenceFromTheWriteTimeReading(t *testing.T) {
	m := Gauge(baseState(), ctxusage.Measure{Tokens: 140_000, TokensKnown: true}, Branch{})
	if !m.GrowthKnown {
		t.Fatal("growth unmeasured with both readings present")
	}
	if m.Growth != 40_000 {
		t.Errorf("Growth = %d, want 40000", m.Growth)
	}
}

// THE COMPACTION CASE. The raw counter RESETS at a boundary — measured 1,001,875 to
// 12,823 — so a naive subtraction goes negative and the stalest note in the file reads
// as the freshest. Growth must be computed through cumulativeDroppedTokens, which is
// monotone.
func TestGrowthStaysPositiveAcrossACompaction(t *testing.T) {
	st := baseState() // 100k held, nothing dropped yet
	// Since then: a compaction dropped 900k, and the live counter restarted at 13k.
	m := Gauge(st, ctxusage.Measure{
		Tokens: 13_000, TokensKnown: true,
		Dropped: 900_000, DroppedKnown: true,
	}, Branch{})

	if !m.GrowthKnown {
		t.Fatal("growth unmeasured across a compaction")
	}
	if m.Growth <= 0 {
		t.Errorf("Growth = %d across a compaction; a negative or zero figure makes the stalest "+
			"note read as the freshest", m.Growth)
	}
	if m.Growth != 813_000 {
		t.Errorf("Growth = %d, want 813000 ((13000+900000) - (100000+0))", m.Growth)
	}
}

// No write-time reading — a note first seen before this shipped — is UNMEASURED, never
// zero. Zero would read as "no work has happened since the note".
func TestGrowthWithoutAWriteTimeReadingIsUnmeasured(t *testing.T) {
	m := Gauge(State{WrittenAtSeen: ts(10)}, ctxusage.Measure{Tokens: 140_000, TokensKnown: true}, Branch{})
	if m.GrowthKnown {
		t.Errorf("GrowthKnown with no write-time reading (got %d)", m.Growth)
	}
	if m.Growth != 0 {
		t.Errorf("unmeasured growth carries a number: %d", m.Growth)
	}
}

// Proximity is the one figure that needs a denominator, and the denominator exists only
// after this session has compacted once. Everything else must survive without it.
func TestProximityOnlyWhenTheCeilingIsKnown(t *testing.T) {
	without := Gauge(baseState(), ctxusage.Measure{Tokens: 140_000, TokensKnown: true}, Branch{})
	if without.ProximityKnown {
		t.Errorf("ProximityKnown with no ceiling (got %v)", without.Proximity)
	}
	with := Gauge(baseState(), ctxusage.Measure{
		Tokens: 140_000, TokensKnown: true,
		Ceiling: 200_000, CeilingKnown: true,
	}, Branch{})
	if !with.ProximityKnown {
		t.Fatal("ceiling known but proximity is not")
	}
	if with.Proximity != 0.7 {
		t.Errorf("Proximity = %v, want 0.7", with.Proximity)
	}
}

// The render is injected into a session's context, so its budget is real: 200 bytes.
func TestRenderFitsTheBudget(t *testing.T) {
	m := Gauge(baseState(), ctxusage.Measure{
		Tokens: 987_654, TokensKnown: true, Turns: 412, TurnsMeasured: true,
		Ceiling: 1_000_000, CeilingKnown: true,
	}, Branch{Commits: 137, Known: true})
	got := Render(m, "research/some-fairly-long-slug/CHECKPOINT.md")
	if len(got) > 200 {
		t.Errorf("render is %d bytes, budget is 200:\n%s", len(got), got)
	}
	if strings.Contains(got, "\n") {
		t.Errorf("render must be ONE line:\n%s", got)
	}
}

// A percentage of a window nobody measured is a confident number with no denominator.
func TestRenderNeverPrintsAPercentageWithoutACeiling(t *testing.T) {
	m := Gauge(baseState(), ctxusage.Measure{Tokens: 140_000, TokensKnown: true}, Branch{})
	got := Render(m, "CHECKPOINT.md")
	if strings.Contains(got, "%") {
		t.Errorf("render printed a percentage with no ceiling measured:\n%s", got)
	}
}

// An unmeasured figure must not appear as a number. Silence about it is fine; a zero
// is a claim.
func TestRenderOmitsUnmeasuredFiguresRatherThanZeroingThem(t *testing.T) {
	m := Gauge(State{WrittenAtSeen: ts(10)}, ctxusage.Measure{}, Branch{})
	got := Render(m, "CHECKPOINT.md")
	for _, bad := range []string{"0 turns", "+0 tokens", "0 commits"} {
		if strings.Contains(got, bad) {
			t.Errorf("render shows %q for an unmeasured figure:\n%s", bad, got)
		}
	}
}

// Spike §3b's surviving finding: the hook adds no imperative of its own. The line states
// what was measured; deciding what to do about it is the session's business.
func TestRenderCarriesNoInstruction(t *testing.T) {
	m := Gauge(baseState(), ctxusage.Measure{Tokens: 140_000, TokensKnown: true, Turns: 30, TurnsMeasured: true}, Branch{})
	got := strings.ToLower(Render(m, "CHECKPOINT.md"))
	for _, imperative := range []string{"you must", "you should", "please", "write the", "update the", "consider"} {
		if strings.Contains(got, imperative) {
			t.Errorf("render carries an instruction (%q):\n%s", imperative, got)
		}
	}
}

// The write-time reading is stamped ONCE per note, when the gauge first sees a
// written_at it has not seen before. Re-stamping on every tick would make growth
// permanently zero — the measure would report the interval since the last tick.
func TestObserveStampsTheWriteReadingOncePerNote(t *testing.T) {
	var st State
	st = Observe(st, ts(10), ctxusage.Measure{Tokens: 100_000, TokensKnown: true})
	if !st.HasWriteReading || st.TokensAtWrite != 100_000 {
		t.Fatalf("first sight did not stamp: %+v", st)
	}
	// Same note, later in the session: the reading must NOT move.
	st = Observe(st, ts(10), ctxusage.Measure{Tokens: 400_000, TokensKnown: true})
	if st.TokensAtWrite != 100_000 {
		t.Errorf("TokensAtWrite moved to %d on an unchanged note; growth would always read ~0", st.TokensAtWrite)
	}
	// The note is rewritten: a new reference point, and a new reading.
	st = Observe(st, ts(50), ctxusage.Measure{Tokens: 400_000, TokensKnown: true})
	if st.TokensAtWrite != 400_000 || !st.WrittenAtSeen.Equal(ts(50)) {
		t.Errorf("a rewritten note did not re-stamp: %+v", st)
	}
}

// Without a token reading there is nothing to stamp, and stamping zero would make the
// next growth reading enormous and wrong.
func TestObserveDoesNotStampAnUnmeasuredReading(t *testing.T) {
	st := Observe(State{}, ts(10), ctxusage.Measure{})
	if st.HasWriteReading {
		t.Errorf("stamped a write reading from an unmeasured transcript: %+v", st)
	}
}
