package fuzz

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// seededCopies are the copies the drive puts in the cache for a corroboration's url before red reads
// it: the url's own bytes, each route by which another copy stands in for it, a recovery whose body
// is the tool's own statement, and — the last — no copy at all. Turned by a counter, never by the
// run's rng, so every later draw of the run is the one it was.
var seededCopies = []struct{ substitution, copyURL, word string }{
	{"", "", "none"},
	{fetchcache.SubstitutionFollowedLink, "https://fuzz.invalid/full-text", "followed_link"},
	{fetchcache.SubstitutionRecovered, "https://web.archive.org/web/2026id_/https://fuzz.invalid/", "recovered"},
	{fetchcache.SubstitutionFullerCopy, "https://fuzz.invalid/open-copy.html", "fuller_copy"},
	{fetchcache.SubstitutionNamedBackend, "https://arxiv.org/html/2601.00001", "named_backend"},
	{fetchcache.SubstitutionRecovered, "", "recovered"},
	{"uncached", "", "not_cached"},
}

var copySeeds atomic.Int64

// seedCopy caches the next seeded copy under url and returns the word and copy url a reading of it
// records; the uncached seed caches nothing.
func (r *runner) seedCopy(url string) (word, copyURL string) {
	s := seededCopies[int(copySeeds.Add(1)-1)%len(seededCopies)]
	if s.substitution == "uncached" {
		return s.word, ""
	}
	body := []byte(sourceBody(url))
	entry := fetchcache.EntryFor(r.run(), url, &fetchcache.Attempt{Body: body, ContentType: "text/plain", TextRetrieved: true, URL: s.copyURL}, s.substitution)
	if _, err := fetchcache.Store(r.run(), entry, body); err != nil {
		r.noteActErr("seeding the cache for " + url + ": " + err.Error())
	}
	return s.word, s.copyURL
}

// copyOf is the copy a verb's --json result names.
func copyOf(out string) (word, copyURL string) {
	var env struct {
		Result struct {
			CopySubstitution string `json:"copy_substitution"`
			CopyURL          string `json:"copy_url"`
		} `json:"result"`
	}
	if json.Unmarshal([]byte(out), &env) != nil {
		return "", ""
	}
	return env.Result.CopySubstitution, env.Result.CopyURL
}

// citationCopy is the copy the record holds for citation anchor's source — blue's cite or a labelled
// corroboration — as `show evidence` renders its source row; ok is false where no row names it.
func (r *runner) citationCopy(anchor string) (word, copyURL string, ok bool) {
	b, bok := r.board()
	if !bok {
		return "", "", false
	}
	for _, s := range record.EvidenceJSONOf(b.Events, b.At).Sources {
		if s.Anchor == anchor {
			return s.CopySubstitution, s.CopyURL, true
		}
	}
	return "", "", false
}

// noteCopy holds a reading's result to the copy expected of it, and counts the word under verb.
func (r *runner) noteCopy(verb, out, wantWord, wantURL string) {
	word, copyURL := copyOf(out)
	if word != wantWord || copyURL != wantURL {
		r.noteActErr(fmt.Sprintf("%s recorded copy %q %q, and the cache held %q %q: %s", verb, word, copyURL, wantWord, wantURL, firstLine(out)))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.copiesRead == nil {
		r.copiesRead = map[string]map[string]int{}
	}
	if r.copiesRead[verb] == nil {
		r.copiesRead[verb] = map[string]int{}
	}
	r.copiesRead[verb][word]++
}
