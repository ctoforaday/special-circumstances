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
	"errors"
	"fmt"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"os"
	"path/filepath"
	"strings"
)

// Fetcher performs the one live read. Prod is an SSRF-capped net/http client (httpfetcher.go);
// tests supply a deterministic stub so the cache is testable offline.
type Fetcher interface {
	// Fetch GETs url and returns its body plus the response facts the cache records, or a
	// non-nil error. It never touches the cache.
	Fetch(url string) (*Response, error)
}

// Response is what one live read produced: the bytes, and the two facts the response header
// carried about them.
//
// THE HEADER SAYS WHAT THE BYTES ARE, AND WE USED TO THROW IT AWAY. `Fetch` returned a bare
// []byte, so `Content-Type` — measured at the source, by the source — was discarded at the
// only moment it was ever available, and every downstream reader had to sniff magic bytes or
// guess an extension off the URL. Guessing off the URL demonstrably fails on real sources:
// the Auer paper is served from `https://inria.hal.science/inria-00574987/document`, which
// carries no extension at all. This is the [[facts-are-fields]] repair, and it needed no new
// record invented — <run>/cache/index already existed and already had a writer.
type Response struct {
	// Body is the response body, already size-capped by the Fetcher.
	Body []byte
	// ContentType is the raw Content-Type header, parameters and all. Callers wanting the
	// bare media type use MediaType; the raw form is kept here so nothing silently
	// reinterprets a header the cache did not parse.
	ContentType string
	// Disposition is the raw Content-Disposition header, or "" — the middle rung of Label's
	// filename chain. Measured across the cited corpus: not one source sent it, which is
	// exactly why it is a rung and not the rule.
	Disposition string
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

// Entry is one line of the manifest: what this run fetched from a URL, and what it turned
// out to be. It is the record BOTH sides of the cache speak — Store writes exactly this and
// Lookup returns exactly this — so a fact cannot be written in one shape and read in another.
//
// EVERY ADDED FIELD IS `omitempty` AND OPTIONAL ON READ. An index written by an older binary
// has only {sha,url}; it must keep resolving, with the new fields reading as "" rather than
// as a corrupt line. The zero value therefore means NOT MEASURED, and the printers say so
// rather than rendering an empty string as a measured empty answer.
type Entry struct {
	Sha string `json:"sha"`
	URL string `json:"url"`
	// ContentType is the bare media type from the response header ("application/pdf"), or ""
	// where the header was absent or unparseable.
	ContentType string `json:"content_type,omitempty"`
	// Filename is the human-readable LABEL from Label — never an identity, never a lookup key.
	Filename string `json:"filename,omitempty"`
	// TextSha is the sha256 of the extracted text stored beside the body, or "" where nothing
	// was extracted. It is what makes an audit of the text a seat actually READ possible: the
	// body's own sha proves the document is unchanged, and says nothing about the extraction.
	TextSha string `json:"text_sha,omitempty"`
	// Extractor identifies what produced the text, as library + semver (#629 D3).
	Extractor string `json:"extractor,omitempty"`
	// TextExtracted is a three-state fact carried as a pointer on purpose: nil means the
	// question was never asked (an older index line, or a content type nothing extracts),
	// false means it was asked and the answer was no. A plain bool would collapse those into
	// the same false — the [[facts-are-fields]] miss that reads exactly like an honest zero.
	TextExtracted *bool `json:"text_extracted,omitempty"`
	// TextReason states WHY, whenever TextExtracted is false. An empty extraction is never
	// recorded as a silent zero.
	TextReason string `json:"text_reason,omitempty"`
	// OCRDerived marks text that came from optical recognition rather than a text layer, so
	// its weaker reproducibility is stated up front rather than discovered when a `reproduce`
	// fails mysteriously.
	OCRDerived bool `json:"ocr_derived,omitempty"`
	// Pages is the document's page count where the format has one, else 0.
	Pages int `json:"pages,omitempty"`

	// HTTPStatus is the status the origin (or whatever answered for it) returned. It is here
	// because a REFUSED fetch used to leave no trace at all: the error went back to the seat and
	// the index recorded nothing, so "we could not read this" survived only as prose in a
	// report — where a run then counted it by grep and got the count wrong twice.
	HTTPStatus int `json:"http_status,omitempty"`

	// RefusalClass says WHO refused, and its most important value is the one that admits it does
	// not know. A 403 from an egress proxy refusing the host and a 403 from an origin refusing
	// the client are the same status line, and nothing in the response separates them — one is a
	// fact about this container, the other a fact about the source, and they license completely
	// different next actions. Where a proxy is configured the honest answer is `unknown`, and
	// recording that is the point: an ambiguity carried is not an ambiguity resolved by guess.
	//
	//	origin   — no proxy is configured, so the refusal is the source's own
	//	unknown   — a proxy is configured and the two readings cannot be told apart
	RefusalClass string `json:"refusal_class,omitempty"`

	// RetrievedVia names the archive snapshot these bytes came from, when the live source refused
	// and the fallback found one. Empty means the bytes are the live source's own.
	//
	// IT IS PROVENANCE, NOT A FOOTNOTE. A snapshot is a different artifact: what the source said
	// on a date, retrieved from a third party. A citation that could not say so would claim to
	// have read something it did not.
	RetrievedVia string `json:"retrieved_via,omitempty"`
}

// Refusal is a fetch that was answered and REFUSED, carried as a typed error so the status
// survives the trip back to a caller that would otherwise see only prose.
type Refusal struct {
	URL    string
	Status int
	Note   string
}

func (r *Refusal) Error() string {
	return fmt.Sprintf("fetch: %s returned HTTP %d%s", r.URL, r.Status, r.Note)
}

// Sha is the lowercase-hex sha256 of b — the cache key and the hash a citation records.
func Sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// Read returns the cached bytes for a content hash.
func Read(run record.Run, sha string) ([]byte, error) { return os.ReadFile(Path(run, sha)) }

// Lookup returns the cached entry and bytes for url if this run has already fetched it AND
// the content file is still present. A first match in the index wins (download-once: the first
// fetch's hash is canonical). A missing content file behind an index line is treated as a
// miss, so a crash between the content write and the index append self-heals on re-fetch.
func Lookup(run record.Run, url string) (e Entry, b []byte, ok bool, err error) {
	f, err := os.Open(indexPath(run))
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, nil, false, nil // no reads yet this run
		}
		return Entry{}, nil, false, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20) // a URL line can be long
	for sc.Scan() {
		var got Entry
		if json.Unmarshal(sc.Bytes(), &got) != nil || got.URL != url {
			continue
		}
		bytes, rerr := Read(run, got.Sha)
		if rerr != nil {
			continue // index points at a content file that is gone → treat as a miss
		}
		return got, bytes, true, nil
	}
	return Entry{}, nil, false, sc.Err()
}

// Store writes b to the content-addressed cache and records the entry in the index. The
// caller supplies everything measured about the fetch except the hash; Store fills Sha and
// returns the completed entry. The content write is atomic (temp + rename) and dedups on the
// hash; the index line is appended only if that exact line is not already present.
func Store(run record.Run, e Entry, b []byte) (Entry, error) {
	if err := os.MkdirAll(Dir(run), 0o755); err != nil {
		return Entry{}, err
	}
	e.Sha = Sha(b)
	if err := writeAtomic(Path(run, e.Sha), b); err != nil {
		return Entry{}, err
	}
	if err := appendIndexIfAbsent(run, e); err != nil {
		return Entry{}, err
	}
	return e, nil
}

// TextPath is the extraction stored beside a document, <run>/cache/<sha256>.txt.
//
// THE NAME IS DERIVED, AND THAT IS NOT THE [[facts-are-fields]] SMELL IT RESEMBLES. Nothing
// ever recovers a fact by parsing this path: the extraction's own hash, its extractor and
// whether it exists at all are FIELDS on the index Entry, and a reader consults those. The
// suffix exists so a human listing the directory can see which blob is which — and because
// the alternative, a second content-addressed name, would need its own index to find.
func TextPath(run record.Run, sha string) string { return Path(run, sha) + ".txt" }

// StoreText writes an extraction beside its source document and returns the extraction's own
// sha256. It does NOT touch the index — the caller folds the hash into the Entry it stores,
// so one write produces one record rather than two that can disagree.
func StoreText(run record.Run, sourceSha string, text []byte) (string, error) {
	if err := os.MkdirAll(Dir(run), 0o755); err != nil {
		return "", err
	}
	if err := writeAtomic(TextPath(run, sourceSha), text); err != nil {
		return "", err
	}
	return Sha(text), nil
}

// writeAtomic writes b to dst via a temp file and a rename, and treats an existing dst as
// already done. Content-addressed callers get dedup for free; TextPath callers get an
// idempotent rewrite of identical bytes.
func writeAtomic(dst string, b []byte) error {
	if _, err := os.Stat(dst); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}

// appendIndexIfAbsent records the entry as a JSON line unless that exact line is already
// present. A duplicate line would be harmless (Lookup takes the first match), so the
// read-then-append is best-effort, not locked: two identical appends still resolve to the
// same bytes.
func appendIndexIfAbsent(run record.Run, e Entry) error {
	entry, err := json.Marshal(e)
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
// returns the stored entry untouched; a miss fetches ONCE via f, extracts, caches both
// artifacts, and returns. hit reports which path was taken (so `fetch` can show a second read
// was served from cache). A fetch error is returned verbatim — the CALLER decides whether that
// is a friction (an unusable CITED source) or a bare miss (a read that may legitimately fail).
//
// EXTRACTION HAPPENS HERE, not in `fetch`, because `blue cite` resolves through this same door
// and cite is the act that records what backs a claim. Putting it one level up would give the
// citation path a document with no recoverable text and call that finished.
//
// AN EXTRACTION FAILURE IS NEVER A FETCH FAILURE. A scanned 1998 standard with no text layer is
// a perfectly good cached source; what it is not is readable. So the reason travels on the
// Entry and the fetch succeeds, rather than the whole read failing and a seat concluding the
// source was unreachable.
func Resolve(run record.Run, url string, f Fetcher) (e Entry, b []byte, hit bool, err error) {
	if got, cached, ok, lerr := Lookup(run, url); lerr != nil {
		return Entry{}, nil, false, lerr
	} else if ok {
		return got, cached, true, nil
	}
	resp, ferr := f.Fetch(url)
	if ferr != nil {
		// A REFUSAL IS EVIDENCE, and it used to have nowhere to live. The index now carries the
		// attempt — url, status, and who refused — so a later reader can tell a source that does
		// not exist from one this container could not reach. The error still goes back: recording
		// the refusal does not make it a success.
		var ref *Refusal
		if !errors.As(ferr, &ref) {
			return Entry{}, nil, false, ferr
		}
		_ = appendIndexIfAbsent(run, Entry{
			URL: url, HTTPStatus: ref.Status, RefusalClass: refusalClass(ref.Status),
		})
		// THE REFUSAL IS NOT THE END OF THE ATTEMPT. Every recovery in the measured run was a seat
		// doing this by hand against the Wayback CDX; the tool now does it, because a capability
		// that must be rebuilt by hand each time is one most seats will skip.
		snapURL, stamp, _ := SnapshotFor(f, url)
		if snapURL == "" {
			return Entry{}, nil, false, ferr
		}
		snapResp, serr := f.Fetch(snapURL)
		if serr != nil {
			return Entry{}, nil, false, ferr
		}
		// STORED UNDER THE URL THE SEAT ASKED FOR, so a later read of the same source is served
		// these bytes — and carrying the provenance, so nothing can mistake them for the live page.
		entry := Entry{
			URL: url, ContentType: MediaType(snapResp.ContentType),
			HTTPStatus: ref.Status, RefusalClass: refusalClass(ref.Status),
			RetrievedVia: ArchiveNote(snapURL, stamp),
		}
		entry, serr = Store(run, entry, snapResp.Body)
		if serr != nil {
			return Entry{}, nil, false, ferr
		}
		return entry, snapResp.Body, false, nil
	}
	entry := Entry{URL: url, ContentType: MediaType(resp.ContentType)}
	entry.Sha = Sha(resp.Body)

	ex := DefaultExtractor.Extract(Dir(run), entry.ContentType, resp.Body)
	entry.Filename = Label(ex.Title, resp.Disposition, url)
	entry.Pages = ex.Pages
	if ex.Attempted {
		extracted := ex.Text != ""
		entry.TextExtracted = &extracted
		entry.Extractor = ex.ExtractorID
		entry.OCRDerived = ex.OCRDerived
		if extracted {
			textSha, terr := StoreText(run, entry.Sha, []byte(ex.Text))
			if terr != nil {
				return Entry{}, nil, false, terr
			}
			entry.TextSha = textSha
		} else {
			// NO .txt FILE IS WRITTEN. An empty extraction on disk is indistinguishable from a
			// successful extraction of an empty document, and a seat that opens it learns
			// nothing and concludes the wrong thing. The absence plus the stated reason is the
			// honest record — [[facts-are-fields]] clause 3.
			entry.TextReason = ex.Reason
		}
	}
	stored, serr := Store(run, entry, resp.Body)
	if serr != nil {
		return Entry{}, nil, false, serr
	}
	return stored, resp.Body, false, nil
}

// LookupSha returns the index entry for a content hash, and whether one exists.
//
// Lookup answers "have we fetched this URL"; this answers "what do we know about these
// bytes". A verb that operates on something already cached — `ocr`, which takes a --sha a
// seat read out of a fetch summary — needs the second question, and had no way to ask it.
//
// FIRST MATCH WINS, the same rule Lookup states, because it is the same index and one file
// must not have two merge semantics. A reader that took the LAST match here while Lookup
// took the first would make "the entry for this document" mean two different lines.
//
// The content file is NOT required to still be present. Lookup treats a missing file as a
// cache miss because its caller is about to serve those bytes; this caller wants the
// RECORD — what the document is, whether text came out of it — and answering "no entry" for
// a document the index plainly describes would be a false statement about the index.
// URLsSharingBody lists every OTHER url already cached under these exact bytes.
//
// THE SAME BODY FOR TWO DIFFERENT SOURCES IS NOT WHAT A DOCUMENT LOOKS LIKE. It is what a bot
// wall, a login interstitial or a "no results" page looks like — and because the cache is
// content-addressed, those collapse onto ONE entry that satisfies "fetch-once, hash-verified"
// perfectly. The hash then certifies the BLOCKADE rather than the source.
//
// MEASURED in research/2026-09-02_quadratic-formula: three unrelated JSTOR articles — Savage 1989,
// Heaton 1896, and one more — all cached as sha 32ed6315…, a single 3,038-byte challenge page. The
// bibliography cited one of them as "the record for John Savage, Factoring Quadratics", whose
// cached bytes are a wall it shares with an unrelated paper. A citation to an unreadable source,
// presented as a source.
//
// It is deliberately NOT content sniffing. No keyword list, no page-shape heuristic — just the
// arithmetic the cache already does: these bytes were already served for something else. That is
// evidence a reader can check, and it cannot be fooled by a wall nobody has seen before.
func URLsSharingBody(run record.Run, sha, exceptURL string) ([]string, error) {
	f, err := os.Open(indexPath(run))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var out []string
	seen := map[string]bool{exceptURL: true}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var got Entry
		if json.Unmarshal(sc.Bytes(), &got) != nil || got.Sha != sha || seen[got.URL] {
			continue
		}
		seen[got.URL] = true
		out = append(out, got.URL)
	}
	return out, sc.Err()
}

func LookupSha(run record.Run, sha string) (Entry, bool, error) {
	f, err := os.Open(indexPath(run))
	if err != nil {
		if os.IsNotExist(err) {
			return Entry{}, false, nil // no reads yet this run
		}
		return Entry{}, false, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		var got Entry
		if json.Unmarshal(sc.Bytes(), &got) != nil || got.Sha != sha {
			continue
		}
		return got, true, nil
	}
	return Entry{}, false, sc.Err()
}
