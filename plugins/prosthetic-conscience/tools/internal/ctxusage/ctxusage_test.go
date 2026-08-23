package ctxusage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func at(min int) time.Time { return time.Date(2026, 8, 23, 0, min, 0, 0, time.UTC) }

func assistant(min, in, cacheRead, cacheCreate int) string {
	return fmt.Sprintf(`{"type":"assistant","timestamp":%q,"message":{"usage":{"input_tokens":%d,"cache_read_input_tokens":%d,"cache_creation_input_tokens":%d,"output_tokens":7}}}`,
		at(min).Format(time.RFC3339), in, cacheRead, cacheCreate)
}

func boundary(min, pre, post, dropped int) string {
	return fmt.Sprintf(`{"type":"system","subtype":"compact_boundary","timestamp":%q,"compactMetadata":{"trigger":"auto","preTokens":%d,"postTokens":%d,"cumulativeDroppedTokens":%d}}`,
		at(min).Format(time.RFC3339), pre, post, dropped)
}

// filler pads a transcript so a later entry falls outside a bounded scan. It is a
// user entry, so it never contributes a turn or a token figure.
func filler(min, bytes int) string {
	return fmt.Sprintf(`{"type":"user","timestamp":%q,"pad":%q}`, at(min).Format(time.RFC3339), strings.Repeat("x", bytes))
}

func transcript(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// The numerator is exact and is a SUM: input plus both cache figures. Reading
// input_tokens alone reports 2 on a session holding a megabyte of context.
func TestTokensSumTheThreeInputFigures(t *testing.T) {
	p := transcript(t, assistant(1, 5, 100, 20), assistant(2, 2, 900, 30))
	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
	}
	if !m.TokensKnown || m.Tokens != 932 {
		t.Errorf("Tokens = %d (known=%v), want 932 from the LAST assistant entry", m.Tokens, m.TokensKnown)
	}
}

// THE CASE THIS PACKAGE EXISTS FOR. A 300-turn-old note is exactly the one whose
// earliest turns fall outside a bounded scan, so a truncated count would report the
// stalest notes as the freshest — and a small number is indistinguishable from an
// honestly small one.
func TestTurnsIsUnmeasuredWhenTheWindowDoesNotReachTheNote(t *testing.T) {
	var lines []string
	lines = append(lines, assistant(1, 1, 1, 1)) // long before the window
	for i := range 40 {
		lines = append(lines, filler(2+i, 40*1024)) // ~1.6 MB of padding
	}
	lines = append(lines, assistant(50, 1, 10, 1), assistant(51, 1, 10, 1))
	p := transcript(t, lines...)

	m, err := Read(p, at(1)) // the note was written at the very start
	if err != nil {
		t.Fatal(err)
	}
	if m.TurnsMeasured {
		t.Errorf("Turns reported as measured (%d) when the scan could not reach the note's write time; "+
			"a partial count here reports the stalest note as the freshest", m.Turns)
	}
	if m.Turns != 0 {
		t.Errorf("Turns = %d on an unmeasured read; it must not carry a partial count", m.Turns)
	}
	// The token figure is still exact — it comes from the newest entry, which is always in view.
	if !m.TokensKnown {
		t.Error("Tokens must remain measurable even when Turns is not")
	}
}

func TestTurnsCountsAssistantEntriesAfterTheNote(t *testing.T) {
	p := transcript(t,
		assistant(1, 1, 1, 1), assistant(2, 1, 1, 1), // before the note
		assistant(10, 1, 1, 1), assistant(11, 1, 1, 1), assistant(12, 1, 1, 1)) // after
	m, err := Read(p, at(9))
	if err != nil {
		t.Fatal(err)
	}
	if !m.TurnsMeasured {
		t.Fatal("a small in-window transcript must be measurable")
	}
	if m.Turns != 3 {
		t.Errorf("Turns = %d, want 3", m.Turns)
	}
}

// Ceiling is a TRI-STATE. A session that has never compacted has no trigger point to
// report, and defaulting it to a guess prints a confident percentage of a denominator
// nobody measured.
func TestCeilingIsUnknownUntilTheSessionHasCompacted(t *testing.T) {
	p := transcript(t, assistant(1, 1, 1, 1))
	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
	}
	if m.CeilingKnown {
		t.Errorf("CeilingKnown on a session with no compact_boundary (got %d)", m.Ceiling)
	}
	if m.DroppedKnown {
		t.Errorf("DroppedKnown with no boundary (got %d)", m.Dropped)
	}
}

func TestMostRecentBoundaryWins(t *testing.T) {
	p := transcript(t,
		boundary(5, 100_000, 12_000, 88_000),
		assistant(6, 1, 1, 1),
		boundary(7, 120_000, 14_000, 194_000),
		assistant(8, 2, 30, 4))
	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
	}
	if !m.CeilingKnown || m.Ceiling != 120_000 {
		t.Errorf("Ceiling = %d (known=%v), want 120000 from the most recent boundary", m.Ceiling, m.CeilingKnown)
	}
	if !m.DroppedKnown || m.Dropped != 194_000 {
		t.Errorf("Dropped = %d, want 194000 — growth is monotone only if this is the cumulative figure", m.Dropped)
	}
}

// A transcript is appended to while we read it, so the last line is routinely half
// written. That must not lose the whole read.
func TestATruncatedFinalLineIsToleratedNotFatal(t *testing.T) {
	p := transcript(t, assistant(1, 5, 100, 20))
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"type":"assistant","timestamp":"2026-08-23T00:`)
	f.Close()

	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
	}
	if !m.TokensKnown || m.Tokens != 125 {
		t.Errorf("Tokens = %d (known=%v), want 125 from the last COMPLETE entry", m.Tokens, m.TokensKnown)
	}
}

// No assistant entry anywhere is "I could not measure", never zero.
func TestNoAssistantEntryIsUnmeasuredNotZero(t *testing.T) {
	p := transcript(t, filler(1, 10), filler(2, 10))
	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
	}
	if m.TokensKnown {
		t.Errorf("TokensKnown with no assistant entry (got %d)", m.Tokens)
	}
}

// A hook must not fail over provenance: a missing transcript is an unmeasured read,
// not an error that takes the session's restore with it.
func TestAMissingTranscriptIsUnmeasuredNotAnError(t *testing.T) {
	m, err := Read(filepath.Join(t.TempDir(), "nope.jsonl"), at(0))
	if err != nil {
		t.Fatalf("missing transcript returned an error: %v", err)
	}
	if m.TokensKnown || m.TurnsMeasured || m.CeilingKnown {
		t.Errorf("missing transcript reported something as measured: %+v", m)
	}
}
