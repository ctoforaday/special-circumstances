package cli

import (
	"sort"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/claimcount"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// §V.5 — the lockdown class-sweep: `blue edit` protects a citation anchor exactly as it
// protects a finding marker, so the set of cite events stays a strict bijection with the
// citation anchors in the document.

// TestBlueEditRejectsSpanContainingCitation pins the spanMarker site: an --old span that
// straddles a "<!--cite:-->" anchor is refused, and the message names it as a citation.
func TestBlueEditRejectsSpanContainingCitation(t *testing.T) {
	runDir := t.TempDir()
	// A citation anchor sits between "value" and "is".
	writeReport(t, runDir, "# H\n\nThe value<!--cite:c-abc123--> is stable over time.\n")
	registerBlue(t, runDir)

	_, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "value is stable", "--new", "value is steady", "--reason", "wording")
	if err == nil {
		t.Fatal("an edit straddling a citation anchor was accepted")
	}
	if !strings.Contains(err.Error(), "citation anchor") {
		t.Errorf("rejection message = %q, want it to name the citation anchor", err.Error())
	}
	// The report is unchanged — the anchor and its prose both survive.
	if got := readReport(t, runDir); !strings.Contains(got, "value<!--cite:c-abc123--> is stable over time") {
		t.Errorf("a rejected edit changed the report:\n%s", got)
	}
}

// TestCiteAnchorBijection drives real cites + an edit and asserts the cite-event label set
// equals the citation-anchor set in the document — the record shows exactly what is cited.
func TestCiteAnchorBijection(t *testing.T) {
	runDir := t.TempDir()
	writeReport(t, runDir, "# H\n\nAlpha holds under load. Beta holds under load too.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{
		"https://a": []byte("source a"),
		"https://b": []byte("source b"),
	}})

	if _, err := run(t, "blue", "cite", "--run", runDir, "--seat-id", blueSeat,
		"--location", `"Alpha holds under load."`, "--url", "https://a", "--title", "A"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "blue", "cite", "--run", runDir, "--seat-id", blueSeat,
		"--location", `"Beta holds under load too."`, "--url", "https://b", "--title", "B"); err != nil {
		t.Fatal(err)
	}

	assertBijection(t, runDir)

	// An edit that does NOT touch an anchor keeps the bijection.
	if _, err := run(t, "blue", "edit", "--run", runDir, "--seat-id", blueSeat,
		"--key", "E1", "--old", "under load too", "--new", "under heavy load too", "--reason", "precision"); err != nil {
		t.Fatalf("clean edit: %v", err)
	}
	assertBijection(t, runDir)
}

func assertBijection(t *testing.T, runDir string) {
	t.Helper()
	events, err := record.CitationLabels(runDir)
	if err != nil {
		t.Fatal(err)
	}
	anchors := claimcount.CitationAnchorIDs(readReport(t, runDir))
	if !sameSet(events, anchors) {
		t.Errorf("bijection broken: cite events %v != document anchors %v", events, anchors)
	}
}

func sameSet(a, b []string) bool {
	as := append([]string(nil), a...)
	bs := append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}
