package fetchcache

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/hybridsearch"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/nativehelper"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/vecembed"
)

var (
	fakeEmbedOnce sync.Once
	fakeEmbedDir  string
	fakeEmbedErr  error
)

// useFakeEmbedder points vecembed at a fake feov-embed (nativehelper/testdata/fakehelper) built with
// the pins this package carries, so a ranked search runs end to end without the native build. A test
// that needs the embedder and cannot build it fails: a skip would read as a pass.
func useFakeEmbedder(t *testing.T) {
	t.Helper()
	fakeEmbedOnce.Do(func() {
		dir, err := os.MkdirTemp("", "feov-fake-embed-")
		if err != nil {
			fakeEmbedErr = err
			return
		}
		exe := vecembed.HelperName
		if runtime.GOOS == "windows" {
			exe += ".exe"
		}
		ldflags := fmt.Sprintf("-X main.digest=%s -X main.assets=%s", nativehelper.HelperDigest(vecembed.HelperName),
			strings.Join(nativehelper.HelperAssets(vecembed.HelperName), ","))
		cmd := exec.Command("go", "build", "-o", filepath.Join(dir, exe), "-ldflags", ldflags, "../nativehelper/testdata/fakehelper")
		if msg, err := cmd.CombinedOutput(); err != nil {
			fakeEmbedErr = fmt.Errorf("building the fake helper: %v\n%s", err, msg)
			return
		}
		fakeEmbedDir = dir
	})
	if fakeEmbedErr != nil {
		t.Fatal(fakeEmbedErr)
	}
	t.Setenv(nativehelper.EnvDir, fakeEmbedDir)
}

// recordingEmbedder embeds each text as its words hashed into the vector, and records every call.
type recordingEmbedder struct{ calls [][]string }

func (r *recordingEmbedder) Embed(texts []string) ([][]float32, error) {
	r.calls = append(r.calls, slices.Clone(texts))
	out := make([][]float32, len(texts))
	for i, s := range texts {
		v := make([]float32, vecembed.Dim)
		for _, w := range strings.Fields(strings.ToLower(s)) {
			h := sha256.Sum256([]byte(w))
			v[binary.LittleEndian.Uint32(h[:])%vecembed.Dim]++
		}
		var n float64
		for _, x := range v {
			n += float64(x * x)
		}
		for d := range v {
			if n > 0 {
				v[d] /= float32(math.Sqrt(n))
			}
		}
		out[i] = v
	}
	return out, nil
}
func (r *recordingEmbedder) ModelID() string {
	return vecembed.ModelIDFor("rec", runtime.GOOS, runtime.GOARCH)
}
func (r *recordingEmbedder) Close() error { return nil }

// A SEARCH AFTER AN EDIT EMBEDS THE QUERY AND THE PARAGRAPH THE EDIT CHANGED, never the report
// again: each paragraph's vector is cached under its visible text's hash.
func TestReportSearchCachesUnchangedParagraphs(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	paras := make([]string, 20)
	for i := range paras {
		paras[i] = fmt.Sprintf("Paragraph %d says the harbour costs rose by %d percent.<!--cite:C-%08x-->", i, i, i)
	}
	report := func() string { return "# Report\n\n" + strings.Join(paras, "\n\n") + "\n" }
	emb := &recordingEmbedder{}
	search := func() []hybridsearch.ReportChunk {
		t.Helper()
		got, n, err := SearchReport(run, report(), "harbour costs", emb, 10)
		if err != nil || n != 20 || len(got) != 10 {
			t.Fatalf("SearchReport = %d chunks of %d, %v; want the top 10 of 20", len(got), n, err)
		}
		return got
	}

	search()
	if last := emb.calls[len(emb.calls)-1]; len(last) != 21 || last[0] != "harbour costs" {
		t.Fatalf("the cold search embedded %d texts, want the query and all 20 paragraphs", len(last))
	}
	search()
	if last := emb.calls[len(emb.calls)-1]; len(last) != 1 {
		t.Errorf("a search of the unchanged report embedded %d texts, want the query alone: %q", len(last), last)
	}
	paras[7] = "Paragraph 7 now says the dredging was the harbour's largest cost.<!--cite:C-00000007-->"
	got := search()
	last := emb.calls[len(emb.calls)-1]
	if len(last) != 2 || last[1] != "Paragraph 7 now says the dredging was the harbour's largest cost." {
		t.Errorf("the search after one edit embedded %q, want the query and the edited paragraph's visible text", last)
	}
	for _, ch := range got {
		if strings.Contains(ch.Text, "dredging") && ch.Anchor != "C-00000007" {
			t.Errorf("the edited paragraph came back without its anchor: %+v", ch)
		}
	}
	// The file holds the report's paragraphs now, and no other model reads it.
	if cached, err := hybridsearch.ReadReportVecCache(ReportVecCachePath(run), emb.ModelID(), hybridsearch.ReportChunkSpec); err != nil || len(cached) != 20 {
		t.Errorf("report_vecs.bin holds %d paragraphs (%v), want the 20 the report holds", len(cached), err)
	}
}

// PREWARM WRITES THE CACHE THE SEARCH READS. A fetch's extraction and an OCR reading each embed the
// text at the address the search computes — the sha256 of the bytes SourceTextPath holds — so the
// search finds the vectors and writes none of its own.
func TestPrewarmMatchesFullSourceTextSha(t *testing.T) {
	useFakeEmbedder(t)
	vecFiles := func(run record.Run) []string {
		m, _ := filepath.Glob(filepath.Join(Dir(run), "*.vec"))
		return m
	}
	text := "Costs in the region are moderate and rising. The harbour was dredged twice in the decade."
	for _, origin := range []string{"embedded", "ocr_derived", "ocr_derived_at_fetch"} {
		t.Run(origin, func(t *testing.T) {
			run := runtest.New(t, t.TempDir())
			var e Entry
			switch origin {
			case "embedded":
				yes := true
				var err error
				if e, err = Store(run, Entry{URL: "https://ex/a.html", ContentType: "text/html", TextExtracted: &yes, Extractor: htmlExtractorID}, []byte("<html>"+text+"</html>")); err != nil {
					t.Fatal(err)
				}
				if err := foldExtraction(run, &e, Extraction{Attempted: true, Text: text, ExtractorID: htmlExtractorID}); err != nil {
					t.Fatal(err)
				}
			case "ocr_derived_at_fetch":
				// fetch's automatic path: render and read a stored scan in one call.
				run2, scan := storeScanned(t, 1)
				run, e = run2, scan
				withEngine(t, textEngine(func(int) (string, error) { return text, nil }))
				if _, err := (RenderAndRead{}).ReadScanned(context.Background(), run, e); err != nil {
					t.Fatal(err)
				}
			default:
				run2, scan := storeScanned(t, 2)
				run, e = run2, scan
				rd := fakeRender(t, run, e.Sha, [][]byte{[]byte("p1"), []byte("p2")}, tessocr.RenderDPI)
				pages := []string{"Costs in the region are moderate and rising.", "The harbour was dredged twice in the decade."}
				withEngine(t, textEngine(func(n int) (string, error) { return pages[n-1], nil }))
				if _, err := ReadRenderedPages(run, e.Sha, rd); err != nil {
					t.Fatal(err)
				}
			}
			o, _, err := TextOrigin(run, e)
			if err != nil {
				t.Fatal(err)
			}
			b, err := os.ReadFile(SourceTextPath(run, e, o))
			if err != nil {
				t.Fatal(err)
			}
			want := VectorCachePath(run, Sha(b))
			if got := vecFiles(run); !slices.Equal(got, []string{want}) {
				t.Fatalf("after the %s text was written the cache holds %v, want %s — the sha256 of %s", origin, got, filepath.Base(want), SourceTextPath(run, e, o))
			}
			before, _ := os.ReadFile(want)
			if _, err := SearchSourceHybrid(run, e, "harbour dredged", 3); err != nil {
				t.Fatal(err)
			}
			after, _ := os.ReadFile(want)
			if got := vecFiles(run); !slices.Equal(got, []string{want}) || string(before) != string(after) {
				t.Errorf("the search wrote %v; want it to read the prewarmed %s and write nothing", got, filepath.Base(want))
			}
		})
	}
}

// A RANKED HIT OFFERS WHAT CITE TAKES. The chunk is the passage, and its best sentence — the one
// holding the query's words, on the page it actually sits on — is the span `cite --source-span`
// accepts, even when the chunk starts on the page before.
func TestHybridSourceHitSelectsCiteableSentenceAndPage(t *testing.T) {
	useFakeEmbedder(t)
	filler := strings.Repeat("The committee met in the spring and adjourned without a vote. ", 6)
	pages := []string{
		strings.TrimSpace(filler) + "\nThe minutes were filed late.",
		"Ledgers were kept by the clerk. The harbour dredging cost rose by forty percent in the decade. Nothing else changed.",
	}
	run, e := storeScanned(t, 2)
	rd := fakeRender(t, run, e.Sha, [][]byte{[]byte("p1"), []byte("p2")}, tessocr.RenderDPI)
	withEngine(t, textEngine(func(n int) (string, error) { return pages[n-1], nil }))
	if _, err := ReadRenderedPages(run, e.Sha, rd); err != nil {
		t.Fatal(err)
	}
	text := storeTextSource(t, run, "https://ex/notes", "Notes on the port.\nThe harbour dredging cost rose sharply after the survey.\nNo other works were done.")

	const query = "harbour dredging cost"
	const want2 = "The harbour dredging cost rose by forty percent in the decade."
	single, err := SearchSourceHybrid(run, e, query, 3)
	if err != nil {
		t.Fatal(err)
	}
	cross, err := SearchCrossSourcesHybrid(run, query, 10)
	if err != nil {
		t.Fatal(err)
	}
	var scanHit *CrossSourceHit
	for i := range cross.Hits {
		if cross.Hits[i].Sha == e.Sha {
			scanHit = &cross.Hits[i].CrossSourceHit
		}
	}
	if len(single.Hits) != 1 || scanHit == nil {
		t.Fatalf("one chunk of the scan should rank: single %d hits, cross %+v", len(single.Hits), cross.Hits)
	}
	for name, h := range map[string]SourceHit{"single": single.Hits[0].SourceHit, "cross": {Page: scanHit.Page, Start: scanHit.Start, End: scanHit.End,
		TextSha: scanHit.TextSha, Matched: scanHit.Matched, Line: scanHit.Line, LastLine: scanHit.LastLine, Sentence: scanHit.Sentence,
		SentenceStart: scanHit.SentenceStart, SentenceEnd: scanHit.SentenceEnd, SourceSpan: scanHit.SourceSpan, NotCiteable: scanHit.NotCiteable}} {
		if h.Page != 2 || h.TextSha != Sha([]byte(pages[1])) || h.SourceSpan != want2 || h.Sentence != want2 ||
			pages[1][h.SentenceStart:h.SentenceEnd] != want2 {
			t.Errorf("%s: the hit offers page %d %q (span %q); want page 2's sentence %q", name, h.Page, h.Sentence, h.SourceSpan, want2)
		}
		// The chunk starts on page 1; the hit states its part on page 2, in page 2's text.
		if h.Start != 0 || h.End != len(pages[1]) || h.Matched != pages[1] || h.Line != 1 || h.LastLine != 1 {
			t.Errorf("%s: the chunk is stated as [%d,%d) lines %d-%d %q; want all of page 2", name, h.Start, h.End, h.Line, h.LastLine, h.Matched)
		}
		loc, err := LocateSourceSpan(run, e, h.SourceSpan)
		if err != nil || !slices.Equal(loc.Pages, []int{2}) {
			t.Errorf("%s: cite's locator on the offered span = %+v, %v; want it placed on page 2", name, loc.Pages, err)
		}
	}

	// An embedded text: the chunk is the whole text, over its three lines; the sentence is the second.
	th, err := SearchSourceHybrid(run, text, query, 3)
	if err != nil || len(th.Hits) != 1 {
		t.Fatalf("the text source: %+v, %v", th.Hits, err)
	}
	h := th.Hits[0]
	full := "Notes on the port.\nThe harbour dredging cost rose sharply after the survey.\nNo other works were done."
	if h.Matched != full || h.Start != 0 || h.End != len(full) || h.Line != 1 || h.LastLine != 3 {
		t.Errorf("the chunk is stated as [%d,%d) lines %d-%d; want the whole text, lines 1-3", h.Start, h.End, h.Line, h.LastLine)
	}
	if want := "The harbour dredging cost rose sharply after the survey."; h.SourceSpan != want || h.Page != 0 {
		t.Errorf("the hit offers %q on page %d; want %q", h.SourceSpan, h.Page, want)
	}
	if _, err := LocateSourceSpan(run, text, h.SourceSpan); err != nil {
		t.Errorf("cite's locator refuses the offered span: %v", err)
	}
}

// A SCAN WITH NO READING IS NAMED, NOT DROPPED. A ranked search across sources ranks the sources that
// hold text and names the image-only PDF as unsearched, saying what it needs.
func TestACrossSourceSearchNamesAnImageOnlyScanUnsearched(t *testing.T) {
	useFakeEmbedder(t)
	run, scan := storeScanned(t, 1)
	storeTextSource(t, run, "https://ex/notes", "The harbour dredging cost rose sharply after the survey.")
	for name, scope := range map[string]func() ([]hybridsearch.UnsearchedSource, int){
		"ranked": func() ([]hybridsearch.UnsearchedSource, int) {
			res, err := SearchCrossSourcesHybrid(run, "harbour cost", 5)
			if err != nil {
				t.Fatal(err)
			}
			return res.SearchScope.UnsearchedSources, len(res.Hits)
		},
		"exhaustive": func() ([]hybridsearch.UnsearchedSource, int) {
			res, err := SearchCrossSources(run, "", "harbour")
			if err != nil {
				t.Fatal(err)
			}
			return res.Selection.UnsearchedSources, len(res.Hits)
		},
	} {
		un, hits := scope()
		want := hybridsearch.UnsearchedSource{Sha: scan.Sha, URL: scan.URL, Reason: "image_only: no OCR reading yet; run fetch --url " + scan.URL}
		if !slices.Equal(un, []hybridsearch.UnsearchedSource{want}) || hits != 1 {
			t.Errorf("%s: unsearched %+v and %d hits; want the scan named image_only and the text's hit", name, un, hits)
		}
	}
}
