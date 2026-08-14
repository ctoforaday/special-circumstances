package report

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// §V.4 — the citation weave: invisible anchors → visible [^N] + a composed ## Bibliography.

func TestWeaveCitationsNumbersInOrderAndComposesBibliography(t *testing.T) {
	md := "Alpha holds<!--cite:c-1-->. Beta holds<!--cite:c-2-->. Alpha again<!--cite:c-1-->."
	sources := []record.Source{
		{Label: "c-1", URL: "https://a", Title: "Source A", AccessDate: "2026-08-03"},
		{Label: "c-2", URL: "https://b", Title: "Source B", AccessDate: "2026-08-03"},
	}
	got := weaveCitations(md, sources)

	// First-appearance order: c-1 → [^1], c-2 → [^2]; the repeat of c-1 shares [^1].
	if !strings.Contains(got, "Alpha holds[^1]. Beta holds[^2]. Alpha again[^1].") {
		t.Errorf("anchors not woven in first-appearance order:\n%s", got)
	}
	// No invisible anchors survive.
	if strings.Contains(got, "<!--cite:") {
		t.Errorf("a citation anchor survived the weave:\n%s", got)
	}
	// The bibliography is composed, one line per distinct source, in [^N] order.
	if !strings.Contains(got, "## Bibliography") ||
		!strings.Contains(got, "[^1]: Source A. https://a (accessed 2026-08-03)") ||
		!strings.Contains(got, "[^2]: Source B. https://b (accessed 2026-08-03)") {
		t.Errorf("bibliography not composed as expected:\n%s", got)
	}
}

func TestWeaveCitationsNoCitationsIsUnchanged(t *testing.T) {
	md := "Plain prose with no citations at all."
	if got := weaveCitations(md, nil); got != md {
		t.Errorf("weave of an uncited report changed it: %q", got)
	}
	if strings.Contains(weaveCitations(md, nil), "## Bibliography") {
		t.Error("an uncited report got an empty bibliography")
	}
}

// A dangling anchor — present in the document but with no source on the record (the
// crash-window orphan; bijection-impossible under the lockdown) — is SURFACED as an
// unresolved line, never dropped or crashed.
func TestWeaveCitationsSurfacesDanglingAnchor(t *testing.T) {
	md := "A claim with a stranded anchor<!--cite:c-dead-->."
	got := weaveCitations(md, nil) // no sources
	if !strings.Contains(got, "A claim with a stranded anchor[^1].") {
		t.Errorf("dangling anchor not rewritten to a footnote ref:\n%s", got)
	}
	if !strings.Contains(got, "[^1]: _(unresolved citation c-dead — no source on the record)_") {
		t.Errorf("dangling anchor not surfaced as unresolved:\n%s", got)
	}
}

// Findings are STRIPPED and citations are RESOLVED — orthogonal passes over one report.
func TestAssembleStripsFindingsAndResolvesCitations(t *testing.T) {
	runDir := t.TempDir()
	blue := strings.Join([]string{
		"# Coherence — research report",
		"",
		"## TL;DR",
		// One finding marker (must be STRIPPED) and one citation anchor (must be RESOLVED),
		// on the same sentence.
		"The cache is coherent<!--fx:f-aaa--><!--cite:c-1-->.",
		"",
		// Blue also hand-authors a Footnotes section — it must be DROPPED (tool composes it).
		"## Footnotes",
		"[^old]: blue's own hand-typed footnote.",
		"",
	}, "\n")
	if err := os.MkdirAll(filepath.Join(runDir, "blue"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(runDir, "blue", "report.md"), []byte(blue), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, err := record.RegisterSeat(runDir, "blue-synth-r0"); err != nil {
		t.Fatal(err)
	}
	cite := record.NewPayload().
		Set("label", "c-1").Set("url", "https://ex/coherence").Set("sha256", "deadbeef").
		Set("title", "Coherence Proof").Set("access_date", "2026-08-03")
	if _, err := record.Append(runDir, "blue-synth-r0", "cite", cite); err != nil {
		t.Fatal(err)
	}

	path, err := Assemble(runDir)
	if err != nil {
		t.Fatalf("Assemble: %v", err)
	}
	b, _ := os.ReadFile(path)
	got := string(b)

	if strings.Contains(got, "<!--fx:") {
		t.Error("a finding marker survived assembly")
	}
	if strings.Contains(got, "<!--cite:") {
		t.Error("a citation anchor survived assembly (should be woven to [^N])")
	}
	if !strings.Contains(got, "The cache is coherent[^1].") {
		t.Errorf("citation not woven into the lifted TL;DR:\n%s", got)
	}
	if !strings.Contains(got, "## Bibliography") || !strings.Contains(got, "[^1]: Coherence Proof. https://ex/coherence (accessed 2026-08-03)") {
		t.Errorf("composed bibliography missing:\n%s", got)
	}
	if strings.Contains(got, "blue's own hand-typed footnote") {
		t.Error("blue's hand-authored footnote leaked through (must be dropped)")
	}
}

// The claim count is invariant across the weave: it is computed on the PRE-assembly anchor
// report, and the visible [^N] the weave produces is never counted (nothing counts the
// assembled report). This pins that the unit is the anchor, uniformly.
func TestClaimCountIsAnchorBasedNotFootnoteBased(t *testing.T) {
	pre := "One<!--cite:c-1-->. Two<!--cite:c-2-->."
	if got := claimcount.Count(pre); got != 2 {
		t.Fatalf("pre-assembly Count = %d, want 2", got)
	}
	// After the weave the anchors are [^1]/[^2] — and those are NOT counted (the migration
	// dropped [^N] matching), so a post-assembly count would be zero. The system never counts
	// the assembled report; this asserts the anchor is the sole claim unit.
	post := weaveCitations(pre, []record.Source{{Label: "c-1", URL: "u1", Title: "t1"}, {Label: "c-2", URL: "u2", Title: "t2"}})
	if got := claimcount.Count(post); got != 0 {
		t.Errorf("post-weave Count = %d, want 0 (footnote refs are not claims; only anchors are)", got)
	}
}
