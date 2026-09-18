package tessocr_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/corpus"
)

// THE CORPUS'S RECORDS, CHECKED WITHOUT THE ENGINE.
//
// These run in the default build, on every box, because what they refuse is not about pixels: a
// page we may not redistribute, a record nobody filled in, a file no gate reads, and prose that has
// drifted from the record it was generated from.
//
// The reading itself is pinned by TestCorpusGoldens in internal/fetchcache, which needs the engine
// and runs only on the tagged Linux leg.

const corpusRoot = "testdata/corpus"

func TestCorpusTreeHasNothingUnaccountedFor(t *testing.T) {
	if err := corpus.CheckTree(corpusRoot); err != nil {
		t.Fatal(err)
	}
}

func TestEveryCorpusRecordCanBeActedOn(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if err := c.Prov.Validate(); err != nil {
			t.Errorf("%v", err)
		}
	}
}

// The page the harness serves must be the page the record names, or the golden was pinned against
// bytes nobody can identify afterwards.
func TestEveryCorpusPageIsTheOneItsRecordNames(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		gz, rerr := os.ReadFile(filepath.Join(c.Dir, corpus.FilePage))
		if rerr != nil {
			t.Fatalf("%s: %v", c.Slug, rerr)
		}
		page, uerr := corpus.Decompress(gz)
		if uerr != nil {
			t.Fatalf("%s: %v", c.Slug, uerr)
		}
		if got := corpus.Sha(page); got != c.Prov.PageSha256 {
			t.Errorf("%s: page.pdf.gz decompresses to %s, record says %s", c.Slug, got, c.Prov.PageSha256)
		}
	}
}

func TestEveryCorpusREADMEIsTheGeneratedOne(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		path := filepath.Join(c.Dir, corpus.FileREADME)
		want := corpus.README(c.Prov)
		have, rerr := os.ReadFile(path)
		if rerr != nil {
			if !os.IsNotExist(rerr) || os.Getenv("UPDATE_GOLDENS") != "1" {
				t.Fatalf("%s: %v", c.Slug, rerr)
			}
			have = nil
		}
		if string(have) != want {
			if os.Getenv("UPDATE_GOLDENS") == "1" {
				if werr := os.WriteFile(path, []byte(want), 0o644); werr != nil {
					t.Fatal(werr)
				}
				t.Logf("%s: README regenerated", c.Slug)
				continue
			}
			t.Errorf("%s: README.md has drifted from provenance.json. It is GENERATED — edit the "+
				"record, then rerun with UPDATE_GOLDENS=1", c.Slug)
		}
	}
}

func TestREFERENCESIsTheGeneratedOne(t *testing.T) {
	refs, err := corpus.LoadReferences(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(refs) == 0 {
		t.Fatal("references.json is empty — the sources we may NOT commit are a record too, and an " +
			"empty one reads like a search nobody has run")
	}
	path := filepath.Join(corpusRoot, corpus.FileRefsDoc)
	want := corpus.References(refs)
	have, rerr := os.ReadFile(path)
	if rerr != nil {
		if !os.IsNotExist(rerr) || os.Getenv("UPDATE_GOLDENS") != "1" {
			t.Fatal(rerr)
		}
		have = nil
	}
	if string(have) != want {
		if os.Getenv("UPDATE_GOLDENS") == "1" {
			if werr := os.WriteFile(path, []byte(want), 0o644); werr != nil {
				t.Fatal(werr)
			}
			return
		}
		t.Error("REFERENCES.md has drifted from references.json — it is GENERATED")
	}
}

// A source cannot be both committed and refused: one of the two records is then wrong, and which
// one is not recoverable afterwards.
func TestNoSourceIsBothCommittedAndRefused(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := corpus.LoadReferences(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	committed := map[string]bool{}
	for _, c := range cases {
		committed[c.Slug] = true
	}
	for _, r := range refs {
		if committed[r.Name] {
			t.Errorf("%q is in references.json as NOT committable and is also a corpus page", r.Name)
		}
	}
}

// STATUS.md is written by the tagged harness, which is the only thing that has read the pages. This
// gate can compare the SET of pages it names with the corpus, and cannot re-establish the verdicts —
// said plainly here so the weaker check is not read as the stronger one.
func TestStatusNamesExactlyTheCorpusPages(t *testing.T) {
	cases, err := corpus.Load(corpusRoot)
	if err != nil {
		t.Fatal(err)
	}
	status, rerr := os.ReadFile(filepath.Join(corpusRoot, corpus.FileStatus))
	if rerr != nil {
		t.Fatal(rerr)
	}
	var want []string
	for _, c := range cases {
		want = append(want, c.Slug)
	}
	if got := corpus.SlugsIn(string(status)); !reflect.DeepEqual(got, want) {
		t.Errorf("STATUS.md names %v, the corpus holds %v — regenerate it with the tagged harness", got, want)
	}
}
