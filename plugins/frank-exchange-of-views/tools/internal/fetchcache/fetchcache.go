// Package fetchcache is the per-run, content-addressed source cache behind
// `feov-record fetch` and `blue cite`.
//
// THE PROBLEM. A research debate cites the web, and both sides must reason about the
// SAME bytes: blue evaluates a source, red re-verifies the claim it backs. If each seat
// fetched live, red could audit against a page that changed since blue read it — the
// disagreement would be an artifact of time, not of judgement. And nothing pinned what
// blue actually saw, so a citation was a URL and a hope.
//
// THE MODEL. Fetch a URL ONCE per run; cache the bytes at <run>/cache/<sha256> keyed by
// their own hash; hand every later reader those exact bytes. The cache is:
//   - content-addressed: the filename IS the sha256, so identical bytes dedup and a cited
//     source carries a hash a reader can verify.
//   - download-once: the first fetch's bytes are canonical for the run (a URL that drifts
//     mid-run does not silently give two readers two truths).
//   - url-indexed: <run>/cache/index maps url -> sha256 so a second fetch of the same URL
//     is a cache hit, not a second download.
//
// The Fetcher is injected (Default in prod is net/http with SSRF caps; tests stub it) so
// the cache logic is exercised without ever touching the network — the one real-data
// check is a dev exercise, not a CI test.
package fetchcache

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"os"
	"path/filepath"
	"strings"
)

// Fetcher performs the one live read. Prod is an SSRF-capped net/http client (httpfetcher.go);
// tests supply a deterministic stub so the cache is testable offline.
type Fetcher interface {
	// Fetch GETs url and returns its body, or a non-nil error. It never touches the cache.
	Fetch(url string) ([]byte, error)
}

// Default is the process-wide Fetcher `fetch`/`blue cite` use. It is a variable for the
// same reason record.Now is: the command tree is built once by newRoot(), so a test pins
// behaviour by swapping this (and restoring it via defer), exactly as the golden harness
// pins the clock. Nothing else mutates it.
var Default Fetcher = NewHTTPFetcher()

// Dir is a run's cache directory, <run>/cache.
func Dir(run record.Run) string { return filepath.Join(run.Dir(), "cache") }

// Path is the cache file for a given content hash, <run>/cache/<sha256>.
func Path(run record.Run, sha string) string { return filepath.Join(Dir(run), sha) }

// indexPath is the url->sha256 manifest, a JSON-lines file. "index" is not a 64-char hex
// string, so it never collides with a content file in the same directory. JSONL (not a
// tab-delimited line) so a URL carrying a tab or newline round-trips instead of corrupting
// the manifest — the caller's URL is untrusted text.
func indexPath(run record.Run) string { return filepath.Join(Dir(run), "index") }

// indexEntry is one url->sha256 mapping line in the manifest.
type indexEntry struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
}

// Sha is the lowercase-hex sha256 of b — the cache key and the hash a citation records.
func Sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Read returns the cached bytes for a content hash.
func Read(run record.Run, sha string) ([]byte, error) { return os.ReadFile(Path(run, sha)) }

// Lookup returns the cached bytes for url if this run has already fetched it AND the
// content file is still present. A first match in the index wins (download-once: the first
// fetch's hash is canonical). A missing content file behind an index line is treated as a
// miss, so a crash between the content write and the index append self-heals on re-fetch.
func Lookup(run record.Run, url string) (sha string, b []byte, ok bool, err error) {
	f, err := os.Open(indexPath(run))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, false, nil // no reads yet this run
		}
		return "", nil, false, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20) // a URL line can be long
	for sc.Scan() {
		var e indexEntry
		if json.Unmarshal(sc.Bytes(), &e) != nil || e.URL != url {
			continue
		}
		bytes, rerr := Read(run, e.Sha)
		if rerr != nil {
			continue // index points at a content file that is gone → treat as a miss
		}
		return e.Sha, bytes, true, nil
	}
	return "", nil, false, sc.Err()
}

// Store writes b to the content-addressed cache and records url -> sha256 in the index.
// The content write is atomic (temp + rename) and dedups on the hash; the index line is
// appended only if this exact (sha, url) pair is not already present. Returns the hash.
func Store(run record.Run, url string, b []byte) (string, error) {
	if err := os.MkdirAll(Dir(run), 0o755); err != nil {
		return "", err
	}
	sha := Sha(b)
	dst := Path(run, sha)
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		tmp, err := os.CreateTemp(Dir(run), ".tmp-*")
		if err != nil {
			return "", err
		}
		tmpName := tmp.Name()
		if _, err := tmp.Write(b); err != nil {
			tmp.Close()
			os.Remove(tmpName)
			return "", err
		}
		if err := tmp.Close(); err != nil {
			os.Remove(tmpName)
			return "", err
		}
		if err := os.Rename(tmpName, dst); err != nil {
			os.Remove(tmpName)
			return "", err
		}
	} else if err != nil {
		return "", err
	}
	if err := appendIndexIfAbsent(run, sha, url); err != nil {
		return "", err
	}
	return sha, nil
}

// appendIndexIfAbsent records a {sha,url} JSON line unless that exact pair is already
// present. A duplicate line would be harmless (Lookup takes the first match), so the
// read-then-append is best-effort, not locked: two identical appends still resolve to the
// same bytes.
func appendIndexIfAbsent(run record.Run, sha, url string) error {
	entry, err := json.Marshal(indexEntry{Sha: sha, URL: url})
	if err != nil {
		return err
	}
	if existing, err := os.ReadFile(indexPath(run)); err == nil {
		for _, l := range strings.Split(string(existing), "\n") {
			if l == string(entry) {
				return nil
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(indexPath(run), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(entry, '\n')); err != nil {
		return err
	}
	return nil
}

// Resolve is the download-once entry point both `fetch` and `blue cite` call: a cache hit
// returns the stored bytes untouched; a miss fetches ONCE via f, caches, and returns. hit
// reports which path was taken (so `fetch` can show a second read was served from cache).
// A fetch error is returned verbatim — the CALLER decides whether that is a friction
// (an unusable CITED source) or a bare miss (a read that may legitimately fail).
func Resolve(run record.Run, url string, f Fetcher) (sha string, b []byte, hit bool, err error) {
	if s, cached, ok, lerr := Lookup(run, url); lerr != nil {
		return "", nil, false, lerr
	} else if ok {
		return s, cached, true, nil
	}
	fetched, ferr := f.Fetch(url)
	if ferr != nil {
		return "", nil, false, ferr
	}
	s, serr := Store(run, url, fetched)
	if serr != nil {
		return "", nil, false, serr
	}
	return s, fetched, false, nil
}
