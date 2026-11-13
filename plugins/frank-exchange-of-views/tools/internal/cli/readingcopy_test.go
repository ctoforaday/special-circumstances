package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/fetchcache"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/recordpb"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// cacheAs makes the run's cached copy of url one of completeness c: a copy the cite already stored is
// restamped in the index (the cite needed a copy that is the work; the reading resolves to it after),
// and a url the cache does not hold is stored afresh.
func (b readingBoard) cacheAs(url, c string) {
	b.t.Helper()
	run := runtest.Open(b.t, b.runDir)
	if _, _, held, err := fetchcache.Lookup(run, url); err != nil {
		b.t.Fatal(err)
	} else if !held {
		body := []byte("<html>a copy of " + url + "</html>")
		entry := fetchcache.EntryFor(run, url, &fetchcache.Attempt{FetchedAt: fetchcache.FetchTime(), Body: body, ContentType: "text/html", TextRetrieved: true}, "")
		if _, err := fetchcache.Store(run, entry, body); err != nil {
			b.t.Fatal(err)
		}
	}
	index := filepath.Join(fetchcache.Dir(run), "index")
	raw, err := os.ReadFile(index)
	if err != nil {
		b.t.Fatal(err)
	}
	var out []string
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		var e fetchcache.Entry
		if json.Unmarshal([]byte(line), &e) == nil && e.URL == url {
			e.Completeness, e.CompletenessReason = c, "the fixture says so"
			nb, err := json.Marshal(e)
			if err != nil {
				b.t.Fatal(err)
			}
			line = string(nb)
		}
		out = append(out, line)
	}
	if err := os.WriteFile(index, []byte(strings.Join(out, "\n")+"\n"), 0o644); err != nil {
		b.t.Fatal(err)
	}
}

// failedOf is the standing failed checks of anchor.
func failedOf(t *testing.T, runDir, anchor string) []record.Blocker {
	t.Helper()
	bs, err := record.PassBlockers(runtest.Open(t, runDir))
	if err != nil {
		t.Fatal(err)
	}
	var out []record.Blocker
	for _, b := range bs {
		if b.Kind == record.BlockerFailedCheck && b.Subject == anchor {
			out = append(out, b)
		}
	}
	return out
}

// 19. A reading of a copy that is not the work reads nothing of the source: it leaves a citation
// unread, whatever its outcome, and owes no reading debt (GAP 1, GAP 2).
func TestAReadingOfACopyThatIsNotTheWorkMintsNothing(t *testing.T) {
	ntw := func(t *testing.T) readingBoard {
		b := readingRun(t)
		b.cacheAs(rdU1, fetchcache.CompletenessNotTheWork)
		b.cacheAs(rdU3, fetchcache.CompletenessNotTheWork)
		return b
	}
	refused := func(t *testing.T, b readingBoard, args []string, says ...string) {
		t.Helper()
		_, err := b.verify(lensL, args...)
		mustRefuse(t, err, "", says...)
		if c := b.counts(); c != [3]int{} {
			t.Errorf("the refusal wrote verify/mint/anchor %v", c)
		}
	}
	readingAlone := func(t *testing.T, b readingBoard, r readingJSON) *recordpb.Verify {
		t.Helper()
		if r.Debt != "" {
			t.Errorf("--json carries a reading debt: %q — a reading that reads nothing owes nothing", r.Debt)
		}
		if c := b.counts(); c[1] != 0 || c[2] != 0 {
			t.Errorf("the reading wrote mint/anchor %v", c)
		}
		v, _ := b.lastVerify()
		if v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_THE_WORK {
			t.Errorf("the reading's stamp is %v, want not_the_work", v.GetSourceCompleteness())
		}
		return v
	}
	unread := func(t *testing.T, b readingBoard) {
		t.Helper()
		fs := failedOf(t, b.runDir, b.c1)
		if len(fs) != 1 || fs[0].CheckState() != record.CheckUnread || fs[0].Owner != lensL {
			t.Errorf("C's failed checks = %+v, want one unread owned by %s", fs, lensL)
		}
	}

	t.Run("b: --anchor refutes", func(t *testing.T) {
		b := ntw(t)
		r := b.mustRead(lensL, citeReading(b.c1, "refutes")...)
		readingAlone(t, b, r)
		unread(t, b)
		ev := evidenceJSON(t, b.runDir)
		var row *record.EvidenceVerificationJSON
		for _, s := range ev.Sources {
			for i := range s.Verified {
				if s.Verified[i].Anchor == b.c1 {
					row = &s.Verified[i]
				}
			}
		}
		if row == nil || row.SourceCompleteness != "not_the_work" || row.FindsAgainst() {
			t.Errorf("show evidence's row = %+v, want not_the_work and not finding against", row)
		}
	})
	t.Run("b: its first human text", func(t *testing.T) {
		b := ntw(t)
		out, err := b.verify(lensL, citeReading(b.c1, "refutes")...)
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range []string{"read nothing of its source", lensL + " owes its gap about the anchor"} {
			if !strings.Contains(out, w) {
				t.Errorf("the text does not say %q:\n%s", w, out)
			}
		}
		if !strings.HasPrefix(out, "this reading reads nothing of the source") || strings.Contains(out, "REFUTES") {
			t.Errorf("the text does not open on what the reading read, or says REFUTES:\n%s", out)
		}
	})
	t.Run("c: --anchor absent", func(t *testing.T) {
		b := ntw(t)
		refused(t, b, citeReading(b.c1, "absent"), "not_the_work")
	})
	t.Run("e: --url refutes", func(t *testing.T) {
		b := ntw(t)
		readingAlone(t, b, b.mustRead(lensL, urlReading(rdU3, rdFree1, "refutes")...))
	})
	t.Run("g: --anchor supports", func(t *testing.T) {
		b := ntw(t)
		out, err := b.verify(lensL, citeReading(b.c1, "supports")...)
		if err != nil {
			t.Fatal(err)
		}
		readingAlone(t, b, readingJSON{})
		unread(t, b)
		if !strings.Contains(out, "read nothing of its source") {
			t.Errorf("the text does not say the citation stays unread:\n%s", out)
		}
	})
	t.Run("h: --url supports", func(t *testing.T) {
		b := ntw(t)
		readingAlone(t, b, b.mustRead(lensL, urlReading(rdU3, rdFree1, "supports")...))
		srcs, err := record.CitedSources(runtest.Open(t, b.runDir))
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range srcs {
			if s.URL == rdU3 {
				t.Errorf("the reading placed a cited source: %+v", s)
			}
		}
		for _, s := range evidenceJSON(t, b.runDir).Sources {
			if s.URL == rdU3 {
				t.Errorf("show evidence lists the found source among the report's: %+v", s)
			}
		}
		// The same source for another sentence is another reading.
		b.mustRead(lensL, urlReading(rdU3, rdFree2, "supports")...)
		if c := b.counts(); c != [3]int{2, 0, 0} {
			t.Errorf("verify/mint/anchor %v, want two readings", c)
		}
	})
	t.Run("h: the stamp rows on verify --url", func(t *testing.T) {
		b := readingRun(t)
		b.cacheAs(rdU4, fetchcache.CompletenessAbstract)
		for _, c := range []struct {
			url, sentence string
			want          recordpb.SourceCompleteness
		}{
			{rdU4, rdFree1, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_ABSTRACT},
			{rdU5, rdFree2, recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED},
		} {
			b.mustRead(lensL, urlReading(c.url, c.sentence, "supports")...)
			if v, _ := b.lastVerify(); v.GetSourceCompleteness() != c.want {
				t.Errorf("%s: stamp %v, want %v", c.url, v.GetSourceCompleteness(), c.want)
			}
		}
	})
	t.Run("i: --anchor supports of a full copy and of an uncached source", func(t *testing.T) {
		b := readingRun(t)
		b.cacheAs(rdU2, fetchcache.CompletenessFull)
		b.mustRead(lensL, citeReading(b.c2, "supports")...)
		if v, _ := b.lastVerify(); v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_FULL {
			t.Errorf("a full copy's reading is stamped %v", v.GetSourceCompleteness())
		}
		if err := os.RemoveAll(fetchcache.Dir(runtest.Open(t, b.runDir))); err != nil {
			t.Fatal(err)
		}
		b.mustRead(lensL, citeReading(b.c1, "supports")...)
		if v, _ := b.lastVerify(); v.GetSourceCompleteness() != recordpb.SourceCompleteness_SOURCE_COMPLETENESS_NOT_ASKED {
			t.Errorf("an uncached source's reading is stamped %v", v.GetSourceCompleteness())
		}
	})
}
