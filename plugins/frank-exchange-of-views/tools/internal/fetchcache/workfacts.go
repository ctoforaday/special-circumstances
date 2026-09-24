package fetchcache

import (
	"encoding/json"
	"net/url"
	"sort"
	"strings"
)

// WHAT THE MAINTAINED INDEXES ALREADY KNOW, TAKEN FROM THE CALL WE ALREADY MAKE.
//
// This tool fetched an OpenAlex work record and read ONE field out of it — a url — then threw
// the rest away. The rest is the part a citation actually turns on: whether the paper was
// RETRACTED, which copy is the published one, what licence the copy carries, whether the venue
// was assessed as open.
//
// Rebuilding any of that from fetched bytes would be building a staler, worse copy of something a
// funded team maintains. The division of labour is therefore explicit: the indexes answer what
// the WORK is and where its copies are; this tool only characterises what a given host's page
// mechanics do, which is the part no index tracks.

// WorkFacts are the index's statements about a work, as opposed to about a url.
type WorkFacts struct {
	// Retracted is the one that can invalidate a citation outright, and it costs nothing: it is a
	// boolean in a record already being fetched. Measured: OpenAlex reports is_retracted true for
	// the Wakefield MMR paper. A seat citing withdrawn work as evidence is a failure no amount of
	// fetch quality redeems.
	//
	// A POINTER, because "not retracted" and "nobody was asked" are different facts and a citation
	// may rest on the first but never on the second.
	Retracted *bool `json:"retracted,omitempty"`
	// Paratext marks front matter — an editorial, a table of contents, a masthead. RECORDED, NOT
	// GATED: a seat may legitimately cite any of it, and the tool's job is to say what the thing
	// is, not to decide what may be cited.
	Paratext *bool `json:"paratext,omitempty"`
	// WorkType is the index's own vocabulary: article, book-chapter, dataset, preprint…
	WorkType string `json:"work_type,omitempty"`
	// OAStatus is gold, green, bronze, hybrid or closed. BRONZE IS THE ONE THAT MATTERS: readable
	// on the publisher's site with no licence at all, so it may be read and not redistributed.
	OAStatus string `json:"oa_status,omitempty"`
	// License is the best open location's licence, where the index knows one.
	License string `json:"license,omitempty"`
}

// OALocation is one place a copy sits, with the index's own typing of it.
type OALocation struct {
	URL string
	// IsDocument says this location is the DOCUMENT ITSELF rather than a page about it — a pdf,
	// or full text in another form. It was called IsPDF, and the name was wrong the moment
	// Europe PMC's JATS route joined the list: a field named for one format, carrying the
	// concept "try this before any landing page", is a name that will be believed by whoever
	// reads it next.
	IsDocument bool
	Version    string // publishedVersion, acceptedVersion, submittedVersion
	License    string
}

// versionRank orders the copies of a work by how close each is to the version of record.
//
// A CITATION SHOULD QUOTE THE PUBLISHED COPY WHERE ONE IS REACHABLE. An accepted manuscript has
// the same content and different pagination; a submitted preprint may differ in substance from
// what was published. Taking whichever copy answered first meant a quote could come from a draft
// while a published pdf sat one entry further down the same list.
func versionRank(v string) int {
	switch strings.ToLower(v) {
	case "publishedversion":
		return 0
	case "acceptedversion":
		return 1
	case "submittedversion":
		return 2
	}
	return 3 // the index did not say; after everything it did say
}

// rankLocations orders candidates: a direct pdf before a landing page, then by how close the copy
// is to the version of record. Stable, so the index's own order survives inside a tier.
func rankLocations(locs []OALocation) []OALocation {
	sort.SliceStable(locs, func(i, j int) bool {
		if locs[i].IsDocument != locs[j].IsDocument {
			return locs[i].IsDocument
		}
		return versionRank(locs[i].Version) < versionRank(locs[j].Version)
	})
	return locs
}

// openAlexWork reads the whole record rather than one field of it.
func openAlexWork(f Fetcher, doi string) (WorkFacts, []OALocation, bool) {
	if doi == "" {
		return WorkFacts{}, nil, true
	}
	resp, err := f.Fetch("https://api.openalex.org/works/doi:" + url.PathEscape(doi) + "?mailto=" + ContactEmail)
	if err != nil {
		return WorkFacts{}, nil, false
	}
	var w struct {
		ID          string `json:"id"` // the discriminator: an error body decodes without it
		IsRetracted *bool  `json:"is_retracted"`
		IsParatext  *bool  `json:"is_paratext"`
		Type        string `json:"type"`
		OpenAccess  struct {
			OAStatus string `json:"oa_status"`
			OAURL    string `json:"oa_url"`
		} `json:"open_access"`
		BestOA *struct {
			License string `json:"license"`
		} `json:"best_oa_location"`
		Locations []struct {
			PDFURL  string `json:"pdf_url"`
			Landing string `json:"landing_page_url"`
			IsOA    bool   `json:"is_oa"`
			Version string `json:"version"`
			License string `json:"license"`
		} `json:"locations"`
	}
	if json.Unmarshal(resp.Body, &w) != nil || w.ID == "" {
		return WorkFacts{}, nil, false
	}
	facts := WorkFacts{
		Retracted: w.IsRetracted, Paratext: w.IsParatext,
		WorkType: w.Type, OAStatus: w.OpenAccess.OAStatus,
	}
	if w.BestOA != nil {
		facts.License = w.BestOA.License
	}
	var locs []OALocation
	for _, l := range w.Locations {
		// `pdf_url` IS OPENALEX'S CLAIM, NOT A CONTENT TYPE. It files an abstract page there
		// often enough to matter: measured over ten works whose free copy this tool failed to
		// reach, five listed a `pubmed.ncbi.nlm.nih.gov/<pmid>` abstract as `pdf_url`, and each
		// consumed one of four candidate slots ahead of a location that would have worked. The
		// url is still tried — as the landing page it is, where its own citation_pdf_url may
		// name the real thing.
		if l.PDFURL != "" {
			locs = append(locs, OALocation{URL: l.PDFURL, IsDocument: looksLikePDF(l.PDFURL),
				Version: l.Version, License: l.License})
		}
		if l.IsOA && l.Landing != "" {
			locs = append(locs, OALocation{URL: l.Landing, Version: l.Version, License: l.License})
		}
	}
	if w.OpenAccess.OAURL != "" {
		locs = append(locs, OALocation{URL: w.OpenAccess.OAURL,
			IsDocument: strings.Contains(w.OpenAccess.OAURL, ".pdf")})
	}
	return facts, locs, true
}

// Retraction is the index's retraction answer for the work these bytes are a copy of, or nil
// where no index was consulted. A method rather than a reach into Work, because every caller
// wants the three-state answer and a nil Work is one of the three.
func (e Entry) Retraction() *bool {
	if e.Work == nil {
		return nil
	}
	return e.Work.Retracted
}
