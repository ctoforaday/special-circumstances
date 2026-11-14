package fetchcache

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// A COPY TAKEN FROM THE WEB ARCHIVE CARRIES WHEN THE ARCHIVE TOOK IT. The fetch time says when this
// run took the bytes; a snapshot's bytes are what the url said at the capture, which can be years
// earlier, and the entry is where the citation reads it from. A live copy has no capture time.
func TestAnArchiveCopyCarriesItsCaptureTimeAndALiveCopyNone(t *testing.T) {
	t.Run("the archive rung, from CDX", func(t *testing.T) {
		run := runtest.New(t, t.TempDir())
		cdx := `[["timestamp","original","digest"],["20190520083015","https://ex/a","D1"]]`
		f := fake(func(u string) (*Response, error) {
			switch {
			case strings.Contains(u, "cdx/search"):
				return &Response{Body: []byte(cdx), ContentType: "application/json"}, nil
			case strings.Contains(u, "web.archive.org/web/"):
				return &Response{Body: []byte("<html><body>" + strings.Repeat("the archived page. ", 60) + "</body></html>"), ContentType: "text/html"}, nil
			}
			if r, ok := emptyIndexAnswer(u); ok {
				return r, nil
			}
			return nil, &Refusal{URL: u, Status: 403}
		})
		e, _, _, err := Resolve(run, "https://ex/a", f)
		if err != nil {
			t.Fatalf("resolve: %v", err)
		}
		if e.Substitution != SubstitutionRecovered {
			t.Fatalf("the stored copy is %q, not the archive recovery this test is about", e.Substitution)
		}
		if e.CopyCapturedAt != "2019-05-20T08:30:15Z" {
			t.Errorf("the archive copy carries capture time %q, want CDX's 20190520083015 as 2019-05-20T08:30:15Z", e.CopyCapturedAt)
		}
		stored, _, ok, lerr := Lookup(run, "https://ex/a")
		if lerr != nil || !ok || stored.CopyCapturedAt != e.CopyCapturedAt {
			t.Errorf("the index holds capture time %q (%v %v), Resolve returned %q", stored.CopyCapturedAt, ok, lerr, e.CopyCapturedAt)
		}
	})
	t.Run("an open-access location the archive held", func(t *testing.T) {
		withText(t, "the paper")
		const pdfURL = "https://publisher.example/x.pdf"
		site := &keepGoingSite{asked: "unused",
			locs: `{"pdf_url":"` + pdfURL + `","is_oa":true}`,
			bodies: map[string]*Response{
				pdfURL: {Body: loadBody(t, "springer-link-paywalled.html.gz"), ContentType: "text/html"},
				"https://archive.org/wayback/available?url=https%3A%2F%2Fpublisher.example%2Fx.pdf": {Body: []byte(
					`{"archived_snapshots":{"closest":{"available":true,"status":"200","timestamp":"20240101123456","url":"https://web.archive.org/web/20240101123456/` + pdfURL + `"}}}`)},
				"https://web.archive.org/web/20240101123456if_/" + pdfURL: {Body: []byte("%PDF-1.7 the paper"), ContentType: "application/pdf"},
			}}
		att := Recover(site, keepGoingDOI, ViaOA, "")
		if att == nil || string(att.Body) != "%PDF-1.7 the paper" {
			t.Fatalf("the archived pdf was not taken: %+v", att)
		}
		if att.CapturedAt != "2024-01-01T12:34:56Z" {
			t.Errorf("the archived open-access copy carries capture time %q, want the archive's 20240101123456 as 2024-01-01T12:34:56Z", att.CapturedAt)
		}
		if e := EntryFor(runtest.New(t, t.TempDir()), keepGoingDOI, att, SubstitutionFullerCopy); e.CopyCapturedAt != att.CapturedAt {
			t.Errorf("the entry built from the attempt carries %q, the attempt %q", e.CopyCapturedAt, att.CapturedAt)
		}
	})
	t.Run("a live copy", func(t *testing.T) {
		run := runtest.New(t, t.TempDir())
		e, _, _, err := Resolve(run, "https://ex/1", &stubFetcher{resp: map[string][]byte{"https://ex/1": []byte("the source")}})
		if err != nil {
			t.Fatal(err)
		}
		if e.CopyCapturedAt != "" {
			t.Errorf("a live copy carries capture time %q", e.CopyCapturedAt)
		}
	})
}

// A CAPTURE WHOSE TIME CANNOT BE READ IS NOT TAKEN SILENTLY. CDX's timestamp is the only statement of
// when the archive took the bytes; a copy stored without it would cite exactly like a live one.
// Asked by name, the archive rung says what came back; reached under auto, it declines.
func TestACaptureWithAnUnreadableTimeIsRefusedLoudly(t *testing.T) {
	cdx := `[["timestamp","original","digest"],["2019052","https://ex/a","D1"]]`
	f := fake(func(u string) (*Response, error) {
		switch {
		case strings.Contains(u, "cdx/search"):
			return &Response{Body: []byte(cdx), ContentType: "application/json"}, nil
		case strings.Contains(u, "web.archive.org/web/"):
			return &Response{Body: []byte("<html>the archived page</html>"), ContentType: "text/html"}, nil
		}
		return nil, &Refusal{URL: u, Status: 403}
	})
	att := Recover(f, "https://ex/a", ViaArchive, "")
	if att == nil || att.TextRetrieved || !strings.Contains(att.Via, "2019052") {
		t.Fatalf("a capture with an unreadable time was taken or not named: %+v", att)
	}
}
