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

func plannerResponse(min, in, cacheRead, out int) string {
	return fmt.Sprintf(`{"step_index":%d,"source":"MODEL","type":"PLANNER_RESPONSE","created_at":%q,"input_tokens":%d,"output_tokens":%d,"cache_read_tokens":%d}`,
		min, at(min).Format(time.RFC3339), in, out, cacheRead)
}

func agyUser(min, bytes int) string {
	return fmt.Sprintf(`{"step_index":%d,"source":"USER_EXPLICIT","type":"USER_INPUT","created_at":%q,"content":%q}`,
		min, at(min).Format(time.RFC3339), strings.Repeat("x", bytes))
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

// The compaction boundary is read ONLY for its cumulative dropped figure, which growth
// needs to stay monotone. preTokens is deliberately not read: it was the sole available
// denominator for a percentage, and this design renders none.
func TestDroppedIsUnknownUntilTheSessionHasCompacted(t *testing.T) {
	p := transcript(t, assistant(1, 1, 1, 1))
	m, err := Read(p, at(0))
	if err != nil {
		t.Fatal(err)
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
	if m.TokensKnown || m.TurnsMeasured || m.DroppedKnown {
		t.Errorf("missing transcript reported something as measured: %+v", m)
	}
}

func TestAntigravityPlannerResponseTokensAndTurns(t *testing.T) {
	p := transcript(t,
		plannerResponse(1, 500, 1000, 50),  // before the note
		plannerResponse(2, 600, 1200, 60),  // before the note
		plannerResponse(10, 700, 2000, 70), // after the note
		plannerResponse(11, 800, 3000, 80), // after the note
		plannerResponse(12, 18420, 92100, 512))
	m, err := Read(p, at(9))
	if err != nil {
		t.Fatal(err)
	}
	if !m.TokensKnown || m.Tokens != 110520 {
		t.Errorf("Tokens = %d (known=%v), want 110520 (18420+92100) from the LAST PLANNER_RESPONSE entry", m.Tokens, m.TokensKnown)
	}
	if !m.TurnsMeasured || m.Turns != 3 {
		t.Errorf("Turns = %d (measured=%v), want 3", m.Turns, m.TurnsMeasured)
	}
}

func TestAntigravityPlannerResponseOutOfWindowStillReadsTokens(t *testing.T) {
	var lines []string
	lines = append(lines, plannerResponse(1, 100, 200, 10))
	for i := range 40 {
		lines = append(lines, agyUser(2+i, 40*1024))
	}
	lines = append(lines, plannerResponse(50, 1000, 5000, 100), plannerResponse(51, 2000, 8000, 200))
	p := transcript(t, lines...)

	m, err := Read(p, at(1))
	if err != nil {
		t.Fatal(err)
	}
	if m.TurnsMeasured || m.Turns != 0 {
		t.Errorf("Turns = %d (measured=%v), want unmeasured 0 when scan does not reach note", m.Turns, m.TurnsMeasured)
	}
	if !m.TokensKnown || m.Tokens != 10000 {
		t.Errorf("Tokens = %d (known=%v), want 10000 from newest PLANNER_RESPONSE", m.Tokens, m.TokensKnown)
	}
}
