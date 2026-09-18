// Command corpusadd puts ONE page of a scan into the corpus: it cuts the page out, removes its
// text layer so the reader meets pixels, gzips it, and writes the record beside it.
//
// It is a command and not a test on purpose. A test driven by environment variables reads a
// misspelled name as a SKIP, and a skip is green — the failure mode this corpus exists to refuse.
// Here a missing or malformed input is a refusal that names the flag, and NOTHING is written.
//
// The file is one the operator has already downloaded: nothing here fetches -url, so nothing here
// can attest the bytes came from it. The record says so in the field's name.
//
// Run from the tools module:
//
//	go run ./internal/tessocr/testdata/corpusadd \
//	  -file ~/dl/sp602.pdf -url https://… -page 44 -slug nbs602-dashed-matrix \
//	  -publisher "National Bureau of Standards" -date 1981-04 \
//	  -rights us-government-work -rights-evidence "https://…" \
//	  -defect dashed-rules -why "rules drawn as 20 px dashes"
//
// It writes page.pdf.gz and provenance.json with an EMPTY expect block, which every gate then
// refuses until a human transcribes from the page image what a correct reader would produce.
package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/enums"
	"github.com/klippa-app/go-pdfium/references"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/tessocr/corpus"
)

func main() {
	var (
		file      = flag.String("file", "", "the file you downloaded (PDF, or a JPEG for a page published as an image)")
		url       = flag.String("url", "", "where you downloaded it from — recorded, never fetched")
		page      = flag.Int("page", 0, "which page of it, 1-based (1 for a JPEG)")
		slug      = flag.String("slug", "", "the corpus directory name")
		publisher = flag.String("publisher", "", "who published it")
		date      = flag.String("date", "", "when, as printed (1981-04)")
		rights    = flag.String("rights", "", "one of: "+strings.Join(corpus.Rights, ", "))
		evidence  = flag.String("rights-evidence", "", "a URL or a quoted licence line")
		caveat    = flag.String("rights-caveat", "", "anything a reader of the licence should know")
		blankOK   = flag.Bool("blank-ok", false, "accept a page with almost no ink — say so deliberately, as for a blank page kept as a specimen")
		defects   = flag.String("defect", "", "comma-separated defect classes this page is a specimen of")
		why       = flag.String("why", "", "one sentence: what this page does to the reader")
		out       = flag.String("out", "internal/tessocr/testdata/corpus", "the corpus root")
	)
	flag.Parse()

	if err := run(*file, *url, *page, *slug, *publisher, *date, *rights, *evidence, *caveat, *defects, *why, *out, *blankOK); err != nil {
		fmt.Fprintf(os.Stderr, "corpusadd: %v\n", err)
		os.Exit(1)
	}
}

func run(file, url string, page int, slug, publisher, date, rights, evidence, caveat, defects, why, out string, blankOK bool) error {
	// Every refusal comes BEFORE a byte is written: an unrightsed page never enters a commit.
	var missing []string
	for _, f := range []struct{ name, val string }{
		{"-file", file}, {"-url", url}, {"-slug", slug}, {"-publisher", publisher},
		{"-date", date}, {"-rights", rights}, {"-rights-evidence", evidence},
		{"-defect", defects}, {"-why", why},
	} {
		if strings.TrimSpace(f.val) == "" {
			missing = append(missing, f.name)
		}
	}
	if page < 1 {
		missing = append(missing, "-page")
	}
	if len(missing) > 0 {
		return fmt.Errorf("these are required and were not given: %s", strings.Join(missing, " "))
	}
	if !allowed(rights) {
		return fmt.Errorf("-rights %q is not one of %s — a licence we cannot name is a licence we do not have",
			rights, strings.Join(corpus.Rights, ", "))
	}

	body, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(body)
	local := hex.EncodeToString(sum[:])

	pdf, removed, err := onePage(body, page)
	if err != nil {
		return err
	}
	ink, err := inkFraction(pdf)
	if err != nil {
		return err
	}
	// A page whose CONTENT was the text layer leaves almost nothing behind when the layer goes.
	// The threshold is measured, not guessed: over the pages considered for this corpus, ink runs
	// 1.64% (a contents page, the sparsest real one), 3.7% (a dot-matrix listing), 5.1% (a dashed
	// matrix and a photographic cover), 7.5% (a ruled form) and 61.6% (a shaded mark grid), against
	// 0.47% for an Internet Archive journal whose transcription WAS the text layer and which left
	// nothing but speckle behind. 1% separates the two with margin on both sides.
	if ink < 0.01 && !blankOK {
		return fmt.Errorf("after removing the text layer this page is %.2f%% ink — its content WAS "+
			"the text layer, so it is not a scan. Pass -blank-ok only if a blank page is the specimen", ink*100)
	}
	gz, err := squeeze(pdf)
	if err != nil {
		return err
	}
	pageSum := sha256.Sum256(pdf)

	dir := filepath.Join(out, slug)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("%s already exists — remove it by hand if you mean to replace it", dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, corpus.FilePage), gz, 0o644); err != nil {
		return err
	}
	p := corpus.Provenance{
		Slug: slug, SourceURL: url, SourcePage: page, Publisher: publisher, Date: date,
		Rights: rights, RightsNote: evidence, RightsCaveat: caveat,
		LocalFileSha256: local, PageSha256: hex.EncodeToString(pageSum[:]),
		TextObjectsRemove: removed, AddedAt: time.Now().UTC().Format("2006-01-02"),
		DefectClass: split(defects), Why: why,
	}
	if err := corpus.Write(dir, p); err != nil {
		return err
	}
	fmt.Printf("%s: page %d, %d text objects removed, %.1f%% ink, %d bytes gzipped\n", slug, page, removed, ink*100, len(gz))
	fmt.Printf("NOW FILL IN expect IN %s — until you do, the corpus gates refuse this page, and\n",
		filepath.Join(dir, corpus.FileRecord))
	fmt.Printf("then regenerate the README, the golden and STATUS.md with the tagged harness.\n")
	return nil
}

func allowed(s string) bool {
	for _, r := range corpus.Rights {
		if r == s {
			return true
		}
	}
	return false
}

func split(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func squeeze(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	// BestCompression and no header fields, so the same page gzips to the same bytes on any box.
	zw, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(b); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// onePage returns a one-page, pixels-only PDF holding page n of body, and how many text objects it
// dropped. A JPEG is wrapped without re-encoding, so in both paths the bytes are the publisher's.
func onePage(body []byte, n int) ([]byte, int, error) {
	if !bytes.HasPrefix(body, []byte("%PDF")) {
		if n != 1 {
			return nil, 0, fmt.Errorf("-page %d: an image file has one page", n)
		}
		pdf, err := wrapJPEG(body)
		return pdf, 0, err
	}
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 1})
	if err != nil {
		return nil, 0, err
	}
	defer pool.Close()
	inst, err := pool.GetInstance(0)
	if err != nil {
		return nil, 0, err
	}
	defer inst.Close()
	src, err := inst.OpenDocument(&requests.OpenDocument{File: &body})
	if err != nil {
		return nil, 0, err
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: src.Document})
	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: src.Document})
	if err != nil {
		return nil, 0, err
	}
	if n > pc.PageCount {
		return nil, 0, fmt.Errorf("-page %d: the document has %d pages", n, pc.PageCount)
	}
	dst, err := inst.FPDF_CreateNewDocument(&requests.FPDF_CreateNewDocument{})
	if err != nil {
		return nil, 0, err
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: dst.Document})
	rng := fmt.Sprint(n)
	if _, err := inst.FPDF_ImportPages(&requests.FPDF_ImportPages{
		Source: src.Document, Destination: dst.Document, PageRange: &rng, Index: 0,
	}); err != nil {
		return nil, 0, err
	}
	removed, err := stripText(inst, dst.Document)
	if err != nil {
		return nil, 0, err
	}
	var buf bytes.Buffer
	if _, err := inst.FPDF_SaveAsCopy(&requests.FPDF_SaveAsCopy{Document: dst.Document, FileWriter: &buf}); err != nil {
		return nil, 0, err
	}
	return buf.Bytes(), removed, nil
}

// stripText removes every TEXT object from every page: a published scan usually carries the
// publisher's own OCR as invisible text, and with it present the reader extracts and never looks at
// the pixels the page was kept for.
func stripText(inst pdfium.Pdfium, doc references.FPDF_DOCUMENT) (int, error) {
	pc, err := inst.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc})
	if err != nil {
		return 0, err
	}
	removed := 0
	for i := 0; i < pc.PageCount; i++ {
		page := requests.Page{ByIndex: &requests.PageByIndex{Document: doc, Index: i}}
		n, cerr := inst.FPDFPage_CountObjects(&requests.FPDFPage_CountObjects{Page: page})
		if cerr != nil {
			return removed, cerr
		}
		// Backwards: removing an object renumbers everything after it.
		for j := n.Count - 1; j >= 0; j-- {
			o, gerr := inst.FPDFPage_GetObject(&requests.FPDFPage_GetObject{Page: page, Index: j})
			if gerr != nil {
				return removed, gerr
			}
			ty, terr := inst.FPDFPageObj_GetType(&requests.FPDFPageObj_GetType{PageObject: o.PageObject})
			if terr != nil {
				return removed, terr
			}
			if ty.Type != enums.FPDF_PAGEOBJ_TEXT {
				continue
			}
			if _, rerr := inst.FPDFPage_RemoveObject(&requests.FPDFPage_RemoveObject{Page: page, PageObject: o.PageObject}); rerr != nil {
				return removed, rerr
			}
			if _, derr := inst.FPDFPageObj_Destroy(&requests.FPDFPageObj_Destroy{PageObject: o.PageObject}); derr != nil {
				return removed, derr
			}
			removed++
		}
		if _, gerr := inst.FPDFPage_GenerateContent(&requests.FPDFPage_GenerateContent{Page: page}); gerr != nil {
			return removed, gerr
		}
	}
	return removed, nil
}

// inkFraction renders the page small and returns the share of pixels that are not near-white. It
// is the cheapest answer to "is there anything on this page to read".
func inkFraction(pdf []byte) (float64, error) {
	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 1})
	if err != nil {
		return 0, err
	}
	defer pool.Close()
	inst, err := pool.GetInstance(0)
	if err != nil {
		return 0, err
	}
	defer inst.Close()
	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &pdf})
	if err != nil {
		return 0, err
	}
	defer inst.FPDF_CloseDocument(&requests.FPDF_CloseDocument{Document: doc.Document})
	r, err := inst.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: 0}}, DPI: 72,
	})
	if err != nil {
		return 0, err
	}
	defer r.Cleanup()
	img := r.Result.RenderedImage
	b := img.Bounds()
	dark, total := 0, 0
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			gray := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			if gray.Y < 200 {
				dark++
			}
			total++
		}
	}
	if total == 0 {
		return 0, nil
	}
	return float64(dark) / float64(total), nil
}

// wrapJPEG puts a JPEG into a one-page PDF at 300 DPI, passing the compressed bytes through
// untouched (DCTDecode), so nothing re-encodes the publisher's image.
func wrapJPEG(jpg []byte) ([]byte, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(jpg))
	if err != nil {
		return nil, fmt.Errorf("not a JPEG this can wrap: %w", err)
	}
	w := float64(cfg.Width) * 72 / 300
	h := float64(cfg.Height) * 72 / 300
	content := fmt.Sprintf("q %.2f 0 0 %.2f 0 0 cm /Im0 Do Q\n", w, h)
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.2f %.2f] /Contents 4 0 R /Resources << /XObject << /Im0 5 0 R >> >> >>", w, h),
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content), content),
		fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n%s\nendstream",
			cfg.Width, cfg.Height, len(jpg), jpg),
	}
	var b bytes.Buffer
	b.WriteString("%PDF-1.4\n")
	offs := make([]int, len(objs))
	for i, o := range objs {
		offs[i] = b.Len()
		fmt.Fprintf(&b, "%d 0 obj\n%s\nendobj\n", i+1, o)
	}
	xref := b.Len()
	fmt.Fprintf(&b, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offs {
		fmt.Fprintf(&b, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&b, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)
	return b.Bytes(), nil
}
