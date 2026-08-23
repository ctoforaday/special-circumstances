// Package freshness turns a session's raw readings into the note's AGE, in three
// units that fail independently.
//
// Three measures, because one cannot see the others' failures: a session can burn
// 400k tokens in twelve turns of bulk reading, or take 300 turns without moving the
// token count, or land a day's commits with the note untouched. Any single measure
// reports one of those as fresh.
//
// NO PERCENTAGES, and that is a decision rather than a limitation to work around. A
// fraction-of-window needs a window, nothing in any payload or transcript carries one,
// and the only obtainable denominator — this session's own first compaction — exists
// on some sessions and not others. So the design does not compute one at all: the
// measures are absolute, and a reader who wants a ratio has to supply the denominator
// knowingly rather than receive a guess dressed as a measurement.
//
// Every measure is a tri-state, and that is the package's whole discipline. An
// unmeasured figure is NEVER rendered as zero: "no work since the note" and "I could
// not tell" are different claims, and a reader cannot distinguish them once one has
// been printed as the other.
//
// This package decides nothing. It does not emit, it does not choose a band, and it
// carries no threshold — Phase 1 collects the distribution those thresholds are
// supposed to come from, and a threshold picked before that distribution exists is a
// guess this design will not launder as a default.
package freshness

import (
	"fmt"
	"strings"
	"time"

	"github.com/ctoforaday/special-circumstances/plugins/prosthetic-conscience/tools/internal/ctxusage"
)

// State is what .claude/checkpoints/freshness.json holds. It exists because growth
// needs a reading nobody kept: "tokens now minus tokens when the note was written"
// requires the second term, and neither the transcript nor the note carries it. An
// agent cannot supply it either — a session can no more read its own context size
// than hash its own bytes.
type State struct {
	// WrittenAtSeen is the note's written_at as last observed, and the key for
	// deciding whether the reading below still belongs to this note.
	WrittenAtSeen time.Time `json:"written_at_seen"`

	TokensAtWrite  int `json:"tokens_at_write"`
	DroppedAtWrite int `json:"dropped_at_write"`
	// HasWriteReading separates "stamped at zero" from "never stamped". Without it a
	// note first seen before this shipped would report its whole context as growth.
	HasWriteReading bool `json:"has_write_reading"`

	// StampedAt is WHEN the reading above was taken, which is not when the note was
	// written and can be far from it.
	//
	// Nothing observes a note at the moment it is written: the only callers of Of are
	// the sealer and the compaction observer, which run at boundaries. A note written at
	// turn 10 and first sealed at turn 200 gets its baseline at turn 200, so "growth"
	// afterwards is growth since the SEAL. That is a real interval and a useful one, but
	// it is not the interval the field name promises, and a reader cannot tell the two
	// apart from the number alone. The row carries this timestamp so the gap is visible
	// rather than assumed away.
	StampedAt time.Time `json:"stamped_at"`
}

// Branch is the third measure, passed in rather than computed here: it costs a git
// subprocess (measured p50 33 ms, p95 170 ms on this repository), which is one to two
// orders of magnitude above the transcript budget, so the caller owns the cache and
// this package stays cheap enough to call on a tick.
type Branch struct {
	Commits int
	Known   bool
}

// Measures is the note's age in every unit that could be established.
type Measures struct {
	// Growth is measured from StampedAt, which is when the baseline reading was taken —
	// NOT from the note's written_at. GrowthSince carries that moment so a consumer can
	// see the gap instead of inferring an interval the number does not describe.
	Growth      int
	GrowthKnown bool
	GrowthSince time.Time

	Turns         int
	TurnsMeasured bool

	BranchCommits int
	BranchKnown   bool
}

// GaugeAfter is Gauge with the stamping flag from ObserveAndSay: when this call created
// the write-time reading, growth has no interval behind it and is reported unmeasured.
func GaugeAfter(st State, m ctxusage.Measure, b Branch, justStamped bool) Measures {
	out := Gauge(st, m, b)
	if justStamped {
		out.Growth, out.GrowthKnown = 0, false
	}
	return out
}

// Gauge composes the readings into the measures. Nothing here re-reads anything: it is
// arithmetic over values the caller already has, so it stays callable on a hot path.
func Gauge(st State, m ctxusage.Measure, b Branch) Measures {
	out := Measures{
		Turns:         m.Turns,
		TurnsMeasured: m.TurnsMeasured,
		BranchCommits: b.Commits,
		BranchKnown:   b.Known,
	}
	if !out.TurnsMeasured {
		out.Turns = 0
	}
	if !out.BranchKnown {
		out.BranchCommits = 0
	}

	// Growth through the CUMULATIVE dropped figure on both sides. The live counter
	// resets at a compaction — 1,001,875 to 12,823, measured — so the naive difference
	// goes negative across one and the stalest note in the file reads as the freshest.
	// Both terms are monotone, so their difference is too.
	if st.HasWriteReading || m.TokensKnown && comparable(st, m) {
		now := m.Tokens
		if m.DroppedKnown {
			now += m.Dropped
		}
		then := st.TokensAtWrite + st.DroppedAtWrite
		// Growth CANNOT be negative: both terms are monotone by construction. A negative
		// result is proof the two sides are not comparable, whatever the flags said, so
		// it is reported as unmeasured rather than rendered. Belt and braces with
		// comparable() below, because the two catch different mistakes — one a missing
		// term, the other an arithmetic that has gone wrong for a reason not foreseen.
		if g := now - then; g >= 0 {
			out.Growth, out.GrowthKnown, out.GrowthSince = g, true, st.StampedAt
		}
	}

	return out
}

// comparable reports whether the two sides of the growth subtraction are on the same
// footing.
//
// THE ASYMMETRY THIS EXISTS TO STOP: the current side adds its dropped figure only when
// known, while the stamped side always carries whatever was recorded. A note stamped
// just after a compaction — when cumulativeDroppedTokens was visible — and gauged later
// from a window that no longer holds a boundary would compute
// 13,000 − 1,000,000 = −987,000 and report it as measured. That is the stalest note
// reading as the freshest, which is the failure the design cites as its own reason.
//
// So: if the stamped side carries a dropped figure, the current side must too.
func comparable(st State, m ctxusage.Measure) bool {
	return st.DroppedAtWrite == 0 || m.DroppedKnown
}

// Observe stamps the write-time reading the first time a given note is seen, and
// leaves it alone afterwards.
//
// Once per note, not once per tick: re-stamping would move the reference point to the
// present on every call, and growth would report the interval since the last tick
// rather than since the note — a measure that is always near zero and always looks
// healthy.
func Observe(st State, writtenAt time.Time, m ctxusage.Measure, now time.Time) State {
	out, _ := ObserveAndSay(st, writtenAt, m, now)
	return out
}

// ObserveAndSay is Observe, and it also reports whether THIS call created the reading.
//
// The caller needs to know, because a gauge run in the same breath as the stamping
// would compare the current figure against itself and report growth 0 as MEASURED. For
// a note written long before its first seal that zero is manufactured, and it lands in
// criterion 1's distribution looking exactly like a genuinely fresh note.
func ObserveAndSay(st State, writtenAt time.Time, m ctxusage.Measure, now time.Time) (State, bool) {
	before := st
	st = observe(st, writtenAt, m, now)
	stamped := st.HasWriteReading && (!before.HasWriteReading || !before.WrittenAtSeen.Equal(st.WrittenAtSeen))
	return st, stamped
}

func observe(st State, writtenAt time.Time, m ctxusage.Measure, now time.Time) State {
	if writtenAt.IsZero() {
		return st
	}
	if st.HasWriteReading && st.WrittenAtSeen.Equal(writtenAt) {
		return st
	}
	if !m.TokensKnown {
		// Nothing to stamp. Recording zero would make the next growth reading the whole
		// context, which is a large confident number that is entirely wrong.
		st.WrittenAtSeen = writtenAt
		return st
	}
	st.WrittenAtSeen = writtenAt
	st.TokensAtWrite = m.Tokens
	st.DroppedAtWrite = m.Dropped
	st.HasWriteReading = true
	st.StampedAt = now
	return st
}

// Render is the one line, ≤ 200 bytes, that a session would see.
//
// It states measurements and stops. Spike §3b's surviving finding is that the hook
// adds no imperative of its own: what to do about an old note is the session's
// business, and an injected instruction was twice treated as a suspected prompt
// injection. An unmeasured figure is omitted entirely rather than shown as zero.
func Render(m Measures, notePath string) string {
	var parts []string
	if m.GrowthKnown {
		parts = append(parts, fmt.Sprintf("+%s tokens", thousands(m.Growth)))
	}
	if m.TurnsMeasured {
		parts = append(parts, fmt.Sprintf("%d turns", m.Turns))
	}
	if m.BranchKnown {
		parts = append(parts, fmt.Sprintf("%d commits", m.BranchCommits))
	}
	if len(parts) == 0 {
		return "checkpoint age not measurable this session — " + notePath
	}
	line := "checkpoint age: " + strings.Join(parts, ", ") + " — " + notePath
	if len(line) > 200 {
		line = line[:197] + "..."
	}
	return line
}

func thousands(n int) string {
	s := fmt.Sprintf("%d", n)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	var out []string
	for len(s) > 3 {
		out = append([]string{s[len(s)-3:]}, out...)
		s = s[:len(s)-3]
	}
	out = append([]string{s}, out...)
	joined := strings.Join(out, ",")
	if neg {
		return "-" + joined
	}
	return joined
}
