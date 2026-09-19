// Package corpus holds the records behind the page corpus — the scans that BEAT the reader, kept
// as fixtures so a change to the reader arrives as a diff a human reads.
//
// The pages are committed, unlike the generated layout cases beside them, so two facts have to be
// refusable rather than hoped for: that we may redistribute a page (Rights, a closed set), and what
// a CORRECT reader would do with it (Expect, checked against the reading). A corpus whose licence
// lives in prose and whose verdict lives in a hand-set flag reports a clean board in exactly the
// state where it has stopped measuring anything.
package corpus

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The four files every case directory holds, and nothing else. An orphan beside them is a fixture
// nothing reads or a record nothing pins, and CheckTree refuses both.
const (
	FilePage     = "page.pdf.gz"
	FileRecord   = "provenance.json"
	FileREADME   = "README.md"
	FileGolden   = "reading.golden"
	FileStatus   = "STATUS.md"
	FileRefs     = "references.json"
	FileRefsDoc  = "REFERENCES.md"
	FileCorpusMD = "README.md"
)

// Rights is the CLOSED set of licences a committed page may carry. A value outside it is refused
// at the write, not described in prose: "probably public domain" is the sentence that puts bytes
// in a public repository on a guess.
var Rights = []string{
	"us-government-work",
	"public-domain-expired",
	"cc0-1.0",
	"cdla-permissive-1.0",
	"cdla-permissive-2.0",
	"apache-2.0",
}

// Provenance is one page's record: where it came from, what we may do with it, and what a correct
// reader would produce.
type Provenance struct {
	Slug         string `json:"slug"`
	SourceURL    string `json:"source_url"`
	SourcePage   int    `json:"source_page"`
	Publisher    string `json:"publisher"`
	Date         string `json:"date"`
	Rights       string `json:"rights"`
	RightsNote   string `json:"rights_evidence"`
	RightsCaveat string `json:"rights_caveat,omitempty"`
	// LocalFileSha256 is the sha of the file the operator downloaded, and the name says exactly
	// that. Nothing here fetches SourceURL, so nothing here can attest the bytes came from it.
	LocalFileSha256 string `json:"local_file_sha256"`
	// PageSha256 is the sha of the DECOMPRESSED page — the bytes the reader is handed and the
	// golden was pinned against. The gzip is a container; the fact is the page.
	PageSha256        string   `json:"page_sha256"`
	TextObjectsRemove int      `json:"text_objects_removed"`
	AddedAt           string   `json:"added_at"`
	DefectClass       []string `json:"defect_class"`
	Why               string   `json:"why"`
	Expect            Expect   `json:"expect"`
}

// Expect is what a correct reader would produce, transcribed from the PAGE IMAGE by a human. It is
// the direction the golden cannot supply: a golden says what changed, this says which way is right.
//
// The pointer fields are claims that may be absent; MustContain holds strings printed on the page.
type Expect struct {
	Table       *bool    `json:"table,omitempty"`
	Tables      *int     `json:"tables,omitempty"`
	Rows        *int     `json:"rows,omitempty"`
	Columns     *int     `json:"columns,omitempty"`
	MustContain []string `json:"must_contain,omitempty"`
}

// Observed is what the reader actually produced for a page. The harness fills it; this package
// stays free of the engine and of fetchcache.
type Observed struct {
	Table   bool
	Tables  int
	Rows    int
	Columns int
	Text    string
}

// Case is one directory of the corpus.
type Case struct {
	Slug string
	Dir  string
	Prov Provenance
}

// Validate refuses a record that cannot be acted on. Every message names the field.
func (p Provenance) Validate() error {
	var bad []string
	if p.Slug == "" {
		bad = append(bad, "slug is empty")
	}
	if p.SourceURL == "" {
		bad = append(bad, "source_url is empty")
	}
	if p.SourcePage < 1 {
		bad = append(bad, "source_page must be 1 or more")
	}
	if p.Publisher == "" {
		bad = append(bad, "publisher is empty")
	}
	if !knownRight(p.Rights) {
		bad = append(bad, fmt.Sprintf("rights %q is not one of %s", p.Rights, strings.Join(Rights, ", ")))
	}
	if strings.TrimSpace(p.RightsNote) == "" {
		bad = append(bad, "rights_evidence is empty — a URL or a quoted licence line")
	}
	if len(p.PageSha256) != 64 {
		bad = append(bad, "page_sha256 is not a sha256")
	}
	if len(p.LocalFileSha256) != 64 {
		bad = append(bad, "local_file_sha256 is not a sha256")
	}
	if len(p.DefectClass) == 0 {
		bad = append(bad, "defect_class is empty — a page enters only as a specimen of something")
	}
	if strings.TrimSpace(p.Why) == "" {
		bad = append(bad, "why is empty")
	}
	if p.Expect.empty() {
		bad = append(bad, "expect is empty — without it nothing says which way a change is an improvement")
	}
	if len(bad) > 0 {
		return fmt.Errorf("%s: %s", p.Slug, strings.Join(bad, "; "))
	}
	return nil
}

func knownRight(s string) bool {
	for _, r := range Rights {
		if r == s {
			return true
		}
	}
	return false
}

func (e Expect) empty() bool {
	return e.Table == nil && e.Tables == nil && e.Rows == nil && e.Columns == nil && len(e.MustContain) == 0
}

// Check returns the claims the reading FAILED, in the order they are written. An empty return is a
// page the reader gets right.
func (e Expect) Check(o Observed) []string {
	var failed []string
	if e.Table != nil && *e.Table != o.Table {
		failed = append(failed, fmt.Sprintf("table: want %v, got %v", *e.Table, o.Table))
	}
	if e.Tables != nil && *e.Tables != o.Tables {
		failed = append(failed, fmt.Sprintf("tables: want %d, got %d", *e.Tables, o.Tables))
	}
	if e.Rows != nil && *e.Rows != o.Rows {
		failed = append(failed, fmt.Sprintf("rows: want %d, got %d", *e.Rows, o.Rows))
	}
	if e.Columns != nil && *e.Columns != o.Columns {
		failed = append(failed, fmt.Sprintf("columns: want %d, got %d", *e.Columns, o.Columns))
	}
	for _, want := range e.MustContain {
		if !strings.Contains(o.Text, want) {
			failed = append(failed, fmt.Sprintf("must_contain: %q is printed on the page and is not in the reading", want))
		}
	}
	return failed
}

// Load reads every case under root, in slug order. It returns an error rather than an empty slice
// when root holds nothing: a corpus that matches nothing reads exactly like a corpus that passes.
func Load(root string) ([]Case, error) {
	ents, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var cases []Case
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		b, rerr := os.ReadFile(filepath.Join(dir, FileRecord))
		if rerr != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), rerr)
		}
		var p Provenance
		if jerr := json.Unmarshal(b, &p); jerr != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), jerr)
		}
		if p.Slug != e.Name() {
			return nil, fmt.Errorf("%s: record names slug %q — the directory name IS the slug", e.Name(), p.Slug)
		}
		cases = append(cases, Case{Slug: p.Slug, Dir: dir, Prov: p})
	}
	if len(cases) == 0 {
		return nil, fmt.Errorf("%s holds no cases — an empty corpus matches forever and pins nothing", root)
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].Slug < cases[j].Slug })
	return cases, nil
}

// CheckTree refuses anything under root that no gate accounts for: a case directory missing one of
// its four files, and any other file or directory at either level.
func CheckTree(root string) error {
	want := map[string]bool{FilePage: true, FileRecord: true, FileREADME: true, FileGolden: true}
	top := map[string]bool{FileStatus: true, FileRefs: true, FileRefsDoc: true, FileCorpusMD: true}
	ents, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	var bad []string
	for _, e := range ents {
		if !e.IsDir() {
			if !top[e.Name()] {
				bad = append(bad, fmt.Sprintf("%s is not a file this corpus accounts for", e.Name()))
			}
			continue
		}
		seen := map[string]bool{}
		inner, ierr := os.ReadDir(filepath.Join(root, e.Name()))
		if ierr != nil {
			return ierr
		}
		for _, f := range inner {
			if !want[f.Name()] {
				bad = append(bad, fmt.Sprintf("%s/%s is an orphan — no gate reads it", e.Name(), f.Name()))
				continue
			}
			seen[f.Name()] = true
		}
		for name := range want {
			if !seen[name] {
				bad = append(bad, fmt.Sprintf("%s/%s is missing", e.Name(), name))
			}
		}
	}
	for _, name := range []string{FileStatus, FileRefs, FileRefsDoc, FileCorpusMD} {
		if _, err := os.Stat(filepath.Join(root, name)); err != nil {
			bad = append(bad, fmt.Sprintf("%s is missing", name))
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return fmt.Errorf("corpus tree:\n  %s", strings.Join(bad, "\n  "))
	}
	return nil
}

// Reference is a source we WANT and may not commit, or could not reach. It is a record for the
// same reason the pages are: the next round otherwise re-finds it and guesses again.
type Reference struct {
	Name          string `json:"name"`
	URL           string `json:"url"`
	WhyWanted     string `json:"why_wanted"`
	Rights        string `json:"rights"`
	BlockedReason string `json:"blocked_reason"`
}

// LoadReferences reads references.json from the corpus root.
func LoadReferences(root string) ([]Reference, error) {
	b, err := os.ReadFile(filepath.Join(root, FileRefs))
	if err != nil {
		return nil, err
	}
	var refs []Reference
	if err := json.Unmarshal(b, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

// Write stores a record, formatted the one way, so a hand edit and a tool write cannot differ in
// whitespace alone.
func Write(dir string, p Provenance) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, FileRecord), append(b, '\n'), 0o644)
}

// Decompress returns the page a stored .gz holds.
func Decompress(gz []byte) ([]byte, error) {
	zr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	return io.ReadAll(zr)
}

// Sha is the page identity used in the records and in the harness.
func Sha(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
