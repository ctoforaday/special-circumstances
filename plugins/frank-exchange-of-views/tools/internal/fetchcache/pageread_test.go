package fetchcache

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record/runtest"
)

// NOTHING HERE CALLS A MODEL. DefaultPageReader is a variable for exactly this reason: a test
// that needed credentials would be a test nobody can run, and one that spent money on every
// `go test ./...` would be worse. The stub is where the two-pass behaviour is actually
// exercised, because what is under test is the COMPARISON, not the transcription.
type stubReader struct {
	// perCall returns the text for the nth call (1-based), so a test can make two passes of
	// one page disagree — which is the case the whole record exists to represent.
	perCall func(n int) (string, error)
	calls   int
	in, out int64
}

func (s *stubReader) ReadPage(_ context.Context, _ string, _ []byte) (string, int64, int64, error) {
	s.calls++
	text, err := s.perCall(s.calls)
	return text, s.in, s.out, err
}

func withReader(t *testing.T, r PageReader) {
	t.Helper()
	prev := DefaultPageReader
	DefaultPageReader = r
	t.Cleanup(func() { DefaultPageReader = prev })
}

// renderOnePage puts a rendered single-page document in a run and returns it.
func renderOnePage(t *testing.T) (record.Run, string, RenderRecord) {
	t.Helper()
	run := runtest.New(t, t.TempDir())
	body := pdfWithNoTextLayer()
	sha := Sha(body)
	rd, err := RenderPages(run, sha, body, MinRenderDPI)
	if err != nil {
		t.Fatal(err)
	}
	return run, sha, rd
}

func TestReadingAgreesWhenBothPassesSayTheSameThing(t *testing.T) {
	withReader(t, &stubReader{in: 1200, out: 40, perCall: func(int) (string, error) {
		return "Hello leaf verification\n", nil
	}})
	run, sha, rd := renderOnePage(t)

	got, err := ReadRenderedPages(context.Background(), run, sha, "test-model", rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	if len(got.Pages) != 1 {
		t.Fatalf("Pages = %d, want 1", len(got.Pages))
	}
	p := got.Pages[0]
	if !p.Agreed {
		t.Errorf("identical passes reported as disagreeing")
	}
	if p.FirstDifferenceAt != -1 {
		t.Errorf("FirstDifferenceAt = %d on agreement, want -1", p.FirstDifferenceAt)
	}
	if len(p.PassShas) != PassCount {
		t.Errorf("PassShas = %d, want %d — every pass is recorded, not just the winner", len(p.PassShas), PassCount)
	}
	if len(got.Divergences()) != 0 {
		t.Errorf("Divergences = %v on an agreed page", got.Divergences())
	}
	// THE ATTESTATION, which is what replaces reproducibility here.
	if got.Model != "test-model" || got.ReadAt.IsZero() || got.DPI != rd.DPI {
		t.Errorf("attestation incomplete: model=%q readAt=%v dpi=%d", got.Model, got.ReadAt, got.DPI)
	}
	if len(got.RenderShas) != 1 || got.RenderShas[0] != rd.PageShas[0] {
		t.Error("the reading does not name the exact images it read — a re-render would go unnoticed")
	}
	if got.InTokens != 2400 || got.OutTok != 80 {
		t.Errorf("tokens = %d in / %d out, want both passes counted (2400/80)", got.InTokens, got.OutTok)
	}
	// The agreed text is the deliverable, and both passes survive as evidence.
	body, err := os.ReadFile(OCRTextPath(run, sha))
	if err != nil {
		t.Fatalf("assembled text: %v", err)
	}
	if !strings.Contains(string(body), "Hello leaf verification") {
		t.Errorf("assembled text = %q", body)
	}
	if Sha(body) != got.TextSha {
		t.Error("text_sha does not hash the text on disk")
	}
	for pass := 1; pass <= PassCount; pass++ {
		if _, err := os.Stat(PassPath(run, sha, 1, pass)); err != nil {
			t.Errorf("pass %d was not kept: %v — a disagreement with no exhibit is an accusation", pass, err)
		}
	}
}

// THE CASE THE WHOLE DESIGN EXISTS FOR. Two readings that differ must not be resolved by
// picking one: an uncorroborated reading in the position a citation is taken from is exactly
// the fluent-failure #644 names.
func TestReadingMarksADisagreementInPlaceAndNeverPicksAWinner(t *testing.T) {
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 1 {
			return "the method of types", nil
		}
		return "the method of typos", nil
	}})
	run, sha, rd := renderOnePage(t)

	got, err := ReadRenderedPages(context.Background(), run, sha, "test-model", rd)
	if err != nil {
		t.Fatalf("ReadRenderedPages: %v", err)
	}
	p := got.Pages[0]
	if p.Agreed {
		t.Fatal("two different readings were reported as agreeing")
	}
	if p.FirstDifferenceAt != 17 {
		t.Errorf("FirstDifferenceAt = %d, want 17 (the 'e'/'o' of types/typos)", p.FirstDifferenceAt)
	}
	if len(got.Divergences()) != 1 || got.Divergences()[0] != 1 {
		t.Errorf("Divergences = %v, want [1]", got.Divergences())
	}

	text, err := os.ReadFile(OCRTextPath(run, sha))
	if err != nil {
		t.Fatal(err)
	}
	// NEITHER READING may appear as if it were the page's text.
	for _, forbidden := range []string{"the method of types", "the method of typos"} {
		if strings.Contains(string(text), forbidden) {
			t.Errorf("the assembled text contains %q — a disagreement was silently resolved", forbidden)
		}
	}
	if !strings.Contains(string(text), "disagree") {
		t.Errorf("the assembled text does not mark the disagreement:\n%s", text)
	}
	// And both sides are on disk to be compared.
	one, _ := os.ReadFile(PassPath(run, sha, 1, 1))
	two, _ := os.ReadFile(PassPath(run, sha, 1, 2))
	if string(one) == string(two) || len(one) == 0 || len(two) == 0 {
		t.Error("the two passes were not both kept")
	}
}

// WHITESPACE IS NOT A DISAGREEMENT. A check that fired on a trailing newline would be ignored
// within a week, which is the failure mode of a guard that cries wolf.
func TestReadingIgnoresWhitespaceDifferencesButNotStructuralOnes(t *testing.T) {
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 1 {
			return "line one\nline two", nil
		}
		return "  line one   \r\nline two\n\n  ", nil // same content, different whitespace
	}})
	run, sha, rd := renderOnePage(t)
	got, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Pages[0].Agreed {
		t.Error("passes differing only in trailing whitespace and line endings were called a disagreement")
	}

	// But a merged line IS a different reading of the page.
	withReader(t, &stubReader{perCall: func(n int) (string, error) {
		if n == 1 {
			return "line one\nline two", nil
		}
		return "line one line two", nil
	}})
	run2, sha2, rd2 := renderOnePage(t)
	got2, err := ReadRenderedPages(context.Background(), run2, sha2, "m", rd2)
	if err != nil {
		t.Fatal(err)
	}
	if got2.Pages[0].Agreed {
		t.Error("one pass merged two lines and that was reported as agreement")
	}
}

// A READER THAT FAILED MUST NOT LOOK LIKE A BLANK PAGE. An empty string is a legitimate
// reading — #644's own prompt tells the model to return nothing for a blank page — so a
// failure has to arrive as an error or the two are indistinguishable.
func TestReadingPropagatesAReaderFailureRatherThanRecordingAnEmptyPage(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) {
		return "", errors.New("credentials refused (401)")
	}})
	run, sha, rd := renderOnePage(t)
	_, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err == nil {
		t.Fatal("a failing reader produced a reading record")
	}
	if !strings.Contains(err.Error(), "credentials refused") {
		t.Errorf("err = %v, want the reader's own message", err)
	}
	// And nothing was recorded, so a retry is clean.
	if _, ok, _ := ReadReadingRecord(run, sha); ok {
		t.Error("a reading record was written despite the failure")
	}
}

// A BLANK PAGE IS A READING, and must be recorded as one rather than as a failure.
func TestReadingAcceptsAGenuinelyBlankPage(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "", nil }})
	run, sha, rd := renderOnePage(t)
	got, err := ReadRenderedPages(context.Background(), run, sha, "m", rd)
	if err != nil {
		t.Fatalf("a blank page was treated as a failure: %v", err)
	}
	if !got.Pages[0].Agreed {
		t.Error("two blank readings disagreed")
	}
	if got.Pages[0].Lengths[0] != 0 {
		t.Errorf("Lengths = %v, want 0 for a blank page", got.Pages[0].Lengths)
	}
}

// THE CAP REFUSES; IT DOES NOT TRUNCATE. Reading the first N pages and recording it as the
// document's reading would be a false statement nothing downstream could detect.
func TestReadingRefusesADocumentOverTheCapRatherThanReadingPartOfIt(t *testing.T) {
	withReader(t, &stubReader{perCall: func(int) (string, error) { return "x", nil }})
	run := runtest.New(t, t.TempDir())
	over := make([]string, MaxReadPages+1)
	for i := range over {
		over[i] = "sha"
	}
	_, err := ReadRenderedPages(context.Background(), run, "s", "m",
		RenderRecord{Sha: "s", DPI: 200, PageShas: over})
	if err == nil {
		t.Fatal("a document over the cap was read")
	}
	for _, want := range []string{"cap", "model calls"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
}

func TestReadReadingRecordReportsAbsenceAndRefusesAMisfiledRecord(t *testing.T) {
	run := runtest.New(t, t.TempDir())
	if _, ok, err := ReadReadingRecord(run, "never-read"); err != nil || ok {
		t.Errorf("ReadReadingRecord on an unread document = (%v, %v), want (false, nil)", ok, err)
	}
	if err := os.MkdirAll(PagesDir(run, "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(readingRecordPath(run, "x"), []byte(`{"sha":"someone-else"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadReadingRecord(run, "x"); err == nil || !strings.Contains(err.Error(), "names document") {
		t.Errorf("a record naming another document was accepted: %v", err)
	}
}

// The offset is in RUNES. A byte offset into a CJK transcription points inside a character
// and reads as a wrong answer to the human it is shown to.
func TestFirstRuneDifferenceCountsCharactersNotBytes(t *testing.T) {
	for _, tc := range []struct {
		name, a, b string
		want       int
	}{
		{"ascii", "abcdef", "abcXef", 3},
		{"multibyte before the difference", "日本語です", "日本語でした", 4},
		{"a prefix differs where the shorter ends", "abc", "abcdef", 3},
		{"identical", "same", "same", 4},
	} {
		if got := firstRuneDifference(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: firstRuneDifference(%q,%q) = %d, want %d", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}

// AN OVER-LIMIT IMAGE IS REFUSED BY NAME, NOT BY THE API. `ocr pages` accepts --dpi up to 400,
// and a letter page of scan-like noise measures 12.69 MB of PNG at that resolution — 16.92 MB
// once base64-encoded, past the API's 10 MB per-image ceiling. Without this the operator gets
// an opaque request-too-large from inside the SDK, with nothing connecting it to the flag that
// caused it.
//
// The rule is a pure function of a length precisely so this test needs no model: asserting it
// through ReadPage would mean handing ReadPage a legal image and letting it call out.
func TestAnImageOverTheAPILimitIsRefusedBeforeTheCall(t *testing.T) {
	err := imageWithinAPILimit(MaxImageBytes + 1)
	if err == nil {
		t.Fatal("an image past the per-image limit was accepted")
	}
	for _, want := range []string{"per-image limit", "--dpi 200", "base64"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal = %q, want it to mention %q", err, want)
		}
	}
	// THE LIMIT ITSELF IS USABLE. A guard that refused its own boundary would be a different
	// bug wearing this one's clothes.
	if err := imageWithinAPILimit(MaxImageBytes); err != nil {
		t.Errorf("an image at exactly the limit was refused: %v", err)
	}
	// And the binary ceiling really does leave room once base64 inflates it by 4/3.
	if MaxImageBytes*4/3 >= 10<<20 {
		t.Errorf("MaxImageBytes=%d base64-encodes to %d, past the API's 10 MB limit",
			MaxImageBytes, MaxImageBytes*4/3)
	}
}
