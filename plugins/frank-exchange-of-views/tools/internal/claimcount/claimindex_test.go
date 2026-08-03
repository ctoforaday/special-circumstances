package claimcount

import "testing"

func occCount(idx []LabelOccurrences) int {
	n := 0
	for _, l := range idx {
		n += len(l.Occurrences)
	}
	return n
}

func labelOf(idx []LabelOccurrences, label string) *LabelOccurrences {
	for i := range idx {
		if idx[i].Label == label {
			return &idx[i]
		}
	}
	return nil
}

// A cited claim asserted at N sites resolves to N occurrences — the completeness the
// propagation guarantee rests on. The label is the real c-<hex> citation id.
func TestIndexEnumeratesEverySite(t *testing.T) {
	md := "The cache evicts safely<!--cite:c-1-->.\n\nElsewhere we rely on that same safety<!--cite:c-1-->.\n\nAnd once more<!--cite:c-1-->.\n"
	idx := Index(md)
	l := labelOf(idx, "c-1")
	if l == nil {
		t.Fatalf("label c-1 missing from index: %+v", idx)
	}
	if len(l.Occurrences) != 3 {
		t.Errorf("c-1 at 3 sites resolved to %d occurrences, want 3", len(l.Occurrences))
	}
}

// The exclusions are the SAME as Count's: fenced code, footnote-definition lines, and
// headings never enter the index. A citation anchor in a heading or a fence is not a claim.
func TestIndexExcludesFencesDefsHeadings(t *testing.T) {
	md := "# A heading with an anchor<!--cite:c-h-->\n" +
		"A real claim<!--cite:c-1-->.\n" +
		"```\ncode with a fake anchor<!--cite:c-c-->\n```\n" +
		"[^c-1]: https://example.com  a footnote definition line, not a claim\n"
	idx := Index(md)
	if labelOf(idx, "c-h") != nil {
		t.Error("a heading's anchor leaked into the index")
	}
	if labelOf(idx, "c-c") != nil {
		t.Error("a fenced-code anchor leaked into the index")
	}
	l := labelOf(idx, "c-1")
	if l == nil || len(l.Occurrences) != 1 {
		t.Errorf("c-1 should have exactly 1 occurrence (the inline anchor): %+v", l)
	}
}

// The reconciliation caveat, both directions. Under per-sentence-unique authoring
// sum(occurrences) == Count. A segment carrying TWO distinct anchors makes the index
// exceed Count by exactly one — one site, two claims — and that is by design, not a bug.
func TestIndexReconcilesWithCount(t *testing.T) {
	// per-sentence-unique: three segments, one anchor each.
	unique := "First<!--cite:c-a-->.\nSecond<!--cite:c-b-->.\nThird<!--cite:c-c-->.\n"
	if got, want := occCount(Index(unique)), Count(unique); got != want {
		t.Errorf("per-sentence-unique: occurrences %d != Count %d (they must agree)", got, want)
	}

	// one segment, TWO distinct anchors: Count sees one claim-bearing segment; the index
	// sees two claims at that one site.
	twoLabel := "A single sentence citing two sources<!--cite:c-a--><!--cite:c-b-->.\n"
	if Count(twoLabel) != 1 {
		t.Fatalf("Count of a one-segment two-anchor line = %d, want 1", Count(twoLabel))
	}
	if got := occCount(Index(twoLabel)); got != 2 {
		t.Errorf("index of a two-anchor segment = %d occurrences, want 2 (exceeds Count by design)", got)
	}

	// the same anchor twice in one segment is ONE site, not two (distinct labels only).
	dupe := "Repeating the same anchor<!--cite:c-a--> twice<!--cite:c-a--> in one sentence.\n"
	if got := occCount(Index(dupe)); got != 1 {
		t.Errorf("an anchor repeated in one segment = %d occurrences, want 1 (one site)", got)
	}
}

// Each occurrence carries the nearest preceding heading, so a seat knows which section
// the site sits in.
func TestIndexTracksHeading(t *testing.T) {
	md := "# Foundations\nA claim under foundations<!--cite:c-1-->.\n## Risks\nA claim under risks<!--cite:c-2-->.\n"
	idx := Index(md)
	if l := labelOf(idx, "c-1"); l == nil || l.Occurrences[0].Heading != "Foundations" {
		t.Errorf("c-1 heading = %+v, want Foundations", l)
	}
	if l := labelOf(idx, "c-2"); l == nil || l.Occurrences[0].Heading != "Risks" {
		t.Errorf("c-2 heading = %+v, want Risks", l)
	}
}

// sentence_hash is stable under whitespace and line-position drift: the durable locator.
// The same claim text hashes identically regardless of surrounding whitespace, and a
// different claim hashes differently.
func TestSentenceHashStableUnderWhitespace(t *testing.T) {
	a := Index("the   cache\tevicts<!--cite:c-1-->.\n")                // irregular whitespace
	b := Index("prefix line.\n\n  the cache evicts<!--cite:c-1-->.\n") // same claim, different line + spacing
	if a[0].Occurrences[0].SentenceHash == "" {
		t.Fatal("empty sentence hash")
	}
	if a[0].Occurrences[0].SentenceHash != b[0].Occurrences[0].SentenceHash {
		t.Errorf("whitespace/position changed the hash: %q vs %q — it must be normalization-stable",
			a[0].Occurrences[0].SentenceHash, b[0].Occurrences[0].SentenceHash)
	}
	c := Index("a different claim<!--cite:c-1-->.\n")
	if a[0].Occurrences[0].SentenceHash == c[0].Occurrences[0].SentenceHash {
		t.Error("different claim text produced the same hash")
	}
}
