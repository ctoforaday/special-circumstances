package anchortext

import (
	"errors"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/anchor"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/repotree"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// reFindingMarker mirrors the anchor kinds table's finding token: the downstream parser is position-
// agnostic, so a test that re-extracts an inserted id proves the round-trip regardless of
// where in the sentence the marker landed.
var reFindingMarker = regexp.MustCompile(`<!--fx:(` + anchor.IDPattern("finding") + `)-->`)

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
	if InBlock(report, end, anchor.Fence) {
		t.Fatalf("locateEnd(%q) resolved inside a fence at %d", quote, end)
	}
	return string(insertMarker([]byte(report), end, marker)), end
}

func TestLocateSpanReturnsRawSpan(t *testing.T) {
	// A marker sits AFTER "time" (trailing), a footnote after "grows".
	report := "The cost is rising over time<!--fx:F-00000001--> now. Volume grows[^v] fast."
	start, end := LocateSpan(report, "rising over time")
	if start < 0 {
		t.Fatal("LocateSpan failed to match")
	}
	if report[start:end] != "rising over time" {
		t.Errorf("span = %q, want the raw matched text", report[start:end])
	}
	// End stops before the trailing marker (a content boundary), so the marker is OUT of span.
	if strings.Contains(report[start:end], "<!--fx:") {
		t.Errorf("span wrongly includes a trailing marker: %q", report[start:end])
	}
	// A span that INTERNALLY crosses a footnote includes it (caller rejects such an edit).
	s2, e2 := LocateSpan(report, "Volume grows fast")
	if s2 < 0 || !strings.Contains(report[s2:e2], "[^v]") {
		t.Errorf("marker-crossing span should include the footnote: %q", report[s2:e2])
	}
	// Absent → (-1,-1).
	if s, e := LocateSpan(report, "not present here"); s != -1 || e != -1 {
		t.Errorf("absent quote → (%d,%d), want (-1,-1)", s, e)
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
	got := string(insertMarker(report, end, "<!--fx:F-00000abc-->"))
	if got != "The sky is blue<!--fx:F-00000abc-->." {
		t.Errorf("insertMarker = %q", got)
	}
}

func TestInsideFenceGuardsCode(t *testing.T) {
	report := "prose here\n```go\ncode line\n```\nmore prose\n"
	codeAt := locateEnd(report, "code line")
	if codeAt < 0 || !InBlock(report, codeAt, anchor.Fence) {
		t.Errorf("a match inside a fence must report InBlock=true (at=%d)", codeAt)
	}
	proseAt := locateEnd(report, "more prose")
	if InBlock(report, proseAt, anchor.Fence) {
		t.Error("a prose match must report InBlock=false")
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
			"Acceptability depends on context<!--fx:F-762c1674-->: approximate financial models tolerate floating-point; transaction settlement does not.<!--fx:F-2462cfb8-->",
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
			"Every major production payment system surveyed—including Modern Treasury, Stripe, Square, PayPal—uses integer minor units (cents) or decimal types, never binary floating-point[^NoProductionFloats].<!--fx:F-96ae5702-->",
			"Every major production payment system surveyed—including Modern Treasury, Stripe, Square, PayPal—uses integer minor units (cents) or decimal types, never binary floating-point",
		},
		{
			"marker between two words",
			"The scheduler is<!--fx:F-00000011-->preemptive and fair.",
			"The scheduler ispreemptive and fair",
		},
		{
			"footnote between a word and its punctuation (possible[^L2].)",
			"Deterministic rounding makes it possible[^L2].",
			"Deterministic rounding makes it possible",
		},
		{
			"several markers/footnotes in one sentence",
			"Floats[^a] are<!--fx:F-00000001--> unsafe[^b] for money<!--fx:F-00000002--> here.",
			"Floats are unsafe for money here",
		},
		{
			"marker mid-token (splices a word)",
			"Reconcili<!--fx:F-00000009-->ation catches the drift.",
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
	report := "Floats[^a] are<!--fx:F-00000001--> unsafe for money here."
	quote := "Floats are unsafe for money here"
	out, _ := mustLocateInsert(t, report, quote, "<!--fx:F-00abc123-->")
	ids := reFindingMarker.FindAllStringSubmatch(out, -1)
	found := false
	for _, m := range ids {
		if m[1] == "F-00abc123" {
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
	root, err := repotree.Root()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "research", "2026-08-02_finding-markers-validation-2", "blue", "report.md")
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

// ATTACH PLACES BY InsertAnchor's RULE AND BUILDS THE TOKEN FROM THE TABLE: the id's token lands
// at the quote's end, before its trailing punctuation, and a refusal is InsertAnchor's sentinel,
// unmapped, so each placing verb keeps its own message.
func TestAttachPlacesTheTablesTokenAtTheQuotesEnd(t *testing.T) {
	doc := "# H\n\nWater is wet. The sky is blue.\n"
	got, err := Attach(doc, "C-00001a2b", "The sky is blue.")
	if err != nil {
		t.Fatal(err)
	}
	if want := "# H\n\nWater is wet. The sky is blue<!--cite:C-00001a2b-->.\n"; got != want {
		t.Errorf("Attach = %q, want %q", got, want)
	}
	if _, err := Attach(doc, "F-00001a2b", "not in the report"); !errors.Is(err, ErrMisQuote) {
		t.Errorf("a mis-quote = %v, want ErrMisQuote", err)
	}
	if _, err := Attach("```\ncode here\n```\n", "P-00001a2b", "code here"); !errors.Is(err, ErrInFence) {
		t.Errorf("a quote in a fence = %v, want ErrInFence", err)
	}
}

// NO ANCHOR SITS ON A HEADING. Attach refuses a quote whose one occurrence ends in a heading, as
// anchor.Blocks reads one, whatever the heading's level and whatever follows it; a one-line bold
// paragraph is a paragraph, and so is the line under a heading. Replay's InsertAnchor reproduces an
// archived placement on a heading rather than re-authorising it.
func TestAttachRefusesAHeading(t *testing.T) {
	const doc = "# Title\n\n## Method\nThe sieve runs once.\n\n**Is 91 prime?**\n\n###### Deep\n"
	for _, c := range []struct {
		quote string
		err   error
	}{
		{"Title", ErrOnHeading},
		{"Method", ErrOnHeading},
		{"Deep", ErrOnHeading},
		{"The sieve runs once.", nil},
		{"Is 91 prime?", nil},
	} {
		if _, err := Attach(doc, "F-00001a2b", c.quote); !errors.Is(err, c.err) {
			t.Errorf("Attach(%q) = %v, want %v", c.quote, err, c.err)
		}
	}
	if _, err := InsertAnchor([]byte(doc), "Method", anchor.Token("F-00001a2b")); err != nil {
		t.Errorf("replay refused an archived placement on a heading: %v", err)
	}
}

// A FENCE'S OWN LINES ARE THE FENCE. InBlock reads anchor.Blocks, where a fence runs from its
// opener line through its closer line, so a quote ending on either line is refused: a marker there
// would ship inside the code block's delimiters. The line after the closer is prose again.
func TestInsideFenceReadsTheBlockReader(t *testing.T) {
	const report = "Intro.\n```text\ncode here\n```\nAfter the fence.\n"
	for _, c := range []struct {
		quote string
		err   error
	}{
		{"```text", ErrInFence},
		{"code here\n```", ErrInFence},
		{"After the fence", nil},
	} {
		if _, err := InsertAnchor([]byte(report), c.quote, "<!--cite:C-00000001-->"); !errors.Is(err, c.err) {
			t.Errorf("InsertAnchor(%q) = %v, want %v", c.quote, err, c.err)
		}
	}
}

// LocateOnce REFUSES IN ONE ORDER. A second occurrence is named before the first one's blank line,
// so a quote that also stands inside one paragraph is told it repeats; an accepted span ends before
// the anchor layer and the trimmed terminator, where a token goes.
func TestLocateOnceRefusesInItsOrder(t *testing.T) {
	for _, c := range []struct {
		doc, quote string
		scope      SpanScope
		want       error
		end        int
	}{
		{"One two.", "three", StopAtParagraph, ErrMisQuote, -1},
		{"Costs rose sharply. Costs rose.", "Costs rose.", StopAtParagraph, ErrAmbiguous, -1},
		{"Costs\n\nrose. Costs rose.", "Costs rose", StopAtParagraph, ErrAmbiguous, -1},
		{"Costs\n\nrose.", "Costs rose", StopAtParagraph, ErrCrossesParagraph, -1},
		{"Costs\n\nrose.", "Costs rose", CrossParagraphs, nil, 11},
		{"Costs rosed.", "Costs rose", StopAtParagraph, ErrSplitsWord, -1},
		{"Costs rose<!--fx:F-00000001-->.", "Costs rose.", StopAtParagraph, nil, 10},
	} {
		if _, end, err := LocateOnce(c.doc, c.quote, c.scope); err != c.want || end != c.end {
			t.Errorf("LocateOnce(%q, %q) = %d, %v; want %d, %v", c.doc, c.quote, end, err, c.end, c.want)
		}
	}
}
