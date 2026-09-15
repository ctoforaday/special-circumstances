// Command gen writes internal/seatprobe/testdata/scanned-1p.pdf: one page carrying one sentence as
// PIXELS and no text layer — the document class a citation of OCR text comes from.
//
// It sets the sentence in Helvetica, rasterises that page with PDFium, and embeds the raster as the
// only content of a new page, so the bytes are reproducible from this file and carry no licensed
// page. Run from the tools module: go run ./internal/seatprobe/testdata/gen
package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"

	"github.com/ctoforaday/special-circumstances/plugins/frank-exchange-of-views/tools/internal/seatprobe"
)

const dpi = 200

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run() error {
	vector := pdf([]string{
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}, "BT /F1 30 Tf 60 600 Td ("+seatprobe.ScannedFixtureSentence+") Tj ET", "/Font << /F1 3 0 R >>")

	pool, err := webassembly.Init(webassembly.Config{MinIdle: 1, MaxIdle: 1, MaxTotal: 1})
	if err != nil {
		return err
	}
	defer pool.Close()
	inst, err := pool.GetInstance(0)
	if err != nil {
		return err
	}
	defer inst.Close()
	doc, err := inst.OpenDocument(&requests.OpenDocument{File: &vector})
	if err != nil {
		return err
	}
	r, err := inst.RenderPageInDPI(&requests.RenderPageInDPI{
		Page: requests.Page{ByIndex: &requests.PageByIndex{Document: doc.Document, Index: 0}}, DPI: dpi,
	})
	if err != nil {
		return err
	}
	defer r.Cleanup()
	src := r.Result.RenderedImage
	gray := image.NewGray(src.Bounds())
	draw.Draw(gray, gray.Bounds(), src, src.Bounds().Min, draw.Src)

	var z bytes.Buffer
	zw, _ := zlib.NewWriterLevel(&z, zlib.BestCompression)
	if _, err := zw.Write(gray.Pix); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	w, h := gray.Bounds().Dx(), gray.Bounds().Dy()
	img := fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width %d /Height %d /ColorSpace /DeviceGray "+
		"/BitsPerComponent 8 /Filter /FlateDecode /Length %d >>\nstream\n%s\nendstream", w, h, z.Len(), z.String())
	scan := pdf([]string{img}, "q 612 0 0 792 0 0 cm /Im1 Do Q", "/XObject << /Im1 3 0 R >>")
	return os.WriteFile(filepath.Join("internal", "seatprobe", "testdata", "scanned-1p.pdf"), scan, 0o644)
}

// pdf writes a one-page PDF: catalog, pages, the given resource objects from 3, the page, and its
// content stream last.
func pdf(resources []string, content, resourceDict string) []byte {
	n := 3 + len(resources)
	objs := []string{"<< /Type /Catalog /Pages 2 0 R >>", fmt.Sprintf("<< /Type /Pages /Kids [%d 0 R] /Count 1 >>", n)}
	objs = append(objs, resources...)
	objs = append(objs,
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents %d 0 R /Resources << %s >> >>", n+1, resourceDict),
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content))
	var b strings.Builder
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
	return []byte(b.String())
}
