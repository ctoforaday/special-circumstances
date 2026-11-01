package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

const (
	fermatURL = "https://en.example/Fermat_pseudoprime"
	// fermatPage is the definition universe-m30's blue cited: the base a is the liar, and the
	// composite n the pseudoprime.
	fermatPage = "<html><body><h1>Fermat pseudoprime</h1>" +
		"<p>Fermat's little theorem states that if p is prime and a is coprime to p, then a<sup>p-1</sup> - 1 is divisible by p. " +
		"For a composite n, an integer a for which a<sup>n-1</sup> - 1 is divisible by n when n is composite is known as a Fermat liar. " +
		"In this case n is called Fermat pseudoprime to base a.</p>" +
		"<p>The photo-<br>synthesis section is unrelated filler.</p></body></html>"
	fermatClaim  = "A composite number that passes for some bases is called a Fermat liar for those bases."
	fermatReport = "# Findings\n\n" + fermatClaim + "\n"
	// fermatReason is blue's paraphrase as m30 stored it: inverted, and worded like the source.
	fermatReason = "Source: a composite n that passes the Fermat congruence for some base a is known as a Fermat liar; n is then called Fermat pseudoprime to base a."
	fermatSpan   = "an integer a for which a^(n-1) - 1 is divisible by n when n is composite is known as a Fermat liar. In this case n is called Fermat pseudoprime to base a."
)

func fermatRun(t *testing.T, contentType string) string {
	t.Helper()
	runDir := newRun(t)
	writeReport(t, runDir, fermatReport)
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{fermatURL: []byte(fermatPage)}, contentType: contentType})
	return runDir
}

func fermatCite(t *testing.T, runDir string, args ...string) (string, error) {
	t.Helper()
	return run(t, append([]string{"cite", "--run", runDir, "--seat-id", citeSeat, "--quote", fermatClaim,
		"--url", fermatURL, "--title", "Fermat pseudoprime", "--reason", fermatReason}, args...)...)
}

// THE m30 CASE. Blue's reason inverts the definition and reads like the source; the evidence entry
// a lens resolves the citation through shows the source's own words at the span, and blue's
// reason under a name that says whose it is — never the paraphrase as the source.
func TestALensSeesTheSourcesWordsAtTheSpanBesideBluesReason(t *testing.T) {
	runDir := fermatRun(t, "text/html")
	if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
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
	src := raw.Sources[0]
	if _, ok := src["text"]; ok {
		t.Errorf("the source entry carries a bare `text`, which a lens reads as the source's:\n%s", b)
	}
	var reason string
	if err := json.Unmarshal(src["blue_reason"], &reason); err != nil || reason != fermatReason {
		t.Errorf("blue_reason = %q (%v), want blue's reason labelled as blue's", reason, err)
	}
	var st record.EvidenceSourceText
	if err := json.Unmarshal(src["source_text"], &st); err != nil {
		t.Fatalf("source_text: %v\n%s", err, b)
	}
	if st.Matched != fermatSpan || !strings.Contains(st.Before, "For a composite n, ") || !strings.Contains(st.After, "photo-") {
		t.Errorf("source_text = %+v; want the page's words at the span and around it", st)
	}
	if strings.Contains(st.Before+st.Matched+st.After, "passes the Fermat congruence") {
		t.Error("the source text holds blue's paraphrase")
	}
	c := firstCiteEvent(t, runDir)
	if c.GetSourceTextOrigin() != recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_EMBEDDED || len(c.GetPages()) != 0 || c.GetSpanEnd() <= c.GetSpanStart() {
		t.Errorf("cite: origin %v, pages %v, span [%d,%d)", c.GetSourceTextOrigin(), c.GetPages(), c.GetSpanStart(), c.GetSpanEnd())
	}
}

// BLUE'S OWN WORDING IS NOT A SPAN: the paraphrase is not in the page, so the cite is refused and
// nothing is recorded. So is a cite with no span at all of a source whose text the tool holds.
func TestACiteOfTextTheSourceDoesNotHoldIsRefused(t *testing.T) {
	runDir := fermatRun(t, "text/html")
	out, err := fermatCite(t, runDir, "--source-span", "a composite n that passes the Fermat congruence for some base a is known as a Fermat liar")
	mustRefuse(t, err, out, "--source-span", "not in the text at")
	out, err = fermatCite(t, runDir)
	mustRefuse(t, err, out, "the tool holds this source's text", "--source-span")
	if c := firstCiteEvent(t, runDir); c != nil {
		t.Errorf("a refused cite left an event: %v", c)
	}
}

// A PAGE'S LINE BREAK INSIDE A HYPHENATED WORD is found with the word typed whole.
func TestAnHTMLSpanIsFoundAcrossALineEndHyphen(t *testing.T) {
	runDir := fermatRun(t, "text/html")
	if out, err := fermatCite(t, runDir, "--source-span", "The photosynthesis section is unrelated filler"); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if got := firstCiteEvent(t, runDir).GetSourceMatched(); got != "The photo-\nsynthesis section is unrelated filler" {
		t.Errorf("matched %q, want the page's own bytes across the break", got)
	}
}

// A SOURCE WITH NO TEXT is cited without a span, and the citation and the verb's answer both say
// so; a span given for one is refused, never dropped.
func TestASourceWithNoTextIsCitedWithoutASpanAndSaysSo(t *testing.T) {
	runDir := fermatRun(t, "image/png")
	out, err := fermatCite(t, runDir, "--source-span", fermatSpan)
	mustRefuse(t, err, out, "the tool holds none for", "(none)")
	out, err = fermatCite(t, runDir)
	if err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if !strings.Contains(out, "carries no source span: the tool holds no text for this source") {
		t.Errorf("the answer does not say the citation carries no span:\n%s", out)
	}
	c := firstCiteEvent(t, runDir)
	if c.GetSourceTextOrigin() != recordpb.SourceTextOrigin_SOURCE_TEXT_ORIGIN_NONE || c.SourceSpan != nil || c.SpanStart != nil {
		t.Errorf("cite: origin %v, span %q", c.GetSourceTextOrigin(), c.GetSourceSpan())
	}
}
