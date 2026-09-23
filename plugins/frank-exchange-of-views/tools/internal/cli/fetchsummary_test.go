package cli

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// THE WARNING IS TOLD ABOUT THE BACKEND THAT ANSWERED. The archive paragraph used to print over
// every recovered fetch, so a seat holding an arXiv PDF taken live from arxiv.org read that the
// live source had refused this container and that its bytes were a third party's snapshot —
// neither true. A warning that misnames what happened spends the seat's attention on the wrong
// risk, and it did so on the majority of recoveries, because only one of four backends is the
// archive.
func TestOnlyAnArchiveAnswerIsCalledASnapshot(t *testing.T) {
	const snapshot = "ARCHIVE SNAPSHOT"
	arx := fetchSummary{URL: "https://arxiv.org/abs/hep-th/9711200", Bytes: 246327,
		Backend: fetchcache.ViaArxiv, RetrievedVia: "arXiv hep-th/9711200 (…/pdf/…)", TextRetrieved: true}
	got := arx.render()
	if strings.Contains(got, snapshot) {
		t.Errorf("an arXiv PDF is reported as an archive snapshot:\n%s", got)
	}
	// AND THE PROVENANCE DUTY SURVIVES: these bytes still did not come from the url asked for.
	if !strings.Contains(got, "DID NOT COME FROM THE URL YOU ASKED FOR") {
		t.Errorf("gating the archive warning dropped the provenance warning entirely:\n%s", got)
	}

	arc := fetchSummary{URL: "https://ex.org/p", Bytes: 10,
		Backend: fetchcache.ViaArchive, RetrievedVia: "archive.org capture of 2019-05-20", TextRetrieved: true}
	if !strings.Contains(arc.render(), snapshot) {
		t.Errorf("an archive capture is no longer called a snapshot:\n%s", arc.render())
	}

	// A PLAIN LIVE FETCH GETS NEITHER. Nothing was recovered, so there is no provenance to warn
	// about and a warning here would be noise on the common path.
	live := fetchSummary{URL: "https://ex.org/p", Bytes: 10, TextRetrieved: true}
	if out := live.render(); strings.Contains(out, snapshot) || strings.Contains(out, "DID NOT COME FROM") {
		t.Errorf("a live fetch carries a provenance warning:\n%s", out)
	}
}

// A LIVE FETCH THAT RETURNED THE DOCUMENT MUST SAY SO. `text_retrieved` is populated by the
// recovery path, so for years every live fetch published `false` to --json — which by this
// field's own documented meaning says the bytes are a record that the source exists rather than
// its text. The human summary hid it, printing the warning only beside a recovery, so the lie
// lived entirely on the machine-readable surface where nothing would question it. It produced
// three contradictory readings of one source sweep before the field was doubted.
func TestALiveDocumentReportsItsTextAsRetrieved(t *testing.T) {
	no := false
	yes := true
	for _, tc := range []struct {
		name string
		e    fetchcache.Entry
		want bool
	}{
		{"a live article", fetchcache.Entry{Sha: "abc", URL: "https://ex/a", ContentType: "text/html", NotRenderable: &no}, true},
		{"a live PDF", fetchcache.Entry{Sha: "abc", URL: "https://ex/a", ContentType: "application/pdf"}, true},
		{"a wall", fetchcache.Entry{Sha: "abc", URL: "https://ex/a", ContentType: "text/html", NotRenderable: &yes}, false},
		{"a bibliographic record", fetchcache.Entry{Sha: "abc", URL: "https://ex/a", RetrievedVia: "Crossref record", TextRetrieved: false}, false},
		{"a recovered document", fetchcache.Entry{Sha: "abc", URL: "https://ex/a", RetrievedVia: "an open copy", TextRetrieved: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (fetchSummary{TextRetrieved: tc.e.TextRetrieved || liveTextRetrieved(tc.e)}).TextRetrieved; got != tc.want {
				t.Errorf("text_retrieved = %v, want %v — this is what a machine reading --json believes about "+
					"whether it holds the source's text", got, tc.want)
			}
		})
	}
}

// THE INDEX'S VERDICT ON THE PAPER REACHES THE SEAT. A retraction is invisible in the bytes: the
// pdf of a withdrawn paper is the same file it always was, and the notice is a separate document
// nobody fetched. It is on the record because OpenAlex already answers it in the call this tool
// makes to find a url — and it is worth nothing there unless it is also rendered.
//
// The assertion goes through summarize, not through a hand-built fetchSummary. A field that is
// written to the index and never copied into the summary stores perfectly and renders nowhere,
// and a renderer test cannot see that gap because it supplies the value itself.
func TestARetractionIsRenderedFromTheStoredRecord(t *testing.T) {
	yes := true
	e := fetchcache.Entry{
		URL: "https://ex.org/p.pdf", ContentType: "application/pdf", TextRetrieved: true,
		CopyVersion: "publishedVersion", CopyLicense: "cc-by",
		Work: &fetchcache.WorkFacts{Retracted: &yes, WorkType: "article", OAStatus: "bronze", License: "cc-by"},
	}
	got := summarize(record.Run{}, e, 4096, false).render()
	for _, want := range []string{"retracted: true", "THIS WORK IS RETRACTED", "AS RETRACTED",
		"work_type: article", "oa_status: bronze", "copy_version: publishedVersion"} {
		if !strings.Contains(got, want) {
			t.Errorf("the summary does not carry %q:\n%s", want, got)
		}
	}

	// AND A WORK NOBODY ASKED AN INDEX ABOUT SAYS NOTHING. A plain live fetch of a url with no doi
	// consults no index, and printing `retracted: false` there would be a reassurance nobody
	// earned — the difference between "checked, sound" and "never checked" is the whole reason
	// the field is a pointer.
	quiet := summarize(record.Run{}, fetchcache.Entry{URL: "https://ex.org/p", TextRetrieved: true}, 10, false).render()
	if strings.Contains(quiet, "retracted") || strings.Contains(quiet, "oa_status") {
		t.Errorf("a fetch that consulted no index reports on retraction anyway:\n%s", quiet)
	}
}
