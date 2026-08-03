package lens

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// reFindingMarker mirrors claimcount.findingMarkerRe: the downstream parser is position-
// agnostic, so a test that re-extracts an inserted id proves the round-trip regardless of
// where in the sentence the marker landed.
var reFindingMarker = regexp.MustCompile(`<!--fx:(f-[0-9a-f]+)-->`)

// reAnyAnnotation strips the whole invisible layer (any fx marker + any footnote ref) so
// a test can compare the semantic prose before and after an insertion.
var reAnyAnnotation = regexp.MustCompile(`<!--fx:[^>]*-->|\[\^[^\]]*\]`)

func stripAnnotations(s string) string { return reAnyAnnotation.ReplaceAllString(s, "") }

// mustLocateInsert locates quote in report, asserts it was found, inserts marker at the
// returned offset (guarding against a fence like the caller does), and returns the new
// report plus the offset. Fails the test on a mis-locate or a fence hit.
func mustLocateInsert(t *testing.T, report, quote, marker string) (string, int) {
	t.Helper()
	end := locateEnd(report, quote)
	if end < 0 {
		t.Fatalf("locateEnd(%q) = -1, want a match", quote)
	}
	if insideFence(report, end) {
		t.Fatalf("locateEnd(%q) resolved inside a fence at %d", quote, end)
	}
	return string(insertMarker([]byte(report), end, marker)), end
}

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

// §V.1 + §V.2 — the annotation-interleaving matrix, seeded from the REAL 2026-08-02
// report.md sentences that surfaced #248. A verbatim-semantic quote (the prose, with the
// invisible markers/footnotes removed) must anchor even though the raw bytes are
// interrupted by "<!--fx:…-->" and "[^label]".
func TestLocateEndSkipsAnnotations(t *testing.T) {
	cases := []struct {
		name, report, quote string
	}{
		// Real line 5: an fx marker between a word and its colon, and another at sentence end.
		{
			"fx-between-word-and-colon (real)",
			"Acceptability depends on context<!--fx:f-762c1674-->: approximate financial models tolerate floating-point; transaction settlement does not.<!--fx:f-2462cfb8-->",
			"Acceptability depends on context: approximate financial models tolerate floating-point; transaction settlement does not.",
		},
		// Real line 16: a footnote ref before the terminal period; internal punctuation
		// (parens, commas, "e.g.", decimals) must be preserved as CONTENT and match.
		{
			"footnote-before-period, internal punctuation (real)",
			"Binary floating-point cannot exactly represent most decimal fractional values (e.g., 0.1, 0.2, 0.78) because decimal fractions require infinitely repeating binary expansions[^TechIncompat].",
			"Binary floating-point cannot exactly represent most decimal fractional values (e.g., 0.1, 0.2, 0.78) because decimal fractions require infinitely repeating binary expansions",
		},
		// Real line 25: an em-dash aside plus a trailing footnote AND fx marker.
		{
			"em-dash + trailing footnote + fx (real)",
			"Every major production payment system surveyed—including Modern Treasury, Stripe, Square, PayPal—uses integer minor units (cents) or decimal types, never binary floating-point[^NoProductionFloats].<!--fx:f-96ae5702-->",
			"Every major production payment system surveyed—including Modern Treasury, Stripe, Square, PayPal—uses integer minor units (cents) or decimal types, never binary floating-point",
		},
		{
			"marker between two words",
			"The scheduler is<!--fx:f-11-->preemptive and fair.",
			"The scheduler ispreemptive and fair",
		},
		{
			"footnote between a word and its punctuation (possible[^L2].)",
			"Deterministic rounding makes it possible[^L2].",
			"Deterministic rounding makes it possible",
		},
		{
			"several markers/footnotes in one sentence",
			"Floats[^a] are<!--fx:f-1--> unsafe[^b] for money<!--fx:f-2--> here.",
			"Floats are unsafe for money here",
		},
		{
			"marker mid-token (splices a word)",
			"Reconcili<!--fx:f-9-->ation catches the drift.",
			"Reconciliation catches the drift",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			end := locateEnd(c.report, c.quote)
			if end < 0 {
				t.Fatalf("locateEnd returned -1; quote should anchor through the annotation layer")
			}
			// The offset lands just past the last matched CONTENT char. It may sit right
			// BEFORE a trailing annotation (a legal boundary — a new marker inserts there),
			// but must never split one. Inserting there and re-stripping must recover prose
			// byte-identical to the report minus its annotation layer.
			out := string(insertMarker([]byte(c.report), end, "<!--fx:f-probe0-->"))
			if stripAnnotations(out) != stripAnnotations(c.report) {
				t.Errorf("insert at %d altered the semantic prose:\n got  %q\n want %q",
					end, stripAnnotations(out), stripAnnotations(c.report))
			}
		})
	}
}

// §V.3 — formatting tolerance: reflowed whitespace, a newline within the sentence,
// leading/trailing spaces, and TRAILING-punctuation variance (both directions).
func TestLocateEndFormatting(t *testing.T) {
	report := "The scheduler is preemptive\nand fair.\n"
	if locateEnd(report, "scheduler   is\tpreemptive and fair") < 0 {
		t.Error("reflowed whitespace + newline-within-sentence should locate")
	}
	if locateEnd(report, "  scheduler is preemptive and fair  ") < 0 {
		t.Error("leading/trailing spaces should be trimmed and locate")
	}
	// Quote omits the trailing period the report has.
	if locateEnd("It settles cleanly.", "It settles cleanly") < 0 {
		t.Error("quote without the report's trailing period should locate")
	}
	// Quote carries a trailing period the report lacks.
	if locateEnd("It settles cleanly", "It settles cleanly.") < 0 {
		t.Error("quote with an extra trailing period should locate (trailing punct trimmed)")
	}
}

// §V.4 — guardrails: no false match. Absent, near-variant, empty, INTERNAL-punctuation
// discrimination, and the blank-line boundary.
func TestLocateEndGuardrails(t *testing.T) {
	report := "The scheduler is preemptive and fair.\n"
	if locateEnd(report, "the scheduler is cooperative") != -1 {
		t.Error("a genuinely absent quote must return -1")
	}
	if locateEnd(report, "The scheduler is preemptive and slow") != -1 {
		t.Error("a one-word-changed near-variant must return -1")
	}
	if locateEnd(report, "") != -1 || locateEnd(report, "  .,;  ") != -1 {
		t.Error("an empty / all-trailing-punctuation quote must return -1 (no match at offset 0)")
	}

	// INTERNAL-punctuation discrimination: the report holds BOTH sentences; internal
	// punctuation is content, so they do not normalize-equal. A quote of one anchors THAT
	// one and never the other. (If internal punctuation were a separator, "costs rise
	// sharply" would false-match the "costs rise, sharply" occurrence.)
	twin := "Analysts warn costs rise, sharply as volume grows. Engineers note costs rise sharply as volume grows.\n"
	endComma := locateEnd(twin, "costs rise, sharply as volume grows")
	endPlain := locateEnd(twin, "costs rise sharply as volume grows")
	if endComma < 0 || endPlain < 0 {
		t.Fatalf("both distinct sentences must locate (comma=%d plain=%d)", endComma, endPlain)
	}
	if endComma == endPlain {
		t.Error("internal-punctuation-distinct sentences must NOT resolve to the same offset")
	}
	// The comma variant appears FIRST, so if internal punctuation collapsed, the plain
	// quote would false-match it (earlier offset). It must instead find the LATER plain one.
	if endPlain <= endComma {
		t.Errorf("the plain quote false-matched the comma sentence (plain=%d <= comma=%d)", endPlain, endComma)
	}

	// Blank-line boundary: a quote must not match across a paragraph break.
	para := "Costs rise as volume grows.\n\nDemand also grows over time.\n"
	if locateEnd(para, "Costs rise as volume grows. Demand also grows over time") != -1 {
		t.Error("a quote spanning a paragraph break must not match across it")
	}
}

// §V.5 — offset correctness: after inserting at the returned offset, the marker is
// present, re-parses, and never lands inside an annotation or a fence; the semantic prose
// is unchanged apart from the added marker.
func TestLocateEndOffsetInsertRoundTrips(t *testing.T) {
	report := "Floats[^a] are<!--fx:f-1--> unsafe for money here."
	quote := "Floats are unsafe for money here"
	out, _ := mustLocateInsert(t, report, quote, "<!--fx:f-abc123-->")
	ids := reFindingMarker.FindAllStringSubmatch(out, -1)
	found := false
	for _, m := range ids {
		if m[1] == "f-abc123" {
			found = true
		}
	}
	if !found {
		t.Errorf("inserted marker did not re-parse out of %q", out)
	}
	// The insertion must change ONLY the annotation layer — the semantic prose is intact.
	if got, want := stripAnnotations(out), stripAnnotations(report); got != want {
		t.Errorf("semantic prose changed:\n got  %q\n want %q", got, want)
	}
}

// §V.6 — multiple occurrences: first-match wins (documented behavior).
func TestLocateEndFirstMatch(t *testing.T) {
	report := "the claim holds. Later, the claim holds again.\n"
	end := locateEnd(report, "the claim holds")
	if end < 0 || end != strings.Index(report, "the claim holds")+len("the claim holds") {
		t.Errorf("first-match expected, got end=%d", end)
	}
}

// §V.9 — live re-check against the REAL artifact if present in the working tree. The
// research corpus is untracked, so a clean/CI checkout skips (the seeded real-sentence
// cases in TestLocateEndSkipsAnnotations carry the invariant unconditionally).
func TestLocateEndAgainstRealArtifact(t *testing.T) {
	// package dir plugins/frank-exchange-of-views/tools/internal/cli/lens → repo root is 6 up.
	path := filepath.Join("..", "..", "..", "..", "..", "..",
		"research", "2026-08-02_finding-markers-validation-2", "blue", "report.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("real artifact absent (%v) — seeded cases cover the invariant", err)
	}
	report := string(data)
	// Drive verbatim-semantic quotes of real sentences (annotations manually stripped).
	quotes := []string{
		"Acceptability depends on context",
		"Binary floating-point cannot exactly represent most decimal fractional values",
	}
	for _, q := range quotes {
		if locateEnd(report, q) < 0 {
			t.Errorf("real-artifact quote failed to anchor: %q", q)
		}
	}
}
