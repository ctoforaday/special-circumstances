package fetchcache

import (
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

// THE SOURCES THAT CONTRIBUTE LOCATIONS, AND WHY EACH IS A DIFFERENT OPINION.
//
// One index's list is not the world: eight works whose publisher pdf_url answered 403 were
// re-checked by hand and four carried a PubMed Central copy the first index never surfaced. The
// lesson generalises past that one case — what matters is not which index is best but that each
// is a DIFFERENT ORGANISATION with its own crawl. Unpaywall was dropped for failing exactly that
// test: it is a wrapper over OpenAlex and contributed 0 unique urls across 30 works.
//
// Each source here answers (locations, answered). `answered` false means the service did not
// speak — a timeout, a 5xx, a body that is not what it should be — and it is never the same as
// "this service knows of no copy". The distinction is load-bearing: the caller assembles a
// determinate "no open copy exists anywhere" only when EVERY source answered, because that claim
// cannot be built out of a timeout.

// semanticScholarCandidates asks Semantic Scholar, which runs its own crawl and its own PDF
// discovery rather than deriving from Crossref.
//
// MEASURED VALUE: MODEST, AND WORTH SAYING SO. On the eight works that had defeated the other
// route it mostly echoed the same publisher urls that had already refused us. It is included
// because it is genuinely independent and free, not because it carried that test.
func semanticScholarCandidates(f Fetcher, doi string) (locs []string, answered bool) {
	if doi == "" {
		return nil, true
	}
	resp, err := f.Fetch("https://api.semanticscholar.org/graph/v1/paper/DOI:" +
		url.PathEscape(doi) + "?fields=openAccessPdf,externalIds")
	if err != nil {
		return nil, false
	}
	var r struct {
		PaperID       string `json:"paperId"` // the discriminator: an error body decodes without it
		OpenAccessPDF *struct {
			URL string `json:"url"`
		} `json:"openAccessPdf"`
		ExternalIDs map[string]any `json:"externalIds"`
	}
	if json.Unmarshal(resp.Body, &r) != nil {
		return nil, false
	}
	// A 404 for a doi it does not hold is an ANSWER — it knows of no copy — and arrives as a
	// refusal above, so reaching here with no paperId means the shape moved.
	if r.PaperID == "" {
		return nil, false
	}
	if r.OpenAccessPDF != nil && r.OpenAccessPDF.URL != "" {
		locs = append(locs, r.OpenAccessPDF.URL)
	}
	return locs, true
}

// doajCandidates asks the Directory of Open Access Journals, which indexes peer-reviewed open
// journals NATIVELY — a publisher applies and is assessed — rather than inferring openness from
// a crawl. That makes it a different KIND of evidence from the others, not merely another crawl:
// where DOAJ lists a full-text url, a human has checked that the journal is open.
func doajCandidates(f Fetcher, doi string) (locs []string, answered bool) {
	if doi == "" {
		return nil, true
	}
	resp, err := f.Fetch("https://doaj.org/api/search/articles/doi:" + url.PathEscape(doi))
	if err != nil {
		return nil, false
	}
	var r struct {
		Total   *int `json:"total"` // present even on a zero-result answer; absent on an error body
		Results []struct {
			Bibjson struct {
				Link []struct {
					Type        string `json:"type"`
					URL         string `json:"url"`
					ContentType string `json:"content_type"`
				} `json:"link"`
			} `json:"bibjson"`
		} `json:"results"`
	}
	if json.Unmarshal(resp.Body, &r) != nil || r.Total == nil {
		return nil, false
	}
	for _, res := range r.Results {
		for _, l := range res.Bibjson.Link {
			if l.Type == "fulltext" && l.URL != "" {
				locs = append(locs, l.URL)
			}
		}
	}
	return locs, true
}

// citationMetaRe reads a `citation_*_url` meta tag, in either attribute order, for whichever name
// it is handed.
//
// THE PDF IS NOT THE GOAL — the CONTENT is. A publisher that serves its full text as html and its
// abstract as another page says both in these tags, and taking only `citation_pdf_url` skipped
// the case where the readable copy is the html one. The names are Google Scholar's inclusion
// vocabulary, which is why they are almost universally present.
func citationMetaRe(name string) *regexp.Regexp {
	n := regexp.QuoteMeta(name)
	return regexp.MustCompile(`(?is)<meta[^>]+(?:name\s*=\s*["']?` + n + `["']?[^>]*content\s*=\s*["']([^"']+)["']` +
		`|content\s*=\s*["']([^"']+)["'][^>]*name\s*=\s*["']?` + n + `["']?)`)
}

// fullTextMetaNames are the pointers a landing page may carry to its own readable copy, most
// specific first.
//
// MEASURED 2026-09-24 over the pages this tool fetched and recorded as documents: of 16 that
// turned out to be abstract or landing pages rather than the paper, 12 carried one of these.
// The publisher states where its full text is; nothing here has to guess.
var fullTextMetaNames = []string{
	"citation_fulltext_html_url",
	"citation_full_html_url",
	"citation_xml_url",
	"citation_pdf_url",
}

// LandingPageFullText reads a landing page's own statement of where its READABLE COPY is.
//
// It was LandingPagePDF and looked only at `citation_pdf_url`. The pdf is not the goal — the
// content is, in whatever form arrives — and a page whose full text is html said so in a tag this
// never read.
//
// THE PAGE IN FRONT OF US IS A SOURCE OF LOCATIONS TOO, and the cheapest one: it costs no request
// at all, because the bytes are already in hand. Google Scholar's inclusion guidelines are why
// the tag is so widely present — publishers emit `citation_pdf_url` to be indexed, and it points
// at the file rather than the viewer.
//
// It is also the one signal a publisher cannot be wrong about. An index may hold a stale url; the
// page saying "my pdf is here" was generated by the system serving it.
//
// Signposting (`Link: rel="item"`) is the standards-track version of the same idea and is read
// from the header where a host sends one — repository platforms do, commercial publishers do not.
func LandingPageFullText(contentType string, body []byte, linkHeader string, base *url.URL) string {
	if base == nil {
		return ""
	}
	if rel := signpostItem(linkHeader); rel != "" {
		if u := resolveWeb(base, rel); u != "" {
			return u
		}
	}
	if !strings.Contains(strings.ToLower(contentType), "html") {
		return ""
	}
	for _, name := range fullTextMetaNames {
		m := citationMetaRe(name).FindSubmatch(body)
		if m == nil {
			continue
		}
		raw := string(m[1])
		if raw == "" {
			raw = string(m[2])
		}
		u := resolveWeb(base, strings.TrimSpace(raw))
		// A POINTER AT THE PAGE WE ARE ALREADY ON IS NOT A LEAD. Nature emits
		// `citation_fulltext_html_url` naming the very url just fetched; following it would
		// re-fetch the same abstract and spend a paced request to learn nothing.
		if u == "" || u == base.String() {
			continue
		}
		return u
	}
	return ""
}

// signpostItem pulls the FAIR Signposting `rel="item"` target out of a Link header, preferring a
// pdf where the header types its links.
func signpostItem(header string) string {
	if header == "" {
		return ""
	}
	for _, part := range strings.Split(header, ",") {
		if !strings.Contains(strings.ToLower(part), `rel="item"`) && !strings.Contains(strings.ToLower(part), "rel=item") {
			continue
		}
		open, close := strings.Index(part, "<"), strings.Index(part, ">")
		if open < 0 || close <= open {
			continue
		}
		return strings.TrimSpace(part[open+1 : close])
	}
	return ""
}

func resolveWeb(base *url.URL, ref string) string {
	if ref == "" {
		return ""
	}
	r, err := url.Parse(ref)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(r)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

// ---------- the registered target of a doi ----------

// doiResolverHosts are the DOI resolver's own names. A url on one of them is an IDENTIFIER
// pointing at a publisher page, not a document.
var doiResolverHosts = map[string]bool{
	"doi.org": true, "dx.doi.org": true, "www.doi.org": true,
}

// crossrefWorks is where a work record is asked for. ONE HOME, because two verbs ask Crossref the
// same question about the same doi and a second spelling of the endpoint is a second thing to get
// wrong. It is a var so a test can point the whole path at a fixture server; nothing else writes it.
var crossrefWorks = "https://api.crossref.org/works/"

// RegisteredTarget asks Crossref where a doi's publisher page IS, so the fetch can go straight
// there instead of through the resolver.
//
// WHY NOT JUST FOLLOW THE REDIRECT. doi.org publishes no robots.txt and no rate headers, so this
// tool's kind default applies to it: fifteen seconds, jittered, per fetch, shared across every
// process. Measured on the source sweep, 2,564 of 2,599 urls were doi.org urls — the whole scan
// ran at one host's floor and projected to 41 hours, nearly all of it waiting on a redirect
// service. Crossref answers the same question from a table, at 200ms, and it is an API built to
// be asked. The registered target is also a REAL host, so its own robots.txt and its own floor
// apply to the fetch that follows, which the resolver's hop obscured.
//
// A MISS IS THE RESOLVER'S CUE, NOT AN ERROR. Where Crossref does not answer — no record, an
// error envelope, a doi it does not mint — this returns "" and the caller follows the doi.org
// redirect exactly as before. Measured over 300 corpus works, `resource.primary.URL` was present
// on 298; the two absences are the case this fallback is for.
func RegisteredTarget(f Fetcher, rawURL string) string {
	u, err := url.Parse(rawURL)
	// Host() rather than Hostname() as a second key so a fixture server on a port can stand in
	// for the resolver; a real doi.org url carries no port and matches on the bare name.
	if err != nil || !(doiResolverHosts[strings.ToLower(u.Hostname())] || doiResolverHosts[strings.ToLower(u.Host)]) {
		return ""
	}
	doi := DOIOf(rawURL)
	if doi == "" {
		return ""
	}
	resp, err := f.Fetch(crossrefWorks + doi + "?mailto=" + ContactEmail)
	if err != nil {
		return ""
	}
	var cr struct {
		Message struct {
			DOI      string `json:"DOI"` // the discriminator: an error envelope decodes without it
			Resource struct {
				Primary struct {
					URL string `json:"URL"`
				} `json:"primary"`
			} `json:"resource"`
		} `json:"message"`
	}
	if json.Unmarshal(resp.Body, &cr) != nil || cr.Message.DOI == "" {
		return ""
	}
	target := strings.TrimSpace(cr.Message.Resource.Primary.URL)
	// A TARGET THAT IS ITSELF A DOI URL BUYS NOTHING and would send the next hop back to the
	// resolver — the loop this exists to leave.
	if t, terr := url.Parse(target); terr != nil || t.Scheme == "" || doiResolverHosts[strings.ToLower(t.Hostname())] {
		return ""
	}
	return target
}

// ---------- the publisher's own machine-readable copy ----------

// keyOnlyTDMHosts are the text-mining endpoints that answer nothing without an API key.
//
// MEASURED, NOT ASSUMED: 55 Crossref-registered text-mining links were fetched on 2026-09-23,
// five per host, interleaved. api.elsevier.com returned HTTP 400 five times out of five, and
// api.wiley.com likewise. No other host in the sample refused for that reason.
//
// They are excluded rather than merely allowed to fail because of their SHARE. Across 300 corpus
// works those two hosts carry 133 of the 136 registered text-mining links between them, so trying
// them costs a request per work, on works whose publisher has already been asked once, to be told
// no by a host that will always say no. Being a good citizen means not making that request.
//
// A KEY WOULD CHANGE THIS AND IS NOT OURS TO GET (see #1129): registering for a publisher API
// credential makes this tool's traffic somebody's account.
var keyOnlyTDMHosts = map[string]bool{
	"api.elsevier.com": true,
	"api.wiley.com":    true,
}

// crossrefTextMiningCandidates returns the full-text urls a PUBLISHER REGISTERED for machine
// reading, which is the one location source that is the publisher's own statement rather than a
// third party's crawl.
//
// WHAT IT IS WORTH, MEASURED. 136 of 300 corpus works (45%) carry at least one such link, and 79
// of those works are ones OpenAlex calls CLOSED — a sanctioned route to full text on papers the
// open-access indexes have nothing for. Of 45 links fetched outside the two key-only hosts, 19
// yielded a readable document (42%), and the failures are per-host rather than scattered: three
// hosts answered every time, and the rest fail for reasons that are facts about the host.
//
// `similarity-checking` links are NOT taken. They are registered for plagiarism services under
// separate agreements, and the intent recorded on the link is the publisher's statement of what
// it is offering; reading one as an invitation to fetch is helping ourselves to a different
// permission from the one that was given.
func crossrefTextMiningCandidates(f Fetcher, doi string) (locs []string, answered bool) {
	if doi == "" {
		return nil, true
	}
	resp, err := f.Fetch(crossrefWorks + doi + "?mailto=" + ContactEmail)
	if err != nil {
		return nil, false
	}
	var cr struct {
		Message struct {
			DOI  string `json:"DOI"` // the discriminator: an error envelope decodes without it
			Link []struct {
				URL                 string `json:"URL"`
				ContentType         string `json:"content-type"`
				IntendedApplication string `json:"intended-application"`
			} `json:"link"`
		} `json:"message"`
	}
	if json.Unmarshal(resp.Body, &cr) != nil || cr.Message.DOI == "" {
		return nil, false
	}
	for _, l := range cr.Message.Link {
		if l.IntendedApplication != "text-mining" || l.URL == "" {
			continue
		}
		u, perr := url.Parse(l.URL)
		if perr != nil || keyOnlyTDMHosts[strings.ToLower(u.Hostname())] {
			continue
		}
		locs = append(locs, l.URL)
	}
	return locs, true
}

// notAPDFHosts are hosts whose pages an index sometimes files under `pdf_url` and which never
// serve a pdf there. They are demoted to landing pages rather than dropped: the page may still
// name the real pdf in its own `citation_pdf_url`, and dropping a location loses that lead.
//
// MEASURED: across ten works where an index said a free copy existed and this tool returned none,
// OpenAlex listed a `pubmed.ncbi.nlm.nih.gov/<pmid>` abstract page as `pdf_url` on five of them.
// Each one occupied a slot in a four-candidate budget ahead of a location that would have worked.
var notAPDFHosts = map[string]bool{
	"pubmed.ncbi.nlm.nih.gov": true,
	"www.ncbi.nlm.nih.gov":    true, // the browser route into PMC, and the one behind the challenge
	"europepmc.org":           false,
}

// looksLikePDF is the fallback for a source that does NOT type its locations — which is every one
// of them except Europe PMC. It is a guess about a url and is labelled as one, because the
// alternative is pretending a suffix is a content type.
func looksLikePDF(u string) bool {
	p, err := url.Parse(u)
	if err != nil {
		return false
	}
	if deny, known := notAPDFHosts[strings.ToLower(p.Hostname())]; known && deny {
		return false
	}
	// `?pdf=render` and `/pdfdirect/…` are pdf routes that do not end in `.pdf`; the suffix test
	// alone called them landing pages and sorted them behind every guessed pdf.
	lower := strings.ToLower(u)
	return strings.HasSuffix(p.Path, ".pdf") ||
		strings.Contains(lower, "pdf=render") ||
		strings.Contains(p.Path, "/pdfdirect/") ||
		strings.HasSuffix(p.Path, "/pdf")
}
