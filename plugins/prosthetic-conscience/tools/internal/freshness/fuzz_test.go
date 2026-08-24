package freshness

import (
	"strings"
	"testing"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
)

// The invariants below are UNIVERSAL claims — "growth is never negative when reported",
// "the render never contains a percentage" — and every test until now asserted them on
// inputs I chose. That is the weakness the example-based suite cannot fix: I pick the
// cases I already believe in, and the case that breaks the claim is by definition one I
// did not think of.
//
// Fuzzing asks the same questions of inputs nobody chose.

// GROWTH IS NEVER NEGATIVE WHEN REPORTED. Both terms are monotone by construction, so a
// negative result means the two sides were not comparable — and a negative growth renders
// as "+-987,000 tokens", making the stalest note in the file read as the freshest. That
// is §II's stated reason for the whole design, so it is the invariant most worth
// attacking with values I did not choose.
func FuzzGrowthIsNeverNegativeWhenReported(f *testing.F) {
	f.Add(100_000, 0, 140_000, 0, true, true)
	f.Add(13_000, 1_000_000, 400_000, 0, true, false)    // the asymmetric case
	f.Add(900_000, 500_000, 12_000, 100_000, true, true) // a compaction between
	f.Add(0, 0, 0, 0, false, false)

	f.Fuzz(func(t *testing.T, atWrite, droppedAtWrite, now, droppedNow int,
		hasWriteReading, droppedKnown bool) {

		st := State{
			TokensAtWrite:   atWrite,
			DroppedAtWrite:  droppedAtWrite,
			HasWriteReading: hasWriteReading,
			WrittenAtSeen:   time.Unix(1, 0),
		}
		m := ctxusage.Measure{
			Tokens: now, TokensKnown: true,
			Dropped: droppedNow, DroppedKnown: droppedKnown,
		}
		got := Gauge(st, m, Branch{})

		if got.GrowthKnown && got.Growth < 0 {
			t.Errorf("reported growth %d as MEASURED from state=%+v measure=%+v; "+
				"a negative figure renders as the freshest note in the file", got.Growth, st, m)
		}
		// And the other half of the tri-state contract: an unmeasured figure carries no
		// number at all, so a reader cannot mistake it for an answer.
		if !got.GrowthKnown && got.Growth != 0 {
			t.Errorf("unmeasured growth carries %d", got.Growth)
		}
	})
}

// THE RENDER'S CONTRACT, against measures nobody chose: one line, at most 200 bytes,
// never a percentage.
//
// The byte budget is real — it is injected into a session's context — and the percentage
// ban is absolute rather than conditional: a fraction needs a window, nothing carries
// one, so there is no input for which printing one is correct.
func FuzzRenderHonoursItsContract(f *testing.F) {
	f.Add(40_000, 12, 3, true, true, true, "CHECKPOINT.md")
	f.Add(0, 0, 0, false, false, false, "")
	f.Add(-5, -5, -5, true, true, true, strings.Repeat("deep/", 40)+"CHECKPOINT.md")

	f.Fuzz(func(t *testing.T, growth, turns, commits int,
		growthKnown, turnsMeasured, branchKnown bool, path string) {

		m := Measures{
			Growth: growth, GrowthKnown: growthKnown,
			Turns: turns, TurnsMeasured: turnsMeasured,
			BranchCommits: commits, BranchKnown: branchKnown,
		}
		got := Render(m, path)

		if len(got) > 200 {
			t.Errorf("render is %d bytes for %+v path=%q:\n%s", len(got), m, path, got)
		}
		if strings.Contains(got, "\n") {
			t.Errorf("render is not one line for %+v:\n%q", m, got)
		}
		if strings.Contains(got, "%") {
			t.Errorf("render printed a percentage for %+v:\n%s", m, got)
		}
		// An unmeasured figure must not appear as a number. The rendered line may say
		// nothing about it; it may not say zero.
		if !turnsMeasured && strings.Contains(got, " turns") {
			t.Errorf("render mentions turns while unmeasured:\n%s", got)
		}
		if !growthKnown && strings.Contains(got, " tokens") {
			t.Errorf("render mentions tokens while growth is unmeasured:\n%s", got)
		}
		if !branchKnown && strings.Contains(got, " commits") {
			t.Errorf("render mentions commits while unmeasured:\n%s", got)
		}
	})
}

// Observe must be IDEMPOTENT for a given note. Re-stamping on every call would move the
// reference point to the present, and growth would then report the interval since the
// last tick — a measure that is always near zero and always looks healthy.
//
// The property holds for any sequence of readings, which is what makes it worth fuzzing:
// the example test uses two calls, and the failure mode is about the third and fortieth.
func FuzzObserveIsIdempotentForOneNote(f *testing.F) {
	f.Add(100_000, 400_000, 250_000)
	f.Add(0, 0, 0)

	f.Fuzz(func(t *testing.T, first, second, third int) {
		note := time.Unix(1000, 0)
		var st State
		st = Observe(st, note, ctxusage.Measure{Tokens: first, TokensKnown: true}, time.Unix(2000, 0))
		stamped := st.TokensAtWrite
		hadReading := st.HasWriteReading

		for _, later := range []int{second, third} {
			st = Observe(st, note, ctxusage.Measure{Tokens: later, TokensKnown: true}, time.Unix(3000, 0))
			if st.HasWriteReading != hadReading {
				t.Errorf("HasWriteReading flipped on an unchanged note: %v -> %v", hadReading, st.HasWriteReading)
			}
			if st.TokensAtWrite != stamped {
				t.Errorf("TokensAtWrite moved from %d to %d on an UNCHANGED note; growth would "+
					"report the interval since the last tick", stamped, st.TokensAtWrite)
			}
		}
	})
}
