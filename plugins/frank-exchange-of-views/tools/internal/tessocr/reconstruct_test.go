package tessocr

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// The reconstruction oracle (plan §V.2), as checked-in fixtures so CI enforces it with no
// OCR run: p0054's real 300-DPI TSV and rotated-band header recovery in, the full
// |-separated Table 3 out.
//
// THE EXPECTATION IS PINNED AT THE CORRECTED ORACLE, NOT THE RAW TRANSCRIPT. The trusted
// model transcript carries one documented fabricated cell (Inspection/Test case x
// Requirements: the page has no X there; the model invented one). The Wave 0 prototype
// measured 380/385 cells against that transcript and 381/385 against the pixels — the
// delta IS the fabricated cell — so the golden here is the reconstruction that matches
// the pixels. The four remaining known misses are glyphs tesseract never emitted at any
// confidence, unrecoverable from word TSV by construction.
func TestReconstructP0054(t *testing.T) {
	tsv := readFixture(t, "p0054.300.tsv")
	headers := ParseRotatedBandHeaders(readFixture(t, "p0054.headers.300.txt"))
	if len(headers) != 11 {
		t.Fatalf("rotated-band fixture parsed to %d headers, want 11: %q", len(headers), headers)
	}

	table, st, err := Reconstruct(tsv, headers)
	if err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}
	if want := readFixture(t, "p0054.expected.md"); table != want {
		t.Errorf("reconstructed table differs from the pinned expectation:\n%s",
			cmp.Diff(want, table))
	}

	// The Wave 0 measured stats, exactly. These are what the page record will carry;
	// a drift here is a behaviour change in the reconstruction even if the table
	// happens to survive it.
	want := Stats{
		ColumnsFound:     11,
		SubColumnsFound:  11,
		RowsFound:        35,
		HeaderNamesFound: 11,
		MarksTotal:       191,
		MarksPlaced:      189,
		MarksUnplaced:    2,
		LabellessRows:    0,
		EmptyRows:        1,
	}
	if st != want {
		t.Errorf("stats = %+v, want %+v", st, want)
	}
}

func TestReconstructNoMarks(t *testing.T) {
	// A prose-only TSV must refuse loudly, never emit a plausible empty table.
	tsv := "5\t1\t1\t1\t1\t1\t100\t100\t50\t20\t96.0\thello\n" +
		"5\t1\t1\t1\t1\t2\t160\t100\t50\t20\t96.0\tworld\n"
	if _, _, err := Reconstruct(tsv, nil); !errors.Is(err, ErrNoMarks) {
		t.Fatalf("Reconstruct on prose = %v, want ErrNoMarks", err)
	}
}

func TestParseRotatedBandHeaders(t *testing.T) {
	// A multi-line group is ONE wrapped name — that is how "Installation and\nCheckout"
	// stays a single column.
	got := ParseRotatedBandHeaders("Acquisition\n\nInstallation and\nCheckout\n\nOperation\n")
	want := []string{"Acquisition", "Installation and Checkout", "Operation"}
	if !cmp.Equal(got, want) {
		t.Errorf("headers = %q, want %q", got, want)
	}
}

func TestMarkTokenClassification(t *testing.T) {
	cases := []struct {
		token  string
		strong bool
	}{
		{"X", true},
		{"xX", true},
		{"D4", false},                    // no x/X after stripping — weak at best
		{"X|xX", true},                   // separators stripped
		{"X|XPX/X|X|X{X|X/X|KEX|", true}, // a whole column-run merged into one token
		{"hello", false},
		{"4", false}, // weak, not strong: no x/X
		{"|", false}, // pure separator strips to nothing
	}
	for _, c := range cases {
		if got := isStrongMark(c.token); got != c.strong {
			t.Errorf("isStrongMark(%q) = %v, want %v", c.token, got, c.strong)
		}
	}
	if !isWeakMark("4") || !isWeakMark("K|") || isWeakMark("hello") {
		t.Errorf("weak-mark grammar: want 4 and K| weak, hello not")
	}
}

func TestMarkTokenCount(t *testing.T) {
	// Geometry fields are irrelevant to counting; only level-5 rows with mark-shaped
	// text count. The count feeds PSMDisagreement, so a miscount here is a silent
	// dropout signal miscalibrated.
	tsv := "5\t1\t1\t1\t1\t1\t10\t10\t20\t20\t90\tX\n" +
		"5\t1\t1\t1\t1\t2\t40\t10\t20\t20\t90\tword\n" +
		"4\t1\t1\t1\t1\t0\t0\t0\t0\t0\t-1\tX\n" + // level-4 line row: not a word
		"5\t1\t1\t1\t1\t3\t70\t10\t20\t20\t90\txX\n"
	if got := MarkTokenCount(tsv); got != 2 {
		t.Errorf("MarkTokenCount = %d, want 2", got)
	}
}

func TestPSMDisagreement(t *testing.T) {
	cases := []struct {
		auto, sparse int
		want         float64
	}{
		{0, 0, 0},   // prose page: no marks under either mode is the honest zero
		{26, 26, 0}, // full agreement
		{0, 26, 1},  // p0052: total silent dropout under PSMAuto
		{26, 0, 1},  // symmetric — the signal does not care which mode dropped
		{13, 26, 0.5},
	}
	for _, c := range cases {
		if got := PSMDisagreement(c.auto, c.sparse); got != c.want {
			t.Errorf("PSMDisagreement(%d, %d) = %v, want %v", c.auto, c.sparse, got, c.want)
		}
	}
}

func TestExpectedIntersections(t *testing.T) {
	// The OCR-independent denominator: the rule lattice implied by the reconstruction,
	// to compare against the detector's measured intersection count. p0054's shape:
	// 35 rows x 11 sub-columns implies a 36x12 lattice.
	s := Stats{RowsFound: 35, SubColumnsFound: 11}
	if got := s.ExpectedIntersections(); got != 432 {
		t.Errorf("ExpectedIntersections = %d, want 432", got)
	}
}

func readFixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// The engine identity is the record's re-derivation key; its pin constants must agree
// with the manifest the C stack is actually built from. This is the in-repo half of the
// staleness gate (build-cstack.sh is the build-time half): neither carrier can move
// without the other noticing.
func TestIdentityMatchesPins(t *testing.T) {
	if got, want := Identity(), "tesseract@5.5.3+leptonica@1.87.0"; got != want {
		t.Fatalf("Identity() = %q, want %q", got, want)
	}
	pins, err := os.ReadFile(filepath.Join("..", "..", "third_party", "pins", "PINS.txt"))
	if err != nil {
		t.Fatalf("reading PINS.txt: %v", err)
	}
	for _, tarball := range []string{
		"tesseract-" + tesseractPin + ".tar.gz",
		"leptonica-" + leptonicaPin + ".tar.gz",
	} {
		if !strings.Contains(string(pins), tarball) {
			t.Errorf("PINS.txt does not pin %s — the Go pin constants and the download "+
				"manifest have drifted", tarball)
		}
	}
}
