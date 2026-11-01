package cli

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// corroborationRow is the one corroboration's source row `show evidence` renders.
func corroborationRow(t *testing.T, runDir string) record.EvidenceSourceJSON {
	t.Helper()
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var got record.EvidenceJSON
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, s := range got.Sources {
		if s.CorroboratedBy != "" {
			return s
		}
	}
	t.Fatalf("no corroboration on the evidence view:\n%s", b)
	return record.EvidenceSourceJSON{}
}

// A CORROBORATION'S ROW NAMES THE COPY IT READ, as a cite's does. On #1677's smoke the row's sha256 was
// empty (C-20761e81), so `show source --anchor` searched for the empty sha and a lens had no way
// to name the copy behind a footnote of the report.
func TestACorroborationsRowNamesTheCopyItRead(t *testing.T) {
	runDir := corroborateRun(t)
	cachedCopy(t, runDir, "", "")
	held, _, cached, err := fetchcache.Lookup(runtest.Open(t, runDir), fermatURL)
	if err != nil || !cached {
		t.Fatalf("the seeded copy is not in the cache: %v, %v", cached, err)
	}
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", fermatURL, "--title", "A source red found", "--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the source says exactly this at page 4"); err != nil {
		t.Fatalf("a supporting corroboration was refused: %v", err)
	}
	row := corroborationRow(t, runDir)
	if row.Sha256 != held.Sha {
		t.Errorf("the corroboration's row says sha256 %q, and the copy it read is %q", row.Sha256, held.Sha)
	}
	out, err := run(t, "show", "source", "--run", runDir, "--seat-id", "red-lens-evidence", "--anchor", row.Anchor, "--quote", "Fermat pseudoprime")
	if err != nil {
		t.Fatalf("show source of the corroboration: %v\n%s", err, out)
	}
	if !strings.Contains(out, held.TextSha) {
		t.Errorf("show source --anchor did not search the copy the corroboration read:\n%s", out)
	}
}

// A CORROBORATION THE CACHE HELD NO COPY OF HAS NO SHA TO NAME, AND THE SEARCH SAYS SO, rather than
// asking the cache for the empty sha.
func TestACorroborationOfNoCachedCopyHasNothingToSearch(t *testing.T) {
	runDir := corroborateRun(t)
	if _, err := run(t, "corroborate", "--run", runDir, "--seat-id", "red-lens-evidence",
		"--url", "https://example.org/red-found-this", "--title", "A source red found", "--quote", corroborated, "--as", "supports", "--confidence", "high",
		"--reason", "the source says exactly this at page 4"); err != nil {
		t.Fatalf("a supporting corroboration was refused: %v", err)
	}
	row := corroborationRow(t, runDir)
	if row.Sha256 != "" || row.CopySubstitution != "not_cached" {
		t.Fatalf("a corroboration of no cached copy reads sha256 %q, copy_substitution %q", row.Sha256, row.CopySubstitution)
	}
	_, err := run(t, "show", "source", "--run", runDir, "--seat-id", "red-lens-evidence", "--anchor", row.Anchor, "--quote", "anything")
	if err == nil || !strings.Contains(err.Error(), "held no copy of that source") {
		t.Errorf("show source of a corroboration with no copy = %v, want the refusal saying the cache held none", err)
	}
}
