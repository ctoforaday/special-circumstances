package cli

import (
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
)

// blue cite is blue's ONLY citation mechanism: it fetches a source through the run cache
// and splices an INVISIBLE, IMMORTAL <!--cite:c-…--> anchor at the quoted sentence. These
// drive the real command tree so the fetch-once cache, the invisible anchor, the reject
// paths, and the auto-friction on an unreachable source are pinned where they are enforced.

const citeSeat = blueSeat

var citeAnchorRe = regexp.MustCompile(`<!--cite:(c-[0-9a-f]+)-->`)

// firstCiteEvent returns the cite BODY, typed. The label is tool-assigned, so the event is the
// source of truth for it and the test has to read the record rather than the command's output.
func firstCiteEvent(t *testing.T, runDir string) *recordpb.Cite {
	t.Helper()
	m, err := record.MergedEvents(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range m.Events {
		if c, ok := recordpb.BodyAs[*recordpb.Cite](e); ok {
			return c
		}
	}
	return nil
}

func TestBlueCiteAnchorsInvisiblyAndRecordsEvent(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# Findings\n\nThe sky is blue and the grass is green.\n")
	registerBlue(t, runDir)
	body := []byte("<html>a source that says the sky is blue</html>")
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://sky/1": body}})

	out, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `# Findings: "The sky is blue and the grass is green."`,
		"--url", "https://sky/1", "--title", "Sky Facts")
	if err != nil {
		t.Fatalf("blue cite: %v (out %q)", err, out)
	}

	// The label is tool-assigned; read it from the cite event (the source of truth).
	ev := firstCiteEvent(t, runDir)
	if ev == nil {
		t.Fatal("no cite event recorded")
	}
	label := ev.GetLabel()

	report := readReport(t, runDir)
	// The anchor is spliced just past the last CONTENT byte of the quote (before the
	// trailing period — #248 trims trailing punctuation) and is INVISIBLE (an HTML comment).
	if !strings.Contains(report, "grass is green<!--cite:"+label+"-->.") {
		t.Errorf("anchor not spliced at the quote's content end:\n%s", report)
	}
	// Visible prose is unchanged: stripping the anchor restores the seeded text.
	if stripped := citeAnchorRe.ReplaceAllString(report, ""); stripped != "# Findings\n\nThe sky is blue and the grass is green.\n" {
		t.Errorf("citing changed the visible prose: %q", stripped)
	}
	// The source was cached at <run>/cache/<sha256>.
	sha := fetchcache.Sha(body)
	if _, statErr := os.Stat(fetchcache.Path(runtest.Open(t, runDir), sha)); statErr != nil {
		t.Errorf("source not cached: %v", statErr)
	}
	// The one cite event carries the full citation record.
	if ev.GetUrl() != "https://sky/1" ||
		ev.GetSha256() != sha || ev.GetTitle() != "Sky Facts" {
		t.Errorf("cite event payload wrong: label=%q url=%q sha=%q title=%q",
			ev.GetLabel(), ev.GetUrl(), ev.GetSha256(), ev.GetTitle())
	}
	if ev.GetAccessDate() == "" {
		t.Error("cite event missing engine-supplied access_date")
	}
}

func TestBlueCiteMisQuoteRejectedNoEffect(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe scheduler is preemptive.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://x": []byte("src")}})

	_, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `"the scheduler is cooperative"`, "--url", "https://x", "--title", "T")
	if err == nil {
		t.Fatal("a mis-quoted cite was accepted")
	}
	if strings.Contains(readReport(t, runDir), "<!--cite:") {
		t.Error("a rejected cite still spliced an anchor")
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); n != 0 {
		t.Errorf("a rejected cite recorded %d cite events, want 0", n)
	}
}

func TestBlueCiteInFenceRejected(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\n```\ncode line here\n```\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{resp: map[string][]byte{"https://x": []byte("src")}})

	_, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `"code line here"`, "--url", "https://x", "--title", "T")
	if err == nil || !strings.Contains(err.Error(), "fence") {
		t.Fatalf("citing inside a fence = %v, want a fence rejection", err)
	}
}

func TestBlueCiteFetchFailureRejectsAndFrictions(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe claim holds under load.\n")
	registerBlue(t, runDir)
	withFetcher(t, &fakeFetcher{err: errFake})

	_, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `"The claim holds under load."`, "--url", "https://gone", "--title", "T")
	if err == nil {
		t.Fatal("citing an unreachable source was accepted")
	}
	// Rejected: no anchor, no cite event.
	if strings.Contains(readReport(t, runDir), "<!--cite:") {
		t.Error("an unusable cite still spliced an anchor")
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); n != 0 {
		t.Errorf("an unusable cite recorded %d cite events, want 0", n)
	}
	// But the DECISION to cite an unreachable source IS surfaced as friction.
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_LOG); n != 1 {
		t.Errorf("an unusable cite recorded %d friction events, want 1", n)
	}
}

func TestBlueCiteReusesCacheAcrossTwoCites(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nFirst claim stands. Second claim stands too.\n")
	registerBlue(t, runDir)
	f := &fakeFetcher{resp: map[string][]byte{"https://same": []byte("one source")}}
	withFetcher(t, f)

	if _, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `"First claim stands."`, "--url", "https://same", "--title", "T"); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, "cite", "--run", runDir, "--seat-id", citeSeat,
		"--quote", `"Second claim stands too."`, "--url", "https://same", "--title", "T"); err != nil {
		t.Fatal(err)
	}
	if f.calls != 1 {
		t.Errorf("two cites of one URL fetched %d times, want 1 (download-once)", f.calls)
	}
	if got := len(citeAnchorRe.FindAllString(readReport(t, runDir), -1)); got != 2 {
		t.Errorf("two cites produced %d anchors, want 2", got)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); n != 2 {
		t.Errorf("two cites recorded %d cite events, want 2", n)
	}
}

func TestBlueCiteKeyIsIdempotent(t *testing.T) {
	runDir := newRun(t)
	writeReport(t, runDir, "# H\n\nThe measured latency is bounded.\n")
	registerBlue(t, runDir)
	f := &fakeFetcher{resp: map[string][]byte{"https://x": []byte("src")}}
	withFetcher(t, f)

	args := []string{"cite", "--run", runDir, "--seat-id", citeSeat, "--key", "C1",
		"--quote", `"The measured latency is bounded."`, "--url", "https://x", "--title", "T"}
	if _, err := run(t, args...); err != nil {
		t.Fatal(err)
	}
	if _, err := run(t, args...); err != nil { // retry same key
		t.Fatal(err)
	}
	if got := len(citeAnchorRe.FindAllString(readReport(t, runDir), -1)); got != 1 {
		t.Errorf("a retried cite produced %d anchors, want 1 (idempotent)", got)
	}
	if n := countType(t, runDir, recordpb.EventType_EVENT_TYPE_CITE); n != 1 {
		t.Errorf("a retried cite recorded %d cite events, want 1", n)
	}
}
