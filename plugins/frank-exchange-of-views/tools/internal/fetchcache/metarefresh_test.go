package fetchcache

import (
	"net/url"
	"strings"
	"testing"
)

func mustURL(t *testing.T, s string) *url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

// THE SHAPE ELSEVIER ACTUALLY SENDS, copied from a live capture: a relative target, an HTML
// entity in the query string, and eleven characters of visible text. One url in five of the
// top-cited works across 26 fields landed on one of these.
const elsevierBouncer = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
<html><head><meta charset="utf-8">
<meta HTTP-EQUIV="REFRESH" content="2; url='/retrieve/articleSelectSinglePerm?Redirect=https%3A%2F%2Fwww.sciencedirect.com%2Fscience%2Farticle%2Fpii%2F0043135482902494&amp;key=806643d4'"/>
<meta name="tdm-reservation" content="1" />
<title>Redirecting</title></head><body>Redirecting</body></html>`

func TestAMarkupRedirectIsFollowedWhereTheHeadersHaveNone(t *testing.T) {
	base := mustURL(t, "https://www.sciencedirect.com/science/article/pii/0043135482902494")
	got := metaRefreshTarget("text/html; charset=utf-8", []byte(elsevierBouncer), base)
	// RESOLVED AGAINST THE URL AFTER THE HEADER REDIRECTS, not the one the seat typed. These
	// urls arrive via doi.org, so resolving a relative target against the original would build
	// it on doi.org — a host that has never heard of the path.
	want := "https://www.sciencedirect.com/retrieve/articleSelectSinglePerm?Redirect=https%3A%2F%2Fwww.sciencedirect.com%2Fscience%2Farticle%2Fpii%2F0043135482902494&key=806643d4"
	if got != want {
		t.Errorf("target = %q\nwant     %q", got, want)
	}
}

func TestAMarkupRedirectIsNotFollowedAwayFromARealDocument(t *testing.T) {
	// A PAGE WITH PROSE IS THE ANSWER, whatever it also asks the browser to do. Following this
	// would discard an article for whatever it points at, which is the failure direction that
	// manufactures an absence.
	doc := `<html><head><meta http-equiv="refresh" content="600"></head><body><article><p>` +
		strings.Repeat("This is a real sentence of the paper. ", 30) + `</p></article></html>`
	if got := metaRefreshTarget("text/html", []byte(doc), mustURL(t, "https://ex.org/a")); got != "" {
		t.Errorf("a document that refreshes was abandoned for %q", got)
	}
}

func TestAMarkupRedirectIsRefusedWhereItLeavesTheWeb(t *testing.T) {
	for _, tc := range []struct{ name, target string }{
		{"javascript", "javascript:alert(1)"},
		{"data", "data:text/html,hi"},
		{"file", "file:///etc/passwd"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := `<html><head><meta http-equiv="refresh" content="0; url=` + tc.target + `"></head><body>x</body></html>`
			if got := metaRefreshTarget("text/html", []byte(b), mustURL(t, "https://ex.org/a")); got != "" {
				t.Errorf("followed a refresh out of http(s) to %q", got)
			}
		})
	}
}

func TestANonBouncerYieldsNoTarget(t *testing.T) {
	base := mustURL(t, "https://ex.org/a")
	for _, tc := range []struct{ name, ct, body string }{
		{"no refresh at all", "text/html", "<html><body>hi</body></html>"},
		{"a refresh with no url", "text/html", `<html><head><meta http-equiv="refresh" content="5"></head><body>x</body></html>`},
		{"not html", "application/pdf", `%PDF-1.7 <meta http-equiv="refresh" content="0; url=/x">`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := metaRefreshTarget(tc.ct, []byte(tc.body), base); got != "" {
				t.Errorf("target = %q, want none", got)
			}
		})
	}
}
