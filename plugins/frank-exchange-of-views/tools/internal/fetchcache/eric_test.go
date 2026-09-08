package fetchcache

import "testing"

// A DOI IS NOT AN ERIC QUERY, and this is measured rather than reasoned. ERIC exposes no DOI
// field, so a DOI can only go in as free text — and free text matches a DOI CITED INSIDE another
// record. Asked for the Savage column's DOI the index returned a 2022 calculus-textbook paper
// whose abstract cites a different 10.5951 DOI, and the backend would have presented that as the
// source. A confident wrong paper is worse here than no answer: the seat cannot tell.
func TestADOIIsNotTurnedIntoAnEricFreeTextSearch(t *testing.T) {
	if got := ericQueryOf("https://doi.org/10.5951/MT.82.1.0035"); got != "" {
		t.Errorf("a DOI became an ERIC query %q — free text is not identity", got)
	}
}

// AN ACCESSION NUMBER IS IDENTITY, and it is the one key this corpus can be asked by exactly.
func TestAnEricAccessionNumberIsQueriedByID(t *testing.T) {
	for _, u := range []string{"https://eric.ed.gov/?id=EJ387875", "https://files.eric.ed.gov/fulltext/ED313351.pdf"} {
		got := ericQueryOf(u)
		if got != "id:EJ387875" && got != "id:ED313351" {
			t.Errorf("ericQueryOf(%q) = %q, want an id: query", u, got)
		}
	}
}

// A BARE HOSTNAME IS NOT A TITLE. A query built from one matches the corpus rather than the
// source, and the first hit would come back as though it were the document.
func TestATooThinURLProducesNoQuery(t *testing.T) {
	for _, u := range []string{"https://example.org/", "https://example.org/paper", "not a url at all"} {
		if got := ericQueryOf(u); got != "" {
			t.Errorf("ericQueryOf(%q) = %q, want no query", u, got)
		}
	}
}

// A TITLE-SHAPED PATH IS ALLOWED, because a landing-page url is what a seat usually holds.
func TestATitleShapedPathBecomesATitleQuery(t *testing.T) {
	got := ericQueryOf("https://example.org/articles/factoring-quadratics-in-algebra.html")
	if got != "factoring quadratics in algebra" {
		t.Errorf("ericQueryOf = %q, want the title words", got)
	}
}
