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
// #658 rendered the pages. This asks a model what they say, ONCE.
//
// IT ASKED TWICE UNTIL THE COST OF THE SECOND PASS WAS PUT AGAINST ITS EVIDENCE. The argument
// for two was that a model cannot be re-run for the same answer, so an independent second
// reading is the only corroboration available. The argument against is that this doubles the
// price of every page to catch an error rate nobody had measured — and the measurement was
// never made, because it needs credentials this container does not have. A check whose value
// is assumed rather than measured does not get to double the bill: #659 ruled one pass, and
// the second returns only if a real error rate is shown to need it. The two-pass comparison
// is recoverable from history at 9e899d27 if that day comes.
//
// WHAT THAT COSTS IS CORROBORATION, AND NOTHING HERE PRETENDS OTHERWISE. A single reading is
// uncorroborated: the record attests which model read which images when, and claims no more.
// The alternative considered was keeping the fields and filling them from one pass — `agreed:
// true` on every page forever. That is worse than dropping them, because a field that reports
// the healthy value whether or not anything was checked is indistinguishable from a check
// that ran and passed. Absent is honest; always-true is a lie with a schema.

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
//
// TWO RULES WERE ADDED FROM THE FIRST REAL MEASUREMENT (#644, IEEE 1012, 63 pages read).
// The transcription was near-verbatim, and its two systematic deviations were both
// prompt-shaped: every bold or italic passage came back wrapped in markdown the page does
// not carry (a seat quoting the transcript would quote asterisks that are not on the
// page), and the one blank page came back as `[illegible]` — which claims unreadable
// characters EXIST where there are none, a different fact from empty. Tables are the
// deliberate exception: the model rendered an 11-column matrix as |-separated rows
// unprompted, and against the pixels that table was right in 329 of ~330 cells — column
// alignment is content on a table, so the layout is kept and now stated rather than
// accidental.
const readPrompt = `Transcribe the text on this page image, exactly as it appears.

Rules:
- Output ONLY the transcription. No preamble, no commentary, no summary.
- Preserve the reading order, line breaks and paragraph breaks of the page.
- Plain characters only: do not add markup the page does not carry — no **bold**, no
  *italics*, no # headings, no <sup></sup>. The one exception is a tabular layout, which
  you should transcribe as rows with cells separated by | so the columns survive.
- Where characters are too degraded to read, write [illegible] rather than guessing. A wrong
  word that reads plausibly is worse than a marked gap.
- If the page carries no text at all (a blank page, or a figure with no caption), output an
  empty response — nothing, not [illegible]: that marker claims unreadable characters exist,
  and a blank page has none.
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
// A PAGE IS A MODEL CALL. The 534-page Unicode CJK chart in #636's corpus is 534 calls, which
// is not a thing to reach by typo. The cap is a refusal a caller lifts deliberately, not a
// silent truncation — truncating would produce a reading of "the document" that is really a
// reading of its first N pages, and nothing downstream could see the difference.
//
// THE NUMBER IS DENOMINATED IN MODEL CALLS, NOT PAGES, which is why it moved when the second
// pass went. At two passes it was 50 pages = 100 calls; at one pass the same budget is 100
// pages. Leaving it at 50 would have quietly halved a ceiling nobody asked to lower, and it
// would refuse the 80-page IEEE 1012 scan that is the corpus's most obvious real test.
const MaxReadPages = 100

// PageReading is what the reader produced for one page.
type PageReading struct {
	// Page is 1-based, matching PagePath and how a document is cited.
	Page int `json:"page"`
	// TextSha is the sha256 of this page's transcription. The text itself is on disk beside
	// the image; the hash keeps the record small and makes a changed page visible.
	TextSha string `json:"text_sha"`
	// Length is the transcription's length in runes — a cheap shape check on a page that came
	// back far shorter than its neighbours, which is what a refusal or a truncation looks like.
	Length int `json:"length"`
	// Blocked marks a page the API's content filter refused to transcribe (#679). The page's
	// text is the named hole, not a transcription; this field is what machinery reads, the
	// marker in the text is what a citing seat reads, and both say the same fact.
	Blocked bool `json:"blocked,omitempty"`
	// BlockedReason is the API's message, verbatim with its request id, so the refusal can be
	// cited and re-examined without re-paying for it.
	BlockedReason string `json:"blocked_reason,omitempty"`
}

// blockedPolicySentence is the API's content-filter refusal, quoted exactly.
//
// THIS IS A STRING-SHAPED KEY AND THE FAILURE DIRECTION IS CLOSED, which is the answer
// [[facts-are-fields]] is owed for it: the API reports the block as a 400
// invalid_request_error with this sentence in the body and NO structured code to key on
// (measured on #679, 4/4 attempts, two models). If Anthropic rewords it, blockedByPolicy
// stops classifying, the error falls through to the fatal path, and the read refuses
// loudly — a wrong-way drift costs a usable document, never a silent hole.
const blockedPolicySentence = "Output blocked by content filtering policy"

// blockedByPolicy reports whether err is the content filter refusing to emit a
// transcription: a typed API 400 carrying the sentence. Anything else — credentials,
// rate limits, network — keeps today's fatal semantics.
func blockedByPolicy(err error) bool {
	var apiErr *anthropic.Error
	return errors.As(err, &apiErr) && apiErr.StatusCode == 400 &&
		strings.Contains(err.Error(), blockedPolicySentence)
}

// blockedMarker is the hole as a citing seat sees it, in the page text and the assembled
// document. The page number is IN the marker because the assembled text is where a reader
// meets it, offset from any per-page file.
func blockedMarker(page int) string {
	return fmt.Sprintf("[page %d: output blocked by content filtering policy]", page)
}

// pageReceipt is one page's provenance row: what was rendered, what read it, what came
// back, and what it cost. Receipts are written as the loop runs — the reading record is a
// PROJECTION of them, assembled last — so a crash mid-document leaves rows that a re-run
// validates page by page instead of re-paying (#679's page-79 problem).
type pageReceipt struct {
	Page      int       `json:"page"`
	RenderSha string    `json:"render_sha"`
	DPI       int       `json:"dpi"`
	Model     string    `json:"model"`
	ReadAt    time.Time `json:"read_at"`
	TextSha   string    `json:"text_sha"`
	Length    int       `json:"length"`
	InTokens  int64     `json:"input_tokens"`
	OutTok    int64     `json:"output_tokens"`
	Blocked   bool      `json:"blocked,omitempty"`
	Reason    string    `json:"blocked_reason,omitempty"`
}

// receiptPath is one page's receipt, beside its text.
func receiptPath(run record.Run, sha string, n int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.receipt.json", n))
}

// readReceipt returns a page's receipt, and whether one exists. A malformed receipt is an
// ERROR, never an absence: treated as absence it would re-spend a model call over a row
// somebody may already have cited from — the same rule every record reader here follows.
func readReceipt(run record.Run, sha string, n int) (pageReceipt, bool, error) {
	b, err := os.ReadFile(receiptPath(run, sha, n))
	if os.IsNotExist(err) {
		return pageReceipt{}, false, nil
	}
	if err != nil {
		return pageReceipt{}, false, err
	}
	var r pageReceipt
	if err := json.Unmarshal(b, &r); err != nil {
		return pageReceipt{}, false, fmt.Errorf("receipt for page %d of %s is unreadable: %w — "+
			"`ocr read --force` discards receipts and reads again", n, sha, err)
	}
	return r, true, nil
}

// ClearReceipts removes every receipt, which is what `ocr read --force` means: discard
// the paid-for rows, blocked ones included, and read everything again.
func ClearReceipts(run record.Run, sha string) error {
	matches, err := filepath.Glob(filepath.Join(PagesDir(run, sha), "p????.receipt.json"))
	if err != nil {
		return err
	}
	for _, m := range matches {
		if err := os.Remove(m); err != nil {
			return err
		}
	}
	return nil
}

// readPageStep is the ONE per-page act both loops share — the deliberate `ocr read` over
// rendered images and fetch's fused rasterise-and-read (#679 requires them identical, and
// a copy would let them drift on exactly the semantics that took a plan to settle).
//
// Given the page's PNG bytes as rendered NOW, it: reuses a receipt whose render sha and
// model match — including a blocked receipt, because a deterministic refusal is a fact
// already paid for — after verifying the text on disk still matches the receipt's hash;
// otherwise spends the model call, classifying a content-filter refusal into a named hole
// (marker text, blocked fields) and letting every other error stay fatal. Text first,
// receipt second, so a crash between the two leaves a page that re-reads rather than a
// receipt naming text that was never written.
func readPageStep(ctx context.Context, run record.Run, sha string, page int, png []byte, model string, dpi int) (pageReceipt, string, error) {
	renderSha := Sha(png)
	if r, ok, err := readReceipt(run, sha, page); err != nil {
		return pageReceipt{}, "", err
	} else if ok && r.RenderSha == renderSha && r.Model == model {
		b, rerr := os.ReadFile(PageTextPath(run, sha, page))
		if rerr != nil || Sha(b) != r.TextSha {
			return pageReceipt{}, "", fmt.Errorf("page %d of %s has a receipt but its text is missing or "+
				"altered (%v) — the pair is corrupt; `ocr read --force` discards receipts and reads again",
				page, sha, rerr)
		}
		return r, string(b), nil
	}

	text, in, outTok, rerr := DefaultPageReader.ReadPage(ctx, model, png)
	blocked := false
	reason := ""
	if rerr != nil {
		if !blockedByPolicy(rerr) {
			// A READER ERROR IS AN ERROR, NOT AN EMPTY PAGE — unchanged from the original
			// design; only the filter's deterministic refusal becomes a hole.
			return pageReceipt{}, "", rerr
		}
		blocked = true
		reason = rerr.Error()
		text = blockedMarker(page)
	}

	norm := normalizeReading(text)
	if err := writeReplacing(PageTextPath(run, sha, page), []byte(norm)); err != nil {
		return pageReceipt{}, "", err
	}
	r := pageReceipt{
		Page: page, RenderSha: renderSha, DPI: dpi, Model: model, ReadAt: time.Now().UTC(),
		TextSha: Sha([]byte(norm)), Length: len([]rune(norm)), InTokens: in, OutTok: outTok,
		Blocked: blocked, Reason: reason,
	}
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return pageReceipt{}, "", err
	}
	if err := writeReplacing(receiptPath(run, sha, page), append(b, '\n')); err != nil {
		return pageReceipt{}, "", err
	}
	return r, norm, nil
}

// BlockedPages is how many pages of this reading the content filter refused. Derived from
// the pages slice rather than stored — a count field beside it would be a second copy of
// the same fact, free to disagree.
func (r ReadingRecord) BlockedPages() int {
	n := 0
	for _, p := range r.Pages {
		if p.Blocked {
			n++
		}
	}
	return n
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
// it is. What this offers instead is: which model, at what time, against which image hashes.
// That is an attestation of provenance, NOT of accuracy — one reading, uncorroborated, and no
// field on this record should be read as saying otherwise. It is weaker than re-derivation
// and it is stated rather than left to be discovered.
type ReadingRecord struct {
	Sha   string `json:"sha"`
	Model string `json:"model"`
	// ReadAt is when. A model changes underneath a fixed name, so the name alone does not
	// identify what did the reading.
	ReadAt time.Time `json:"read_at"`
	// RenderShas binds this reading to the exact images it read. On the deliberate path those
	// images are on disk under `ocr pages`; on fetch's automatic path they were released as
	// they were read and these hashes are the only identity the pixels have. A re-render at a
	// new DPI makes the reading stale, and this is what lets a reader notice rather than assume.
	RenderShas []string `json:"render_shas"`
	// DPI the images were rendered at, carried so a poor reading can be attributed.
	DPI      int           `json:"dpi"`
	Pages    []PageReading `json:"pages"`
	TextSha  string        `json:"text_sha"`
	InTokens int64         `json:"input_tokens"`
	OutTok   int64         `json:"output_tokens"`
}

func readingRecordPath(run record.Run, sha string) string {
	return filepath.Join(PagesDir(run, sha), "reading.json")
}

// PageTextPath is one page's transcription, beside the image it was read from.
//
// THE PER-PAGE FILES ARE KEPT EVEN THOUGH THE ASSEMBLED TEXT CONTAINS THEM. A citation lands
// on a page, and a reader checking one against the scan wants that page's text next to that
// page's image — not an offset into a document-length file they must count into.
func PageTextPath(run record.Run, sha string, page int) string {
	return filepath.Join(PagesDir(run, sha), fmt.Sprintf("p%04d.txt", page))
}

// OCRTextPath is the assembled reading, beside the extraction TextPath would hold.
// The suffix says which is which to a human listing the directory; no reader recovers a fact
// from it — `ocr_derived` on the record is what states the difference.
func OCRTextPath(run record.Run, sha string) string { return Path(run, sha) + ".ocr.txt" }

// ReadRenderedPages reads every rendered page and records what came back.
func ReadRenderedPages(ctx context.Context, run record.Run, sha, model string, rd RenderRecord) (ReadingRecord, error) {
	if rd.Pages() == 0 {
		return ReadingRecord{}, fmt.Errorf("the render record for %s names no pages", sha)
	}
	if rd.Pages() > MaxReadPages {
		return ReadingRecord{}, fmt.Errorf("%d pages is over the %d-page cap for one reading, and each "+
			"page costs a model call. Re-render a subset, or raise the cap deliberately — this refuses "+
			"rather than reading part of the document and calling it the document",
			rd.Pages(), MaxReadPages)
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
		// The shared per-page step: reuse a matching receipt without a call, turn a
		// content-filter refusal into a NAMED hole, keep every other error fatal (#679).
		r, norm, rerr := readPageStep(ctx, run, sha, i, png, model, rd.DPI)
		if rerr != nil {
			return ReadingRecord{}, fmt.Errorf("page %d: %w", i, rerr)
		}
		out.InTokens += r.InTokens
		out.OutTok += r.OutTok
		out.Pages = append(out.Pages, PageReading{
			Page: i, TextSha: r.TextSha, Length: r.Length,
			Blocked: r.Blocked, BlockedReason: r.Reason,
		})

		assembled.WriteString(norm)
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

// normalizeReading settles line endings and strips trailing whitespace.
//
// THIS IS HYGIENE ON STORED TEXT, NOT INTERPRETATION OF IT. It served a comparison once and no
// longer does; what remains is that a transcription is about to become a citable artifact, and
// CRLF or a trailing run of spaces in it is noise a reader would have to look past. Interior
// line structure is PRESERVED — paragraph breaks are how the page reads, and flattening them
// would be this function editing the content rather than tidying its edges.
func normalizeReading(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}
