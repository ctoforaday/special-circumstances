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
//	arxiv     the preprint itself: abstract, PDF, LaTeX source, or arXiv's HTML rendering. The
//	          forms differ in size by an order of magnitude in BOTH directions, so the backend
//	          probes rather than picks — see ArxivURLs.
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
	Body        []byte
	ContentType string
	Via         string
	// Backend is the Via constant that produced this answer — the fact a reader branches on,
	// where Via is the sentence a seat reads. Recover sets it; a backend never does.
	Backend       string
	TextRetrieved bool
}

// ---------- identifiers ----------

var (
	doiRe = regexp.MustCompile(`10\.\d{4,9}/[^\s"'<>&?#]+`)
	// BOTH arXiv SCHEMES, and the old one is the one that carries the literature. `YYMM.NNNNN`
	// dates from April 2007; before it, identifiers were `archive/YYMMNNN` with an optional
	// subject class — `hep-th/9711200`, `cond-mat.stat-mech/0605194`. Measured on INSPIRE, 75 of
	// the 100 most-cited hep-th papers carry an old-scheme identifier and the top nine are all
	// old, so a pattern that knows only the new scheme answers "no arXiv id here" for the most
	// cited paper in the field and the `arxiv` backend returns nothing for it. Defunct archives
	// (`q-alg`, `alg-geom`) still resolve, hence `[a-z-]+` rather than a list. Lowercasing the
	// url first is safe because arxiv.org serves `math.ag/0611800` and `math.AG/0611800` alike.
	arxivRe = regexp.MustCompile(`arxiv\.org/(?:abs|pdf|e-print)/([0-9]{4}\.[0-9]{4,5}(?:v[0-9]+)?|[a-z-]+(?:\.[a-z-]+)?/[0-9]{7}(?:v[0-9]+)?)`)
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
		return nil, fmt.Errorf("the archive did not answer: %w", err)
	}
	// AN ARCHIVE THAT CANNOT ANSWER IS NOT AN ARCHIVE WITH NOTHING IN IT, and until this
	// distinction existed the two were the same bytes. Measured live on 2026-09-22: CDX served
	// `<title>Internet Archive: Temporarily Offline</title>` as HTML with a 200, this function
	// failed to unmarshal it, returned no captures and no error, and a seat was told the url had
	// never been archived. The honest zero and the outage are indistinguishable at exactly the
	// moment the outage is most likely — under the load that caused it.
	var rows [][]string
	if uerr := json.Unmarshal(resp.Body, &rows); uerr != nil {
		head := strings.TrimSpace(string(resp.Body))
		if len(head) > 120 {
			head = head[:120] + "…"
		}
		return nil, fmt.Errorf("the archive answered with something that is not a capture list "+
			"(%d bytes, %s): %s", len(resp.Body), ContentTypeOrUnknown(resp.ContentType), head)
	}
	// A WELL-FORMED EMPTY LIST IS THE HONEST ZERO, and it keeps the nil error: CDX returns just
	// the header row, or nothing at all, for a url it has never captured.
	if len(rows) < 2 {
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

// OpenAccessCandidates gathers EVERY location the open-access indexes know of, ranked, rather
// than the single one each index nominates.
//
// WHY A LIST AND NOT A BEST. Both providers publish a chosen location and a full list, and this
// backend read only the chosen one — Unpaywall's `best_oa_location`, or OpenAlex's `oa_url`.
// Measured over the sweep's own failures: of the fetches that ended at a wall or a bibliographic
// record, 48% had a fetchable `pdf_url` sitting in a list neither read path touched. Of those
// candidates 4 in 14 actually returned a PDF at the leaf, the rest being publisher platforms
// advertising a PDF they then refuse — so the list is worth trying and worth VERIFYING, and
// neither is worth doing from a single nomination.
//
// ONE REQUEST, NOT TWO, AND UNPAYWALL IS NOT THE ONE. This function used to ask Unpaywall first
// and OpenAlex second, on the belief that two indexes were two opinions. They are one:
// OurResearch states that "Unpaywall is kept as a legacy-compatible wrapper and format over
// OpenAlex", and measured across five works its `oa_locations` is a STRICT SUBSET of OpenAlex's
// `locations` every time — 0 urls unique to Unpaywall, and 2 to 33 unique to OpenAlex. So the
// second call added no location this one does not have, and cost a request against a service
// that asks callers to limit their use.
//
// The consequence for the caller matters more than the saved request. `answered` licenses the
// determinate "no open copy exists anywhere", and while two calls were being made it looked as
// though that claim rested on two silences. It never did. It rests on OpenAlex, which is why the
// `ID` discriminator below is load-bearing rather than belt-and-braces. A GENUINE second opinion
// has to come from a different organisation — Semantic Scholar, CORE or DOAJ — and none is
// consulted here yet.
func OpenAccessCandidates(f Fetcher, doi string) (locs []string, answered bool) {
	var indexFailed bool
	if doi == "" {
		return nil, false
	}
	seen := map[string]bool{}
	var pdfs, pages []string
	add := func(u string, isPDF bool) {
		u = strings.TrimSpace(u)
		// A doi.org url is an identifier, not a copy: following it returns to the publisher this
		// lookup exists to get around.
		if u == "" || seen[u] || strings.Contains(u, "doi.org/") {
			return
		}
		seen[u] = true
		if isPDF {
			pdfs = append(pdfs, u)
			return
		}
		pages = append(pages, u)
	}

	if resp, err := f.Fetch("https://api.openalex.org/works/doi:" + doi + "?mailto=" + ContactEmail); err == nil {
		var w struct {
			// ID IS THE DISCRIMINATOR, AND WITHOUT IT THIS DECODE CANNOT FAIL. A struct carrying
			// only the fields we want unmarshals successfully from ANY json object — an error
			// envelope, a metering message, a response whose shape has moved — and the zero value
			// is indistinguishable from a work with no open copy. The caller turns the second into
			// a fact about the WORLD.
			ID         string `json:"id"`
			OpenAccess struct {
				OAURL string `json:"oa_url"`
			} `json:"open_access"`
			Locations []struct {
				PDFURL  string `json:"pdf_url"`
				Landing string `json:"landing_page_url"`
				IsOA    bool   `json:"is_oa"`
			} `json:"locations"`
		}
		if json.Unmarshal(resp.Body, &w) == nil && w.ID != "" {
			answered = true
			add(w.OpenAccess.OAURL, strings.Contains(w.OpenAccess.OAURL, ".pdf"))
			for _, l := range w.Locations {
				add(l.PDFURL, true)
				if l.IsOA {
					add(l.Landing, false)
				}
			}
		}
	}
	// EUROPE PMC, AND THROUGH IT PUBMED CENTRAL, WHICH IS THE ADDITION THAT PAID.
	//
	// Measured on eight works whose OpenAlex pdf_url answered 403 from the publisher that listed
	// it: four carry a PMCID that OpenAlex's locations did not surface, and three of those four
	// serve a real PDF from PMC's open-access bucket — 615 KB, 767 KB, 983 KB. Papers this tool
	// had recorded as unreachable are sitting in an archive whose machine routes NCBI publishes.
	//
	// One index's list is not the world. Taking OpenAlex's alone was the same mistake in miniature
	// as taking Unpaywall's best nomination: a union of everything each index knows, tried and
	// verified, is the only shape that survives a publisher refusing the copy it advertises.
	epmc, epmcOK := europePMCCandidates(f, doi)
	if !epmcOK {
		// AN INDEX THAT DID NOT ANSWER IS NOT AN INDEX WITH NOTHING IN IT. Europe PMC is visibly
		// flaky — it answers a bare nginx 503 under load — and losing it loses the PubMed Central
		// route, which is where the copies a publisher refuses actually live. Measured: the same
		// doi returned a 983 KB PDF on one call and "no open copy exists anywhere" on the next,
		// which is a claim about the WORLD produced by a timeout. Whatever else this function
		// says, it must not say that.
		indexFailed = true
	}
	for _, u := range epmc {
		add(u, strings.HasSuffix(u, ".pdf"))
	}

	// A DIRECT PDF BEFORE A LANDING PAGE, because the landing page is the thing that walls us and
	// the pdf is the thing a citation can be checked against. Beyond that the order is the
	// indexes' own: neither publishes a ranking this could improve on, and the measured predictor
	// of success was not source type — every location that worked in the sample was typed
	// `journal`, including the ones that did not.
	return append(pdfs, pages...), answered && !indexFailed
}

// maxOACandidates bounds how many listed locations one lookup will try. The indexes list up to
// nine; each attempt is a paced request against a different host, and a seat waiting on a ninth
// long-shot has already lost more than the ninth was worth.
const maxOACandidates = 4

// europePMCCandidates asks Europe PMC what it holds for a doi, and turns a PMCID into the
// sanctioned PubMed Central route rather than the browser one.
//
// THE PMC BUCKET IS THE POINT. NCBI names the Cloud Service as one of the "only services that may
// be used for automated retrieval of PMC content", and the `/articles/PMC…/pdf/` path a browser
// would take is not among them — it is the one behind the proof-of-work challenge. So a PMCID is
// resolved to `pmc-oa-opendata.s3.amazonaws.com`, which answers anonymously with no challenge,
// and a miss there is a positive `KeyCount 0` rather than an error.
func europePMCCandidates(f Fetcher, doi string) (locs []string, answered bool) {
	if doi == "" {
		return nil, true // nothing to ask about is not a failure to ask
	}
	resp, err := f.Fetch("https://www.ebi.ac.uk/europepmc/webservices/rest/search?query=DOI:%22" +
		url.QueryEscape(doi) + "%22&resultType=core&format=json&pageSize=1")
	if err != nil {
		return nil, false
	}
	var r struct {
		ResultList struct {
			Result []struct {
				PMCID        string `json:"pmcid"`
				FullTextURLs struct {
					URL []struct {
						Availability  string `json:"availability"`
						DocumentStyle string `json:"documentStyle"`
						URL           string `json:"url"`
					} `json:"fullTextUrl"`
				} `json:"fullTextUrlList"`
			} `json:"result"`
		} `json:"resultList"`
	}
	if json.Unmarshal(resp.Body, &r) != nil {
		return nil, false // a body that is not a result list means the service did not answer
	}
	if len(r.ResultList.Result) == 0 {
		return nil, true // an answered "we hold nothing for this doi"
	}
	rec := r.ResultList.Result[0]
	var out []string
	if rec.PMCID != "" {
		if k := pmcOpenAccessPDF(f, rec.PMCID); k != "" {
			out = append(out, k)
		}
	}
	for _, u := range rec.FullTextURLs.URL {
		// Europe PMC types each location, which is the field that separates a copy we may read
		// from a publisher page that will refuse us.
		if u.Availability == "Free" || u.Availability == "Open access" {
			out = append(out, u.URL)
		}
	}
	return out, true
}

// pmcOpenAccessPDF returns the bucket url for a PMCID's pdf, or "" where the open-access subset
// does not hold one — which the bucket states positively, as a zero key count.
func pmcOpenAccessPDF(f Fetcher, pmcid string) string {
	resp, err := f.Fetch("https://pmc-oa-opendata.s3.amazonaws.com/?list-type=2&max-keys=40&prefix=" +
		url.QueryEscape(pmcid))
	if err != nil {
		return ""
	}
	m := pmcPDFKey.FindSubmatch(resp.Body)
	if m == nil {
		return ""
	}
	return "https://pmc-oa-opendata.s3.amazonaws.com/" + string(m[1])
}

var pmcPDFKey = regexp.MustCompile(`<Key>([^<]+\.pdf)</Key>`)

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
	resp, err := f.Fetch("https://api.crossref.org/works/" + doi + "?mailto=" + ContactEmail)
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

// ArxivURLs gives the abstract page, the PDF, the LaTeX source and the HTML rendering of an
// arXiv id. The e-print form is the one the retired `arxiv-latex` tooling existed to reach, and
// it needs no MCP server.
//
// THE FOUR FORMS ARE NOT INTERCHANGEABLE AND DIFFER IN SIZE BY AN ORDER OF MAGNITUDE EITHER WAY.
// Measured: Maldacena's PDF is 246 KB and its e-print 24 KB, while CLIP's PDF is 6.8 MB and its
// HTML 874 KB. The e-print wins on old TeX-only papers and LOSES on modern figure-heavy ones
// (separate high-resolution figures do not compress the way one PDF does); HTML is the mirror
// image — it exists for recent papers and 404s on every pre-2007 identifier tried. Neither is a
// general answer, which is why the caller probes rather than picks.
func ArxivURLs(id string) (abs, pdf, eprint, html string) {
	return "https://arxiv.org/abs/" + id,
		"https://arxiv.org/pdf/" + id,
		"https://arxiv.org/e-print/" + id,
		"https://arxiv.org/html/" + id
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
	// named SAYS WHETHER THE SEAT ASKED FOR THIS BACKEND BY NAME. It changes what a backend does
	// with a failure it can describe: asked directly, `arxiv` states that the paper is there and
	// this run could not take it, which is the answer to the question that was put. Reached as one
	// rung of `auto`, that same stated failure would END the chain — the seat would never learn
	// that an open-access copy was one rung further down — so under auto it declines instead and
	// the next backend runs. A fact about THIS FETCH must not foreclose the other routes; a fact
	// about the WORLD, as `oa`'s answered "no open copy exists" is, legitimately does.
	try := func(name string, named bool) *Attempt {
		switch name {
		case ViaArchive:
			caps, cerr := CapturesFor(f, rawURL)
			if cerr != nil {
				// THE SAME RULE AS arxiv's STATED FAILURE. Asked for by name, say the archive
				// could not answer, because "no captures" would be a claim about the world drawn
				// from the archive's silence. Reached as a rung of auto, decline so the chain runs
				// on — an outage at one backend must not end the search.
				if !named {
					return nil
				}
				return &Attempt{
					Body:        []byte(fmt.Sprintf("the web archive could not be consulted for %s\n", rawURL)),
					ContentType: "text/plain", TextRetrieved: false,
					Via: fmt.Sprintf("archive lookup for %s: THE ARCHIVE DID NOT ANSWER — %v. This is NOT "+
						"'there are no captures': nothing was learned about whether this url was ever archived, "+
						"and it must not be recorded as absent. Try again later, or ask a different backend", rawURL, cerr),
				}
			}
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
			return &Attempt{Body: resp.Body, ContentType: SniffedMediaType(resp.ContentType, resp.Body), TextRetrieved: true,
				Via: fmt.Sprintf("archive.org capture of %s (%s) — a snapshot, NOT the live source; for a subscription "+
					"article this is usually the landing page and not the text, so read it before citing it as read",
					when, c.SnapshotURL())}
		case ViaOA:
			locs, answered := OpenAccessCandidates(f, doi)
			if len(locs) == 0 {
				if answered && doi != "" {
					// AN ANSWERED "NO" IS A FINDING. It is the determinate half of #736: not
					// "unreachable from this container" but "no open copy exists anywhere". It is
					// now drawn from BOTH indexes' full location lists, so it takes two silences
					// rather than one nomination — checked against the sweep's own verdicts, where
					// 0 of 45 works we called closed had a usable location either index knew of.
					return &Attempt{Body: []byte(fmt.Sprintf("no open-access copy of doi %s exists\n", doi)),
						ContentType: "text/plain", TextRetrieved: false,
						Via: fmt.Sprintf("open-access lookup for doi %s: NO OPEN COPY EXISTS anywhere the OA indexes know of. "+
							"This is a fact about the WORLD, not about this container — do not record it as unreachable-from-here", doi)}
				}
				return nil
			}
			// EVERY CANDIDATE IS TRIED, AND EACH IS VERIFIED AT THE LEAF. An index naming a pdf_url
			// is not a copy: measured, 10 of 14 such urls answered 403 from the publisher platform
			// that published them. So a location counts only once its bytes arrive and pass the
			// same test any other fetch passes.
			var refused []string
			for i, loc := range locs {
				if i >= maxOACandidates {
					break
				}
				resp, ferr := f.Fetch(loc)
				if ferr != nil {
					refused = append(refused, loc)
					continue
				}
				if why := ShellReason(resp.ContentType, resp.Body); why != "" {
					refused = append(refused, loc)
					continue
				}
				return &Attempt{Body: resp.Body, ContentType: SniffedMediaType(resp.ContentType, resp.Body), TextRetrieved: true,
					Via: fmt.Sprintf("open-access copy, located by doi %s at %s (candidate %d of %d the OA indexes listed)",
						doi, loc, i+1, len(locs))}
			}
			// LOCATED BUT BLOCKED IS NOT "NO OPEN COPY". The indexes say a copy exists and named
			// where; this container could not take it from any of them. That is a fact about the
			// fetch, and it is ACTIONABLE — a human with a browser can open these urls — so the
			// urls travel with the refusal rather than being summarised away.
			if !named {
				return nil
			}
			return &Attempt{
				Body:        []byte(strings.Join(refused, "\n") + "\n"),
				ContentType: "text/plain", TextRetrieved: false,
				Via: fmt.Sprintf("open-access lookup for doi %s: A COPY EXISTS AND THIS CONTAINER COULD NOT TAKE IT. "+
					"The indexes list %d location(s) and every one tried refused us or answered with a wall. This is NOT "+
					"'no open copy exists' — do not record the source as closed. The urls are in the body: a human with a "+
					"browser can open them", doi, len(locs)),
			}
		case ViaArxiv:
			id := ArxivIDOf(rawURL)
			if id == "" {
				return nil
			}
			_, pdf, eprint, html := ArxivURLs(id)
			resp, err := f.Fetch(pdf)
			if err == nil {
				return &Attempt{Body: resp.Body, ContentType: SniffedMediaType(resp.ContentType, resp.Body), TextRetrieved: true,
					Via: "arXiv " + id + " (" + pdf + "); the LaTeX source is at " + eprint}
			}
			// A REFUSED PDF IS NOT AN ABSENT PAPER, and returning nil here said it was. arXiv
			// answers 200 for CLIP and the bytes are 6.8 MB, over this tool's cap; the seat then
			// read "that backend has no answer for this url" — the same sentence it gets for a
			// url with no arXiv identifier in it at all. One of the most-cited papers in machine
			// learning was reported as not being on arXiv.
			//
			// So: try the HTML rendering, which for that paper is 874 KB and under the cap, and
			// where even that does not answer, say what happened rather than nothing. The HTML is
			// a DIFFERENT DOCUMENT — it has no pages, so a citation of it cannot name one, and
			// the Via says so where a seat will read it.
			if hr, herr := f.Fetch(html); herr == nil {
				return &Attempt{Body: hr.Body, ContentType: MediaType(hr.ContentType), TextRetrieved: true,
					Via: fmt.Sprintf("arXiv %s as arXiv's OWN HTML RENDERING (%s), because the PDF at %s could not be "+
						"taken: %v. This is not the PDF and HAS NO PAGE NUMBERS — quote it as the HTML, never as a page "+
						"of the paper. The LaTeX source is at %s", id, html, pdf, err, eprint)}
			}
			if !named {
				return nil
			}
			return &Attempt{
				Body:        []byte(fmt.Sprintf("arXiv %s exists and neither form of it could be taken here\n", id)),
				ContentType: "text/plain", TextRetrieved: false,
				Via: fmt.Sprintf("arXiv %s: the paper IS on arXiv and THIS RUN COULD NOT TAKE IT — %v — and the HTML "+
					"rendering at %s did not answer either (arXiv has none for pre-2007 identifiers). This is a fact "+
					"about this fetch, NOT about the paper's existence: do not record the source as absent. The LaTeX "+
					"source at %s is often far smaller and is not tried automatically", id, err, html, eprint)}
		case ViaMetadata:
			att, _ := MetadataRecord(f, doi)
			return att
		case ViaEric:
			att, _ := EricRecord(f, rawURL)
			return att
		}
		return nil
	}
	stamp := func(name string, a *Attempt) *Attempt {
		if a != nil {
			a.Backend = name
		}
		return a
	}
	if via != ViaAuto && via != "" {
		return stamp(via, try(via, true))
	}
	for _, name := range AutoOrder() {
		if a := stamp(name, try(name, false)); a != nil {
			return a
		}
	}
	return nil
}

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
