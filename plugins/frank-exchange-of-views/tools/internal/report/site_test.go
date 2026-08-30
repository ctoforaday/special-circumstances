package report

import (
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
)

func siteFixture(t *testing.T) (string, *record.Board) {
	t.Helper()
	board := &record.Board{
		GapOrder: []string{"R1-1"},
		Gaps:     map[string]*record.Gap{"R1-1": {ID: "R1-1", Open: true}},
		Events: []*record.Event{
			recordtest.Event(t, "", 0, &recordpb.Outcome{Verdict: recordtest.P(recordpb.RunOutcome_RUN_OUTCOME_CEILING)}),
		},
	}
	docs := []Doc{
		{File: FileReport, Nav: "Report", Blurb: "the research", Body: "## Read this first\n\nR1-1 is still open.\n"},
		{File: FileDocket, Nav: "Docket", Blurb: "the gaps", Body: "### R1-1 — eviction races the reader\n\ncache.go:88\n"},
	}
	return RenderSite("# Whether the cache is coherent", docs, board), board
}

// ONE FILE. A reader opens it out of a tarball months later, offline — so nothing it needs may
// live anywhere else. This is the property the whole tier is for, and it is the one a later
// convenience (a CDN stylesheet, a font, an icon set) silently takes away.
func TestTheSiteIsSelfContained(t *testing.T) {
	html, _ := siteFixture(t)
	for _, forbidden := range []string{"http://", "https://cdn", "<link", "<img", "src=\"http"} {
		if strings.Contains(html, forbidden) {
			t.Errorf("the site reaches outside itself (%q) — it must open offline from an archive", forbidden)
		}
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
	if w, cls := verdictBadge(&record.Board{}); w != "no terminal outcome recorded" || cls != "unknown" {
		t.Errorf("a run with no outcome must say so rather than show a verdict: %q %q", w, cls)
	}
	_ = board
}

// The join the markdown set cannot make: a mention in one document links into another.
func TestCrossDocumentIdsAreLinkedInTheSite(t *testing.T) {
	html, _ := siteFixture(t)
	if !strings.Contains(html, `data-doc="docket.md">R1-1</a>`) {
		t.Errorf("the report's mention of R1-1 does not link to the docket that defines it:\n%s", html)
	}
}

// THE SET'S OWN LINKS ARE DEAD IN ONE FILE unless they are rewritten. `[the docket](docket.md)`
// is exactly right in the markdown tier and points at a document that does not exist as a file
// when the whole set is one page.
func TestSiblingLinksBecomeTabSwitches(t *testing.T) {
	html, _ := siteFixture(t)
	if strings.Contains(html, `<a href="docket.md"`) {
		t.Errorf("a sibling link survived as a file path — it opens nothing in a single-file site:\n%s", html)
	}
	files := map[string]bool{FileEvidence: true}
	got := siteLinks(`<a href="evidence.md#p-1">P</a> and <a href="https://example.invalid/x">out</a>`, files)
	if !strings.Contains(got, `<a class="idref" data-doc="evidence.md" href="#p-1"`) {
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
	ev := mdToHTML("### P1 — the computation\n", FileEvidence, anchor)
	rep := mdToHTML("A claim[^P1].\n\n[^P1]: the computation.\n", FileReport, anchor)
	got := linkIDs(rep, anchor, FileReport)
	if strings.Contains(got, `<span class="fnlabel" data-nolink><a`) {
		t.Errorf("the footnote label was turned into a link to somewhere else:\n%s", got)
	}
	_ = ev
}
