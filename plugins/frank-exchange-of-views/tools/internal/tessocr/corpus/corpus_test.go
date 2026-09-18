package corpus

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func good() Provenance {
	yes := true
	return Provenance{
		Slug: "a-page", SourceURL: "https://example.invalid/x.pdf", SourcePage: 3,
		Publisher: "Somebody", Date: "1981-04",
		Rights: "us-government-work", RightsNote: "17 U.S.C. 105",
		LocalFileSha256: strings.Repeat("a", 64), PageSha256: strings.Repeat("b", 64),
		DefectClass: []string{"dashed-rules"}, Why: "it beat the reader",
		Expect: Expect{Table: &yes},
	}
}

// The licence is the one field whose failure is not ours to absorb: a page we may not redistribute
// is in a public repository the moment it is committed.
func TestARightsValueOutsideTheClosedSetIsRefused(t *testing.T) {
	for _, bad := range []string{"", "probably public domain", "public domain", "MIT", "us government work"} {
		p := good()
		p.Rights = bad
		err := p.Validate()
		if err == nil {
			t.Errorf("rights %q was accepted; the set is %s", bad, strings.Join(Rights, ", "))
			continue
		}
		if !strings.Contains(err.Error(), "rights") {
			t.Errorf("rights %q: the refusal does not name the field: %v", bad, err)
		}
	}
	for _, ok := range Rights {
		p := good()
		p.Rights = ok
		if err := p.Validate(); err != nil {
			t.Errorf("rights %q is in the set and was refused: %v", ok, err)
		}
	}
}

// Without an expect block the corpus pins a reading and says nothing about which way is better,
// which is the state the hand-set status of the first draft would have left it in permanently.
func TestARecordWithNoExpectIsRefused(t *testing.T) {
	p := good()
	p.Expect = Expect{}
	err := p.Validate()
	if err == nil {
		t.Fatal("a record with an empty expect was accepted")
	}
	if !strings.Contains(err.Error(), "expect") {
		t.Fatalf("the refusal does not name expect: %v", err)
	}
}

func TestEveryRequiredFieldIsNamedWhenItIsMissing(t *testing.T) {
	for name, break_ := range map[string]func(*Provenance){
		"slug":              func(p *Provenance) { p.Slug = "" },
		"source_url":        func(p *Provenance) { p.SourceURL = "" },
		"source_page":       func(p *Provenance) { p.SourcePage = 0 },
		"publisher":         func(p *Provenance) { p.Publisher = "" },
		"rights_evidence":   func(p *Provenance) { p.RightsNote = " " },
		"page_sha256":       func(p *Provenance) { p.PageSha256 = "short" },
		"local_file_sha256": func(p *Provenance) { p.LocalFileSha256 = "" },
		"defect_class":      func(p *Provenance) { p.DefectClass = nil },
		"why":               func(p *Provenance) { p.Why = "" },
	} {
		p := good()
		break_(&p)
		err := p.Validate()
		if err == nil {
			t.Errorf("%s: missing, and the record was accepted", name)
			continue
		}
		if !strings.Contains(err.Error(), name) {
			t.Errorf("%s: the refusal does not name it: %v", name, err)
		}
	}
}

func TestCheckReportsEveryClaimThatFailedAndNothingElse(t *testing.T) {
	yes, three := true, 3
	e := Expect{Table: &yes, Rows: &three, MustContain: []string{"Spot Welding", "5.4"}}

	failed := e.Check(Observed{Table: true, Rows: 3, Text: "Spot Welding 5.4"})
	if len(failed) != 0 {
		t.Fatalf("a reading that meets every claim reported %v", failed)
	}

	failed = e.Check(Observed{Table: false, Rows: 2, Text: "Spot Welding 3.4"})
	if len(failed) != 3 {
		t.Fatalf("want three failures (table, rows, the missing string), got %v", failed)
	}
	joined := strings.Join(failed, " | ")
	for _, want := range []string{"table", "rows", "5.4"} {
		if !strings.Contains(joined, want) {
			t.Errorf("the report does not name %q: %s", want, joined)
		}
	}
}

// An empty corpus is the honest zero this package exists to refuse: every gate below walks the set
// Load returns, so a set of none passes them all.
func TestAnEmptyCorpusIsAnError(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err == nil {
		t.Fatal("an empty corpus loaded without complaint")
	}
}

func TestATreeWithAnOrphanOrAMissingFileIsRefused(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{FileStatus, FileRefs, FileRefsDoc, FileCorpusMD} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	slug := filepath.Join(root, "a-page")
	if err := os.MkdirAll(slug, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{FilePage, FileRecord, FileREADME, FileGolden} {
		if err := os.WriteFile(filepath.Join(slug, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := CheckTree(root); err != nil {
		t.Fatalf("a complete tree was refused: %v", err)
	}

	stray := filepath.Join(slug, "crop.png")
	if err := os.WriteFile(stray, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CheckTree(root)
	if err == nil || !strings.Contains(err.Error(), "crop.png") {
		t.Fatalf("an orphan beside the four files was not named: %v", err)
	}
	if err := os.Remove(stray); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(slug, FileGolden)); err != nil {
		t.Fatal(err)
	}
	err = CheckTree(root)
	if err == nil || !strings.Contains(err.Error(), FileGolden) {
		t.Fatalf("a missing golden was not named: %v", err)
	}
}

func TestStatusCountsOnlyThePagesThatFailTheirExpect(t *testing.T) {
	s := Status([]Result{
		{Slug: "b-page", Failed: []string{"table: want true, got false"}},
		{Slug: "a-page"},
	})
	if !strings.Contains(s, "**1 of 2 pages do not yet read as they should.**") {
		t.Errorf("the meter does not count what the results say:\n%s", s)
	}
	// Slug order, not result order: the file is read by humans across rounds.
	if strings.Index(s, "a-page") > strings.Index(s, "b-page") {
		t.Error("STATUS.md is not in slug order")
	}
	if got := SlugsIn(s); len(got) != 2 || got[0] != "a-page" || got[1] != "b-page" {
		t.Errorf("SlugsIn read %v out of the table it renders", got)
	}
}
