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
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	neturl "net/url"
	"os"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
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
	// TDMReserved and TDMPolicy carry a text-and-data-mining reservation declared by ANY hop of
	// this fetch, not only the last one.
	//
	// THE DECLARATION IS USUALLY MADE ON A PAGE WE PASS THROUGH. Elsevier puts it on the
	// markup-redirect bouncer, so once the fetcher began following those, the reservation stopped
	// reaching the cache entirely — the entry that lands is the article, or a metadata record
	// after the article refuses, and neither carries the tag. A rights signal that the follow
	// silently drops is worse than one never read, because the absence looks like an answer.
	TDMReserved bool
	TDMPolicy   string
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
	// Pages is the document's page count where the format has one, else 0.
	Pages int `json:"pages,omitempty"`
	// NotRenderable marks a response that arrived as something other than the document its url
	// names — bytes, a 200, a sha, and no paper. Three causes, distinguished by the reason below
	// because they license different next acts: an ACCESS CHALLENGE, which usually means the host
	// publishes a sanctioned machine route to the same text; an unrendered client-side app, which
	// means nothing will render without a browser; and a page too starved of prose to be the
	// document either way. A seat verifying a citation against one of these is checking the
	// furniture, and the miss reads exactly like an honest check.
	//
	// A POINTER, for the reason TextExtracted is one: nil means nobody asked (an older index
	// line, or a content type the question does not apply to) and false means asked and
	// answered no. A plain bool collapses those into the reading that flatters.
	NotRenderable *bool `json:"not_renderable,omitempty"`
	// NotRenderableReason states what the detector saw, whenever NotRenderable is true — which of
	// the three it was, with the measurement. A flag with no reason is a verdict a reader cannot
	// check, and here it is also the only place the three causes are told apart.
	NotRenderableReason string `json:"not_renderable_reason,omitempty"`

	// TDMReserved says the source reserved text-and-data-mining rights in its own markup (the
	// W3C TDM Reservation Protocol), and TDMPolicy is where it says it states the terms.
	//
	// A THREE-STATE POINTER, like TextExtracted: nil means nobody asked — a content type the
	// question does not apply to — and false means the page was read and reserved nothing. A
	// plain bool would report every PDF and every JSON record as unreserved, which is a claim
	// about a document nobody examined. See TDMReservation for what the reservation covers, and
	// what it does not: not reading, and not quotation.
	TDMReserved *bool  `json:"tdm_reserved,omitempty"`
	TDMPolicy   string `json:"tdm_policy,omitempty"`

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
	//	robots    — nothing was asked of the origin at all; this host's robots.txt disallows the
	//	            path, so there is no HTTP status to record and inventing one would attribute a
	//	            refusal to a source that never made it
	RefusalClass string `json:"refusal_class,omitempty"`

	// RetrievedVia names the archive snapshot these bytes came from, when the live source refused
	// and the fallback found one. Empty means the bytes are the live source's own.
	//
	// IT IS PROVENANCE, NOT A FOOTNOTE. A snapshot is a different artifact: what the source said
	// on a date, retrieved from a third party. A citation that could not say so would claim to
	// have read something it did not.
	RetrievedVia string `json:"retrieved_via,omitempty"`

	// Backend NAMES WHICH BACKEND PRODUCED THESE BYTES — one of the Via constants, empty for a
	// plain live fetch. RetrievedVia beside it is a sentence for a human; this is the field a
	// reader may branch on, and it exists because a reader WAS branching on the sentence's mere
	// presence: fetch's summary printed "these bytes are an archive snapshot, retrieved from a
	// third party" over every recovered fetch, including an arXiv PDF taken live from arxiv.org
	// seconds earlier. The provenance duty is real for all of them; the archive's particular
	// warning is true of exactly one.
	Backend string `json:"backend,omitempty"`

	// TextRetrieved says whether the bytes are the SOURCE'S TEXT or only a record that it exists.
	// A metadata answer is a real finding and a legitimate citation — as `source_text_read:
	// unread`. It is not a reading, and nothing may cite it as one.
	TextRetrieved bool `json:"text_retrieved,omitempty"`
	// TextRetrievedReason states WHY the bytes are not the source's text on a fetch that reached
	// the source and got its real document — the paired reason this flag lacked, in the shape
	// TextExtracted and NotRenderable already use here. A withdrawn claim with no reason is
	// indistinguishable from a fetch that never happened.
	TextRetrievedReason string `json:"text_retrieved_reason,omitempty"`

	// Work carries WHAT THE INDEX SAYS ABOUT THE PAPER, as against every other field here, which
	// says something about this url. Retraction is the reason it is on the record at all: it is a
	// fact about the work that survives a perfect fetch, and a seat that read the pdf cleanly has
	// no other way to learn it. Nil where no index was consulted — a plain live fetch of a url
	// with no doi asks nobody, and an empty WorkFacts there would read as "asked, nothing found".
	Work *WorkFacts `json:"work,omitempty"`

	// CopyVersion and CopyLicense describe THESE BYTES, not the work: which of its copies this is
	// (publishedVersion, acceptedVersion, submittedVersion) and under what licence that copy sits.
	// They are separate from Work.License, which is the best open copy's — and the best open copy
	// is often not the one that answered.
	CopyVersion string `json:"copy_version,omitempty"`
	CopyLicense string `json:"copy_license,omitempty"`
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

// Classify records what the stored bytes ARE, and is the one place that asks.
//
// IT EXISTS BECAUSE ONLY THE LIVE PATH ASKED. The wall detector ran where a fetch reached the
// source directly, and every RECOVERED document — an archive snapshot, an open-access copy, a
// bibliographic record — was stored without it. That is exactly backwards: the archive is the
// rung most likely to hand back a landing page rather than a paper, and this tool's own summary
// says so in as many words.
//
// Measured on a 50-url scan: an archived Journal of Chemical Physics page was recorded as a
// retrieved document at 1,525 bytes carrying TEN visible characters. On the live path the same
// bytes are refused as starved. A seat would have read that as the paper.
//
// So both paths call this, and a third path cannot silently skip it: there is one place that
// decides, and it takes the bytes rather than the route they arrived by.
func Classify(entry *Entry, body []byte) {
	// HTML ONLY, and the answer is recorded either way — "we looked and it is a document" is the
	// fact that makes the flag's ABSENCE mean something. DefaultExtractor is a PDF extractor and
	// reports Attempted=false for HTML deliberately, so this cannot live inside the extraction
	// block: it would be dead code on precisely the content type it is about.
	if strings.Contains(entry.ContentType, "html") {
		shell := ShellReason(entry.ContentType, body)
		notRenderable := shell != ""
		entry.NotRenderable = &notRenderable
		entry.NotRenderableReason = shell
	}
}

// SniffedMediaType is the media type a response ACTUALLY carries, preferring the bytes over the
// header whenever the header declines to say.
//
// THE HEADER IS AN OPINION AND THE MAGIC BYTES ARE THE DOCUMENT. PubMed Central's open-access
// bucket serves its PDFs as `binary/octet-stream`; measured, a 983,106-byte body beginning
// `%PDF-1.4` was cached with that type, which meant the PDF extractor never ran and the OCR path
// could not fire either — `applicableToOCR` asks for `application/pdf`. The tool had gone to the
// trouble of finding a paper three indexes had hidden, and then could not read it.
//
// It only ever overrides a type that declines to be specific. A source calling its bytes
// `text/html` is making a claim this does not second-guess; a source saying `octet-stream` is
// saying it does not know, and here we do.
func SniffedMediaType(declared string, body []byte) string {
	mt := MediaType(declared)
	switch mt {
	case "", "application/octet-stream", "binary/octet-stream", "application/force-download", "application/download":
	default:
		return mt
	}
	switch {
	case bytes.HasPrefix(body, []byte("%PDF")):
		return "application/pdf"
	case bytes.HasPrefix(bytes.TrimLeft(body, " \t\r\n"), []byte("<?xml")):
		return "application/xml"
	}
	return mt
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
		// A ROBOTS REFUSAL IS ALSO A REFUSAL TO RECOVER FROM, and it used not to be.
		//
		// The recovery chain keyed on *Refusal alone, so a page the operator disallows ended the
		// fetch outright — while the refusal's own text tells the seat to "ask `metadata` whether
		// it exists, or `oa` whether a copy that is meant to be read exists elsewhere". The tool
		// named the route and declined to take it.
		//
		// It is not a corner case. IOP publishes `Disallow: /`, so every IOP paper was lost;
		// measured on one of them, `--via oa` returns a 633 KB PDF from arXiv. Being told we may
		// not read the publisher's copy says nothing about the copy the author put in a
		// repository — which is the whole reason the open-access rung exists.
		//
		// Nothing here evades the instruction: the recovery chain fetches through the same client,
		// so a candidate on the disallowed host is refused again by the same rule. What changes is
		// that the OTHER hosts are now asked.
		var ref *Refusal
		var rob *RobotsRefusal
		isRefusal, isRobots := errors.As(ferr, &ref), errors.As(ferr, &rob)
		if !isRefusal && !isRobots {
			return Entry{}, nil, false, ferr
		}
		stub := Entry{URL: url}
		if isRefusal {
			stub.HTTPStatus, stub.RefusalClass = ref.Status, refusalClass(ref.Status)
		} else {
			// NOT an HTTP status: nothing was asked of the origin. Recording one would invent a
			// refusal the source never made.
			stub.RefusalClass = "robots"
		}
		_ = appendIndexIfAbsent(run, stub)
		// THE REFUSAL IS NOT THE END OF THE ATTEMPT, but the right next move depends on WHAT this
		// source is — and choosing wrongly is how a run ends up citing a landing page.
		att := Recover(f, url, ViaAuto, "")
		if att == nil {
			return Entry{}, nil, false, ferr
		}
		entry := EntryFor(url, att)
		entry.HTTPStatus, entry.RefusalClass = stub.HTTPStatus, stub.RefusalClass
		entry, serr := Store(run, entry, att.Body)
		if serr != nil {
			return Entry{}, nil, false, ferr
		}
		return entry, att.Body, false, nil
	}
	entry := Entry{URL: url, ContentType: SniffedMediaType(resp.ContentType, resp.Body)}
	entry.Sha = Sha(resp.Body)

	ex := DefaultExtractor.Extract(Dir(run), entry.ContentType, resp.Body)
	entry.Filename = Label(ex.Title, resp.Disposition, url)
	entry.Pages = ex.Pages
	if ex.Attempted {
		extracted := ex.Text != ""
		entry.TextExtracted = &extracted
		entry.Extractor = ex.ExtractorID
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
	// OUTSIDE THE EXTRACTION BLOCK, BECAUSE HTML NEVER ENTERS IT. DefaultExtractor is a PDF
	// extractor and reports Attempted=false for HTML deliberately; a check placed inside would
	// be dead code on precisely the content type it is about.
	//
	// The answer is recorded either way for HTML — "we looked and it is a document" is the fact
	// that makes the flag's ABSENCE mean something.
	Classify(&entry, resp.Body)
	// READ OFF THE RESPONSE, NOT THE FINAL BODY, so a reservation declared on a hop this fetch
	// passed through is still recorded — Elsevier declares it on the markup-redirect bouncer,
	// which the fetcher now follows past.
	//
	// THE POINTER IS SET ONLY WHERE THE QUESTION WAS ASKABLE, which is what its three states
	// mean: a PDF or a JSON record leaves it nil, because nothing looked, and writing `false`
	// there would report a document nobody examined as declaring nothing.
	if resp.TDMReserved || strings.Contains(entry.ContentType, "html") {
		reserved := resp.TDMReserved
		entry.TDMReserved = &reserved
		entry.TDMPolicy = resp.TDMPolicy
	}
	// THE PAGE MAY BE THE ABSTRACT, AND IT USUALLY SAYS SO ITSELF.
	//
	// MEASURED over 120 works: of 25 html bodies this tool recorded as documents, ELEVEN were
	// abstract or landing pages — 401 to 1,885 words of navigation, abstract and references, with
	// none of the sections a paper has. `text_retrieved: true` on those is the same false claim
	// as calling a zip the source's text, and it is worse, because an abstract reads like a
	// paper: a seat quoting one would be quoting the summary and saying it read the study.
	//
	// The fix is not a word-count threshold. Of those sixteen short pages, TWELVE carried the
	// publisher's own pointer to the readable copy — `citation_fulltext_html_url`,
	// `citation_pdf_url` — which is a stated fact rather than a guess about length, and the page
	// that emits it is the system serving the article.
	//
	// One hop, only when the first answer was html. A pdf or an xml body is already the document.
	base, _ := neturl.Parse(url)
	if fullText := LandingPageFullText(entry.ContentType, resp.Body, "", base); fullText != "" {
		if hop, herr := f.Fetch(fullText); herr == nil && len(hop.Body) > len(resp.Body) {
			// LONGER IS THE TEST, AND IT IS DELIBERATELY CRUDE. The pointer is the publisher's, so
			// this is not deciding WHICH is the paper — it is refusing to trade a page for a
			// smaller one, which is what a paywall stub or an error page would be.
			resp = hop
			entry.ContentType = SniffedMediaType(hop.ContentType, hop.Body)
			entry.RetrievedVia = fmt.Sprintf("the full text at %s, which the page at %s names in its own citation metadata", fullText, url)
		}
	}
	// AND THE WORK'S OWN FACTS, ON THE PATH THAT SUCCEEDS.
	//
	// THIS WAS BACKWARDS AND A SWEEP SHOWED IT. The index lookup lived in Recover's stamp, which
	// runs only when the live fetch FAILED — so every paper this tool actually read came back
	// with no retraction check, and every paper it could not read came back with one. Measured
	// over 47 works: all 28 rows carrying work facts were `metadata` or `oa` answers; all 8 rows
	// that returned a readable document carried none. The flagship fact was present on exactly
	// the documents nobody could quote.
	//
	// One request, to a host with a 500ms floor, on a path whose median is tens of seconds. It is
	// the cheapest thing here and it is the one that can void a citation.
	if doi := DOIOf(url); doi != "" {
		if facts, _, ok := openAlexWork(f, doi); ok && facts != (WorkFacts{}) {
			entry.Work = &facts
		}
	}
	// AND THE LIVE PATH RECORDS WHY IT IS NOT TEXT, not just that it is not. The summary can
	// derive the flag from the content type, but only the record outlives the run — an archived
	// entry holding `text_retrieved: false` and nothing else cannot say whether the source
	// refused, the page was a wall, or the bytes were a tarball.
	if !TextBearing(entry.ContentType) {
		entry.TextRetrievedReason = textBearingRefusal(entry.ContentType)
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

// EntryFor builds the index record for one backend attempt: every field the Attempt carries, the
// classification of its bytes, and the withdrawal that follows from it.
//
// ONE FUNCTION BECAUSE THERE ARE TWO CALLERS AND THEY DRIFTED. Each route that stores an attempt
// copied the same fields by hand, so the newest field reached whichever site was edited and the
// other stored a record missing it — silently, because a missing field is indistinguishable from
// a fact the backend did not learn.
func EntryFor(url string, att *Attempt) Entry {
	entry := Entry{
		URL: url, ContentType: att.ContentType,
		RetrievedVia: att.Via, Backend: att.Backend, TextRetrieved: att.TextRetrieved,
		CopyVersion: att.Version, CopyLicense: att.License,
	}
	if att.Facts != (WorkFacts{}) {
		facts := att.Facts
		entry.Work = &facts
	}
	Classify(&entry, att.Body)
	// AND BYTES NOTHING CAN READ ARE NOT THE SOURCE'S TEXT, however honestly they were fetched.
	if entry.TextRetrieved && !TextBearing(entry.ContentType) {
		entry.TextRetrieved = false
		entry.TextRetrievedReason = textBearingRefusal(entry.ContentType)
	}
	// A RECOVERED WALL IS NOT A RECOVERED DOCUMENT. Where the bytes turn out to be a landing
	// page or an interstitial, the claim that text was retrieved is withdrawn with them.
	if entry.NotRenderable != nil && *entry.NotRenderable {
		entry.TextRetrieved = false
	}
	return entry
}

// TextBearing says whether a media type can carry the SOURCE'S TEXT AT ALL, as against being a
// container or a binary this tool has no reader for.
//
// MEASURED: jair.org's Crossref-registered text-mining link serves `jair.ps.Z` — a compressed
// PostScript file, 337 KB, delivered as `application/zip`. The fetch reached the source, took its
// real document, and recorded `text_retrieved: true`. Nothing in this tool can open it, so the
// claim was false in the one field a citation's `leaf` reading rests on, and false in the
// optimistic direction: a seat reading the summary is told the source's text is in hand.
//
// AN ALLOWLIST, DELIBERATELY. The unknown type is the interesting case and a denylist answers it
// `true` — every new container this tool has never seen becomes a claim that its text was
// retrieved. Listing what can be read makes the unknown answer `false`, which is also what a
// reader can check by looking at the cached bytes.
func TextBearing(contentType string) bool {
	mt := MediaType(contentType)
	if strings.HasPrefix(mt, "text/") {
		return true
	}
	switch mt {
	// Each of these is either plain text a seat reads directly, or a pdf, which is the one
	// binary this tool has readers for — the extractor's text layer, and the OCR engine's
	// reading of the page images when there is none.
	case "application/pdf", "application/xml", "application/json", "application/xhtml+xml",
		"application/x-tex", "application/ld+json":
		return true
	}
	// POSTSCRIPT AND RTF ARE NOT ON THAT LIST, and the first draft of it had them. They carry
	// text in a sense no reader here can act on: DefaultExtractor handles pdf alone, and a seat
	// handed a .ps has the document and no way to quote it. Listing a format nothing reads
	// recreates the exact claim this function exists to withdraw.
	// A type that declines to say. SniffedMediaType has already looked at the magic bytes and
	// found neither a pdf nor xml, so there is nothing further to go on, and the honest answer to
	// "is the source's text in these bytes" is that nobody knows.
	return false
}

// textBearingRefusal is the sentence recorded when the bytes cannot carry text. It names the
// type, because "we could not read it" without saying what arrived is a dead end for whoever
// reads the record next.
func textBearingRefusal(contentType string) string {
	return fmt.Sprintf("the source answered with %s — a container or binary this tool has no reader for, so these bytes are the document and NOT its text", MediaType(contentType))
}

// TRANSPORT COMPRESSION IS ALREADY HANDLED, AND THE WAY TO BREAK IT IS TO HELP.
//
// net/http adds `Accept-Encoding: gzip` itself and decompresses the response transparently —
// but ONLY while the caller sets no Accept-Encoding header of its own. Setting one, even to the
// same value, switches the transport to raw mode and hands back the compressed bytes with the
// Content-Encoding header still attached. Every gzipped page would then arrive as binary, sniff
// as a container, and be recorded as "not the source's text" with a plausible reason.
//
// So the invariant is a SILENCE, which is why it is written down here and pinned by a test: a
// future edit adding an Accept-Encoding header, or DisableCompression, looks helpful and is not.
//
// This is a different question from a compressed ARTIFACT. `jair.ps.Z` served as
// `application/zip` carries no Content-Encoding at all (measured, 2026-09-23: the response has
// Content-Type and Content-Disposition and no encoding header). The compression is the document,
// not the transport, and no Accept-Encoding would have changed what the server sent.
