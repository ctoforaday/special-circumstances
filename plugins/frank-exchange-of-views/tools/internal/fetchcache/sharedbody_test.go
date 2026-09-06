package fetchcache

import (
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// THE BOT WALL, REPRODUCED. Three unrelated JSTOR articles — Savage 1989, Heaton 1896, and one
// more — all cached as sha 32ed6315…, a single 3,038-byte challenge page, in
// research/2026-09-02_quadratic-formula. The cache is content-addressed, so they collapse onto ONE
// entry that satisfies "fetch-once, hash-verified" perfectly: the hash certifies the BLOCKADE.
//
// The bibliography then cited one of them as "the record for John Savage, Factoring Quadratics",
// whose cached bytes are a wall it shares with an unrelated paper — a citation to an unreadable
// source, presented as a source.
func TestTheSameBodyServedForTwoSourcesIsNamed(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	wall := []byte("Access check — please enable JavaScript.\n")

	for _, u := range []string{
		"https://www.jstor.org/stable/27966169", // Savage
		"https://www.jstor.org/stable/2971099",  // Heaton
		"https://www.jstor.org/stable/27966220",
	} {
		if _, err := Store(run, Entry{URL: u}, wall); err != nil {
			t.Fatal(err)
		}
	}

	sha := Sha(wall)
	shared, err := URLsSharingBody(run, sha, "https://www.jstor.org/stable/27966169")
	if err != nil {
		t.Fatal(err)
	}
	if len(shared) != 2 {
		t.Fatalf("a body served for three urls named %d others, want 2: %v", len(shared), shared)
	}
	// The url asked about must never be listed as sharing with itself.
	for _, u := range shared {
		if u == "https://www.jstor.org/stable/27966169" {
			t.Error("the url under inspection was listed as sharing a body with itself")
		}
	}
}

// A GENUINE DOCUMENT SHARES ITS BODY WITH NOTHING, and must not be flagged. Content-addressing
// dedupes a re-fetch of the SAME url onto one entry, which is the cache working, not a wall.
func TestAnOrdinaryDocumentNamesNoSharers(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	body := []byte("On the quadratic formula.\n")
	url := "https://example.org/paper.html"
	for i := 0; i < 2; i++ { // fetched twice: same url, same bytes
		if _, err := Store(run, Entry{URL: url}, body); err != nil {
			t.Fatal(err)
		}
	}
	shared, err := URLsSharingBody(run, Sha(body), url)
	if err != nil {
		t.Fatal(err)
	}
	if len(shared) != 0 {
		t.Errorf("a document fetched twice from one url was reported as sharing a body: %v", shared)
	}
}
