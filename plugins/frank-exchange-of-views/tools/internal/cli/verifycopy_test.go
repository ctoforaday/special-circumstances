package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// copyRoutes are the copies the cache can hold for a url: its own bytes, and each route by which
// another copy stands in for it.
var copyRoutes = []struct {
	substitution string
	copyURL      string
	want         recordpb.CopySubstitution
}{
	{"", "", recordpb.CopySubstitution_COPY_SUBSTITUTION_NONE},
	{fetchcache.SubstitutionFollowedLink, "https://en.example/full", recordpb.CopySubstitution_COPY_SUBSTITUTION_FOLLOWED_LINK},
	{fetchcache.SubstitutionRecovered, "https://web.archive.org/web/2019id_/" + fermatURL, recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED},
	{fetchcache.SubstitutionFullerCopy, "https://repo.example/open.html", recordpb.CopySubstitution_COPY_SUBSTITUTION_FULLER_COPY},
	{fetchcache.SubstitutionNamedBackend, "https://arxiv.org/html/2101.00001", recordpb.CopySubstitution_COPY_SUBSTITUTION_NAMED_BACKEND},
}

// cachedCopy stores fermatPage under fermatURL as the route leaves it, and refuses any fetch.
func cachedCopy(t *testing.T, runDir, substitution, copyURL string) {
	t.Helper()
	withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
		t.Fatalf("fetched %s past a cached entry", u)
		return nil, nil
	}))
	run := runtest.Open(t, runDir)
	entry := fetchcache.EntryFor(run, fermatURL, &fetchcache.Attempt{Body: []byte(fermatPage), ContentType: "text/html",
		TextRetrieved: true, URL: copyURL}, substitution)
	if _, err := fetchcache.Store(run, entry, []byte(fermatPage)); err != nil {
		t.Fatal(err)
	}
}

// jsonResult runs a verb with --json and returns its result object.
func jsonResult(t *testing.T, args ...string) map[string]any {
	t.Helper()
	out, err := run(t, append([]string{"--json"}, args...)...)
	if err != nil {
		t.Fatalf("%s: %v\n%s", args[0], err, out)
	}
	var env map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &env); err != nil {
		t.Fatalf("%s --json is not JSON (%v): %s", args[0], err, out)
	}
	r, _ := env["result"].(map[string]any)
	return r
}

// evidenceReadings is every verification row `show evidence` renders, anchored and independent.
func evidenceReadings(t *testing.T, runDir string) []map[string]any {
	t.Helper()
	b, err := record.EvidenceJSONBytes(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Sources []struct {
			Verified []map[string]any `json:"verified"`
		} `json:"sources"`
		Independent []map[string]any `json:"independent"`
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("evidence: %v\n%s", err, b)
	}
	out := got.Independent
	for _, s := range got.Sources {
		out = append(out, s.Verified...)
	}
	return out
}

// copyWordsSaid reports whether a verb's human line names the copy as it should: nothing on the
// url's own bytes, the route and where it was read on a substituted copy.
func copyWordsSaid(out, kind, copyURL string) bool {
	if kind == "none" {
		return !strings.Contains(out, "copy_substitution")
	}
	return strings.Contains(out, "copy_substitution "+kind+", read at "+copyURL)
}

// BLUE IS TOLD WHOSE BYTES IT CITED, by the act that recorded them, in its human line and its JSON —
// read from the stored row, so the report is not the first place a substituted copy surfaces.
func TestACiteSaysWhichCopyItRead(t *testing.T) {
	for _, tc := range copyRoutes {
		t.Run(recordpb.Word(tc.want), func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, fermatReport)
			registerBlue(t, runDir)
			cachedCopy(t, runDir, tc.substitution, tc.copyURL)
			r := jsonResult(t, "cite", "--run", runDir, "--seat-id", citeSeat, "--quote", fermatClaim, "--url", fermatURL,
				"--title", "Fermat pseudoprime", "--reason", fermatReason, "--source-span", fermatSpan, "--key", "k1")
			if r["copy_substitution"] != recordpb.Word(tc.want) {
				t.Errorf("cite --json copy_substitution = %v, want %s", r["copy_substitution"], recordpb.Word(tc.want))
			}
			if got, _ := r["copy_url"].(string); got != tc.copyURL {
				t.Errorf("cite --json copy_url = %q, want %q", got, tc.copyURL)
			}
			// The keyed retry returns the stored row, and says the same.
			out, err := fermatCite(t, runDir, "--source-span", fermatSpan, "--key", "k1")
			if err != nil {
				t.Fatalf("cite retry: %v\n%s", err, out)
			}
			if !copyWordsSaid(out, recordpb.Word(tc.want), tc.copyURL) {
				t.Errorf("cite's retry line does not name the copy %s %s:\n%s", recordpb.Word(tc.want), tc.copyURL, out)
			}
		})
	}
	t.Run("the human line", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, fermatReport)
		registerBlue(t, runDir)
		cachedCopy(t, runDir, fetchcache.SubstitutionRecovered, "https://web.archive.org/x")
		out, err := fermatCite(t, runDir, "--source-span", fermatSpan)
		if err != nil {
			t.Fatalf("cite: %v\n%s", err, out)
		}
		if !copyWordsSaid(out, "recovered", "https://web.archive.org/x") {
			t.Errorf("cite's line does not name the recovered copy:\n%s", out)
		}
	})
}

// A VERIFICATION RECORDS THE COPY ITS READING WAS MADE AGAINST — the entry the cache resolves the
// citation's url to — for every route, says so in its line, and `show evidence` shows it on the row.
func TestAVerifyRecordsTheCopyItsReadingWasMadeAgainst(t *testing.T) {
	for _, tc := range copyRoutes {
		t.Run(recordpb.Word(tc.want), func(t *testing.T) {
			runDir := newRun(t)
			writeReport(t, runDir, fermatReport)
			registerBlue(t, runDir)
			cachedCopy(t, runDir, tc.substitution, tc.copyURL)
			if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
				t.Fatalf("cite: %v\n%s", err, out)
			}
			registerLensOnce(t, runDir)
			out, err := verifyAs(t, runDir, firstCiteEvent(t, runDir).GetLabel(), "supports")
			if err != nil {
				t.Fatalf("verify: %v\n%s", err, out)
			}
			v := firstVerify(t, runDir)
			if v.GetCopySubstitution() != tc.want || v.GetCopyUrl() != tc.copyURL || (tc.copyURL == "") != (v.CopyUrl == nil) {
				t.Errorf("recorded copy %v %v, want %v %q", v.GetCopySubstitution(), v.CopyUrl, tc.want, tc.copyURL)
			}
			if !copyWordsSaid(out, recordpb.Word(tc.want), tc.copyURL) {
				t.Errorf("verify's line does not name the copy:\n%s", out)
			}
			rows := evidenceReadings(t, runDir)
			if len(rows) != 1 || rows[0]["copy_substitution"] != recordpb.Word(tc.want) {
				t.Fatalf("the evidence row does not carry copy_substitution %s: %v", recordpb.Word(tc.want), rows)
			}
			if got, _ := rows[0]["copy_url"].(string); got != tc.copyURL {
				t.Errorf("the evidence row's copy_url = %q, want %q", got, tc.copyURL)
			}
		})
	}
}

// THE COPY IS READ AT THE VERIFICATION, not copied off the citation: a cache that no longer holds the
// source — the run resumed in another container — records not_cached, never the cite's word and
// never none.
func TestAVerifyOfASourceTheCacheNoLongerHoldsIsNotCached(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, fermatReport)
	registerBlue(t, runDir)
	cachedCopy(t, runDir, fetchcache.SubstitutionRecovered, "https://web.archive.org/x")
	if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	if err := os.RemoveAll(fetchcache.Dir(runtest.Open(t, runDir))); err != nil {
		t.Fatal(err)
	}
	registerLensOnce(t, runDir)
	out, err := verifyAs(t, runDir, firstCiteEvent(t, runDir).GetLabel(), "supports")
	if err != nil {
		t.Fatalf("verify: %v\n%s", err, out)
	}
	v := firstVerify(t, runDir)
	if v.GetCopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_CACHED || v.CopyUrl != nil {
		t.Errorf("recorded copy %v %v, want not_cached and no url", v.GetCopySubstitution(), v.CopyUrl)
	}
	if !strings.Contains(out, "recorded as not_cached") {
		t.Errorf("verify's line does not say the cache held no copy:\n%s", out)
	}
}

// A CORROBORATION RECORDS THE COPY TOO: the cache's entry for the url red names, or not_cached for a
// source red read outside the tool. A labelled one is a source of the report, so its source row and
// the citation set the report assembles from carry the copy.
func TestACorroborationRecordsTheCopyItsReadingWasMadeAgainst(t *testing.T) {
	corroborate := func(t *testing.T, runDir string) string {
		t.Helper()
		out, err := run(t, "corroborate", "--run", runDir, "--seat-id", lensSeat, "--url", fermatURL, "--title", "Fermat pseudoprime",
			"--quote", fermatClaim, "--as", "supports", "--confidence", "high", "--reason", "read at the leaf")
		if err != nil {
			t.Fatalf("corroborate: %v\n%s", err, out)
		}
		return out
	}
	t.Run("recovered", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, fermatReport)
		cachedCopy(t, runDir, fetchcache.SubstitutionRecovered, "https://web.archive.org/x")
		registerLensOnce(t, runDir)
		out := corroborate(t, runDir)
		v := firstVerify(t, runDir)
		if v.GetCopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED || v.GetCopyUrl() != "https://web.archive.org/x" {
			t.Errorf("recorded copy %v %q, want recovered from the snapshot", v.GetCopySubstitution(), v.GetCopyUrl())
		}
		if !copyWordsSaid(out, "recovered", "https://web.archive.org/x") {
			t.Errorf("corroborate's line does not name the copy:\n%s", out)
		}
		srcs, err := record.CitedSources(runtest.Open(t, runDir))
		if err != nil {
			t.Fatal(err)
		}
		if len(srcs) != 1 || srcs[0].CopySubstitution != recordpb.CopySubstitution_COPY_SUBSTITUTION_RECOVERED || srcs[0].CopyURL != "https://web.archive.org/x" {
			t.Errorf("the corroboration's footnote does not carry its copy: %+v", srcs)
		}
		word, copyURL := evidenceCopy(t, runDir)
		if word != "recovered" || copyURL == nil || *copyURL != "https://web.archive.org/x" {
			t.Errorf("the corroboration's source row: copy_substitution %q copy_url %v", word, copyURL)
		}
	})
	t.Run("not cached", func(t *testing.T) {
		runDir := newRun(t)
		writeReport(t, runDir, fermatReport)
		withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
			t.Fatalf("corroborate fetched %s", u)
			return nil, nil
		}))
		registerLensOnce(t, runDir)
		corroborate(t, runDir)
		v := firstVerify(t, runDir)
		if v.GetCopySubstitution() != recordpb.CopySubstitution_COPY_SUBSTITUTION_NOT_CACHED || v.CopyUrl != nil {
			t.Errorf("recorded copy %v %v, want not_cached and no url", v.GetCopySubstitution(), v.CopyUrl)
		}
		// The check renders under independent and on the corroboration's own source row.
		rows := evidenceReadings(t, runDir)
		if len(rows) != 2 || rows[0]["copy_substitution"] != "not_cached" || rows[1]["copy_substitution"] != "not_cached" {
			t.Errorf("the evidence rows do not say not_cached: %v", rows)
		}
	})
}

// A VERIFICATION READS THE CITATION'S COPY, so absent is refused on a citation whose copy the cache
// holds as the abstract, as it is on a corroboration of one.
func TestAbsentIsRefusedOnACitationWhoseCopyIsOnlyTheAbstract(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, fermatReport)
	registerBlue(t, runDir)
	withFetcher(t, fetchFunc(func(u string) (*fetchcache.Response, error) {
		t.Fatalf("fetched %s past a cached entry", u)
		return nil, nil
	}))
	run := runtest.Open(t, runDir)
	entry := fetchcache.EntryFor(run, fermatURL, &fetchcache.Attempt{Body: []byte(fermatPage), ContentType: "text/html", TextRetrieved: true}, "")
	entry.Completeness, entry.CompletenessReason = fetchcache.CompletenessAbstract, "the platform printed its paywall in place of the body"
	if _, err := fetchcache.Store(run, entry, []byte(fermatPage)); err != nil {
		t.Fatal(err)
	}
	if out, err := fermatCite(t, runDir, "--source-span", fermatSpan); err != nil {
		t.Fatalf("cite: %v\n%s", err, out)
	}
	registerLensOnce(t, runDir)
	if out, err := verifyAs(t, runDir, firstCiteEvent(t, runDir).GetLabel(), "absent"); err == nil || !strings.Contains(err.Error(), "abstract") {
		t.Fatalf("absent was recorded against a citation whose copy is only the abstract: %v\n%s", err, out)
	}
}
