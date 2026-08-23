// Package freshness turns a session's raw readings into the note's AGE, in three
// units that fail independently.
//
// Three measures, because one cannot see the others' failures: a session can burn
// 400k tokens in twelve turns of bulk reading, or take 300 turns without moving the
// token count, or land a day's commits with the note untouched. Any single measure
// reports one of those as fresh.
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
	Growth      int
	GrowthKnown bool

	Turns         int
	TurnsMeasured bool

	BranchCommits int
	BranchKnown   bool

	// Proximity is Tokens/Ceiling, and it is the one figure with a denominator this
	// design cannot always obtain: no payload and no transcript field carries the
	// context window, so the only ceiling is this session's own first compaction.
	Proximity      float64
	ProximityKnown bool
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
	if st.HasWriteReading && m.TokensKnown {
		now := m.Tokens
		if m.DroppedKnown {
			now += m.Dropped
		}
		then := st.TokensAtWrite + st.DroppedAtWrite
		out.Growth, out.GrowthKnown = now-then, true
	}

	if m.TokensKnown && m.CeilingKnown && m.Ceiling > 0 {
		out.Proximity, out.ProximityKnown = float64(m.Tokens)/float64(m.Ceiling), true
	}
	return out
}

// Observe stamps the write-time reading the first time a given note is seen, and
// leaves it alone afterwards.
//
// Once per note, not once per tick: re-stamping would move the reference point to the
// present on every call, and growth would report the interval since the last tick
// rather than since the note — a measure that is always near zero and always looks
// healthy.
func Observe(st State, writtenAt time.Time, m ctxusage.Measure) State {
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
	// Proximity rides only with its basis named, so the number cannot be read as a
	// fraction of some assumed window.
	if m.ProximityKnown {
		parts = append(parts, fmt.Sprintf("%.0f%% of this session's compaction point", m.Proximity*100))
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
