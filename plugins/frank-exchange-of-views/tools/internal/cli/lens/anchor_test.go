package lens

import "testing"

func TestExtractQuotePrefersQuotedSpan(t *testing.T) {
	if got := extractQuote(`§ Foundations: "the scheduler is preemptive"`); got != "the scheduler is preemptive" {
		t.Errorf("extractQuote = %q, want the quoted span", got)
	}
	if got := extractQuote("  a bare location  "); got != "a bare location" {
		t.Errorf("extractQuote bare = %q", got)
	}
}

func TestLocateEndExactAndFlexible(t *testing.T) {
	report := "# H1\n\nThe scheduler is preemptive and fair.\n"
	// Exact substring.
	if end := locateEnd(report, "scheduler is preemptive"); end < 0 || report[:end] != "# H1\n\nThe scheduler is preemptive" {
		t.Errorf("exact locate wrong end=%d", end)
	}
	// Whitespace-flexible (reflowed spacing in the quote).
	if end := locateEnd(report, "scheduler   is\tpreemptive"); end < 0 {
		t.Error("flexible locate failed to match reflowed spacing")
	}
	// A mis-quote → -1 (reject).
	if end := locateEnd(report, "the scheduler is cooperative"); end != -1 {
		t.Errorf("a mis-quote must not locate, got %d", end)
	}
}

func TestInsertMarkerAtOffset(t *testing.T) {
	report := []byte("The sky is blue.")
	end := locateEnd(string(report), "sky is blue")
	got := string(insertMarker(report, end, "<!--fx:f-abc-->"))
	if got != "The sky is blue<!--fx:f-abc-->." {
		t.Errorf("insertMarker = %q", got)
	}
}

func TestInsideFenceGuardsCode(t *testing.T) {
	report := "prose here\n```go\ncode line\n```\nmore prose\n"
	codeAt := locateEnd(report, "code line")
	if codeAt < 0 || !insideFence(report, codeAt) {
		t.Errorf("a match inside a fence must report insideFence=true (at=%d)", codeAt)
	}
	proseAt := locateEnd(report, "more prose")
	if insideFence(report, proseAt) {
		t.Error("a prose match must report insideFence=false")
	}
}
