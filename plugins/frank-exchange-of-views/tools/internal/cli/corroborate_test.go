package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/report"
)

// corroborateRun is a run with a registered lens and a seeded report.
func corroborateRun(t *testing.T) string {
	t.Helper()
	runDir := recordtest.TmpRun(t)
	if _, err := run(t, "register", "--run", runDir, "--seat-id", "red-lens-evidence"); err != nil {
		t.Fatal(err)
	}
	seedBlueReport(t, runDir)
	return runDir
}

const corroborated = "§2 the finding prose lands in a quoted sentence."

// assembled runs the composer and returns the report's TEXT. Assemble writes the file and
// returns what it wrote, and its report path reads like markdown to a Contains check and always fails to match.
func assembled(t *testing.T, runDir string) string {
	t.Helper()
	a, err := report.Assemble(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	md, err := os.ReadFile(a.Report)
	if err != nil {
		t.Fatal(err)
	}
	return string(md)
}

// A SUPPORTING CORROBORATION BECOMES AN ORDINARY FOOTNOTE.
//
// A human reader cares that the text has appropriate references, not which team inserted them.
// Before this, red's independent corroboration reached no reader of the document at all:
// `internal/report` has no reader for a Verify event, and citationid.go stated red's exclusion
// as a property ("Red's `lens cite` carries no label and is EXCLUDED").
func TestASupportingCorroborationRendersAsAFootnote(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/red-found-this", "--title", "A source red found",
		"--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the source says exactly this at page 4"); err != nil {
		t.Fatalf("a supporting corroboration was refused: %v", err)
	}

	// The anchor is spliced into the report, invisibly, exactly as blue's cite and red's
	// finding markers already are.
	md := readReport(t, runDir)
	if !strings.Contains(string(md), "<!--cite:C-") {
		t.Fatalf("no citation anchor was spliced at the corroborated sentence:\n%s", md)
	}

	// And it reaches the READER, which is the whole point: assembly weaves it into the
	// bibliography with no knowledge of which seat wrote it.
	out := assembled(t, runDir)
	if !strings.Contains(out, "## Bibliography") || !strings.Contains(out, "example.org/red-found-this") {
		t.Errorf("red's corroborating source is absent from the assembled bibliography:\n%s", out)
	}
	if !strings.Contains(out, "A source red found") {
		t.Error("the source's title did not travel into the footnote")
	}
}

// A REFUTING CORROBORATION IS NOT A REFERENCE, and must not become one.
//
// A source that CONTRADICTS the sentence, rendered in the bibliography, reads as backing it —
// and the report's own assembly check already treats a live refuted citation as a failure.
func TestARefutingCorroborationIsNotSplicedAsAFootnote(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/contradicts", "--title", "A source that disagrees",
		"--quote", corroborated, "--as", "refutes", "--confidence", "high",
		"--reason", "the source says the opposite at page 9"); err != nil {
		t.Fatalf("a refuting corroboration was refused — it is red's strongest finding on this axis and must still record: %v", err)
	}
	md := readReport(t, runDir)
	if strings.Contains(string(md), "<!--cite:C-") {
		t.Errorf("a REFUTING source was spliced as a citation — the reader would meet it as a reference backing the sentence it contradicts:\n%s", md)
	}
	if strings.Contains(assembled(t, runDir), "example.org/contradicts") {
		t.Error("a refuting source reached the bibliography")
	}
}

// THE QUOTE MUST BE IN THE LIVE REPORT, so a corroboration of a claim blue has since edited
// away is refused rather than spliced blind — the same rule blue's own cite is held to.
func TestACorroborationOfAnAbsentClaimIsRefused(t *testing.T) {
	runDir := corroborateRun(t)
	_, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/x", "--title", "T",
		"--quote", "a sentence that is nowhere in the report", "--as", "supports",
		"--confidence", "high", "--reason", "r")
	if err == nil {
		t.Fatal("a corroboration of a claim absent from the report was accepted")
	}
	if !strings.Contains(err.Error(), "--quote") {
		t.Errorf("the refusal does not name the flag to fix, so a seat cannot act on it: %v", err)
	}
	// AND NOTHING WAS RECORDED. A rejected splice must leave no event behind, or the record
	// carries a corroboration whose footnote does not exist.
	srcs, err := record.CitedSources(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 0 {
		t.Errorf("the splice was refused and %d source(s) still recorded: %+v", len(srcs), srcs)
	}
}

// A `weak` READING STILL POINTS AT A SOURCE, so it joins the bibliography.
//
// THE ARGUMENT IS SYMMETRY, NOT STRENGTH. When blue cites a source and red grades it `weak`
// through `lens verify`, the footnote STAYS — verify adjudicates and never touches the report.
// Excluding a weak corroboration rendered the same (claim, source, grade) triple as a footnote
// when blue found the source and as nothing when red did, which is the one difference a reader
// must not be able to see. It also left red's reading with no reader and no duty: neither cited
// nor owing a finding, visible only in the evidence projection.
//
// `unreachable` is the honest exclusion and is asserted beside it: red could not read the thing,
// so there is nothing to point a reader at.
func TestAWeakReadingCitesButAnUnreachableOneDoesNot(t *testing.T) {
	for _, tc := range []struct {
		outcome string
		cites   bool
		why     string
	}{
		{"weak", true, "thin support is still support, and the footnote is a pointer rather than an endorsement"},
		{"unreachable", false, "red could not read it, so there is nothing to point the reader at"},
	} {
		t.Run(tc.outcome, func(t *testing.T) {
			runDir := corroborateRun(t)
			if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
				"--url", "https://example.org/"+tc.outcome, "--title", "A source",
				"--quote", corroborated, "--as", tc.outcome, "--confidence", "medium",
				"--reason", "what the source says"); err != nil {
				t.Fatalf("a %s corroboration was refused — every reading must RECORD, whatever it renders: %v", tc.outcome, err)
			}
			got := strings.Contains(assembled(t, runDir), "example.org/"+tc.outcome)
			if got != tc.cites {
				t.Errorf("%s in the bibliography = %v, want %v — %s", tc.outcome, got, tc.cites, tc.why)
			}
		})
	}
}

// SILENCE IN AN ABSTRACT IS NOT SILENCE IN THE WORK. A copy the fetch recorded as the abstract
// cannot establish that a claim is absent from the work; the same finding against a copy that
// carries the body — or one nobody classified — still records.
func TestAbsentIsRefusedOnACopyThatIsOnlyTheAbstract(t *testing.T) {
	runDir := corroborateRun(t)
	const src = "https://doi.org/10.1038/nature06964"
	page := []byte("<html><body>the abstract</body></html>")
	if _, err := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: src, Sha: fetchcache.Sha(page),
		ContentType: "text/html", Completeness: fetchcache.CompletenessAbstract,
		CompletenessReason: "the platform printed its paywall in place of the body"}, page); err != nil {
		t.Fatal(err)
	}
	absent := func(url string) error {
		_, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
			"--url", url, "--title", "A source", "--quote", corroborated, "--as", "absent", "--confidence", "high",
			"--reason", "the claim is not in it")
		return err
	}
	if err := absent(src); err == nil || !strings.Contains(err.Error(), "abstract") {
		t.Fatalf("absent was recorded against a copy that is only the abstract: %v", err)
	}
	if err := absent("https://example.org/a-source-never-fetched"); err != nil {
		t.Fatalf("absent against an unclassified source was refused: %v", err)
	}
}

// RED MAY CONFIRM FROM AN ABSTRACT, AND THE RECORD SAYS IT DID (gblock, 2026-09-26). A `supports`
// read off the abstract confirms what the abstract says; the verdict carries the copy's
// completeness so that nothing downstream reads it as a verdict on the study. A source outside the
// run's cache was never classified, and says so.
func TestASupportFromAnAbstractRecordsThatItWasOne(t *testing.T) {
	runDir := corroborateRun(t)
	const src = "https://doi.org/10.1038/nature06964"
	page := []byte("<html><body>the abstract</body></html>")
	if _, err := fetchcache.Store(runtest.Open(t, runDir), fetchcache.Entry{URL: src, Sha: fetchcache.Sha(page),
		ContentType: "text/html", Completeness: fetchcache.CompletenessAbstract,
		CompletenessReason: "the platform printed its paywall in place of the body"}, page); err != nil {
		t.Fatal(err)
	}
	supports := func(url string) {
		t.Helper()
		if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
			"--url", url, "--title", "A source", "--quote", corroborated, "--as", "supports", "--confidence", "high",
			"--reason", "the abstract states it"); err != nil {
			t.Fatalf("a supports from %s was refused: %v", url, err)
		}
	}
	supports(src)
	supports("https://example.org/never-fetched")
	got := map[string]recordpb.SourceCompleteness{}
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if v, ok := recordpb.BodyAs[*recordpb.Verify](e); ok {
			got[v.GetUrl()] = v.GetSourceCompleteness()
		}
	}
	if got[src] != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_ABSTRACT {
		t.Errorf("a supports read off an abstract records %v", got[src])
	}
	if g := got["https://example.org/never-fetched"]; g != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED {
		t.Errorf("a source outside the cache records %v, want not_asked", g)
	}
	// AND RED'S LOOKUP TABLE SHOWS IT, which is where the verdict is read.
	ev := record.EvidenceJSONOf(m.Events, m.At)
	seen := false
	for _, v := range ev.Independent {
		if v.URL == src && v.SourceCompleteness == "abstract" {
			seen = true
		}
	}
	if !seen {
		t.Errorf("the evidence view does not show the verdict rests on an abstract: %+v", ev.Independent)
	}
	// A SUPPORTING CORROBORATION IS A FOOTNOTE, and the reader is told it rests on an abstract.
	if out := assembled(t, runDir); !strings.Contains(out, "https://doi.org/10.1038/nature06964") ||
		!strings.Contains(out, "**[ABSTRACT ONLY]**") {
		t.Errorf("red's abstract-backed corroboration is not marked for the reader:\n%s", out)
	}
}
