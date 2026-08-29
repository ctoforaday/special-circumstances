package fetchcache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// THE FIXTURES ARE BUILT, NOT PASTED. A base64 blob would say "a PDF" and nothing more; these
// say exactly what a text layer IS (a content stream carrying a `Tj` operator) and what its
// absence IS (a page whose only content is an image XObject). That distinction is the entire
// subject of this file, so it should be legible rather than encoded.
//
// buildPDF assembles a minimal PDF and computes the cross-reference offsets from the bytes it
// actually wrote. Hand-written offsets in a pasted literal go stale the moment anyone edits a
// string, and a PDF with a broken xref fails for a reason that has nothing to do with the test.
func buildPDF(objects []string, infoObj bool) []byte {
	var out strings.Builder
	out.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects))
	for i, o := range objects {
		offsets = append(offsets, out.Len())
		fmt.Fprintf(&out, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := out.Len()
	fmt.Fprintf(&out, "xref\n0 %d\n0000000000 65535 f \n", len(objects)+1)
	for _, off := range offsets {
		fmt.Fprintf(&out, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&out, "trailer\n<< /Size %d /Root 1 0 R", len(objects)+1)
	if infoObj {
		fmt.Fprintf(&out, " /Info %d 0 R", len(objects))
	}
	fmt.Fprintf(&out, " >>\nstartxref\n%d\n%%%%EOF\n", xref)
	return []byte(out.String())
}

// stream wraps a content stream with the /Length the parser needs.
func stream(body string) string {
	return fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(body), body)
}

const fixtureSentence = "Hello leaf verification"

// pdfWithTextLayer draws one sentence with a real font — the ordinary case, and the one every
// paper in the 2026-08-23 bibliography except IEEE 1012 falls into.
func pdfWithTextLayer() []byte {
	return buildPDF([]string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R " +
			"/Resources << /Font << /F1 5 0 R >> >> >>",
		stream("BT /F1 24 Tf 72 700 Td (" + fixtureSentence + ") Tj ET"),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		"<< /Title (A Tiny Test Document) >>",
	}, true)
}

// pdfWithNoTextLayer paints an image over the whole page and writes no text operator at all —
// structurally what a 1998 Acrobat PDFWriter scan is, and what IEEE 1012 turned out to be.
// The image data is ASCIIHex so this file stays entirely printable.
func pdfWithNoTextLayer() []byte {
	const hexPixels = "000000FFFFFF000000FFFFFF>" // 4 RGB pixels, ASCIIHexDecode-terminated
	return buildPDF([]string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R " +
			"/Resources << /XObject << /Im1 5 0 R >> >> >>",
		stream("q 612 0 0 792 0 0 cm /Im1 Do Q"),
		fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 2 /Height 2 /ColorSpace /DeviceRGB "+
			"/BitsPerComponent 8 /Filter /ASCIIHexDecode /Length %d >>\nstream\n%s\nendstream",
			len(hexPixels), hexPixels),
	}, false)
}

func TestPDFExtractorReadsATextLayerAndItsTitle(t *testing.T) {
	got := PDFExtractor{}.Extract(t.TempDir(), "application/pdf", pdfWithTextLayer())
	if !got.Attempted {
		t.Fatal("Attempted = false for application/pdf — the extractor did not look")
	}
	if got.Reason != "" {
		t.Errorf("Reason = %q on a successful extraction; it must be empty", got.Reason)
	}
	if !strings.Contains(got.Text, fixtureSentence) {
		t.Errorf("Text = %q, want it to contain %q", got.Text, fixtureSentence)
	}
	if got.Title != "A Tiny Test Document" {
		t.Errorf("Title = %q, want the document's own /Title (D4 rung 1)", got.Title)
	}
	if got.Pages != 1 {
		t.Errorf("Pages = %d, want 1", got.Pages)
	}
	if got.ExtractorID != extractorIdentity() {
		t.Errorf("ExtractorID = %q, want %q", got.ExtractorID, extractorIdentity())
	}
	if got.OCRDerived {
		t.Error("OCRDerived = true for text read from a text layer")
	}
}

// THE FAILURE THIS WHOLE CHANGE EXISTS FOR. IEEE 1012 was 11 of the 33 PDF reads in the
// 2026-08-23 programme and has no text layer on any page; the seat got nothing and rendered
// 3.61 MB of base64 page images instead. Empty must arrive as a STATED fact, and the reason
// must be specific enough that a reader knows OCR is what is missing.
func TestPDFExtractorStatesWhyAScanHasNoText(t *testing.T) {
	got := PDFExtractor{}.Extract(t.TempDir(), "application/pdf", pdfWithNoTextLayer())
	if !got.Attempted {
		t.Fatal("Attempted = false — a scan must be looked at and reported on, not skipped")
	}
	if got.Text != "" {
		t.Errorf("Text = %q, want empty for a page with no text operators", got.Text)
	}
	if got.Reason == "" {
		t.Fatal("empty extraction with no Reason — the miss and the honest zero are then the same bytes")
	}
	// The reason must name the REMEDY, not merely the symptom: "no text" leaves a reader to
	// guess between a broken tool and an image-only document, and those call for opposite acts.
	for _, want := range []string{"no text layer", "recognition"} {
		if !strings.Contains(got.Reason, want) {
			t.Errorf("Reason = %q, want it to mention %q", got.Reason, want)
		}
	}
	if got.Pages != 1 {
		t.Errorf("Pages = %d, want 1 — page count is known even when text is not", got.Pages)
	}
}

// A CONTENT TYPE NOTHING EXTRACTS IS "NOT MEASURED", NEVER "NO TEXT". An HTML page is entirely
// text; recording text_extracted:false for one would be a false statement about the document
// rather than a true statement about the extractor.
func TestPDFExtractorDoesNotLookAtOtherContentTypes(t *testing.T) {
	for _, mt := range []string{"text/html", "text/plain", "application/json", ""} {
		got := PDFExtractor{}.Extract(t.TempDir(), mt, []byte("<html>a page of words</html>"))
		if got.Attempted {
			t.Errorf("Attempted = true for %q — only application/pdf is extracted", mt)
		}
		if got.Reason != "" || got.Text != "" {
			t.Errorf("%q produced Reason=%q Text=%q, want a wholly zero Extraction", mt, got.Reason, got.Text)
		}
	}
}

// GARBAGE IN, A NAMED ERROR OUT. Measured over 159 files: encrypted, permission-locked and
// malformed documents all land here with a specific message. What must never happen is the
// quiet empty string, which a reader cannot tell from a scan.
func TestPDFExtractorNamesWhatItCouldNotOpen(t *testing.T) {
	got := PDFExtractor{}.Extract(t.TempDir(), "application/pdf", []byte("this is not a PDF at all"))
	if !got.Attempted {
		t.Fatal("Attempted = false — it was handed application/pdf and must report on it")
	}
	if got.Text != "" {
		t.Errorf("Text = %q from unparseable bytes", got.Text)
	}
	if !strings.Contains(got.Reason, "could not be opened") {
		t.Errorf("Reason = %q, want it to say the document could not be opened", got.Reason)
	}
}

// THE FALLBACK CANNOT GO STALE SILENTLY. extractorIdentity reads the real module graph out of
// the linked binary, so a shipped build never uses this constant — but a TEST binary carries an
// empty Deps list (measured: main is populated, deps=0), and so does anything else built without
// module info. This gate reads go.mod ITSELF, which is the record, rather than comparing the
// constant to another copy of it.
func TestExtractorFallbackMatchesGoMod(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var mod string
	for {
		candidate := filepath.Join(dir, "go.mod")
		if _, statErr := os.Stat(candidate); statErr == nil {
			mod = candidate
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("no go.mod found above the package directory — this gate cannot measure anything")
		}
		dir = parent
	}
	b, err := os.ReadFile(mod)
	if err != nil {
		t.Fatal(err)
	}
	var got string
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) >= 2 && fields[0] == extractorModule {
			got = fields[1]
			break
		}
	}
	if got == "" {
		t.Fatalf("%s requires no %s — extractorModule names a dependency the module does not have", mod, extractorModule)
	}
	if got != extractorFallbackVersion {
		t.Errorf("extractorFallbackVersion = %q but %s requires %q — a build without module info "+
			"would key every cached extraction to a version that did not produce it",
			extractorFallbackVersion, mod, got)
	}
}

// And the identity a real extraction records must name the module, whichever path produced it.
func TestExtractorIdentityNamesTheModule(t *testing.T) {
	id := extractorIdentity()
	module, version, found := strings.Cut(id, "@")
	if !found || module != extractorModule {
		t.Fatalf("extractorIdentity() = %q, want %s@<version>", id, extractorModule)
	}
	if !strings.HasPrefix(version, "v") {
		t.Errorf("extractorIdentity() = %q — the version half must be a semver, since D3 makes "+
			"semver the audit key", id)
	}
}

func TestNormalizeExtractedStripsTheHyphenationMarkerAndKeepsTheWords(t *testing.T) {
	// U+0002 is what PDFium writes at a hyphenation point — 337 of them in the 67-page Settles
	// survey. Left in, they land mid-word in text a seat quotes into a report.
	for _, tc := range []struct{ name, in, want string }{
		{"hyphenation marker", "re\x02search", "research"},
		{"CRLF becomes LF", "a\r\nb", "a\nb"},
		{"tabs and newlines survive", "a\tb\nc", "a\tb\nc"},
		{"C1 controls go", "a\u009fb", "ab"},
		{"letters and marks survive", "l = \u03bbw \u2014 \u2318", "l = \u03bbw \u2014 \u2318"},
		{"nothing but controls is empty", "\x02\r\x01", ""},
		{"whitespace only is empty", "  \n\t ", ""},
	} {
		if got := normalizeExtracted(tc.in); got != tc.want {
			t.Errorf("%s: normalizeExtracted(%q) = %q, want %q", tc.name, tc.in, got, tc.want)
		}
	}
}
