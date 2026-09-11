package fetchcache

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// THE BACKENDS ANSWER DIFFERENT QUESTIONS, WHICH IS WHY THERE IS MORE THAN ONE.
//
// A refusal used to end the attempt, and every recovery in
// research/2026-09-02_quadratic-formula was a seat rebuilding one of these by hand. But the
// strategies are not interchangeable, and choosing wrongly is how a run ends up citing a landing
// page:
//
//	archive   what this URL SAID ON A DATE. Right for web pages. Wrong for subscription
//	          articles, where the snapshot is usually the landing page and not the text.
//	oa        is there a LEGAL OPEN COPY, and where. Right for scholarship.
//	metadata  does this EXIST, in what venue, at what pages. Retrieves no text at all — and is
//	          the honest answer when there is none to retrieve.
//	arxiv     the preprint itself: abstract, PDF, or LaTeX source.
//
// THE MOST VALUABLE ANSWER IS OFTEN "NO". Measured: Crossref, OpenAlex and Unpaywall all agree
// that 10.5951/MT.82.1.0033 — Sharing Teaching Ideas, The Mathematics Teacher 82(1) pp.33-35,
// which is the Savage record — has no open copy anywhere. That is a DETERMINATE FACT where the
// run could only say "unreachable from this container", and the difference between those two is
// the whole of #736: one is a claim about this container, the other about the world.
const (
	ViaLive     = "live"
	ViaArchive  = "archive"
	ViaOA       = "oa"
	ViaMetadata = "metadata"
	ViaArxiv    = "arxiv"
	ViaEric     = "eric"
	ViaAuto     = "auto"
)

// Vias lists the backends a seat may name, in the order the help should present them.
func Vias() []string {
	return []string{ViaLive, ViaArchive, ViaOA, ViaMetadata, ViaArxiv, ViaEric, ViaAuto}
}

// AutoOrder is the order `auto` — and a refused live fetch — tries the backends in. One list, read
// by Recover and by fetch's help, so the page cannot state an order the code does not follow (it
// did: the help said archive, oa, metadata, arxiv while this ran arxiv first).
func AutoOrder() []string {
	return []string{ViaArxiv, ViaArchive, ViaOA, ViaMetadata}
}

// Attempt is one backend's answer: the bytes it got, and what a citation is entitled to say
// about them. TextRetrieved false means NO TEXT WAS FETCHED — a record that the source exists,
// which is `source_text_read: unread` and must never be cited as a reading.
type Attempt struct {
	Body          []byte
	ContentType   string
	Via           string
	TextRetrieved bool
}

// ---------- identifiers ----------

var (
	doiRe   = regexp.MustCompile(`10\.\d{4,9}/[^\s"'<>&?#]+`)
	arxivRe = regexp.MustCompile(`arxiv\.org/(?:abs|pdf|e-print)/([0-9]{4}\.[0-9]{4,5}(?:v[0-9]+)?)`)
)

// DOIOf pulls a DOI out of a url — doi.org links, and publisher urls that embed one. Empty when
// there is none, which is an ordinary answer: most web pages are not articles.
func DOIOf(rawURL string) string {
	m := doiRe.FindString(rawURL)
	return strings.TrimRight(m, ".,;)")
}

// ArxivIDOf pulls an arXiv identifier out of a url.
func ArxivIDOf(rawURL string) string {
	if m := arxivRe.FindStringSubmatch(strings.ToLower(rawURL)); m != nil {
		return m[1]
	}
	return ""
}

// ---------- archive (CDX) ----------

// Capture is one Wayback capture, as CDX reports it.
type Capture struct {
	Timestamp string
	Original  string
	Digest    string
}

// SnapshotURL is where the archived bytes live.
func (c Capture) SnapshotURL() string {
	return "https://web.archive.org/web/" + c.Timestamp + "id_/" + c.Original
}

// CapturesFor lists a url's captures, newest last.
//
// CDX RATHER THAN THE AVAILABILITY API, and the difference is not cosmetic. `availability`
// answers with ONE "closest" capture and no way to ask for another; CDX returns every capture
// with its timestamp, status and digest, so a caller can take the EARLIEST — which is what the
// measured run needed, because its load-bearing evidence was a snapshot 146 days BEFORE the
// preprint, and "closest to now" can never produce that. An earlier draft of this file used the
// weaker endpoint and was therefore below the capability the seats had by hand.
func CapturesFor(f Fetcher, rawURL string) ([]Capture, error) {
	q := "https://web.archive.org/cdx/search/cdx?output=json&fl=timestamp,original,digest" +
		"&filter=statuscode:200&collapse=digest&limit=200&url=" + url.QueryEscape(strings.TrimPrefix(strings.TrimPrefix(rawURL, "https://"), "http://"))
	resp, err := f.Fetch(q)
	if err != nil {
		return nil, nil // the archive being unreachable is a fact about this container, not a failure of the source
	}
	var rows [][]string
	if json.Unmarshal(resp.Body, &rows) != nil || len(rows) < 2 {
		return nil, nil
	}
	out := make([]Capture, 0, len(rows)-1)
	for _, r := range rows[1:] { // row 0 is the header
		if len(r) >= 3 {
			out = append(out, Capture{Timestamp: r[0], Original: r[1], Digest: r[2]})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Timestamp < out[j].Timestamp })
	return out, nil
}

// PickCapture chooses among captures. `at` is a YYYYMMDD bound: the LATEST capture at or before
// it, which is how "what did this say before date X" is answered. Empty `at` takes the earliest,
// because a priority question wants the first time a thing was visible, not the most recent.
func PickCapture(cs []Capture, at string) (Capture, bool) {
	if len(cs) == 0 {
		return Capture{}, false
	}
	if at == "" {
		return cs[0], true
	}
	best, ok := Capture{}, false
	for _, c := range cs {
		if len(c.Timestamp) >= 8 && c.Timestamp[:8] <= at {
			best, ok = c, true
		}
	}
	return best, ok
}

// ---------- open access ----------

// OpenAccessURL asks whether a legal open copy exists, and where. The second return says whether
// the question was ANSWERED: false means no backend could speak to it, which is different from
// an answered "there is none" and must not be reported as one.
func OpenAccessURL(f Fetcher, doi string) (loc string, answered bool) {
	if doi == "" {
		return "", false
	}
	if resp, err := f.Fetch("https://api.unpaywall.org/v2/" + doi + "?email=feov@ctoforaday.com"); err == nil {
		var u struct {
			IsOA           bool `json:"is_oa"`
			BestOALocation *struct {
				URLForPDF string `json:"url_for_pdf"`
				URL       string `json:"url"`
			} `json:"best_oa_location"`
		}
		if json.Unmarshal(resp.Body, &u) == nil {
			if u.BestOALocation != nil {
				if u.BestOALocation.URLForPDF != "" {
					return u.BestOALocation.URLForPDF, true
				}
				return u.BestOALocation.URL, true
			}
			return "", true // answered: there is no open copy
		}
	}
	if resp, err := f.Fetch("https://api.openalex.org/works/doi:" + doi); err == nil {
		var w struct {
			OpenAccess struct {
				IsOA  bool   `json:"is_oa"`
				OAURL string `json:"oa_url"`
			} `json:"open_access"`
		}
		if json.Unmarshal(resp.Body, &w) == nil {
			return w.OpenAccess.OAURL, true
		}
	}
	return "", false
}

// ---------- metadata ----------

// MetadataRecord fetches the bibliographic record: what the source IS, without its text.
//
// THIS IS THE BACKEND THAT MAKES "UNREAD" HONEST. A run that cannot get an article can still
// establish that it exists, in which venue, at which pages — which is a real finding and the
// exact thing `source_text_read: unread` was added to carry. The run proved the route by hand:
// when five full-text hosts refused, Crossref and ERIC returned registered records.
func MetadataRecord(f Fetcher, doi string) (*Attempt, error) {
	if doi == "" {
		return nil, nil
	}
	resp, err := f.Fetch("https://api.crossref.org/works/" + doi)
	if err != nil {
		return nil, nil
	}
	var cr struct {
		Message struct {
			Title          []string `json:"title"`
			ContainerTitle []string `json:"container-title"`
			Page           string   `json:"page"`
			Issued         struct {
				DateParts [][]int `json:"date-parts"`
			} `json:"issued"`
		} `json:"message"`
	}
	if json.Unmarshal(resp.Body, &cr) != nil || len(cr.Message.Title) == 0 {
		return nil, nil
	}
	year := ""
	if p := cr.Message.Issued.DateParts; len(p) > 0 && len(p[0]) > 0 {
		year = fmt.Sprint(p[0][0])
	}
	venue := ""
	if len(cr.Message.ContainerTitle) > 0 {
		venue = cr.Message.ContainerTitle[0]
	}
	return &Attempt{
		Body:        resp.Body,
		ContentType: "application/json",
		Via: fmt.Sprintf("bibliographic record only (Crossref, doi %s): %q, %s %s, pp. %s — THE TEXT WAS NOT RETRIEVED",
			doi, cr.Message.Title[0], venue, year, cr.Message.Page),
		TextRetrieved: false,
	}, nil
}

// ---------- arxiv ----------

// ArxivURLs gives the abstract page, the PDF and the LaTeX source for an arXiv id. The e-print
// form is the one the retired `arxiv-latex` tooling existed to reach, and it needs no MCP server.
func ArxivURLs(id string) (abs, pdf, eprint string) {
	return "https://arxiv.org/abs/" + id,
		"https://arxiv.org/pdf/" + id,
		"https://arxiv.org/e-print/" + id
}

// Recover runs one backend, or tries them in order under ViaAuto, and returns the first answer
// that carries bytes. nil means nothing spoke — which leaves the caller's ORIGINAL refusal
// standing, because replacing it would hide which of the two actually happened.
//
// THE AUTO ORDER IS THE ORDER OF DIMINISHING CLAIM. The archive may hold the document itself; an
// open-access copy is the document; a bibliographic record is not the document at all and says so.
// Ending on metadata means a run that cannot read a source still learns whether the source EXISTS,
// which is the difference between "I could not reach it" and "there is nothing to reach".
func Recover(f Fetcher, rawURL, via, at string) *Attempt {
	doi := DOIOf(rawURL)
	try := func(name string) *Attempt {
		switch name {
		case ViaArchive:
			caps, _ := CapturesFor(f, rawURL)
			c, ok := PickCapture(caps, at)
			if !ok {
				return nil
			}
			resp, err := f.Fetch(c.SnapshotURL())
			if err != nil {
				return nil
			}
			when := c.Timestamp
			if len(when) >= 8 {
				when = fmt.Sprintf("%s-%s-%s", when[0:4], when[4:6], when[6:8])
			}
			return &Attempt{Body: resp.Body, ContentType: MediaType(resp.ContentType), TextRetrieved: true,
				Via: fmt.Sprintf("archive.org capture of %s (%s) — a snapshot, NOT the live source; for a subscription "+
					"article this is usually the landing page and not the text, so read it before citing it as read",
					when, c.SnapshotURL())}
		case ViaOA:
			loc, answered := OpenAccessURL(f, doi)
			if loc == "" {
				if answered && doi != "" {
					// AN ANSWERED "NO" IS A FINDING. It is the determinate half of #736: not
					// "unreachable from this container" but "no open copy exists anywhere".
					return &Attempt{Body: []byte(fmt.Sprintf("no open-access copy of doi %s exists\n", doi)),
						ContentType: "text/plain", TextRetrieved: false,
						Via: fmt.Sprintf("open-access lookup for doi %s: NO OPEN COPY EXISTS anywhere the OA indexes know of. "+
							"This is a fact about the WORLD, not about this container — do not record it as unreachable-from-here", doi)}
				}
				return nil
			}
			resp, err := f.Fetch(loc)
			if err != nil {
				return nil
			}
			return &Attempt{Body: resp.Body, ContentType: MediaType(resp.ContentType), TextRetrieved: true,
				Via: "open-access copy, located by doi " + doi + " at " + loc}
		case ViaArxiv:
			id := ArxivIDOf(rawURL)
			if id == "" {
				return nil
			}
			_, pdf, _ := ArxivURLs(id)
			resp, err := f.Fetch(pdf)
			if err != nil {
				return nil
			}
			return &Attempt{Body: resp.Body, ContentType: MediaType(resp.ContentType), TextRetrieved: true,
				Via: "arXiv " + id + " (" + pdf + "); the LaTeX source is at " + mustEprint(id)}
		case ViaMetadata:
			att, _ := MetadataRecord(f, doi)
			return att
		case ViaEric:
			att, _ := EricRecord(f, rawURL)
			return att
		}
		return nil
	}
	if via != ViaAuto && via != "" {
		return try(via)
	}
	for _, name := range AutoOrder() {
		if a := try(name); a != nil {
			return a
		}
	}
	return nil
}

func mustEprint(id string) string { _, _, e := ArxivURLs(id); return e }

// EricRecord answers from ERIC, the US Department of Education's index — the education literature,
// which none of the other backends covers.
//
// # Why it earns a backend rather than being folded into `metadata`
//
// Crossref, OpenAlex and Unpaywall are DOI-generic and OA-generic. ERIC is a CORPUS, and it holds
// things they do not. Measured on the Savage record this file already cites: Crossref returns the
// title and pages and NO AUTHOR AT ALL; ERIC returns `["Oliver, June", "Savage, John"]` and a
// descriptive abstract naming the setting ("factoring quadratics in a 10th-grade enrichment
// class"). A run cited it as a sole-authored article. It is a shared two-page column, and no
// existing backend could have said so — which changes how much weight a novelty argument can put
// on it.
//
// # What it can and cannot deliver
//
// 527,037 of its 2,116,390 records carry authorised full text at files.eric.ed.gov; the rest do
// not, and ERIC SAYS WHICH — `e_fulltextauth` is a field, so "there is no free copy" is a fact it
// states rather than a fetch that failed. Where the text exists this returns it; where it does
// not, this returns the record and says the text was not retrieved, which is the same contract
// `metadata` keeps.
//
// It is deliberately NOT in the `auto` chain. `auto` walks backends that answer for any source;
// asking a US education index about a physics preprint wastes a request to be told no, and a seat
// that wants this corpus knows it wants this corpus.
func EricRecord(f Fetcher, rawURL string) (*Attempt, error) {
	q := ericQueryOf(rawURL)
	if q == "" {
		return nil, nil
	}
	resp, err := f.Fetch("https://api.ies.ed.gov/eric/?search=" + url.QueryEscape(q) +
		"&format=json&rows=1&fields=id,title,author,description,source,publicationdateyear,e_fulltextauth")
	if err != nil {
		return nil, nil
	}
	var er struct {
		Response struct {
			NumFound int `json:"numFound"`
			Docs     []struct {
				ID           string   `json:"id"`
				Title        string   `json:"title"`
				Author       []string `json:"author"`
				Description  string   `json:"description"`
				Source       string   `json:"source"`
				Year         int      `json:"publicationdateyear"`
				FullTextAuth int      `json:"e_fulltextauth"`
			} `json:"docs"`
		} `json:"response"`
	}
	if json.Unmarshal(resp.Body, &er) != nil || er.Response.NumFound == 0 || len(er.Response.Docs) == 0 {
		return nil, nil
	}
	d := er.Response.Docs[0]

	// THE FULL TEXT WHERE ERIC SAYS IT HAS IT. The flag is the record's own answer, so a miss here
	// is a broken link rather than an absent document — and it is reported as the record, not as a
	// silent failure to fetch.
	if d.FullTextAuth == 1 {
		if pdf, perr := f.Fetch("https://files.eric.ed.gov/fulltext/" + d.ID + ".pdf"); perr == nil && len(pdf.Body) > 0 {
			return &Attempt{
				Body: pdf.Body, ContentType: "application/pdf", TextRetrieved: true,
				Via: fmt.Sprintf("ERIC %s (https://files.eric.ed.gov/fulltext/%s.pdf): %q, %s %d — full text, authorised by ERIC",
					d.ID, d.ID, d.Title, d.Source, d.Year),
			}, nil
		}
	}
	who := "no author on the record"
	if len(d.Author) > 0 {
		who = strings.Join(d.Author, "; ")
	}
	return &Attempt{
		Body:        resp.Body,
		ContentType: "application/json",
		Via: fmt.Sprintf("bibliographic record only (ERIC %s): %q — %s, %s %d. %s "+
			"ERIC states it holds NO authorised full text for this record, which is a fact about the "+
			"world and not about this container — THE TEXT WAS NOT RETRIEVED",
			d.ID, d.Title, who, d.Source, d.Year, strings.TrimSpace(d.Description)),
	}, nil
}

// ericQueryOf turns a url into an ERIC search: the accession number where the url carries one,
// otherwise the title-ish tail of the path.
//
// A DOI IS DELIBERATELY NOT USED, and this was measured rather than reasoned. ERIC exposes no DOI
// field, so a DOI can only go in as free text — and free text matches a DOI CITED INSIDE another
// record. Asked for 10.5951/MT.82.1.0035 (the Savage column, 1989) it returned "How an
// Inquiry-Oriented Textbook Shaped a Calculus Instructor's Planning" (2022), whose abstract cites
// a different 10.5951 DOI, and the backend would have presented that as the source. A confident
// wrong paper is worse here than no answer: the seat has no way to tell, and the whole point of
// this surface is to say what a source says.
//
// IT RETURNS "" RATHER THAN GUESSING WIDELY, for the same reason. A query built from a bare
// hostname would match thousands of records and the first would come back as though it were the
// source.
func ericQueryOf(rawURL string) string {
	// url.Parse ACCEPTS ALMOST ANYTHING — "not a url at all" parses as a relative path, and its
	// words would then be sent as a title query. A fetch target is an absolute URL; anything else
	// is not a source this can look up.
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return ""
	}
	if m := ericIDRe.FindStringSubmatch(rawURL); m != nil {
		return "id:" + m[1]
	}
	if DOIOf(rawURL) != "" {
		return "" // see above: a DOI can only go in as free text, and free text is not identity
	}
	seg := strings.Trim(u.Path, "/")
	if i := strings.LastIndex(seg, "/"); i >= 0 {
		seg = seg[i+1:]
	}
	seg = strings.TrimSuffix(strings.TrimSuffix(seg, ".html"), ".pdf")
	seg = strings.NewReplacer("-", " ", "_", " ").Replace(seg)
	if len(strings.Fields(seg)) < 2 {
		return "" // one word is not a title; it would match the corpus, not the source
	}
	return seg
}

// ericIDRe recognises an ERIC accession number, the identity this corpus is keyed on.
var ericIDRe = regexp.MustCompile(`\b(E[JD]\d{6,})\b`)
