package fetchcache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/record"
)

// Reading a rendered page is the second half of #644, and the half that spends money.
//
// #658 rendered the pages. This asks a model what they say — TWICE, independently, because
// the two-pass agreement is the only check available on a reader that cannot be re-run for
// the same answer. Tesseract's determinism guarantees agreement even when it is wrong; a
// model's nondeterminism is what makes a second opinion mean something. The property that
// looks like the weakness supplies the check.

// PageReader asks something what a page image says. It exists so the whole verb is testable
// without a network call — the same reason DefaultExtractor is a variable.
//
// It returns the text and the tokens it cost. A reader that could not answer returns an
// error: an empty string is a legitimate reading (a blank page) and must never be how a
// failure arrives.
type PageReader interface {
	ReadPage(ctx context.Context, model string, png []byte) (text string, in, out int64, err error)
}

// DefaultPageReader is the process-wide reader. A test replaces it; production never does.
var DefaultPageReader PageReader = AnthropicPageReader{}

// readPrompt is what the model is asked. It is deliberately narrow.
//
// THE INSTRUCTION IS "TRANSCRIBE", NOT "DESCRIBE", and the difference is the whole risk. A
// model asked what a page is about will summarise, and a summary of a source document is a
// paraphrase presented in the position a quotation occupies. What this record is for is
// citation at the leaf, so the only useful output is the characters that are on the page.
//
// It is also told to mark illegibility rather than guess. A model handed a smudge will
// produce fluent plausible text if nothing tells it not to, and that failure is invisible
// downstream — see #644. `[illegible]` is a fact; an invented word is not.
const readPrompt = `Transcribe the text on this page image, exactly as it appears.

Rules:
- Output ONLY the transcription. No preamble, no commentary, no summary.
- Preserve the reading order, line breaks and paragraph breaks of the page.
- Where characters are too degraded to read, write [illegible] rather than guessing. A wrong
  word that reads plausibly is worse than a marked gap.
- If the page carries no text at all (a blank page, or a figure with no caption), output
  nothing at all.
- Do not translate, correct spelling, expand abbreviations, or normalise anything.`

// AnthropicPageReader is the production reader: the Anthropic Go SDK, one message per page,
// the image as base64 PNG.
//
// CREDENTIALS ARE THE CALLER'S ENVIRONMENT, not this tool's business — the SDK resolves
// ANTHROPIC_API_KEY, then ANTHROPIC_AUTH_TOKEN, then an `ant auth login` profile. What this
// type owes is a LOUD failure when none resolves, which is why the error below names what to
// set rather than surfacing a bare 401.
type AnthropicPageReader struct{}

func (AnthropicPageReader) ReadPage(ctx context.Context, model string, png []byte) (string, int64, int64, error) {
	if err := imageWithinAPILimit(len(png)); err != nil {
		return "", 0, 0, err
	}
	client := anthropic.NewClient()
	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: 16000,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewImageBlockBase64("image/png", base64.StdEncoding.EncodeToString(png)),
				anthropic.NewTextBlock(readPrompt),
			),
		},
	})
	if err != nil {
		var apiErr *anthropic.Error
		if errors.As(err, &apiErr) && (apiErr.StatusCode == 401 || apiErr.StatusCode == 403) {
			return "", 0, 0, fmt.Errorf("the model refused this request for want of credentials (%d). "+
				"`ocr read` is the one verb in this tool that calls out: set ANTHROPIC_API_KEY, or run "+
				"`ant auth login`, then retry. Nothing has been written: %w", apiErr.StatusCode, err)
		}
		return "", 0, 0, err
	}
	var sb strings.Builder
	for _, block := range msg.Content {
		if t, okBlock := block.AsAny().(anthropic.TextBlock); okBlock {
			sb.WriteString(t.Text)
		}
	}
	return sb.String(), msg.Usage.InputTokens, msg.Usage.OutputTokens, nil
}

// MaxImageBytes is the largest page image the reader will send, in PNG bytes.
//
// THE PUBLISHED LIMIT IS ON THE BASE64 FORM — 10 MB per image on the Claude API — and base64
// inflates by 4/3, so the binary ceiling is 7.5 MB. Stating it in the units the caller has (a
// file on disk) is the point: a limit expressed in the encoded form is one every caller has to
// convert before it can be checked, and most will not.
//
// It is a real ceiling, not a theoretical one. A letter page of scan-like noise measures
// 3.37 MB at 200 DPI, 7.33 MB at 300 and 12.69 MB at 400 — and `ocr pages` accepts --dpi up to
// 400. Above ~200 DPI the resolution buys nothing anyway: the model's high-resolution tier caps
// an image at 2576 px on the long edge and 4784 visual tokens (one per 28x28 patch), and a
// letter page at 200 DPI is 1700x2200 = 4819 tokens — already at the ceiling. Everything past
// that is downscaled by the API before the model ever sees it.
const MaxImageBytes = 7 << 20 // 7 MB binary ~= 9.3 MB base64, inside the 10 MB limit

// imageWithinAPILimit refuses a page image the API would reject, by name and before the call.
//
// IT IS A SEPARATE FUNCTION SO IT CAN BE TESTED WITHOUT A MODEL. Checked inline in ReadPage,
// the only way to exercise the boundary would be to hand ReadPage an image at exactly the
// limit and let it proceed — into a real API call, in a package where nothing else does that.
// A pure function of a length is the whole rule, and the whole rule is testable offline.
//
// `ocr pages` accepts --dpi up to 400 and will happily render past this, so without the guard
// the operator gets an opaque request-too-large from inside the SDK with nothing connecting it
// to the flag that caused it.
func imageWithinAPILimit(n int) error {
	if n <= MaxImageBytes {
		return nil
	}
	return fmt.Errorf("page image is %.1f MB, which base64-encodes to %.1f MB and exceeds the "+
		"API's %d MB per-image limit; re-render lower with `ocr pages --dpi 200` (200 DPI is "+
		"already at the model's resolution ceiling for a letter-size page, so nothing is lost)",
		float64(n)/(1<<20), float64(n)*4/3/(1<<20), MaxImageBytes*4/3/(1<<20))
}

// MaxReadPages bounds one invocation.
//
// A PAGE IS A MODEL CALL, AND TWO PASSES DOUBLE IT. The 534-page Unicode CJK chart in #636's
// corpus is 1,068 calls, which is not a thing to reach by typo. The cap is a refusal a caller
// lifts deliberately, not a silent truncation — truncating would produce a reading of "the
// document" that is really a reading of its first N pages, and nothing downstream could see
// the difference.
const MaxReadPages = 50

// PassCount is how many independent readings each page gets.
//
// TWO IS THE MINIMUM THAT MEANS ANYTHING and the most that is worth paying for by default.
// One reading has no check at all. Three would break ties by majority, which sounds better
// and is worse here: a majority vote MANUFACTURES a confidence that two-of-three agreement
// does not earn on a degraded page, and it hides the disagreement this record exists to
// surface. Two readings either agree — evidence — or they do not, and then a human is told
// exactly where.
const PassCount = 2

// PageReading is what the two passes produced for one page.
type PageReading struct {
	// Page is 1-based, matching PagePath and how a document is cited.
	Page int `json:"page"`
	// PassShas is the sha256 of each pass's text, in pass order. The texts themselves are on
	// disk beside the image; hashes here keep the record small and make a changed pass visible.
	PassShas []string `json:"pass_shas"`
	// Agreed is whether every pass produced the same text after whitespace normalisation.
	// Exact equality would report disagreement over a trailing newline, which is not a
	// disagreement about what the page says.
	Agreed bool `json:"agreed"`
	// FirstDifferenceAt is the rune offset where the passes first diverge, or -1 when they
	// agree. A rune offset rather than a byte offset, so it points at a character in text a
	// human is going to read.
	FirstDifferenceAt int `json:"first_difference_at"`
	// Lengths is each pass's normalised length in runes. Two readings that differ only in
	// length by a lot is a different failure from two that differ early and then re-converge.
	Lengths []int `json:"lengths"`
}

// ReadingRecord is one document's reading, written beside its page images.
//
// IT IS ITS OWN RECORD FOR THE REASON RenderRecord IS: the cache index is append-only and
// Lookup takes the first match, so it cannot express an update. A reading is also a distinct
// later act with its own facts — which model, when, against which rendered images.
//
// THE ATTESTATION IS THE POINT, AND IT REPLACES REPRODUCIBILITY. #636 keyed an extraction to
// library@semver so an audit could RE-RUN it and compare. A model re-reading the same page
// returns different bytes, so that check is not available here and the record must not imply
// it is. What this offers instead is: which model, at what time, against which image hashes,
// and whether a second independent reading agreed. That is weaker than re-derivation and it
// is stated rather than left to be discovered.
type ReadingRecord struct {
	Sha   string `json:"sha"`
	Model string `json:"model"`
	// ReadAt is when. A model changes underneath a fixed name, so the name alone does not
	// identify what did the reading.
	ReadAt time.Time `json:"read_at"`
	// RenderShas binds this reading to the exact images it read. A re-render at a new DPI
	// makes the reading stale, and this is what lets a reader notice rather than assume.
	RenderShas []string `json:"render_shas"`
	// DPI the images were rendered at, carried so a poor reading can be attributed.
	DPI      int           `json:"dpi"`
	Pages    []PageReading `json:"pages"`
	TextSha  string        `json:"text_sha"`
	InTokens int64         `json:"input_tokens"`
	OutTok   int64         `json:"output_tokens"`
}

// Divergences is the pages whose passes disagreed.
//
// A DERIVED COUNT WOULD BE A SECOND COPY. Nothing stores "how many pages diverged": it is
// computed from the rows, so it cannot drift from them.
func (r ReadingRecord) Divergences() []int {
	var out []int
	for _, p := range r.Pages {
		if !p.Agreed {
			out = append(out, p.Page)
		}
	}
	return out
}

func readingRecordPath(run record.Run, sha string) string {
	return filepath.Join(PagesDir(run, sha), "reading.json")
}

// PassPath is one pass's transcription for one page. Both passes are kept, not just the
// agreed text: when they disagree, the two files ARE the evidence a human needs, and a record
// that named a disagreement without keeping the sides would be an accusation with no exhibit.
func PassPath(run record.Run, sha string, page, pass int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.pass%d.txt", page, pass))
}

// OCRTextPath is the assembled reading, beside the extraction TextPath would hold.
// The suffix says which is which to a human listing the directory; no reader recovers a fact
// from it — `ocr_derived` on the record is what states the difference.
func OCRTextPath(run record.Run, sha string) string { return Path(run, sha) + ".ocr.txt" }

// ReadRenderedPages runs the passes over every rendered page and records what came back.
func ReadRenderedPages(ctx context.Context, run record.Run, sha, model string, rd RenderRecord) (ReadingRecord, error) {
	if rd.Pages() == 0 {
		return ReadingRecord{}, fmt.Errorf("the render record for %s names no pages", sha)
	}
	if rd.Pages() > MaxReadPages {
		return ReadingRecord{}, fmt.Errorf("%d pages is over the %d-page cap for one reading, and each "+
			"page costs %d model calls (%d in total here). Re-render a subset, or raise the cap "+
			"deliberately — this refuses rather than reading part of the document and calling it the "+
			"document", rd.Pages(), MaxReadPages, PassCount, rd.Pages()*PassCount)
	}

	out := ReadingRecord{
		Sha: sha, Model: model, ReadAt: time.Now().UTC(),
		RenderShas: rd.PageShas, DPI: rd.DPI,
	}
	var assembled strings.Builder

	for i := 1; i <= rd.Pages(); i++ {
		png, err := os.ReadFile(PagePath(run, sha, i))
		if err != nil {
			return ReadingRecord{}, fmt.Errorf("page %d image: %w", i, err)
		}
		pr := PageReading{Page: i, FirstDifferenceAt: -1}
		texts := make([]string, 0, PassCount)
		for pass := 1; pass <= PassCount; pass++ {
			text, in, outTok, rerr := DefaultPageReader.ReadPage(ctx, model, png)
			if rerr != nil {
				return ReadingRecord{}, fmt.Errorf("page %d pass %d: %w", i, pass, rerr)
			}
			out.InTokens += in
			out.OutTok += outTok
			norm := normalizeReading(text)
			texts = append(texts, norm)
			if err := writeReplacing(PassPath(run, sha, i, pass), []byte(norm)); err != nil {
				return ReadingRecord{}, err
			}
			pr.PassShas = append(pr.PassShas, Sha([]byte(norm)))
			pr.Lengths = append(pr.Lengths, len([]rune(norm)))
		}
		pr.Agreed, pr.FirstDifferenceAt = comparePasses(texts)
		out.Pages = append(out.Pages, pr)

		// THE ASSEMBLED TEXT IS THE AGREED TEXT, AND A DISAGREEMENT IS MARKED IN PLACE rather
		// than resolved. Picking a winner would put a reading nothing corroborated into the
		// position a citation is taken from; the marker sends a reader to the two passes on disk.
		if pr.Agreed {
			assembled.WriteString(texts[0])
		} else {
			fmt.Fprintf(&assembled, "[page %d: the two readings disagree from character %d; see %s and %s]",
				i, pr.FirstDifferenceAt,
				filepath.Base(PassPath(run, sha, i, 1)), filepath.Base(PassPath(run, sha, i, 2)))
		}
		assembled.WriteString("\n\n")
	}

	text := strings.TrimRight(assembled.String(), "\n")
	if err := writeReplacing(OCRTextPath(run, sha), []byte(text)); err != nil {
		return ReadingRecord{}, err
	}
	out.TextSha = Sha([]byte(text))

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return ReadingRecord{}, err
	}
	// The record is written LAST, so a crash leaves passes with no record — which reads as
	// "not read" and re-reads cleanly, the same ordering RenderPages relies on.
	if err := writeReplacing(readingRecordPath(run, sha), append(b, '\n')); err != nil {
		return ReadingRecord{}, err
	}
	return out, nil
}

// ReadReadingRecord returns a document's reading, and whether one exists. A malformed record
// is an error rather than an absence, for the reason ReadRenderRecord gives.
func ReadReadingRecord(run record.Run, sha string) (ReadingRecord, bool, error) {
	b, err := os.ReadFile(readingRecordPath(run, sha))
	if os.IsNotExist(err) {
		return ReadingRecord{}, false, nil
	}
	if err != nil {
		return ReadingRecord{}, false, err
	}
	var r ReadingRecord
	if err := json.Unmarshal(b, &r); err != nil {
		return ReadingRecord{}, false, fmt.Errorf("reading record for %s is unreadable: %w", sha, err)
	}
	if r.Sha != sha {
		return ReadingRecord{}, false, fmt.Errorf("reading record under %s names document %s", sha, r.Sha)
	}
	return r, true, nil
}

// writeReplacing writes b to dst, REPLACING whatever is there.
//
// Not writeAtomic, and the difference is the bug #658 shipped and caught: writeAtomic treats
// an existing destination as already done, which is right for a content-addressed path and
// wrong for a stable name whose content varies. A second reading of the same page is new
// content at the same path, exactly as a re-render at a new DPI was.
func writeReplacing(dst string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		os.Remove(name)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(name)
		return err
	}
	if err := os.Rename(name, dst); err != nil {
		os.Remove(name)
		return err
	}
	return nil
}

// normalizeReading collapses the differences that are not disagreements about content.
//
// Two readings of one page routinely differ in trailing whitespace and in whether a line ends
// CRLF. Reporting those as a divergence would fill the record with noise and teach a reader
// to ignore the field, which is the failure mode of a check that cries wolf. Interior line
// structure is PRESERVED — a model that merged two paragraphs read the page differently, and
// that is a real disagreement.
func normalizeReading(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// comparePasses reports whether every pass agrees, and where the first two diverge.
//
// The offset is in RUNES because it is shown to a human who will count characters in text,
// not bytes in a buffer — a byte offset in a CJK transcription points into the middle of a
// character and reads as a wrong answer.
func comparePasses(texts []string) (agreed bool, firstDiff int) {
	if len(texts) < 2 {
		// One pass cannot disagree with itself, and reporting `agreed: true` for it would
		// overstate what a single reading establishes. It is "no disagreement found", which is
		// what -1 with agreed=true has to mean here; the record's PassShas length is what tells
		// a reader how many opinions stand behind it.
		return true, -1
	}
	for i := 1; i < len(texts); i++ {
		if texts[i] != texts[0] {
			return false, firstRuneDifference(texts[0], texts[i])
		}
	}
	return true, -1
}

func firstRuneDifference(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := len(ra)
	if len(rb) < n {
		n = len(rb)
	}
	for i := 0; i < n; i++ {
		if ra[i] != rb[i] {
			return i
		}
	}
	// One is a prefix of the other: they first differ where the shorter one ends.
	return n
}
