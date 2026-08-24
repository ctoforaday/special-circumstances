package freshness

import (
	"os"
	"path/filepath"
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

// The render is injected into a session's context, so its budget is real: 200 bytes.
func TestRenderFitsTheBudget(t *testing.T) {
	m := Gauge(baseState(), ctxusage.Measure{
		Tokens: 987_654, TokensKnown: true, Turns: 412, TurnsMeasured: true,
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
// NO PERCENTAGE, EVER — not "none without a ceiling". A fraction needs a window,
// nothing carries one, and the only obtainable denominator exists on some sessions and
// not others. The design renders absolute figures and lets a reader supply a
// denominator knowingly rather than receive a guess dressed as a measurement.
func TestRenderNeverPrintsAPercentage(t *testing.T) {
	for _, m := range []Measures{
		Gauge(baseState(), ctxusage.Measure{Tokens: 140_000, TokensKnown: true}, Branch{}),
		Gauge(baseState(), ctxusage.Measure{
			Tokens: 990_000, TokensKnown: true, Dropped: 400_000, DroppedKnown: true,
		}, Branch{Commits: 9, Known: true}),
	} {
		if got := Render(m, "CHECKPOINT.md"); strings.Contains(got, "%") {
			t.Errorf("render printed a percentage:\n%s", got)
		}
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
	st = Observe(st, ts(10), ctxusage.Measure{Tokens: 100_000, TokensKnown: true}, ts(99))
	if !st.HasWriteReading || st.TokensAtWrite != 100_000 {
		t.Fatalf("first sight did not stamp: %+v", st)
	}
	// Same note, later in the session: the reading must NOT move.
	st = Observe(st, ts(10), ctxusage.Measure{Tokens: 400_000, TokensKnown: true}, ts(99))
	if st.TokensAtWrite != 100_000 {
		t.Errorf("TokensAtWrite moved to %d on an unchanged note; growth would always read ~0", st.TokensAtWrite)
	}
	// The note is rewritten: a new reference point, and a new reading.
	st = Observe(st, ts(50), ctxusage.Measure{Tokens: 400_000, TokensKnown: true}, ts(99))
	if st.TokensAtWrite != 400_000 || !st.WrittenAtSeen.Equal(ts(50)) {
		t.Errorf("a rewritten note did not re-stamp: %+v", st)
	}
}

// Without a token reading there is nothing to stamp, and stamping zero would make the
// next growth reading enormous and wrong.
func TestObserveDoesNotStampAnUnmeasuredReading(t *testing.T) {
	st := Observe(State{}, ts(10), ctxusage.Measure{}, ts(99))
	if st.HasWriteReading {
		t.Errorf("stamped a write reading from an unmeasured transcript: %+v", st)
	}
}

// THE MANUFACTURED ZERO. Of() stamps the write-time reading and gauges in the same
// call, so the FIRST seal after a note appears would compute growth against a reading
// taken moments earlier and report 0 as MEASURED.
//
// For a note written long before its first seal that zero is invented, and it enters
// criterion 1's growth distribution indistinguishable from a genuinely fresh note —
// the precise effect the "omitted when unmeasured" rule exists to prevent, arriving
// through the door that rule was built to guard.
func TestGrowthIsUnmeasuredOnTheObservationThatCreatedTheReading(t *testing.T) {
	var st State // nothing seen yet
	st, fresh := ObserveAndSay(st, ts(10), ctxusage.Measure{Tokens: 100_000, TokensKnown: true}, ts(99))
	if !fresh {
		t.Fatal("first sight of a note must report that it stamped")
	}
	m := GaugeAfter(st, ctxusage.Measure{Tokens: 100_000, TokensKnown: true}, Branch{}, fresh)
	if m.GrowthKnown {
		t.Errorf("growth reported as measured (%d) on the observation that created its own baseline", m.Growth)
	}

	// The NEXT boundary has a real interval behind it.
	st2, fresh2 := ObserveAndSay(st, ts(10), ctxusage.Measure{Tokens: 180_000, TokensKnown: true}, ts(99))
	if fresh2 {
		t.Fatal("an unchanged note must not re-stamp")
	}
	m2 := GaugeAfter(st2, ctxusage.Measure{Tokens: 180_000, TokensKnown: true}, Branch{}, fresh2)
	if !m2.GrowthKnown || m2.Growth != 80_000 {
		t.Errorf("Growth = %d (known=%v), want 80000 on the second observation", m2.Growth, m2.GrowthKnown)
	}
}

// THE ASYMMETRY. Growth adds the CURRENT dropped figure only when it is known, but
// always subtracts the stamped one — so a note stamped after a compaction (when
// cumulativeDroppedTokens was visible) and gauged later from a window that no longer
// contains a boundary computes 13,000 − 1,000,000 and reports −987,000 as MEASURED.
//
// That is "the stalest note reads as the freshest", which is the failure §II gives as
// the reason for the whole design, arriving through the arithmetic meant to prevent it.
// The original test could not see it: its baseline had DroppedAtWrite: 0.
func TestGrowthIsUnmeasuredWhenTheDroppedTermsAreNotComparable(t *testing.T) {
	stamped := State{
		WrittenAtSeen: ts(10), TokensAtWrite: 13_000,
		DroppedAtWrite: 1_000_000, HasWriteReading: true,
	}
	// Later read: no boundary in the window, so DroppedKnown is false.
	m := Gauge(stamped, ctxusage.Measure{Tokens: 400_000, TokensKnown: true}, Branch{})
	if m.GrowthKnown {
		t.Errorf("growth reported as measured (%d) from incomparable terms: the stamped side "+
			"carries a dropped figure and the current side does not", m.Growth)
	}
}

// Growth cannot be negative — both terms are monotone by construction. A negative result
// is proof the two sides are not comparable, whatever the flags say, so it is reported
// as unmeasured rather than rendered.
func TestNegativeGrowthIsNeverReported(t *testing.T) {
	st := State{WrittenAtSeen: ts(10), TokensAtWrite: 900_000, DroppedAtWrite: 500_000, HasWriteReading: true}
	m := Gauge(st, ctxusage.Measure{
		Tokens: 12_000, TokensKnown: true, Dropped: 100_000, DroppedKnown: true,
	}, Branch{})
	if m.GrowthKnown {
		t.Errorf("negative growth reported as measured: %d", m.Growth)
	}
}

// A note with no written_at — schema 2, or one written wrong — has no reference point,
// so it has no age. Every measure comes back unmeasured and NONE comes back zero.
//
// §V lists this as required Phase 1 coverage and nothing exercised it: Of's early
// return was reachable only through a real project directory, which no test built.
func TestANoteWithoutWrittenAtIsUnmeasuredNotAgeZero(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".claude", "checkpoints"), 0o755); err != nil {
		t.Fatal(err)
	}
	const schema2 = "---\nschema: 2\nupdated: 2026-08-23T00:10:00Z\n---\n## Validation loop\n1. x\n"
	m := Of(dir, "", schema2, Branch{}, ts(99))
	if m.GrowthKnown || m.TurnsMeasured || m.BranchKnown {
		t.Errorf("a note with no written_at reported something as measured: %+v", m)
	}
	if m.Growth != 0 || m.Turns != 0 || m.BranchCommits != 0 {
		t.Errorf("unmeasured figures carry numbers: %+v", m)
	}
	// And it must not have written a state file for a note it could not place in time.
	if _, err := os.Stat(filepath.Join(dir, ".claude", "checkpoints", "freshness.json")); err == nil {
		t.Error("state file written for a note with no reference point")
	}
}
