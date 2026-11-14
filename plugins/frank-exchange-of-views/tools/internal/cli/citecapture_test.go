package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// evidenceCapture is the one source's copy_captured_at as `show evidence` renders it, and whether
// the key is there at all.
func evidenceCapture(t *testing.T, runDir string) (string, bool) {
	t.Helper()
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var raw struct {
		Sources []map[string]json.RawMessage `json:"sources"`
	}
	if err := json.Unmarshal(b, &raw); err != nil || len(raw.Sources) != 1 {
		t.Fatalf("evidence: %v\n%s", err, b)
	}
	v, ok := raw.Sources[0]["copy_captured_at"]
	if !ok {
		return "", false
	}
	var s string
	if err := json.Unmarshal(v, &s); err != nil {
		t.Fatal(err)
	}
	return s, true
}

// A CITATION OF A COPY THE TOOL TOOK FROM THE WEB ARCHIVE CARRIES WHEN THE ARCHIVE TOOK IT — the
// capture time the fetch index holds for those bytes, the CDX timestamp — and `show evidence` shows
// it beside the fetch time. A citation of the url's own bytes carries none. The seat types nothing.
func TestACitationOfAnArchiveCopyCarriesItsCaptureTime(t *testing.T) {
	t.Run("archive copy", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, fermatReport)
		registerBlue(t, runDir)
		withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
			switch {
			case strings.Contains(u, "web.archive.org/cdx/search"):
				return &fetchcache.Response{Body: []byte(`[["timestamp","original","digest"],["20190520083015","` + fermatURL + `","D1"]]`), ContentType: "application/json"}, nil
			case strings.HasPrefix(u, "https://web.archive.org/web/20190520083015id_/"):
				return &fetchcache.Response{Body: []byte(fermatPage), ContentType: "text/html"}, nil
			}
			return nil, &fetchcache.Refusal{URL: u, Status: 403}
		}))
		if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
			t.Fatalf("cite: %v\n%s", err, out)
		}
		c := firstCiteEvent(t, runDir)
		entry, _, ok, err := fetchcache.Lookup(runtest.Open(t, runDir), fermatURL)
		if err != nil || !ok {
			t.Fatalf("the fetch index holds no entry for %s: %v", fermatURL, err)
		}
		if entry.CopyCapturedAt != "2019-05-20T08:30:15Z" {
			t.Fatalf("the fetch index holds capture time %q, want the CDX timestamp as 2019-05-20T08:30:15Z", entry.CopyCapturedAt)
		}
		if c.CopyCapturedAt == nil || c.GetCopyCapturedAt() != entry.CopyCapturedAt {
			t.Errorf("the cite carries copy_captured_at %v; the fetch index holds %q for the copy it read", c.CopyCapturedAt, entry.CopyCapturedAt)
		}
		if got, ok := evidenceCapture(t, runDir); !ok || got != entry.CopyCapturedAt {
			t.Errorf("show evidence renders copy_captured_at %q (present %v), want %q", got, ok, entry.CopyCapturedAt)
		}
	})
	t.Run("live copy", func(t *testing.T) {
		runDir := fermatRun(t, "text/html")
		if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
			t.Fatalf("cite: %v\n%s", err, out)
		}
		if c := firstCiteEvent(t, runDir); c.CopyCapturedAt != nil {
			t.Errorf("a citation of the url's own bytes carries copy_captured_at %q", c.GetCopyCapturedAt())
		}
		if got, ok := evidenceCapture(t, runDir); ok {
			t.Errorf("show evidence renders copy_captured_at %q for a live copy", got)
		}
	})
}
