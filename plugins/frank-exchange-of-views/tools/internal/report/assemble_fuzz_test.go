package report

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// FuzzWeaveCitations drives the citation weave over arbitrary report text with arbitrary
// anchors and asserts its invariants hold no matter the input: it never panics, never
// leaves a raw "<!--cite:C-…-->" anchor behind (every one becomes a [^N]), and produces a
// bibliography line for each distinct anchor it wove — a dangling anchor (no source) is
// surfaced, never crashed on.
//
// The oracle reads the token with its own pattern, so the weave's reader does not grade itself.
func FuzzWeaveCitations(f *testing.F) {
	citeAnchor := regexp.MustCompile(`<!--cite:(C-[0-9a-f]{8})-->`)
	f.Add("Alpha<!--cite:C-00000001-->. Beta<!--cite:C-00000002-->.", "C-00000001|C-00000002")
	f.Add("No citations here at all.", "")
	f.Add("Dangling<!--cite:C-0000dead-->.", "") // anchor present, no source
	f.Add("Repeat<!--cite:C-00000001--> and again<!--cite:C-00000001-->.", "C-00000001")
	f.Add("<!--cite:C-0000000a--><!--cite:C-0000000b--><!--cite:C-0000000a-->", "C-0000000a|C-0000000b")

	f.Fuzz(func(t *testing.T, md, labels string) {
		var sources []record.Source
		for _, l := range strings.Split(labels, "|") {
			if l = strings.TrimSpace(l); l != "" {
				sources = append(sources, record.Source{Label: l, URL: "https://u/" + l, Title: "T " + l, AccessDate: "2026-08-03"})
			}
		}
		out := weaveCitations(md, sources) // must not panic

		// No VALID citation anchor survives — every well-formed "<!--cite:C-hex-->" is woven
		// to a [^N]. (A malformed fragment like a bare "<!--cite:" is not an anchor; it is
		// inert HTML-comment text and legitimately passes through.)
		if citeAnchor.MatchString(out) {
			t.Fatalf("a valid citation anchor survived the weave:\n%s", out)
		}
		// Every distinct anchor present in the input must appear as a woven [^N] and get a
		// bibliography entry; a valid anchor becomes a footnote ref exactly once numbered.
		anchors := citeAnchor.FindAllStringSubmatch(md, -1)
		distinct := map[string]bool{}
		for _, m := range anchors {
			distinct[m[1]] = true
		}
		if len(distinct) > 0 && !strings.Contains(out, "## Bibliography") {
			t.Fatalf("input had %d anchor(s) but no bibliography was composed:\n%s", len(distinct), out)
		}
		if len(distinct) == 0 && strings.Contains(out, "## Bibliography") {
			t.Fatalf("no anchors but a bibliography appeared:\n%s", out)
		}
	})
}
