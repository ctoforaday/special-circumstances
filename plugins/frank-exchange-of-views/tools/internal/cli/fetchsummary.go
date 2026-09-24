package cli

import (
	"fmt"
	"strings"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// fetchSummary is what `fetch` prints instead of the document (#629 D1, D2).
//
// THE BODY USED TO GO STRAIGHT TO STDOUT, AND FOR A PDF THAT WAS BINARY. `--json` was worse:
// it put invalid UTF-8 in a JSON string field. But the repair is not "print the PDF's text
// instead" — the Settles survey is 67 pages and IEEE 1012 is 80, and pasting either into a
// seat's context is the same waste whether it is legible or not. So NOTHING is returned inline,
// for ANY content type, and what comes back is a set of paths plus the facts a seat needs to
// decide what to open.
//
// EVERY FACT IS A FIELD, WHICH IS THE POINT. The alternative considered — hand back a bare path
// — would have left the seat to infer the content type from the bytes and to assume text
// existed. Here the absence of text is a field with a reason beside it, so the honest zero and
// the unlooked-at case cannot be confused (see [[facts-are-fields]]).
type fetchSummary struct {
	URL         string `json:"url"`
	Sha256      string `json:"sha256"`
	Path        string `json:"path"`
	ContentType string `json:"content_type,omitempty"`
	Filename    string `json:"filename,omitempty"`
	Bytes       int    `json:"bytes"`
	CacheHit    bool   `json:"cache_hit"`
	// SharedBodyWith names other urls already cached under these EXACT bytes. Two different
	// sources answering with one body is not what a document looks like — it is what a bot wall,
	// a login interstitial or a "no results" page looks like, and content-addressing collapses
	// those onto one entry that satisfies "fetch-once, hash-verified" perfectly while the hash
	// certifies the BLOCKADE rather than the source.
	SharedBodyWith []string `json:"shared_body_with,omitempty"`
	// HTTPStatus and RefusalClass carry a REFUSAL THAT WAS THEN RECOVERED FROM. Both were
	// recorded on the index entry and rendered by nothing, so a seat handed an archive snapshot
	// after a 403 learned where the bytes came from and never that the live source had refused,
	// still less whether the refusal was the origin's or this container's proxy. The bare-failure
	// path already tells a seat this, on the error; the recovery path dropped it, which is the
	// path where it matters more, because the fetch looks like a success.
	HTTPStatus   int    `json:"http_status,omitempty"`
	RefusalClass string `json:"refusal_class,omitempty"`
	// RetrievedVia is the SENTENCE naming where these bytes came from, when the live source
	// refused and a backend recovered them. Empty means the bytes are the live source's own.
	RetrievedVia string `json:"retrieved_via,omitempty"`
	// FollowedTo names where the walk ended when a page pointed at its own full text. It is not
	// RetrievedVia and must not be rendered as one: the bytes are the same work from the place
	// its publisher named, not another artifact standing in for it.
	FollowedTo string `json:"followed_to,omitempty"`
	// Backend is which backend that was, as a value rather than a sentence to match on.
	Backend string `json:"backend,omitempty"`
	// TextRetrieved says THESE BYTES ARE THE SOURCE'S TEXT. False means they are a record that
	// the source exists — a bibliographic entry, or an answered "no open copy exists" — and
	// nothing about what the source SAYS may rest on them.
	//
	// IT USED TO BE FALSE ON EVERY LIVE FETCH, which inverted it exactly where it is most often
	// read. The field was populated only by the recovery path, so a 240 KB article fetched
	// straight from its publisher and passed by the shell detector was published to `--json` as
	// `"text_retrieved": false` — by this field's own documented meaning, "not its text". The
	// human summary was spared because it prints the warning only alongside a recovery, so the
	// defect lived entirely in the machine-readable surface, where nothing would question it. It
	// cost three wrong readings of one sweep before the field was doubted rather than the data.
	TextRetrieved bool `json:"text_retrieved"`
	Pages         int  `json:"pages,omitempty"`

	// TextExtracted is a pointer for the same three-state reason the Entry field is: nil means
	// nothing was attempted (this content type is not one anything extracts), false means it was
	// attempted and there was none. Rendering both as `false` would tell a seat that an HTML
	// page has no text.
	TextExtracted *bool  `json:"text_extracted,omitempty"`
	TextPath      string `json:"text_path,omitempty"`
	TextSha256    string `json:"text_sha256,omitempty"`
	TextReason    string `json:"text_reason,omitempty"`
	Extractor     string `json:"extractor,omitempty"`
	OCRDerived    bool   `json:"ocr_derived,omitempty"`

	// The automatic read's own facts (#644). They appear ONLY where the document was a PDF
	// with no text layer — the one case fetch runs the OCR engine on — and they are
	// separate fields from the extraction's because they answer a different question.
	// TextReason says why the DOCUMENT had no text layer; OCRReason says why there is no
	// reading of its pixels either, and a seat that saw only the first would conclude the
	// source is unreadable when in fact automatic reading was switched off — or this
	// binary was built without the engine, which OCRReason states in as many words.
	OCRReason string `json:"ocr_reason,omitempty"`
	// OCREngineAbsent says the automatic read failed because this binary was built without the
	// engine — the fact behind one of OCRReason's sentences, as a field.
	OCREngineAbsent bool `json:"ocr_engine_absent,omitempty"`
	// TextRetrievedReason says why the bytes are not the source's text where the fetch reached the
	// source and took its real document — a container or a binary with no reader here. It is the
	// reason TextRetrieved never had: a bare false is indistinguishable from a fetch that never
	// happened, and the two license different next moves.
	TextRetrievedReason string `json:"text_retrieved_reason,omitempty"`
	// NotRenderable says the bytes are not the document the url names — a challenge, an
	// unrendered app, or a page with too little prose to be either. It is printed EVEN WHEN text
	// was extracted, because all three yield a little text — the nav, a cookie banner, the
	// challenge's own explanation — and a seat that saw `text_extracted: true` and stopped there
	// would verify a citation against the furniture.
	NotRenderable       *bool  `json:"not_renderable,omitempty"`
	NotRenderableReason string `json:"not_renderable_reason,omitempty"`
	// TDMReserved says the source reserved text-and-data-mining rights in its markup. It bears on
	// KEEPING this content, not on reading or quoting it, so it is reported and nothing is gated
	// on it — see fetchcache.TDMReservation.
	TDMReserved *bool  `json:"tdm_reserved,omitempty"`
	TDMPolicy   string `json:"tdm_policy,omitempty"`
	// Retracted, Paratext, WorkType, OAStatus and WorkLicense are what the INDEX says about the
	// paper — not about this url, and not about these bytes. They are on the summary because a
	// perfect fetch cannot discover any of them: the pdf of a retracted paper reads exactly like
	// the pdf of a sound one, and the retraction notice is a separate document the seat never
	// asked for. OpenAlex answers all five in the record this tool already fetches to find a url.
	//
	// REPORTED, NEVER GATED. A retracted paper is a legitimate thing to cite — as retracted — and
	// so is an editorial, a dataset or a book chapter. The tool's duty is to say what the thing
	// is; what may be cited is the seat's judgement and the bench's.
	Retracted   *bool  `json:"retracted,omitempty"`
	Paratext    *bool  `json:"paratext,omitempty"`
	WorkType    string `json:"work_type,omitempty"`
	OAStatus    string `json:"oa_status,omitempty"`
	WorkLicense string `json:"work_license,omitempty"`
	// CopyVersion and CopyLicense describe THESE BYTES: which copy of the work arrived, and under
	// what licence it sits. A submitted preprint and the version of record are different
	// documents to quote from, and until this was recorded a quote could come from either with
	// nothing in the summary saying which.
	CopyVersion string `json:"copy_version,omitempty"`
	CopyLicense string `json:"copy_license,omitempty"`
	// TablePages counts pages whose ruled grid the engine detected — their reconstruction
	// stats live on the reading record. Present only when nonzero, so a prose-only reading
	// renders without it.
	TablePages int `json:"table_pages,omitempty"`
	// TextCellPages counts those table pages rebuilt from the rules' own cells — a ruled table
	// of TEXT, whose rows would otherwise arrive column by column with the row binding lost.
	// A table page in neither count fell back to plain text, and the reading record says why.
	TextCellPages int     `json:"text_cell_pages,omitempty"`
	Engine        string  `json:"engine,omitempty"`
	DPILow        float64 `json:"dpi_low,omitempty"`
	DPIHigh       float64 `json:"dpi_high,omitempty"`
}

// applyReading folds the engine's reading of the page images into the summary.
//
// IT OVERWRITES TextExtracted, and that is the point of the whole change: the extraction
// said false with a reason, the read then succeeded, and what the seat must be told is that
// the text IS there — with `ocr_derived: true` beside it so nobody mistakes a machine
// reading for an author's text layer. The path is the reading's own file, never TextPath:
// two producers writing one name is how a re-render at another resolution silently replaces
// the text a citation was taken from.
func (s *fetchSummary) applyReading(run record.Run, r fetchcache.ReadingRecord) {
	yes := true
	s.TextExtracted = &yes
	s.TextReason = ""
	s.TextPath = fetchcache.OCRTextPath(run, r.Sha)
	s.TextSha256 = r.TextSha
	s.OCRDerived = true
	// THE EXTRACTOR DID NOT PRODUCE THIS TEXT, so its id must not be left beside it. PDFium
	// opened the document, counted 80 pages and found no glyphs; the reading came from the
	// OCR engine. `engine` is the word the reading record already uses for this fact, so it
	// is the word used here rather than a second one meaning the same thing — and unlike
	// the extractor's id it names a producer an audit CAN re-run for the same bytes.
	s.Extractor = ""
	s.Engine = r.Engine
	s.DPILow, s.DPIHigh = r.DPIRange()
	s.TablePages = r.TablePages()
	s.TextCellPages = r.TextCellPages()
}

// summarize projects a cache entry into what the seat is shown. Paths are absolute — a Run
// resolves its directory once, at construction — because the seat's next act is to Read one.
func summarize(run record.Run, e fetchcache.Entry, bodyLen int, hit bool) fetchSummary {
	// Best-effort: a cache whose index cannot be read is a fetch that still succeeded, and
	// refusing here would trade a real document for a missing warning.
	shared, _ := fetchcache.URLsSharingBody(run, e.Sha, e.URL)
	s := fetchSummary{
		URL:            e.URL,
		Sha256:         e.Sha,
		Path:           fetchcache.Path(run, e.Sha),
		ContentType:    e.ContentType,
		Filename:       e.Filename,
		Bytes:          bodyLen,
		CacheHit:       hit,
		SharedBodyWith: shared,
		HTTPStatus:     e.HTTPStatus,
		RefusalClass:   e.RefusalClass,
		RetrievedVia:   e.RetrievedVia,
		FollowedTo:     e.FollowedTo,
		Backend:        e.Backend,
		// A LIVE FETCH THAT YIELDED A DOCUMENT HAS THE SOURCE'S TEXT. The entry's own flag speaks
		// only for the recovery path, so it is the wrong answer for the common one.
		TextRetrieved:       e.TextRetrieved || liveTextRetrieved(e),
		TextRetrievedReason: e.TextRetrievedReason,
		Pages:               e.Pages,
		TextExtracted:       e.TextExtracted,
		TextSha256:          e.TextSha,
		TextReason:          e.TextReason,
		NotRenderable:       e.NotRenderable, NotRenderableReason: e.NotRenderableReason,
		TDMReserved: e.TDMReserved, TDMPolicy: e.TDMPolicy,
		CopyVersion: e.CopyVersion, CopyLicense: e.CopyLicense,
		Extractor: e.Extractor,
	}
	if e.Work != nil {
		s.Retracted, s.Paratext = e.Work.Retracted, e.Work.Paratext
		s.WorkType, s.OAStatus, s.WorkLicense = e.Work.WorkType, e.Work.OAStatus, e.Work.License
	}
	// THE PATH IS NAMED ONLY WHEN THE FILE IS THERE. A text_path pointing at a file that was
	// never written is worse than no field at all: a seat would Read it, get a not-found, and
	// have to guess whether the tool or the document was at fault.
	if e.TextSha != "" {
		s.TextPath = fetchcache.TextPath(run, e.Sha)
	}
	return s
}

// applicableToOCR reports whether this is the one case fetch runs the OCR engine on: a PDF
// the extractor looked at and found no text layer in.
//
// THE THREE-STATE POINTER IS LOAD-BEARING HERE. nil means nothing looked — an HTML page, an
// index line written by an older binary — and rendering either to pixels would be absurd. A
// plain bool would have collapsed that into the same false the scanned standard produces,
// and fetch would rasterise every web page it read.
func applicableToOCR(e fetchcache.Entry) bool {
	return e.ContentType == "application/pdf" && e.TextExtracted != nil && !*e.TextExtracted
}

// render is the human-facing form: the same fields, one per line, in the order a seat needs
// them — what it is, then where to read it, then whether the text is there.
//
// It is deliberately NOT a paraphrase of the JSON. Every line is `key: value`, so the two
// renderings carry identical facts and neither has to be parsed back out of prose.
func (s fetchSummary) render() string {
	var b strings.Builder
	line := func(k, v string) {
		if v != "" {
			fmt.Fprintf(&b, "%s: %s\n", k, v)
		}
	}
	line("url", s.URL)
	line("sha256", s.Sha256)
	line("path", s.Path)
	line("content_type", s.ContentType)
	line("filename", s.Filename)
	line("bytes", fmt.Sprint(s.Bytes))
	line("cache_hit", fmt.Sprint(s.CacheHit))
	// PROVENANCE FIRST: these bytes may not be the live source, and every later judgement about
	// what the source SAYS depends on knowing that before reading a word of it.
	// THE LIVE SOURCE'S REFUSAL, ABOVE WHERE THE RECOVERY IS NAMED, because a seat that reads no
	// further has still been told the bytes below are not the source's own answer.
	if s.HTTPStatus != 0 {
		line("live_fetch_refused", fmt.Sprintf("HTTP %d", s.HTTPStatus))
		if s.RefusalClass == "unknown" {
			fmt.Fprintf(&b, "refusal_class: unknown\n"+
				"  ^ WHO REFUSED IS NOT KNOWN. An egress proxy refusing the host and the origin refusing this client\n"+
				"    are the same status line, and a proxy is configured here. One is a fact about this container, the\n"+
				"    other a fact about the source, and they license different next acts. Do not record an unreached\n"+
				"    source as evidence of absence: say it was UNREACHABLE FROM HERE, not that the question is open.\n")
		} else {
			line("refusal_class", s.RefusalClass)
		}
	}
	if s.RetrievedVia != "" {
		fmt.Fprintf(&b, "retrieved_via: %s\n", s.RetrievedVia)
		if !s.TextRetrieved {
			fmt.Fprintf(&b, "text_retrieved: false\n"+
				"  ^ THESE BYTES ARE A RECORD THAT THE SOURCE EXISTS, NOT ITS TEXT. That is a real finding and a\n"+
				"    legitimate citation — as source-text `unread`. It is not a reading, and nothing about what the\n"+
				"    source SAYS may rest on it.\n")
		}
		// THE WARNING IS TOLD ABOUT THE BACKEND THAT ACTUALLY ANSWERED. This paragraph used to
		// print whenever anything at all was recovered, so a seat holding an arXiv PDF taken live
		// from arxiv.org read that the live source had refused this container and that its bytes
		// were a third party's snapshot. Neither was true, and a warning that misnames what
		// happened spends the seat's attention on the wrong risk.
		if s.Backend == fetchcache.ViaArchive {
			fmt.Fprintf(&b, "  ^ THESE BYTES ARE AN ARCHIVE SNAPSHOT — what the source said on that date, retrieved from a\n"+
				"    third party. Usable, and NOT the same artifact: say so in any claim about what the source\n"+
				"    currently says.\n"+
				"    READ IT BEFORE YOU CITE IT AS READ. Measured against the sources that actually failed in\n"+
				"    2026-09-02_quadratic-formula, a snapshot of a SUBSCRIPTION article is usually the landing page —\n"+
				"    title, abstract and analytics — not the text. That is a source you have NOT read at the leaf, and\n"+
				"    a citation must say so rather than inherit the snapshot's apparent success.\n")
		} else {
			fmt.Fprintf(&b, "  ^ THESE BYTES DID NOT COME FROM THE URL YOU ASKED FOR — a backend found them elsewhere, and\n"+
				"    the line above says where. It may be a different artifact from the one the url names: another\n"+
				"    version, another format, or a record ABOUT the source rather than the source. READ IT BEFORE\n"+
				"    YOU CITE IT AS READ, and let the citation say which artifact you actually read.\n")
		}
	}
	// LOUD, AND ABOVE THE TEXT LINES, because a seat that reads no further has still been told the
	// one thing that decides whether this is a source at all.
	if len(s.SharedBodyWith) > 0 {
		fmt.Fprintf(&b, "shared_body_with: %s\n", strings.Join(s.SharedBodyWith, ", "))
		fmt.Fprintf(&b, "  ^ THESE EXACT BYTES WERE ALREADY SERVED FOR THE URL(S) ABOVE. Two different sources do not\n"+
			"    answer with one body: this is the shape of a bot wall, a login interstitial or a no-results page.\n"+
			"    The sha is hash-verified and certifies the BLOCKADE, not the document. Read the file before citing\n"+
			"    it, and if it is a wall, this source is UNREAD — cite it as such or not at all.\n")
	}
	if s.Pages > 0 {
		line("pages", fmt.Sprint(s.Pages))
	}
	switch {
	case s.TextExtracted == nil:
		// Nothing attempted. Say so rather than printing a text_extracted line a seat would read
		// as a measured "no".
	case *s.TextExtracted:
		line("text_extracted", "true")
		line("text_path", s.TextPath)
		line("text_sha256", s.TextSha256)
		line("extractor", s.Extractor)
		if s.OCRDerived {
			line("ocr_derived", "true")
			// STATED BESIDE THE TEXT IT QUALIFIES, and only when nonzero: a seat about to
			// open the file learns that some pages are reconstructed tables whose
			// confidence stats live on the reading record, and a prose-only reading
			// renders without the line.
			if s.TablePages > 0 {
				line("table_pages", fmt.Sprint(s.TablePages))
			}
			if s.TextCellPages > 0 {
				line("text_cell_pages", fmt.Sprint(s.TextCellPages))
			}
		}
	default:
		line("text_extracted", "false")
		line("text_reason", s.TextReason)
		line("extractor", s.Extractor)
	}
	// OUTSIDE THE SWITCH, DELIBERATELY. A skeleton normally DOES extract a little text — the
	// navigation, a cookie banner — so this arrives on the `text_extracted: true` arm, which is
	// the arm a seat reads as success. Printed there or not at all, the flag would be absent in
	// exactly the case it exists for.
	// STATED, NEVER ENFORCED. A reservation is the Article 4 opt-out: it covers mining, not
	// reading and not quotation, so it changes nothing this tool does today and would change
	// everything about keeping a corpus. Reporting it is what makes that a decision later rather
	// than a thing nobody noticed.
	if s.TDMReserved != nil && *s.TDMReserved {
		line("tdm_reserved", "true")
		if s.TDMPolicy != "" {
			line("tdm_policy", s.TDMPolicy)
		}
		fmt.Fprintf(&b, "  ^ THE SOURCE RESERVES TEXT-AND-DATA-MINING RIGHTS over this page. That reservation covers\n"+
			"    MINING — analysing a body of works to derive patterns — and not READING it or QUOTING it, which\n"+
			"    are separate rights it does not reach. Read and cite this source as you would any other. What it\n"+
			"    forbids is keeping a corpus of it or training on it, neither of which this tool does.\n")
	}
	// A LIVE FETCH THAT GOT BYTES NOBODY CAN READ SAYS SO IN THE HUMAN SUMMARY TOO. The
	// `text_retrieved` line printed only under `retrieved_via`, so on the live path — the common
	// one — the flag existed in `--json` and nowhere a seat reading the summary would see it. A
	// container arrived, the fetch looked like every other success, and only the content type
	// hinted otherwise.
	if s.FollowedTo != "" {
		fmt.Fprintf(&b, "followed_to: %s\n"+
			"  ^ THE PAGE YOU ASKED FOR NAMED ITS OWN FULL TEXT AND THIS FOLLOWED IT. Same work, from where the\n"+
			"    publisher said to look — not a substitute copy. Cite the url you asked for; this records how the\n"+
			"    text was reached.\n", s.FollowedTo)
	}
	if !s.TextRetrieved && s.RetrievedVia == "" && s.TextRetrievedReason != "" {
		fmt.Fprintf(&b, "text_retrieved: false\n  ^ %s\n", s.TextRetrievedReason)
	}
	if s.NotRenderable != nil && *s.NotRenderable {
		line("not_renderable", "true")
		line("not_renderable_reason", s.NotRenderableReason)
	}
	// WHAT THE INDEX KNOWS THAT THE BYTES CANNOT TELL YOU.
	line("work_type", s.WorkType)
	line("oa_status", s.OAStatus)
	line("work_license", s.WorkLicense)
	line("copy_version", s.CopyVersion)
	line("copy_license", s.CopyLicense)
	if s.Paratext != nil && *s.Paratext {
		line("paratext", "true")
		fmt.Fprintf(&b, "  ^ THE INDEX CALLS THIS PARATEXT — an editorial, a masthead, a table of contents, rather than\n"+
			"    the research article. Cite it for what it is; do not cite it as the study.\n")
	}
	// LAST, AND UNMISSABLE. A retraction is the one fact here that can void a citation outright,
	// and it survives a flawless fetch: the pdf of a retracted paper is byte-identical to what it
	// was before the retraction, and the notice is a different document nobody asked for.
	if s.Retracted != nil && *s.Retracted {
		line("retracted", "true")
		fmt.Fprintf(&b, "  ^ THIS WORK IS RETRACTED. The bytes are genuine and the fetch was sound — retraction is a\n"+
			"    judgement about the paper, not about this retrieval, so nothing upstream could have caught it.\n"+
			"    You MAY cite it, and you MUST cite it AS RETRACTED: its findings do not support a claim, and a\n"+
			"    quotation from it stands only as evidence of what the withdrawn paper said.\n")
	}
	// THE REASON IS PRINTED WHETHER OR NOT THERE IS TEXT. Where a reading succeeded it is
	// empty and this line does not appear; where it did not — switched off, over the render
	// budget, the engine absent from this binary — it is the only thing standing between
	// the seat and the conclusion that the source itself is unreadable.
	line("ocr_reason", s.OCRReason)
	if s.OCRDerived {
		line("engine", s.Engine)
		line("dpi", dpiSpan(s.DPILow, s.DPIHigh))
	}
	return b.String()
}

// liveTextRetrieved answers whether a fetch that reached the source itself came back with the
// document, rather than with a wall or an empty answer. It is deliberately conservative: a page
// the shell detector refused is not the source's text however many bytes it carried.
func liveTextRetrieved(e fetchcache.Entry) bool {
	if e.RetrievedVia != "" || e.Sha == "" {
		return false // a recovery speaks for itself through the entry's own flag
	}
	// A CONTAINER IS NOT TEXT. The live path asked only whether the page was a wall, so every
	// media type that is not html passed — a zip, a tarball, an image — and the flag a citation's
	// `leaf` reading rests on said the source's text was in hand.
	if !fetchcache.TextBearing(e.ContentType) {
		return false
	}
	if e.NotRenderable != nil && *e.NotRenderable {
		return false
	}
	return true
}
