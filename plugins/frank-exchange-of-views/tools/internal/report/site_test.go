package report

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func siteFixture(t *testing.T) (string, record.Family) {
	t.Helper()
	board := &boardT{
		GapOrder: []string{"G-00000001"},
		Gaps:     map[string]*record.Gap{"G-00000001": {ID: "G-00000001", Open: true}},
		Events: []*record.Event{
			recordtest.Event(t, "", &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING)}),
		},
	}
	docs := []Doc{
		{File: FileReport, Nav: "Report", Blurb: "the research", Body: "## Read this first\n\nG-00000001 is still open.\n"},
		{File: FileDocket, Nav: "Board", Blurb: "the gaps", Body: "### G-00000001 — eviction races the reader\n\ncache.go:88\n"},
	}
	return RenderSite("# Whether the cache is coherent", Ship(docs), board.fam()), board.fam()
}

// ONE FILE. A reader opens it out of a tarball months later, offline — so nothing the page
// NEEDS may live anywhere else. This is the property the whole tier is for, and it is the one
// a later convenience (a CDN stylesheet, a font, an icon set) silently takes away.
//
// One address is allowed, and it is an enhancement, not a need: the pinned mermaid module,
// imported only when a document carries a diagram fence, degrading offline to the fenced
// source with a note. It must appear exactly once, inside the script, never as a fetched
// resource attribute — the page must render complete before it resolves.
func TestTheSiteIsSelfContained(t *testing.T) {
	html, _ := siteFixture(t)
	for _, forbidden := range []string{"http://", "<link", "<img", "src=\"http"} {
		if strings.Contains(html, forbidden) {
			t.Errorf("the site reaches outside itself (%q) — it must open offline from an archive", forbidden)
		}
	}
	if got := strings.Count(html, mermaidCDN); got != 1 {
		t.Errorf("the pinned mermaid URL must appear exactly once (the dynamic import), got %d", got)
	}
	if got := strings.Count(html, "https://cdn"); got != strings.Count(html, mermaidCDN) {
		t.Errorf("a CDN reference other than the pinned mermaid URL crept in")
	}
	if !strings.Contains(html, "rendering it needs network access") {
		t.Errorf("the offline fallback note for diagrams is missing — without it an offline miss is silent")
	}
	for _, want := range []string{"<!DOCTYPE html>", "<style>", "<script>"} {
		if !strings.Contains(html, want) {
			t.Errorf("the site is missing %q", want)
		}
	}
}

// The tab strip is the set, and the first document is the one that opens.
func TestTabsAreTheDocumentSet(t *testing.T) {
	html, _ := siteFixture(t)
	if !strings.Contains(html, `data-doc="report.md" aria-selected="true"`) {
		t.Errorf("the research document is not the tab that opens:\n%s", html)
	}
	if !strings.Contains(html, `<button role="tab" data-doc="docket.md"`) {
		t.Errorf("the docket has no tab")
	}
	if strings.Contains(html, `data-doc="judgments.md"`) {
		t.Errorf("a tab was rendered for a document the set does not contain")
	}
}

// The verdict is a BADGE off the record, and a ceiling termination is not painted as a
// failure — it is not one, and the report says so at length.
func TestTheVerdictBadgeComesOffTheRecord(t *testing.T) {
	html, board := siteFixture(t)
	if !strings.Contains(html, `<span class="badge warn">CEILING-TERMINATED</span>`) {
		t.Errorf("the verdict badge is missing or misgraded:\n%s", html)
	}
	if !strings.Contains(html, "1 open") {
		t.Errorf("the board's shape is not in the header")
	}
	if w, cls := outcomeBadge((record.NewFamily(nil, record.Merged{}))); w != "PENDING" || cls != "unknown" {
		t.Errorf("a run with no outcome must say so rather than show a verdict: %q %q", w, cls)
	}
	_ = board
}

// The join the markdown set cannot make: a mention in one document links into another.
func TestCrossDocumentIdsAreLinkedInTheSite(t *testing.T) {
	html, _ := siteFixture(t)
	if !strings.Contains(html, `data-doc="docket.md">G-00000001</a>`) {
		t.Errorf("the report's mention of G-00000001 does not link to the docket that defines it:\n%s", html)
	}
}

// THE SET'S OWN LINKS ARE DEAD IN ONE FILE unless they are rewritten. `[the board](docket.md)`
// is exactly right in the markdown tier and points at a document that does not exist as a file
// when the whole set is one page.
func TestSiblingLinksBecomeTabSwitches(t *testing.T) {
	html, _ := siteFixture(t)
	if strings.Contains(html, `<a href="docket.md"`) {
		t.Errorf("a sibling link survived as a file path — it opens nothing in a single-file site:\n%s", html)
	}
	files := map[string]bool{FileEvidence: true}
	got := mdToHTML("[P](evidence.md#P-00000001) and [out](https://example.invalid/x)\n", FileReport, anchors{}, map[string]int{}, files).html
	if !strings.Contains(got, `<a class="idref" data-doc="evidence.md" href="#P-00000001"`) {
		t.Errorf("a fragment link into a sibling document was not rewritten: %s", got)
	}
	if !strings.Contains(got, `<a href="https://example.invalid/x">`) {
		t.Errorf("an external link must be left alone: %s", got)
	}
}

// A footnote's LABEL is the definition of that id — linking it away makes every definition
// read as a reference to somewhere else.
func TestFootnoteLabelsAreNotIdLinked(t *testing.T) {
	anchor := anchors{}
	ev := mdToHTML("### P1 — the computation\n", FileEvidence, anchor, map[string]int{}, nil)
	rep := mdToHTML("A claim[^P1].\n\n[^P1]: the computation.\n", FileReport, anchor, map[string]int{}, nil)
	got := linkIDs(rep, anchor, FileReport)
	if strings.Contains(got, `<span class="fnlabel" data-nolink><a`) {
		t.Errorf("the footnote label was turned into a link to somewhere else:\n%s", got)
	}
	_ = ev
}

// ONE PAGE, ONE ELEMENT PER FOOTNOTE ID. Every document with a citation numbers a note 1, and the
// documents are tabs of one page: the second tab's marker and entry under "fn-1" were the first
// tab's to the browser, so the entry's way back and the marker's way down led to a hidden tab.
// Each tab's notes link among themselves, and the label a reader sees is the one written.
func TestFootnoteIdsAreTheSetsNotTheDocuments(t *testing.T) {
	docs := []Doc{
		{File: FileReport, Nav: "Report", Body: "## Fn 1\n\nA claim[^1], again[^1], a proof[^P1] and an aside[^scope].\n\n[^1]: The first source.\n[^P1]: The computation.\n[^scope]: Of the naturals.\n"},
		{File: FileDebate, Nav: "Debate", Body: "A position[^1] on the proof[^P1].\n\n[^1]: Another source.\n[^P1]: The computation.\n"},
		{File: FileRun, Nav: "Run", Body: "A third[^1].\n\n[^1]: A third source.\n"},
	}
	html := RenderSite("# T", Ship(docs), record.Family{})
	arts := regexp.MustCompile(`(?s)<article[^>]*data-doc="([^"]+)"[^>]*>(.*?)</article>`).FindAllStringSubmatch(html, -1)
	if len(arts) != len(docs) {
		t.Fatalf("the page holds %d article(s), want %d", len(arts), len(docs))
	}
	seen := map[string]string{}
	for _, a := range arts {
		here := map[string]bool{}
		for _, m := range regexp.MustCompile(` id="([^"]+)"`).FindAllStringSubmatch(a[2], -1) {
			if in, dup := seen[m[1]]; dup {
				t.Errorf("id %q is an element of %s and of %s", m[1], in, a[1])
			}
			seen[m[1]], here[m[1]] = a[1], true
		}
		for _, m := range regexp.MustCompile(`href="#(fn[^"]*)"`).FindAllStringSubmatch(a[2], -1) {
			if !here[m[1]] {
				t.Errorf("%s links to #%s, which is no element of that tab", a[1], m[1])
			}
		}
	}
	for doc, wants := range map[string][]string{
		FileReport: {`<h2 id="fn-1">`, `id="fnref-1-2" href="#fn-1-2"`, `<li id="fn-1-2">The first source.`, `href="#fnref-1-2"`, `<li id="fn-P1">`, `<li id="fn-scope">`,
			`<a class="cite" href="#fn-1-2" title="1 — The first source.">[1]</a>`, `<span class="fnlabel" data-nolink>1</span>`},
		FileDebate: {`id="fnref-1-3" href="#fn-1-3"`, `<li id="fn-1-3">Another source.`, `href="#fnref-1-3"`, `id="fnref-P1-2" href="#fn-P1-2" title="P1 — The computation."`, `<li id="fn-P1-2">`, `<span class="fnlabel" data-nolink>P1</span>`},
		FileRun:    {`id="fnref-1-4" href="#fn-1-4"`, `<li id="fn-1-4">A third source.`},
	} {
		for _, a := range arts {
			if a[1] != doc {
				continue
			}
			for _, w := range wants {
				if !strings.Contains(a[2], w) {
					t.Errorf("%s does not carry %s:\n%s", doc, w, a[2])
				}
			}
		}
	}
}

// AN ID WITH A COUNT IS AN ID TOO: the second "A" is "a-2", and a heading that reads "A 2" is not
// given that id again.
func TestACountedIdIsNotGivenOutTwice(t *testing.T) {
	html := mdToHTML("## A\n\n## A\n\n## A 2\n\n## A\n", FileReport, anchors{}, map[string]int{}, nil).html
	for _, id := range []string{"a", "a-2", "a-2-2", "a-3"} {
		if n := strings.Count(html, `<h2 id="`+id+`">`); n != 1 {
			t.Errorf("%d heading(s) carry id %q, want one:\n%s", n, id, html)
		}
	}
}
