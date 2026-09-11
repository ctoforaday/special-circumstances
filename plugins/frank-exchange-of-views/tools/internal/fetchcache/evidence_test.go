package fetchcache

import (
	"os"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
)

// The engine's evidence for a page is kept beside the reading it came from (#644: the
// diagnostics that name a page's resolution estimate, and the word boxes a failed
// reconstruction was built from, were computed and thrown away). Three properties, each a
// way it could quietly go wrong: an empty kind leaves no file rather than an empty one; a
// page served from its receipt keeps the evidence of the read that produced it; and a page
// re-read under a new engine keeps NONE of the old evidence beside the new reading.
func TestPageEvidenceIsKeptBesideTheReading(t *testing.T) {
	first := tessocr.PageResult{
		Text:        "reading\n",
		Diagnostics: "Estimating resolution as 282\n",
		Evidence:    tessocr.PageEvidence{PortraitTSV: "portrait tsv", SparseTSV: "sparse tsv"},
	}
	fe := &fakeEngine{perCall: func(int) (tessocr.PageResult, error) { return first, nil }}
	withEngine(t, fe)
	run, sha, rd := onePageRender(t)

	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	want := map[string]string{
		"engine.log":   first.Diagnostics,
		"portrait.tsv": "portrait tsv",
		"sparse.tsv":   "sparse tsv",
	}
	assertEvidence := func(when string) {
		t.Helper()
		for _, k := range evidenceKinds {
			b, err := os.ReadFile(EvidencePath(run, sha, 1, k.suffix))
			w, kept := want[k.suffix]
			switch {
			case kept && err != nil:
				t.Errorf("%s: %s not kept: %v", when, k.suffix, err)
			case kept && string(b) != w:
				t.Errorf("%s: %s = %q, want %q", when, k.suffix, b, w)
			case !kept && !os.IsNotExist(err):
				t.Errorf("%s: %s exists (err %v) though the engine produced none", when, k.suffix, err)
			}
		}
	}
	assertEvidence("after the read")

	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatalf("second ReadRenderedPages: %v", err)
	}
	if fe.calls != 1 {
		t.Fatalf("the engine ran %d times; the second call should have been served by the receipt", fe.calls)
	}
	assertEvidence("after a receipt reuse")

	withEngine(t, &fakeEngine{id: "fake@new", perCall: func(int) (tessocr.PageResult, error) {
		return tessocr.PageResult{Text: "reading\n"}, nil
	}})
	if _, err := ReadRenderedPages(run, sha, rd); err != nil {
		t.Fatalf("re-read under a new engine: %v", err)
	}
	want = map[string]string{}
	assertEvidence("after a re-read that produced no evidence")
}
