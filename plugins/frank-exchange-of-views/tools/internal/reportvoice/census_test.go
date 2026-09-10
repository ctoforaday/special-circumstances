package reportvoice

import "testing"

// PRESENCE AND CENSUS ARE TWO SHAPES, and the whole of #873 is that one call site had the wrong
// one. This pins both against the same text, because the bug is invisible unless you compare them:
// Find's answer is not WRONG, it is a different question, and the defect was asking it of a
// document.
func TestFindIsPresenceAndFindAllIsACensus(t *testing.T) {
	doc := `# On 91

[minority: lane-3] The factorisation is 7 x 13.
This run checked it three ways. [minority: lane-1] The debate also
considered Fermat witnesses. [minority: lane-3] See this report's appendix.`

	// Find: at most one per TELL. Two process-voice patterns and one lane pattern match, so 3 —
	// for a text carrying seven occurrences.
	if got := len(Find(doc)); got != 3 {
		t.Errorf("Find returned %d; presence is one per tell and three tells match here", got)
	}

	all := FindAll(doc)
	byClass := CountByClass(all)
	if byClass[LaneAttribution] != 3 {
		t.Errorf("lane-attribution census = %d, want 3", byClass[LaneAttribution])
	}
	// "This run", "The debate", "this report" — the two ProcessVoice patterns across three places.
	if byClass[ProcessVoice] != 3 {
		t.Errorf("process-voice census = %d, want 3", byClass[ProcessVoice])
	}
	if len(all) != 6 {
		t.Errorf("census total = %d, want 6", len(all))
	}
	// THE POINT OF THE CENSUS: strictly more than presence, on the same text.
	if len(all) <= len(Find(doc)) {
		t.Error("the census found no more than presence did, so it is not a census")
	}
}

// ORDERED BY POSITION, NOT GROUPED BY TELL. An author reads a document top to bottom; the grouping
// a summary wants is one line of code from this, while the ordering cannot be recovered once
// thrown away.
func TestFindAllIsOrderedByPosition(t *testing.T) {
	doc := "line one\n[minority: lane-2] two\nthis run three\n[lane-4] four\n"
	all := FindAll(doc)
	if len(all) < 3 {
		t.Fatalf("expected at least 3 occurrences, got %d", len(all))
	}
	for i := 1; i < len(all); i++ {
		if all[i].Offset < all[i-1].Offset {
			t.Fatalf("occurrence %d is at offset %d, behind its predecessor at %d — not ordered",
				i, all[i].Offset, all[i-1].Offset)
		}
	}
}

// THE LINE NUMBER IS THE PART AN AUTHOR USES. An offset alone sends them counting bytes.
func TestFindAllReportsTheLine(t *testing.T) {
	doc := "alpha\nbeta\n[minority: lane-1] gamma\ndelta\nthis run again\n"
	all := FindAll(doc)
	want := map[string]int{"[minority: lane-1]": 3, "this run": 5}
	seen := map[string]int{}
	for _, o := range all {
		seen[o.Match] = o.Line
	}
	for m, line := range want {
		if seen[m] != line {
			t.Errorf("%q reported at line %d, want %d (all: %+v)", m, seen[m], line, seen)
		}
	}
}

// Clean subject prose is silent under BOTH shapes. An advisory that fires on ordinary writing is
// noise, and noise is how a real note stops being read.
func TestCleanProseIsSilent(t *testing.T) {
	doc := "91 is composite: 7 x 13. Euclid's Elements, Book VII, Definition 11 defines a prime."
	if got := Find(doc); len(got) != 0 {
		t.Errorf("Find fired on clean prose: %+v", got)
	}
	if got := FindAll(doc); len(got) != 0 {
		t.Errorf("FindAll fired on clean prose: %+v", got)
	}
}

// Every occurrence carries its tell's Redirect — the "where it belongs instead" half. A census
// that says only WHERE without saying WHAT TO DO leaves an author deleting the sentence, which
// loses the epistemic half with the operational one.
func TestEveryOccurrenceCarriesItsRedirect(t *testing.T) {
	for _, o := range FindAll("this run is the debate [minority: lane-1]") {
		if o.Redirect == "" {
			t.Errorf("%q (%s) carries no redirect", o.Match, o.Class)
		}
	}
}
